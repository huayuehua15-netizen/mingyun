package liuyao

import (
	"strings"
	"testing"
)

func testUseShenLines() []LinePayload {
	return []LinePayload{
		{Position: 1, LiuQin: "子孙", Shi: false, Ying: true},
		{Position: 2, LiuQin: "妻财"},
		{Position: 3, LiuQin: "父母"},
		{Position: 4, LiuQin: "兄弟", Shi: true},
		{Position: 5, LiuQin: "官鬼"},
		{Position: 6, LiuQin: "妻财"},
	}
}

func TestSelectUseShenByTopic(t *testing.T) {
	cases := []struct {
		name      string
		question  string
		override  string
		gender    string
		wantTopic Topic
		wantFirst string
		wantConf  string
	}{
		{name: "考试", question: "这次考试能否通过", wantTopic: TopicExam, wantFirst: "父母", wantConf: "definite"},
		{name: "求财", question: "这笔生意能赚钱吗", wantTopic: TopicWealth, wantFirst: "妻财", wantConf: "definite"},
		{name: "事业", question: "什么时候能升职", wantTopic: TopicCareer, wantFirst: "官鬼", wantConf: "definite"},
		{name: "健康", question: "父亲的病什么时候好", wantTopic: TopicHealth, wantFirst: "世爻", wantConf: "definite"},
		{name: "失物", question: "身份证丢了能找回吗", wantTopic: TopicLost, wantFirst: "妻财", wantConf: "definite"},
		{name: "感情未标性别", question: "这段感情还能复合吗", wantTopic: TopicRelationship, wantConf: "needs_gender"},
		{name: "男占感情", question: "这段感情还能复合吗", gender: "male", wantTopic: TopicRelationship, wantFirst: "妻财", wantConf: "definite"},
		{name: "女占感情", question: "这段感情还能复合吗", gender: "female", wantTopic: TopicRelationship, wantFirst: "官鬼", wantConf: "definite"},
		{name: "六亲爸爸", question: "爸爸今年运势如何", wantTopic: TopicFamily, wantFirst: "父母", wantConf: "definite"},
		{name: "未知", question: "最近运势如何", wantTopic: TopicGeneral, wantConf: "open"},
		{name: "显式覆盖", question: "这次考试能否通过", override: "wealth", wantTopic: TopicWealth, wantFirst: "妻财", wantConf: "definite"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := SelectUseShen(tc.question, tc.override, tc.gender, testUseShenLines(), nil)
			if plan.Topic != tc.wantTopic {
				t.Fatalf("topic = %s，期望 %s", plan.Topic, tc.wantTopic)
			}
			if tc.wantFirst != "" && (len(plan.Primary) == 0 || plan.Primary[0] != tc.wantFirst) {
				t.Fatalf("主用神 = %v，期望首项 %s", plan.Primary, tc.wantFirst)
			}
			if plan.Confidence != tc.wantConf {
				t.Fatalf("confidence = %s，期望 %s", plan.Confidence, tc.wantConf)
			}
			if !strings.Contains(plan.Rationale, "规则") && plan.RuleID == "" {
				t.Fatal("取用神结果缺少规则标识")
			}
		})
	}
}

func TestUseShenCandidatesAndFactValue(t *testing.T) {
	plan := SelectUseShen("这笔生意能赚钱吗", "", "", testUseShenLines(), nil)
	if len(plan.Candidates) == 0 {
		t.Fatal("缺少六亲候选")
	}
	var wifePositions []int
	for _, candidate := range plan.Candidates {
		if candidate.Name == "妻财" {
			wifePositions = candidate.Positions
		}
	}
	if len(wifePositions) != 2 || wifePositions[0] != 2 || wifePositions[1] != 6 {
		t.Fatalf("妻财爻位 = %v，期望 [2 6]", wifePositions)
	}
	fact := plan.FactValue()
	for _, want := range []string{"主用神：妻财", "子孙", "兄弟", "规则 YS-01"} {
		if !strings.Contains(fact, want) {
			t.Fatalf("用神事实缺少 %q：%s", want, fact)
		}
	}
}

func TestRelationshipNeedsGenderFact(t *testing.T) {
	plan := SelectUseShen("这段感情还能复合吗", "", "", testUseShenLines(), nil)
	fact := plan.FactValue()
	if !strings.Contains(fact, "男占妻财") || !strings.Contains(fact, "女占官鬼") {
		t.Fatalf("未标性别的感情盘应并列两种取法：%s", fact)
	}
	if !strings.Contains(fact, "待确认：求测者性别") {
		t.Fatalf("应提示待确认性别：%s", fact)
	}
}

func TestUseShenMissingOnChart(t *testing.T) {
	lines := testUseShenLines()
	for i := range lines {
		if lines[i].LiuQin == "父母" {
			lines[i].LiuQin = "兄弟"
		}
	}
	plan := SelectUseShen("考试能否通过", "", "", lines, nil)
	fact := plan.FactValue()
	if !strings.Contains(fact, "用神不现：父母") {
		t.Fatalf("父母不上卦时应明示，而不是默认可断：%s", fact)
	}
}

func TestUseShenCanBeHiddenAsFuShen(t *testing.T) {
	lines := testUseShenLines()
	for i := range lines {
		if lines[i].LiuQin == "父母" {
			lines[i].LiuQin = "兄弟"
		}
	}
	fu := []FuShenPayload{{UnderPosition: 3, LiuQin: "父母"}}
	plan := SelectUseShen("考试能否通过", "", "", lines, fu)
	fact := plan.FactValue()
	if !strings.Contains(fact, "用神伏藏：父母伏于三爻下") {
		t.Fatalf("父母伏藏时应标伏神，而不是当成完全不上卦：%s", fact)
	}
	if strings.Contains(fact, "用神不现：父母") {
		t.Fatalf("有伏神时不应同时说用神不现：%s", fact)
	}
}
