// Package meihua 实现梅花易数时间起卦与报数起卦。
//
// 时间起卦按“年支数 + 农历月 + 农历日”取上卦，再加时支数取下卦，
// 总数除六取动爻；报数起卦按两组或三组数字取卦。断法只提供本卦、互卦、
// 变卦与体用生克，不直接替用户判定成败。
package meihua

import (
	"fmt"
	"strings"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 梅花易数引擎版本。
const Version = "meihua@0.1.0"

// Mode 起卦方式。
type Mode string

const (
	ModeTime    Mode = "time"
	ModeNumbers Mode = "numbers"
)

// Input 起卦输入。
type Input struct {
	Question string
	Mode     Mode
	Numbers  []int
	School   map[string]string
}

// Trigram 八卦。
type Trigram struct {
	Name    string `json:"name"`
	Element string `json:"element"`
}

// Hexagram 六爻卦。
type Hexagram struct {
	Name  string `json:"name"`
	Upper string `json:"upper"`
	Lower string `json:"lower"`
	Lines []bool `json:"lines"` // 自下而上，true 为阳
}

// Payload 梅花易数盘面。
type Payload struct {
	Question        string            `json:"question,omitempty"`
	Mode            string            `json:"mode"`
	Numbers         []int             `json:"numbers,omitempty"`
	YearZhi         string            `json:"year_zhi,omitempty"`
	LunarDate       string            `json:"lunar_date,omitempty"`
	HourZhi         string            `json:"hour_zhi,omitempty"`
	UpperNumber     int               `json:"upper_number"`
	LowerNumber     int               `json:"lower_number"`
	MovingNumber    int               `json:"moving_number"`
	BenGua          Hexagram          `json:"ben_gua"`
	HuGua           Hexagram          `json:"hu_gua"`
	BianGua         Hexagram          `json:"bian_gua"`
	MovingLine      int               `json:"moving_line"`
	TiGua           Trigram           `json:"ti_gua"`
	YongGua         Trigram           `json:"yong_gua"`
	Relation        string            `json:"relation"`
	ChangedRelation string            `json:"changed_relation"`
	School          map[string]string `json:"school,omitempty"`
}

// Summary 一行摘要。
func (p Payload) Summary() string {
	relation := p.Relation
	if relation == "" {
		relation = "体用待定"
	}
	return fmt.Sprintf("梅花易数 · %s 之 %s · 第%d爻动 · %s", p.BenGua.Name, p.BianGua.Name, p.MovingLine, relation)
}

// Cast 起一张梅花易数卦。
func Cast(in Input, cal calendar.Context) (schema.Chart, error) {
	mode := in.Mode
	if mode == "" {
		mode = ModeTime
	}
	if mode != ModeTime && mode != ModeNumbers {
		return schema.Chart{}, fmt.Errorf("mode 只能是 time 或 numbers")
	}

	var upperNum, lowerNum, movingNum int
	yearZhi, lunarDate, hourZhi := "", "", ""
	var numbers []int
	if mode == ModeTime {
		if !cal.Year.Zhi.Valid() || !cal.Hour.Zhi.Valid() {
			return schema.Chart{}, fmt.Errorf("历法上下文缺少年支或时支")
		}
		t := cal.TrueSolarTime
		if cal.DayBoundary == "zi_chu" && t.Hour() >= 23 {
			t = t.Add(24 * time.Hour)
		}
		_, month, day, leap := calendar.LunarParts(t)
		yearNum := int(cal.Year.Zhi) + 1
		hourNum := int(cal.Hour.Zhi) + 1
		upperNum = yearNum + month + day
		lowerNum = upperNum + hourNum
		movingNum = lowerNum
		yearZhi = cal.Year.Zhi.String()
		hourZhi = cal.Hour.Zhi.String()
		if leap {
			lunarDate = fmt.Sprintf("闰%d月%d日", month, day)
		} else {
			lunarDate = fmt.Sprintf("%d月%d日", month, day)
		}
	} else {
		if len(in.Numbers) != 2 && len(in.Numbers) != 3 {
			return schema.Chart{}, fmt.Errorf("报数起卦需 2 或 3 个正整数")
		}
		for i, n := range in.Numbers {
			if n <= 0 || n > 999999 {
				return schema.Chart{}, fmt.Errorf("第%d个数必须是 1—999999 的正整数", i+1)
			}
		}
		numbers = append([]int(nil), in.Numbers...)
		upperNum = numbers[0]
		lowerNum = numbers[1]
		if len(numbers) == 3 {
			movingNum = numbers[2]
		} else {
			movingNum = upperNum + lowerNum
		}
	}

	upper := trigramFromNumber(upperNum)
	lower := trigramFromNumber(lowerNum)
	movingLine := movingNum % 6
	if movingLine == 0 {
		movingLine = 6
	}

	var lines [6]bool
	for i := 0; i < 3; i++ {
		lines[i] = trigramLines(lower)[i]
		lines[i+3] = trigramLines(upper)[i]
	}
	ben := hexagramFromLines(lines)

	// 互卦：二三四爻为下卦，三四五爻为上卦。
	var huLines [6]bool
	huLines[0], huLines[1], huLines[2] = lines[1], lines[2], lines[3]
	huLines[3], huLines[4], huLines[5] = lines[2], lines[3], lines[4]
	hu := hexagramFromLines(huLines)

	bianLines := lines
	bianLines[movingLine-1] = !bianLines[movingLine-1]
	bian := hexagramFromLines(bianLines)

	tiIndex, yongIndex := upper, lower
	if movingLine > 3 {
		tiIndex, yongIndex = lower, upper
	}
	ti := Trigram{Name: trigrams[tiIndex].Name, Element: trigrams[tiIndex].Element}
	yong := Trigram{Name: trigrams[yongIndex].Name, Element: trigrams[yongIndex].Element}
	relation := elementRelation(ti.Element, yong.Element)

	changedTrigram := lower
	if movingLine > 3 {
		changedTrigram = upper
	}
	changedRelation := changedRelationText(ti.Element, trigrams[changedTrigram].Element)

	school := map[string]string{
		"qigua":     string(mode),
		"year_base": "zi=1",
		"moving":    "total_mod6",
		"tiyong":    "moving_trigram_is_yong",
	}
	for k, v := range in.School {
		if k != "" && v != "" {
			school[k] = v
		}
	}

	payload := Payload{
		Question:        in.Question,
		Mode:            string(mode),
		Numbers:         numbers,
		YearZhi:         yearZhi,
		LunarDate:       lunarDate,
		HourZhi:         hourZhi,
		UpperNumber:     upperNum,
		LowerNumber:     lowerNum,
		MovingNumber:    movingNum,
		BenGua:          ben,
		HuGua:           hu,
		BianGua:         bian,
		MovingLine:      movingLine,
		TiGua:           ti,
		YongGua:         yong,
		Relation:        relation,
		ChangedRelation: changedRelation,
		School:          school,
	}

	chart := schema.Chart{
		Method:        schema.MethodMeiHua,
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
	modeText := "时间起卦"
	if p.Mode == string(ModeNumbers) {
		modeText = "报数起卦：" + joinNumbers(p.Numbers)
	}
	timeText := ""
	if p.LunarDate != "" {
		timeText = fmt.Sprintf("%s年 %s %s时", p.YearZhi, p.LunarDate, p.HourZhi)
	}
	facts := []schema.Fact{
		{
			Key:    "meihua.mode",
			Label:  "起卦方式",
			Value:  modeText,
			Weight: schema.WeightCore,
		},
		{
			Key:    "meihua.ben",
			Label:  "本卦",
			Value:  fmt.Sprintf("%s（上%s下%s）", p.BenGua.Name, p.BenGua.Upper, p.BenGua.Lower),
			Weight: schema.WeightCore,
			Anchor: "/payload/ben_gua",
		},
		{
			Key:    "meihua.hu",
			Label:  "互卦",
			Value:  fmt.Sprintf("%s（上%s下%s）", p.HuGua.Name, p.HuGua.Upper, p.HuGua.Lower),
			Weight: schema.WeightNormal,
			Anchor: "/payload/hu_gua",
		},
		{
			Key:    "meihua.bian",
			Label:  "变卦",
			Value:  fmt.Sprintf("%s（上%s下%s）", p.BianGua.Name, p.BianGua.Upper, p.BianGua.Lower),
			Weight: schema.WeightCore,
			Anchor: "/payload/bian_gua",
		},
		{
			Key:    "meihua.moving",
			Label:  "动爻",
			Value:  fmt.Sprintf("第%d爻动", p.MovingLine),
			Weight: schema.WeightCore,
		},
		{
			Key:    "meihua.tiyong",
			Label:  "体用",
			Value:  fmt.Sprintf("体卦%s（%s）· 用卦%s（%s）", p.TiGua.Name, p.TiGua.Element, p.YongGua.Name, p.YongGua.Element),
			Weight: schema.WeightCore,
		},
		{
			Key:    "meihua.relation",
			Label:  "体用关系",
			Value:  p.Relation,
			Weight: schema.WeightCore,
		},
		{
			Key:    "meihua.changed",
			Label:  "变卦关系",
			Value:  p.ChangedRelation,
			Weight: schema.WeightNormal,
		},
		{
			Key:    "meihua.method",
			Label:  "取象提示",
			Value:  "本卦看当下，互卦看过程，变卦看结果；体为己身，用为所问之事。考试重文书与名次，比赛重体用胜负与变卦收束。",
			Weight: schema.WeightNormal,
		},
	}
	if timeText != "" {
		facts = append(facts, schema.Fact{
			Key:    "meihua.time",
			Label:  "起卦时间",
			Value:  timeText,
			Weight: schema.WeightNormal,
		})
	}
	return facts
}

func joinNumbers(numbers []int) string {
	parts := make([]string, 0, len(numbers))
	for _, n := range numbers {
		parts = append(parts, fmt.Sprintf("%d", n))
	}
	return strings.Join(parts, "、")
}

func elementRelation(a, b string) string {
	if a == b {
		return "体用比和"
	}
	switch {
	case generates(a, b):
		return "体生用"
	case generates(b, a):
		return "用生体"
	case overcomes(a, b):
		return "体克用"
	case overcomes(b, a):
		return "用克体"
	default:
		return "体用关系不明"
	}
}

// changedRelationText 只看变卦对体卦的生克，不把变卦称为“用卦”。
func changedRelationText(tiElement, changedElement string) string {
	if tiElement == changedElement {
		return "体变比和"
	}
	switch {
	case generates(changedElement, tiElement):
		return "变生体"
	case generates(tiElement, changedElement):
		return "体生变"
	case overcomes(tiElement, changedElement):
		return "体克变"
	case overcomes(changedElement, tiElement):
		return "变克体"
	default:
		return "体变关系不明"
	}
}

// 五行生克。字符串只用于八个卦的五行。
func generates(a, b string) bool {
	switch a {
	case "木":
		return b == "火"
	case "火":
		return b == "土"
	case "土":
		return b == "金"
	case "金":
		return b == "水"
	case "水":
		return b == "木"
	}
	return false
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
