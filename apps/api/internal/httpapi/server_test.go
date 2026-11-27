package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
)

var testNow = time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)

type testEnv struct {
	srv *httptest.Server
	src *entropy.Source
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	seq := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	r := bytes.NewReader(bytes.Repeat(seq, 64))
	src := entropy.NewSourceWithReader(bytes.Repeat([]byte{0x2a}, 32), r, func() time.Time { return testNow })
	srv := httptest.NewServer(New("0.1.0-test", store.NewMemory(), src))
	t.Cleanup(srv.Close)
	return &testEnv{srv: srv, src: src}
}

type testEnvelope struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data"`
	RequestID  string          `json:"request_id"`
	ServerTime string          `json:"server_time"`
}

func doRequest(t *testing.T, ts *testEnv, method, path, body string, wantStatus int) testEnvelope {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, ts.srv.URL+path, rdr)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s 状态码 = %d，期望 %d", method, path, resp.StatusCode, wantStatus)
	}
	if id := resp.Header.Get("X-Request-Id"); id == "" {
		t.Error("响应应带 X-Request-Id")
	}
	var env testEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	return env
}

func TestHealthz(t *testing.T) {
	ts := newTestEnv(t)
	resp, err := http.Get(ts.srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" || body.Version != "0.1.0-test" {
		t.Errorf("健康检查响应不符: %+v", body)
	}
}

func TestNotFoundReturnsJSON(t *testing.T) {
	ts := newTestEnv(t)
	env := doRequest(t, ts, http.MethodGet, "/api/v1/nope", "", http.StatusNotFound)
	if env.Code != 10301 {
		t.Errorf("错误码应为 10301，实为 %d", env.Code)
	}
}

func TestRequestIDPassthrough(t *testing.T) {
	ts := newTestEnv(t)
	req, err := http.NewRequest(http.MethodGet, ts.srv.URL+"/api/v1/meta/capabilities", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Request-Id", "caller-provided-id")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("X-Request-Id"); got != "caller-provided-id" {
		t.Errorf("应透传调用方的 request id，实为 %q", got)
	}

	var env testEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.RequestID != "caller-provided-id" {
		t.Errorf("响应体 request_id = %q", env.RequestID)
	}
}

func TestCapabilities(t *testing.T) {
	ts := newTestEnv(t)
	env := doRequest(t, ts, http.MethodGet, "/api/v1/meta/capabilities", "", http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var caps []schema.Capability
	if err := json.Unmarshal(env.Data, &caps); err != nil {
		t.Fatal(err)
	}
	if len(caps) != 3 || caps[0].Method != schema.MethodLiuYao {
		t.Errorf("能力清单不符: %+v", caps)
	}
}

func TestMetaCalendar(t *testing.T) {
	ts := newTestEnv(t)
	env := doRequest(t, ts, http.MethodGet, "/api/v1/meta/calendar?at=2026-11-25T14:30:00%2B08:00&lon=113.26", "", http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var cal struct {
		DayGZ   string `json:"day_gz"`
		MonthGZ string `json:"month_gz"`
	}
	if err := json.Unmarshal(env.Data, &cal); err != nil {
		t.Fatal(err)
	}
	if cal.DayGZ != "癸卯" || cal.MonthGZ != "己亥" {
		t.Errorf("日月柱 = %s / %s，期望 癸卯 / 己亥", cal.DayGZ, cal.MonthGZ)
	}
}

func TestCastManual(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","question":"这次考试能否通过","at":"2026-11-25T14:30:00+08:00","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}

	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	if view.ChartID == "" || len(view.ChartID) != 36 {
		t.Errorf("chart_id 应为 UUID，实为 %q", view.ChartID)
	}
	if view.Question != "这次考试能否通过" {
		t.Errorf("question = %q", view.Question)
	}
	if view.Entropy != nil {
		t.Error("手动录入不应产生摇卦记录")
	}
	payload, ok := view.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload 类型 = %T", view.Payload)
	}
	if payload["gua_name"] != "乾为天" {
		t.Errorf("卦名 = %v", payload["gua_name"])
	}
	useShen, ok := payload["use_shen"].(map[string]any)
	if !ok {
		t.Fatalf("payload 缺少 use_shen：%T", payload["use_shen"])
	}
	if useShen["topic"] != "exam" || useShen["rule_id"] != "YS-03" {
		t.Errorf("考试用神规则 = %v / %v", useShen["topic"], useShen["rule_id"])
	}
	var useShenFact bool
	for _, f := range view.Facts {
		if f.Label == "用神" {
			useShenFact = true
		}
	}
	if !useShenFact {
		t.Error("core facts 缺少用神事实")
	}
	for _, f := range view.Facts {
		if f.Weight != schema.WeightCore {
			t.Errorf("排盘响应默认只回 core 事实，出现了 %s", f.Weight)
		}
	}
}

