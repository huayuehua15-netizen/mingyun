package liuyao

import (
	"fmt"
	"strings"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

var yaoNames = [6]string{"初", "二", "三", "四", "五", "上"}

func yaoLabel(pos int) string {
	if pos < 1 || pos > 6 {
		return "?"
	}
	return yaoNames[pos-1] + "爻"
}

// buildFacts 把盘面压成给 AI 的事实清单。
// 事实要短、要准、要能指回 payload；模型只依据事实说话，不允许自己推演。
func buildFacts(p Payload) []schema.Fact {
	facts := make([]schema.Fact, 0, 16)

	facts = append(facts, schema.Fact{
		Key:    "liuyao.gua",
		Label:  "本卦",
		Value:  fmt.Sprintf("%s（%s·%s，宫五行%s）", p.GuaName, p.Palace, p.Rank, p.PalaceWuXing),
		Anchor: "/payload/gua_name",
		Weight: schema.WeightCore,
	})

	if p.ChangedGua != "" {
		facts = append(facts, schema.Fact{
			Key:    "liuyao.changed_gua",
			Label:  "变卦",
			Value:  fmt.Sprintf("%s（%s）", p.ChangedGua, p.ChangedPalace),
			Anchor: "/payload/changed_gua",
			Weight: schema.WeightCore,
		})
	}

	shi, ying := findLine(p.Lines, p.ShiYao), findLine(p.Lines, p.YingYao)
	if shi != nil && ying != nil {
		facts = append(facts, schema.Fact{
			Key:   "liuyao.shi_ying",
			Label: "世应",
			Value: fmt.Sprintf("世在%s（%s%s%s·%s）；应在%s（%s%s%s·%s）",
				yaoLabel(shi.Position), shi.Gan, shi.Zhi, shi.WuXing, shi.LiuQin,
				yaoLabel(ying.Position), ying.Gan, ying.Zhi, ying.WuXing, ying.LiuQin),
			Anchor: "/payload/shi_yao",
			Weight: schema.WeightCore,
		})
	}

	facts = append(facts, schema.Fact{
		Key:    "liuyao.calendar",
		Label:  "月建日辰",
		Value:  fmt.Sprintf("月建%s，日辰%s", p.MonthGanZhi, p.DayGanZhi),
		Anchor: "/payload/month_zhi",
		Weight: schema.WeightCore,
	})

	facts = append(facts, schema.Fact{
		Key:    "liuyao.xunkong",
		Label:  "旬空",
		Value:  fmt.Sprintf("%s、%s", p.XunKong[0], p.XunKong[1]),
		Anchor: "/payload/xun_kong",
		Weight: schema.WeightCore,
	})

	if p.UseShen != nil {
		facts = append(facts, schema.Fact{
			Key:    "liuyao.topic",
			Label:  "问事类别",
			Value:  p.UseShen.TopicLabel,
			Anchor: "/payload/use_shen/topic",
			Weight: schema.WeightCore,
		})
		facts = append(facts, schema.Fact{
			Key:    "liuyao.use_shen",
			Label:  "用神",
			Value:  p.UseShen.FactValue(),
			Anchor: "/payload/use_shen",
			Weight: schema.WeightCore,
		})
	}

	for i, l := range p.Lines {
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("liuyao.line.%d", l.Position),
			Label:  yaoLabel(l.Position),
			Value:  lineValue(l),
			Anchor: fmt.Sprintf("/payload/lines/%d", i),
			Weight: lineWeight(l),
		})
	}

	for i, l := range p.Lines {
		if l.Change == nil {
			continue
		}
		value := fmt.Sprintf("%s变%s%s%s·%s", yaoLabel(l.Position), l.Change.Gan, l.Change.Zhi, l.Change.WuXing, l.Change.LiuQin)
		if l.Change.Relation != "" {
			value += "，" + l.Change.Relation
		}
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("liuyao.change.%d", l.Position),
			Label:  yaoLabel(l.Position) + "变爻",
			Value:  value,
			Anchor: fmt.Sprintf("/payload/lines/%d/change", i),
			Weight: schema.WeightCore,
		})
	}

	for i, f := range p.FuShen {
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("liuyao.fushen.%d", i+1),
			Label:  "伏神",
			Value:  fmt.Sprintf("%s%s%s伏于%s下", f.LiuQin, f.Gan, f.Zhi, yaoLabel(f.UnderPosition)),
			Anchor: fmt.Sprintf("/payload/fu_shen/%d", i),
			Weight: schema.WeightMinor,
		})
	}

	return facts
}

func findLine(lines []LinePayload, pos int) *LinePayload {
	for i := range lines {
		if lines[i].Position == pos {
			return &lines[i]
		}
	}
	return nil
}

func lineValue(l LinePayload) string {
	parts := []string{l.Kind, l.Gan + l.Zhi + l.WuXing, l.LiuQin, l.LiuShen}
	if l.Moving {
		parts = append(parts, "动")
	} else {
		parts = append(parts, "静")
	}
	if l.Shi {
		parts = append(parts, "世")
	}
	if l.Ying {
		parts = append(parts, "应")
	}
	if len(l.Marks) > 0 {
		parts = append(parts, strings.Join(l.Marks, "、"))
	}
	return strings.Join(parts, "｜")
}

// lineWeight 动爻是断卦关键，给核心权重；静爻可被裁剪。
func lineWeight(l LinePayload) schema.Weight {
	if l.Moving {
		return schema.WeightCore
	}
	return schema.WeightNormal
}
