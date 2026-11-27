// Package ziwei 实现紫微斗数排盘。
//
// 本版按 docs/03 第 4 节的通行三合派规则实现：命身宫、五行局、十四主星、
// 常用辅星、生年四化、十二宫与大限。庚干四化采用文档默认“天同科、太阴忌”，
// 并在 school 中记录派别，便于后续扩展。
package ziwei

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	lunar "github.com/6tail/lunar-go/calendar"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 紫微斗数引擎版本。
const Version = "ziwei@0.1.0"

// Input 生盘输入。BirthDate 为 YYYY-MM-DD，BirthTime 为 HH:MM 或 HH:MM:SS。
type Input struct {
	Question        string
	Gender          string
	CalendarType    string // solar（默认）或 lunar
	BirthDate       string
	BirthTime       string
	IsLeapMonth     bool
	LeapMonthPolicy string // as_this_month（默认）或 as_next_month
	School          map[string]string
}

// BirthOptions 历法/真太阳时选项。
type BirthOptions struct {
	Location         *time.Location
	Longitude        float64
	HasLongitude     bool
	UseTrueSolarTime bool
}

// Birth 归一化后的出生信息。
type Birth struct {
	SolarTime      time.Time
	LunarYear      int
	LunarMonth     int // 1—12，闰月时 LunarLeap=true
	LunarDay       int
	LunarLeap      bool
	LunarDate      string
	YearGan        calendar.Gan
	YearZhi        calendar.Zhi
	YearGanZhi     string
	HourZhi        calendar.Zhi
	LateZi         bool
	TrueSolarUsed  bool
	CalendarType   string
	InputBirthDate string
	InputBirthTime string
}

// Star 盘面星曜。
type Star struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Hua  string `json:"hua,omitempty"`
}

// Sihua 生年四化。
type Sihua struct {
	Gan  string `json:"gan"`
	Star string `json:"star"`
	Hua  string `json:"hua"`
}

// DaXian 大限。
type DaXian struct {
	Index    int          `json:"index"`
	Name     string       `json:"name"`
	Branch   calendar.Zhi `json:"branch"`
	StartAge int          `json:"start_age"`
	EndAge   int          `json:"end_age"`
	Range    string       `json:"range"`
}

// Palace 十二宫之一。
type Palace struct {
	Index       int          `json:"index"`
	Branch      calendar.Zhi `json:"branch"`
	Stem        calendar.Gan `json:"stem"`
	StemBranch  string       `json:"stem_branch"`
	Name        string       `json:"name"`
	IsMing      bool         `json:"is_ming,omitempty"`
	IsShen      bool         `json:"is_shen,omitempty"`
	MajorStars  []Star       `json:"major_stars,omitempty"`
	MinorStars  []Star       `json:"minor_stars,omitempty"`
	Sihua       []string     `json:"sihua,omitempty"`
	DaxianStart int          `json:"daxian_start,omitempty"`
	DaxianEnd   int          `json:"daxian_end,omitempty"`
	DaxianRange string       `json:"daxian_range,omitempty"`
}

// Payload 紫微斗数盘面。
type Payload struct {
	Question        string            `json:"question,omitempty"`
	Gender          string            `json:"gender"`
	CalendarType    string            `json:"calendar_type"`
	BirthDate       string            `json:"birth_date"`
	BirthTime       string            `json:"birth_time"`
	IsLeapMonth     bool              `json:"is_leap_month"`
	LeapMonthPolicy string            `json:"leap_month_policy"`
	LunarDate       string            `json:"lunar_date"`
	SolarDate       string            `json:"solar_date"`
	YearGanZhi      string            `json:"year_gz"`
	MingGong        calendar.Zhi      `json:"ming_gong"`
	ShenGong        calendar.Zhi      `json:"shen_gong"`
	WuXingJu        string            `json:"wu_xing_ju"`
	ZiWeiPos        calendar.Zhi      `json:"ziwei_pos"`
	TianFuPos       calendar.Zhi      `json:"tianfu_pos"`
	Palaces         []Palace          `json:"palaces"`
	DaXian          []DaXian          `json:"da_xian"`
	SihuaYear       []Sihua           `json:"sihua_year"`
	School          map[string]string `json:"school,omitempty"`
}

