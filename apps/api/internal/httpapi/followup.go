package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/ai"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
	"github.com/huayuehua15-netizen/mingyun/paipan/daliuren"
	"github.com/huayuehua15-netizen/mingyun/paipan/liuyao"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
	"github.com/huayuehua15-netizen/mingyun/paipan/ziwei"
)

const maxFollowupRounds = 6

type followupTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type followupRequest struct {
	Message   string         `json:"message"`
	History   []followupTurn `json:"history"`
	Chart     *schema.Chart  `json:"chart"`
	Style     string         `json:"style"`
	MaxTokens int            `json:"max_tokens"`
}

// handleFollowup 在同一张盘内追问，SSE 流式返回。
//
// 正常情况下按 chart_id 从服务端内存取盘；如果 API 重启导致旧盘不在内存，
// 允许请求携带本地保存的 chart 快照恢复上下文。快照只用于构造提示词，
// 不会写回 store，也不会改变盘面事实。
func (s *Server) handleFollowup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req followupRequest
	r.Body = http.MaxBytesReader(w, r.Body, 3<<20)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeAPIError(w, r, badRequest("请求体不是合法 JSON", nil))
		return
	}

	rec, found := s.store.Get(id)
	if !found {
		if req.Chart == nil || len(req.Chart.Facts) == 0 || req.Chart.Payload == nil {
			writeAPIError(w, r, &apiError{Status: http.StatusNotFound, Code: 10301, Message: "盘面不存在或已失效，请重新起卦/起课/排盘"})
			return
		}
		if req.Chart.ChartID != "" && req.Chart.ChartID != id {
			writeAPIError(w, r, badRequest("chart_id 与盘面快照不一致", map[string]string{"chart_id": req.Chart.ChartID}))
			return
		}
		if err := restoreChartFacts(req.Chart); err != nil {
			writeAPIError(w, r, badRequest("盘面快照无法恢复完整事实", map[string]string{"chart": "invalid payload"}))
			return
		}
		req.Chart.ChartID = id
		rec = store.Record{ChartID: id, Chart: *req.Chart}
	}

	style := strings.TrimSpace(req.Style)
	if style == "" {
		style = "plain"
	}
	if style != "plain" && style != "classic" {
		writeAPIError(w, r, badRequest("style 只能是 plain 或 classic", map[string]string{"style": style}))
		return
	}

	rawQuestion := strings.TrimSpace(req.Message)
	if rawQuestion == "" {
		writeAPIError(w, r, badRequest("请先填写追问", map[string]string{"message": "empty"}))
		return
	}
	if utf8.RuneCountInString(rawQuestion) > 200 {
		writeAPIError(w, r, badRequest("追问不得超过 200 字", map[string]string{"message": "too long"}))
		return
	}
	question := ai.SanitizeQuestion(rawQuestion)

	history, userTurns := normalizeFollowupHistory(req.History)
	if userTurns >= maxFollowupRounds {
		writeAPIError(w, r, badRequest("这张盘已追问 6 轮，建议另起一卦/课/盘", map[string]string{"history": "round limit"}))
		return
	}

	if s.aiClient == nil || !s.aiClient.Enabled() {
		writeAPIError(w, r, &apiError{Status: http.StatusServiceUnavailable, Code: 30003, Message: "当前未配置模型，无法追问；首轮离线摘要仍可查看。"})
		return
	}

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
	if maxTokens > 2048 {
		maxTokens = 2048
	}

	messages := ai.BuildFollowupMessages(rec.Chart, history, question, style)

	// 和首轮解读一样，先缓冲完整回答再做事实校验，避免编造的爻位或干支
	// 已经流到前端。追问的校验不要求四段标题。
	var raw strings.Builder
	result, err := s.aiClient.Stream(ctx, messages, maxTokens, func(text string) error {
		raw.WriteString(text)
		return nil
	})
	if err != nil {
		log.Printf("followup chart=%s error=%v", rec.ChartID, err)
		_ = writeSSE(w, flusher, "error", interpretError{Code: 30001, Message: interpretErrorText(err)})
		return
	}

	if issues := ai.AuditFollowup(raw.String(), rec.Chart); len(issues) > 0 {
		log.Printf("followup chart=%s audit rejected: %s", rec.ChartID, strings.Join(issues, "；"))
		_ = writeSSE(w, flusher, "error", interpretError{
			Code:    30002,
			Message: "追问未通过盘面事实校验，已拦截。请换个问法，或重新解读这张盘。",
		})
		return
	}

	text := raw.String()
	for _, chunk := range chunkRunes(text, 24) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(14 * time.Millisecond):
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

func normalizeFollowupHistory(in []followupTurn) ([]ai.DialogueTurn, int) {
	out := make([]ai.DialogueTurn, 0, len(in))
	userTurns := 0
	for _, turn := range in {
		role := strings.TrimSpace(turn.Role)
		content := sanitizeDialogue(turn.Content)
		if content == "" || (role != "user" && role != "assistant") {
			continue
		}
		if role == "user" {
			userTurns++
		}
		out = append(out, ai.DialogueTurn{Role: role, Content: content})
	}
	if len(out) > 12 {
		out = out[len(out)-12:]
	}
	return out, userTurns
}

func sanitizeDialogue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "<<<", "‹‹‹")
	s = strings.ReplaceAll(s, ">>>", "›››")
	runes := []rune(s)
	if len(runes) > 2500 {
		s = string(runes[:2500])
	}
	return s
}

// restoreChartFacts 从本地盘面快照的 payload 重建完整 facts。前端正常起卦
// 只拿到 core facts；若服务端内存已丢失，直接拿 core facts 做追问校验会
// 误判模型提到的爻位干支为“事实之外”。这里按原引擎规则重建一次。
func restoreChartFacts(chart *schema.Chart) error {
	raw, err := json.Marshal(chart.Payload)
	if err != nil {
		return err
	}

	switch chart.Method {
	case schema.MethodDaLiuRen:
		var payload daliuren.Payload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		if len(payload.TianPan) != 12 || len(payload.SiKe) != 4 || len(payload.SanChuan) != 3 {
			return errors.New("大六壬盘面快照不完整")
		}
		facts := daliuren.BuildFacts(payload)
		if len(facts) == 0 {
			return errors.New("大六壬盘面快照重建 facts 为空")
		}
		chart.Facts = facts
	case schema.MethodZiWei:
		var payload ziwei.Payload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		if len(payload.Palaces) != 12 || !payload.MingGong.Valid() {
			return errors.New("紫微盘面快照不完整")
		}
		facts := ziwei.BuildFacts(payload)
		if len(facts) == 0 {
			return errors.New("紫微盘面快照重建 facts 为空")
		}
		chart.Facts = facts
	default:
		var payload liuyao.Payload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		if payload.GuaName == "" || len(payload.Lines) == 0 {
			return errors.New("盘面快照缺少卦名或六爻")
		}
		facts := liuyao.BuildFacts(payload)
		if len(facts) == 0 {
			return errors.New("盘面快照重建 facts 为空")
		}
		chart.Facts = facts
	}
	return nil
}