func TestFollowupRouteRequiresAI(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","question":"同盘追问接口","at":"2026-11-25T14:30:00+08:00","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	followup := doRequest(t, ts, http.MethodPost, "/api/v1/charts/"+view.ChartID+"/messages", `{"message":"应期为什么这么看？"}`, http.StatusServiceUnavailable)
	if followup.Code != 30003 {
		t.Errorf("无模型追问 code = %d，期望 30003", followup.Code)
	}
	if followup.Message == "" {
		t.Error("无模型追问应给出可展示的错误消息")
	}

	// 服务端内存里没有这张盘，且没有本地快照时，应明确 404，而不是空流。
	missing := doRequest(t, ts, http.MethodPost, "/api/v1/charts/not-a-chart/messages", `{"message":"还在吗？"}`, http.StatusNotFound)
	if missing.Code != 10301 {
		t.Errorf("缺盘追问 code = %d，期望 10301", missing.Code)
	}
}

func TestCastSurnameTopic(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","question":"测一下我的姓氏，提示在生肖中","at":"2026-11-26T21:55:00+08:00","lines":["少阳","少阴","少阳","少阴","少阴","少阴"]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	payload := view.Payload.(map[string]any)
	useShen, ok := payload["use_shen"].(map[string]any)
	if !ok {
		t.Fatalf("payload 缺少 use_shen：%T", payload["use_shen"])
	}
	if useShen["topic"] != "surname" || useShen["rule_id"] != "YS-10" {
		t.Errorf("测姓规则 = %v / %v", useShen["topic"], useShen["rule_id"])
	}
	if useShen["primary"].([]any)[0] != "世爻" {
		t.Errorf("自测姓氏应以世爻为主：%v", useShen["primary"])
	}
	var surnameFact string
	for _, f := range view.Facts {
		if f.Label == "姓氏候选" {
			surnameFact = f.Value
		}
	}
	if !strings.Contains(surnameFact, "牛") || strings.Contains(surnameFact, "姓兔") {
		t.Errorf("姓氏候选应只落真实姓氏牛，实际为 %q", surnameFact)
	}
}

func TestCastCoinServerToss(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"coin","question":"求财","at":"2026-11-25T14:30:00+08:00"}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}

	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	if view.Entropy == nil {
		t.Fatal("服务端摇卦应返回 entropy 记录")
	}
	if !ts.src.Verify(*view.Entropy) {
		t.Fatal("entropy_proof 复核失败")
	}
	if view.Entropy.BacksString() == "" {
		t.Fatal("背数串为空")
	}
}

func TestCastCoinClientCoins(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"coin","coins":[[1,1,1],[1,1,0],[1,0,0],[0,0,0],[1,1,1],[1,0,0]]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	if view.Entropy != nil {
		t.Error("客户端自带铜钱时不应产生服务端摇卦记录")
	}
	payload := view.Payload.(map[string]any)
	if payload["changed_gua"] != "火山旅" {
		t.Errorf("变卦 = %v", payload["changed_gua"])
	}
}

