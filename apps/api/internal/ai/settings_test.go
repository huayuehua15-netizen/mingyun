package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestSettingsStoreEncryptsAPIKeyAtRest(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	path := filepath.Join(t.TempDir(), "ai-settings.json")
	apiKey := "sk-very-secret-123456"
	store, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(filepath.Dir(path), settingsKeyFileName)
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(keyData) != settingsKeySize {
		t.Fatalf("本地密钥长度 = %d, want %d", len(keyData), settingsKeySize)
	}
	if err := store.Update(Settings{
		ProviderID:  "deepseek",
		BaseURL:     "https://api.deepseek.com/v1",
		APIKey:      apiKey,
		Model:       "deepseek-chat",
		Temperature: 0.3,
		MaxTokens:   1200,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), apiKey) {
		t.Fatal("ai-settings.json 不应包含明文 API Key")
	}
	var file settingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	if file.Version != settingsFileVersion {
		t.Fatalf("配置文件版本 = %d, want %d", file.Version, settingsFileVersion)
	}
	if file.APIKeyCipher == "" || file.Settings.APIKey != "" {
		t.Fatalf("加密字段不符合预期: %+v", file)
	}

	reloaded, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Config().APIKey; got != apiKey {
		t.Fatalf("重载 Key = %q", got)
	}
	if !reloaded.View().KeyEncrypted {
		t.Fatal("重载后应标记 Key 已加密")
	}
}

func TestSettingsStoreAppSecretEncryption(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	path := filepath.Join(t.TempDir(), "ai-settings.json")
	secret := []byte("0123456789abcdef")
	apiKey := "sk-app-secret-abcdef"
	store, err := NewSettingsStoreWithSecret(path, Config{}, secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), settingsKeyFileName)); !os.IsNotExist(err) {
		t.Fatal("APP_SECRET 模式不应生成本地密钥文件")
	}
	if err := store.Update(Settings{
		ProviderID: "deepseek",
		BaseURL:    "https://api.deepseek.com/v1",
		APIKey:     apiKey,
		Model:      "deepseek-chat",
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), apiKey) {
		t.Fatal("APP_SECRET 模式下配置文件不应包含明文 Key")
	}
	reloaded, err := NewSettingsStoreWithSecret(path, Config{}, secret)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Config().APIKey; got != apiKey {
		t.Fatalf("同 APP_SECRET 重载 Key = %q", got)
	}
	if _, err := NewSettingsStoreWithSecret(path, Config{}, []byte("fedcba9876543210")); err == nil {
		t.Fatal("不同 APP_SECRET 不应解密成功")
	}
}

func TestSettingsStoreMigratesLegacyPlaintext(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	path := filepath.Join(t.TempDir(), "ai-settings.json")
	legacyKey := "sk-legacy-plaintext"
	legacy := settingsFile{
		Version: 1,
		Settings: Settings{
			ProviderID: "deepseek",
			BaseURL:    "https://api.deepseek.com/v1",
			APIKey:     legacyKey,
			Model:      "deepseek-chat",
			MaxTokens:  1200,
		},
	}
	raw, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Config().APIKey; got != legacyKey {
		t.Fatalf("迁移后内存 Key = %q", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), legacyKey) {
		t.Fatal("迁移后文件不应保留明文 Key")
	}
	var migrated settingsFile
	if err := json.Unmarshal(data, &migrated); err != nil {
		t.Fatal(err)
	}
	if migrated.Version != settingsFileVersion || migrated.APIKeyCipher == "" {
		t.Fatalf("迁移格式 = %+v", migrated)
	}
	reloaded, err := NewSettingsStore(path, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Config().APIKey; got != legacyKey {
		t.Fatalf("迁移后重载 Key = %q", got)
	}
}

func TestSettingsStoreRejectsPlaintextV2(t *testing.T) {
	t.Setenv("MINGYUN_MODEL_SETTINGS", "on")
	path := filepath.Join(t.TempDir(), "ai-settings.json")
	raw := []byte(`{
  "version": 2,
  "settings": {
    "provider_id": "deepseek",
    "base_url": "https://api.deepseek.com/v1",
    "api_key": "sk-plain-v2-must-not-load",
    "model": "deepseek-chat"
  }
}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSettingsStore(path, Config{}); err == nil || !strings.Contains(err.Error(), "仍含明文") {
		t.Fatalf("version 2 明文配置应拒绝载入，err=%v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "sk-plain-v2-must-not-load") {
		t.Fatal("拒绝载入时不应改写原文件")
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
