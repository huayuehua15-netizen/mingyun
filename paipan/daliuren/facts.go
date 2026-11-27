package daliuren

import (
	"fmt"
	"strings"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// BuildFacts 从盘面重建完整事实清单，供追问快照恢复使用。
func BuildFacts(p Payload) []schema.Fact {
	return buildFacts(p)
}

func buildFacts(p Payload) []schema.Fact {
	facts := []schema.Fact{
		{
			Key:    "daliuren.month_general",
			Label:  "月将",
			Value:  p.MonthGeneral.String(),
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.day",
			Label:  "日干支",
			Value:  p.DayGan.String() + p.DayZhi.String(),
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.hour",
			Label:  "占时",
			Value:  p.HourZhi.String() + "时",
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.sike",
			Label:  "四课",
			Value:  siKeText(p.SiKe),
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.sanchuan",
			Label:  "三传",
			Value:  chuanText(p.SanChuan),
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.keti",
			Label:  "课体",
			Value:  p.Gate + " · " + p.KeTi,
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.tianjiang",
			Label:  "十二天将",
			Value:  jiangText(p.TianPan),
			Weight: schema.WeightCore,
		},
		{
			Key:    "daliuren.xunkong",
			Label:  "旬空",
			Value:  p.XunKong[0].String() + "、" + p.XunKong[1].String(),
			Weight: schema.WeightCore,
		},
	}

	for i, k := range p.SiKe {
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("daliuren.sike.%d", i+1),
			Label:  k.Name,
			Value:  fmt.Sprintf("%s上%s · %s", k.Lower, k.Upper, k.Relation),
			Weight: schema.WeightNormal,
			Anchor: fmt.Sprintf("/payload/si_ke/%d", i),
		})
	}

	for i, c := range p.SanChuan {
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("daliuren.chuan.%d", i+1),
			Label:  c.Name,
			Value:  fmt.Sprintf("%s（%s宫）· %s · %s", c.Branch, c.Palace, c.LiuQin, c.TianJiang),
			Weight: schema.WeightCore,
			Anchor: fmt.Sprintf("/payload/san_chuan/%d", i),
		})
	}

	for i, tp := range p.TianPan {
		value := fmt.Sprintf("天盘%s · %s · %s", tp.Heaven, tp.LiuQin, tp.TianJiang)
		if len(tp.ShenSha) > 0 {
			value += " · " + strings.Join(tp.ShenSha, "、")
		}
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("daliuren.palace.%d", i+1),
			Label:  "地盘" + tp.Earth.String(),
			Value:  value,
			Weight: schema.WeightNormal,
			Anchor: fmt.Sprintf("/payload/tian_pan/%d", i),
		})
	}

	if len(p.ShenSha) > 0 {
		parts := make([]string, 0, len(p.ShenSha))
		for _, s := range p.ShenSha {
			parts = append(parts, s.Name+"在"+s.Zhi.String())
		}
		facts = append(facts, schema.Fact{
			Key:    "daliuren.shensha",
			Label:  "神煞",
			Value:  strings.Join(parts, "，"),
			Weight: schema.WeightNormal,
		})
	}

	return facts
}

func siKeText(sike []SiKe) string {
	parts := make([]string, 0, len(sike))
	for _, k := range sike {
		parts = append(parts, fmt.Sprintf("%d课%s上%s（%s）", k.Index, k.Lower, k.Upper, k.Relation))
	}
	return strings.Join(parts, "；")
}

func chuanText(chuan []Chuan) string {
	parts := make([]string, 0, len(chuan))
	for _, c := range chuan {
		parts = append(parts, fmt.Sprintf("%s%s（%s，%s，%s）", c.Name, c.Branch, c.Palace, c.LiuQin, c.TianJiang))
	}
	return strings.Join(parts, " → ")
}

func jiangText(tian []TianPanPalace) string {
	parts := make([]string, 0, len(tian))
	for _, tp := range tian {
		parts = append(parts, fmt.Sprintf("%s宫%s·%s", tp.Earth, tp.Heaven, tp.TianJiang))
	}
	return strings.Join(parts, "，")
}
