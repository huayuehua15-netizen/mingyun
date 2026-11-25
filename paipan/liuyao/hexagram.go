package liuyao

import "fmt"

// Hexagram 六十四卦。Lines 自下而上，true 为阳爻。
type Hexagram struct {
	Name   string
	Upper  Trigram
	Lower  Trigram
	Lines  [6]bool
	Palace Trigram
	Rank   int // 0 本宫、1—5 一世至五世、6 游魂、7 归魂
}

var rankNames = [8]string{"本宫", "一世", "二世", "三世", "四世", "五世", "游魂", "归魂"}

// 六十四卦名，按 [上卦][下卦] 排列，两卦顺序均为乾兑离震巽坎艮坤。
var hexagramNames = [8][8]string{
	// 上乾
	{"乾为天", "天泽履", "天火同人", "天雷无妄", "天风姤", "天水讼", "天山遁", "天地否"},
	// 上兑
	{"泽天夬", "兑为泽", "泽火革", "泽雷随", "泽风大过", "泽水困", "泽山咸", "泽地萃"},
	// 上离
	{"火天大有", "火泽睽", "离为火", "火雷噬嗑", "火风鼎", "火水未济", "火山旅", "火地晋"},
	// 上震
	{"雷天大壮", "雷泽归妹", "雷火丰", "震为雷", "雷风恒", "雷水解", "雷山小过", "雷地豫"},
	// 上巽
	{"风天小畜", "风泽中孚", "风火家人", "风雷益", "巽为风", "风水涣", "风山渐", "风地观"},
	// 上坎
	{"水天需", "水泽节", "水火既济", "水雷屯", "水风井", "坎为水", "水山蹇", "水地比"},
	// 上艮
	{"山天大畜", "山泽损", "山火贲", "山雷颐", "山风蛊", "山水蒙", "艮为山", "山地剥"},
	// 上坤
	{"地天泰", "地泽临", "地火明夷", "地雷复", "地风升", "地水师", "地山谦", "坤为地"},
}

// 八宫卦序，按 [卦宫][0 本宫、1—5 一世至五世、6 游魂、7 归魂] 排列。
var palaceHexagrams = [8][8]string{
	// 乾宫
	{"乾为天", "天风姤", "天山遁", "天地否", "风地观", "山地剥", "火地晋", "火天大有"},
	// 兑宫
	{"兑为泽", "泽水困", "泽地萃", "泽山咸", "水山蹇", "地山谦", "雷山小过", "雷泽归妹"},
	// 离宫
	{"离为火", "火山旅", "火风鼎", "火水未济", "山水蒙", "风水涣", "天水讼", "天火同人"},
	// 震宫
	{"震为雷", "雷地豫", "雷水解", "雷风恒", "地风升", "水风井", "泽风大过", "泽雷随"},
	// 巽宫
	{"巽为风", "风天小畜", "风火家人", "风雷益", "天雷无妄", "火雷噬嗑", "山雷颐", "山风蛊"},
	// 坎宫
	{"坎为水", "水泽节", "水雷屯", "水火既济", "泽火革", "雷火丰", "地火明夷", "地水师"},
	// 艮宫
	{"艮为山", "山火贲", "山天大畜", "山泽损", "火泽睽", "天泽履", "风泽中孚", "风山渐"},
	// 坤宫
	{"坤为地", "地雷复", "地泽临", "地天泰", "雷天大壮", "泽天夬", "水天需", "水地比"},
}

// 世爻位置，按卦序（本宫、一世至五世、游魂、归魂）。
var shiYaoByRank = [8]int{6, 1, 2, 3, 4, 5, 4, 3}

type hexagramIndex struct {
	byBits map[uint8]Hexagram
	byName map[string]Hexagram
	all    []Hexagram
}

var hexagrams = buildHexagrams()

