package calendar

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

// Gan 天干。取值为 0—9，依次为甲乙丙丁戊己庚辛壬癸。
type Gan int

const (
	GanJia  Gan = iota // 甲
	GanYi              // 乙
	GanBing            // 丙
	GanDing            // 丁
	GanWu              // 戊
	GanJi              // 己
	GanGeng            // 庚
	GanXin             // 辛
	GanRen             // 壬
	GanGui             // 癸

	ganCount
)

var ganNames = [10]string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// String 返回天干的中文名。非法值返回 "?"。
func (g Gan) String() string {
	if !g.Valid() {
		return "?"
	}
	return ganNames[g]
}

// Valid 判断是否为合法天干。
func (g Gan) Valid() bool { return g >= 0 && g < ganCount }

// WuXing 返回天干五行。
func (g Gan) WuXing() WuXing {
	switch g {
	case GanJia, GanYi:
		return WuXingMu
	case GanBing, GanDing:
		return WuXingHuo
	case GanWu, GanJi:
		return WuXingTu
	case GanGeng, GanXin:
		return WuXingJin
	default:
		return WuXingShui
	}
}

// YinYang 返回天干阴阳。偶数为阳。
func (g Gan) YinYang() YinYang {
	if g%2 == 0 {
		return Yang
	}
	return Yin
}

// Zhi 地支。取值为 0—11，依次为子丑寅卯辰巳午未申酉戌亥。
type Zhi int

const (
	ZhiZi   Zhi = iota // 子
	ZhiChou            // 丑
	ZhiYin             // 寅
	ZhiMao             // 卯
	ZhiChen            // 辰
	ZhiSi              // 巳
	ZhiWu              // 午
	ZhiWei             // 未
	ZhiShen            // 申
	ZhiYou             // 酉
	ZhiXu              // 戌
	ZhiHai             // 亥

	zhiCount
)

var zhiNames = [12]string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// String 返回地支的中文名。非法值返回 "?"。
func (z Zhi) String() string {
	if !z.Valid() {
		return "?"
	}
	return zhiNames[z]
}

// Valid 判断是否为合法地支。
func (z Zhi) Valid() bool { return z >= 0 && z < zhiCount }

// WuXing 返回地支五行。
func (z Zhi) WuXing() WuXing {
	switch z {
	case ZhiYin, ZhiMao:
		return WuXingMu
	case ZhiSi, ZhiWu:
		return WuXingHuo
	case ZhiShen, ZhiYou:
		return WuXingJin
	case ZhiHai, ZhiZi:
		return WuXingShui
	default: // 辰戌丑未
		return WuXingTu
	}
}

// YinYang 返回地支阴阳。偶数为阳。
func (z Zhi) YinYang() YinYang {
	if z%2 == 0 {
		return Yang
	}
	return Yin
}

// MarshalJSON 让地支以中文名序列化，而不是数字。
func (z Zhi) MarshalJSON() ([]byte, error) { return json.Marshal(z.String()) }

// UnmarshalJSON 支持从中文名解析地支。
func (z *Zhi) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseZhi(s)
	if err != nil {
		return err
	}
	*z = parsed
	return nil
}

// WuXing 五行。取值为 0—4，依次为木火土金水。
type WuXing int

const (
	WuXingMu   WuXing = iota // 木
	WuXingHuo                // 火
	WuXingTu                 // 土
	WuXingJin                // 金
	WuXingShui               // 水

	wuXingCount
)

var wuXingNames = [5]string{"木", "火", "土", "金", "水"}

// String 返回五行中文名。
func (w WuXing) String() string {
	if w < 0 || w >= wuXingCount {
		return "?"
	}
	return wuXingNames[w]
}

// Valid 判断是否为合法五行。
func (w WuXing) Valid() bool { return w >= 0 && w < wuXingCount }

// Generates 判断 w 是否生 target（木生火、火生土、土生金、金生水、水生木）。
func (w WuXing) Generates(target WuXing) bool {
	if !w.Valid() || !target.Valid() {
		return false
	}
	return (target-w+wuXingCount)%wuXingCount == 1
}

// Overcomes 判断 w 是否克 target（木克土、土克水、水克火、火克金、金克木）。
func (w WuXing) Overcomes(target WuXing) bool {
	if !w.Valid() || !target.Valid() {
		return false
	}
	return (target-w+wuXingCount)%wuXingCount == 2
}

// YinYang 阴阳。
type YinYang int

const (
	Yang YinYang = iota // 阳
	Yin                 // 阴
)

// String 返回阴阳中文名。
func (y YinYang) String() string {
	if y == Yin {
		return "阴"
	}
	return "阳"
}

// GanZhi 干支组合。干支必须同阴阳，否则为非法组合。
type GanZhi struct {
	Gan Gan
	Zhi Zhi
}

