package calendar

import (
	"fmt"
	"time"

	lunar "github.com/6tail/lunar-go/calendar"
)

// cst 北京时间。节气时刻与年月柱的判定以北京时间为基准：
// 节气是全球同一时刻，用固定时区表示只是取值方便。
var cst = time.FixedZone("CST", 8*3600)

// Term 二十四节气之一。
type Term struct {
	Name  string    `json:"name"`
	At    time.Time `json:"at"`
	IsJie bool      `json:"is_jie"`
}

// String 返回 "立春 2026-02-04 04:02:08" 形式的描述。
func (t Term) String() string {
	return fmt.Sprintf("%s %s", t.Name, t.At.Format("2006-01-02 15:04:05"))
}

func solarFor(t time.Time) *lunar.Solar {
	bj := t.In(cst)
	return lunar.NewSolar(bj.Year(), int(bj.Month()), bj.Day(), bj.Hour(), bj.Minute(), bj.Second())
}

func instantOf(s *lunar.Solar) time.Time {
	return time.Date(s.GetYear(), time.Month(s.GetMonth()), s.GetDay(),
		s.GetHour(), s.GetMinute(), s.GetSecond(), 0, cst)
}

func termOf(jq *lunar.JieQi) Term {
	return Term{Name: jq.GetName(), At: instantOf(jq.GetSolar()), IsJie: jq.IsJie()}
}

// PrevTerm 返回 t 之前（含当刻）最近的节气。
func PrevTerm(t time.Time) (Term, error) {
	jq := solarFor(t).GetLunar().GetPrevJieQi()
	if jq == nil || jq.GetSolar() == nil {
		return Term{}, fmt.Errorf("无法确定 %s 之前的节气", t.Format(time.RFC3339))
	}
	return termOf(jq), nil
}

// NextTerm 返回 t 之后最近的节气。
func NextTerm(t time.Time) (Term, error) {
	jq := solarFor(t).GetLunar().GetNextJieQi()
	if jq == nil || jq.GetSolar() == nil {
		return Term{}, fmt.Errorf("无法确定 %s 之后的节气", t.Format(time.RFC3339))
	}
	return termOf(jq), nil
}

// YearPillar 返回年柱，以立春交节时刻换年。
func YearPillar(t time.Time) (GanZhi, error) {
	return ParseGanZhi(solarFor(t).GetLunar().GetYearInGanZhiExact())
}

// MonthPillar 返回月柱，以节气交节时刻换月。
func MonthPillar(t time.Time) (GanZhi, error) {
	return ParseGanZhi(solarFor(t).GetLunar().GetMonthInGanZhiExact())
}

// MonthGeneral 月将。大六壬起课用，由中气换将。
type MonthGeneral struct {
	Zhi       Zhi    `json:"zhi"`
	Name      string `json:"name"`
	AfterTerm string `json:"after_term"`
}

// String 返回 "寅将功曹（小雪后）" 形式。
func (m MonthGeneral) String() string {
	return fmt.Sprintf("%s将%s（%s后）", m.Zhi, m.Name, m.AfterTerm)
}

// generalTable 中气换将表。键为中气名，值为月将地支与将名。
var generalTable = map[string]struct {
	zhi  Zhi
	name string
}{
	"雨水": {ZhiHai, "登明"},
	"春分": {ZhiXu, "河魁"},
	"谷雨": {ZhiYou, "从魁"},
	"小满": {ZhiShen, "传送"},
	"夏至": {ZhiWei, "小吉"},
	"大暑": {ZhiWu, "胜光"},
	"处暑": {ZhiSi, "太乙"},
	"秋分": {ZhiChen, "天罡"},
	"霜降": {ZhiMao, "太冲"},
	"小雪": {ZhiYin, "功曹"},
	"冬至": {ZhiChou, "大吉"},
	"大寒": {ZhiZi, "神后"},
}

// MonthGeneralAt 返回 t 时刻所用月将（取最近一个中气之后）。
func MonthGeneralAt(t time.Time) (MonthGeneral, error) {
	q := solarFor(t).GetLunar().GetPrevQi()
	if q == nil {
		return MonthGeneral{}, fmt.Errorf("无法确定 %s 之前的节中气", t.Format(time.RFC3339))
	}
	entry, ok := generalTable[q.GetName()]
	if !ok {
		return MonthGeneral{}, fmt.Errorf("未识别的中气: %s", q.GetName())
	}
	return MonthGeneral{Zhi: entry.zhi, Name: entry.name, AfterTerm: q.GetName()}, nil
}
