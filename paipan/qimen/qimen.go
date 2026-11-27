// Package qimen 实现时家转盘奇门的基础排盘。
//
// 本版覆盖：节气定阴阳遁与局数、符头定上中下元、地盘三奇六仪、
// 旬首值符值使、天盘九星、八门、八神、时空亡与马星。采用拆补法，
// 不含超神接气置闰、十干克应、格局与用神体系；这些边界记录在 school。
package qimen

import (
	"fmt"
	"strings"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 奇门遁甲引擎版本。
const Version = "qimen@0.1.0"

// Input 起局输入。奇门以时刻起局，问题只参与后续解读。
type Input struct {
	Question string
	School   map[string]string
}

// Palace 九宫之一。
type Palace struct {
	Index        int          `json:"index"`
	Name         string       `json:"name"`
	Gua          string       `json:"gua"`
	Direction    string       `json:"direction"`
	EarthGan     calendar.Gan `json:"earth_gan"`
	HeavenGan    calendar.Gan `json:"heaven_gan"`
	EarthWuXing  string       `json:"earth_wu_xing"`
	HeavenWuXing string       `json:"heaven_wu_xing"`
	Star         string       `json:"star"`
	Door         string       `json:"door"`
	God          string       `json:"god"`
	MenPo        bool         `json:"men_po,omitempty"`
	Kong         bool         `json:"kong,omitempty"`
	Horse        bool         `json:"horse,omitempty"`
	IsZhiFu      bool         `json:"is_zhi_fu,omitempty"`
	IsZhiShi     bool         `json:"is_zhi_shi,omitempty"`
}

// Payload 奇门盘面。
type Payload struct {
	Question     string            `json:"question,omitempty"`
	Dun          string            `json:"dun"`
	Ju           int               `json:"ju"`
	Yuan         string            `json:"yuan"`
	SolarTerm    string            `json:"solar_term"`
	TermAt       time.Time         `json:"term_at"`
	YearGZ       string            `json:"year_gz"`
	MonthGZ      string            `json:"month_gz"`
	DayGZ        string            `json:"day_gz"`
	HourGZ       string            `json:"hour_gz"`
	XunShou      string            `json:"xun_shou"`
	Instrument   string            `json:"instrument"`
	ZhiFu        string            `json:"zhi_fu"`
	ZhiFuPalace  int               `json:"zhi_fu_palace"`
	ZhiShi       string            `json:"zhi_shi"`
	ZhiShiPalace int               `json:"zhi_shi_palace"`
	Kong         [2]calendar.Zhi   `json:"kong"`
	Horse        calendar.Zhi      `json:"horse"`
	HorsePalace  int               `json:"horse_palace"`
	Palaces      []Palace          `json:"palaces"`
	School       map[string]string `json:"school,omitempty"`
}

// Summary 一行摘要。
func (p Payload) Summary() string {
	return fmt.Sprintf("奇门遁甲 · %s%d局 · %s值符 · %s值使", p.Dun, p.Ju, p.ZhiFu, p.ZhiShi)
}

type termJu struct {
	Yang bool
	Ju   [3]int
}

var termTable = map[string]termJu{
	"冬至": {Yang: true, Ju: [3]int{1, 7, 4}},
	"小寒": {Yang: true, Ju: [3]int{2, 8, 5}},
	"大寒": {Yang: true, Ju: [3]int{3, 9, 6}},
	"立春": {Yang: true, Ju: [3]int{8, 5, 2}},
	"雨水": {Yang: true, Ju: [3]int{9, 6, 3}},
	"惊蛰": {Yang: true, Ju: [3]int{1, 7, 4}},
	"春分": {Yang: true, Ju: [3]int{3, 9, 6}},
	"清明": {Yang: true, Ju: [3]int{4, 1, 7}},
	"谷雨": {Yang: true, Ju: [3]int{5, 2, 8}},
	"立夏": {Yang: true, Ju: [3]int{4, 1, 7}},
	"小满": {Yang: true, Ju: [3]int{5, 2, 8}},
	"芒种": {Yang: true, Ju: [3]int{6, 3, 9}},

	"夏至": {Yang: false, Ju: [3]int{9, 3, 6}},
	"小暑": {Yang: false, Ju: [3]int{8, 2, 5}},
	"大暑": {Yang: false, Ju: [3]int{7, 1, 4}},
	"立秋": {Yang: false, Ju: [3]int{2, 5, 8}},
	"处暑": {Yang: false, Ju: [3]int{1, 4, 7}},
	"白露": {Yang: false, Ju: [3]int{9, 3, 6}},
	"秋分": {Yang: false, Ju: [3]int{7, 1, 4}},
	"寒露": {Yang: false, Ju: [3]int{6, 9, 3}},
	"霜降": {Yang: false, Ju: [3]int{5, 8, 2}},
	"立冬": {Yang: false, Ju: [3]int{6, 9, 3}},
	"小雪": {Yang: false, Ju: [3]int{5, 8, 2}},
	"大雪": {Yang: false, Ju: [3]int{4, 7, 1}},
}

var palaceNames = [10]string{"", "坎一宫", "坤二宫", "震三宫", "巽四宫", "中五宫", "乾六宫", "兑七宫", "艮八宫", "离九宫"}
var palaceGua = [10]string{"", "坎", "坤", "震", "巽", "中", "乾", "兑", "艮", "离"}
var palaceDirection = [10]string{"", "北", "西南", "东", "东南", "中", "西北", "西", "东北", "南"}
var palaceElement = [10]string{"", "水", "土", "木", "木", "土", "金", "金", "土", "火"}
var starHome = [10]string{"", "天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英"}
var doorHome = [10]string{"", "休门", "死门", "伤门", "杜门", "", "开门", "惊门", "生门", "景门"}
var ringOrder = []int{1, 8, 3, 4, 9, 2, 7, 6}
var doorOrder = []string{"休门", "生门", "伤门", "杜门", "景门", "死门", "惊门", "开门"}
var godOrder = []string{"值符", "腾蛇", "太阴", "六合", "白虎", "玄武", "九地", "九天"}

var instrumentByXun = []calendar.Gan{
	calendar.GanWu,   // 甲子戊
	calendar.GanJi,   // 甲戌己
	calendar.GanGeng, // 甲申庚
	calendar.GanXin,  // 甲午辛
	calendar.GanRen,  // 甲辰壬
	calendar.GanGui,  // 甲寅癸
}

// Cast 起一张时家奇门局。
func Cast(in Input, cal calendar.Context) (schema.Chart, error) {
	if !cal.Day.Valid() || !cal.Hour.Valid() || cal.PrevTerm.At.IsZero() {
		return schema.Chart{}, fmt.Errorf("历法上下文缺少年月日时或节气")
	}
	entry, ok := termTable[cal.PrevTerm.Name]
	if !ok {
		return schema.Chart{}, fmt.Errorf("未识别的节气: %s", cal.PrevTerm.Name)
	}
	fuTou := fuTouOfDay(cal.Day)
	yuan, err := yuanOf(fuTou)
	if err != nil {
		return schema.Chart{}, err
	}
	ju := entry.Ju[yuan]
	dun := "阳遁"
	if !entry.Yang {
		dun = "阴遁"
	}

	earth := buildEarth(ju, entry.Yang)
	xun := calendar.Xun(cal.Hour)
	xunIdx := xun.Index() / 10
	if xunIdx < 0 || xunIdx >= len(instrumentByXun) {
		return schema.Chart{}, fmt.Errorf("无法确定时旬")
	}
	instrument := instrumentByXun[xunIdx]
	instrumentPalace := findStem(earth, instrument)
	if instrumentPalace == 0 {
		return schema.Chart{}, fmt.Errorf("地盘找不到旬首六仪 %s", instrument)
	}

	zhiFuHome := instrumentPalace
	zhiFu := starHome[zhiFuHome]
	targetStem := cal.Hour.Gan
	if targetStem == calendar.GanJia {
		targetStem = instrument
	}
	targetPalace := findStem(earth, targetStem)
	if targetPalace == 0 {
		return schema.Chart{}, fmt.Errorf("地盘找不到时干 %s", targetStem)
	}

	stars, heaven := buildHeaven(earth, zhiFuHome, targetPalace)
	zhiShiHome := zhiFuHome
	if zhiShiHome == 5 {
		zhiShiHome = 2
	}
	zhiShi := doorHome[zhiFuHome]
	if zhiShi == "" {
		zhiShi = "死门"
	}
	steps := calendar.Mod(int(cal.Hour.Zhi)-int(xun.Zhi), 12)
	sign := 1
	if !entry.Yang {
		sign = -1
	}
	zhiShiPalace := moveNumeric(zhiShiHome, steps, sign)
	if zhiShiPalace == 5 {
		zhiShiPalace = 2
	}

	doors := buildDoors(zhiShi, zhiShiPalace)
	gods := buildGods(targetPalace, entry.Yang)

	hourKongA, hourKongB := calendar.XunKong(cal.Hour)
	horse := horseOf(cal.Hour.Zhi)
	horsePalace := branchPalace(horse)

	palaces := make([]Palace, 0, 9)
	for p := 1; p <= 9; p++ {
		door := doors[p]
		menPo := false
		if door != "" && doorElement(door) != "" && overcomes(doorElement(door), palaceElement[p]) {
			menPo = true
		}
		// 中五宫寄坤，八门八神不落中宫；地盘/天盘干与天禽仍显示。
		palaces = append(palaces, Palace{
			Index:        p,
			Name:         palaceNames[p],
			Gua:          palaceGua[p],
			Direction:    palaceDirection[p],
			EarthGan:     earth[p],
			HeavenGan:    heaven[p],
			EarthWuXing:  earth[p].WuXing().String(),
			HeavenWuXing: heaven[p].WuXing().String(),
			Star:         stars[p],
			Door:         door,
			God:          gods[p],
			MenPo:        menPo,
			Kong:         branchPalace(hourKongA) == p || branchPalace(hourKongB) == p,
			Horse:        horsePalace == p,
			IsZhiFu:      p == targetPalace,
			IsZhiShi:     p == zhiShiPalace,
		})
	}

	school := map[string]string{
		"pan":   "zhuanpan",
		"ju":    "chaibu_futou",
		"yuan":  "futou_branch",
		"stars": "clockwise",
		"doors": "numeric_target_clockwise_fill",
		"gods":  "yang_fwd_yin_rev",
		"qin":   "ji_kun",
		"geju":  "not_included",
	}
	for k, v := range in.School {
		if k != "" && v != "" {
			school[k] = v
		}
	}

	payload := Payload{
		Question:     in.Question,
		Dun:          dun,
		Ju:           ju,
		Yuan:         []string{"上元", "中元", "下元"}[yuan],
		SolarTerm:    cal.PrevTerm.Name,
		TermAt:       cal.PrevTerm.At,
		YearGZ:       cal.Year.String(),
		MonthGZ:      cal.Month.String(),
		DayGZ:        cal.Day.String(),
		HourGZ:       cal.Hour.String(),
		XunShou:      xun.String() + instrument.String(),
		Instrument:   instrument.String(),
		ZhiFu:        zhiFu,
		ZhiFuPalace:  targetPalace,
		ZhiShi:       zhiShi,
		ZhiShiPalace: zhiShiPalace,
		Kong:         [2]calendar.Zhi{hourKongA, hourKongB},
		Horse:        horse,
		HorsePalace:  horsePalace,
		Palaces:      palaces,
		School:       school,
	}

	chart := schema.Chart{
		Method:        schema.MethodQiMen,
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

func fuTouOfDay(day calendar.GanZhi) calendar.GanZhi {
	idx := day.Index()
	if idx < 0 {
		return day
	}
	for i := 0; i < 5; i++ {
		gz := calendar.GanZhiFromIndex(idx - i)
		if gz.Gan == calendar.GanJia || gz.Gan == calendar.GanJi {
			return gz
		}
	}
	return day
}

func yuanOf(fuTou calendar.GanZhi) (int, error) {
	switch fuTou.Zhi {
	case calendar.ZhiZi, calendar.ZhiWu, calendar.ZhiMao, calendar.ZhiYou:
		return 0, nil
	case calendar.ZhiYin, calendar.ZhiShen, calendar.ZhiSi, calendar.ZhiHai:
		return 1, nil
	case calendar.ZhiChen, calendar.ZhiXu, calendar.ZhiChou, calendar.ZhiWei:
		return 2, nil
	}
	return 0, fmt.Errorf("无法由符头 %s 定元", fuTou)
}

func buildEarth(ju int, yang bool) [10]calendar.Gan {
	stems := []calendar.Gan{
		calendar.GanWu, calendar.GanJi, calendar.GanGeng, calendar.GanXin,
		calendar.GanRen, calendar.GanGui, calendar.GanDing, calendar.GanBing, calendar.GanYi,
	}
	var earth [10]calendar.Gan
	sign := 1
	if !yang {
		sign = -1
	}
	for i, stem := range stems {
		p := 1 + calendar.Mod(ju-1+sign*i, 9)
		earth[p] = stem
	}
	return earth
}

func findStem(earth [10]calendar.Gan, stem calendar.Gan) int {
	for p := 1; p <= 9; p++ {
		if earth[p] == stem {
			return p
		}
	}
	return 0
}

func ringIndex(p int) int {
	if p == 5 {
		p = 2
	}
	for i, v := range ringOrder {
		if v == p {
			return i
		}
	}
	return 0
}

func buildHeaven(earth [10]calendar.Gan, zhiFuHome, target int) ([10]string, [10]calendar.Gan) {
	var stars [10]string
	var heaven [10]calendar.Gan
	// 中五宫天禽寄坤，先固定中宫；八星沿八宫顺时针转。
	stars[5] = "天禽"
	heaven[5] = earth[5]
	if zhiFuHome == 5 {
		// 旬首在中宫时，值符按天禽处理，外八星不转。
		for _, p := range ringOrder {
			stars[p] = starHome[p]
			heaven[p] = earth[p]
		}
		return stars, heaven
	}
	homeRing := ringIndex(zhiFuHome)
	targetRing := ringIndex(target)
	for i, homePalace := range ringOrder {
		newPalace := ringOrder[calendar.Mod(targetRing+(i-homeRing), 8)]
		stars[newPalace] = starHome[homePalace]
		heaven[newPalace] = earth[homePalace]
	}
	return stars, heaven
}

func moveNumeric(start, steps, sign int) int {
	p := start
	for i := 0; i < steps; i++ {
		p = 1 + calendar.Mod(p-1+sign, 9)
	}
	return p
}

func buildDoors(zhiShi string, target int) [10]string {
	var doors [10]string
	start := ringIndex(target)
	doorIdx := 0
	for i, door := range doorOrder {
		if door == zhiShi {
			doorIdx = i
			break
		}
	}
	for i, door := range doorOrder {
		p := ringOrder[calendar.Mod(start+(i-doorIdx), 8)]
		doors[p] = door
	}
	return doors
}

func buildGods(zhiFuPalace int, yang bool) [10]string {
	var gods [10]string
	start := ringIndex(zhiFuPalace)
	sign := 1
	if !yang {
		sign = -1
	}
	for i, god := range godOrder {
		p := ringOrder[calendar.Mod(start+sign*i, 8)]
		gods[p] = god
	}
	return gods
}

func horseOf(hour calendar.Zhi) calendar.Zhi {
	switch hour {
	case calendar.ZhiShen, calendar.ZhiZi, calendar.ZhiChen:
		return calendar.ZhiYin
	case calendar.ZhiYin, calendar.ZhiWu, calendar.ZhiXu:
		return calendar.ZhiShen
	case calendar.ZhiSi, calendar.ZhiYou, calendar.ZhiChou:
		return calendar.ZhiHai
	default: // 亥卯未
		return calendar.ZhiSi
	}
}

func branchPalace(z calendar.Zhi) int {
	switch z {
	case calendar.ZhiZi:
		return 1
	case calendar.ZhiChou, calendar.ZhiYin:
		return 8
	case calendar.ZhiMao:
		return 3
	case calendar.ZhiChen, calendar.ZhiSi:
		return 4
	case calendar.ZhiWu:
		return 9
	case calendar.ZhiWei, calendar.ZhiShen:
		return 2
	case calendar.ZhiYou:
		return 7
	case calendar.ZhiXu, calendar.ZhiHai:
		return 6
	}
	return 0
}

func doorElement(door string) string {
	switch door {
	case "休门":
		return "水"
	case "生门", "死门":
		return "土"
	case "伤门", "杜门":
		return "木"
	case "景门":
		return "火"
	case "惊门", "开门":
		return "金"
	}
	return ""
}

func overcomes(a, b string) bool {
	switch a {
	case "木":
		return b == "土"
	case "土":
		return b == "水"
	case "水":
		return b == "火"
	case "火":
		return b == "金"
	case "金":
		return b == "木"
	}
	return false
}

// BuildFacts 从盘面重建事实清单，供追问快照恢复使用。
func BuildFacts(p Payload) []schema.Fact {
	facts := []schema.Fact{
		{
			Key:    "qimen.ju",
			Label:  "起局",
			Value:  fmt.Sprintf("%s%d局 · %s · %s", p.Dun, p.Ju, p.SolarTerm, p.Yuan),
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.sizhu",
			Label:  "四柱",
			Value:  fmt.Sprintf("%s年 %s月 %s日 %s时", p.YearGZ, p.MonthGZ, p.DayGZ, p.HourGZ),
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.xunshou",
			Label:  "旬首",
			Value:  p.XunShou,
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.zhifu",
			Label:  "值符",
			Value:  fmt.Sprintf("%s落%s", p.ZhiFu, palaceNames[p.ZhiFuPalace]),
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.zhishi",
			Label:  "值使",
			Value:  fmt.Sprintf("%s落%s", p.ZhiShi, palaceNames[p.ZhiShiPalace]),
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.kong",
			Label:  "时空亡",
			Value:  p.Kong[0].String() + "、" + p.Kong[1].String(),
			Weight: schema.WeightCore,
		},
		{
			Key:    "qimen.horse",
			Label:  "马星",
			Value:  fmt.Sprintf("%s在%s", p.Horse, palaceNames[p.HorsePalace]),
			Weight: schema.WeightNormal,
		},
	}
	for _, palace := range p.Palaces {
		value := fmt.Sprintf("天盘%s / 地盘%s · %s · %s", palace.HeavenGan, palace.EarthGan, palace.Star, palace.Door)
		if palace.God != "" {
			value += " · " + palace.God
		}
		if palace.MenPo {
			value += " · 门迫"
		}
		if palace.Kong {
			value += " · 空"
		}
		if palace.Horse {
			value += " · 马"
		}
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("qimen.palace.%d", palace.Index),
			Label:  palace.Name,
			Value:  value,
			Weight: schema.WeightNormal,
			Anchor: fmt.Sprintf("/payload/palaces/%d", palace.Index-1),
		})
	}

	// 针对近期考试、比赛的盘面参考。这里只给落宫，不替用户断成败。
	var jing, fu, shang, chong, zhiFuPalace, shiGanPalace int
	for _, palace := range p.Palaces {
		if palace.Door == "景门" {
			jing = palace.Index
		}
		if palace.Star == "天辅" {
			fu = palace.Index
		}
		if palace.Door == "伤门" {
			shang = palace.Index
		}
		if palace.Star == "天冲" {
			chong = palace.Index
		}
		if palace.IsZhiFu {
			zhiFuPalace = palace.Index
		}
	}
	// 时干落宫：天盘时干所在宫；若时干为甲，取旬首六仪。
	targetStem := calendar.GanJia
	if len(p.HourGZ) >= 1 {
		if g, err := calendar.ParseGan(string([]rune(p.HourGZ)[0])); err == nil {
			targetStem = g
		}
	}
	if targetStem == calendar.GanJia {
		if g, err := calendar.ParseGan(p.Instrument); err == nil {
			targetStem = g
		}
	}
	for _, palace := range p.Palaces {
		if palace.HeavenGan == targetStem {
			shiGanPalace = palace.Index
		}
	}
	refParts := []string{
		"考试：景门在" + palaceLabel(jing) + "、天辅在" + palaceLabel(fu),
		"比赛：值符在" + palaceLabel(zhiFuPalace) + "、时干在" + palaceLabel(shiGanPalace) + "、伤门在" + palaceLabel(shang) + "、天冲在" + palaceLabel(chong),
	}
	facts = append(facts, schema.Fact{
		Key:    "qimen.recent",
		Label:  "近期考试/比赛参考",
		Value:  strings.Join(refParts, "；"),
		Weight: schema.WeightNormal,
	})

	var po []string
	for _, palace := range p.Palaces {
		if palace.MenPo {
			po = append(po, palace.Name+"·"+palace.Door)
		}
	}
	if len(po) > 0 {
		facts = append(facts, schema.Fact{
			Key:    "qimen.menpo",
			Label:  "门迫",
			Value:  strings.Join(po, "、"),
			Weight: schema.WeightNormal,
		})
	}
	return facts
}

func palaceLabel(p int) string {
	if p < 1 || p > 9 {
		return "未定宫"
	}
	return palaceNames[p] + "（" + palaceGua[p] + "）"
}
