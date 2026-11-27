package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/ai"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
)

type aiSettingsData struct {
	Enabled      bool                `json:"enabled"`
	ProviderID   string              `json:"provider_id"`
	ProviderName string              `json:"provider_name"`
	Model        string              `json:"model"`
	BaseURL      string              `json:"base_url"`
	HasKey       bool                `json:"has_key"`
	KeyHint      string              `json:"key_hint"`
	Source       string              `json:"source"`
	Writable     bool                `json:"writable"`
	Presets      []ai.ProviderPreset `json:"presets"`
}

type aiTestData struct {
	OK        bool   `json:"ok"`
	LatencyMS int64  `json:"latency_ms"`
	Message   string `json:"message"`
	Preview   string `json:"preview"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
}

func newSettingsTestEnv(t *testing.T, writable string) (*testEnv, *ai.SettingsStore, string) {
	t.Helper()
	t.Setenv("MINGYUN_MODEL_SETTINGS", writable)
	t.Setenv("MINGYUN_AI_BASE_URL", "")
	t.Setenv("MINGYUN_AI_API_KEY", "")
	t.Setenv("MINGYUN_AI_MODEL", "")
	path := filepath.Join(t.TempDir(), "ai-settings.json")
	settings, err := ai.NewSettingsStore(path, ai.Config{})
	if err != nil {
		t.Fatal(err)
	}
	seq := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	src := entropy.NewSourceWithReader(
		bytes.Repeat([]byte{0x2a}, 32),
		bytes.NewReader(bytes.Repeat(seq, 64)),
		func() time.Time { return testNow },
	)
	srv := httptest.NewServer(NewWithSettings("0.1.0-test", store.NewMemory(), src, settings))
	t.Cleanup(srv.Close)
	return &testEnv{srv: srv, src: src}, settings, path
}

func decodeSettingsData(t *testing.T, raw json.RawMessage) aiSettingsData {
	t.Helper()
	var data aiSettingsData
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAISettingsGetAndPut(t *testing.T) {
	ts, settings, path := newSettingsTestEnv(t, "on")

	env := doRequest(t, ts, http.MethodGet, "/api/v1/settings/ai", "", http.StatusOK)
	data := decodeSettingsData(t, env.Data)
	if data.Enabled || data.Source != "none" || !data.Writable {
		t.Fatalf("初始设置 = %+v", data)
	}
	if len(data.Presets) < 8 {
		t.Fatalf("预置数量 = %d", len(data.Presets))
	}

	body := `{"provider_id":"deepseek","model":"deepseek-chat","api_key":"sk-verysecret-1234","temperature":0.3,"max_tokens":1800}`
	env = doRequest(t, ts, http.MethodPut, "/api/v1/settings/ai", body, http.StatusOK)
	data = decodeSettingsData(t, env.Data)
	if !data.Enabled || data.ProviderName != "深度求索 DeepSeek" || data.Model != "deepseek-chat" {
		t.Fatalf("保存后设置 = %+v", data)
	}
	if data.BaseURL != "https://api.deepseek.com/v1" {
		t.Fatalf("应自动填入预置地址，实际 = %q", data.BaseURL)
	}
	if !data.HasKey || data.KeyHint != "••••1234" {
		t.Fatalf("Key 掩码 = %+v", data)
	}
	if strings.Contains(string(env.Data), "sk-verysecret-1234") {
		t.Fatal("响应中不应出现明文 Key")
	}
	if settings.Settings().APIKey != "sk-verysecret-1234" {
		t.Fatal("服务端内存配置未保留 Key")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("配置文件未生成: %v", err)
	}

	reloaded, err := ai.NewSettingsStore(path, ai.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.View().Model != "deepseek-chat" || reloaded.Config().APIKey != "sk-verysecret-1234" {
		t.Fatalf("重载配置 = %+v", reloaded.View())
	}
}

func TestAISettingsClearRestoresOffline(t *testing.T) {
	ts, _, path := newSettingsTestEnv(t, "on")
	body := `{"provider_id":"deepseek","model":"deepseek-chat","api_key":"sk-verysecret-1234"}`
	doRequest(t, ts, http.MethodPut, "/api/v1/settings/ai", body, http.StatusOK)
	env := doRequest(t, ts, http.MethodDelete, "/api/v1/settings/ai", "", http.StatusOK)
	data := decodeSettingsData(t, env.Data)
	if data.Enabled || data.HasKey || data.Source != "none" {
		t.Fatalf("清除后 = %+v", data)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("配置文件未删除: %v", err)
	}
}

func TestAISettingsLocked(t *testing.T) {
	ts, _, _ := newSettingsTestEnv(t, "off")
	env := doRequest(t, ts, http.MethodGet, "/api/v1/settings/ai", "", http.StatusOK)
	if decodeSettingsData(t, env.Data).Writable {
		t.Fatal("只读模式应返回 writable=false")
	}
	body := `{"provider_id":"deepseek","model":"deepseek-chat","api_key":"sk-test"}`
	doRequest(t, ts, http.MethodPut, "/api/v1/settings/ai", body, http.StatusForbidden)
}

func TestAISettingsTestConnection(t *testing.T) {
	ts, _, _ := newSettingsTestEnv(t, "on")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"正常\"},\"finish_reason\":\"stop\"}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()

	body := fmt.Sprintf(`{"provider_id":"custom","base_url":%q,"model":"fake-model","api_key":"test-key"}`, upstream.URL+"/v1")
	env := doRequest(t, ts, http.MethodPost, "/api/v1/settings/ai/test", body, http.StatusOK)
	var data aiTestData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.OK || data.Model != "fake-model" || !strings.Contains(data.Preview, "正常") {
		t.Fatalf("测试结果 = %+v", data)
	}
	if data.LatencyMS < 0 {
		t.Fatalf("延迟不应为负: %d", data.LatencyMS)
	}
}

func TestAISettingsRejectsMissingKey(t *testing.T) {
	ts, _, _ := newSettingsTestEnv(t, "on")
	body := `{"provider_id":"deepseek","model":"deepseek-chat"}`
	doRequest(t, ts, http.MethodPut, "/api/v1/settings/ai", body, http.StatusBadRequest)
}
