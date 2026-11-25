package liuyao

import (
	"fmt"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

// Trigram 八经卦（三爻卦）。取值为 0—7，依次为乾兑离震巽坎艮坤。
type Trigram int

const (
	TrigramQian Trigram = iota // 乾 ☰
	TrigramDui                 // 兑 ☱
	TrigramLi                  // 离 ☲
	TrigramZhen                // 震 ☳
	TrigramXun                 // 巽 ☴
	TrigramKan                 // 坎 ☵
	TrigramGen                 // 艮 ☶
	TrigramKun                 // 坤 ☷

	trigramCount
)

var (
	trigramNames   = [8]string{"乾", "兑", "离", "震", "巽", "坎", "艮", "坤"}
	trigramSymbols = [8]string{"☰", "☱", "☲", "☳", "☴", "☵", "☶", "☷"}
	// trigramBits 以 bit0 为初爻、bit2 为三爻，1 表示阳。
	trigramBits = [8]uint8{7, 3, 5, 1, 6, 2, 4, 0}
)

// Valid 判断是否为合法经卦。
func (t Trigram) Valid() bool { return t >= 0 && t < trigramCount }

// String 返回卦名。
func (t Trigram) String() string {
	if !t.Valid() {
		return "?"
	}
	return trigramNames[t]
}

// Symbol 返回卦符。
func (t Trigram) Symbol() string {
	if !t.Valid() {
		return "?"
	}
	return trigramSymbols[t]
}

// Bits 返回三爻的二进制表示，bit0 为初爻。
func (t Trigram) Bits() uint8 {
	if !t.Valid() {
		return 0
	}
	return trigramBits[t]
}

// Yaos 返回三爻阴阳，依次为初、二、三爻，true 表示阳。
func (t Trigram) Yaos() [3]bool {
	b := t.Bits()
	return [3]bool{b&1 == 1, b&2 == 2, b&4 == 4}
}

// TrigramFromBits 由三爻二进制还原经卦。
func TrigramFromBits(b uint8) (Trigram, bool) {
	b &= 7
	for t := Trigram(0); t < trigramCount; t++ {
		if trigramBits[t] == b {
			return t, true
		}
	}
	return 0, false
}

// WuXing 返回经卦五行，即该宫卦的五行。
func (t Trigram) WuXing() calendar.WuXing {
	switch t {
	case TrigramQian, TrigramDui:
		return calendar.WuXingJin
	case TrigramLi:
		return calendar.WuXingHuo
	case TrigramZhen, TrigramXun:
		return calendar.WuXingMu
	case TrigramKan:
		return calendar.WuXingShui
	default: // 艮、坤
		return calendar.WuXingTu
	}
}

// NaJia 一个经卦的纳甲配置。
type NaJia struct {
	// GanInner 内卦所用天干；外卦见 GanOuter。
	GanInner calendar.Gan
	// GanOuter 外卦所用天干。乾纳甲壬、坤纳乙癸，其余内外同干。
	GanOuter calendar.Gan
	// Zhi 六支，前三位为内卦初、二、三爻，后三位为外卦四、五、六爻。
	Zhi [6]calendar.Zhi
}

var naJiaTable = [8]NaJia{
	TrigramQian: {GanInner: calendar.GanJia, GanOuter: calendar.GanRen, Zhi: [6]calendar.Zhi{calendar.ZhiZi, calendar.ZhiYin, calendar.ZhiChen, calendar.ZhiWu, calendar.ZhiShen, calendar.ZhiXu}},
	TrigramDui:  {GanInner: calendar.GanDing, GanOuter: calendar.GanDing, Zhi: [6]calendar.Zhi{calendar.ZhiSi, calendar.ZhiMao, calendar.ZhiChou, calendar.ZhiHai, calendar.ZhiYou, calendar.ZhiWei}},
	TrigramLi:   {GanInner: calendar.GanJi, GanOuter: calendar.GanJi, Zhi: [6]calendar.Zhi{calendar.ZhiMao, calendar.ZhiChou, calendar.ZhiHai, calendar.ZhiYou, calendar.ZhiWei, calendar.ZhiSi}},
	TrigramZhen: {GanInner: calendar.GanGeng, GanOuter: calendar.GanGeng, Zhi: [6]calendar.Zhi{calendar.ZhiZi, calendar.ZhiYin, calendar.ZhiChen, calendar.ZhiWu, calendar.ZhiShen, calendar.ZhiXu}},
	TrigramXun:  {GanInner: calendar.GanXin, GanOuter: calendar.GanXin, Zhi: [6]calendar.Zhi{calendar.ZhiChou, calendar.ZhiHai, calendar.ZhiYou, calendar.ZhiWei, calendar.ZhiSi, calendar.ZhiMao}},
	TrigramKan:  {GanInner: calendar.GanWu, GanOuter: calendar.GanWu, Zhi: [6]calendar.Zhi{calendar.ZhiYin, calendar.ZhiChen, calendar.ZhiWu, calendar.ZhiShen, calendar.ZhiXu, calendar.ZhiZi}},
	TrigramGen:  {GanInner: calendar.GanBing, GanOuter: calendar.GanBing, Zhi: [6]calendar.Zhi{calendar.ZhiChen, calendar.ZhiWu, calendar.ZhiShen, calendar.ZhiXu, calendar.ZhiZi, calendar.ZhiYin}},
	TrigramKun:  {GanInner: calendar.GanYi, GanOuter: calendar.GanGui, Zhi: [6]calendar.Zhi{calendar.ZhiWei, calendar.ZhiSi, calendar.ZhiMao, calendar.ZhiChou, calendar.ZhiHai, calendar.ZhiYou}},
}

// NaJiaOf 返回该经卦的纳甲配置。
func (t Trigram) NaJiaOf() (NaJia, error) {
	if !t.Valid() {
		return NaJia{}, fmt.Errorf("非法经卦: %d", t)
	}
	return naJiaTable[t], nil
}

// NaJiaAt 返回经卦在指定爻位（1—3）上的干支。
// outer 为 true 时表示该经卦处于外卦（四、五、六爻）。
func (t Trigram) NaJiaAt(yao int, outer bool) (calendar.Gan, calendar.Zhi, error) {
	if yao < 1 || yao > 3 {
		return 0, 0, fmt.Errorf("经卦爻位必须为 1—3，实为 %d", yao)
	}
	spec, err := t.NaJiaOf()
	if err != nil {
		return 0, 0, err
	}
	if outer {
		return spec.GanOuter, spec.Zhi[yao+2], nil
	}
	return spec.GanInner, spec.Zhi[yao-1], nil
}
