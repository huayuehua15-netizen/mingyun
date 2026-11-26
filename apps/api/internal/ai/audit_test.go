package ai

import (
	"strings"
	"testing"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

func auditFixture() schema.Chart {
	return schema.Chart{
		Method: schema.MethodLiuYao,
		Facts: []schema.Fact{
			{Key: "liuyao.gua", Label: "本卦", Value: "泽火革（坎宫·六世，宫五行水）", Weight: schema.WeightCore},
			{Key: "liuyao.changed_gua", Label: "变卦", Value: "泽雷随（震宫）", Weight: schema.WeightCore},
			{Key: "liuyao.shi_ying", Label: "世应", Value: "世在四爻（丁亥水·兄弟）；应在初爻（己卯木·子孙）", Weight: schema.WeightCore},
			{Key: "liuyao.calendar", Label: "月建日辰", Value: "月建己亥，日辰甲辰", Weight: schema.WeightCore},
			{Key: "liuyao.xunkong", Label: "旬空", Value: "寅、卯", Weight: schema.WeightCore},
			{Key: "liuyao.line.1", Label: "初爻", Value: "老阳｜己卯木｜子孙｜勾陈｜动", Weight: schema.WeightCore},
			{Key: "liuyao.line.3", Label: "三爻", Value: "老阴｜丁丑土｜父母｜腾蛇｜动", Weight: schema.WeightCore},
			{Key: "liuyao.line.4", Label: "四爻", Value: "少阳｜丁亥水｜兄弟｜白虎｜静｜世", Weight: schema.WeightNormal},
		},
	}
}

func TestAuditReadingAcceptsGroundedReading(t *testing.T) {
	text := `## 盘面要点
本卦泽火革，变卦泽雷随。世在四爻，应在初爻；月建己亥、日辰甲辰，旬空寅、卯。初爻子孙发动，三爻父母发动。

## 断语
因为初爻子孙在月建亥水与日辰甲辰之间受制，所以先看它是否能承担所问之事；三爻父母发动，说明事情有推动，但仍需看世爻能否承接。
因为世爻兄弟在四爻，得月建亥水同类相助，所以自身并非全无支撑，只是不宜把判断落在单一的吉凶词上。

## 应期与走向
先等寅、卯出空，再看初爻与三爻的动变是否落地；本月内可先观察，不把日期说死。

## 建议
先把问题收窄到一件可验证的事，再看初爻与三爻的动变。
记录这次起卦的时间与铜钱结果，方便回看。`

	issues := AuditReading(text, auditFixture())
	if len(issues) > 0 {
		t.Fatalf("合规解读不应被拦截: %v", issues)
	}
}

func TestAuditReadingRejectsUngroundedReadings(t *testing.T) {
	cases := []struct {
		name string
		text string
		want string
	}{
		{
			name: "缺少四段",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为世爻在四爻，所以先观察。\n## 应期与走向\n本月内观察。",
			want: "## 建议",
		},
		{
			name: "事实之外的卦名",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为变卦是水火既济，所以事情会转顺。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "水火既济",
		},
		{
			name: "静爻说成动爻",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为四爻发动，所以事情会变化。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "四爻",
		},
		{
			name: "六亲安错爻位",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为官鬼在初爻，所以压力来自下方。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "初爻",
		},
		{
			name: "世应安错",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为世在二爻，所以先看自身。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "世爻",
		},
		{
			name: "绝对化承诺",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n这次必成，不用担心。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "必成",
		},
		{
			name: "事实之外的干支",
			text: "## 盘面要点\n本卦泽火革。\n## 断语\n因为现在是丙午月，所以火旺。\n## 应期与走向\n本月内观察。\n## 建议\n先观察。",
			want: "丙午",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issues := AuditReading(tc.text, auditFixture())
			if len(issues) == 0 {
				t.Fatal("应该拦截，但没有拦截")
			}
			joined := strings.Join(issues, "；")
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("拦截原因应包含 %q，实际为 %s", tc.want, joined)
			}
		})
	}
}

func TestAuditReadingAllowsUncertainty(t *testing.T) {
	text := `## 盘面要点
本卦泽火革，变卦泽雷随。世在四爻，应在初爻。

## 断语
因为盘面只列出这些事实，所以不能判断事情必成，也不能说必败；只能先看世爻与用神关系。

## 应期与走向
盘面不足以稳妥推出应期。

## 建议
先观察，再决定下一步。`
	if issues := AuditReading(text, auditFixture()); len(issues) > 0 {
		t.Fatalf("不确定表述不应被误拦: %v", issues)
	}
}

func TestAuditReadingChecksUseShen(t *testing.T) {
	chart := auditFixture()
	chart.Facts = append(chart.Facts, schema.Fact{
		Key:    "liuyao.use_shen",
		Label:  "用神",
		Value:  "主用神：妻财；次用神：子孙；忌神：兄弟；取法：求财以妻财为用（规则 YS-01）",
		Weight: schema.WeightCore,
	})
	text := `## 盘面要点
本卦泽火革，变卦泽雷随。

## 断语
以官鬼为用神，因为官鬼在三爻，所以压力来自工作。

## 应期与走向
本月内观察。

## 建议
先观察。`
	issues := AuditReading(text, chart)
	if len(issues) == 0 {
		t.Fatal("用神与规则不一致，应拦截")
	}
	if joined := strings.Join(issues, "；"); !strings.Contains(joined, "妻财") {
		t.Fatalf("拦截原因应指出规则用神为妻财: %s", joined)
	}
}

func TestAuditReadingRejectsHardUseShenWhenGenderUnknown(t *testing.T) {
	chart := auditFixture()
	chart.Facts = append(chart.Facts, schema.Fact{
		Key:    "liuyao.use_shen",
		Label:  "用神",
		Value:  "主用神：男占妻财 / 女占官鬼；待确认：求测者性别",
		Weight: schema.WeightCore,
	})
	text := `## 盘面要点
本卦泽火革，变卦泽雷随。

## 断语
以妻财为用神，因为妻财在三爻，所以对方仍然在意。

## 应期与走向
本月内观察。

## 建议
先观察。`
	issues := AuditReading(text, chart)
	if len(issues) == 0 {
		t.Fatal("性别未确认时硬取妻财，应拦截")
	}
	if joined := strings.Join(issues, "；"); !strings.Contains(joined, "待确认") {
		t.Fatalf("拦截原因应指出用神待确认: %s", joined)
	}
}

func TestAuditReadingAllowsDualGenderUseShen(t *testing.T) {
	chart := auditFixture()
	chart.Facts = append(chart.Facts, schema.Fact{
		Key:    "liuyao.use_shen",
		Label:  "用神",
		Value:  "主用神：男占妻财 / 女占官鬼；待确认：求测者性别",
		Weight: schema.WeightCore,
	})
	text := `## 盘面要点
本卦泽火革，变卦泽雷随。

## 断语
男占以妻财为用神，女占以官鬼为用神；性别未确认前不合并断。

## 应期与走向
本月内观察。

## 建议
先确认求测者性别，再继续。`
	if issues := AuditReading(text, chart); len(issues) > 0 {
		t.Fatalf("并列说明两种取法不应被误拦: %v", issues)
	}
}
