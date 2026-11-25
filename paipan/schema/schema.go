// Package schema 定义排盘引擎与外部世界之间的数据契约。
//
// 该包只做类型定义与校验，不做任何计算。所有术数的输出都必须满足 Chart
// 的结构，AI 与前端只依赖这里定义的字段。
package schema

import (
	"fmt"
	"strings"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

// Method 术数类型。
type Method string

const (
	MethodLiuYao   Method = "liuyao"
	MethodDaLiuRen Method = "daliuren"
	MethodZiWei    Method = "ziwei"
)

// Valid 判断是否为已支持的术数。
func (m Method) Valid() bool {
	switch m {
	case MethodLiuYao, MethodDaLiuRen, MethodZiWei:
		return true
	}
	return false
}

// ParseMethod 解析术数名。
func ParseMethod(s string) (Method, error) {
	m := Method(s)
	if !m.Valid() {
		return "", fmt.Errorf("不支持的术数: %q", s)
	}
	return m, nil
}

// Weight 事实权重。注入提示词时按权重裁剪。
type Weight string

const (
	// WeightCore 核心事实，永不裁剪。
	WeightCore Weight = "core"
	// WeightNormal 一般事实，预算紧张时保留。
	WeightNormal Weight = "normal"
	// WeightMinor 次要事实，优先裁剪。
	WeightMinor Weight = "minor"
)

// Valid 判断权重是否合法。
func (w Weight) Valid() bool {
	return w == WeightCore || w == WeightNormal || w == WeightMinor
}

// Fact 盘面事实。它是排盘结果与 AI 之间的唯一桥梁。
//
// Fact 写得好，提示词就短；Fact 写得含糊，模型就会开始自由发挥。
type Fact struct {
	// Key 稳定的机器标识，如 liuyao.line.3.najia。
	Key string `json:"key"`
	// Label 中文可读标签，如 "三爻纳甲"。
	Label string `json:"label"`
	// Value 已格式化的值，如 "辰土·父母·青龙·静"。
	Value string `json:"value"`
	// Anchor 指回 payload 的 JSON Pointer，用于前端高亮。
	Anchor string `json:"anchor,omitempty"`
	// Weight 权重。
	Weight Weight `json:"weight"`
}

// Validate 校验事实的完整性。
func (f Fact) Validate() error {
	if f.Key == "" {
		return fmt.Errorf("事实缺少 key")
	}
	if f.Label == "" {
		return fmt.Errorf("事实 %s 缺少 label", f.Key)
	}
	if f.Anchor != "" && !strings.HasPrefix(f.Anchor, "/") {
		return fmt.Errorf("事实 %s 的 anchor 必须是 JSON Pointer（以 / 开头），实为 %q", f.Key, f.Anchor)
	}
	if !f.Weight.Valid() {
		return fmt.Errorf("事实 %s 的 weight 非法: %q", f.Key, f.Weight)
	}
	return nil
}

// Chart 一次排盘的完整输出。
type Chart struct {
	ChartID       string            `json:"chart_id,omitempty"`
	Method        Method            `json:"method"`
	EngineVersion string            `json:"engine_version"`
	Calendar      calendar.Context  `json:"calendar"`
	Question      string            `json:"question,omitempty"`
	School        map[string]string `json:"school,omitempty"`
	Payload       any               `json:"payload"`
	Facts         []Fact            `json:"facts"`
	Warnings      []string          `json:"warnings,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// Validate 校验盘面契约。引擎在返回结果前必须调用它。
func (c Chart) Validate() error {
	if !c.Method.Valid() {
		return fmt.Errorf("非法的 method: %q", c.Method)
	}
	if c.EngineVersion == "" {
		return fmt.Errorf("缺少 engine_version")
	}
	if c.Payload == nil {
		return fmt.Errorf("缺少 payload")
	}
	if len(c.Facts) == 0 {
		return fmt.Errorf("facts 不能为空：AI 没有事实可依据")
	}
	for _, f := range c.Facts {
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Capability 描述一种术数的能力与输入要求，供前端动态渲染表单。
type Capability struct {
	Method   Method   `json:"method"`
	Name     string   `json:"name"`
	Modes    []string `json:"modes"`
	Requires []string `json:"requires"`
}

// Capabilities 已支持术数的能力清单。
var Capabilities = []Capability{
	{
		Method:   MethodLiuYao,
		Name:     "六爻",
		Modes:    []string{"coin", "time", "manual"},
		Requires: []string{"question"},
	},
	{
		Method:   MethodDaLiuRen,
		Name:     "大六壬",
		Modes:    []string{"time"},
		Requires: []string{"at", "question"},
	},
	{
		Method:   MethodZiWei,
		Name:     "紫微斗数",
		Modes:    []string{"birth"},
		Requires: []string{"gender", "birth_date", "birth_time"},
	},
}
