package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

// 旧盘追问的关键回归：前端只拿到 core facts，服务端内存丢盘后必须能从
// payload 重建普通爻的纳甲、六亲、六神等完整事实，否则合法回答会被审计误杀。
func TestRestoreChartFactsFromSnapshot(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","question":"旧盘追问回归","at":"2026-11-25T14:30:00+08:00","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}

	snapshot := schema.Chart{
		ChartID:       view.ChartID,
		Method:        view.Method,
		EngineVersion: view.EngineVersion,
		Question:      view.Question,
		Calendar:      view.Calendar,
		School:        view.School,
		Payload:       view.Payload,
		Facts:         view.Facts,
		Warnings:      view.Warnings,
		CreatedAt:     view.CreatedAt,
	}
	if err := restoreChartFacts(&snapshot); err != nil {
		t.Fatalf("restoreChartFacts: %v", err)
	}
	if len(snapshot.Facts) <= len(view.Facts) {
		t.Fatalf("重建 facts = %d 条，未多于原有 %d 条 core facts", len(snapshot.Facts), len(view.Facts))
	}
	found := false
	for _, f := range snapshot.Facts {
		if f.Key == "liuyao.line.2" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("重建 facts 缺少二爻的普通爻事实")
	}
}

func TestRestoreChartFactsRejectsEmptySnapshot(t *testing.T) {
	chart := schema.Chart{Payload: map[string]any{"gua_name": "", "lines": []any{}}}
	if err := restoreChartFacts(&chart); err == nil {
		t.Fatal("空快照应返回错误，而不是伪造 facts")
	}
}
