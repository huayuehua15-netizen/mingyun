package schema

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
)

func TestParseMethod(t *testing.T) {
	for _, s := range []string{"liuyao", "daliuren", "ziwei"} {
		m, err := ParseMethod(s)
		if err != nil {
			t.Errorf("%s 应可解析: %v", s, err)
		}
		if string(m) != s {
			t.Errorf("解析结果应为 %s，实为 %s", s, m)
		}
	}
	if _, err := ParseMethod("qimen"); err == nil {
		t.Error("qimen 尚未支持，应报错")
	}
}

func TestChartValidate(t *testing.T) {
	valid := Chart{
		Method:        MethodLiuYao,
		EngineVersion: "liuyao@0.1.0",
		Payload:       map[string]string{"gua": "乾为天"},
		Facts: []Fact{
			{Key: "liuyao.gua", Label: "卦名", Value: "乾为天", Anchor: "/payload/gua", Weight: WeightCore},
		},
		CreatedAt: time.Now(),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("合法盘面不应报错: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(c Chart) Chart
		want   string
	}{
		{"非法术数", func(c Chart) Chart { c.Method = "qimen"; return c }, "method"},
		{"缺版本", func(c Chart) Chart { c.EngineVersion = ""; return c }, "engine_version"},
		{"缺盘面", func(c Chart) Chart { c.Payload = nil; return c }, "payload"},
		{"空事实", func(c Chart) Chart { c.Facts = nil; return c }, "facts"},
		{"锚点非指针", func(c Chart) Chart {
			c.Facts = []Fact{{Key: "k", Label: "l", Value: "v", Anchor: "payload/gua", Weight: WeightCore}}
			return c
		}, "JSON Pointer"},
		{"权重非法", func(c Chart) Chart {
			c.Facts = []Fact{{Key: "k", Label: "l", Value: "v", Weight: "high"}}
			return c
		}, "weight"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mutate(valid).Validate()
			if err == nil {
				t.Fatalf("%s 应校验失败", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("错误信息应包含 %q，实为 %q", tc.want, err.Error())
			}
		})
	}
}

func TestCapabilitiesCoverAllMethods(t *testing.T) {
	seen := map[Method]bool{}
	for _, c := range Capabilities {
		if !c.Method.Valid() {
			t.Errorf("能力清单含非法术数 %q", c.Method)
		}
		if c.Name == "" || len(c.Modes) == 0 {
			t.Errorf("%s 的能力描述不完整", c.Method)
		}
		seen[c.Method] = true
	}
	for _, m := range []Method{MethodLiuYao, MethodDaLiuRen, MethodZiWei} {
		if !seen[m] {
			t.Errorf("能力清单缺少 %s", m)
		}
	}
}

func TestChartJSONKeepsGanZhiReadable(t *testing.T) {
	c := Chart{
		Method:        MethodZiWei,
		EngineVersion: "ziwei@0.1.0",
		Calendar:      calendar.Context{},
		Payload:       map[string]string{"ming_gong": "子"},
		Facts:         []Fact{{Key: "k", Label: "l", Value: "v", Weight: WeightNormal}},
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"method":"ziwei"`) {
		t.Errorf("method 字段应出现在 JSON 中: %s", data)
	}
}