// ParseBirth 归一化出生时间。
//
// 输入为公历时直接解析；输入为农历时先转换为公历。晚子时（23:00—23:59）
// 按“归次日”处理；早子时（00:00—00:59）仍归当日。真太阳时校正后若落入
// 晚子时，同样归次日。
func ParseBirth(in Input, opt BirthOptions) (Birth, error) {
	loc := opt.Location
	if loc == nil {
		loc = time.Local
	}
	if strings.TrimSpace(in.BirthDate) == "" || strings.TrimSpace(in.BirthTime) == "" {
		return Birth{}, fmt.Errorf("出生日期与出生时刻不能为空")
	}

	y, m, d, err := parseDate(in.BirthDate)
	if err != nil {
		return Birth{}, err
	}
	hour, minute, second, err := parseClock(in.BirthTime)
	if err != nil {
		return Birth{}, err
	}

	calendarType := strings.ToLower(strings.TrimSpace(in.CalendarType))
	if calendarType == "" {
		calendarType = "solar"
	}
	if calendarType != "solar" && calendarType != "lunar" {
		return Birth{}, fmt.Errorf("calendar_type 只能是 solar 或 lunar")
	}

	var solarTime time.Time
	if calendarType == "solar" {
		solarTime = time.Date(y, time.Month(m), d, hour, minute, second, 0, loc)
	} else {
		month := m
		if in.IsLeapMonth {
			month = -month
		}
		l, err := newLunarSafe(y, month, d, hour, minute, second)
		if err != nil {
			return Birth{}, err
		}
		s := l.GetSolar()
		solarTime = time.Date(s.GetYear(), time.Month(s.GetMonth()), s.GetDay(), hour, minute, second, 0, loc)
	}

	trueSolar := false
	if opt.UseTrueSolarTime && opt.HasLongitude {
		_, offset := solarTime.Zone()
		solarTime = calendar.TrueSolarTime(solarTime, opt.Longitude, float64(offset)/3600)
		trueSolar = true
	}

	lateZi := false
	if solarTime.Hour() == 23 {
		// 晚子时归次日；保留 23 点，只把日期推后。
		solarTime = time.Date(solarTime.Year(), solarTime.Month(), solarTime.Day()+1,
			solarTime.Hour(), solarTime.Minute(), solarTime.Second(), 0, loc)
		lateZi = true
	}

	l := lunar.NewLunarFromSolar(lunar.NewSolar(solarTime.Year(), int(solarTime.Month()), solarTime.Day(), solarTime.Hour(), solarTime.Minute(), solarTime.Second()))
	lunarMonth := l.GetMonth()
	leap := lunarMonth < 0
	if leap {
		lunarMonth = -lunarMonth
	}

	yearGZ := l.GetYearInGanZhi()
	yg, err := calendar.ParseGan(string([]rune(yearGZ)[0]))
	if err != nil {
		return Birth{}, fmt.Errorf("解析生年天干: %w", err)
	}
	yz, err := calendar.ParseZhi(string([]rune(yearGZ)[1]))
	if err != nil {
		return Birth{}, fmt.Errorf("解析生年地支: %w", err)
	}

	cal, err := calendar.Build(calendar.Request{At: solarTime})
	if err != nil {
		return Birth{}, fmt.Errorf("计算出生历法: %w", err)
	}

	policy := strings.ToLower(strings.TrimSpace(in.LeapMonthPolicy))
	if policy == "" {
		policy = "as_this_month"
	}
	if policy != "as_this_month" && policy != "as_next_month" {
		return Birth{}, fmt.Errorf("leap_month_policy 只能是 as_this_month 或 as_next_month")
	}

	return Birth{
		SolarTime:      solarTime,
		LunarYear:      l.GetYear(),
		LunarMonth:     lunarMonth,
		LunarDay:       l.GetDay(),
		LunarLeap:      leap,
		LunarDate:      fmt.Sprintf("%s年%s月%s", yearGZ, l.GetMonthInChinese(), l.GetDayInChinese()),
		YearGan:        yg,
		YearZhi:        yz,
		YearGanZhi:     yearGZ,
		HourZhi:        cal.Hour.Zhi,
		LateZi:         lateZi,
		TrueSolarUsed:  trueSolar,
		CalendarType:   calendarType,
		InputBirthDate: in.BirthDate,
		InputBirthTime: in.BirthTime,
	}, nil
}

