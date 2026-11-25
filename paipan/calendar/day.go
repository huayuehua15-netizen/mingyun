package calendar

import (
	"fmt"
	"time"
)

// dayAnchorJDN 为日柱锚点：1949-10-01 是甲子日（儒略日 2433191）。
// 该锚点与 2000-01-01 戊午日、2024-01-01 甲子日相互印证，测试中三点同时校验。
const dayAnchorJDN = 2433191

// DayBoundary 日界规则。
type DayBoundary int

const (
	// ZiChu 子初换日：23:00 起算次日。传统术数默认。
	ZiChu DayBoundary = iota
	// ZiZheng 子正换日：00:00 起算次日。
	ZiZheng
)

// String 返回配置名，用于盘面记录。
func (b DayBoundary) String() string {
	if b == ZiZheng {
		return "zi_zheng"
	}
	return "zi_chu"
}

// ParseDayBoundary 解析配置名。
func ParseDayBoundary(s string) (DayBoundary, error) {
	switch s {
	case "", "zi_chu":
		return ZiChu, nil
	case "zi_zheng":
		return ZiZheng, nil
	}
	return 0, fmt.Errorf("非法日界规则: %q", s)
}

// DayGanZhi 由公历日期求日柱。
func DayGanZhi(y, m, d int) GanZhi {
	return dayGanZhiFromJDN(JulianDayNumber(y, m, d))
}

// DayGanZhiOf 由时刻求日柱，按日界规则处理晚子时。
// t 必须携带正确的本地时区，否则 hour 判定会错。
func DayGanZhiOf(t time.Time, b DayBoundary) GanZhi {
	jdn := JulianDayNumberOf(t)
	if b == ZiChu && t.Hour() >= 23 {
		jdn++
	}
	return dayGanZhiFromJDN(jdn)
}

func dayGanZhiFromJDN(jdn int) GanZhi {
	return GanZhiFromIndex(jdn - dayAnchorJDN)
}

// Xun 返回日柱所属旬的旬首，如 "甲子"。
func Xun(gz GanZhi) GanZhi {
	idx := gz.Index()
	if idx < 0 {
		return GanZhi{}
	}
	return GanZhiFromIndex(idx - idx%10)
}

// XunKong 返回日柱所属旬的空亡二支，如甲子旬返回 (戌, 亥)。
func XunKong(gz GanZhi) (Zhi, Zhi) {
	idx := gz.Index()
	if idx < 0 {
		return ZhiZi, ZhiZi
	}
	head := idx - idx%10
	headZhi := head % 12
	return Zhi((headZhi + 10) % 12), Zhi((headZhi + 11) % 12)
}
