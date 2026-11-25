package calendar

import (
	"encoding/json"
	"testing"
	"time"
)

func mustLoadShanghai(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestTermBoundaries2026(t *testing.T) {
	loc := mustLoadShanghai(t)
	at := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)

	prev, err := PrevTerm(at)
	if err != nil {
		t.Fatal(err)
	}
	if prev.Name != "小满" {
		t.Errorf("2026-06-01 前一节气应为小满，实为 %s", prev.Name)
	}
	want := time.Date(2026, 5, 21, 8, 36, 45, 0, cst)
	if !prev.At.Equal(want) {
		t.Errorf("小满时刻应为 %s，实为 %s", want.Format("2006-01-02 15:04:05"), prev.At.Format("2006-01-02 15:04:05"))
	}

	next, err := NextTerm(at)
	if err != nil {
		t.Fatal(err)
	}
	if next.Name != "芒种" {
		t.Errorf("2026-06-01 后一节气应为芒种，实为 %s", next.Name)
	}
}

func TestYearPillarChangesAtLiChun(t *testing.T) {
	loc := mustLoadShanghai(t)
	const liChun = "2026-02-04 04:02:08"
	wantLiChun, _ := time.ParseInLocation("2006-01-02 15:04:05", liChun, cst)

	before := time.Date(2026, 2, 4, 4, 0, 0, 0, loc)
	gz, err := YearPillar(before)
	if err != nil {
		t.Fatal(err)
	}
	if gz.String() != "乙巳" {
		t.Errorf("立春前年柱应为乙巳，实为 %s", gz)
	}

	after := time.Date(2026, 2, 4, 4, 10, 0, 0, loc)
	gz, err = YearPillar(after)
	if err != nil {
		t.Fatal(err)
	}
	if gz.String() != "丙午" {
		t.Errorf("立春后年柱应为丙午，实为 %s", gz)
	}

	// 确认立春时刻本身与探针一致，避免测试与实现各说各话。
	if !before.Before(wantLiChun) || !wantLiChun.Before(after) {
		t.Errorf("立春时刻 %s 应落在 %s 与 %s 之间", wantLiChun, before, after)
	}
}

func TestMonthGeneral(t *testing.T) {
	loc := mustLoadShanghai(t)
	at := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)
	mg, err := MonthGeneralAt(at)
	if err != nil {
		t.Fatal(err)
	}
	if mg.Zhi != ZhiShen || mg.Name != "传送" || mg.AfterTerm != "小满" {
		t.Errorf("2026-06-01 月将应为申将传送（小满后），实为 %s", mg)
	}

	at2 := time.Date(2026, 11, 25, 12, 0, 0, 0, loc)
	mg2, err := MonthGeneralAt(at2)
	if err != nil {
		t.Fatal(err)
	}
	if mg2.Zhi != ZhiYin || mg2.Name != "功曹" || mg2.AfterTerm != "小雪" {
		t.Errorf("2026-11-25 月将应为寅将功曹（小雪后），实为 %s", mg2)
	}
}

func TestBuildContext(t *testing.T) {
	loc := mustLoadShanghai(t)
	at := time.Date(2026, 6, 1, 12, 0, 0, 0, loc)

	ctx, err := Build(Request{
		At:               at,
		Longitude:        121.47,
		HasLongitude:     true,
		UseTrueSolarTime: true,
		DayBoundary:      ZiChu,
	})
	if err != nil {
		t.Fatal(err)
	}

	if ctx.Year.String() != "丙午" || ctx.Month.String() != "癸巳" {
		t.Errorf("年月柱应为 丙午/癸巳，实为 %s/%s", ctx.Year, ctx.Month)
	}
	if ctx.Day.String() != "丙午" {
		t.Errorf("日柱应为丙午，实为 %s", ctx.Day)
	}
	if ctx.XunKong[0].String()+ctx.XunKong[1].String() != "寅卯" {
		t.Errorf("甲辰旬空应为寅卯，实为 %s%s", ctx.XunKong[0], ctx.XunKong[1])
	}
	if !ctx.TrueSolarUsed {
		t.Error("应标记为已使用真太阳时")
	}
	if len(ctx.Warnings) == 0 {
		t.Error("启用真太阳时应给出提示")
	}

	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["day_gz"] != "丙午" {
		t.Errorf("JSON 中 day_gz 应为字符串 丙午，实为 %v", m["day_gz"])
	}
	kong, ok := m["xun_kong"].([]any)
	if !ok || len(kong) != 2 || kong[0] != "寅" {
		t.Errorf("JSON 中 xun_kong 应为 [\"寅\",\"卯\"]，实为 %v", m["xun_kong"])
	}
}

func TestBuildRejectsOutOfRange(t *testing.T) {
	loc := mustLoadShanghai(t)
	_, err := Build(Request{At: time.Date(1899, 12, 31, 12, 0, 0, 0, loc)})
	if err == nil {
		t.Error("1899 年应被拒绝")
	}
}

func TestCrossCheckDayGanZhiWithLunarGo(t *testing.T) {
	// 每 5 年抽 12 个月，与 lunar-go 的精确日柱交叉校验。
	for year := 1900; year <= 2100; year += 5 {
		for month := 1; month <= 12; month++ {
			at := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, cst)
			ours := DayGanZhi(year, month, 15).String()
			theirs := solarFor(at).GetLunar().GetDayInGanZhiExact()
			if ours != theirs {
				t.Fatalf("%d-%02d-15 日柱不一致：本包 %s，lunar-go %s", year, month, ours, theirs)
			}
		}
	}
}
