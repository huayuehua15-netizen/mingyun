package ai

import (
	"fmt"
	"strings"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

const OfflineModel = "offline-rulebook@0.1"

// BuildOfflineReading 在没有模型时生成四段式盘面摘要。
//
// 它明确不判吉凶、不推应期，避免半成品规则冒充专业解读。用户能看到
// 盘面事实与下一步该做什么，而不是一个假的「AI 断卦」。
func BuildOfflineReading(chart schema.Chart, question string) string {
	switch chart.Method {
	case schema.MethodDaLiuRen:
		return buildDaLiuRenOffline(chart, question)
	case schema.MethodZiWei:
		return buildZiWeiOffline(chart, question)
	}
	var b strings.Builder

	b.WriteString("## 盘面要点\n")
	b.WriteString("引擎已完成排盘，以下只转录盘面事实：\n")
	for _, label := range []string{"本卦", "变卦", "世应", "月建日辰", "旬空", "问事类别", "用神"} {
		if v := factValue(chart.Facts, label); v != "" {
			fmt.Fprintf(&b, "· %s：%s\n", label, v)
		}
	}
	moving := movingFacts(chart.Facts)
	if moving == "" {
		b.WriteString("· 动爻：无，六爻皆静。\n")
	} else {
		fmt.Fprintf(&b, "· 动爻：%s\n", moving)
	}

	b.WriteString("\n## 断语\n")
	b.WriteString("离线模式只列盘面，不作吉凶判断。当前没有配置大模型，系统不会伪装成断卦结果；")
	b.WriteString("盘面上的用神取法、旺衰与生克，需要结合所问之事再定。")
	if q := strings.TrimSpace(question); q != "" {
		fmt.Fprintf(&b, "你问的是「%s」，建议先把它拆成可验证的一件事，再看对应爻位。", q)
	}
	b.WriteString("\n")

	b.WriteString("\n## 应期与走向\n")
	b.WriteString("离线模式不推应期。时间判断依赖用神旺衰、动爻与月建日辰的配合，")
	b.WriteString("仅凭当前事实清单给出日期会变成编造。配置模型后可获得这一段的完整解读。\n")

	b.WriteString("\n## 建议\n")
	b.WriteString("1. 先核对盘面事实：卦名、世应、月建日辰、旬空与动爻是否与你的起卦记录一致。\n")
	b.WriteString("2. 把问题收窄到一件具体、可验证的事，避免同时问感情、工作与财运。\n")
	b.WriteString("3. 在服务端配置 MINGYUN_AI_BASE_URL 与 MINGYUN_AI_MODEL 后重新解读。\n")

	return b.String()
}

func buildDaLiuRenOffline(chart schema.Chart, question string) string {
	var b strings.Builder
	b.WriteString("## 盘面要点\n")
	b.WriteString("引擎已完成大六壬起课，以下只转录课盘事实：\n")
	for _, label := range []string{"月将", "日干支", "占时", "四课", "三传", "课体", "十二天将", "旬空", "神煞"} {
		if v := factValue(chart.Facts, label); v != "" {
			fmt.Fprintf(&b, "· %s：%s\n", label, v)
		}
	}
	b.WriteString("\n## 断语\n")
	b.WriteString("离线模式只列课盘，不作吉凶判断。当前没有配置大模型，系统不会伪装成断课结果；")
	b.WriteString("三传的承接、课体与天将的组合，需要结合所问之事再定。")
	if q := strings.TrimSpace(question); q != "" {
		fmt.Fprintf(&b, "你问的是「%s」，建议先把它拆成一件具体、可验证的事。", q)
	}
	b.WriteString("\n\n## 应期与走向\n")
	b.WriteString("离线模式不推应期。时间判断依赖三传、旬空与神煞的配合，仅凭当前事实清单给日期会变成编造。配置模型后可获得这一段解读。\n")
	b.WriteString("\n## 建议\n")
	b.WriteString("1. 先核对月将、日干支、四课与三传是否与起课记录一致。\n")
	b.WriteString("2. 把问题收窄到一件具体、可验证的事，避免同时问多个对象。\n")
	b.WriteString("3. 在服务端配置 MINGYUN_AI_BASE_URL 与 MINGYUN_AI_MODEL 后重新解读。\n")
	return b.String()
}

func buildZiWeiOffline(chart schema.Chart, question string) string {
	var b strings.Builder
	b.WriteString("## 盘面要点\n")
	b.WriteString("引擎已完成紫微斗数排盘，以下只转录命盘事实：\n")
	for _, label := range []string{"生年", "农历生日", "命宫", "身宫", "五行局", "紫微位置", "天府位置", "生年四化", "大限起运", "命宫三方四正"} {
		if v := factValue(chart.Facts, label); v != "" {
			fmt.Fprintf(&b, "· %s：%s\n", label, v)
		}
	}
	b.WriteString("\n## 断语\n")
	b.WriteString("离线模式只列命盘，不作吉凶判断。当前没有配置大模型，系统不会伪装成批命结果；")
	b.WriteString("命宫、三方四正、四化落宫与大限的组合，需要结合所问之事再定。")
	if q := strings.TrimSpace(question); q != "" {
		fmt.Fprintf(&b, "你问的是「%s」，建议先把它拆成一件具体、可验证的事。", q)
	}
	b.WriteString("\n\n## 应期与走向\n")
	b.WriteString("离线模式不推应期。阶段判断依赖大限、宫位与四化的配合，仅凭当前事实清单给年份会变成编造。配置模型后可获得这一段解读。\n")
	b.WriteString("\n## 建议\n")
	b.WriteString("1. 先核对命宫、身宫、五行局、主星与四化是否与出生信息一致。\n")
	b.WriteString("2. 把问题收窄到一件具体、可验证的事，避免同时问事业、感情与财运。\n")
	b.WriteString("3. 在服务端配置 MINGYUN_AI_BASE_URL 与 MINGYUN_AI_MODEL 后重新解读。\n")
	return b.String()
}

func factValue(facts []schema.Fact, label string) string {
	for _, f := range facts {
		if f.Label == label {
			return f.Value
		}
	}
	return ""
}

func movingFacts(facts []schema.Fact) string {
	var parts []string
	for _, f := range facts {
		if strings.Contains(f.Label, "变爻") {
			parts = append(parts, f.Value)
		}
	}
	return strings.Join(parts, "；")
}
