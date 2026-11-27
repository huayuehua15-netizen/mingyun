// Package xiaoliuren 实现小六壬月日时起课。
//
// 本版按通行的“大安起正月，月上起日，日上起时”计算，等价于
// (农历月 + 农历日 + 时支序数 - 3) mod 6。它只做近期事情的落宫与取象，
// 不预测具体名次或分数；解读由上层按盘面事实生成。
package xiaoliuren

import (
	"fmt"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 小六壬引擎版本。
const Version = "xiaoliuren@0.1.0"

// Input 起课输入。小六壬以时刻起课，问题只参与后续解读。
type Input struct {
	Question string
	School   map[string]string
}

// Palace 六宫之一。
type Palace struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Element   string `json:"element"`
	Deity     string `json:"deity"`
	Direction string `json:"direction"`
	Luck      string `json:"luck"`
	Meaning   string `json:"meaning"`
}

// Payload 小六壬盘面。
type Payload struct {
	Question    string            `json:"question,omitempty"`
	LunarMonth  int               `json:"lunar_month"`
	LunarDay    int               `json:"lunar_day"`
	LunarLeap   bool              `json:"lunar_leap,omitempty"`
	HourZhi     calendar.Zhi      `json:"hour_zhi"`
	MonthPalace Palace            `json:"month_palace"`
	DayPalace   Palace            `json:"day_palace"`
	FinalPalace Palace            `json:"final_palace"`
	SixPalaces  []Palace          `json:"six_palaces"`
	School      map[string]string `json:"school,omitempty"`
}

// Summary 一行摘要。
func (p Payload) Summary() string {
	return fmt.Sprintf("小六壬 · %s · 月%s 日%s 时%s", p.FinalPalace.Name, p.MonthPalace.Name, p.DayPalace.Name, p.FinalPalace.Name)
}

var sixPalaces = []Palace{
	{Index: 0, Name: "大安", Element: "木", Deity: "青龙", Direction: "东", Luck: "吉", Meaning: "安定守成，事缓可成"},
	{Index: 1, Name: "留连", Element: "水", Deity: "玄武", Direction: "北", Luck: "凶", Meaning: "拖延纠缠，反复难决"},
	{Index: 2, Name: "速喜", Element: "火", Deity: "朱雀", Direction: "南", Luck: "吉", Meaning: "喜讯速至，宜快不宜拖"},
	{Index: 3, Name: "赤口", Element: "金", Deity: "白虎", Direction: "西", Luck: "凶", Meaning: "口舌争执，文书须慎"},
	{Index: 4, Name: "小吉", Element: "木", Deity: "六合", Direction: "东", Luck: "吉", Meaning: "小成和合，可进一步"},
	{Index: 5, Name: "空亡", Element: "土", Deity: "勾陈", Direction: "中", Luck: "凶", Meaning: "落空迟滞，宜另寻路径"},
}

func palaceAt(index int) Palace {
	return sixPalaces[calendar.Mod(index, 6)]
}

// Cast 起一张小六壬课。
func Cast(in Input, cal calendar.Context) (schema.Chart, error) {
	if !cal.Hour.Zhi.Valid() {
		return schema.Chart{}, fmt.Errorf("历法上下文缺少时支")
	}

	t := cal.TrueSolarTime
	if cal.DayBoundary == "zi_chu" && t.Hour() >= 23 {
		// 小六壬用农历月日，晚子时归次日，才能与日柱口径一致。
		t = t.Add(24 * time.Hour)
	}
	lunarYear, lunarMonth, lunarDay, leap := calendar.LunarParts(t)
	if lunarMonth < 1 || lunarMonth > 12 || lunarDay < 1 || lunarDay > 30 {
		return schema.Chart{}, fmt.Errorf("农历月日超出范围：%d月%d日", lunarMonth, lunarDay)
	}
	_ = lunarYear

	// 大安起正月：月宫下标 = 农历月 - 1；日从月宫起；时从日宫起，
	// 子时为 0，因此最终下标 = 月 + 日 + 时支序数 - 3。
	monthIndex := lunarMonth - 1
	dayIndex := monthIndex + lunarDay - 1
	finalIndex := dayIndex + int(cal.Hour.Zhi)
	monthPalace := palaceAt(monthIndex)
	dayPalace := palaceAt(dayIndex)
	finalPalace := palaceAt(finalIndex)

	school := map[string]string{
		"method":           "month_day_hour",
		"month_start":      "daan",
		"hour_branch_base": "zi=1",
		"formula":          "lunar_month+lunar_day+hour_branch-3 mod 6",
		"xiaojie_element":  "wood_liuhe",
	}
	for k, v := range in.School {
		if k != "" && v != "" {
			school[k] = v
		}
	}

	payload := Payload{
		Question:    in.Question,
		LunarMonth:  lunarMonth,
		LunarDay:    lunarDay,
		LunarLeap:   leap,
		HourZhi:     cal.Hour.Zhi,
		MonthPalace: monthPalace,
		DayPalace:   dayPalace,
		FinalPalace: finalPalace,
		SixPalaces:  append([]Palace(nil), sixPalaces...),
		School:      school,
	}

	chart := schema.Chart{
		Method:        schema.MethodXiaoLiuRen,
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

// BuildFacts 从盘面重建事实清单，供追问快照恢复使用。
func BuildFacts(p Payload) []schema.Fact {
	monthDay := fmt.Sprintf("%d月%d日", p.LunarMonth, p.LunarDay)
	if p.LunarLeap {
		monthDay = fmt.Sprintf("闰%d月%d日", p.LunarMonth, p.LunarDay)
	}
	facts := []schema.Fact{
		{
			Key:    "xiaoliuren.lunar",
			Label:  "农历月日",
			Value:  monthDay + " · " + p.HourZhi.String() + "时",
			Weight: schema.WeightCore,
		},
		{
			Key:    "xiaoliuren.three",
			Label:  "三宫",
			Value:  fmt.Sprintf("月%s · 日%s · 时%s", p.MonthPalace.Name, p.DayPalace.Name, p.FinalPalace.Name),
			Weight: schema.WeightCore,
		},
		{
			Key:    "xiaoliuren.final",
			Label:  "落宫",
			Value:  fmt.Sprintf("%s（%s·%s·%s）· %s", p.FinalPalace.Name, p.FinalPalace.Element, p.FinalPalace.Deity, p.FinalPalace.Direction, p.FinalPalace.Luck),
			Weight: schema.WeightCore,
			Anchor: "/payload/final_palace",
		},
		{
			Key:    "xiaoliuren.window",
			Label:  "近事窗口",
			Value:  "以起课时刻起 1—15 日内；月宫看背景，日宫看事体，时宫看当下动向",
			Weight: schema.WeightNormal,
		},
	}
	for _, palace := range p.SixPalaces {
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("xiaoliuren.palace.%d", palace.Index),
			Label:  palace.Name,
			Value:  fmt.Sprintf("%s · %s · %s · %s", palace.Element, palace.Deity, palace.Direction, palace.Meaning),
			Weight: schema.WeightNormal,
			Anchor: fmt.Sprintf("/payload/six_palaces/%d", palace.Index),
		})
	}
	return facts
}
