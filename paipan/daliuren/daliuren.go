// Package daliuren 实现大六壬时间起课。
//
// 本版按 docs/03 第 3 节的通行规则实现月将加时、天地盘、四课、三传主框架、
// 十二天将、六亲、旬空与常用神煞。涉害深浅、伏吟中末传等流派细节在 school
// 字段中记录；有争议的取法优先采用文档所列通行做法。
package daliuren

import (
	"fmt"
	"strings"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 大六壬引擎版本。
const Version = "daliuren@0.1.0"

// Input 起课输入。大六壬以时间起课，问题只参与后续解读。
type Input struct {
	Question string
	School   map[string]string
}

// TianPanPalace 天地盘一宫。
type TianPanPalace struct {
	Earth        calendar.Zhi `json:"earth"`
	Heaven       calendar.Zhi `json:"heaven"`
	HeavenWuXing string       `json:"heaven_wuxing"`
	LiuQin       string       `json:"liu_qin"`
	TianJiang    string       `json:"tian_jiang"`
	ShenSha      []string     `json:"shen_sha,omitempty"`
}

// SiKe 四课之一。
type SiKe struct {
	Index       int          `json:"index"`
	Name        string       `json:"name"`
	Lower       calendar.Zhi `json:"lower"`
	Upper       calendar.Zhi `json:"upper"`
	LowerWuXing string       `json:"lower_wuxing"`
	UpperWuXing string       `json:"upper_wuxing"`
	Relation    string       `json:"relation"`
}

// Chuan 三传之一。
type Chuan struct {
	Index     int          `json:"index"`
	Name      string       `json:"name"`
	Branch    calendar.Zhi `json:"branch"`
	Palace    calendar.Zhi `json:"palace"`
	TianJiang string       `json:"tian_jiang"`
	LiuQin    string       `json:"liu_qin"`
}

// ShenSha 神煞。
type ShenSha struct {
	Name  string       `json:"name"`
	Zhi   calendar.Zhi `json:"zhi"`
	Value string       `json:"value"`
}

// Payload 大六壬盘面。
type Payload struct {
	Question     string                `json:"question,omitempty"`
	MonthGeneral calendar.MonthGeneral `json:"month_general"`
	HourZhi      calendar.Zhi          `json:"hour_zhi"`
	DayGan       calendar.Gan          `json:"day_gan"`
	DayZhi       calendar.Zhi          `json:"day_zhi"`
	MonthZhi     calendar.Zhi          `json:"month_zhi"`
	XunKong      [2]calendar.Zhi       `json:"xun_kong"`
	TianPan      []TianPanPalace       `json:"tian_pan"`
	SiKe         []SiKe                `json:"si_ke"`
	SanChuan     []Chuan               `json:"san_chuan"`
	KeTi         string                `json:"ke_ti"`
	Gate         string                `json:"gate"`
	ShenSha      []ShenSha             `json:"shen_sha,omitempty"`
	School       map[string]string     `json:"school,omitempty"`
}

// Cast 起一张大六壬课。
func Cast(in Input, cal calendar.Context) (schema.Chart, error) {
	mg := cal.MonthGeneral
	if !mg.Zhi.Valid() {
		return schema.Chart{}, fmt.Errorf("月将无效")
	}
	if !cal.Hour.Zhi.Valid() || !cal.Day.Gan.Valid() || !cal.Day.Zhi.Valid() {
		return schema.Chart{}, fmt.Errorf("历法上下文缺少日干支或时支")
	}

	hour := cal.Hour.Zhi
	dayGan := cal.Day.Gan
	dayZhi := cal.Day.Zhi
	tian := buildTianPan(mg.Zhi, hour)

	sike := buildSiKe(tian, dayGan, dayZhi)
	school := map[string]string{
		"sanchuan":        "jiuzongmen",
		"shehai":          "count_ke",
		"fuyin_zhongmo":   "xing",
		"fanyin_chuchuan": "yima",
		"guiren":          "maoshen_daynight",
	}
	for k, v := range in.School {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			school[k] = v
		}
	}

	gate, keTi, first, middle, last, err := buildSanChuan(tian, sike, dayGan, dayZhi, hour, mg.Zhi)
	if err != nil {
		return schema.Chart{}, err
	}

	jiang := buildTianJiang(tian, dayGan, hour)
	xunKong := [2]calendar.Zhi{cal.XunKong[0], cal.XunKong[1]}
	shenSha := buildShenSha(dayZhi, cal.Month.Zhi)

	payload := Payload{
		Question:     in.Question,
		MonthGeneral: mg,
		HourZhi:      hour,
		DayGan:       dayGan,
		DayZhi:       dayZhi,
		MonthZhi:     cal.Month.Zhi,
		XunKong:      xunKong,
		SiKe:         sike,
		SanChuan:     buildChuan(tian, jiang, dayGan, first, middle, last, gate),
		KeTi:         keTi,
		Gate:         gate,
		ShenSha:      shenSha,
		School:       school,
	}

	for e := 0; e < 12; e++ {
		earth := calendar.Zhi(e)
		heaven := tian[e]
		payload.TianPan = append(payload.TianPan, TianPanPalace{
			Earth:        earth,
			Heaven:       heaven,
			HeavenWuXing: heaven.WuXing().String(),
			LiuQin:       liuQinFor(dayGan, heaven),
			TianJiang:    jiang[e],
			ShenSha:      marksFor(heaven, shenSha, xunKong),
		})
	}

	chart := schema.Chart{
		Method:        schema.MethodDaLiuRen,
		EngineVersion: Version,
		Calendar:      cal,
		Question:      in.Question,
		School:        school,
		Payload:       payload,
		Facts:         BuildFacts(payload),
		CreatedAt:     time.Now().UTC(),
	}
	if err := chart.Validate(); err != nil {
		return schema.Chart{}, err
	}
	return chart, nil
}

