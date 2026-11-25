package calendar

import (
	"strconv"
	"testing"
	"time"
)

func TestHourZhiBoundaries(t *testing.T) {
	cases := []struct {
		hour int
		want string
	}{
		{23, "子"},
		{0, "子"},
		{1, "丑"},
		{2, "丑"},
		{3, "寅"},
		{4, "寅"},
		{5, "卯"},
		{11, "午"},
		{13, "未"},
		{21, "亥"},
		{22, "亥"},
	}
	for _, c := range cases {
		if got := HourZhi(c.hour).String(); got != c.want {
			t.Errorf("%d 点应为%s时，实为%s", c.hour, c.want, got)
		}
	}
}

func TestHourGanZhiWuShuDun(t *testing.T) {
	// 甲日子时起甲子，顺行十二辰。
	day := MustGanZhi(GanJia, ZhiZi)
	want := []string{"甲子", "乙丑", "丙寅", "丁卯", "戊辰", "己巳", "庚午", "辛未", "壬申", "癸酉", "甲戌", "乙亥"}
	for i := 0; i < 12; i++ {
		hour := i * 2
		if i == 0 {
			hour = 0
		}
		if got := HourGanZhi(day, hour).String(); got != want[i] {
			t.Errorf("甲日 %s时 应为 %s，实为 %s", HourZhi(hour), want[i], got)
		}
	}

	// 五鼠遁其余四组起点。日柱需为合法干支（干支同阴阳）。
	starts := []struct {
		day  string
		want string
	}{
		{"乙丑", "丙子"},
		{"丙寅", "戊子"},
		{"丁卯", "庚子"},
		{"戊辰", "壬子"},
	}
	for _, s := range starts {
		d, err := ParseGanZhi(s.day)
		if err != nil {
			t.Fatal(err)
		}
		if got := HourGanZhi(d, 0).String(); got != s.want {
			t.Errorf("%s日子时应为 %s，实为 %s", s.day, s.want, got)
		}
	}
}

func TestHourGanZhiOfLateZi(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	late := time.Date(2024, 1, 1, 23, 30, 0, 0, loc)
	// 子初换日：日柱变为乙丑，乙日子时起丙子。
	if got := HourGanZhiOf(late, ZiChu).String(); got != "丙子" {
		t.Errorf("2024-01-01 23:30 时柱应为丙子，实为 %s", got)
	}
	// 子正换日：日柱仍为甲子，甲日子时起甲子。
	if got := HourGanZhiOf(late, ZiZheng).String(); got != "甲子" {
		t.Errorf("子正换日下 2024-01-01 23:30 时柱应为甲子，实为 %s", got)
	}
}

func TestHourZhiValidRange(t *testing.T) {
	for h := 0; h < 24; h++ {
		z := HourZhi(h)
		if !z.Valid() {
			t.Errorf("小时 %s 得到非法地支", strconv.Itoa(h))
		}
	}
}
