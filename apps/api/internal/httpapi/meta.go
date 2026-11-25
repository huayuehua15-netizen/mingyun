package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

const (
	defaultTZ    = "Asia/Shanghai"
	defaultLon   = 120.0
	minYear      = 1900
	maxYear      = 2100
	defaultLimit = 20
	maxLimit     = 50
)

// calendarQuery 历法参数，HTTP 与排盘共用。
type calendarQuery struct {
	at        string
	tz        string
	lon       *float64
	trueSolar *bool
	boundary  string
}

func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	writeOK(w, r, schema.Capabilities)
}

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lon, err := parseLon(q.Get("lon"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	trueSolar, err := parseBool(q.Get("use_true_solar_time"))
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	cq := calendarQuery{
		at:        q.Get("at"),
		tz:        q.Get("tz"),
		lon:       lon,
		trueSolar: trueSolar,
		boundary:  q.Get("day_boundary"),
	}
	cal, err := s.buildCalendar(cq)
	if err != nil {
		writeAPIError(w, r, err)
		return
	}
	writeOK(w, r, cal)
}

func (s *Server) buildCalendar(q calendarQuery) (calendar.Context, *apiError) {
	tz := q.tz
	if tz == "" {
		tz = defaultTZ
	}
	loc, loadErr := time.LoadLocation(tz)
	if loadErr != nil {
		return calendar.Context{}, badRequest("时区无法识别", map[string]string{"tz": tz})
	}

	var at time.Time
	if q.at == "" {
		at = s.now().In(loc)
	} else {
		parsed, parseErr := time.Parse(time.RFC3339, q.at)
		if parseErr != nil {
			// 兼容不带偏移的写法，按请求时区解释。
			parsed, parseErr = time.ParseInLocation("2006-01-02 15:04", q.at, loc)
			if parseErr != nil {
				return calendar.Context{}, badRequest(
					"时间格式应为 RFC3339，如 2026-11-25T14:30:00+08:00",
					map[string]string{"at": q.at},
				)
			}
		}
		at = parsed
	}

	if at.Year() < minYear || at.Year() > maxYear {
		return calendar.Context{}, &apiError{
			Status:  http.StatusBadRequest,
			Code:    10002,
			Message: "时间超出支持范围（1900—2100）",
			Data:    map[string]any{"year": at.Year()},
		}
	}

	lon := defaultLon
	if q.lon != nil {
		lon = *q.lon
	}
	if lon < -180 || lon > 180 {
		return calendar.Context{}, badRequest("经度应在 -180 到 180 之间", map[string]string{"lon": strconv.FormatFloat(lon, 'f', 2, 64)})
	}

	trueSolar := true
	if q.trueSolar != nil {
		trueSolar = *q.trueSolar
	}

	boundary := calendar.ZiChu
	if q.boundary != "" {
		parsed, parseErr := calendar.ParseDayBoundary(q.boundary)
		if parseErr != nil {
			return calendar.Context{}, badRequest("日界只能是 zi_chu 或 zi_zheng", map[string]string{"day_boundary": q.boundary})
		}
		boundary = parsed
	}

	cal, buildErr := calendar.Build(calendar.Request{
		At:               at,
		Longitude:        lon,
		HasLongitude:     true,
		UseTrueSolarTime: trueSolar,
		DayBoundary:      boundary,
	})
	if buildErr != nil {
		return calendar.Context{}, &apiError{
			Status:  http.StatusInternalServerError,
			Code:    20001,
			Message: "历法计算失败：" + buildErr.Error(),
		}
	}
	return cal, nil
}

func parseLon(raw string) (*float64, *apiError) {
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, badRequest("经度必须是数字", map[string]string{"lon": raw})
	}
	return &v, nil
}

func parseBool(raw string) (*bool, *apiError) {
	if raw == "" {
		return nil, nil
	}
	switch raw {
	case "1", "true", "True", "yes":
		v := true
		return &v, nil
	case "0", "false", "False", "no":
		v := false
		return &v, nil
	}
	return nil, badRequest("布尔参数只能是 true / false / 1 / 0", map[string]string{"value": raw})
}