func buildHexagrams() *hexagramIndex {
	idx := &hexagramIndex{
		byBits: make(map[uint8]Hexagram, 64),
		byName: make(map[string]Hexagram, 64),
	}

	for u := Trigram(0); u < trigramCount; u++ {
		for l := Trigram(0); l < trigramCount; l++ {
			name := hexagramNames[u][l]
			bits := l.Bits() | u.Bits()<<3

			palace, rank := Trigram(0), -1
		search:
			for p := Trigram(0); p < trigramCount; p++ {
				for r := 0; r < 8; r++ {
					if palaceHexagrams[p][r] == name {
						palace, rank = p, r
						break search
					}
				}
			}
			if rank < 0 {
				panic(fmt.Sprintf("六十四卦表不一致：%s 未归宫", name))
			}
			if _, dup := idx.byName[name]; dup {
				panic(fmt.Sprintf("六十四卦名重复：%s", name))
			}

			h := Hexagram{
				Name:   name,
				Upper:  u,
				Lower:  l,
				Lines:  linesFromBits(bits),
				Palace: palace,
				Rank:   rank,
			}
			idx.byBits[bits] = h
			idx.byName[name] = h
			idx.all = append(idx.all, h)
		}
	}
	if len(idx.all) != 64 {
		panic(fmt.Sprintf("六十四卦数量错误：%d", len(idx.all)))
	}
	return idx
}

func linesFromBits(b uint8) [6]bool {
	return [6]bool{b&1 == 1, b&2 == 2, b&4 == 4, b&8 == 8, b&16 == 16, b&32 == 32}
}

func bitsFromLines(l [6]bool) uint8 {
	var b uint8
	for i, yang := range l {
		if yang {
			b |= 1 << uint(i)
		}
	}
	return b
}

// Bits 返回六爻二进制，bit0 为初爻。
func (h Hexagram) Bits() uint8 { return bitsFromLines(h.Lines) }

// FromBits 由六爻二进制取卦。
func FromBits(b uint8) (Hexagram, bool) {
	h, ok := hexagrams.byBits[b&0x3f]
	return h, ok
}

// FromLines 由六爻阴阳取卦。
func FromLines(lines [6]bool) (Hexagram, bool) {
	return FromBits(bitsFromLines(lines))
}

// FromName 由卦名取卦。
func FromName(name string) (Hexagram, bool) {
	h, ok := hexagrams.byName[name]
	return h, ok
}

// AllHexagrams 返回全部六十四卦，按二进制序排列。
func AllHexagrams() []Hexagram { return append([]Hexagram(nil), hexagrams.all...) }

// PalaceLabel 返回 "乾宫" 形式。
func (h Hexagram) PalaceLabel() string { return h.Palace.String() + "宫" }

// RankLabel 返回 "本宫"、"一世" 等。
func (h Hexagram) RankLabel() string {
	if h.Rank < 0 || h.Rank > 7 {
		return "?"
	}
	return rankNames[h.Rank]
}

// String 返回 "乾为天（乾宫·本宫）"。
func (h Hexagram) String() string {
	return fmt.Sprintf("%s（%s·%s）", h.Name, h.PalaceLabel(), h.RankLabel())
}

// ShiYao 返回世爻位置（1—6）。
func (h Hexagram) ShiYao() int {
	if h.Rank < 0 || h.Rank > 7 {
		return 0
	}
	return shiYaoByRank[h.Rank]
}

// YingYao 返回应爻位置：与世爻相隔两位。
func (h Hexagram) YingYao() int {
	shi := h.ShiYao()
	if shi == 0 {
		return 0
	}
	if shi > 3 {
		return shi - 3
	}
	return shi + 3
}

// Changed 返回动爻变化后的卦。moving 为 true 的爻翻转阴阳。
func (h Hexagram) Changed(moving [6]bool) (Hexagram, bool) {
	lines := h.Lines
	for i := range lines {
		if moving[i] {
			lines[i] = !lines[i]
		}
	}
	return FromLines(lines)
}
