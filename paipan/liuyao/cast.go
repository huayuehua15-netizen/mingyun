package liuyao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// Version 六爻引擎版本。规则变更时升 minor，盘面结构变更时升 major。
const Version = "liuyao@0.1.0"

// Mode 起卦方式。
type Mode string

const (
	// ModeManual 手动录入六爻。
	ModeManual Mode = "manual"
	// ModeCoin 铜钱摇卦，由调用方提供六组三枚铜钱的背数。
	ModeCoin Mode = "coin"
)

// Input 起卦输入。
type Input struct {
	Mode     Mode
	Lines    []LineKind // ModeManual 时使用，自下而上六爻
	Backs    [6]int     // ModeCoin 时使用，每爻三枚铜钱的背数（0—3）
	Question string
}

// LinePayload 一爻的盘面数据。
type LinePayload struct {
	Position int          `json:"position"`
	Kind     string       `json:"kind"`
	YinYang  string       `json:"yin_yang"`
	Moving   bool         `json:"moving"`
	Gan      string       `json:"gan"`
	Zhi      string       `json:"zhi"`
	WuXing   string       `json:"wu_xing"`
	LiuQin   string       `json:"liu_qin"`
	LiuShen  string       `json:"liu_shen"`
	Shi      bool         `json:"shi,omitempty"`
	Ying     bool         `json:"ying,omitempty"`
	Marks    []string     `json:"marks,omitempty"`
	Change   *ChangedLine `json:"change,omitempty"`
}

// ChangedLine 动爻变化后的爻。
// 六亲默认以本卦卦宫为“我”来定（流派开关 school.changed_liuqin=ben_gong）。
type ChangedLine struct {
	Kind     string `json:"kind"`
	Gan      string `json:"gan"`
	Zhi      string `json:"zhi"`
	WuXing   string `json:"wu_xing"`
	LiuQin   string `json:"liu_qin"`
	Relation string `json:"relation,omitempty"`
}

// FuShenPayload 伏神。
type FuShenPayload struct {
	UnderPosition int    `json:"under_position"`
	LiuQin        string `json:"liu_qin"`
	Gan           string `json:"gan"`
	Zhi           string `json:"zhi"`
	WuXing        string `json:"wu_xing"`
}

// Payload 六爻盘面。
type Payload struct {
	Question      string          `json:"question,omitempty"`
	GuaName       string          `json:"gua_name"`
	Upper         string          `json:"upper"`
	Lower         string          `json:"lower"`
	Palace        string          `json:"palace"`
	PalaceWuXing  string          `json:"palace_wuxing"`
	Rank          string          `json:"rank"`
	ShiYao        int             `json:"shi_yao"`
	YingYao       int             `json:"ying_yao"`
	Lines         []LinePayload   `json:"lines"`
	ChangedGua    string          `json:"changed_gua,omitempty"`
	ChangedPalace string          `json:"changed_palace,omitempty"`
	FuShen        []FuShenPayload `json:"fu_shen,omitempty"`
	MonthGanZhi   string          `json:"month_gz"`
	DayGanZhi     string          `json:"day_gz"`
	MonthZhi      string          `json:"month_zhi"`
	DayZhi        string          `json:"day_zhi"`
	XunKong       [2]string       `json:"xun_kong"`
}

