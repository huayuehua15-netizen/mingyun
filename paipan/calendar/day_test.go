package calendar

import (
	"testing"
	"time"
)

func TestDayGanZhiAnchors(t *testing.T) {
	cases := []struct {
		y, m, d int
		want    string
	}{
		{1949, 10, 1, "甲子"},
		{2000, 1, 1, "戊午"},
		{2024, 1, 1, "甲子"},
	}
	for _, c := range cases {
		if got := DayGanZhi(c.y, c.m, c.d).String(); got != c.want {
			t.Errorf("%d-%02d-%02d 日柱应为 %s，实为 %s", c.y, c.m, c.d, c.want, got)
		}
	}
}

func TestDayGanZhiSixtyCycle(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	seen := map[int]bool{}
	for i := 0; i < 60; i++ {
		d := start.AddDate(0, 0, i)
		y, m, day := d.Date()
		gz := DayGanZhi(y, int(m), day)
		idx := gz.Index()
		if idx < 0 {
			t.Fatalf("第 %d 天得到非法干支 %s", i, gz)
		}
		if seen[idx] {
			t.Fatalf("六十日内干支重复: %s", gz)
		}
		seen[idx] = true
	}
	if len(seen) != 60 {
		t.Fatalf("六十日应覆盖全部 60 个干支，实为 %d", len(seen))
	}
}

func TestDayBoundary(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	// 2024-01-01 为甲子日。23:30 在子初换日规则下应算作次日乙丑。
	late := time.Date(2024, 1, 1, 23, 30, 0, 0, loc)
	if got := DayGanZhiOf(late, ZiChu).String(); got != "乙丑" {
		t.Errorf("子初换日下 2024-01-01 23:30 应为乙丑，实为 %s", got)
	}
	if got := DayGanZhiOf(late, ZiZheng).String(); got != "甲子" {
		t.Errorf("子正换日下 2024-01-01 23:30 应为甲子，实为 %s", got)
	}

	// 00:30 两种规则都应算当日。
	early := time.Date(2024, 1, 1, 0, 30, 0, 0, loc)
	if got := DayGanZhiOf(early, ZiChu).String(); got != "甲子" {
		t.Errorf("2024-01-01 00:30 应为甲子，实为 %s", got)
	}
}

func TestXunKong(t *testing.T) {
	cases := []struct {
		xun  string
		kong string
	}{
		{"甲子", "戌亥"},
		{"甲戌", "申酉"},
		{"甲申", "午未"},
		{"甲午", "辰巳"},
		{"甲辰", "寅卯"},
		{"甲寅", "子丑"},
	}
	for _, c := range cases {
		gz, err := ParseGanZhi(c.xun)
		if err != nil {
			t.Fatal(err)
		}
		a, b := XunKong(gz)
		if a.String()+b.String() != c.kong {
			t.Errorf("%s 旬空应为 %s，实为 %s%s", c.xun, c.kong, a, b)
		}
	}
}

func TestXun(t *testing.T) {
	gz, err := ParseGanZhi("乙酉")
	if err != nil {
		t.Fatal(err)
	}
	if got := Xun(gz).String(); got != "甲申" {
		t.Errorf("乙酉所属旬首应为甲申，实为 %s", got)
	}
}