// Cast 排紫微斗数盘。cal 应使用 birth.SolarTime 构建，且不再重复做真太阳时校正。
func Cast(in Input, birth Birth, cal calendar.Context) (schema.Chart, error) {
	if !birth.YearGan.Valid() || !birth.YearZhi.Valid() || !birth.HourZhi.Valid() {
		return schema.Chart{}, fmt.Errorf("出生信息不完整")
	}
	gender, err := normalizeGender(in.Gender)
	if err != nil {
		return schema.Chart{}, err
	}
	month, err := monthForPalace(birth, in.LeapMonthPolicy)
	if err != nil {
		return schema.Chart{}, err
	}

	ming := mingPalace(month, int(birth.HourZhi))
	shen := shenPalace(month, int(birth.HourZhi))
	mingStem := palaceStem(birth.YearGan, int(ming))
	gz, err := calendar.NewGanZhi(mingStem, ming)
	if err != nil {
		return schema.Chart{}, err
	}
	wuxing, ok := calendar.NaYinWuXing(gz)
	if !ok {
		return schema.Chart{}, fmt.Errorf("命宫纳音五行无法确定: %s", gz)
	}
	ju := juNumber(wuxing)
	ziwei := ziweiPosition(birth.LunarDay, ju)
	tianfu := calendar.Zhi(calendar.Mod(4-int(ziwei), 12))

	// 先建十二宫骨架。
	palaces := make([]Palace, 12)
	names := []string{"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄", "迁移", "交友", "官禄", "田宅", "福德", "父母"}
	for i, name := range names {
		branch := calendar.Zhi(calendar.Mod(int(ming)-i, 12))
		stem := palaceStem(birth.YearGan, int(branch))
		palaces[branch] = Palace{
			Index:      int(branch),
			Branch:     branch,
			Stem:       stem,
			StemBranch: stem.String() + branch.String(),
			Name:       name,
			IsMing:     branch == ming,
			IsShen:     branch == shen,
		}
	}

	sihua := sihuaFor(birth.YearGan, in.School)
	starHua := map[string]string{}
	for _, s := range sihua.list {
		starHua[s.Star] = s.Hua
	}

	addMajorStars(palaces, ziwei, tianfu, starHua)
	addMinorStars(palaces, birth, month, int(birth.HourZhi), starHua)

	// 大限：命宫起，阳男阴女顺行，阴男阳女逆行。
	daXian, err := buildDaXian(palaces, ming, ju, birth.YearGan, gender)
	if err != nil {
		return schema.Chart{}, err
	}

	payload := Payload{
		Question:        in.Question,
		Gender:          gender,
		CalendarType:    birth.CalendarType,
		BirthDate:       birth.InputBirthDate,
		BirthTime:       birth.InputBirthTime,
		IsLeapMonth:     birth.LunarLeap,
		LeapMonthPolicy: sihua.leapPolicy(in.LeapMonthPolicy),
		LunarDate:       birth.LunarDate,
		SolarDate:       birth.SolarTime.Format("2006-01-02 15:04"),
		YearGanZhi:      birth.YearGanZhi,
		MingGong:        ming,
		ShenGong:        shen,
		WuXingJu:        juLabel(ju),
		ZiWeiPos:        ziwei,
		TianFuPos:       tianfu,
		Palaces:         palaces,
		DaXian:          daXian,
		SihuaYear:       sihua.list,
		School:          sihua.school,
	}
	if birth.LateZi {
		payload.School["late_zi"] = "next_day"
	}
	if birth.TrueSolarUsed {
		payload.School["true_solar"] = "used"
	}

	chart := schema.Chart{
		Method:        schema.MethodZiWei,
		EngineVersion: Version,
		Calendar:      cal,
		Question:      in.Question,
		School:        payload.School,
		Payload:       payload,
		Facts:         BuildFacts(payload),
		CreatedAt:     time.Now().UTC(),
	}
	if err := chart.Validate(); err != nil {
		return schema.Chart{}, err
	}
	return chart, nil
}

