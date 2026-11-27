package ai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Settings 一份可持久化的模型接入配置。
//
// APIKey 只保存在服务端配置文件里，SettingsView 只返回掩码，不把 Key 回传浏览器。
type Settings struct {
	ProviderID  string    `json:"provider_id"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type settingsFile struct {
	Version  int      `json:"version"`
	Settings Settings `json:"settings"`
}

// AISettingsView 是设置接口与前端共用的只读快照，不含明文 Key。
type AISettingsView struct {
	Enabled        bool             `json:"enabled"`
	ProviderID     string           `json:"provider_id"`
	ProviderName   string           `json:"provider_name"`
	Provider       string           `json:"provider"`
	Model          string           `json:"model"`
	BaseURL        string           `json:"base_url"`
	HasKey         bool             `json:"has_key"`
	KeyHint        string           `json:"key_hint,omitempty"`
	Temperature    float64          `json:"temperature"`
	MaxTokens      int              `json:"max_tokens"`
	TimeoutSeconds int              `json:"timeout_seconds"`
	Source         string           `json:"source"`
	Writable       bool             `json:"writable"`
	UpdatedAt      string           `json:"updated_at,omitempty"`
	Presets        []ProviderPreset `json:"presets"`
}

// SettingsStore 管理运行中的模型配置，并可选地落盘。
//
// 优先级：配置文件 > 环境变量。配置文件不存在时，使用 MINGYUN_AI_* 环境变量；
// 保存后立即热切换运行时客户端，不需要重启 API。
type SettingsStore struct {
	mu       sync.RWMutex
	path     string
	fallback Settings
	timeout  time.Duration
	writable bool
	current  Settings
	source   string
}

// NewSettingsStore 读取已有配置文件；文件不存在时使用 fallback。
func NewSettingsStore(path string, fallback Config) (*SettingsStore, error) {
	timeout := fallback.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	st := &SettingsStore{
		path:     strings.TrimSpace(path),
		fallback: settingsFromConfig(fallback),
		timeout:  timeout,
		writable: modelSettingsWritable(),
		current:  settingsFromConfig(fallback),
		source:   sourceForFallback(fallback),
	}
	if st.path == "" {
		return st, nil
	}
	if err := st.load(); err != nil {
		return nil, err
	}
	return st, nil
}

// NewMemorySettings 构造不落盘的设置存储，适合测试或只读降级。
func NewMemorySettings(fallback Config) *SettingsStore {
	st, _ := NewSettingsStore("", fallback)
	return st
}

func settingsFromConfig(cfg Config) Settings {
	base := NormalizeBaseURL(cfg.BaseURL)
	temperature := cfg.Temperature
	if temperature < 0 || temperature > 2 {
		temperature = defaultTemperature
	}
	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	maxTokens = clampInt(maxTokens, 1, maxAllowedTokens)
	return Settings{
		ProviderID:  InferProviderID(base),
		BaseURL:     base,
		APIKey:      strings.TrimSpace(cfg.APIKey),
		Model:       strings.TrimSpace(cfg.Model),
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
}

func sourceForFallback(cfg Config) string {
	if cfg.Enabled() {
		return "env"
	}
	return "none"
}

func modelSettingsWritable() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MINGYUN_MODEL_SETTINGS"))) {
	case "off", "false", "0", "readonly", "read-only":
		return false
	default:
		return true
	}
}

func (s *SettingsStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取模型配置 %s: %w", s.path, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("模型配置 %s 为空", s.path)
	}
	var file settingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("解析模型配置 %s: %w", s.path, err)
	}
	sett := file.Settings
	sett.BaseURL = NormalizeBaseURL(sett.BaseURL)
	sett.Model = strings.TrimSpace(sett.Model)
	sett.APIKey = strings.TrimSpace(sett.APIKey)
	sett.ProviderID = strings.TrimSpace(sett.ProviderID)
	if sett.ProviderID == "" {
		sett.ProviderID = InferProviderID(sett.BaseURL)
	}
	if sett.Temperature < 0 || sett.Temperature > 2 {
		sett.Temperature = defaultTemperature
	}
	if sett.MaxTokens <= 0 {
		sett.MaxTokens = defaultMaxTokens
	}
	sett.MaxTokens = clampInt(sett.MaxTokens, 1, maxAllowedTokens)
	if sett.UpdatedAt.IsZero() {
		if info, statErr := os.Stat(s.path); statErr == nil {
			sett.UpdatedAt = info.ModTime()
		}
	}
	s.current = sett
	s.source = "file"
	return nil
}

// Config 返回当前生效的模型连接配置。
func (s *SettingsStore) Config() Config {
	if s == nil {
		return Config{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := Config{
		BaseURL:     s.current.BaseURL,
		APIKey:      s.current.APIKey,
		Model:       s.current.Model,
		Timeout:     s.timeout,
		MaxTokens:   s.current.MaxTokens,
		Temperature: s.current.Temperature,
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	cfg.MaxTokens = clampInt(cfg.MaxTokens, 1, maxAllowedTokens)
	cfg.Temperature = clampFloat(cfg.Temperature, 0, 2)
	return cfg
}

// Settings 返回当前原始配置（含 Key），只供服务端内部使用。
func (s *SettingsStore) Settings() Settings {
	if s == nil {
		return Settings{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// View 返回脱敏后的设置快照。
func (s *SettingsStore) View() AISettingsView {
	if s == nil {
		return AISettingsView{Writable: false, Presets: ProviderPresets(), Source: "none"}
	}
	s.mu.RLock()
	sett := s.current
	source := s.source
	writable := s.writable
	updatedAt := sett.UpdatedAt
	s.mu.RUnlock()

	cfg := Config{
		BaseURL:     sett.BaseURL,
		APIKey:      sett.APIKey,
		Model:       sett.Model,
		MaxTokens:   sett.MaxTokens,
		Temperature: sett.Temperature,
	}
	preset, ok := FindProviderPreset(sett.ProviderID)
	providerName := "自定义接入"
	switch {
	case sett.BaseURL == "":
		providerName = "未接入"
	case ok:
		providerName = preset.Name
	}
	updated := ""
	if !updatedAt.IsZero() {
		updated = updatedAt.Format(time.RFC3339)
	}
	return AISettingsView{
		Enabled:        cfg.Enabled(),
		ProviderID:     sett.ProviderID,
		ProviderName:   providerName,
		Provider:       cfg.Provider(),
		Model:          sett.Model,
		BaseURL:        sett.BaseURL,
		HasKey:         sett.APIKey != "",
		KeyHint:        maskKey(sett.APIKey),
		Temperature:    sett.Temperature,
		MaxTokens:      sett.MaxTokens,
		TimeoutSeconds: int(s.timeout / time.Second),
		Source:         source,
		Writable:       writable,
		UpdatedAt:      updated,
		Presets:        ProviderPresets(),
	}
}

// Writable 返回服务端是否允许网页修改模型配置。
func (s *SettingsStore) Writable() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writable
}

// Update 校验并保存配置，同时更新内存中的运行时配置。
func (s *SettingsStore) Update(in Settings) error {
	if s == nil {
		return fmt.Errorf("模型设置存储未初始化")
	}
	in.BaseURL = NormalizeBaseURL(in.BaseURL)
	in.Model = strings.TrimSpace(in.Model)
	in.APIKey = strings.TrimSpace(in.APIKey)
	in.ProviderID = strings.TrimSpace(in.ProviderID)
	if in.BaseURL == "" && in.Model != "" {
		return fmt.Errorf("填写模型名时必须填写接口地址")
	}
	if in.BaseURL == "" && in.APIKey != "" {
		return fmt.Errorf("清空接口地址时不能保留 API Key")
	}
	if in.BaseURL != "" && in.Model == "" {
		return fmt.Errorf("填写接口地址时必须填写模型名")
	}
	if in.BaseURL != "" {
		u, err := url.Parse(in.BaseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("接口地址不是合法 URL")
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("接口地址只支持 http / https")
		}
	}
	if in.ProviderID == "" {
		in.ProviderID = InferProviderID(in.BaseURL)
	}
	if in.Temperature < 0 || in.Temperature > 2 {
		return fmt.Errorf("temperature 应在 0 到 2 之间")
	}
	if in.MaxTokens <= 0 {
		in.MaxTokens = defaultMaxTokens
	}
	if in.MaxTokens > maxAllowedTokens {
		in.MaxTokens = maxAllowedTokens
	}
	in.UpdatedAt = time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	previous, previousSource := s.current, s.source
	s.current = in
	s.source = "file"
	if s.path == "" {
		return nil
	}
	if err := s.persistLocked(); err != nil {
		s.current, s.source = previous, previousSource
		return err
	}
	return nil
}

// Clear 删除网页保存的配置，恢复环境变量作为兜底。
func (s *SettingsStore) Clear() error {
	if s == nil {
		return fmt.Errorf("模型设置存储未初始化")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path != "" {
		if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除模型配置: %w", err)
		}
	}
	s.current = s.fallback
	s.source = sourceForFallback(Config{
		BaseURL: s.fallback.BaseURL,
		Model:   s.fallback.Model,
	})
	return nil
}

func (s *SettingsStore) persistLocked() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("创建模型配置目录: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("收紧模型配置目录: %w", err)
	}
	data, err := json.MarshalIndent(settingsFile{Version: 1, Settings: s.current}, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".ai-settings-*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时配置: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("写入临时配置: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("设置配置权限: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时配置: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("替换模型配置: %w", err)
	}
	return nil
}

func maskKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	runes := []rune(key)
	if len(runes) <= 4 {
		return "已保存"
	}
	return "••••" + string(runes[len(runes)-4:])
}
