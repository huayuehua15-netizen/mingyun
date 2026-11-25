package calendar

import (
	"math"
	"testing"
	"time"
)

func TestLongitudeOffset(t *testing.T) {
	// 东八区中央经线 120°E，广州 113.26°E。
	got := LongitudeOffset(113.26, 8)
	want := time.Duration(-26.96 * float64(time.Minute))
	if math.Abs(float64(got-want)) > float64(time.Second) {
		t.Errorf("经度修正应为 %s，实为 %s", want, got)
	}

	// 恰在中央经线上，修正为零。
	if off := LongitudeOffset(120, 8); off != 0 {
		t.Errorf("120°E 在东八区应无经度修正，实为 %s", off)
	}

	// 每偏 1 度 = 4 分钟。
	if off := LongitudeOffset(121, 8); off != 4*time.Minute {
		t.Errorf("偏东 1 度应为 4 分钟，实为 %s", off)
	}
}

func TestEquationOfTimeSeasonal(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	// 1 月初均时差约为 -3 分钟。
	jan := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)
	eot := EquationOfTime(jan)
	if eot > 0 || eot < -6*time.Minute {
		t.Errorf("1 月 1 日均时差应约为 -3 分钟，实为 %s", eot)
	}

	// 11 月初均时差约为 +16 分钟（全年最大）。
	nov := time.Date(2024, 11, 3, 12, 0, 0, 0, loc)
	eot = EquationOfTime(nov)
	if eot < 12*time.Minute || eot > 19*time.Minute {
		t.Errorf("11 月 3 日均时差应约为 +16 分钟，实为 %s", eot)
	}

	// 2 月中约为 -14 分钟（全年最小）。
	feb := time.Date(2024, 2, 11, 12, 0, 0, 0, loc)
	eot = EquationOfTime(feb)
	if eot > -11*time.Minute || eot < -17*time.Minute {
		t.Errorf("2 月 11 日均时差应约为 -14 分钟，实为 %s", eot)
	}
}

func TestTrueSolarTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2024, 6, 1, 12, 0, 0, 0, loc)

	// 位于中央经线时，真太阳时 = 标准时 + 均时差（结果截断到秒，允许 1 秒误差）。
	got := TrueSolarTime(base, 120, 8)
	diff := got.Sub(base)
	eot := EquationOfTime(base)
	if d := diff - eot; d > time.Second || d < -time.Second {
		t.Errorf("中央经线上真太阳时差应等于均时差 %s，实为 %s", eot, diff)
	}

	// 广州应比标准时早约半小时（经度修正为主）。
	guangzhou := TrueSolarTime(base, 113.26, 8)
	if diff := guangzhou.Sub(base); diff > -20*time.Minute {
		t.Errorf("广州真太阳时应比标准时早至少 20 分钟，实为 %s", diff)
	}
}
