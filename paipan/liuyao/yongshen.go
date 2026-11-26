package liuyao

import "strings"

// Topic 问事类别。用神取法先按类别落规则，不交给模型临时发挥。
type Topic string

const (
	TopicGeneral      Topic = "general"
	TopicExam         Topic = "exam"
	TopicWealth       Topic = "wealth"
	TopicCareer       Topic = "career"
	TopicRelationship Topic = "relationship"
	TopicHealth       Topic = "health"
	TopicTravel       Topic = "travel"
	TopicLost         Topic = "lost"
	TopicLegal        Topic = "legal"
	TopicFamily       Topic = "family"
	TopicSurname      Topic = "surname"
)

// UseShenCandidate 六亲或世应落在盘中的哪几爻；FuPositions 为伏神所在爻。
type UseShenCandidate struct {
	Name        string `json:"name"`
	Positions   []int  `json:"positions,omitempty"`
	FuPositions []int  `json:"fu_positions,omitempty"`
}

// SurnameHint 是测姓字谜类问题的地支、生肖与常见姓氏映射。
// 生肖只作线索；Surname 为空表示该生肖没有可直接落到的常见姓氏。
type SurnameHint struct {
	Position int    `json:"position"`
	Zhi      string `json:"zhi"`
	Animal   string `json:"animal"`
	Surname  string `json:"surname,omitempty"`
	IsShi    bool   `json:"is_shi,omitempty"`
	IsYing   bool   `json:"is_ying,omitempty"`
}

// UseShenPlan 是规则引擎产出的取用神结果，不是模型自行判断的结果。
type UseShenPlan struct {
	Topic        Topic              `json:"topic"`
	TopicLabel   string             `json:"topic_label"`
	RuleID       string             `json:"rule_id"`
	Confidence   string             `json:"confidence"` // definite | needs_gender | open
	Primary      []string           `json:"primary,omitempty"`
	Secondary    []string           `json:"secondary,omitempty"`
	Taboo        []string           `json:"taboo,omitempty"`
	Rationale    string             `json:"rationale"`
	Candidates   []UseShenCandidate `json:"candidates,omitempty"`
	SurnameHints []SurnameHint      `json:"surname_hints,omitempty"`
}

type topicRule struct {
	Topic     Topic
	Label     string
	RuleID    string
	Keywords  []string
	Primary   []string
	Secondary []string
	Taboo     []string
	Rationale string
}

