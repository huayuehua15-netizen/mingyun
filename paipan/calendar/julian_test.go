package calendar

import (
	"testing"
	"time"
)

func TestJulianDayNumberKnownPoints(t *testing.T) {
	cases := []struct {
		y, m, d int
		want    int
	}{
		{2000, 1, 1, 2451545}, // J2000 纪元
		{1949, 10, 1, 2433191},
		{2024, 1, 1, 2460311},
		{1900, 1, 1, 2415021},
		{2100, 12, 31, 2488434},
	}
	for _, c := range cases {
		if got := JulianDayNumber(c.y, c.m, c.d); got != c.want {
			t.Errorf("JDN(%d-%02d-%02d) 应为 %d，实为 %d", c.y, c.m, c.d, c.want, got)
		}
	}
}

func TestJulianRoundTrip(t *testing.T) {
	start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2100, 12, 31, 0, 0, 0, 0, time.UTC)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 7) {
		y, m, day := d.Date()
		jdn := JulianDayNumber(y, int(m), day)
		ry, rm, rd := CivilDate(jdn)
		if ry != y || rm != int(m) || rd != day {
			t.Fatalf("儒略日往返失败: %s -> %d -> %d-%02d-%02d", d.Format("2006-01-02"), jdn, ry, rm, rd)
		}
	}
}

func TestCheckRange(t *testing.T) {
	if err := CheckRange(2026); err != nil {
		t.Errorf("2026 应在范围内: %v", err)
	}
	if err := CheckRange(1899); err == nil {
		t.Error("1899 应被拒绝")
	}
	if err := CheckRange(2101); err == nil {
		t.Error("2101 应被拒绝")
	}
}
