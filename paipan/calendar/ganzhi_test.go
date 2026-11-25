package calendar

import (
	"encoding/json"
	"testing"
)

func TestGanZhiFromIndexRoundTrip(t *testing.T) {
	for i := 0; i < 60; i++ {
		gz := GanZhiFromIndex(i)
		if got := gz.Index(); got != i {
			t.Fatalf("index %d 往返失败，得到 %d（%s）", i, got, gz)
		}
	}
}

func TestGanZhiKnownPoints(t *testing.T) {
	cases := []struct {
		index int
		want  string
	}{
		{0, "甲子"},
		{1, "乙丑"},
		{10, "甲戌"},
		{21, "乙酉"},
		{54, "戊午"},
		{59, "癸亥"},
	}
	for _, c := range cases {
		gz := GanZhiFromIndex(c.index)
		if gz.String() != c.want {
			t.Errorf("序号 %d 应为 %s，实为 %s", c.index, c.want, gz)
		}
	}
}

func TestGanZhiFromIndexWraps(t *testing.T) {
	if got := GanZhiFromIndex(60); got.String() != "甲子" {
		t.Errorf("60 应回绕为甲子，实为 %s", got)
	}
	if got := GanZhiFromIndex(-1); got.String() != "癸亥" {
		t.Errorf("-1 应回绕为癸亥，实为 %s", got)
	}
}

func TestNewGanZhiRejectsMismatch(t *testing.T) {
	// 甲为阳干，丑为阴支，阴阳不匹配。
	if _, err := NewGanZhi(GanJia, ZhiChou); err == nil {
		t.Fatal("甲丑应被判为非法组合")
	}
	// 甲子合法。
	if _, err := NewGanZhi(GanJia, ZhiZi); err != nil {
		t.Fatalf("甲子应为合法组合: %v", err)
	}
}

func TestGanZhiJSON(t *testing.T) {
	gz := MustGanZhi(GanJia, ZhiZi)
	data, err := json.Marshal(gz)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"甲子"` {
		t.Errorf("干支 JSON 应为 \"甲子\"，实为 %s", string(data))
	}

	var back GanZhi
	if err := json.Unmarshal([]byte(`"癸亥"`), &back); err != nil {
		t.Fatal(err)
	}
	if back.String() != "癸亥" {
		t.Errorf("反序列化失败，得到 %s", back)
	}
}

func TestWuXingRelations(t *testing.T) {
	// 相生：木→火→土→金→水→木
	for w := WuXingMu; w <= WuXingShui; w++ {
		if !w.Generates((w + 1) % wuXingCount) {
			t.Errorf("%s 应生 %s", w, (w+1)%wuXingCount)
		}
	}
	// 相克：隔一
	for w := WuXingMu; w <= WuXingShui; w++ {
		if !w.Overcomes((w + 2) % wuXingCount) {
			t.Errorf("%s 应克 %s", w, (w+2)%wuXingCount)
		}
	}
	if WuXingMu.Generates(WuXingTu) {
		t.Error("木不应生土")
	}
	if WuXingMu.Overcomes(WuXingHuo) {
		t.Error("木不应克火")
	}
}

func TestZhiRelations(t *testing.T) {
	if !ZhiChong(ZhiZi, ZhiWu) {
		t.Error("子午应相冲")
	}
	if ZhiChong(ZhiZi, ZhiChou) {
		t.Error("子丑不应相冲")
	}
	if !ZhiLiuHe(ZhiZi, ZhiChou) {
		t.Error("子丑应六合")
	}
	if !ZhiLiuHe(ZhiWu, ZhiWei) {
		t.Error("午未应六合")
	}
	if ZhiLiuHe(ZhiZi, ZhiYin) {
		t.Error("子寅不应六合")
	}
	if w, ok := SanHeWuXing(ZhiShen); !ok || w != WuXingShui {
		t.Error("申应属申子辰水局")
	}
	if w, ok := SanHeWuXing(ZhiMao); !ok || w != WuXingMu {
		t.Error("卯应属亥卯未木局")
	}
}

func TestNaYin(t *testing.T) {
	cases := []struct {
		gz   string
		want string
	}{
		{"甲子", "海中金"},
		{"乙丑", "海中金"},
		{"丙寅", "炉中火"},
		{"庚午", "路旁土"},
		{"壬戌", "大海水"},
	}
	for _, c := range cases {
		gz, err := ParseGanZhi(c.gz)
		if err != nil {
			t.Fatal(err)
		}
		if got := NaYin(gz); got != c.want {
			t.Errorf("%s 纳音应为 %s，实为 %s", c.gz, c.want, got)
		}
	}
	// 纳音名末字即五行
	if w, ok := NaYinWuXing(MustGanZhi(GanJia, ZhiZi)); !ok || w != WuXingJin {
		t.Errorf("甲子纳音应属金，实为 %v", w)
	}
}