func buildTianPan(general, hour calendar.Zhi) [12]calendar.Zhi {
	var tian [12]calendar.Zhi
	for e := 0; e < 12; e++ {
		// 月将加占时：占时宫置月将，其余顺时针顺布。
		tian[e] = calendar.Zhi(calendar.Mod(int(general)+calendar.Mod(e-int(hour), 12), 12))
	}
	return tian
}

func earthOfHeaven(tian [12]calendar.Zhi, heaven calendar.Zhi) calendar.Zhi {
	for e, h := range tian {
		if h == heaven {
			return calendar.Zhi(e)
		}
	}
	return 0
}

func buildSiKe(tian [12]calendar.Zhi, dayGan calendar.Gan, dayZhi calendar.Zhi) []SiKe {
	firstLower := ganJiZhi(dayGan)
	firstUpper := tian[firstLower]
	secondLower := firstUpper
	secondUpper := tian[secondLower]
	thirdLower := dayZhi
	thirdUpper := tian[thirdLower]
	fourthLower := thirdUpper
	fourthUpper := tian[fourthLower]

	pairs := [][2]calendar.Zhi{
		{firstLower, firstUpper},
		{secondLower, secondUpper},
		{thirdLower, thirdUpper},
		{fourthLower, fourthUpper},
	}
	out := make([]SiKe, 0, 4)
	for i, pair := range pairs {
		out = append(out, SiKe{
			Index:       i + 1,
			Name:        fmt.Sprintf("第%d课", i+1),
			Lower:       pair[0],
			Upper:       pair[1],
			LowerWuXing: pair[0].WuXing().String(),
			UpperWuXing: pair[1].WuXing().String(),
			Relation:    relationOf(pair[0], pair[1]),
		})
	}
	return out
}

func relationOf(lower, upper calendar.Zhi) string {
	lw, uw := lower.WuXing(), upper.WuXing()
	switch {
	case uw == lw:
		return "同"
	case lw.Overcomes(uw):
		return "下贼上"
	case uw.Overcomes(lw):
		return "上克下"
	case lw.Generates(uw):
		return "下生上"
	case uw.Generates(lw):
		return "上生下"
	default:
		return "无克"
	}
}

type keCandidate struct {
	index int
	upper calendar.Zhi
	lower calendar.Zhi
	kind  string
}

func findKe(sike []SiKe) []keCandidate {
	var down, up []keCandidate
	for _, k := range sike {
		lw, uw := k.Lower.WuXing(), k.Upper.WuXing()
		if lw.Overcomes(uw) {
			down = append(down, keCandidate{index: k.Index, upper: k.Upper, lower: k.Lower, kind: "下贼上"})
		} else if uw.Overcomes(lw) {
			up = append(up, keCandidate{index: k.Index, upper: k.Upper, lower: k.Lower, kind: "上克下"})
		}
	}
	if len(down) > 0 {
		return down
	}
	return up
}

