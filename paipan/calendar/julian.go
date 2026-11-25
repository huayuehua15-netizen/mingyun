package calendar

import (
	"fmt"
	"time"
)

const (
	// MinYear 支持的最早公历年份。
	MinYear = 1900
	// MaxYear 支持的最晚公历年份。
	MaxYear = 2100
)

// JulianDayNumber 返回给定公历日期的儒略日数（整日号）。
// 采用格里高利历，适用于 1582 年 10 月 15 日之后的日期；本项目限定 1900—2100。
func JulianDayNumber(y, m, d int) int {
	a := (14 - m) / 12
	yy := y + 4800 - a
	mm := m + 12*a - 3
	return d + (153*mm+2)/5 + 365*yy + yy/4 - yy/100 + yy/400 - 32045
}

// CivilDate 由儒略日数还原公历年月日。
func CivilDate(jdn int) (y, m, d int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - 146097*b/4
	dd := (4*c + 3) / 1461
	e := c - 1461*dd/4
	mm := (5*e + 2) / 153
	d = e - (153*mm+2)/5 + 1
	m = mm + 3 - 12*(mm/10)
	y = 100*b + dd - 4800 + mm/10
	return y, m, d
}

// JulianDayNumberOf 返回 time.Time 所在日期（按其自身时区）的儒略日数。
func JulianDayNumberOf(t time.Time) int {
	y, m, d := t.Date()
	return JulianDayNumber(y, int(m), d)
}

// CheckRange 校验年份是否在本包支持范围内。
func CheckRange(y int) error {
	if y < MinYear || y > MaxYear {
		return fmt.Errorf("年份 %d 超出支持范围 %d—%d", y, MinYear, MaxYear)
	}
	return nil
}