// Cast 排六爻盘。
func Cast(in Input, cal calendar.Context) (schema.Chart, error) {
	kinds, err := in.lineKinds()
	if err != nil {
		return schema.Chart{}, err
	}

	var yangs, moving [6]bool
	anyMoving := false
	for i, k := range kinds {
		yangs[i] = k.Yang()
		moving[i] = k.Moving()
		if moving[i] {
			anyMoving = true
		}
	}

	ben, ok := FromLines(yangs)
	if !ok {
		return schema.Chart{}, fmt.Errorf("无法识别本卦")
	}

	var bian Hexagram
	if anyMoving {
		bian, ok = ben.Changed(moving)
		if !ok {
			return schema.Chart{}, fmt.Errorf("无法识别变卦")
		}
	}

	palaceWu := ben.Palace.WuXing()
	monthZhi := cal.Month.Zhi
	dayZhi := cal.Day.Zhi
	kongA, kongB := cal.XunKong[0], cal.XunKong[1]

	payload := Payload{
		Question:     in.Question,
		GuaName:      ben.Name,
		Upper:        ben.Upper.String(),
		Lower:        ben.Lower.String(),
		Palace:       ben.PalaceLabel(),
		PalaceWuXing: palaceWu.String(),
		Rank:         ben.RankLabel(),
		ShiYao:       ben.ShiYao(),
		YingYao:      ben.YingYao(),
		Lines:        make([]LinePayload, 0, 6),
		MonthGanZhi:  cal.Month.String(),
		DayGanZhi:    cal.Day.String(),
		MonthZhi:     monthZhi.String(),
		DayZhi:       dayZhi.String(),
	}
	if anyMoving {
		payload.ChangedGua = bian.Name
		payload.ChangedPalace = bian.PalaceLabel()
	}
	k1, k2 := cal.XunKong[0].String(), cal.XunKong[1].String()
	payload.XunKong = [2]string{k1, k2}

	for pos := 1; pos <= 6; pos++ {
		tri, yao := ben.Lower, pos
		outer := false
		if pos > 3 {
			tri, yao, outer = ben.Upper, pos-3, true
		}
		gan, zhi, err := tri.NaJiaAt(yao, outer)
		if err != nil {
			return schema.Chart{}, fmt.Errorf("装第 %d 爻: %w", pos, err)
		}
		wu := zhi.WuXing()
		lq, err := LiuQinFor(palaceWu, wu)
		if err != nil {
			return schema.Chart{}, fmt.Errorf("定第 %d 爻六亲: %w", pos, err)
		}

		line := LinePayload{
			Position: pos,
			Kind:     kinds[pos-1].String(),
			YinYang:  kinds[pos-1].YangString(),
			Moving:   moving[pos-1],
			Gan:      gan.String(),
			Zhi:      zhi.String(),
			WuXing:   wu.String(),
			LiuQin:   lq.String(),
			Marks:    marksFor(zhi, monthZhi, dayZhi, kongA, kongB),
			Shi:      pos == ben.ShiYao(),
			Ying:     pos == ben.YingYao(),
		}
		ls, err := LiuShenAt(cal.Day.Gan, pos)
		if err != nil {
			return schema.Chart{}, fmt.Errorf("装第 %d 爻六神: %w", pos, err)
		}
		line.LiuShen = ls.String()

		if moving[pos-1] {
			tri2, yao2 := bian.Lower, pos
			outer2 := false
			if pos > 3 {
				tri2, yao2, outer2 = bian.Upper, pos-3, true
			}
			gan2, zhi2, err := tri2.NaJiaAt(yao2, outer2)
			if err != nil {
				return schema.Chart{}, fmt.Errorf("装第 %d 爻变爻: %w", pos, err)
			}
			wu2 := zhi2.WuXing()
			lq2, err := LiuQinFor(palaceWu, wu2)
			if err != nil {
				return schema.Chart{}, fmt.Errorf("定第 %d 爻变爻六亲: %w", pos, err)
			}
			line.Change = &ChangedLine{
				Kind:     kinds[pos-1].Changed().String(),
				Gan:      gan2.String(),
				Zhi:      zhi2.String(),
				WuXing:   wu2.String(),
				LiuQin:   lq2.String(),
				Relation: relationFor(wu, wu2),
			}
		}

		payload.Lines = append(payload.Lines, line)
	}

	payload.FuShen = fuShenFor(ben, palaceWu)

	chart := schema.Chart{
		Method:        schema.MethodLiuYao,
		EngineVersion: Version,
		Calendar:      cal,
		Question:      in.Question,
		School: map[string]string{
			"changed_liuqin": "ben_gong",
			"fushen":         "single",
		},
		Payload:   payload,
		Facts:     buildFacts(payload),
		CreatedAt: time.Now().UTC(),
	}
	if err := chart.Validate(); err != nil {
		return schema.Chart{}, err
	}
	return chart, nil
}

// YangString 返回 "阳" 或 "阴"。
func (k LineKind) YangString() string {
	if k.Yang() {
		return "阳"
	}
	return "阴"
}

