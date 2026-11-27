package ai

import (
	"testing"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

func xiaoLiuRenFixture() schema.Chart {
	return schema.Chart{
		Method: schema.MethodXiaoLiuRen,
		Facts: []schema.Fact{
			{Key: "xiaoliuren.calendar", Label: "农历", Value: "农历10月19日 · 酉时", Weight: schema.WeightCore},
			{Key: "xiaoliuren.palaces", Label: "三宫", Value: "月赤口 · 日赤口 · 时大安", Weight: schema.WeightCore},
			{Key: "xiaoliuren.final", Label: "落宫", Value: "大安（木·青龙·东）· 吉", Weight: schema.WeightCore},
			{Key: "xiaoliuren.window", Label: "近事窗口", Value: "1—15 日", Weight: schema.WeightNormal},
		},
	}
}

func meiHuaFixture() schema.Chart {
	return schema.Chart{
		Method: schema.MethodMeiHua,
		Facts: []schema.Fact{
			{Key: "meihua.ben", Label: "本卦", Value: "火雷噬嗑（上离下震）", Weight: schema.WeightCore},
			{Key: "meihua.hu", Label: "互卦", Value: "水山蹇（上坎下艮）", Weight: schema.WeightCore},
			{Key: "meihua.bian", Label: "变卦", Value: "火地晋（上离下坤）", Weight: schema.WeightCore},
			{Key: "meihua.moving", Label: "动爻", Value: "第1爻动", Weight: schema.WeightCore},
			{Key: "meihua.tiyong", Label: "体用", Value: "体卦离（火）· 用卦震（木）· 用生体", Weight: schema.WeightCore},
		},
	}
}

func qiMenFixture() schema.Chart {
	return schema.Chart{
		Method: schema.MethodQiMen,
		Facts: []schema.Fact{
			{Key: "qimen.ju", Label: "起局", Value: "阴遁2局 · 小雪 · 下元", Weight: schema.WeightCore},
			{Key: "qimen.zhifu", Label: "值符", Value: "天英落震三宫", Weight: schema.WeightCore},
			{Key: "qimen.zhishi", Label: "值使", Value: "景门落艮八宫", Weight: schema.WeightCore},
			{Key: "qimen.kong", Label: "时空亡", Value: "午、未", Weight: schema.WeightCore},
			{Key: "qimen.horse", Label: "马星", Value: "亥在乾六宫", Weight: schema.WeightNormal},
			{Key: "qimen.palace.1", Label: "坎一宫", Value: "天盘乙 / 地盘己 · 天冲 · 杜门 · 太阴", Weight: schema.WeightNormal},
			{Key: "qimen.palace.3", Label: "震三宫", Value: "天盘庚 / 地盘乙 · 天英 · 死门 · 值符", Weight: schema.WeightNormal},
			{Key: "qimen.palace.6", Label: "乾六宫", Value: "天盘辛 / 地盘癸 · 天任 · 伤门 · 六合 · 马", Weight: schema.WeightNormal},
			{Key: "qimen.recent", Label: "近期考试/比赛参考", Value: "考试：景门在艮八宫（艮）、天辅在艮八宫（艮）；比赛：值符在震三宫（震）、时干在坎一宫（坎）、伤门在乾六宫（乾）、天冲在坎一宫（坎）", Weight: schema.WeightNormal},
		},
	}
}

func TestAuditFollowupAcceptsGroundedNewMethods(t *testing.T) {
	cases := []struct {
		name  string
		chart schema.Chart
		text  string
	}{
		{
			name:  "小六壬",
			chart: xiaoLiuRenFixture(),
			text:  "月日双赤口主竞争与口舌，时宫和落宫同在 大安，说明节奏可稳，但事缓可成，不宜把日期说死。",
		},
		{
			name:  "梅花",
			chart: meiHuaFixture(),
			text:  "本卦 火雷噬嗑，互卦 水山蹇，变卦 火地晋；体卦 离 火，用卦 震 木，用生体，过程见阻而收束向好。",
		},
		{
			name:  "奇门",
			chart: qiMenFixture(),
			text:  "阴遁2局，值符 天英 落震三宫，值使 景门 落艮八宫；时干 乙 在坎一宫，与 天冲、杜门、太阴 同宫。",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if issues := AuditFollowup(tc.text, tc.chart); len(issues) > 0 {
				t.Fatalf("合规追问不应被拦截: %v", issues)
			}
		})
	}
}

func TestAuditFollowupRejectsUngroundedNewMethods(t *testing.T) {
	cases := []struct {
		name  string
		chart schema.Chart
		text  string
	}{
		{
			name:  "小六壬",
			chart: xiaoLiuRenFixture(),
			text:  "落宫是 小吉，主小成和合。",
		},
		{
			name:  "梅花",
			chart: meiHuaFixture(),
			text:  "互卦其实是 乾为天，所以事情会更强。",
		},
		{
			name:  "奇门",
			chart: qiMenFixture(),
			text:  "值符是 天蓬，落在 阴遁9局 的离九宫。",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if issues := AuditFollowup(tc.text, tc.chart); len(issues) == 0 {
				t.Fatal("盘面事实之外的术语应被拦截")
			}
		})
	}
}
