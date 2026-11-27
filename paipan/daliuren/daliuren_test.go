package daliuren

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

func testCalendar(t *testing.T, at time.Time) calendar.Context {
	t.Helper()
	cal, err := calendar.Build(calendar.Request{At: at})
	if err != nil {
		t.Fatal(err)
	}
	return cal
}

func TestCastBasicInvariants(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	cal := testCalendar(t, time.Date(2026, 11, 25, 14, 32, 0, 0, loc))
	chart, err := Cast(Input{Question: "这次考试能否通过？"}, cal)
	if err != nil {
		t.Fatal(err)
	}
	if chart.Method != schema.MethodDaLiuRen {
		t.Fatalf("method = %q", chart.Method)
	}
	if err := chart.Validate(); err != nil {
		t.Fatal(err)
	}

	payload, ok := chart.Payload.(Payload)
	if !ok {
		t.Fatalf("payload type = %T", chart.Payload)
	}
	if len(payload.TianPan) != 12 {
		t.Fatalf("tianpan = %d", len(payload.TianPan))
	}
	seen := map[calendar.Zhi]bool{}
	for _, p := range payload.TianPan {
		if seen[p.Heaven] {
			t.Fatalf("天盘重复: %s", p.Heaven)
		}
		seen[p.Heaven] = true
	}
	if len(payload.SiKe) != 4 {
		t.Fatalf("sike = %d", len(payload.SiKe))
	}
	if len(payload.SanChuan) != 3 {
		t.Fatalf("sanchuan = %d", len(payload.SanChuan))
	}
	if payload.SanChuan[0].Branch == payload.SanChuan[1].Branch && payload.SanChuan[0].Branch == payload.SanChuan[2].Branch {
		t.Fatalf("三传不应三支全同: %+v", payload.SanChuan)
	}
	if payload.MonthGeneral.Zhi != cal.MonthGeneral.Zhi {
		t.Fatalf("月将 = %s, want %s", payload.MonthGeneral.Zhi, cal.MonthGeneral.Zhi)
	}
	if len(chart.Facts) < 8 {
		t.Fatalf("facts = %d", len(chart.Facts))
	}
}

func TestFuYinAndFanYin(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	// 2026-11-25 小雪后月将为寅。寅时（03:32）月将加占时同宫，为伏吟。
	fuyinCal := testCalendar(t, time.Date(2026, 11, 25, 3, 32, 0, 0, loc))
	fuyinChart, err := Cast(Input{}, fuyinCal)
	if err != nil {
		t.Fatal(err)
	}
	fuyin := fuyinChart.Payload.(Payload)
	if fuyin.Gate != "伏吟" {
		t.Fatalf("寅时 gate = %s, want 伏吟 (%s)", fuyin.Gate, fuyin.Summary())
	}

	// 申时（15:32）与月将寅相冲，为返吟。
	fanyinCal := testCalendar(t, time.Date(2026, 11, 25, 15, 32, 0, 0, loc))
	fanyinChart, err := Cast(Input{}, fanyinCal)
	if err != nil {
		t.Fatal(err)
	}
	fanyin := fanyinChart.Payload.(Payload)
	if fanyin.Gate != "返吟" && fanyin.Gate != "贼克" && fanyin.Gate != "比用" && fanyin.Gate != "涉害" {
		t.Fatalf("申时 gate = %s, want 返吟或贼克 (%s)", fanyin.Gate, fanyin.Summary())
	}

	// 纯规则层验证：返吟无克时取驿马，不被遥克截走。
	tian := buildTianPan(calendar.ZhiYin, calendar.ZhiShen)
	sike := []SiKe{
		{Index: 1, Lower: calendar.ZhiZi, Upper: calendar.ZhiZi},
		{Index: 2, Lower: calendar.ZhiChou, Upper: calendar.ZhiChou},
		{Index: 3, Lower: calendar.ZhiChen, Upper: calendar.ZhiChen},
		{Index: 4, Lower: calendar.ZhiSi, Upper: calendar.ZhiSi},
	}
	gate, _, first, _, _, err := buildSanChuan(tian, sike, calendar.GanJia, calendar.ZhiChen, calendar.ZhiShen, calendar.ZhiYin)
	if err != nil {
		t.Fatal(err)
	}
	if gate != "返吟" {
		t.Fatalf("synthetic 返吟 gate = %s", gate)
	}
	if first != yiMa(calendar.ZhiChen) {
		t.Fatalf("返吟初传 = %s, want 驿马 %s", first, yiMa(calendar.ZhiChen))
	}
}

func TestPayloadMarshalRoundTrip(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	cal := testCalendar(t, time.Date(2026, 11, 25, 14, 32, 0, 0, loc))
	chart, err := Cast(Input{}, cal)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(chart.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var out Payload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.SanChuan[0].Branch != chart.Payload.(Payload).SanChuan[0].Branch {
		t.Fatalf("round trip mismatch: %s", raw)
	}
}
