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
	if strings.TrimSpace(text) == "" {
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

	issues = append(issues, auditMethodContent(text, chart, true)...)
	return issues
}

// AuditFollowup 追问不要求四段标题，但仍逐项检查盘面事实、绝对化措辞和
// 用神/世应/六亲是否被写错。
func AuditFollowup(text string, chart schema.Chart) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if strings.TrimSpace(text) == "" {
		return []string{"输出为空"}
	}
	return auditMethodContent(text, chart, false)
}

func auditMethodContent(text string, chart schema.Chart, requireSurname bool) []string {
	switch chart.Method {
	case schema.MethodLiuYao:
		return auditLiuYaoContent(text, chart, requireSurname)
	case schema.MethodDaLiuRen:
		return append(auditGenericContent(text, chart), auditDaLiuRenTerms(text, chart.Facts)...)
	case schema.MethodZiWei:
		return append(auditGenericContent(text, chart), auditZiWeiTerms(text, chart.Facts)...)
	case schema.MethodXiaoLiuRen:
		return append(auditGenericContent(text, chart), auditXiaoLiuRenTerms(text, chart.Facts)...)
	case schema.MethodMeiHua:
		return append(auditGenericContent(text, chart), auditMeiHuaTerms(text, chart.Facts)...)
	case schema.MethodQiMen:
		return append(auditGenericContent(text, chart), auditQiMenTerms(text, chart.Facts)...)
	default:
		return auditGenericContent(text, chart)
	}
}

func auditGenericContent(text string, chart schema.Chart) []string {
	var issues []string
	if factsBlob(chart.Facts) == "" {
		issues = append(issues, "盘面事实为空，模型没有可依据的盘面")
	}
	for _, phrase := range forbiddenPhrases {
		if containsAffirmative(text, phrase) {
			issues = append(issues, "出现绝对化/诱导性词："+phrase)
		}
	}
	for _, phrase := range []string{"必成", "必败"} {
		if containsAffirmative(text, phrase) {
			issues = append(issues, "出现绝对化吉凶承诺："+phrase)
		}
	}
	return issues
}

var daLiuRenGates = []string{"贼克", "比用", "涉害", "遥克", "昴星", "别责", "八专", "伏吟", "返吟"}

func auditDaLiuRenTerms(text string, facts []schema.Fact) []string {
	factsText := factsBlob(facts)
	var issues []string
	for _, gate := range daLiuRenGates {
		if strings.Contains(text, gate) && !strings.Contains(factsText, gate) {
			issues = append(issues, "出现盘面事实之外的课体："+gate)
		}
	}
	for _, name := range []string{"贵人", "螣蛇", "朱雀", "六合", "勾陈", "青龙", "天空", "白虎", "太常", "玄武", "太阴", "天后"} {
		if strings.Contains(text, name) && !strings.Contains(factsText, name) {
			issues = append(issues, "出现盘面事实之外的天将："+name)
		}
	}
	return issues
}

var ziWeiStars = []string{
	"紫微", "天机", "太阳", "武曲", "天同", "廉贞",
	"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军",
	"左辅", "右弼", "文昌", "文曲", "天魁", "天钺", "禄存", "擎羊", "陀罗",
	"火星", "铃星", "地空", "地劫", "天马",
}

func auditZiWeiTerms(text string, facts []schema.Fact) []string {
	factsText := factsBlob(facts)
	var issues []string
	for _, star := range ziWeiStars {
		if strings.Contains(text, star) && !strings.Contains(factsText, star) {
			issues = append(issues, "出现盘面事实之外的星曜："+star)
		}
	}
	for _, star := range ziWeiStars {
		for _, hua := range []string{"禄", "权", "科", "忌"} {
			claim := star + "化" + hua
			// 允许模型写“武曲星化权”，但落盘校验仍按“武曲化权”。
			if strings.Contains(text, claim) || strings.Contains(text, star+"星化"+hua) {
				if !strings.Contains(factsText, claim) {
					issues = append(issues, "出现盘面事实之外的四化："+claim)
				}
			}
		}
	}
	return issues
}

var xiaoLiuRenPalaces = []string{"大安", "留连", "速喜", "赤口", "小吉", "空亡"}

func auditXiaoLiuRenTerms(text string, facts []schema.Fact) []string {
	factsText := factsBlob(facts)
	var issues []string
	for _, palace := range xiaoLiuRenPalaces {
		if strings.Contains(text, palace) && !strings.Contains(factsText, palace) {
			issues = append(issues, "出现盘面事实之外的六宫："+palace)
		}
	}
	return issues
}

func auditMeiHuaTerms(text string, facts []schema.Fact) []string {
	factsText := factsBlob(facts)
	var issues []string
	// 64 卦名复用六爻的卦名表；梅花只允许出现本卦、互卦、变卦。
	for _, h := range liuyao.AllHexagrams() {
		if strings.Contains(text, h.Name) && !strings.Contains(factsText, h.Name) {
			issues = append(issues, "出现盘面事实之外的卦名："+h.Name)
		}
	}
	for _, trigram := range []string{"乾", "兑", "离", "震", "巽", "坎", "艮", "坤"} {
		if strings.Contains(text, trigram) && !strings.Contains(factsText, trigram) {
			issues = append(issues, "出现盘面事实之外的八卦："+trigram)
		}
	}
	return issues
}