var topicRules = []topicRule{
	{
		Topic: TopicExam, Label: "考试学业", RuleID: "YS-03",
		Keywords:  []string{"考试", "考学", "成绩", "分数", "录取", "升学", "考研", "高考", "中考", "证书", "论文", "答辩", "笔试"},
		Primary:   []string{"父母"},
		Secondary: []string{"官鬼"},
		Taboo:     []string{"子孙"},
		Rationale: "占考试以父母为文书成绩，官鬼为名次与录取；子孙克官，不作主用。",
	},
	{
		Topic: TopicWealth, Label: "求财经营", RuleID: "YS-01",
		Keywords:  []string{"求财", "财运", "赚钱", "收入", "投资", "生意", "买卖", "盈利", "工资", "奖金", "回款", "欠款", "债务", "股票", "基金", "赚", "借钱", "借款", "还钱", "还款", "贷款", "亏", "赔", "收益", "分红"},
		Primary:   []string{"妻财"},
		Secondary: []string{"子孙"},
		Taboo:     []string{"兄弟"},
		Rationale: "占财以妻财为用，子孙为财源，兄弟为劫财。",
	},
	{
		Topic: TopicCareer, Label: "事业官运", RuleID: "YS-02",
		Keywords:  []string{"工作", "事业", "升职", "升迁", "职位", "官运", "求职", "面试", "跳槽", "创业", "录用", "岗位", "职称", "编制", "晋升"},
		Primary:   []string{"官鬼"},
		Secondary: []string{"父母"},
		Taboo:     []string{"子孙"},
		Rationale: "占事业以官鬼为职位与权责，父母为文书与单位；子孙克官，不作主用。",
	},
	{
		Topic: TopicRelationship, Label: "感情婚姻", RuleID: "YS-04",
		Keywords:  []string{"感情", "婚姻", "婚恋", "恋爱", "对象", "男朋友", "女朋友", "男友", "女友", "老公", "老婆", "丈夫", "妻子", "分手", "复合", "结婚", "暗恋", "相亲", "暧昧", "伴侣", "配偶", "前男友", "前女友"},
		Primary:   []string{"世爻"},
		Secondary: []string{"应爻"},
		Rationale: "感情以世应为双方，男占以妻财为对方，女占以官鬼为对方；未标性别时不能合并取一个主用神。",
	},
	{
		Topic: TopicHealth, Label: "疾病健康", RuleID: "YS-05",
		Keywords:  []string{"健康", "生病", "病", "手术", "医药", "医生", "康复", "疼痛", "怀孕", "检查", "治疗"},
		Primary:   []string{"世爻"},
		Secondary: []string{"官鬼", "子孙"},
		Rationale: "占病以世爻为本人，官鬼为病，子孙为医药；先看世爻旺衰，再取病与药。",
	},
	{
		Topic: TopicTravel, Label: "出行远行", RuleID: "YS-06",
		Keywords:  []string{"出行", "旅行", "出差", "搬家", "航班", "车票", "远行", "动身", "旅游", "行程", "签证", "机票"},
		Primary:   []string{"世爻"},
		Secondary: []string{"父母"},
		Taboo:     []string{"官鬼"},
		Rationale: "出行以世爻为行人，父母为车船行李，官鬼为阻碍。",
	},
	{
		Topic: TopicLost, Label: "失物寻找", RuleID: "YS-07",
		Keywords:  []string{"失物", "丢失", "丢了", "找不到", "遗失", "被盗", "寻物", "找东西", "钥匙", "钱包", "手机"},
		Primary:   []string{"妻财"},
		Secondary: []string{"子孙"},
		Taboo:     []string{"官鬼"},
		Rationale: "失物以妻财为物，子孙为找回，官鬼为盗失。",
	},
	{
		Topic: TopicLegal, Label: "官司诉讼", RuleID: "YS-08",
		Keywords:  []string{"官司", "诉讼", "起诉", "仲裁", "合同", "纠纷", "法院", "律师", "报警", "打官司", "索赔"},
		Primary:   []string{"官鬼"},
		Secondary: []string{"父母", "子孙"},
		Rationale: "官司以官鬼为官非，父母为文书，子孙为解忧；先看官鬼旺衰，再看文书与解神。",
	},
	{
		Topic: TopicFamily, Label: "六亲家人", RuleID: "YS-09",
		Keywords:  []string{"父母", "父亲", "母亲", "子女", "孩子", "儿子", "女儿", "兄弟", "姐妹", "家人", "亲戚", "长辈", "爸爸", "妈妈", "爷爷", "奶奶", "姥姥", "姥爷", "岳父", "岳母", "公公", "婆婆"},
		Primary:   []string{"世爻"},
		Secondary: []string{"应爻"},
		Rationale: "六亲问事先按所指对象取六亲；未指明对象时不可硬取，先看世应。",
	},
	{
		Topic: TopicSurname, Label: "测姓字谜", RuleID: "YS-10",
		Keywords:  []string{"姓氏", "测姓", "姓什么", "我姓", "猜姓", "字谜", "测字"},
		Primary:   []string{"世爻"},
		Secondary: []string{"应爻"},
		Rationale: "测自身姓氏以世爻为主，问他人姓氏以应爻为主；生肖只作取象线索，答案必须落到真实姓氏。",
	},
}

var validTopics = map[Topic]bool{
	TopicGeneral: true, TopicExam: true, TopicWealth: true, TopicCareer: true,
	TopicRelationship: true, TopicHealth: true, TopicTravel: true, TopicLost: true,
	TopicLegal: true, TopicFamily: true, TopicSurname: true,
}

// SelectUseShen 按问题类别和性别取用神。它是确定性规则，模型不得越过。
// topicOverride 为空时才做关键词分类；gender 只接受 male/female（或男/女）。
func SelectUseShen(question, topicOverride, gender string, lines []LinePayload, fuShen []FuShenPayload) UseShenPlan {
	topic, matched := classifyTopic(question, topicOverride)
	plan := buildUseShenPlan(topic, question, normalizeGender(gender))
	plan.Topic = topic
	plan.TopicLabel = topicLabel(topic)
	if matched != "" {
		plan.Rationale += "（命中关键词：" + matched + "）"
	}
	plan.Candidates = useShenCandidates(lines, fuShen)
	if topic == TopicSurname {
		plan.SurnameHints = surnameHints(lines)
	}
	return plan
}

