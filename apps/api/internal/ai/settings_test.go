package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProviderPresetsHaveStableIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range ProviderPresets() {
		if p.ID == "" || seen[p.ID] {
			t.Fatalf("预置 ID 为空或重复: %+v", p)
		}
		seen[p.ID] = true
		if p.Custom {
			continue
		}
		if p.Name == "" || p.BaseURL == "" {
			t.Fatalf("预置缺少名称或地址: %+v", p)
		}
		if InferProviderID(p.BaseURL) != p.ID {
			t.Fatalf("地址无法反推预置 ID: %s -> %s", p.BaseURL, InferProviderID(p.BaseURL))
		}
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{" https://api.example.com/v1/ ", "https://api.example.com/v1"},
		{"https://api.example.com/v1/chat/completions/", "https://api.example.com/v1"},
		{"", ""},
	}
	for _, c := range cases {
		if got := NormalizeBaseURL(c.raw); got != c.want {
			t.Fatalf("NormalizeBaseURL(%q)=%q, want %q", c.raw, got, c.want)
		}
	}
}

func TestNewSettingsStoreUsesEnvFallback(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	cfg := Config{
		BaseURL:     "https://api.deepseek.com/v1/",
		APIKey:      "sk-12345678",
		Model:       "deepseek-chat",
		Timeout:     45 * time.Second,
		MaxTokens:   1800,
		Temperature: 0.3,
	}
	store, err := NewSettingsStore(filepath.Join(t.TempDir(), "ai.json"), cfg)
	if err != nil {
		t.Fatal(err)
	}
	view := store.View()
	if !view.Enabled || view.Source != "env" {
		t.Fatalf("view = %+v", view)
	}
	if view.ProviderID != "deepseek" || view.ProviderName != "深度求索 DeepSeek" {
		t.Fatalf("provider = %+v", view)
	}
	if !view.HasKey || view.KeyHint != "••••5678" {
		t.Fatalf("key hint = %q", view.KeyHint)
	}
	if len(view.Presets) < 8 {
		t.Fatalf("presets 太少: %d", len(view.Presets))
	}
	if got := store.Config(); got.APIKey != "sk-12345678" || got.Model != "deepseek-chat" {
		t.Fatalf("config = %+v", got)
	}
}

func TestSettingsStoreSaveReloadAndClear(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	path := filepath.Join(t.TempDir(), "nested", "ai-settings.json")
	store, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = store.Update(Settings{
		ProviderID:  "deepseek",
		BaseURL:     "https://api.deepseek.com/v1/",
		APIKey:      "sk-abcdefgh",
		Model:       "deepseek-chat",
		Temperature: 0.35,
		MaxTokens:   2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("配置文件未落盘: %v", err)
	}

	reloaded, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	view := reloaded.View()
	if !view.Enabled || view.Source != "file" || view.Model != "deepseek-chat" {
		t.Fatalf("reload view = %+v", view)
	}
	if !view.HasKey || view.KeyHint != "••••efgh" {
		t.Fatalf("reload key = %+v", view)
	}
	if got := reloaded.Config().APIKey; got != "sk-abcdefgh" {
		t.Fatalf("reload key = %q", got)
	}

	if err := reloaded.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Clear 后文件仍存在: %v", err)
	}
	after := reloaded.View()
	if after.Enabled || after.Source != "none" {
		t.Fatalf("Clear 后 view = %+v", after)
	}
}

func TestSettingsStoreValidatesAddress(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	store, err := NewSettingsStore("", Config{})
	if err != nil {
		t.Fatal(err)
	}
	tests := []Settings{
		{BaseURL: "https://api.example.com/v1", APIKey: "k"},
		{BaseURL: "ftp://api.example.com/v1", Model: "m", APIKey: "k"},
		{BaseURL: "https://api.example.com/v1", Model: "m", APIKey: "k", Temperature: 3},
	}
	for _, in := range tests {
		if err := store.Update(in); err != nil {
			continue
		}
		t.Fatalf("应拒绝非法配置: %+v", in)
	}
}

func TestIsLocalBaseURL(t *testing.T) {
	for _, base := range []string{"http://localhost:11434/v1", "http://127.0.0.1:8000/v1", "http://host.docker.internal:11434/v1"} {
		if !IsLocalBaseURL(base) {
			t.Errorf("%s 应识别为本机端点", base)
		}
	}
	if IsLocalBaseURL("https://api.deepseek.com/v1") {
		t.Fatal("公网地址不应识别为本机")
	}
}
