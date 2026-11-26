package ai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/huayuehua15-netizen/mingyun/paipan/liuyao"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// AuditReading 是模型输出之后的硬校验。提示词负责约束，校验负责拦住
// 仍然越界的内容：缺段、绝对化承诺、事实之外的卦名/干支、把静爻说成动爻、
// 把六亲或世应安错爻位。校验不通过时，调用方不应把正文发给用户。
func AuditReading(text string, chart schema.Chart) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if !strings.Contains(text, "## 盘面要点") && strings.TrimSpace(text) == "" {
		return []string{"输出为空"}
	}

	var issues []string
	seen := make(map[string]bool)
	add := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		if !seen[msg] {
			seen[msg] = true
			issues = append(issues, msg)
		}
	}

	cursor := 0
	for _, heading := range requiredHeadings {
		idx := strings.Index(text[cursor:], heading)
		if idx < 0 {
			add("缺少或错序的四段标题：%s", heading)
			continue
		}
		cursor += idx + len(heading)
	}

	factsText := factsBlob(chart.Facts)
	if factsText == "" {
		add("盘面事实为空，模型没有可依据的卦象")
	}

	for _, phrase := range forbiddenPhrases {
		if containsAffirmative(text, phrase) {
			add("出现绝对化/诱导性词：%s", phrase)
		}
	}
	for _, phrase := range []string{"必成", "必败"} {
		if containsAffirmative(text, phrase) {
			add("出现绝对化吉凶承诺：%s", phrase)
		}
	}

	for _, h := range liuyao.AllHexagrams() {
		if strings.Contains(text, h.Name) && !strings.Contains(factsText, h.Name) {
			add("出现盘面事实之外的卦名：%s", h.Name)
		}
	}

	for _, gz := range sixtyJiaZi {
		if strings.Contains(text, gz) && !strings.Contains(factsText, gz) {
			add("出现盘面事实之外的干支：%s", gz)
		}
	}

	lines := parseAuditLines(chart.Facts)
	checkMovingLines(text, lines, add)
	checkLineLiuQin(text, lines, add)
	checkShiYing(text, factsText, add)
	checkUseShen(text, chart.Facts, add)
	checkShenSha(text, factsText, add)

	return issues
}

var requiredHeadings = []string{
	"## 盘面要点",
	"## 断语",
	"## 应期与走向",
	"## 建议",
}

var forbiddenPhrases = []string{
	"保证",
	"包治",
	"改运",
	"消灾",
	"开光",
	"百分百",
	"绝对",
	"一定会",
	"一定能",
	"一定是",
	"必然",
	"注定",
	"稳赚",
	"铁定",
	"包你",
}

func containsAffirmative(text, phrase string) bool {
	for offset := 0; offset <= len(text); {
		idx := strings.Index(text[offset:], phrase)
		if idx < 0 {
			return false
		}
		idx += offset
		if !negatedBefore(text, idx) {
			return true
		}
		offset = idx + len(phrase)
	}
	return false
}

func negatedBefore(text string, idx int) bool {
	runes := []rune(text[:idx])
	window := runes
	if len(window) > 8 {
		window = window[len(window)-8:]
	}
	prefix := string(window)
	for _, neg := range []string{"不", "未", "没", "无", "非", "别", "勿", "莫", "岂", "不能", "不会", "未必", "并非", "绝不"} {
		if strings.Contains(prefix, neg) {
			return true
		}
	}
	return false
}

var sixtyJiaZi = func() []string {
	stems := []rune("甲乙丙丁戊己庚辛壬癸")
	branches := []rune("子丑寅卯辰巳午未申酉戌亥")
	out := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		out = append(out, string(stems[i%10])+string(branches[i%12]))
	}
	return out
}()

var shenShaTerms = []string{
	"驿马",
	"天乙",
	"文昌",
	"羊刃",
	"劫煞",
	"灾煞",
	"华盖",
	"将星",
}

type auditLine struct {
	pos    int
	liuQin string
	moving bool
}

var yaoPositions = map[string]int{
	"初爻": 1,
	"二爻": 2,
	"三爻": 3,
	"四爻": 4,
	"五爻": 5,
	"上爻": 6,
}

var liuQinSet = map[string]bool{
	"父母": true,
	"兄弟": true,
	"子孙": true,
	"妻财": true,
	"官鬼": true,
}

func factsBlob(facts []schema.Fact) string {
	var b strings.Builder
	for _, f := range facts {
		fmt.Fprintf(&b, "%s：%s\n", f.Label, f.Value)
	}
	return b.String()
}

func parseAuditLines(facts []schema.Fact) map[int]auditLine {
	out := make(map[int]auditLine)
	for _, f := range facts {
		pos, ok := yaoPositions[f.Label]
		if !ok || strings.Contains(f.Label, "变") {
			continue
		}
		line := auditLine{pos: pos}
		for _, token := range strings.Split(f.Value, "｜") {
			token = strings.TrimSpace(token)
			if token == "动" {
				line.moving = true
			}
			if token == "静" {
				line.moving = false
			}
			if liuQinSet[token] {
				line.liuQin = token
			}
		}
		out[pos] = line
	}
	return out
}