func classifyTopic(question, override string) (Topic, string) {
	override = strings.TrimSpace(override)
	if t := Topic(override); validTopics[t] {
		return t, ""
	}
	for _, rule := range topicRules {
		for _, kw := range rule.Keywords {
			if strings.Contains(question, kw) {
				return rule.Topic, kw
			}
		}
	}
	return TopicGeneral, ""
}

func buildUseShenPlan(topic Topic, question, gender string) UseShenPlan {
	switch topic {
	case TopicExam, TopicWealth, TopicCareer, TopicHealth, TopicTravel, TopicLost, TopicLegal:
		rule := findTopicRule(topic)
		return UseShenPlan{
			Topic: topic, TopicLabel: rule.Label, RuleID: rule.RuleID, Confidence: "definite",
			Primary: rule.Primary, Secondary: rule.Secondary, Taboo: rule.Taboo, Rationale: rule.Rationale,
		}
	case TopicRelationship:
		rule := findTopicRule(topic)
		if gender == "male" {
			return UseShenPlan{
				Topic: topic, TopicLabel: rule.Label, RuleID: "YS-04M", Confidence: "definite",
				Primary: []string{"妻财"}, Secondary: []string{"世爻", "应爻"},
				Rationale: "男占感情以妻财为对方，世应看双方关系。",
			}
		}
		if gender == "female" {
			return UseShenPlan{
				Topic: topic, TopicLabel: rule.Label, RuleID: "YS-04F", Confidence: "definite",
				Primary: []string{"官鬼"}, Secondary: []string{"世爻", "应爻"},
				Rationale: "女占感情以官鬼为对方，世应看双方关系。",
			}
		}
		return UseShenPlan{
			Topic: topic, TopicLabel: rule.Label, RuleID: "YS-04?", Confidence: "needs_gender",
			Primary: []string{"妻财", "官鬼"}, Secondary: []string{"世爻", "应爻"},
			Rationale: "男占以妻财为对方，女占以官鬼为对方；未标性别时只能并列，不能合并取一个主用神。",
		}
	case TopicFamily:
		return familyUseShen(question)
	case TopicSurname:
		return surnameUseShen(question)
	default:
		return UseShenPlan{
			Topic: TopicGeneral, TopicLabel: "未定", RuleID: "YS-00", Confidence: "open",
			Secondary: []string{"世爻", "应爻"},
			Rationale: "未命中明确问事类别：先看世爻与应爻，主用神不硬取；需补充所问对象后再定。",
		}
	}
}

func familyUseShen(question string) UseShenPlan {
	primary := ""
	if containsAny(question, "父母", "父亲", "母亲", "爸爸", "妈妈", "爷爷", "奶奶", "姥姥", "姥爷", "岳父", "岳母", "公公", "婆婆", "长辈") {
		primary = "父母"
	} else if containsAny(question, "子女", "孩子", "儿子", "女儿") {
		primary = "子孙"
	} else if containsAny(question, "兄弟", "姐妹", "哥哥", "弟弟", "姐姐", "妹妹", "表哥", "表姐") {
		primary = "兄弟"
	}
	if primary == "" {
		return UseShenPlan{
			Topic: TopicFamily, TopicLabel: "六亲家人", RuleID: "YS-09", Confidence: "open",
			Secondary: []string{"世爻", "应爻"},
			Rationale: "问六亲但未指明对象，先看世应，不硬取用神。",
		}
	}
	return UseShenPlan{
		Topic: TopicFamily, TopicLabel: "六亲家人", RuleID: "YS-09", Confidence: "definite",
		Primary: []string{primary}, Secondary: []string{"世爻"},
		Rationale: "六亲问事以所指对象为用神；" + primary + "为所问之人，世爻为求测者。",
	}
}

