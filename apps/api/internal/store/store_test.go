package store

import (
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

func rec(id string, method schema.Method, at time.Time) Record {
	return Record{
		ChartID:   id,
		Method:    method,
		Question:  "q-" + id,
		Summary:   "s-" + id,
		CreatedAt: at,
		Chart:     schema.Chart{Method: method, EngineVersion: "test@0.1.0", Payload: struct{}{}},
	}
}

func TestMemorySaveGet(t *testing.T) {
	m := NewMemory()
	if err := m.Save(rec("a", schema.MethodLiuYao, time.Now())); err != nil {
		t.Fatal(err)
	}
	got, ok := m.Get("a")
	if !ok || got.ChartID != "a" || got.Question != "q-a" {
		t.Fatalf("取回记录不符: %+v, ok=%v", got, ok)
	}
	if _, ok := m.Get("missing"); ok {
		t.Fatal("不存在的记录不应命中")
	}
}

func TestMemorySaveDuplicate(t *testing.T) {
	m := NewMemory()
	r := rec("a", schema.MethodLiuYao, time.Now())
	if err := m.Save(r); err != nil {
		t.Fatal(err)
	}
	if err := m.Save(r); err == nil {
		t.Fatal("重复 chart_id 应当报错")
	}
	if err := m.Save(rec("", schema.MethodLiuYao, time.Now())); err == nil {
		t.Fatal("空 chart_id 应当报错")
	}
}

func TestMemoryList(t *testing.T) {
	m := NewMemory()
	base := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	m.Save(rec("a", schema.MethodLiuYao, base))
	m.Save(rec("b", schema.MethodLiuYao, base.Add(time.Minute)))
	m.Save(rec("c", schema.MethodDaLiuRen, base.Add(2*time.Minute)))

	got, total := m.List(ListOptions{Limit: 2})
	if total != 3 || len(got) != 2 {
		t.Fatalf("total=%d len=%d", total, len(got))
	}
	if got[0].ChartID != "c" || got[1].ChartID != "b" {
		t.Fatalf("应按创建时间倒序: %v", []string{got[0].ChartID, got[1].ChartID})
	}

	got, total = m.List(ListOptions{Method: "liuyao", Limit: 10})
	if total != 2 || len(got) != 2 {
		t.Fatalf("按术数过滤 total=%d len=%d", total, len(got))
	}

	got, total = m.List(ListOptions{Offset: 1, Limit: 10})
	if total != 3 || len(got) != 2 || got[0].ChartID != "b" {
		t.Fatalf("游标偏移结果不符: total=%d len=%d first=%s", total, len(got), got[0].ChartID)
	}
	if m.Len() != 3 {
		t.Fatalf("Len = %d", m.Len())
	}
}