func parseDate(s string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(s), "-")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("出生日期格式应为 YYYY-MM-DD")
	}
	y, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	d, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, fmt.Errorf("出生日期格式应为 YYYY-MM-DD")
	}
	if y < 1900 || y > 2100 || m < 1 || m > 12 || d < 1 || d > 31 {
		return 0, 0, 0, fmt.Errorf("出生日期超出支持范围（1900—2100）")
	}
	return y, m, d, nil
}

func parseClock(s string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, 0, 0, fmt.Errorf("出生时刻格式应为 HH:MM 或 HH:MM:SS")
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	sec := 0
	var err3 error
	if len(parts) == 3 {
		sec, err3 = strconv.Atoi(parts[2])
	}
	if err1 != nil || err2 != nil || err3 != nil || h < 0 || h > 23 || m < 0 || m > 59 || sec < 0 || sec > 59 {
		return 0, 0, 0, fmt.Errorf("出生时刻非法")
	}
	return h, m, sec, nil
}

func newLunarSafe(year, month, day, hour, minute, second int) (l *lunar.Lunar, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("农历日期非法: %v", r)
		}
	}()
	return lunar.NewLunar(year, month, day, hour, minute, second), nil
}

func normalizeGender(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "male", "m", "男", "乾":
		return "male", nil
	case "female", "f", "女", "坤":
		return "female", nil
	default:
		return "", fmt.Errorf("gender 必须是 male 或 female")
	}
}

func monthForPalace(birth Birth, policy string) (int, error) {
	m := birth.LunarMonth
	if birth.LunarLeap {
		switch strings.ToLower(strings.TrimSpace(policy)) {
		case "", "as_this_month":
			// 闰月按本月计。
		case "as_next_month":
			m++
		default:
			return 0, fmt.Errorf("leap_month_policy 只能是 as_this_month 或 as_next_month")
		}
	}
	if m > 12 {
		m = 12
	}
	if m < 1 {
		m = 1
	}
	return m, nil
}

func mingPalace(month, hour int) calendar.Zhi {
	monthPalace := calendar.Zhi(calendar.Mod(2+month-1, 12))
	return calendar.Zhi(calendar.Mod(int(monthPalace)-hour, 12))
}

func shenPalace(month, hour int) calendar.Zhi {
	monthPalace := calendar.Zhi(calendar.Mod(2+month-1, 12))
	return calendar.Zhi(calendar.Mod(int(monthPalace)+hour, 12))
}

func palaceStem(yearGan calendar.Gan, branch int) calendar.Gan {
	base := 2 // 甲己年寅宫起丙
	switch yearGan {
	case calendar.GanJia, calendar.GanJi:
		base = 2
	case calendar.GanYi, calendar.GanGeng:
		base = 4
	case calendar.GanBing, calendar.GanXin:
		base = 6
	case calendar.GanDing, calendar.GanRen:
		base = 8
	case calendar.GanWu, calendar.GanGui:
		base = 0
	}
	return calendar.Gan(calendar.Mod(base+calendar.Mod(branch-2, 12), 10))
}

func juNumber(w calendar.WuXing) int {
	switch w {
	case calendar.WuXingShui:
		return 2
	case calendar.WuXingMu:
		return 3
	case calendar.WuXingJin:
		return 4
	case calendar.WuXingTu:
		return 5
	default:
		return 6
	}
}

func juLabel(ju int) string {
	switch ju {
	case 2:
		return "水二局"
	case 3:
		return "木三局"
	case 4:
		return "金四局"
	case 5:
		return "土五局"
	default:
		return "火六局"
	}
}

func ziweiPosition(day, ju int) calendar.Zhi {
	k := 0
	for (day+k)%ju != 0 {
		k++
	}
	q := (day + k) / ju
	base := calendar.Mod(2+q-1, 12)
	pos := base
	if k > 0 {
		pos = calendar.Mod(base-k, 12)
	}
	return calendar.Zhi(pos)
}