func surnameUseShen(question string) UseShenPlan {
	primary, secondary := "世爻", "应爻"
	rationale := "测自身姓氏以世爻为主，应爻为线索；生肖只作取象，答案须落到真实姓氏。"
	if asksOtherSurname(question) {
		primary, secondary = "应爻", "世爻"
		rationale = "问他人姓氏以应爻为主，世爻为求测者；生肖只作取象，答案须落到真实姓氏。"
	}
	return UseShenPlan{
		Topic: TopicSurname, TopicLabel: "测姓字谜", RuleID: "YS-10", Confidence: "definite",
		Primary: []string{primary}, Secondary: []string{secondary}, Rationale: rationale,
	}
}

func asksOtherSurname(question string) bool {
	return containsAny(question, "他", "她", "对方", "别人", "朋友", "孩子", "爸爸", "妈妈", "父亲", "母亲", "爷爷", "奶奶", "家人", "亲戚")
}

// branchSurnameHints 只列常见、可直接落到的姓氏；没有常见姓氏对应的生肖
// 只保留为线索，绝不把动物本身当成姓。
var branchSurnameHints = map[string]struct{ Animal, Surname string }{
	"子": {Animal: "鼠"},
	"丑": {Animal: "牛", Surname: "牛"},
	"寅": {Animal: "虎"},
	"卯": {Animal: "兔"},
	"辰": {Animal: "龙", Surname: "龙"},
	"巳": {Animal: "蛇", Surname: "佘"},
	"午": {Animal: "马", Surname: "马"},
	"未": {Animal: "羊", Surname: "杨"},
	"申": {Animal: "猴", Surname: "侯"},
	"酉": {Animal: "鸡", Surname: "姬"},
	"戌": {Animal: "狗", Surname: "苟"},
	"亥": {Animal: "猪", Surname: "朱"},
}

func surnameHints(lines []LinePayload) []SurnameHint {
	out := make([]SurnameHint, 0, len(lines))
	for _, line := range lines {
		info, ok := branchSurnameHints[line.Zhi]
		if !ok {
			continue
		}
		out = append(out, SurnameHint{
			Position: line.Position,
			Zhi:      line.Zhi,
			Animal:   info.Animal,
			Surname:  info.Surname,
			IsShi:    line.Shi,
			IsYing:   line.Ying,
		})
	}
	return out
}

func findTopicRule(topic Topic) topicRule {
	for _, rule := range topicRules {
		if rule.Topic == topic {
			return rule
		}
	}
	return topicRule{Topic: TopicGeneral, Label: "未定", RuleID: "YS-00"}
}

func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func topicLabel(topic Topic) string {
	for _, rule := range topicRules {
		if rule.Topic == topic {
			return rule.Label
		}
	}
	return "未定"
}

func normalizeGender(gender string) string {
	gender = strings.ToLower(strings.TrimSpace(gender))
	switch {
	case strings.Contains(gender, "男") || gender == "male" || gender == "m":
		return "male"
	case strings.Contains(gender, "女") || gender == "female" || gender == "f":
		return "female"
	default:
		return ""
	}
}

func useShenCandidates(lines []LinePayload, fuShen []FuShenPayload) []UseShenCandidate {
	names := []string{"父母", "兄弟", "子孙", "妻财", "官鬼", "世爻", "应爻"}
	out := make([]UseShenCandidate, 0, len(names))
	for _, name := range names {
		positions := make([]int, 0, 6)
		fuPositions := make([]int, 0, 6)
		for _, line := range lines {
			switch name {
			case "世爻":
				if line.Shi {
					positions = append(positions, line.Position)
				}
			case "应爻":
				if line.Ying {
					positions = append(positions, line.Position)
				}
			default:
				if line.LiuQin == name {
					positions = append(positions, line.Position)
				}
			}
		}
		if name != "世爻" && name != "应爻" {
			for _, fu := range fuShen {
				if fu.LiuQin == name {
					fuPositions = append(fuPositions, fu.UnderPosition)
				}
			}
		}
		candidate := UseShenCandidate{Name: name, Positions: positions}
		if len(fuPositions) > 0 {
			candidate.FuPositions = fuPositions
		}
		out = append(out, candidate)
	}
	return out
}

