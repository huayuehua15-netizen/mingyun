package liuyao

import (
	"strings"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

var cst = time.FixedZone("CST", 8*3600)

func testCal(t *testing.T) calendar.Context {
	t.Helper()
	cal, err := calendar.Build(calendar.Request{
		At:               time.Date(2026, 11, 25, 14, 30, 0, 0, cst),
		Longitude:        113.26,
		HasLongitude:     true,
		UseTrueSolarTime: true,
		DayBoundary:      calendar.ZiChu,
	})
	if err != nil {
		t.Fatalf("构建历法上下文失败: %v", err)
	}
	return cal
}

func TestHexagramTables(t *testing.T) {
	all := AllHexagrams()
	if len(all) != 64 {
		t.Fatalf("六十四卦数量 = %d", len(all))
	}

	seen := make(map[uint8]bool, 64)
	names := make(map[string]bool, 64)
	for _, h := range all {
		if seen[h.Bits()] {
			t.Fatalf("卦重复: %s", h.Name)
		}
		seen[h.Bits()] = true
		if names[h.Name] {
			t.Fatalf("卦名重复: %s", h.Name)
		}
		names[h.Name] = true

		back, ok := FromBits(h.Bits())
		if !ok || back.Name != h.Name {
			t.Fatalf("%s 二进制往返失败: %v", h.Name, back)
		}
		if got := palaceHexagrams[h.Palace][h.Rank]; got != h.Name {
			t.Fatalf("%s 归宫错误: 期望 %s，实为 %s", h.Name, got, palaceHexagrams[h.Palace][h.Rank])
		}
	}

	for p := 0; p < 8; p++ {
		for r := 0; r < 8; r++ {
			if !names[palaceHexagrams[p][r]] {
				t.Fatalf("宫表出现未知卦名: %s", palaceHexagrams[p][r])
			}
		}
	}
}

func TestPureHexagrams(t *testing.T) {
	cases := []struct {
		name  string
		upper Trigram
		lower Trigram
		shi   int
		ying  int
	}{
		{"乾为天", TrigramQian, TrigramQian, 6, 3},
		{"兑为泽", TrigramDui, TrigramDui, 6, 3},
		{"离为火", TrigramLi, TrigramLi, 6, 3},
		{"震为雷", TrigramZhen, TrigramZhen, 6, 3},
		{"巽为风", TrigramXun, TrigramXun, 6, 3},
		{"坎为水", TrigramKan, TrigramKan, 6, 3},
		{"艮为山", TrigramGen, TrigramGen, 6, 3},
		{"坤为地", TrigramKun, TrigramKun, 6, 3},
	}
	for _, c := range cases {
		h, ok := FromName(c.name)
		if !ok {
			t.Fatalf("%s 不存在", c.name)
		}
		if h.Upper != c.upper || h.Lower != c.lower {
			t.Fatalf("%s 上下卦错误: %s/%s", c.name, h.Upper, h.Lower)
		}
		if h.Rank != 0 || h.Palace != c.upper {
			t.Fatalf("%s 应为 %s宫本宫，实为 %s", c.name, c.upper, h)
		}
		if h.ShiYao() != c.shi || h.YingYao() != c.ying {
			t.Fatalf("%s 世应 = %d/%d", c.name, h.ShiYao(), h.YingYao())
		}
	}
}

func TestPalaceRankAndShiYing(t *testing.T) {
	cases := []struct {
		name   string
		palace Trigram
		rank   int
		shi    int
		ying   int
	}{
		{"乾为天", TrigramQian, 0, 6, 3},
		{"天风姤", TrigramQian, 1, 1, 4},
		{"天山遁", TrigramQian, 2, 2, 5},
		{"天地否", TrigramQian, 3, 3, 6},
		{"风地观", TrigramQian, 4, 4, 1},
		{"山地剥", TrigramQian, 5, 5, 2},
		{"火地晋", TrigramQian, 6, 4, 1},
		{"火天大有", TrigramQian, 7, 3, 6},
		{"坤为地", TrigramKun, 0, 6, 3},
		{"地雷复", TrigramKun, 1, 1, 4},
		{"地天泰", TrigramKun, 3, 3, 6},
		{"水天需", TrigramKun, 6, 4, 1},
		{"水地比", TrigramKun, 7, 3, 6},
	}
	for _, c := range cases {
		h, ok := FromName(c.name)
		if !ok {
			t.Fatalf("%s 不存在", c.name)
		}
		if h.Palace != c.palace || h.Rank != c.rank {
			t.Fatalf("%s 归宫 = %s/%s，期望 %s/%s", c.name, h.Palace, h.RankLabel(), c.palace, rankNames[c.rank])
		}
		if h.ShiYao() != c.shi || h.YingYao() != c.ying {
			t.Fatalf("%s 世应 = %d/%d，期望 %d/%d", c.name, h.ShiYao(), h.YingYao(), c.shi, c.ying)
		}
	}
}

func TestNaJia(t *testing.T) {
	cases := []struct {
		name   string
		ganzhi [6]string
	}{
		{"乾为天", [6]string{"甲子", "甲寅", "甲辰", "壬午", "壬申", "壬戌"}},
		{"坤为地", [6]string{"乙未", "乙巳", "乙卯", "癸丑", "癸亥", "癸酉"}},
		{"天风姤", [6]string{"辛丑", "辛亥", "辛酉", "壬午", "壬申", "壬戌"}},
		{"地天泰", [6]string{"甲子", "甲寅", "甲辰", "癸丑", "癸亥", "癸酉"}},
	}
	for _, c := range cases {
		h, ok := FromName(c.name)
		if !ok {
			t.Fatalf("%s 不存在", c.name)
		}
		for pos := 1; pos <= 6; pos++ {
			gan, zhi := naJiaAtPos(h, pos)
			if got := gan.String() + zhi.String(); got != c.ganzhi[pos-1] {
				t.Fatalf("%s 第%d爻纳甲 = %s，期望 %s", c.name, pos, got, c.ganzhi[pos-1])
			}
		}
	}
}

func TestLiuShen(t *testing.T) {
	cases := []struct {
		gan   calendar.Gan
		first LiuShen
	}{
		{calendar.GanJia, LiuShenQingLong},
		{calendar.GanYi, LiuShenQingLong},
		{calendar.GanBing, LiuShenZhuQue},
		{calendar.GanDing, LiuShenZhuQue},
		{calendar.GanWu, LiuShenGouChen},
		{calendar.GanJi, LiuShenTengShe},
		{calendar.GanGeng, LiuShenBaiHu},
		{calendar.GanXin, LiuShenBaiHu},
		{calendar.GanRen, LiuShenXuanWu},
		{calendar.GanGui, LiuShenXuanWu},
	}
	for _, c := range cases {
		ls, err := LiuShenAt(c.gan, 1)
		if err != nil {
			t.Fatalf("%s 日初爻: %v", c.gan, err)
		}
		if ls != c.first {
			t.Fatalf("%s 日初爻六神 = %s，期望 %s", c.gan, ls, c.first)
		}
	}
	if _, err := LiuShenAt(calendar.GanJia, 7); err == nil {
		t.Fatal("爻位 7 应当报错")
	}
}

func TestLiuQin(t *testing.T) {
	cases := []struct {
		me, other calendar.WuXing
		want      LiuQin
	}{
		{calendar.WuXingJin, calendar.WuXingJin, LiuQinXiongDi},
		{calendar.WuXingJin, calendar.WuXingTu, LiuQinFuMu},
		{calendar.WuXingJin, calendar.WuXingShui, LiuQinZiSun},
		{calendar.WuXingJin, calendar.WuXingHuo, LiuQinGuanGui},
		{calendar.WuXingJin, calendar.WuXingMu, LiuQinQiCai},
		{calendar.WuXingTu, calendar.WuXingShui, LiuQinQiCai},
		{calendar.WuXingMu, calendar.WuXingTu, LiuQinQiCai},
	}
	for _, c := range cases {
		got, err := LiuQinFor(c.me, c.other)
		if err != nil {
			t.Fatalf("%s 对 %s: %v", c.me, c.other, err)
		}
		if got != c.want {
			t.Fatalf("%s 对 %s 六亲 = %s，期望 %s", c.me, c.other, got, c.want)
		}
	}
}

func TestLineKind(t *testing.T) {
	cases := []struct {
		backs int
		kind  LineKind
	}{
		{3, OldYang}, {2, YoungYin}, {1, YoungYang}, {0, OldYin},
	}
	for _, c := range cases {
		got, err := LineKindFromCoins(c.backs)
		if err != nil {
			t.Fatalf("%d 背: %v", c.backs, err)
		}
		if got != c.kind {
			t.Fatalf("%d 背 = %s，期望 %s", c.backs, got, c.kind)
		}
	}
	if _, err := LineKindFromCoins(4); err == nil {
		t.Fatal("4 背应当报错")
	}

	if OldYang.Changed() != YoungYin || OldYin.Changed() != YoungYang {
		t.Fatal("动爻变化错误")
	}
	if YoungYang.Changed() != YoungYang || YoungYin.Changed() != YoungYin {
		t.Fatal("静爻不应变化")
	}
	if !OldYang.Moving() || YoungYin.Moving() {
		t.Fatal("动爻判定错误")
	}
}

func TestCastPureQian(t *testing.T) {
	cal := testCal(t)
	chart, err := Cast(Input{
		Mode:     ModeManual,
		Lines:    []LineKind{YoungYang, YoungYang, YoungYang, YoungYang, YoungYang, YoungYang},
		Question: "测试",
	}, cal)
	if err != nil {
		t.Fatalf("排盘失败: %v", err)
	}

	p := chart.Payload.(Payload)
	if p.GuaName != "乾为天" || p.Palace != "乾宫" || p.Rank != "本宫" {
		t.Fatalf("卦 = %s", p.GuaName)
	}
	if p.ShiYao != 6 || p.YingYao != 3 {
		t.Fatalf("世应 = %d/%d", p.ShiYao, p.YingYao)
	}
	if p.XunKong[0] != "辰" || p.XunKong[1] != "巳" {
		t.Fatalf("旬空 = %v", p.XunKong)
	}
	if p.MonthGanZhi != "己亥" || p.DayGanZhi != "癸卯" {
		t.Fatalf("日月 = %s / %s，期望 己亥 / 癸卯", p.MonthGanZhi, p.DayGanZhi)
	}

	wantLiuQin := []string{"子孙", "妻财", "父母", "官鬼", "兄弟", "父母"}
	wantLiuShen := []string{"玄武", "青龙", "朱雀", "勾陈", "螣蛇", "白虎"}
	for i, l := range p.Lines {
		if l.LiuQin != wantLiuQin[i] {
			t.Fatalf("第%d爻六亲 = %s，期望 %s", l.Position, l.LiuQin, wantLiuQin[i])
		}
		if l.LiuShen != wantLiuShen[i] {
			t.Fatalf("第%d爻六神 = %s，期望 %s", l.Position, l.LiuShen, wantLiuShen[i])
		}
		if l.Moving {
			t.Fatalf("第%d爻不应发动", l.Position)
		}
	}

	if !strings.Contains(strings.Join(p.Lines[0].Marks, "、"), "旺") {
		t.Fatalf("初爻月建应旺，实为 %v", p.Lines[0].Marks)
	}
	if !strings.Contains(strings.Join(p.Lines[2].Marks, "、"), "旬空") {
		t.Fatalf("三爻辰土应旬空，实为 %v", p.Lines[2].Marks)
	}
	if !strings.Contains(strings.Join(p.Lines[5].Marks, "、"), "日合") {
		t.Fatalf("上爻戌土应日合，实为 %v", p.Lines[5].Marks)
	}

	if len(chart.Facts) < 10 {
		t.Fatalf("事实过少: %d", len(chart.Facts))
	}
	for _, f := range chart.Facts {
		if err := f.Validate(); err != nil {
			t.Fatalf("事实非法: %v", err)
		}
	}
}

func TestCastMovingToGou(t *testing.T) {
	cal := testCal(t)
	chart, err := Cast(Input{
		Mode:  ModeManual,
		Lines: []LineKind{OldYang, YoungYang, YoungYang, YoungYang, YoungYang, YoungYang},
	}, cal)
	if err != nil {
		t.Fatalf("排盘失败: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.GuaName != "乾为天" || p.ChangedGua != "天风姤" {
		t.Fatalf("卦 = %s 之 %s", p.GuaName, p.ChangedGua)
	}

	change := p.Lines[0].Change
	if change == nil {
		t.Fatal("初爻应发动")
	}
	if change.Zhi != "丑" || change.WuXing != "土" || change.LiuQin != "父母" {
		t.Fatalf("初爻变爻 = %+v", change)
	}
	if change.Relation != "回头克" {
		t.Fatalf("初爻变爻关系 = %q，期望 回头克", change.Relation)
	}
}

func TestCastFuShen(t *testing.T) {
	cal := testCal(t)
	// 天风姤：初爻辛丑、二爻辛亥、三爻辛酉、四爻壬午、五爻壬申、上爻壬戌，缺妻财。
	chart, err := Cast(Input{
		Mode:  ModeManual,
		Lines: []LineKind{YoungYin, YoungYang, YoungYang, YoungYang, YoungYang, YoungYang},
	}, cal)
	if err != nil {
		t.Fatalf("排盘失败: %v", err)
	}
	p := chart.Payload.(Payload)
	if p.GuaName != "天风姤" {
		t.Fatalf("卦 = %s，期望 天风姤", p.GuaName)
	}
	if len(p.FuShen) != 1 {
		t.Fatalf("伏神数量 = %d，期望 1: %+v", len(p.FuShen), p.FuShen)
	}
	f := p.FuShen[0]
	if f.LiuQin != "妻财" || f.Zhi != "寅" || f.WuXing != "木" || f.UnderPosition != 2 {
		t.Fatalf("伏神 = %+v", f)
	}
}

func TestCastCoinMode(t *testing.T) {
	cal := testCal(t)
	chart, err := Cast(Input{
		Mode:  ModeCoin,
		Backs: [6]int{3, 2, 1, 0, 3, 1},
	}, cal)
	if err != nil {
		t.Fatalf("铜钱起卦失败: %v", err)
	}
	p := chart.Payload.(Payload)
	wantMoving := [6]bool{true, false, false, true, true, false}
	for i, l := range p.Lines {
		if l.Moving != wantMoving[i] {
			t.Fatalf("第%d爻动否 = %v，期望 %v", i+1, l.Moving, wantMoving[i])
		}
	}
	if p.Lines[1].YinYang != "阴" || p.Lines[2].YinYang != "阳" {
		t.Fatalf("阴阳判定错误: %v", p.Lines)
	}
}

func TestCastRejectsBadInput(t *testing.T) {
	cal := testCal(t)
	if _, err := Cast(Input{Mode: ModeManual, Lines: []LineKind{YoungYang}}, cal); err == nil {
		t.Fatal("爻数不足应当报错")
	}
	if _, err := Cast(Input{Mode: "time"}, cal); err == nil {
		t.Fatal("未实现的起卦方式应当报错")
	}
}

func TestHexagramChanged(t *testing.T) {
	h, _ := FromName("乾为天")
	ch, ok := h.Changed([6]bool{true, false, false, false, false, false})
	if !ok || ch.Name != "天风姤" {
		t.Fatalf("乾初爻动 = %s", ch.Name)
	}

	ch, ok = h.Changed([6]bool{false, false, true, false, false, false})
	if !ok || ch.Name != "天泽履" {
		t.Fatalf("乾三爻动 = %s", ch.Name)
	}
}