func addMajorStars(palaces []Palace, ziwei, tianfu calendar.Zhi, hua map[string]string) {
	add := func(pos calendar.Zhi, name string) {
		if pos < 0 || pos >= 12 {
			return
		}
		palaces[pos].MajorStars = append(palaces[pos].MajorStars, Star{Name: name, Kind: "主星", Hua: hua[name]})
	}
	for _, item := range []struct {
		offset int
		name   string
	}{
		{0, "紫微"}, {-1, "天机"}, {-3, "太阳"}, {-4, "武曲"}, {-5, "天同"}, {-8, "廉贞"},
	} {
		add(calendar.Zhi(calendar.Mod(int(ziwei)+item.offset, 12)), item.name)
	}
	for _, item := range []struct {
		offset int
		name   string
	}{
		{0, "天府"}, {1, "太阴"}, {2, "贪狼"}, {3, "巨门"}, {4, "天相"}, {5, "天梁"}, {6, "七杀"}, {10, "破军"},
	} {
		add(calendar.Zhi(calendar.Mod(int(tianfu)+item.offset, 12)), item.name)
	}
}

func addMinorStars(palaces []Palace, birth Birth, month, hour int, hua map[string]string) {
	add := func(pos calendar.Zhi, name, kind string) {
		palaces[pos].MinorStars = append(palaces[pos].MinorStars, Star{Name: name, Kind: kind, Hua: hua[name]})
	}

	// 左辅：辰起正月顺行；右弼：戌起正月逆行。
	add(calendar.Zhi(calendar.Mod(4+month-1, 12)), "左辅", "辅星")
	add(calendar.Zhi(calendar.Mod(10-(month-1), 12)), "右弼", "辅星")
	// 文昌：戌起子时逆行；文曲：辰起子时顺行。
	add(calendar.Zhi(calendar.Mod(10-hour, 12)), "文昌", "辅星")
	add(calendar.Zhi(calendar.Mod(4+hour, 12)), "文曲", "辅星")

	kui, yue := kuiYue(birth.YearGan)
	add(kui, "天魁", "辅星")
	add(yue, "天钺", "辅星")

	lu := luCun(birth.YearGan)
	add(lu, "禄存", "辅星")
	add(calendar.Zhi(calendar.Mod(int(lu)+1, 12)), "擎羊", "煞星")
	add(calendar.Zhi(calendar.Mod(int(lu)-1, 12)), "陀罗", "煞星")

	add(huoXing(birth.YearZhi, hour), "火星", "煞星")
	add(lingXing(birth.YearZhi, hour), "铃星", "煞星")
	add(calendar.Zhi(calendar.Mod(11+hour, 12)), "地劫", "煞星")
	add(calendar.Zhi(calendar.Mod(11-hour, 12)), "地空", "煞星")
	add(tianMa(birth.YearZhi), "天马", "辅星")
}

func kuiYue(yearGan calendar.Gan) (calendar.Zhi, calendar.Zhi) {
	switch yearGan {
	case calendar.GanJia, calendar.GanWu, calendar.GanGeng:
		return calendar.ZhiChou, calendar.ZhiWei
	case calendar.GanYi, calendar.GanJi:
		return calendar.ZhiZi, calendar.ZhiShen
	case calendar.GanBing, calendar.GanDing:
		return calendar.ZhiHai, calendar.ZhiYou
	case calendar.GanRen, calendar.GanGui:
		return calendar.ZhiMao, calendar.ZhiSi
	default:
		return calendar.ZhiWu, calendar.ZhiYin
	}
}

func luCun(yearGan calendar.Gan) calendar.Zhi {
	switch yearGan {
	case calendar.GanJia:
		return calendar.ZhiYin
	case calendar.GanYi:
		return calendar.ZhiMao
	case calendar.GanBing, calendar.GanWu:
		return calendar.ZhiSi
	case calendar.GanDing, calendar.GanJi:
		return calendar.ZhiWu
	case calendar.GanGeng:
		return calendar.ZhiShen
	case calendar.GanXin:
		return calendar.ZhiYou
	case calendar.GanRen:
		return calendar.ZhiHai
	default:
		return calendar.ZhiZi
	}
}

