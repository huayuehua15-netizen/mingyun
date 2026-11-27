package xiaoliuren

import (
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

func buildCal(t *testing.T, at string) calendar.Context {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04", at, loc)
	if err != nil {
		t.Fatal(err)
	}
	cal, err := calendar.Build(calendar.Request{At: parsed, DayBoundary: calendar.ZiChu})
	if err != nil {
		t.Fatal(err)
	}
	return cal
}

func TestCastLunarNewYear(t *testing.T) {
	// 2024-02-10 为农历正月初一。月=1、日=1、子时=0，
	// 最终 (1+1+0-3) mod 6 = 0，落大安。
	cal := buildCal(t, "2024-02-10 00:30")
	chart, err := Cast(Input{Question: "这次考试能否通过？"}, cal)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.LunarMonth != 1 || p.LunarDay != 1 {
		t.Fatalf("农历日期错误: %d月%d日", p.LunarMonth, p.LunarDay)
	}
	if p.FinalPalace.Name != "大安" {
		t.Fatalf("落宫 = %s，想要大安", p.FinalPalace.Name)
	}
}

func TestCastHourMovesPalace(t *testing.T) {
	// 同一天巳时（时支序数 6），(1+1+6-3) mod 6 = 5，落空亡。
	cal := buildCal(t, "2024-02-10 10:30")
	chart, err := Cast(Input{}, cal)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.FinalPalace.Name != "空亡" {
		t.Fatalf("落宫 = %s，想要空亡", p.FinalPalace.Name)
	}
	if len(chart.Facts) < 6 {
		t.Fatalf("facts 过少: %d", len(chart.Facts))
	}
}

func TestThreePalacesDiffer(t *testing.T) {
	cal := buildCal(t, "2024-02-10 10:30")
	chart, err := Cast(Input{}, cal)
	if err != nil {
		t.Fatal(err)
	}
	p := chart.Payload.(Payload)
	if p.MonthPalace.Name != "大安" || p.DayPalace.Name != "大安" {
		t.Fatalf("月日宫应从大安起: 月%s 日%s", p.MonthPalace.Name, p.DayPalace.Name)
	}
	if len(p.SixPalaces) != 6 {
		t.Fatalf("六宫数量 = %d", len(p.SixPalaces))
	}
}