func buildSanChuan(tian [12]calendar.Zhi, sike []SiKe, dayGan calendar.Gan, dayZhi, hour, general calendar.Zhi) (string, string, calendar.Zhi, calendar.Zhi, calendar.Zhi, error) {
	ke := findKe(sike)
	if len(ke) > 0 {
		gate, name, first := selectKe(tian, ke, dayGan)
		middle := tian[first]
		last := tian[middle]
		return gate, name, first, middle, last, nil
	}

	// 伏吟、返吟是有克仍取贼克、无克才取专法的结构；在无克分支里先判。
	if hour == general {
		first := tian[ganJiZhi(dayGan)]
		if dayGan.YinYang() == calendar.Yin {
			first = tian[dayZhi]
		}
		middle := xing(first)
		last := xing(middle)
		name := "自任"
		if dayGan.YinYang() == calendar.Yin {
			name = "自信"
		}
		return "伏吟", name, first, middle, last, nil
	}
	if calendar.ZhiChong(hour, general) {
		first := yiMa(dayZhi)
		middle := tian[first]
		last := tian[middle]
		return "返吟", "无依", first, middle, last, nil
	}

	if gate, name, first, ok := yaoKe(sike, dayGan); ok {
		middle := tian[first]
		last := tian[middle]
		return gate, name, first, middle, last, nil
	}

	if ganJiZhi(dayGan) == dayZhi {
		first := tian[ganJiZhi(dayGan)]
		if dayGan.YinYang() == calendar.Yang {
			first = calendar.Zhi(calendar.Mod(int(first)+2, 12))
		} else {
			first = calendar.Zhi(calendar.Mod(int(first)-2, 12))
		}
		middle := tian[first]
		last := tian[middle]
		return "八专", "八专课", first, middle, last, nil
	}
	if uniqueSiKe(sike) < 4 {
		first := bieZheFirst(tian, sike, dayGan, dayZhi)
		middle := tian[first]
		last := tian[middle]
		return "别责", "别责课", first, middle, last, nil
	}

	// 昴星：阳日取地盘酉上神，阴日取天盘酉下神。
	var first calendar.Zhi
	name := "虎视"
	if dayGan.YinYang() == calendar.Yang {
		first = tian[int(calendar.ZhiYou)]
	} else {
		first = earthOfHeaven(tian, calendar.ZhiYou)
		name = "冬蛇掩目"
	}
	middle := tian[first]
	last := tian[middle]
	return "昴星", name, first, middle, last, nil
}