var (
	reMovingYao = regexp.MustCompile(`([初二三四五上])爻(?:发动|为动|是动|动)`)
	reStaticYao = regexp.MustCompile(`([初二三四五上])爻(?:为静|不动|未动)`)
	reQinAToYao = regexp.MustCompile(`([初二三四五上])爻(?:为|临|是|持|带|坐)(子孙|妻财|官鬼|父母|兄弟)`)
	reQinToYaoB = regexp.MustCompile(`(子孙|妻财|官鬼|父母|兄弟)(?:在|临|持|居|落)([初二三四五上])爻`)
	reShiYao    = regexp.MustCompile(`世(?:爻)?[在是落居]([初二三四五上])爻`)
	reYingYao   = regexp.MustCompile(`应(?:爻)?[在是落居]([初二三四五上])爻`)
	reUseShenA  = regexp.MustCompile(`以(子孙|妻财|官鬼|父母|兄弟|世爻)为用神`)
	reUseShenB  = regexp.MustCompile(`主?用神(?:为|是|取|：|:)\s*(子孙|妻财|官鬼|父母|兄弟|世爻)`)
)

func checkMovingLines(text string, lines map[int]auditLine, add func(string, ...any)) {
	for _, match := range reMovingYao.FindAllStringSubmatch(text, -1) {
		pos := yaoPositions[match[1]+"爻"]
		if line, ok := lines[pos]; ok && !line.moving {
			add("%s在盘面中是静爻，输出却说它发动", match[1]+"爻")
		}
	}
	for _, match := range reStaticYao.FindAllStringSubmatch(text, -1) {
		pos := yaoPositions[match[1]+"爻"]
		if line, ok := lines[pos]; ok && line.moving {
			add("%s在盘面中是动爻，输出却说它不动", match[1]+"爻")
		}
	}
}

func checkLineLiuQin(text string, lines map[int]auditLine, add func(string, ...any)) {
	for _, match := range reQinAToYao.FindAllStringSubmatch(text, -1) {
		pos := yaoPositions[match[1]+"爻"]
		checkQinAt(pos, match[2], lines, add)
	}
	for _, match := range reQinToYaoB.FindAllStringSubmatch(text, -1) {
		pos := yaoPositions[match[2]+"爻"]
		checkQinAt(pos, match[1], lines, add)
	}
}

func checkQinAt(pos int, qin string, lines map[int]auditLine, add func(string, ...any)) {
	line, ok := lines[pos]
	if !ok || line.liuQin == "" || line.liuQin == qin {
		return
	}
	add("%s实际为%s，输出却写成%s", yaoName(pos), line.liuQin, qin)
}

func checkShiYing(text, factsText string, add func(string, ...any)) {
	if match := reShiYao.FindStringSubmatch(factsText); len(match) == 2 {
		if statements := reShiYao.FindAllStringSubmatch(text, -1); len(statements) > 0 {
			for _, s := range statements {
				if s[1] != match[1] {
					add("世爻应在%s，输出却写在%s", match[1]+"爻", s[1]+"爻")
				}
			}
		}
	}
	if match := reYingYao.FindStringSubmatch(factsText); len(match) == 2 {
		if statements := reYingYao.FindAllStringSubmatch(text, -1); len(statements) > 0 {
			for _, s := range statements {
				if s[1] != match[1] {
					add("应爻应在%s，输出却写在%s", match[1]+"爻", s[1]+"爻")
				}
			}
		}
	}
}

type useShenClaim struct {
	name       string
	contextual bool
}

func collectUseShenClaims(text string) []useShenClaim {
	var claims []useShenClaim
	for _, re := range []*regexp.Regexp{reUseShenA, reUseShenB} {
		for _, loc := range re.FindAllStringSubmatchIndex(text, -1) {
			name := text[loc[2]:loc[3]]
			prefix := text[:loc[0]]
			if cut := strings.LastIndexAny(prefix, "。；，,\n"); cut >= 0 {
				prefix = prefix[cut+1:]
			}
			contextual := strings.Contains(prefix, "男占") || strings.Contains(prefix, "女占")
			claims = append(claims, useShenClaim{name: name, contextual: contextual})
		}
	}
	return claims
}

func checkUseShen(text string, facts []schema.Fact, add func(string, ...any)) {
	var value string
	for _, f := range facts {
		if f.Label == "用神" {
			value = f.Value
			break
		}
	}
	if value == "" {
		return
	}

	claims := collectUseShenClaims(text)

	if strings.Contains(value, "待确认：求测者性别") || strings.Contains(value, "待确认：所问对象") {
		for _, claim := range claims {
			if !claim.contextual {
				add("用神待确认，输出不能只取一个主用神：%s", claim.name)
			}
		}
		return
	}

	prefix := value
	if idx := strings.Index(prefix, "；"); idx >= 0 {
		prefix = prefix[:idx]
	}
	prefix = strings.TrimPrefix(prefix, "主用神：")
	allowed := make(map[string]bool)
	for _, name := range strings.Split(prefix, "、") {
		allowed[strings.TrimSpace(name)] = true
	}
	if len(allowed) == 0 {
		return
	}

	missing := make(map[string]bool)
	if idx := strings.Index(value, "用神不现："); idx >= 0 {
		rest := value[idx+len("用神不现："):]
		if end := strings.Index(rest, "；"); end >= 0 {
			rest = rest[:end]
		}
		for _, name := range strings.Split(rest, "、") {
			missing[strings.TrimSpace(name)] = true
		}
	}
	for _, claim := range claims {
		if missing[claim.name] {
			add("用神%s在盘面不现，输出不能直接写成可用", claim.name)
			continue
		}
		if !allowed[claim.name] {
			add("用神应为%s，输出却写成%s", prefix, claim.name)
		}
	}
}

func checkShenSha(text, factsText string, add func(string, ...any)) {
	for _, term := range shenShaTerms {
		if strings.Contains(text, term) && !strings.Contains(factsText, term) {
			add("出现盘面事实之外的神煞：%s", term)
		}
	}
}

func yaoName(pos int) string {
	for label, p := range yaoPositions {
		if p == pos {
			return label
		}
	}
	return "该爻"
}
