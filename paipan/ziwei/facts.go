package ziwei

import (
	"fmt"
	"strings"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// BuildFacts 从盘面重建完整事实清单，供追问快照恢复使用。
func BuildFacts(p Payload) []schema.Fact {
	return buildFacts(p)
}

func buildFacts(p Payload) []schema.Fact {
	facts := []schema.Fact{
		{Key: "ziwei.year", Label: "生年", Value: p.YearGanZhi, Weight: schema.WeightCore},
		{Key: "ziwei.lunar", Label: "农历生日", Value: p.LunarDate, Weight: schema.WeightCore},
		{Key: "ziwei.ming_gong", Label: "命宫", Value: palaceFact(p, p.MingGong), Weight: schema.WeightCore},
		{Key: "ziwei.shen_gong", Label: "身宫", Value: palaceFact(p, p.ShenGong), Weight: schema.WeightCore},
		{Key: "ziwei.wu_xing_ju", Label: "五行局", Value: p.WuXingJu, Weight: schema.WeightCore},
		{Key: "ziwei.ziwei_pos", Label: "紫微位置", Value: p.ZiWeiPos.String() + "宫", Weight: schema.WeightCore},
		{Key: "ziwei.tianfu_pos", Label: "天府位置", Value: p.TianFuPos.String() + "宫", Weight: schema.WeightCore},
		{Key: "ziwei.sihua", Label: "生年四化", Value: sihuaText(p.SihuaYear), Weight: schema.WeightCore},
		{Key: "ziwei.daxian", Label: "大限起运", Value: daxianText(p.DaXian), Weight: schema.WeightCore},
		{Key: "ziwei.sanfang", Label: "命宫三方四正", Value: sanfangText(p), Weight: schema.WeightCore},
	}

	for i, palace := range p.Palaces {
		value := fmt.Sprintf("%s · 主星 %s · 辅星 %s", palace.StemBranch, starText(palace.MajorStars), starText(palace.MinorStars))
		if palace.IsShen {
			value += " · 身宫"
		}
		if palace.DaxianRange != "" {
			value += " · 大限" + palace.DaxianRange
		}
		facts = append(facts, schema.Fact{
			Key:    fmt.Sprintf("ziwei.palace.%d", i),
			Label:  palace.Name + "宫",
			Value:  value,
			Weight: schema.WeightNormal,
			Anchor: fmt.Sprintf("/payload/palaces/%d", i),
		})
	}

	return facts
}

func palaceFact(p Payload, branch calendar.Zhi) string {
	for _, palace := range p.Palaces {
		if palace.Branch == branch {
			value := palace.Name + " · " + palace.StemBranch
			if len(palace.MajorStars) > 0 {
				value += " · " + starText(palace.MajorStars)
			}
			return value
		}
	}
	return branch.String()
}

func starText(stars []Star) string {
	if len(stars) == 0 {
		return "无"
	}
	parts := make([]string, 0, len(stars))
	for _, star := range stars {
		text := star.Name
		if star.Hua != "" {
			text += "化" + star.Hua
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "、")
}

func sihuaText(sihua []Sihua) string {
	parts := make([]string, 0, len(sihua))
	for _, s := range sihua {
		parts = append(parts, s.Star+"化"+s.Hua)
	}
	return strings.Join(parts, "、")
}

func daxianText(dx []DaXian) string {
	if len(dx) == 0 {
		return "未定"
	}
	return fmt.Sprintf("%s起，%d—%d岁；%s", dx[0].Name, dx[0].StartAge, dx[0].EndAge, dx[1].Range)
}

func sanfangText(p Payload) string {
	ming := int(p.MingGong)
	branches := []calendar.Zhi{
		p.MingGong,
		calendar.Zhi(calendar.Mod(ming+4, 12)),
		calendar.Zhi(calendar.Mod(ming+8, 12)),
		calendar.Zhi(calendar.Mod(ming+6, 12)),
	}
	parts := make([]string, 0, len(branches))
	for _, branch := range branches {
		for _, palace := range p.Palaces {
			if palace.Branch == branch {
				parts = append(parts, palace.Name+"("+branch.String()+")")
				break
			}
		}
	}
	return strings.Join(parts, "、")
}
