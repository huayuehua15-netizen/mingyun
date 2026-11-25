package calendar

import (
	"math"
	"time"
)

// LongitudeOffset 返回经度相对时区中央经线的时间修正。
// 每偏离 1 度修正 4 分钟；东经为正。
//
// 例：东八区中央经线 120°E，广州 113.26°E，修正约 -27 分钟。
func LongitudeOffset(longitude, tzOffsetHours float64) time.Duration {
	central := 15 * tzOffsetHours
	minutes := (longitude - central) * 4
	return time.Duration(minutes * float64(time.Minute))
}

// EquationOfTime 返回均时差（真太阳时减平太阳时），采用 NOAA 近似公式，
// 全年误差约 ±0.5 分钟，足够时辰判定使用。
//
// t 应携带当地时间，公式中的年积日与小时按 t 自身的时区取值。
func EquationOfTime(t time.Time) time.Duration {
	n := float64(t.YearDay())
	gamma := 2 * math.Pi / 365 * (n - 1 + (float64(t.Hour())-12)/24)
	minutes := 229.18 * (0.000075 +
		0.001868*math.Cos(gamma) -
		0.032077*math.Sin(gamma) -
		0.014615*math.Cos(2*gamma) -
		0.040849*math.Sin(2*gamma))
	return time.Duration(minutes * float64(time.Minute))
}

// TrueSolarTime 由标准时求真太阳时。
//
//	t 为当地时间（必须已设定正确时区）
//	longitude 为东经度数
//	tzOffsetHours 为该时刻的时区偏移小时数（东八区为 8）
//
// 计算顺序为 标准时 + 经度修正 + 均时差。
func TrueSolarTime(t time.Time, longitude, tzOffsetHours float64) time.Time {
	res := t.Add(LongitudeOffset(longitude, tzOffsetHours)).Add(EquationOfTime(t))
	// 均时差为浮点运算，结果截断到秒，避免 JSON 出现纳秒尾数。
	return time.Date(res.Year(), res.Month(), res.Day(),
		res.Hour(), res.Minute(), res.Second(), 0, res.Location())
}
