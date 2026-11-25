package calendar

import (
	"fmt"
	"time"
)

// Request 历法查询的输入。
type Request struct {
	// At 用户输入的本地时刻，必须携带正确时区。
	At time.Time
	// Longitude 东经度数。HasLongitude 为 false 时忽略。
	Longitude float64
	// HasLongitude 是否提供了经度。
	HasLongitude bool
	// UseTrueSolarTime 是否启用真太阳时校正。
	UseTrueSolarTime bool
	// DayBoundary 日界规则，默认子初。
	DayBoundary DayBoundary
}

// Context 一次历法计算的全部结果，是三术共用的上下文。
type Context struct {
	SolarTime     time.Time    `json:"solar_time"`
	TrueSolarTime time.Time    `json:"true_solar_time"`
	LunarDate     string       `json:"lunar_date"`
	Year          GanZhi       `json:"year_gz"`
	Month         GanZhi       `json:"month_gz"`
	Day           GanZhi       `json:"day_gz"`
	Hour          GanZhi       `json:"hour_gz"`
	XunKong       [2]Zhi       `json:"xun_kong"`
	PrevTerm      Term         `json:"prev_term"`
	NextTerm      Term         `json:"next_term"`
	MonthGeneral  MonthGeneral `json:"month_general"`
	DayBoundary   string       `json:"day_boundary"`
	TrueSolarUsed bool         `json:"true_solar_used"`
	Warnings      []string     `json:"warnings,omitempty"`
}

// Build 计算历法上下文。
//
// 年柱与月柱按交节时刻判定（节气为全球同一瞬间），日柱与时柱按请求的
// 本地时刻判定，因此二者使用的基准不同，这是正确的：
// 立春在各地是同一刻，而子时在各地是各自的午夜。
func Build(req Request) (Context, error) {
	if req.At.IsZero() {
		return Context{}, fmt.Errorf("时间不能为空")
	}
	if err := CheckRange(req.At.Year()); err != nil {
		return Context{}, err
	}

	tst := req.At
	used := false
	if req.UseTrueSolarTime && req.HasLongitude {
		_, offset := req.At.Zone()
		tst = TrueSolarTime(req.At, req.Longitude, float64(offset)/3600)
		used = true
	}

	day := DayGanZhiOf(tst, req.DayBoundary)
	hour := HourGanZhi(day, tst.Hour())

	year, err := YearPillar(req.At)
	if err != nil {
		return Context{}, fmt.Errorf("求年柱: %w", err)
	}
	month, err := MonthPillar(req.At)
	if err != nil {
		return Context{}, fmt.Errorf("求月柱: %w", err)
	}
	prev, err := PrevTerm(req.At)
	if err != nil {
		return Context{}, fmt.Errorf("求上一节气: %w", err)
	}
	next, err := NextTerm(req.At)
	if err != nil {
		return Context{}, fmt.Errorf("求下一节气: %w", err)
	}
	mg, err := MonthGeneralAt(req.At)
	if err != nil {
		return Context{}, fmt.Errorf("求月将: %w", err)
	}

	kongA, kongB := XunKong(day)

	l := solarFor(req.At).GetLunar()
	lunarDate := fmt.Sprintf("%s年 %s月%s", l.GetYearInGanZhi(), l.GetMonthInChinese(), l.GetDayInChinese())

	var warnings []string
	if req.DayBoundary == ZiChu && tst.Hour() >= 23 {
		warnings = append(warnings, "已过子初，日柱按次日计算")
	}
	if used {
		warnings = append(warnings, fmt.Sprintf("已按真太阳时校正（经度 %.2f°）", req.Longitude))
	}

	return Context{
		SolarTime:     req.At,
		TrueSolarTime: tst,
		LunarDate:     lunarDate,
		Year:          year,
		Month:         month,
		Day:           day,
		Hour:          hour,
		XunKong:       [2]Zhi{kongA, kongB},
		PrevTerm:      prev,
		NextTerm:      next,
		MonthGeneral:  mg,
		DayBoundary:   req.DayBoundary.String(),
		TrueSolarUsed: used,
		Warnings:      warnings,
	}, nil
}
