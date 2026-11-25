package calendar

import "time"

// HourZhi 返回小时（0—23）对应的时辰地支。23 点与 0 点均为子时。
func HourZhi(hour int) Zhi {
	return Zhi(Mod((hour+1)/2, 12))
}

// HourGanZhi 由日柱与小时求时柱，时干按五鼠遁：
// 甲己起甲子、乙庚起丙子、丙辛起戊子、丁壬起庚子、戊癸起壬子。
func HourGanZhi(day GanZhi, hour int) GanZhi {
	z := HourZhi(hour)
	g := Gan(Mod(int(day.Gan)%5*2+int(z), 10))
	// 时干与地支阴阳必然一致：上式中的偶数项保证了这一点。
	return GanZhi{Gan: g, Zhi: z}
}

// HourGanZhiOf 由时刻求时柱，日柱按日界规则确定。
// t 必须携带正确的本地时区。
func HourGanZhiOf(t time.Time, b DayBoundary) GanZhi {
	return HourGanZhi(DayGanZhiOf(t, b), t.Hour())
}
