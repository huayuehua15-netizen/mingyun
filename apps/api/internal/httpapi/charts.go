package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/daliuren"
	"github.com/huayuehua15-netizen/mingyun/paipan/liuyao"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
	"github.com/huayuehua15-netizen/mingyun/paipan/ziwei"
)

// castRequest 三术共用的排盘请求。六爻使用 mode/coins/lines；大六壬使用 at；
// 紫微使用 gender/calendar_type/birth_date/birth_time 等出生信息。
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
	CalendarType     string            `json:"calendar_type,omitempty"`
	BirthDate        string            `json:"birth_date,omitempty"`
	BirthTime        string            `json:"birth_time,omitempty"`
	IsLeapMonth      bool              `json:"is_leap_month,omitempty"`
	LeapMonthPolicy  string            `json:"leap_month_policy,omitempty"`
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

	var chart schema.Chart
	var toss *entropy.Toss
	var apiErr *apiError
	switch method {
	case schema.MethodLiuYao:
		chart, toss, apiErr = s.castLiuYaoChart(req)
	case schema.MethodDaLiuRen:
		chart, apiErr = s.castDaLiuRenChart(req)
	case schema.MethodZiWei:
		chart, apiErr = s.castZiWeiChart(req)
	default:
		apiErr = &apiError{
			Status:  http.StatusForbidden,
			Code:    10202,
			Message: "该术数尚未开放",
			Data:    map[string]any{"method": method},
		}
	}
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	chartID := newChartID()
	chart.ChartID = chartID
	createdAt := s.now().UTC()
	chart.CreatedAt = createdAt

	rec := store.Record{
		ChartID:   chartID,
		Method:    chart.Method,
		Question:  strings.TrimSpace(req.Question),
		Summary:   summaryText(chart.Payload),
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

func (s *Server) castLiuYaoChart(req castRequest) (schema.Chart, *entropy.Toss, *apiError) {
	mode := liuyao.Mode(req.Mode)
	if mode != liuyao.ModeCoin && mode != liuyao.ModeManual {
		return schema.Chart{}, nil, badRequest("mode 只能是 coin 或 manual", map[string]string{"mode": req.Mode})
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
			return schema.Chart{}, nil, badRequest("手动录入需提供 6 个 lines", map[string]string{"lines": "need 6"})
		}
		for _, s := range req.Lines {
			kind, err := liuyao.ParseLineKind(strings.TrimSpace(s))
			if err != nil {
				return schema.Chart{}, nil, badRequest("无法识别的爻形："+s, map[string]string{"lines": s})
			}
			in.Lines = append(in.Lines, kind)
		}
	case liuyao.ModeCoin:
		switch {
		case req.Coins != nil && req.Backs != nil:
			return schema.Chart{}, nil, badRequest("coins 与 backs 只能提供一个", nil)
		case req.Coins != nil:
			backs, apiErr := backsFromCoins(*req.Coins)
			if apiErr != nil {
				return schema.Chart{}, nil, apiErr
			}
			in.Backs = backs
		case req.Backs != nil:
			for i, v := range *req.Backs {
				if v < 0 || v > 3 {
					return schema.Chart{}, nil, badRequest("背数必须在 0—3 之间", map[string]string{"backs": strconv.Itoa(i + 1)})
				}
			}
			in.Backs = *req.Backs
		default:
			if s.entropy == nil {
				return schema.Chart{}, nil, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "随机源未配置"}
			}
			t, err := s.entropy.Toss()
			if err != nil {
				return schema.Chart{}, nil, &apiError{Status: http.StatusInternalServerError, Code: 20001, Message: "摇卦失败"}
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
		return schema.Chart{}, nil, apiErr
	}
	chart, err := liuyao.Cast(in, cal)
	if err != nil {
		return schema.Chart{}, nil, badRequest("排盘失败："+err.Error(), nil)
	}
	return chart, toss, nil
}

func (s *Server) castDaLiuRenChart(req castRequest) (schema.Chart, *apiError) {
	cal, apiErr := s.buildCalendar(calendarQuery{
		at:        req.At,
		tz:        req.TZ,
		lon:       req.Lon,
		trueSolar: req.UseTrueSolarTime,
		boundary:  req.DayBoundary,
	})
	if apiErr != nil {
		return schema.Chart{}, apiErr
	}
	chart, err := daliuren.Cast(daliuren.Input{
		Question: strings.TrimSpace(req.Question),
		School:   req.School,
	}, cal)
	if err != nil {
		return schema.Chart{}, badRequest("大六壬排盘失败："+err.Error(), nil)
	}
	return chart, nil
}

func (s *Server) castZiWeiChart(req castRequest) (schema.Chart, *apiError) {
	tz := req.TZ
	if tz == "" {
		tz = defaultTZ
	}
	loc, loadErr := time.LoadLocation(tz)
	if loadErr != nil {
		return schema.Chart{}, badRequest("时区无法识别", map[string]string{"tz": tz})
	}

	lon := defaultLon
	if req.Lon != nil {
		lon = *req.Lon
	}
	if lon < -180 || lon > 180 {
		return schema.Chart{}, badRequest("经度应在 -180 到 180 之间", map[string]string{"lon": strconv.FormatFloat(lon, 'f', 2, 64)})
	}
	trueSolar := true
	if req.UseTrueSolarTime != nil {
		trueSolar = *req.UseTrueSolarTime
	}

	birth, err := ziwei.ParseBirth(ziwei.Input{
		Gender:          strings.TrimSpace(req.Gender),
		CalendarType:    strings.TrimSpace(req.CalendarType),
		BirthDate:       strings.TrimSpace(req.BirthDate),
		BirthTime:       strings.TrimSpace(req.BirthTime),
		IsLeapMonth:     req.IsLeapMonth,
		LeapMonthPolicy: strings.TrimSpace(req.LeapMonthPolicy),
	}, ziwei.BirthOptions{
		Location:         loc,
		Longitude:        lon,
		HasLongitude:     true,
		UseTrueSolarTime: trueSolar,
	})
	if err != nil {
		return schema.Chart{}, badRequest("出生信息无法排盘："+err.Error(), nil)
	}

	// birth.SolarTime 已经按真太阳时/晚子时归一化，这里不再二次校正。
	falseVal := false
	cal, apiErr := s.buildCalendar(calendarQuery{
		at:        birth.SolarTime.Format(time.RFC3339),
		tz:        tz,
		trueSolar: &falseVal,
		boundary:  req.DayBoundary,
	})
	if apiErr != nil {
		return schema.Chart{}, apiErr
	}
	if birth.TrueSolarUsed {
		cal.TrueSolarUsed = true
		cal.TrueSolarTime = birth.SolarTime
		cal.Warnings = append(cal.Warnings, fmt.Sprintf("紫微生盘已按真太阳时校正（经度 %.2f°）", lon))
	}

	chart, castErr := ziwei.Cast(ziwei.Input{
		Question:        strings.TrimSpace(req.Question),
		Gender:          strings.TrimSpace(req.Gender),
		CalendarType:    strings.TrimSpace(req.CalendarType),
		BirthDate:       strings.TrimSpace(req.BirthDate),
		BirthTime:       strings.TrimSpace(req.BirthTime),
		IsLeapMonth:     req.IsLeapMonth,
		LeapMonthPolicy: strings.TrimSpace(req.LeapMonthPolicy),
		School:          req.School,
	}, birth, cal)
	if castErr != nil {
		return schema.Chart{}, badRequest("紫微排盘失败："+castErr.Error(), nil)
	}
	return chart, nil
}

func summaryText(payload any) string {
	switch p := payload.(type) {
	case liuyao.Payload:
		return p.Summary()
	case daliuren.Payload:
		return p.Summary()
	case ziwei.Payload:
		return p.Summary()
	default:
		return ""
	}
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
