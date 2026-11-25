package liuyao

import (
	"fmt"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

// LineKind 一爻的形态：少阳、少阴为静爻，老阳、老阴为动爻。
type LineKind int

const (
	YoungYang LineKind = iota // 少阳（单），静
	YoungYin                  // 少阴（拆），静
	OldYang                   // 老阳（重），动，变阴
	OldYin                    // 老阴（交），动，变阳
)

var (
	lineKindNames  = [4]string{"少阳", "少阴", "老阳", "老阴"}
	lineKindMarks  = [4]string{"—", "--", "○", "×"}
	lineKindYangs  = [4]bool{true, false, true, false}
	lineKindMoving = [4]bool{false, false, true, true}
)

// Valid 判断是否为合法爻形。
func (k LineKind) Valid() bool { return k >= 0 && k < 4 }

// String 返回中文名。
func (k LineKind) String() string {
	if !k.Valid() {
		return "?"
	}
	return lineKindNames[k]
}

// Mark 返回记号，动爻为 ○ 或 ×。
func (k LineKind) Mark() string {
	if !k.Valid() {
		return "?"
	}
	return lineKindMarks[k]
}

// Yang 是否为阳爻。
func (k LineKind) Yang() bool {
	if !k.Valid() {
		return false
	}
	return lineKindYangs[k]
}

// Moving 是否为动爻。
func (k LineKind) Moving() bool {
	if !k.Valid() {
		return false
	}
	return lineKindMoving[k]
}

// Changed 返回变后的爻形：老阳变少阴、老阴变少阳，静爻不变。
func (k LineKind) Changed() LineKind {
	switch k {
	case OldYang:
		return YoungYin
	case OldYin:
		return YoungYang
	default:
		return k
	}
}

// LineKindFromCoins 由三枚铜钱的背数（0—3）定爻。
// 三背为老阳、二背为少阴、一背为少阳、无背为老阴。
func LineKindFromCoins(backs int) (LineKind, error) {
	switch backs {
	case 3:
		return OldYang, nil
	case 2:
		return YoungYin, nil
	case 1:
		return YoungYang, nil
	case 0:
		return OldYin, nil
	}
	return 0, fmt.Errorf("铜钱背数必须为 0—3，实为 %d", backs)
}

// ParseLineKind 解析爻形，支持中文名与英文名。
func ParseLineKind(s string) (LineKind, error) {
	switch s {
	case "少阳", "young_yang", "yang":
		return YoungYang, nil
	case "少阴", "young_yin", "yin":
		return YoungYin, nil
	case "老阳", "old_yang", "moving_yang":
		return OldYang, nil
	case "老阴", "old_yin", "moving_yin":
		return OldYin, nil
	}
	return 0, fmt.Errorf("无法识别的爻形: %q", s)
}

// LiuQin 六亲。
type LiuQin int

const (
	LiuQinFuMu    LiuQin = iota // 父母
	LiuQinXiongDi               // 兄弟
	LiuQinZiSun                 // 子孙
	LiuQinQiCai                 // 妻财
	LiuQinGuanGui               // 官鬼
)

var liuQinNames = [5]string{"父母", "兄弟", "子孙", "妻财", "官鬼"}

// String 返回六亲名。
func (l LiuQin) String() string {
	if l < 0 || l >= 5 {
		return "?"
	}
	return liuQinNames[l]
}

// Valid 判断是否为合法六亲。
func (l LiuQin) Valid() bool { return l >= 0 && l < 5 }

// LiuQinFor 以 me（卦宫五行）为我，求 other（爻五行）的六亲。
//
//	同我者兄弟、生我者父母、我生者子孙、克我者官鬼、我克者妻财。
func LiuQinFor(me, other calendar.WuXing) (LiuQin, error) {
	if !me.Valid() || !other.Valid() {
		return 0, fmt.Errorf("五行非法: %v %v", me, other)
	}
	switch {
	case other == me:
		return LiuQinXiongDi, nil
	case other.Generates(me):
		return LiuQinFuMu, nil
	case me.Generates(other):
		return LiuQinZiSun, nil
	case other.Overcomes(me):
		return LiuQinGuanGui, nil
	case me.Overcomes(other):
		return LiuQinQiCai, nil
	}
	return 0, fmt.Errorf("无法判定六亲: %v 对 %v", other, me)
}

// LiuShen 六神（六兽）。
type LiuShen int

const (
	LiuShenQingLong LiuShen = iota // 青龙
	LiuShenZhuQue                  // 朱雀
	LiuShenGouChen                 // 勾陈
	LiuShenTengShe                 // 螣蛇
	LiuShenBaiHu                   // 白虎
	LiuShenXuanWu                  // 玄武
)

var liuShenNames = [6]string{"青龙", "朱雀", "勾陈", "螣蛇", "白虎", "玄武"}

// String 返回六神名。
func (l LiuShen) String() string {
	if l < 0 || l >= 6 {
		return "?"
	}
	return liuShenNames[l]
}

// liuShenStart 按日干定初爻所起六神：甲乙青龙、丙丁朱雀、戊勾陈、己螣蛇、庚辛白虎、壬癸玄武。
var liuShenStart = [10]LiuShen{
	LiuShenQingLong, LiuShenQingLong,
	LiuShenZhuQue, LiuShenZhuQue,
	LiuShenGouChen,
	LiuShenTengShe,
	LiuShenBaiHu, LiuShenBaiHu,
	LiuShenXuanWu, LiuShenXuanWu,
}

// LiuShenAt 返回日干与爻位（1—6）对应的六神，自初爻向上顺排。
func LiuShenAt(dayGan calendar.Gan, position int) (LiuShen, error) {
	if !dayGan.Valid() {
		return 0, fmt.Errorf("非法日干: %d", dayGan)
	}
	if position < 1 || position > 6 {
		return 0, fmt.Errorf("爻位必须为 1—6，实为 %d", position)
	}
	start := liuShenStart[dayGan]
	return LiuShen((int(start) + position - 1) % 6), nil
}
