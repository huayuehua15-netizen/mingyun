// Package httpapi 提供 HTTP 路由与通用中间件。
//
// 这一层只负责协议转换：解析请求、调用服务、写响应。业务规则一律不放在
// 这里，便于后续替换传输层（HTTP、CLI、消息队列）。
package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
)

// Server HTTP 服务。
type Server struct {
	version string
	store   store.ChartStore
	entropy *entropy.Source
	now     func() time.Time
}

// New 构造 HTTP 处理器。
func New(version string, st store.ChartStore, es *entropy.Source) http.Handler {
	s := &Server{
		version: version,
		store:   st,
		entropy: es,
		now:     time.Now,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /version", s.handleVersion)
	mux.HandleFunc("GET /api/v1/meta/capabilities", s.handleCapabilities)
	mux.HandleFunc("GET /api/v1/meta/calendar", s.handleCalendar)
	mux.HandleFunc("POST /api/v1/charts/{method}/cast", s.handleCast)
	mux.HandleFunc("GET /api/v1/charts", s.handleListCharts)
	mux.HandleFunc("GET /api/v1/charts/{id}", s.handleGetChart)
	mux.HandleFunc("/", s.handleNotFound)

	return withRequestID(mux)
}

type healthResponse struct {
	Status     string `json:"status"`
	Version    string `json:"version"`
	ServerTime string `json:"server_time"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:     "ok",
		Version:    s.version,
		ServerTime: s.now().Format(time.RFC3339),
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": s.version})
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, r, &apiError{Status: http.StatusNotFound, Code: 10301, Message: "接口不存在"})
}

// envelope 统一响应信封，字段名与 docs/05 1.1 一致。
type envelope struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	RequestID  string `json:"request_id"`
	ServerTime string `json:"server_time"`
}

// apiError 一个可以直接写给客户端的错误。
type apiError struct {
	Status  int
	Code    int
	Message string
	Data    any
}

func badRequest(msg string, fields map[string]string) *apiError {
	return &apiError{Status: http.StatusBadRequest, Code: 10001, Message: msg, Data: map[string]any{"fields": fields}}
}

func writeOK(w http.ResponseWriter, r *http.Request, data any) {
	writeEnvelope(w, r, http.StatusOK, 0, "ok", data)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, err *apiError) {
	code := err.Code
	if code == 0 {
		code = 20001
	}
	writeEnvelope(w, r, err.Status, code, err.Message, err.Data)
}

func writeEnvelope(w http.ResponseWriter, r *http.Request, status, code int, msg string, data any) {
	writeJSON(w, status, envelope{
		Code:       code,
		Message:    msg,
		Data:       data,
		RequestID:  requestID(r),
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		// 响应头已发出，此处只能记录，不能改状态码。
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
	}
}

type ctxKey int

const ctxKeyRequestID ctxKey = iota

// withRequestID 为每个请求生成追踪 ID，透传到响应头与 context。
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func requestID(r *http.Request) string {
	if id, ok := r.Context().Value(ctxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// newChartID 生成 UUID v4，避免引入额外依赖。
func newChartID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("chart-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