func huoXing(yearZhi calendar.Zhi, hour int) calendar.Zhi {
	base := 2
	switch int(yearZhi) % 4 {
	case 0: // 申子辰
		base = 2
	case 1: // 巳酉丑
		base = 3
	case 2: // 寅午戌
		base = 1
	default: // 亥卯未
		base = 9
	}
	return calendar.Zhi(calendar.Mod(base+hour, 12))
}

func lingXing(yearZhi calendar.Zhi, hour int) calendar.Zhi {
	base := 10
	switch int(yearZhi) % 4 {
	case 0: // 申子辰
		base = 10
	case 1: // 巳酉丑
		base = 10
	case 2: // 寅午戌
		base = 3
	default: // 亥卯未
		base = 4
	}
	return calendar.Zhi(calendar.Mod(base+hour, 12))
}

func tianMa(yearZhi calendar.Zhi) calendar.Zhi {
	switch int(yearZhi) % 4 {
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

type sihuaData struct {
	list   []Sihua
	school map[string]string
}

func sihuaFor(yearGan calendar.Gan, custom map[string]string) sihuaData {
	school := map[string]string{"sihua": "standard", "geng": "tiantong_ke_taiyin_ji"}
	for k, v := range custom {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			school[k] = v
		}
	}
	var stars [4]string
	switch yearGan {
	case calendar.GanJia:
		stars = [4]string{"廉贞", "破军", "武曲", "太阳"}
	case calendar.GanYi:
		stars = [4]string{"天机", "天梁", "紫微", "太阴"}
	case calendar.GanBing:
		stars = [4]string{"天同", "天机", "文昌", "廉贞"}
	case calendar.GanDing:
		stars = [4]string{"太阴", "天同", "天机", "巨门"}
	case calendar.GanWu:
		stars = [4]string{"贪狼", "太阴", "右弼", "天机"}
	case calendar.GanJi:
		stars = [4]string{"武曲", "贪狼", "天梁", "文曲"}
	case calendar.GanGeng:
		stars = [4]string{"太阳", "武曲", "天同", "太阴"}
	case calendar.GanXin:
		stars = [4]string{"巨门", "太阳", "文曲", "文昌"}
	case calendar.GanRen:
		stars = [4]string{"天梁", "紫微", "左辅", "武曲"}
	default: // 癸
		stars = [4]string{"破军", "巨门", "太阴", "贪狼"}
	}
	names := [4]string{"禄", "权", "科", "忌"}
	out := make([]Sihua, 0, 4)
	for i, star := range stars {
		out = append(out, Sihua{Gan: yearGan.String(), Star: star, Hua: names[i]})
	}
	return sihuaData{list: out, school: school}
}

func (s sihuaData) leapPolicy(input string) string {
	if strings.EqualFold(strings.TrimSpace(input), "as_next_month") {
		return "as_next_month"
	}
	return "as_this_month"
}

func buildDaXian(palaces []Palace, ming calendar.Zhi, ju int, yearGan calendar.Gan, gender string) ([]DaXian, error) {
	forward := (yearGan.YinYang() == calendar.Yang && gender == "male") ||
		(yearGan.YinYang() == calendar.Yin && gender == "female")
	step := 1
	if !forward {
		step = -1
	}
	out := make([]DaXian, 0, 12)
	startAge := ju
	for i := 0; i < 12; i++ {
		branch := calendar.Zhi(calendar.Mod(int(ming)+step*i, 12))
		rangeText := fmt.Sprintf("%d—%d", startAge, startAge+9)
		p := &palaces[branch]
		p.DaxianStart = startAge
		p.DaxianEnd = startAge + 9
		p.DaxianRange = rangeText
		out = append(out, DaXian{
			Index:    i + 1,
			Name:     p.Name,
			Branch:   branch,
			StartAge: startAge,
			EndAge:   startAge + 9,
			Range:    rangeText,
		})
		startAge += 10
	}
	if len(out) != 12 {
		return nil, fmt.Errorf("大限数量异常: %d", len(out))
	}
	return out, nil
}

// Summary 返回一行紫微摘要。
func (p Payload) Summary() string {
	return fmt.Sprintf("%s · 命宫%s · 身宫%s · %s · 紫微在%s", p.YearGanZhi, p.MingGong, p.ShenGong, p.WuXingJu, p.ZiWeiPos)
}