var qiMenStars = []string{"天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英"}
var qiMenDoors = []string{"休门", "生门", "伤门", "杜门", "景门", "死门", "惊门", "开门"}
var qiMenGods = []string{"值符", "腾蛇", "太阴", "六合", "白虎", "玄武", "九地", "九天"}

func auditQiMenTerms(text string, facts []schema.Fact) []string {
	factsText := factsBlob(facts)
	var issues []string
	for _, term := range append(append(append([]string{}, qiMenStars...), qiMenDoors...), qiMenGods...) {
		if strings.Contains(text, term) && !strings.Contains(factsText, term) {
			issues = append(issues, "出现盘面事实之外的奇门元素："+term)
		}
	}
	for _, term := range []string{"阳遁", "阴遁", "上元", "中元", "下元", "马星", "空亡", "门迫"} {
		if strings.Contains(text, term) && !strings.Contains(factsText, term) {
			issues = append(issues, "出现盘面事实之外的奇门状态："+term)
		}
	}
	return issues
}

func auditLiuYaoContent(text string, chart schema.Chart, requireSurname bool) []string {
	var issues []string
	seen := make(map[string]bool)
	add := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		if !seen[msg] {
			seen[msg] = true
			issues = append(issues, msg)
		}
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
	checkSurnameAnswer(text, chart.Facts, requireSurname, add)
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

var zodiacAnimals = map[string]bool{
	"鼠": true, "牛": true, "虎": true, "兔": true, "龙": true, "蛇": true,
	"马": true, "羊": true, "猴": true, "鸡": true, "狗": true, "猪": true,
}

func checkSurnameAnswer(text string, facts []schema.Fact, requireAnswer bool, add func(string, ...any)) {
	allowed, ok := surnameCandidates(facts)
	if !ok {
		return
	}
	relevant := make(map[string]bool, len(allowed)+len(zodiacAnimals))
	for name := range allowed {
		relevant[name] = true
	}
	for name := range zodiacAnimals {
		relevant[name] = true
	}

	claims := collectSurnameClaims(text, relevant)
	hasAllowed := false
	for _, name := range claims {
		if allowed[name] {
			hasAllowed = true
			continue
		}
		if zodiacAnimals[name] {
			add("把生肖动物直接当姓氏：%s", name)
		}
	}
	if !hasAllowed && requireAnswer {
		add("测姓类解读没有给出真实姓氏候选")
	}
}

func surnameCandidates(facts []schema.Fact) (map[string]bool, bool) {
	for _, f := range facts {
		if f.Label != "姓氏候选" {
			continue
		}
		value := f.Value
		if cut := strings.Index(value, "；"); cut >= 0 {
			value = value[:cut]
		}
		value = strings.TrimPrefix(value, "常见姓氏候选：")
		allowed := make(map[string]bool)
		for _, token := range strings.Split(value, "、") {
			token = strings.TrimSpace(token)
			if cut := strings.Index(token, "（"); cut >= 0 {
				token = token[:cut]
			}
			if token != "" {
				allowed[token] = true
			}
		}
		return allowed, true
	}
	return nil, false
}

// collectSurnameClaims 只收集明确的“姓 X / X 姓 / 答案 X”句式；仅把候选姓氏与
// 十二生肖字纳入，避免把“百姓的”之类普通文本误判成姓氏声明。
func collectSurnameClaims(text string, relevant map[string]bool) []string {
	runes := []rune(text)
	var out []string
	addToken := func(start int) {
		j := start
		for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n') {
			j++
		}
		j = skipClaimWords(runes, j)
		if j < len(runes) && relevant[string(runes[j])] {
			out = append(out, string(runes[j]))
		}
	}
	addBefore := func(end int) {
		j := end - 1
		for j >= 0 && (runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n' || strings.ContainsRune("「」“”‘’\"'", runes[j])) {
			j--
		}
		if j >= 0 && relevant[string(runes[j])] {
			out = append(out, string(runes[j]))
		}
	}
	for i := 0; i < len(runes); i++ {
		if runes[i] != '姓' {
			continue
		}
		j := i + 1
		if j < len(runes) && runes[j] == '氏' {
			j++
		}
		addToken(j)
		addBefore(i)
	}
	for i := 0; i+1 < len(runes); i++ {
		if string(runes[i:i+2]) == "为姓" {
			addBefore(i)
		}
		if string(runes[i:i+2]) != "答案" && string(runes[i:i+2]) != "谜底" && string(runes[i:i+2]) != "结论" {
			continue
		}
		addToken(i + 2)
	}
	return out
}

func skipClaimWords(runes []rune, j int) int {
	words := [][]rune{
		{'可', '能'}, {'应', '该'}, {'大', '概'}, {'约'}, {'就'}, {'即'},
		{'是'}, {'为'}, {'叫'}, {'：'}, {':'}, {'候', '选'},
		{'指', '向'}, {'落', '在'}, {'对', '应'}, {'取'}, {'以'},
		{'「'}, {'」'}, {'“'}, {'”'}, {'‘'}, {'’'}, {'"'}, {'\''},
	}
	for {
		matched := false
		for _, word := range words {
			if j+len(word) > len(runes) {
				continue
			}
			same := true
			for k := range word {
				if runes[j+k] != word[k] {
					same = false
					break
				}
			}
			if same {
				j += len(word)
				matched = true
				break
			}
		}
		if !matched {
			return j
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
