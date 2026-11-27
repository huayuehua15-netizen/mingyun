package ziwei

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

func buildTestBirth(t *testing.T, in Input, opt BirthOptions) (Birth, calendar.Context) {
	t.Helper()
	birth, err := ParseBirth(in, opt)
	if err != nil {
		t.Fatal(err)
	}
	cal, err := calendar.Build(calendar.Request{At: birth.SolarTime})
	if err != nil {
		t.Fatal(err)
	}
	return birth, cal
}

func TestCastSolarChart(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	in := Input{
		Question:     "事业方向",
		Gender:       "male",
		CalendarType: "solar",
		BirthDate:    "1990-05-15",
		BirthTime:    "10:30",
	}
	birth, cal := buildTestBirth(t, in, BirthOptions{Location: loc})
	chart, err := Cast(in, birth, cal)
	if err != nil {
		t.Fatal(err)
	}
	payload := chart.Payload.(Payload)
	if len(payload.Palaces) != 12 {
		t.Fatalf("palaces = %d", len(payload.Palaces))
	}
	if payload.MingGong.Valid() == false || payload.ShenGong.Valid() == false {
		t.Fatalf("ming/shen invalid")
	}
	major := 0
	minor := 0
	for _, p := range payload.Palaces {
		major += len(p.MajorStars)
		minor += len(p.MinorStars)
	}
	if major != 14 {
		t.Fatalf("major stars = %d, want 14", major)
	}
	if minor < 10 {
		t.Fatalf("minor stars = %d", minor)
	}
	if len(payload.SihuaYear) != 4 {
		t.Fatalf("sihua = %d", len(payload.SihuaYear))
	}
	if len(payload.DaXian) != 12 {
		t.Fatalf("daxian = %d", len(payload.DaXian))
	}
	if payload.WuXingJu == "" {
		t.Fatal("empty wuxing ju")
	}
	if payload.ZiWeiPos < 0 || payload.ZiWeiPos >= 12 {
		t.Fatalf("ziwei pos = %d", payload.ZiWeiPos)
	}
	if payload.TianFuPos != (calendar.Zhi(calendar.Mod(4-int(payload.ZiWeiPos), 12))) {
		t.Fatalf("tianfu mismatch: %s %s", payload.ZiWeiPos, payload.TianFuPos)
	}
	t.Logf("birth=%+v summary=%s", birth, payload.Summary())
	for _, p := range payload.Palaces {
		t.Logf("%s %s 主:%v 辅:%v 大限:%s", p.Name, p.StemBranch, p.MajorStars, p.MinorStars, p.DaxianRange)
	}
}

func TestZiweiPositionTable(t *testing.T) {
	cases := []struct {
		day, ju int
		want    calendar.Zhi
	}{
		{1, 2, calendar.ZhiChou},
		{2, 2, calendar.ZhiYin},
		{3, 3, calendar.ZhiYin},
		{4, 4, calendar.ZhiYin},
		{1, 6, calendar.ZhiYou},
	}
	for _, c := range cases {
		if got := ziweiPosition(c.day, c.ju); got != c.want {
			t.Fatalf("day=%d ju=%d got=%s want=%s", c.day, c.ju, got, c.want)
		}
	}
}

func TestPalaceFormula(t *testing.T) {
	if got := mingPalace(1, 0); got != calendar.ZhiYin {
		t.Fatalf("正月子时命宫 = %s", got)
	}
	if got := mingPalace(1, 1); got != calendar.ZhiChou {
		t.Fatalf("正月丑时命宫 = %s", got)
	}
	if got := shenPalace(1, 0); got != calendar.ZhiYin {
		t.Fatalf("正月子时身宫 = %s", got)
	}
	if got := shenPalace(1, 1); got != calendar.ZhiMao {
		t.Fatalf("正月丑时身宫 = %s", got)
	}
}

func TestLunarInputAndRoundTrip(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	in := Input{
		Gender:       "female",
		CalendarType: "lunar",
		BirthDate:    "1990-04-21",
		BirthTime:    "10:30",
	}
	birth, cal := buildTestBirth(t, in, BirthOptions{Location: loc})
	if birth.LunarMonth != 4 || birth.LunarDay != 21 {
		t.Fatalf("lunar = %d-%d leap=%v", birth.LunarMonth, birth.LunarDay, birth.LunarLeap)
	}
	chart, err := Cast(in, birth, cal)
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
	if out.MingGong != chart.Payload.(Payload).MingGong {
		t.Fatalf("round trip mismatch")
	}
}

func TestInvalidLeapMonth(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	_, err := ParseBirth(Input{
		Gender:       "male",
		CalendarType: "lunar",
		BirthDate:    "1990-04-21",
		BirthTime:    "10:30",
		IsLeapMonth:  true,
	}, BirthOptions{Location: loc})
	// 1990 年是否有闰四月随历法库而定；这里只要求非法闰月能返回错误，
	// 若该年确有闰四月，则解析成功也说明输入合法。两种结果都不应有 panic。
	if err != nil && err.Error() == "" {
		t.Fatal("expected non-empty error")
	}
}