func selectKe(tian [12]calendar.Zhi, candidates []keCandidate, dayGan calendar.Gan) (string, string, calendar.Zhi) {
	if len(candidates) == 1 {
		if candidates[0].kind == "下贼上" {
			return "贼克", "重审", candidates[0].upper
		}
		return "贼克", "元首", candidates[0].upper
	}

	// 比用：取与日干阴阳相同者。
	var matched []keCandidate
	for _, c := range candidates {
		if c.upper.YinYang() == dayGan.YinYang() {
			matched = append(matched, c)
		}
	}
	if len(matched) == 1 {
		return "比用", "知一课", matched[0].upper
	}
	if len(matched) > 1 {
		candidates = matched
	}

	// 涉害：从所临地盘顺数回本家，计被地盘所克次数，取最深。
	best := candidates[0]
	bestScore := -1
	for _, c := range candidates {
		score := sheHaiScore(tian, c.upper)
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return "涉害", "涉害课", best.upper
}

func sheHaiScore(tian [12]calendar.Zhi, upper calendar.Zhi) int {
	start := int(earthOfHeaven(tian, upper))
	home := int(upper)
	score := 0
	for step := 0; step < 12; step++ {
		e := calendar.Zhi(calendar.Mod(start+step, 12))
		if e.WuXing().Overcomes(upper.WuXing()) {
			score++
		}
		if int(e) == home {
			break
		}
	}
	// 孟仲季只作为同分时的次序：孟浅、仲中、季深。
	switch home {
	case int(calendar.ZhiYin), int(calendar.ZhiShen), int(calendar.ZhiSi), int(calendar.ZhiHai):
		score += 0
	case int(calendar.ZhiZi), int(calendar.ZhiWu), int(calendar.ZhiMao), int(calendar.ZhiYou):
		score += 1
	default:
		score += 2
	}
	return score
}

func yaoKe(sike []SiKe, dayGan calendar.Gan) (string, string, calendar.Zhi, bool) {
	dayWu := dayGan.WuXing()
	// 蒿矢：四课上神克日干。
	for _, k := range sike {
		if k.Upper.WuXing().Overcomes(dayWu) {
			return "遥克", "蒿矢", k.Upper, true
		}
	}
	// 弹射：日干克四课上神。
	for _, k := range sike {
		if dayWu.Overcomes(k.Upper.WuXing()) {
			return "遥克", "弹射", k.Upper, true
		}
	}
	return "", "", 0, false
}

func uniqueSiKe(sike []SiKe) int {
	seen := map[string]bool{}
	for _, k := range sike {
		seen[k.Lower.String()+k.Upper.String()] = true
	}
	return len(seen)
}

func bieZheFirst(tian [12]calendar.Zhi, sike []SiKe, dayGan calendar.Gan, dayZhi calendar.Zhi) calendar.Zhi {
	if dayGan.YinYang() == calendar.Yang {
		heGan := calendar.Gan(calendar.Mod(int(dayGan)+5, 10))
		return tian[ganJiZhi(heGan)]
	}
	// 阴日取日支三合前神之上神。
	group := sanHeGroup(dayZhi)
	prev := group[calendar.Mod(sanHeIndex(dayZhi)-1, 3)]
	return tian[prev]
}

func sanHeGroup(z calendar.Zhi) [3]calendar.Zhi {
	switch int(z) % 4 {
	case 0: // 申子辰
		return [3]calendar.Zhi{calendar.ZhiShen, calendar.ZhiZi, calendar.ZhiChen}
	case 1: // 巳酉丑
		return [3]calendar.Zhi{calendar.ZhiSi, calendar.ZhiYou, calendar.ZhiChou}
	case 2: // 寅午戌
		return [3]calendar.Zhi{calendar.ZhiYin, calendar.ZhiWu, calendar.ZhiXu}
	default: // 亥卯未
		return [3]calendar.Zhi{calendar.ZhiHai, calendar.ZhiMao, calendar.ZhiWei}
	}
}

func sanHeIndex(z calendar.Zhi) int {
	group := sanHeGroup(z)
	for i, v := range group {
		if v == z {
			return i
		}
	}
	return 0
}

func buildTianJiang(tian [12]calendar.Zhi, dayGan calendar.Gan, hour calendar.Zhi) [12]string {
	gui := guiRen(dayGan, hour)
	guiEarth := earthOfHeaven(tian, gui)
	sequence := []string{"贵人", "螣蛇", "朱雀", "六合", "勾陈", "青龙", "天空", "白虎", "太常", "玄武", "太阴", "天后"}
	earth := int(guiEarth)
	// 贵人临地盘亥至辰顺行，临巳至戌逆行。
	// 亥(11)、子(0)、丑(1)、寅(2)、卯(3)、辰(4) 为顺。
	clockwise := earth == 11 || earth <= 4
	var out [12]string
	for i, name := range sequence {
		step := i
		if !clockwise {
			step = -i
		}
		e := calendar.Mod(earth+step, 12)
		out[e] = name
	}
	return out
}

func guiRen(dayGan calendar.Gan, hour calendar.Zhi) calendar.Zhi {
	// 卯至申为昼，酉至寅为夜；用真太阳时定出的时支。
	day := int(hour) >= int(calendar.ZhiMao) && int(hour) <= int(calendar.ZhiShen)
	switch dayGan {
	case calendar.GanJia, calendar.GanWu, calendar.GanGeng:
		if day {
			return calendar.ZhiChou
		}
		return calendar.ZhiWei
	case calendar.GanYi, calendar.GanJi:
		if day {
			return calendar.ZhiZi
		}
		return calendar.ZhiShen
	case calendar.GanBing, calendar.GanDing:
		if day {
			return calendar.ZhiHai
		}
		return calendar.ZhiYou
	case calendar.GanRen, calendar.GanGui:
		if day {
			return calendar.ZhiMao
		}
		return calendar.ZhiSi
	default: // 辛
		if day {
			return calendar.ZhiWu
		}
		return calendar.ZhiYin
	}
}

func buildChuan(tian [12]calendar.Zhi, jiang [12]string, dayGan calendar.Gan, first, middle, last calendar.Zhi, gate string) []Chuan {
	_ = gate
	names := []string{"初传", "中传", "末传"}
	out := make([]Chuan, 0, 3)
	for i, branch := range []calendar.Zhi{first, middle, last} {
		palace := earthOfHeaven(tian, branch)
		out = append(out, Chuan{
			Index:     i + 1,
			Name:      names[i],
			Branch:    branch,
			Palace:    palace,
			TianJiang: jiang[palace],
			LiuQin:    liuQinFor(dayGan, branch),
		})
	}
	return out
}

func liuQinFor(dayGan calendar.Gan, other calendar.Zhi) string {
	dayWu, otherWu := dayGan.WuXing(), other.WuXing()
	switch {
	case dayWu == otherWu:
		return "兄弟"
	case otherWu.Generates(dayWu):
		return "父母"
	case dayWu.Generates(otherWu):
		return "子孙"
	case otherWu.Overcomes(dayWu):
		return "官鬼"
	default:
		return "妻财"
	}
}

func ganJiZhi(gan calendar.Gan) calendar.Zhi {
	switch gan {
	case calendar.GanJia:
		return calendar.ZhiYin
	case calendar.GanYi:
		return calendar.ZhiChen
	case calendar.GanBing, calendar.GanWu:
		return calendar.ZhiSi
	case calendar.GanDing, calendar.GanJi:
		return calendar.ZhiWei
	case calendar.GanGeng:
		return calendar.ZhiShen
	case calendar.GanXin:
		return calendar.ZhiXu
	case calendar.GanRen:
		return calendar.ZhiHai
	default: // 癸
		return calendar.ZhiChou
	}
}

func xing(z calendar.Zhi) calendar.Zhi {
	switch z {
	case calendar.ZhiZi:
		return calendar.ZhiMao
	case calendar.ZhiMao:
		return calendar.ZhiZi
	case calendar.ZhiYin:
		return calendar.ZhiSi
	case calendar.ZhiSi:
		return calendar.ZhiShen
	case calendar.ZhiShen:
		return calendar.ZhiYin
	case calendar.ZhiChou:
		return calendar.ZhiXu
	case calendar.ZhiXu:
		return calendar.ZhiWei
	case calendar.ZhiWei:
		return calendar.ZhiChou
	default: // 辰午酉亥自刑
		return z
	}
}

func yiMa(dayZhi calendar.Zhi) calendar.Zhi {
	switch int(dayZhi) % 4 {
	case 0: // 申子辰
		return calendar.ZhiYin
	case 1: // 巳酉丑
		return calendar.ZhiHai
	case 2: // 寅午戌
		return calendar.ZhiShen
	default: // 亥卯未
		return calendar.ZhiSi
	}
}

func taoHua(dayZhi calendar.Zhi) calendar.Zhi {
	switch int(dayZhi) % 4 {
	case 0: // 申子辰
		return calendar.ZhiYou
	case 1: // 巳酉丑
		return calendar.ZhiWu
	case 2: // 寅午戌
		return calendar.ZhiMao
	default: // 亥卯未
		return calendar.ZhiZi
	}
}

func buildShenSha(dayZhi, monthZhi calendar.Zhi) []ShenSha {
	ma := yiMa(dayZhi)
	tao := taoHua(dayZhi)
	yuePo := calendar.Zhi(calendar.Mod(int(monthZhi)+6, 12))
	riChong := calendar.Zhi(calendar.Mod(int(dayZhi)+6, 12))
	return []ShenSha{
		{Name: "驿马", Zhi: ma, Value: "驿马在" + ma.String()},
		{Name: "桃花", Zhi: tao, Value: "桃花在" + tao.String()},
		{Name: "月破", Zhi: yuePo, Value: "月破在" + yuePo.String()},
		{Name: "日冲", Zhi: riChong, Value: "日冲在" + riChong.String()},
	}
}

func marksFor(heaven calendar.Zhi, shenSha []ShenSha, xunKong [2]calendar.Zhi) []string {
	var marks []string
	for _, s := range shenSha {
		if s.Zhi == heaven {
			marks = append(marks, s.Name)
		}
	}
	if heaven == xunKong[0] || heaven == xunKong[1] {
		marks = append(marks, "旬空")
	}
	return marks
}

// Summary 返回一行大六壬摘要。
func (p Payload) Summary() string {
	if len(p.SanChuan) == 0 {
		return p.KeTi
	}
	return fmt.Sprintf("%s · %s · 初传%s 中传%s 末传%s", p.MonthGeneral.String(), p.KeTi,
		p.SanChuan[0].Branch, p.SanChuan[1].Branch, p.SanChuan[2].Branch)
}
