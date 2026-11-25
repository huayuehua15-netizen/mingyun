// Package ai 负责把已排好的盘面 facts 交给大模型，转成可读的解读。
//
// 这一层不排盘、不改盘、不补规则。它只做三件事：读取模型配置、拼装
// 提示词、消费 OpenAI 兼容的流式响应。没有配置模型时，由 offline.go
// 给出一份只列事实的离线摘要，保证站点在零配置下也能用。
package ai

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout     = 90 * time.Second
	defaultMaxTokens   = 1200
	maxAllowedTokens   = 8192
	defaultTemperature = 0.5
)

// Config 模型连接配置，全部来自环境变量。
type Config struct {
	BaseURL     string
	APIKey      string
	Model       string
	Timeout     time.Duration
	MaxTokens   int
	Temperature float64
}

// LoadConfig 从环境变量读取配置，缺省值面向本地小模型。
func LoadConfig() Config {
	cfg := Config{
		BaseURL:     strings.TrimSpace(os.Getenv("MINGYUN_AI_BASE_URL")),
		APIKey:      strings.TrimSpace(os.Getenv("MINGYUN_AI_API_KEY")),
		Model:       strings.TrimSpace(os.Getenv("MINGYUN_AI_MODEL")),
		Timeout:     defaultTimeout,
		MaxTokens:   defaultMaxTokens,
		Temperature: defaultTemperature,
	}

	if raw := strings.TrimSpace(os.Getenv("MINGYUN_AI_TIMEOUT")); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			cfg.Timeout = d
		}
	}
	if raw := strings.TrimSpace(os.Getenv("MINGYUN_AI_MAX_TOKENS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			cfg.MaxTokens = n
		}
	}
	if raw := strings.TrimSpace(os.Getenv("MINGYUN_AI_TEMPERATURE")); raw != "" {
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			cfg.Temperature = f
		}
	}
	cfg.MaxTokens = clampInt(cfg.MaxTokens, 1, maxAllowedTokens)
	cfg.Temperature = clampFloat(cfg.Temperature, 0, 2)
	return cfg
}

// Enabled 只有 base_url 与 model 同时存在才启用在线解读。
func (c Config) Enabled() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.Model) != ""
}

// Endpoint 返回 OpenAI 兼容的 chat completions 地址。
func (c Config) Endpoint() string {
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

// Provider 给前端一个可读的来源标识，不暴露密钥。
func (c Config) Provider() string {
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" {
		return "openai-compatible"
	}
	return u.Host
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
