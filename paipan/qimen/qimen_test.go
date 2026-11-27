package qimen

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

func TestCastKnownChart(t *testing.T) {
	// 与公开排盘样例对照：2026-11-27 06:41，小雪下元，阴遁二局；
	// 旬首甲戌己，值符天蓬落坎一，值使休门落坤二，时空亡申酉，马星巳在巽四。
	cal := buildCal(t, "2026-11-27 06:41")
	chart, err := Cast(Input{Question: "这次比赛能否进入决赛？"}, cal)
	if err != nil {
		t.Fatalf("Cast: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.Dun != "阴遁" || p.Ju != 2 || p.Yuan != "下元" || p.SolarTerm != "小雪" {
		t.Fatalf("起局 = %s%d局 %s %s", p.Dun, p.Ju, p.SolarTerm, p.Yuan)
	}
	if p.XunShou != "甲戌己" || p.ZhiFu != "天蓬" || p.ZhiFuPalace != 1 {
		t.Fatalf("旬首/值符 = %s %s@%d", p.XunShou, p.ZhiFu, p.ZhiFuPalace)
	}
	if p.ZhiShi != "休门" || p.ZhiShiPalace != 2 {
		t.Fatalf("值使 = %s@%d", p.ZhiShi, p.ZhiShiPalace)
	}
	if p.Kong[0].String() != "申" || p.Kong[1].String() != "酉" {
		t.Fatalf("时空亡 = %s%s", p.Kong[0], p.Kong[1])
	}
	if p.Horse.String() != "巳" || p.HorsePalace != 4 {
		t.Fatalf("马星 = %s@%d", p.Horse, p.HorsePalace)
	}
	if len(p.Palaces) != 9 {
		t.Fatalf("九宫数量 = %d", len(p.Palaces))
	}
	if p.Palaces[4].Door != "" || p.Palaces[4].God != "" {
		t.Fatalf("中五宫不应有门神: door=%q god=%q", p.Palaces[4].Door, p.Palaces[4].God)
	}
	if p.Palaces[4].Star != "天禽" || p.Palaces[0].Star != "天蓬" {
		t.Fatalf("九星落宫异常: 中=%s 坎=%s", p.Palaces[4].Star, p.Palaces[0].Star)
	}
	if p.Palaces[1].Door != "休门" {
		t.Fatalf("值使应临坤二，实际坤二 door=%q", p.Palaces[1].Door)
	}
	if p.Palaces[0].God != "值符" {
		t.Fatalf("值符神应临坎一，实际 %q", p.Palaces[0].God)
	}
}

func TestAllStarsAndDoorsPresent(t *testing.T) {
	cal := buildCal(t, "2026-11-27 06:41")
	chart, err := Cast(Input{}, cal)
	if err != nil {
		t.Fatal(err)
	}
	p := chart.Payload.(Payload)
	stars := map[string]bool{}
	doors := map[string]bool{}
	for _, palace := range p.Palaces {
		if palace.Star != "" {
			stars[palace.Star] = true
		}
		if palace.Door != "" {
			doors[palace.Door] = true
		}
	}
	for _, want := range []string{"天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英"} {
		if !stars[want] {
			t.Errorf("缺星: %s", want)
		}
	}
	for _, want := range []string{"休门", "生门", "伤门", "杜门", "景门", "死门", "惊门", "开门"} {
		if !doors[want] {
			t.Errorf("缺门: %s", want)
		}
	}
}

func TestYuanBoundary(t *testing.T) {
	// 符头地支分组：子午卯酉上元、寅申巳亥中元、辰戌丑未下元。
	cases := []struct {
		day  calendar.GanZhi
		want int
	}{
		{calendar.MustGanZhi(calendar.GanJia, calendar.ZhiZi), 0},
		{calendar.MustGanZhi(calendar.GanJi, calendar.ZhiMao), 0},
		{calendar.MustGanZhi(calendar.GanJia, calendar.ZhiYin), 1},
		{calendar.MustGanZhi(calendar.GanJi, calendar.ZhiWei), 2},
	}
	for _, c := range cases {
		got, err := yuanOf(c.day)
		if err != nil || got != c.want {
			t.Errorf("yuanOf(%s) = %d, %v; want %d", c.day, got, err, c.want)
		}
	}
}