// FactValue 把规则结果压成给模型和前端看的一条事实。
func (p UseShenPlan) FactValue() string {
	var parts []string
	if p.Confidence == "needs_gender" && len(p.Primary) == 2 {
		parts = append(parts, "主用神：男占"+p.Primary[0]+" / 女占"+p.Primary[1])
	} else if len(p.Primary) == 0 {
		parts = append(parts, "主用神：待定")
	} else {
		parts = append(parts, "主用神："+strings.Join(p.Primary, "、"))
	}
	if len(p.Secondary) > 0 {
		parts = append(parts, "次用神："+strings.Join(p.Secondary, "、"))
	}
	if len(p.Taboo) > 0 {
		parts = append(parts, "忌神："+strings.Join(p.Taboo, "、"))
	}
	positions := make([]string, 0, len(p.Primary))
	hidden := make([]string, 0, len(p.Primary))
	missing := make([]string, 0, len(p.Primary))
	for _, name := range p.Primary {
		found := false
		for _, candidate := range p.Candidates {
			if candidate.Name != name {
				continue
			}
			if len(candidate.Positions) > 0 {
				found = true
				positions = append(positions, name+"在"+positionText(candidate.Positions))
			}
			if len(candidate.FuPositions) > 0 {
				found = true
				hidden = append(hidden, name+"伏于"+positionText(candidate.FuPositions)+"下")
			}
		}
		if !found {
			missing = append(missing, name)
		}
	}
	if len(positions) > 0 {
		parts = append(parts, "用神爻位："+strings.Join(positions, "；"))
	}
	if len(hidden) > 0 {
		parts = append(parts, "用神伏藏："+strings.Join(hidden, "；"))
	}
	if len(missing) > 0 {
		parts = append(parts, "用神不现："+strings.Join(missing, "、"))
	}
	parts = append(parts, "取法："+p.Rationale+"（规则 "+p.RuleID+"）")
	if p.Confidence == "needs_gender" {
		parts = append(parts, "待确认：求测者性别")
	}
	if p.Confidence == "open" {
		parts = append(parts, "待确认：所问对象")
	}
	return strings.Join(parts, "；")
}

func positionText(positions []int) string {
	names := []string{"初爻", "二爻", "三爻", "四爻", "五爻", "上爻"}
	out := make([]string, 0, len(positions))
	for _, pos := range positions {
		if pos >= 1 && pos <= len(names) {
			out = append(out, names[pos-1])
		}
	}
	return strings.Join(out, "、")
}

// SurnameFactValue 把测姓取象压成一条可供模型引用、也可供审计解析的事实。
// 只列主用神所在爻的真实姓氏候选；没有常见姓氏对应的生肖只作线索，避免“姓兔”。
func (p UseShenPlan) SurnameFactValue() string {
	if len(p.SurnameHints) == 0 {
		return ""
	}
	selected := p.SurnameHints
	if len(p.Primary) > 0 {
		switch p.Primary[0] {
		case "世爻":
			selected = filterSurnameHints(p.SurnameHints, func(h SurnameHint) bool { return h.IsShi })
		case "应爻":
			selected = filterSurnameHints(p.SurnameHints, func(h SurnameHint) bool { return h.IsYing })
		}
	}
	if len(selected) == 0 {
		return "主用神所在爻没有常见姓氏候选；生肖只作线索，不能当姓"
	}

	valid := make([]string, 0, len(selected))
	clues := make([]string, 0, len(selected))
	for _, h := range selected {
		where := positionText([]int{h.Position}) + "·" + h.Zhi + h.Animal
		if h.Surname != "" {
			valid = append(valid, h.Surname+"（"+where+"）")
		} else {
			clues = append(clues, h.Animal+"（"+where+"）")
		}
	}
	if len(valid) == 0 {
		return "主用神所在爻只有生肖线索：" + strings.Join(clues, "、") + "，不能把动物本身当姓"
	}
	value := "常见姓氏候选：" + strings.Join(valid, "、")
	if len(clues) > 0 {
		value += "；同爻生肖线索：" + strings.Join(clues, "、") + "，不能当姓"
	}
	value += "；生肖只作线索，不能当姓"
	return value
}

func filterSurnameHints(hints []SurnameHint, keep func(SurnameHint) bool) []SurnameHint {
	out := make([]SurnameHint, 0, len(hints))
	for _, hint := range hints {
		if keep(hint) {
			out = append(out, hint)
		}
	}
	return out
}
