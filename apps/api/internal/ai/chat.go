package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Message 一条对话消息，字段与 OpenAI Chat Completions 对齐。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage token 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Result 一次流式调用的收尾信息。
type Result struct {
	Usage        Usage
	FinishReason string
}

// Client OpenAI 兼容的流式客户端。
type Client struct {
	cfg  Config
	http *http.Client
}

// NewClient 构造客户端。配置不完整时仍可构造，Enabled() 会返回 false。
func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: timeout},
	}
}

// Enabled 是否已配置可用的在线模型。
func (c *Client) Enabled() bool {
	return c != nil && c.cfg.Enabled()
}

// Model 当前模型名。
func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.cfg.Model
}

// Provider 当前来源标识。
func (c *Client) Provider() string {
	if c == nil {
		return ""
	}
	return c.cfg.Provider()
}

// MaxTokens 环境变量给出的默认输出上限。
func (c *Client) MaxTokens() int {
	if c == nil || c.cfg.MaxTokens <= 0 {
		return defaultMaxTokens
	}
	return c.cfg.MaxTokens
}

// Stream 发起流式请求，每收到一段正文就调用 onDelta。
//
// 只消费 delta.content；reasoning_content 不进入正文。上游的
// stream_options 不强制发送，以兼容更多自部署网关。
func (c *Client) Stream(ctx context.Context, messages []Message, maxTokens int, onDelta func(string) error) (Result, error) {
	if !c.Enabled() {
		return Result{}, fmt.Errorf("AI 未配置")
	}
	if maxTokens <= 0 {
		maxTokens = c.cfg.MaxTokens
	}
	maxTokens = clampInt(maxTokens, 1, maxAllowedTokens)

	body, err := json.Marshal(map[string]any{
		"model":       c.cfg.Model,
		"messages":    messages,
		"stream":      true,
		"max_tokens":  maxTokens,
		"temperature": c.cfg.Temperature,
	})
	if err != nil {
		return Result{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint(), bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("连接模型服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, upstreamError(resp)
	}

	type delta struct {
		Content string `json:"content"`
	}
	type choice struct {
		Delta        delta  `json:"delta"`
		FinishReason string `json:"finish_reason"`
	}
	type chunk struct {
		Choices []choice `json:"choices"`
		Usage   *Usage   `json:"usage"`
		Error   *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	var result Result
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "id:") || strings.HasPrefix(line, "retry:") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			break
		}

		var ck chunk
		if err := json.Unmarshal([]byte(payload), &ck); err != nil {
			return result, fmt.Errorf("解析模型流失败: %w", err)
		}
		if ck.Error != nil && ck.Error.Message != "" {
			return result, fmt.Errorf("模型返回错误: %s", ck.Error.Message)
		}
		if ck.Usage != nil {
			result.Usage = *ck.Usage
		}
		for _, ch := range ck.Choices {
			if ch.FinishReason != "" {
				result.FinishReason = ch.FinishReason
			}
			if ch.Delta.Content != "" && onDelta != nil {
				if err := onDelta(ch.Delta.Content); err != nil {
					return result, err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		return result, fmt.Errorf("读取模型流失败: %w", err)
	}
	if result.FinishReason == "" {
		result.FinishReason = "stop"
	}
	return result, nil
}

// Test 用一条极短消息验证地址、Key、模型名与流式响应是否可用。
//
// 设置页只需要“能不能接通”，不在这里做盘面校验；真正的解读仍走 Stream
// 与 AuditReading。返回值是上游正文预览，供页面显示。
func (c *Client) Test(ctx context.Context) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("AI 未配置")
	}
	var preview strings.Builder
	_, err := c.Stream(ctx, []Message{
		{Role: "user", Content: "只回复两个字：正常"},
	}, 24, func(text string) error {
		preview.WriteString(text)
		return nil
	})
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(preview.String())
	if text == "" {
		return "上游已响应", nil
	}
	return text, nil
}

func upstreamError(resp *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	msg := ""
	if json.Unmarshal(raw, &parsed) == nil {
		msg = strings.TrimSpace(parsed.Error.Message)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}
	if len([]rune(msg)) > 300 {
		msg = string([]rune(msg)[:300]) + "…"
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("模型服务返回 %d: %s", resp.StatusCode, msg)
}
