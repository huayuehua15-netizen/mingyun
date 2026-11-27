package meihua

// trigram 先天八卦与五行。
type trigramInfo struct {
	Name    string
	Element string
}

var trigrams = [8]trigramInfo{
	{Name: "乾", Element: "金"},
	{Name: "兑", Element: "金"},
	{Name: "离", Element: "火"},
	{Name: "震", Element: "木"},
	{Name: "巽", Element: "木"},
	{Name: "坎", Element: "水"},
	{Name: "艮", Element: "土"},
	{Name: "坤", Element: "土"},
}

// 先天八卦数：乾1 兑2 离3 震4 巽5 坎6 艮7 坤8。
func trigramFromNumber(n int) int {
	r := n % 8
	if r == 0 {
		r = 8
	}
	return r - 1
}

func trigramLines(index int) [3]bool {
	switch index {
	case 0: // 乾
		return [3]bool{true, true, true}
	case 1: // 兑
		return [3]bool{true, true, false}
	case 2: // 离
		return [3]bool{true, false, true}
	case 3: // 震
		return [3]bool{true, false, false}
	case 4: // 巽
		return [3]bool{false, true, true}
	case 5: // 坎
		return [3]bool{false, true, false}
	case 6: // 艮
		return [3]bool{false, false, true}
	default: // 坤
		return [3]bool{false, false, false}
	}
}

func trigramIndexFromLines(lines [3]bool) int {
	for i := 0; i < 8; i++ {
		if trigramLines(i) == lines {
			return i
		}
	}
	return 0
}

// hexagramNames 上卦为行、下卦为列。行序：乾兑离震巽坎艮坤。
var hexagramNames = [8][8]string{
	{"乾为天", "天泽履", "天火同人", "天雷无妄", "天风姤", "天水讼", "天山遁", "天地否"},
	{"泽天夬", "兑为泽", "泽火革", "泽雷随", "泽风大过", "泽水困", "泽山咸", "泽地萃"},
	{"火天大有", "火泽睽", "离为火", "火雷噬嗑", "火风鼎", "火水未济", "火山旅", "火地晋"},
	{"雷天大壮", "雷泽归妹", "雷火丰", "震为雷", "雷风恒", "雷水解", "雷山小过", "雷地豫"},
	{"风天小畜", "风泽中孚", "风火家人", "风雷益", "巽为风", "风水涣", "风山渐", "风地观"},
	{"水天需", "水泽节", "水火既济", "水雷屯", "水风井", "坎为水", "水山蹇", "水地比"},
	{"山天大畜", "山泽损", "山火贲", "山雷颐", "山风蛊", "山水蒙", "艮为山", "山地剥"},
	{"地天泰", "地泽临", "地火明夷", "地雷复", "地风升", "地水师", "地山谦", "坤为地"},
}

func hexagramIndexFromLines(lines [6]bool) (upper, lower int) {
	lower = trigramIndexFromLines([3]bool{lines[0], lines[1], lines[2]})
	upper = trigramIndexFromLines([3]bool{lines[3], lines[4], lines[5]})
	return upper, lower
}

func hexagramFromLines(lines [6]bool) Hexagram {
	upper, lower := hexagramIndexFromLines(lines)
	return Hexagram{
		Name:  hexagramNames[upper][lower],
		Upper: trigrams[upper].Name,
		Lower: trigrams[lower].Name,
		Lines: append([]bool(nil), lines[:]...),
	}
}
