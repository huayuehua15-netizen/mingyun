package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/ai"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

type interpretRequest struct {
	Question  string `json:"question"`
	Style     string `json:"style"`
	MaxTokens int    `json:"max_tokens"`
}

type interpretMeta struct {
	MessageID string `json:"message_id"`
	Mode      string `json:"mode"`
	Model     string `json:"model"`
	Provider  string `json:"provider"`
	ChartID   string `json:"chart_id"`
}

type interpretDelta struct {
	Text string `json:"text"`
}

type interpretDone struct {
	MessageID    string `json:"message_id"`
	FinishReason string `json:"finish_reason"`
}

type interpretError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleInterpret 首轮解读，SSE 流式返回。
//
// 有模型时走 OpenAI 兼容接口；没有模型时走离线摘要。两种模式共用同一套
// meta / delta / usage / done / error 事件，前端不需要分支。
func (s *Server) handleInterpret(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rec, ok := s.store.Get(id)
	if !ok {
		writeAPIError(w, r, &apiError{Status: http.StatusNotFound, Code: 10301, Message: "盘面不存在"})
		return
	}

	var req interpretRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeAPIError(w, r, badRequest("请求体不是合法 JSON", nil))
		return
	}

	style := strings.TrimSpace(req.Style)
	if style == "" {
		style = "plain"
	}
	if style != "plain" && style != "classic" {
		writeAPIError(w, r, badRequest("style 只能是 plain 或 classic", map[string]string{"style": style}))
		return
	}

	rawQuestion := strings.TrimSpace(req.Question)
	if rawQuestion == "" {
		rawQuestion = strings.TrimSpace(rec.Question)
	}
	if rawQuestion == "" && rec.Chart.Method == schema.MethodZiWei {
		rawQuestion = "请以这张命盘，看整体格局、当下值得注意的方向。"
	}
	if rawQuestion == "" {
		writeAPIError(w, r, badRequest("请先填写所问之事", map[string]string{"question": "empty"}))
		return
	}
	if utf8.RuneCountInString(rawQuestion) > 200 {
		writeAPIError(w, r, badRequest("问题不得超过 200 字", map[string]string{"question": "too long"}))
		return
	}
	question := ai.SanitizeQuestion(rawQuestion)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAPIError(w, r, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "当前连接不支持流式输出"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	messageID := newChartID()
	ctx := r.Context()

	if s.aiClient == nil || !s.aiClient.Enabled() {
		meta := interpretMeta{
			MessageID: messageID,
			Mode:      "offline",
			Model:     ai.OfflineModel,
			Provider:  "local",
			ChartID:   rec.ChartID,
		}
		if err := writeSSE(w, flusher, "meta", meta); err != nil {
			return
		}
		text := ai.BuildOfflineReading(rec.Chart, question)
		for _, chunk := range chunkRunes(text, 28) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(18 * time.Millisecond):
			}
			if err := writeSSE(w, flusher, "delta", interpretDelta{Text: chunk}); err != nil {
				return
			}
		}
		_ = writeSSE(w, flusher, "done", interpretDone{MessageID: messageID, FinishReason: "offline"})
		return
	}

	meta := interpretMeta{
		MessageID: messageID,
		Mode:      "ai",
		Model:     s.aiClient.Model(),
		Provider:  s.aiClient.Provider(),
		ChartID:   rec.ChartID,
	}
	if err := writeSSE(w, flusher, "meta", meta); err != nil {
		return
	}

	maxTokens := s.aiClient.MaxTokens()
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	// 紫微/大六壬的盘面事实比六爻长，默认 1800 tokens 容易在末段被截断。
	switch rec.Chart.Method {
	case schema.MethodZiWei:
		if maxTokens < 2400 {
			maxTokens = 2400
		}
	case schema.MethodDaLiuRen:
		if maxTokens < 2200 {
			maxTokens = 2200
		}
	case schema.MethodQiMen:
		if maxTokens < 2600 {
			maxTokens = 2600
		}
	case schema.MethodMeiHua:
		if maxTokens < 2200 {
			maxTokens = 2200
		}
	case schema.MethodXiaoLiuRen:
		if maxTokens < 1800 {
			maxTokens = 1800
		}
	}
	if maxTokens > 8192 {
		maxTokens = 8192
	}

	messages := ai.BuildMessages(rec.Chart, question, style)

	// 先缓冲完整输出并做盘面事实校验，避免模型编造的卦名、动爻或干支
	// 已经流到浏览器之后才被发现。校验通过后再按原 SSE 协议分片下发。
	var raw strings.Builder
	result, err := s.aiClient.Stream(ctx, messages, maxTokens, func(text string) error {
		raw.WriteString(text)
		return nil
	})
	if err != nil {
		log.Printf("interpret chart=%s error=%v", rec.ChartID, err)
		_ = writeSSE(w, flusher, "error", interpretError{Code: 30001, Message: interpretErrorText(err)})
		return
	}

	if issues := ai.AuditReading(raw.String(), rec.Chart); len(issues) > 0 {
		log.Printf("interpret chart=%s audit rejected: %s", rec.ChartID, strings.Join(issues, "；"))
		_ = writeSSE(w, flusher, "error", interpretError{
			Code:    30002,
			Message: "解读未通过盘面事实校验，已拦截。请重试，或稍后使用离线摘要。",
		})
		return
	}

	text := raw.String()
	for _, chunk := range chunkRunes(text, 28) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(18 * time.Millisecond):
		}
		if err := writeSSE(w, flusher, "delta", interpretDelta{Text: chunk}); err != nil {
			return
		}
	}

	if result.Usage.PromptTokens > 0 || result.Usage.CompletionTokens > 0 {
		if err := writeSSE(w, flusher, "usage", result.Usage); err != nil {
			return
		}
	}
	finish := result.FinishReason
	if finish == "" {
		finish = "stop"
	}
	_ = writeSSE(w, flusher, "done", interpretDone{MessageID: messageID, FinishReason: finish})
}

// interpretErrorText 把上游错误翻成可展示的话，不把内网地址、原始响应带给浏览器。
func interpretErrorText(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "模型响应超时，请稍后重试。"
	case strings.Contains(err.Error(), "401"), strings.Contains(err.Error(), "403"), strings.Contains(err.Error(), "鉴权"):
		return "模型服务鉴权失败，请检查服务端 MINGYUN_AI_API_KEY。"
	case strings.Contains(err.Error(), "模型服务返回"):
		return "模型服务返回错误，请查看服务端日志。"
	default:
		return "模型服务暂时不可用，请稍后重试。"
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func chunkRunes(s string, size int) []string {
	if size <= 0 {
		size = 1
	}
	runes := []rune(s)
	out := make([]string, 0, (len(runes)+size-1)/size)
	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[i:end]))
	}
	return out
}
