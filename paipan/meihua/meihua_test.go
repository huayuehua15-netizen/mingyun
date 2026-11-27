package meihua

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

func TestTimeCast(t *testing.T) {
	// 2024-02-10 00:30：年支辰=5，农历正月初一，子时=1。
	// 上卦 5+1+1=7 艮，下卦 7+1=8 坤，动爻 8 mod 6=2。
	cal := buildCal(t, "2024-02-10 00:30")
	chart, err := Cast(Input{Question: "这次考试能过吗？", Mode: ModeTime}, cal)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.BenGua.Name != "山地剥" {
		t.Fatalf("本卦 = %s，想要山地剥", p.BenGua.Name)
	}
	if p.MovingLine != 2 {
		t.Fatalf("动爻 = %d，想要 2", p.MovingLine)
	}
	if p.TiGua.Name != "艮" || p.YongGua.Name != "坤" {
		t.Fatalf("体用 = %s/%s，想要艮/坤", p.TiGua.Name, p.YongGua.Name)
	}
	if p.Relation != "体用比和" {
		t.Fatalf("体用关系 = %s，想要体用比和", p.Relation)
	}
}

func TestNumberCast(t *testing.T) {
	// 3 离为上，4 震为下，动爻 7 mod 6=1；初爻动，下卦为用。
	cal := buildCal(t, "2024-02-10 12:00")
	chart, err := Cast(Input{Mode: ModeNumbers, Numbers: []int{3, 4}}, cal)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.BenGua.Name != "火雷噬嗑" {
		t.Fatalf("本卦 = %s", p.BenGua.Name)
	}
	if p.MovingLine != 1 {
		t.Fatalf("动爻 = %d", p.MovingLine)
	}
	if p.TiGua.Name != "离" || p.YongGua.Element != "木" {
		t.Fatalf("体用 = %s/%s", p.TiGua.Name, p.YongGua.Name)
	}
	if p.Relation != "用生体" {
		t.Fatalf("关系 = %s，想要用生体", p.Relation)
	}
	if p.BianGua.Name == "" || p.HuGua.Name == "" {
		t.Fatal("互卦/变卦为空")
	}
}

func TestNumbersValidation(t *testing.T) {
	cal := buildCal(t, "2024-02-10 12:00")
	if _, err := Cast(Input{Mode: ModeNumbers, Numbers: []int{3}}, cal); err == nil {
		t.Fatal("单个数应报错")
	}
	if _, err := Cast(Input{Mode: ModeNumbers, Numbers: []int{0, 4}}, cal); err == nil {
		t.Fatal("零应报错")
	}
}