func TestCastErrors(t *testing.T) {
	ts := newTestEnv(t)

	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/ziwei/cast", `{"mode":"manual"}`, http.StatusBadRequest)
	if env.Code != 10001 {
		t.Errorf("缺少出生信息错误码 = %d", env.Code)
	}

	env = doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", `{"mode":"time"}`, http.StatusBadRequest)
	if env.Code != 10001 {
		t.Errorf("非法 mode 错误码 = %d", env.Code)
	}

	env = doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", `{"mode":"manual","lines":["少阳"]}`, http.StatusBadRequest)
	if env.Code != 10001 {
		t.Errorf("爻数不足错误码 = %d", env.Code)
	}

	env = doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", `not json`, http.StatusBadRequest)
	if env.Code != 10001 {
		t.Errorf("非法 JSON 错误码 = %d", env.Code)
	}
}

func TestCastDaLiuRen(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"question":"这次考试能否通过？","at":"2026-11-25T14:32:00+08:00","tz":"Asia/Shanghai","use_true_solar_time":false}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/daliuren/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	if view.Method != schema.MethodDaLiuRen {
		t.Fatalf("method = %q", view.Method)
	}
	payload := view.Payload.(map[string]any)
	tianPan, _ := payload["tian_pan"].([]any)
	if len(tianPan) != 12 {
		t.Fatalf("tian_pan = %d", len(tianPan))
	}
	sanChuan, _ := payload["san_chuan"].([]any)
	if len(sanChuan) != 3 {
		t.Fatalf("san_chuan = %d", len(sanChuan))
	}
	if !hasFactLabel(view.Facts, "课体") {
		t.Fatal("缺少课体事实")
	}
}

func TestCastZiWei(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"question":"事业方向","gender":"male","calendar_type":"solar","birth_date":"1990-05-15","birth_time":"10:30","tz":"Asia/Shanghai","use_true_solar_time":false}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/ziwei/cast", body, http.StatusOK)
	if env.Code != 0 {
		t.Fatalf("code = %d, message = %s", env.Code, env.Message)
	}
	var view chartView
	if err := json.Unmarshal(env.Data, &view); err != nil {
		t.Fatal(err)
	}
	if view.Method != schema.MethodZiWei {
		t.Fatalf("method = %q", view.Method)
	}
	payload := view.Payload.(map[string]any)
	palaces, _ := payload["palaces"].([]any)
	if len(palaces) != 12 {
		t.Fatalf("palaces = %d", len(palaces))
	}
	if !hasFactLabel(view.Facts, "命宫") || !hasFactLabel(view.Facts, "五行局") {
		t.Fatal("紫微核心事实缺失")
	}
}

func hasFactLabel(facts []schema.Fact, label string) bool {
	for _, f := range facts {
		if f.Label == label {
			return true
		}
	}
	return false
}

func TestGetChart(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	var cast chartView
	if err := json.Unmarshal(env.Data, &cast); err != nil {
		t.Fatal(err)
	}

	env = doRequest(t, ts, http.MethodGet, "/api/v1/charts/"+cast.ChartID, "", http.StatusOK)
	var got chartView
	if err := json.Unmarshal(env.Data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Facts) <= len(cast.Facts) {
		t.Fatalf("详情应返回完整 facts：cast=%d get=%d", len(cast.Facts), len(got.Facts))
	}

	env = doRequest(t, ts, http.MethodGet, "/api/v1/charts/not-exist", "", http.StatusNotFound)
	if env.Code != 10301 {
		t.Errorf("不存在盘面错误码 = %d", env.Code)
	}
}

func TestListCharts(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"manual","lines":["少阳","少阳","少阳","少阳","少阳","少阳"]}`
	for i := 0; i < 2; i++ {
		doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusOK)
	}

	env := doRequest(t, ts, http.MethodGet, "/api/v1/charts?limit=1", "", http.StatusOK)
	var list listChartsResponse
	if err := json.Unmarshal(env.Data, &list); err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 || len(list.Items) != 1 || list.NextCursor != "1" {
		t.Fatalf("列表结果不符: %+v", list)
	}
}

func TestCastBacksOutOfRange(t *testing.T) {
	ts := newTestEnv(t)
	body := `{"mode":"coin","backs":[4,0,0,0,0,0]}`
	env := doRequest(t, ts, http.MethodPost, "/api/v1/charts/liuyao/cast", body, http.StatusBadRequest)
	if env.Code != 10001 {
		t.Errorf("背数越界错误码 = %d", env.Code)
	}
}
