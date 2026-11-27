package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

func testRecord(id string) Record {
	return Record{
		ChartID:   id,
		Method:    schema.MethodLiuYao,
		Question:  "这次考试能否通过",
		Summary:   "六爻 · 泽火革",
		CreatedAt: time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
		Chart: schema.Chart{
			ChartID: id,
			Method:  schema.MethodLiuYao,
			Payload: map[string]any{"gua_name": "泽火革"},
			Facts: []schema.Fact{
				{Key: "liuyao.gua", Label: "本卦", Value: "泽火革", Weight: schema.WeightCore},
			},
		},
	}
}

func TestFileStorePersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "charts.json")
	f, err := NewFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Save(testRecord("c1")); err != nil {
		t.Fatal(err)
	}
	if err := f.Save(testRecord("c2")); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Len() != 2 {
		t.Fatalf("reload len = %d", reloaded.Len())
	}
	rec, ok := reloaded.Get("c1")
	if !ok || rec.Question != "这次考试能否通过" {
		t.Fatalf("reload c1 = %+v, ok=%v", rec, ok)
	}
	list, total := reloaded.List(ListOptions{Limit: 10})
	if total != 2 || len(list) != 2 {
		t.Fatalf("list total=%d len=%d", total, len(list))
	}
	if list[0].ChartID != "c2" {
		t.Fatalf("列表应按新到旧，首条 = %s", list[0].ChartID)
	}
}

func TestFileStoreRejectsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "charts.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFile(path); err == nil {
		t.Fatal("损坏的盘面文件应报错")
	}
}