// NewGanZhi 构造干支，校验合法性与阴阳匹配。
func NewGanZhi(g Gan, z Zhi) (GanZhi, error) {
	if !g.Valid() {
		return GanZhi{}, fmt.Errorf("非法天干: %d", g)
	}
	if !z.Valid() {
		return GanZhi{}, fmt.Errorf("非法地支: %d", z)
	}
	if int(g)%2 != int(z)%2 {
		return GanZhi{}, fmt.Errorf("干支阴阳不匹配: %s%s", g, z)
	}
	return GanZhi{Gan: g, Zhi: z}, nil
}

// MustGanZhi 与 NewGanZhi 相同，但非法组合时 panic。
// 仅用于包内常量与已校验数据，不要在接收外部输入的地方使用。
func MustGanZhi(g Gan, z Zhi) GanZhi {
	gz, err := NewGanZhi(g, z)
	if err != nil {
		panic(err)
	}
	return gz
}

// GanZhiFromIndex 由六十甲子序号（0—59）还原干支，序号自动取模。
func GanZhiFromIndex(i int) GanZhi {
	i = Mod(i, 60)
	return GanZhi{Gan: Gan(i % 10), Zhi: Zhi(i % 12)}
}

// Index 返回六十甲子序号（0—59）。非法组合返回 -1。
func (gz GanZhi) Index() int {
	if !gz.Valid() {
		return -1
	}
	for i := int(gz.Gan); i < 60; i += 10 {
		if i%12 == int(gz.Zhi) {
			return i
		}
	}
	return -1
}

// Valid 判断干支是否为合法组合。
func (gz GanZhi) Valid() bool {
	return gz.Gan.Valid() && gz.Zhi.Valid() && int(gz.Gan)%2 == int(gz.Zhi)%2
}

// String 返回干支中文名，如 "甲子"。
func (gz GanZhi) String() string { return gz.Gan.String() + gz.Zhi.String() }

// GanWuXing 返回天干五行。
func (gz GanZhi) GanWuXing() WuXing { return gz.Gan.WuXing() }

// ZhiWuXing 返回地支五行。
func (gz GanZhi) ZhiWuXing() WuXing { return gz.Zhi.WuXing() }

// MarshalJSON 让干支以 "甲子" 形式序列化。
func (gz GanZhi) MarshalJSON() ([]byte, error) { return json.Marshal(gz.String()) }

// UnmarshalJSON 支持从 "甲子" 形式解析。
func (gz *GanZhi) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseGanZhi(s)
	if err != nil {
		return err
	}
	*gz = parsed
	return nil
}

// ParseGanZhi 解析两字干支，如 "甲子"。
func ParseGanZhi(s string) (GanZhi, error) {
	if utf8.RuneCountInString(s) != 2 {
		return GanZhi{}, fmt.Errorf("干支必须为两个字: %q", s)
	}
	runes := []rune(s)
	g, err := ParseGan(string(runes[0]))
	if err != nil {
		return GanZhi{}, err
	}
	z, err := ParseZhi(string(runes[1]))
	if err != nil {
		return GanZhi{}, err
	}
	return NewGanZhi(g, z)
}

// ParseGan 解析单个天干。
func ParseGan(s string) (Gan, error) {
	for i, name := range ganNames {
		if name == s {
			return Gan(i), nil
		}
	}
	return 0, fmt.Errorf("非法天干: %q", s)
}

// ParseZhi 解析单个地支。
func ParseZhi(s string) (Zhi, error) {
	for i, name := range zhiNames {
		if name == s {
			return Zhi(i), nil
		}
	}
	return 0, fmt.Errorf("非法地支: %q", s)
}

// ZhiChong 判断两地支是否相冲（子午、丑未、寅申、卯酉、辰戌、巳亥）。
func ZhiChong(a, b Zhi) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	return (int(a)-int(b)+12)%12 == 6
}

// ZhiLiuHe 判断两地支是否六合（子丑、寅亥、卯戌、辰酉、巳申、午未）。
func ZhiLiuHe(a, b Zhi) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	return (int(a)+int(b))%12 == 1
}

// SanHeWuXing 返回地支所属三合局的五行。
// 申子辰合水、巳酉丑合金、寅午戌合火、亥卯未合木。
func SanHeWuXing(z Zhi) (WuXing, bool) {
	if !z.Valid() {
		return 0, false
	}
	switch int(z) % 4 {
	case 0:
		return WuXingShui, true
	case 1:
		return WuXingJin, true
	case 2:
		return WuXingHuo, true
	default:
		return WuXingMu, true
	}
}

// Mod 返回非负余数。
func Mod(n, m int) int {
	if m <= 0 {
		return 0
	}
	r := n % m
	if r < 0 {
		r += m
	}
	return r
}
