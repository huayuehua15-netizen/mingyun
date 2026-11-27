package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/ai"
)

type aiSettingsRequest struct {
	ProviderID  string   `json:"provider_id"`
	BaseURL     string   `json:"base_url"`
	Model       string   `json:"model"`
	APIKey      *string  `json:"api_key"`
	Temperature *float64 `json:"temperature"`
	MaxTokens   *int     `json:"max_tokens"`
}

func (s *Server) handleGetAISettings(w http.ResponseWriter, r *http.Request) {
	if s.aiConfig == nil {
		writeAPIError(w, r, &apiError{Status: http.StatusServiceUnavailable, Code: 30004, Message: "服务端未启用模型设置存储"})
		return
	}
	writeOK(w, r, s.aiConfig.View())
}

func (s *Server) handleSaveAISettings(w http.ResponseWriter, r *http.Request) {
	if s.aiConfig == nil {
		writeAPIError(w, r, &apiError{Status: http.StatusServiceUnavailable, Code: 30004, Message: "服务端未启用模型设置存储"})
		return
	}
	if !s.aiConfig.Writable() {
		writeAPIError(w, r, &apiError{Status: http.StatusForbidden, Code: 30005, Message: "模型设置已锁定，只能由环境变量配置"})
		return
	}
	req, apiErr := decodeAISettingsRequest(w, r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sett, apiErr := resolveAISettings(req, s.aiConfig.Settings())
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.aiConfig.Update(sett); err != nil {
		writeAPIError(w, r, badRequest(err.Error(), nil))
		return
	}
	s.replaceAIClient(s.aiConfig.Config())
	writeOK(w, r, s.aiConfig.View())
}

func (s *Server) handleClearAISettings(w http.ResponseWriter, r *http.Request) {
	if s.aiConfig == nil {
		writeAPIError(w, r, &apiError{Status: http.StatusServiceUnavailable, Code: 30004, Message: "服务端未启用模型设置存储"})
		return
	}
	if !s.aiConfig.Writable() {
		writeAPIError(w, r, &apiError{Status: http.StatusForbidden, Code: 30005, Message: "模型设置已锁定，只能由环境变量配置"})
		return
	}
	if err := s.aiConfig.Clear(); err != nil {
		writeAPIError(w, r, &apiError{Status: http.StatusInternalServerError, Code: 30006, Message: "清除模型配置失败: " + err.Error()})
		return
	}
	s.replaceAIClient(s.aiConfig.Config())
	writeOK(w, r, s.aiConfig.View())
}

func (s *Server) handleTestAISettings(w http.ResponseWriter, r *http.Request) {
	if s.aiConfig == nil {
		writeAPIError(w, r, &apiError{Status: http.StatusServiceUnavailable, Code: 30004, Message: "服务端未启用模型设置存储"})
		return
	}
	req, apiErr := decodeAISettingsRequest(w, r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sett, apiErr := resolveAISettings(req, s.aiConfig.Settings())
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	cfg := ai.Config{
		BaseURL:     sett.BaseURL,
		APIKey:      sett.APIKey,
		Model:       sett.Model,
		Timeout:     30 * time.Second,
		MaxTokens:   sett.MaxTokens,
		Temperature: sett.Temperature,
	}
	client := ai.NewClient(cfg)
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	started := time.Now()
	preview, err := client.Test(ctx)
	latency := time.Since(started).Milliseconds()
	result := map[string]any{
		"ok":         err == nil,
		"latency_ms": latency,
		"provider":   cfg.Provider(),
		"model":      cfg.Model,
	}
	if err != nil {
		log.Printf("ai settings test provider=%s model=%s error=%v", cfg.Provider(), cfg.Model, err)
		result["message"] = modelTestErrorText(err)
		writeOK(w, r, result)
		return
	}
	result["message"] = "连接正常"
	result["preview"] = clipText(preview, 120)
	writeOK(w, r, result)
}

func decodeAISettingsRequest(w http.ResponseWriter, r *http.Request) (aiSettingsRequest, *apiError) {
	var req aiSettingsRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		return req, badRequest("请求体不是合法 JSON", nil)
	}
	return req, nil
}

// resolveAISettings 把设置页输入折算成一份完整配置。
//
// 空 API Key 是“沿用已保存的 Key”，但换了接口地址后不会沿用旧 Key，
// 避免把 A 服务商的 Key 发给 B 服务商。本机/容器宿主机端点允许无 Key。
func resolveAISettings(req aiSettingsRequest, current ai.Settings) (ai.Settings, *apiError) {
	providerID := strings.TrimSpace(strings.ToLower(req.ProviderID))
	preset, hasPreset := ai.FindProviderPreset(providerID)
	base := ai.NormalizeBaseURL(req.BaseURL)
	if base == "" && hasPreset {
		base = ai.NormalizeBaseURL(preset.BaseURL)
	}
	model := strings.TrimSpace(req.Model)
	if model == "" && hasPreset && ai.NormalizeBaseURL(preset.BaseURL) == base {
		model = strings.TrimSpace(preset.DefaultModel)
	}

	if base == "" && model != "" {
		return ai.Settings{}, badRequest("填写模型名时必须填写接口地址", map[string]string{"base_url": "empty"})
	}
	if base != "" && model == "" {
		return ai.Settings{}, badRequest("请填写模型名；方舟等部分服务商需填推理接入点 ID", map[string]string{"model": "empty"})
	}
	if base != "" {
		inferred := ai.InferProviderID(base)
		if !hasPreset || ai.NormalizeBaseURL(preset.BaseURL) != base {
			providerID = inferred
		}
	} else {
		providerID = ""
	}

	key := strings.TrimSpace(current.APIKey)
	if req.APIKey != nil {
		key = strings.TrimSpace(*req.APIKey)
	} else if base == "" || ai.NormalizeBaseURL(current.BaseURL) != base {
		key = ""
	}
	if base == "" {
		key = ""
	}
	if base != "" && key == "" && !ai.IsLocalBaseURL(base) {
		return ai.Settings{}, badRequest("请填写 API Key", map[string]string{"api_key": "empty"})
	}

	temperature := current.Temperature
	if req.Temperature != nil {
		temperature = *req.Temperature
	}
	if temperature < 0 || temperature > 2 {
		return ai.Settings{}, badRequest("temperature 应在 0 到 2 之间", map[string]string{"temperature": "out of range"})
	}
	maxTokens := current.MaxTokens
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	if maxTokens <= 0 {
		maxTokens = 1200
	}
	if maxTokens > 8192 {
		maxTokens = 8192
	}

	return ai.Settings{
		ProviderID:  providerID,
		BaseURL:     base,
		APIKey:      key,
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}, nil
}

func modelTestErrorText(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "连接超时：检查接口地址、网络或代理"
	case strings.Contains(err.Error(), "401"), strings.Contains(err.Error(), "403"):
		return "鉴权失败：检查 API Key 是否正确、是否有该模型权限"
	case strings.Contains(err.Error(), "404"):
		return "接口不存在：检查地址是否少了 /v1 或模型名是否填错"
	case strings.Contains(err.Error(), "429"):
		return "请求过多或额度不足：稍后再试，或换一个模型"
	case strings.Contains(err.Error(), "模型服务返回 5"):
		return "服务商暂时不可用：稍后再试"
	default:
		return "连接失败：检查接口地址、模型名与网络"
	}
}

func clipText(s string, limit int) string {
	s = strings.TrimSpace(s)
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit]) + "…"
}
