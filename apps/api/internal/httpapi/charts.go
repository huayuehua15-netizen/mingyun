package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/liuyao"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// castRequest 六爻排盘请求，字段见 docs/05 3.3。
type castRequest struct {
	Mode             string            `json:"mode"`
	Question         string            `json:"question"`
	Topic            string            `json:"topic,omitempty"`
	Gender           string            `json:"gender,omitempty"`
	At               string            `json:"at"`
	TZ               string            `json:"tz"`
	Lon              *float64          `json:"lon"`
	UseTrueSolarTime *bool             `json:"use_true_solar_time"`
	DayBoundary      string            `json:"day_boundary"`
	Lines            []string          `json:"lines"`
	Coins            *[6][3]int        `json:"coins"`
	Backs            *[6]int           `json:"backs"`
	School           map[string]string `json:"school"`
}

// chartView 返回给前端的盘面。fullFacts=false 时只回 core 事实。
type chartView struct {
	ChartID       string            `json:"chart_id"`
	Method        schema.Method     `json:"method"`
	EngineVersion string            `json:"engine_version"`
	Question      string            `json:"question,omitempty"`
	Calendar      calendar.Context  `json:"calendar"`
	School        map[string]string `json:"school,omitempty"`
	Payload       any               `json:"payload"`
	Facts         []schema.Fact     `json:"facts"`
	Warnings      []string          `json:"warnings,omitempty"`
	Entropy       *entropy.Toss     `json:"entropy,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

func (s *Server) handleCast(w http.ResponseWriter, r *http.Request) {
	method, parseErr := schema.ParseMethod(r.PathValue("method"))
	if parseErr != nil {
		writeAPIError(w, r, badRequest("不支持的术数", map[string]string{"method": r.PathValue("method")}))
		return
	}
	if method != schema.MethodLiuYao {
		writeAPIError(w, r, &apiError{
			Status:  http.StatusForbidden,
			Code:    10202,
			Message: "该术数尚未开放",
			Data:    map[string]any{"method": method},
		})
		return
	}

	var req castRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		writeAPIError(w, r, badRequest("请求体不是合法 JSON", nil))
		return
	}

	if utf8.RuneCountInString(req.Question) > 200 {
		writeAPIError(w, r, badRequest("问题不得超过 200 字", map[string]string{"question": "too long"}))
		return
	}

	mode := liuyao.Mode(req.Mode)
	if mode != liuyao.ModeCoin && mode != liuyao.ModeManual {
		writeAPIError(w, r, badRequest("mode 只能是 coin 或 manual", map[string]string{"mode": req.Mode}))
		return
	}

	in := liuyao.Input{
		Mode:     mode,
		Question: strings.TrimSpace(req.Question),
		Topic:    strings.TrimSpace(req.Topic),
		Gender:   strings.TrimSpace(req.Gender),
	}
	var toss *entropy.Toss

	switch mode {
	case liuyao.ModeManual:
		if len(req.Lines) != 6 {
			writeAPIError(w, r, badRequest("手动录入需提供 6 个 lines", map[string]string{"lines": "need 6"}))
			return
		}
		for _, s := range req.Lines {
			kind, err := liuyao.ParseLineKind(strings.TrimSpace(s))
			if err != nil {
				writeAPIError(w, r, badRequest("无法识别的爻形："+s, map[string]string{"lines": s}))
				return
			}
			in.Lines = append(in.Lines, kind)
		}
	case liuyao.ModeCoin:
		switch {
		case req.Coins != nil && req.Backs != nil:
			writeAPIError(w, r, badRequest("coins 与 backs 只能提供一个", nil))
			return
		case req.Coins != nil:
			backs, apiErr := backsFromCoins(*req.Coins)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			in.Backs = backs
		case req.Backs != nil:
			for i, v := range *req.Backs {
				if v < 0 || v > 3 {
					writeAPIError(w, r, badRequest("背数必须在 0—3 之间", map[string]string{"backs": strconv.Itoa(i + 1)}))
					return
				}
			}
			in.Backs = *req.Backs
		default:
			if s.entropy == nil {
				writeAPIError(w, r, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "随机源未配置"})
				return
			}
			t, err := s.entropy.Toss()
			if err != nil {
				writeAPIError(w, r, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "摇卦失败"})
				return
			}
			toss = &t
			in.Backs = t.Backs
		}
	}

	cal, apiErr := s.buildCalendar(calendarQuery{
		at:        req.At,
		tz:        req.TZ,
		lon:       req.Lon,
		trueSolar: req.UseTrueSolarTime,
		boundary:  req.DayBoundary,
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	chart, err := liuyao.Cast(in, cal)
	if err != nil {
		writeAPIError(w, r, badRequest("排盘失败："+err.Error(), nil))
		return
	}

	chartID := newChartID()
	chart.ChartID = chartID
	createdAt := s.now().UTC()
	chart.CreatedAt = createdAt

	summary := ""
	if p, ok := chart.Payload.(liuyao.Payload); ok {
		summary = p.Summary()
	}

	rec := store.Record{
		ChartID:   chartID,
		Method:    chart.Method,
		Question:  in.Question,
		Summary:   summary,
		Chart:     chart,
		CreatedAt: createdAt,
		Entropy:   toss,
	}
	if err := s.store.Save(rec); err != nil {
		writeAPIError(w, r, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "保存盘面失败"})
		return
	}

	writeOK(w, r, newChartView(rec, false))
}

func (s *Server) handleGetChart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rec, ok := s.store.Get(id)
	if !ok {
		writeAPIError(w, r, &apiError{Status: http.StatusNotFound, Code: 10301, Message: "盘面不存在"})
		return
	}
	writeOK(w, r, newChartView(rec, true))
}

type chartSummary struct {
	ChartID   string        `json:"chart_id"`
	Method    schema.Method `json:"method"`
	Question  string        `json:"question,omitempty"`
	Summary   string        `json:"summary,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

type listChartsResponse struct {
	Items      []chartSummary `json:"items"`
	Total      int            `json:"total"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

func (s *Server) handleListCharts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	method := q.Get("method")
	if method != "" {
		if _, err := schema.ParseMethod(method); err != nil {
			writeAPIError(w, r, badRequest("不支持的术数", map[string]string{"method": method}))
			return
		}
	}

	limit := defaultLimit
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			writeAPIError(w, r, badRequest("limit 必须是正整数", map[string]string{"limit": raw}))
			return
		}
		if v > maxLimit {
			v = maxLimit
		}
		limit = v
	}

	offset := 0
	if raw := q.Get("cursor"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			writeAPIError(w, r, badRequest("cursor 非法", map[string]string{"cursor": raw}))
			return
		}
		offset = v
	}

	recs, total := s.store.List(store.ListOptions{Method: method, Limit: limit, Offset: offset})
	items := make([]chartSummary, 0, len(recs))
	for _, rec := range recs {
		items = append(items, chartSummary{
			ChartID:   rec.ChartID,
			Method:    rec.Method,
			Question:  rec.Question,
			Summary:   rec.Summary,
			CreatedAt: rec.CreatedAt,
		})
	}

	resp := listChartsResponse{Items: items, Total: total}
	if offset+len(items) < total {
		resp.NextCursor = strconv.Itoa(offset + len(items))
	}
	writeOK(w, r, resp)
}

func backsFromCoins(coins [6][3]int) ([6]int, *apiError) {
	var backs [6]int
	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			switch coins[i][j] {
			case 0:
			case 1:
				backs[i]++
			default:
				return backs, badRequest(
					"coins 每个元素只能是 0（字）或 1（背）",
					map[string]string{"coins": strconv.Itoa(i + 1)},
				)
			}
		}
	}
	return backs, nil
}

func newChartView(rec store.Record, fullFacts bool) chartView {
	facts := rec.Chart.Facts
	if !fullFacts {
		facts = coreFacts(facts)
	}
	return chartView{
		ChartID:       rec.ChartID,
		Method:        rec.Method,
		EngineVersion: rec.Chart.EngineVersion,
		Question:      rec.Question,
		Calendar:      rec.Chart.Calendar,
		School:        rec.Chart.School,
		Payload:       rec.Chart.Payload,
		Facts:         facts,
		Warnings:      rec.Chart.Warnings,
		Entropy:       rec.Entropy,
		CreatedAt:     rec.CreatedAt,
	}
}

func coreFacts(facts []schema.Fact) []schema.Fact {
	out := make([]schema.Fact, 0, len(facts))
	for _, f := range facts {
		if f.Weight == schema.WeightCore {
			out = append(out, f)
		}
	}
	return out
}
