package calendar

import (
	"time"

	lunar "github.com/6tail/lunar-go/calendar"
)

// LunarParts 返回某时刻对应的农历年月日。month 为 1—12，闰月时 leap=true。
//
// 小六壬、梅花易数等以农历月日取数的术数需要数字而不是格式化字符串。
func LunarParts(t time.Time) (year, month, day int, leap bool) {
	l := lunar.NewLunarFromSolar(lunar.NewSolar(
		t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second(),
	))
	m := l.GetMonth()
	if m < 0 {
		return l.GetYear(), -m, l.GetDay(), true
	}
	return l.GetYear(), m, l.GetDay(), false
}