func (in Input) lineKinds() ([6]LineKind, error) {
	var out [6]LineKind
	switch in.Mode {
	case ModeManual:
		if len(in.Lines) != 6 {
			return out, fmt.Errorf("手动录入需提供 6 爻，实为 %d", len(in.Lines))
		}
		for i, k := range in.Lines {
			if !k.Valid() {
				return out, fmt.Errorf("第 %d 爻爻形非法", i+1)
			}
			out[i] = k
		}
	case ModeCoin:
		for i := 0; i < 6; i++ {
			k, err := LineKindFromCoins(in.Backs[i])
			if err != nil {
				return out, fmt.Errorf("第 %d 爻: %w", i+1, err)
			}
			out[i] = k
		}
	default:
		return out, fmt.Errorf("不支持的起卦方式: %q", in.Mode)
	}
	return out, nil
}

func marksFor(zhi, month, day, kongA, kongB calendar.Zhi) []string {
	marks := make([]string, 0, 6)
	w, mw, dw := zhi.WuXing(), month.WuXing(), day.WuXing()

	switch {
	case w == mw:
		marks = append(marks, "旺")
	case mw.Generates(w):
		marks = append(marks, "相")
	case w.Generates(mw):
		marks = append(marks, "休")
	case w.Overcomes(mw):
		marks = append(marks, "囚")
	case mw.Overcomes(w):
		marks = append(marks, "死")
	}

	switch {
	case dw.Generates(w):
		marks = append(marks, "日生")
	case dw == w:
		marks = append(marks, "日扶")
	case dw.Overcomes(w):
		marks = append(marks, "日克")
	}
	if calendar.ZhiChong(day, zhi) {
		marks = append(marks, "日冲")
	}
	if calendar.ZhiLiuHe(day, zhi) {
		marks = append(marks, "日合")
	}
	if calendar.ZhiChong(month, zhi) {
		marks = append(marks, "月破")
	}
	if zhi == kongA || zhi == kongB {
		marks = append(marks, "旬空")
	}
	return marks
}

func relationFor(original, changed calendar.WuXing) string {
	switch {
	case changed.Generates(original):
		return "回头生"
	case changed.Overcomes(original):
		return "回头克"
	default:
		return ""
	}
}

var pureHexagramNames = [8]string{"乾为天", "兑为泽", "离为火", "震为雷", "巽为风", "坎为水", "艮为山", "坤为地"}

// fuShenFor 取本宫首卦中、本卦所缺六亲的同位爻为伏神。
// 同一六亲只取最低爻位的一支（school.fushen=single）。
func fuShenFor(h Hexagram, palaceWu calendar.WuXing) []FuShenPayload {
	pure, ok := FromName(pureHexagramNames[h.Palace])
	if !ok {
		return nil
	}
	if h.Name == pure.Name {
		return nil
	}

	present := make(map[LiuQin]bool, 6)
	for pos := 1; pos <= 6; pos++ {
		_, zhi := naJiaAtPos(h, pos)
		wu := zhi.WuXing()
		lq, err := LiuQinFor(palaceWu, wu)
		if err == nil {
			present[lq] = true
		}
	}

	out := make([]FuShenPayload, 0, 2)
	used := make(map[LiuQin]bool, 2)
	for pos := 1; pos <= 6; pos++ {
		gan, zhi := naJiaAtPos(pure, pos)
		wu := zhi.WuXing()
		lq, err := LiuQinFor(palaceWu, wu)
		if err != nil || present[lq] || used[lq] {
			continue
		}
		used[lq] = true
		out = append(out, FuShenPayload{
			UnderPosition: pos,
			LiuQin:        lq.String(),
			Gan:           gan.String(),
			Zhi:           zhi.String(),
			WuXing:        wu.String(),
		})
	}
	return out
}

func naJiaAtPos(h Hexagram, pos int) (calendar.Gan, calendar.Zhi) {
	tri, yao := h.Lower, pos
	outer := false
	if pos > 3 {
		tri, yao, outer = h.Upper, pos-3, true
	}
	gan, zhi, _ := tri.NaJiaAt(yao, outer)
	return gan, zhi
}

// Summary 返回一行摘要，用于 CLI 与日志。
func (p Payload) Summary() string {
	parts := []string{
		fmt.Sprintf("%s（%s·%s）", p.GuaName, p.Palace, p.Rank),
		fmt.Sprintf("月建%s 日辰%s", p.MonthGanZhi, p.DayGanZhi),
		fmt.Sprintf("旬空%s%s", p.XunKong[0], p.XunKong[1]),
	}
	if p.ChangedGua != "" {
		parts = append(parts, "变"+p.ChangedGua)
	}
	return strings.Join(parts, "，")
}
