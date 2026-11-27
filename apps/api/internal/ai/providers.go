package ai

import (
	"net/url"
	"strings"
)

// ProviderPreset 一组可以直接选择的服务商接入参数。
//
// BaseURL 与 Models 只是默认值，最终以用户在设置页填写的内容为准。
// 模型迭代很快，模型名允许自由输入，不把预设当成白名单。
type ProviderPreset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	BaseURL      string   `json:"base_url"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models"`
	KeyHint      string   `json:"key_hint,omitempty"`
	Note         string   `json:"note,omitempty"`
	DocsURL      string   `json:"docs_url,omitempty"`
	Custom       bool     `json:"custom,omitempty"`
}

var providerPresets = []ProviderPreset{
	{
		ID:           "deepseek",
		Name:         "深度求索 DeepSeek",
		Region:       "cn",
		BaseURL:      "https://api.deepseek.com/v1",
		DefaultModel: "deepseek-chat",
		Models:       []string{"deepseek-chat", "deepseek-reasoner"},
		KeyHint:      "sk-",
		Note:         "官方 OpenAI 兼容接口",
		DocsURL:      "https://api-docs.deepseek.com/",
	},
	{
		ID:           "dashscope",
		Name:         "通义千问",
		Region:       "cn",
		BaseURL:      "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultModel: "qwen-plus",
		Models:       []string{"qwen-plus", "qwen-max", "qwen-turbo", "qwen-long"},
		KeyHint:      "sk-",
		Note:         "阿里云百炼兼容模式；模型名按控制台为准",
		DocsURL:      "https://help.aliyun.com/zh/model-studio/",
	},
	{
		ID:           "moonshot",
		Name:         "Kimi",
		Region:       "cn",
		BaseURL:      "https://api.moonshot.cn/v1",
		DefaultModel: "kimi-k2.7",
		Models:       []string{"kimi-k2.7", "kimi-k3", "kimi-k2.5"},
		KeyHint:      "sk-",
		Note:         "月之暗面 OpenAI 兼容接口",
		DocsURL:      "https://platform.moonshot.cn/docs/",
	},
	{
		ID:           "zhipu",
		Name:         "智谱 GLM",
		Region:       "cn",
		BaseURL:      "https://open.bigmodel.cn/api/paas/v4",
		DefaultModel: "glm-5.3",
		Models:       []string{"glm-5.3", "glm-5.2", "glm-4.6"},
		KeyHint:      "",
		Note:         "智谱开放平台 OpenAI 兼容接口",
		DocsURL:      "https://open.bigmodel.cn/dev/api",
	},
	{
		ID:           "sensenova",
		Name:         "商汤日日新",
		Region:       "cn",
		BaseURL:      "https://token.sensenova.cn/v1",
		DefaultModel: "sensenova-6.8-flash",
		Models:       []string{"sensenova-6.8-flash", "sensenova-6.5-flash"},
		KeyHint:      "sk-",
		Note:         "商汤 Token Plan，当前项目默认实测模型",
		DocsURL:      "https://platform.sensenova.cn/",
	},
	{
		ID:           "ark",
		Name:         "火山方舟",
		Region:       "cn",
		BaseURL:      "https://ark.cn-beijing.volces.com/api/v3",
		DefaultModel: "",
		KeyHint:      "sk-",
		Note:         "模型名填方舟控制台里的模型 ID 或推理接入点 ID",
		DocsURL:      "https://www.volcengine.com/docs/82379",
	},
	{
		ID:           "openai",
		Name:         "OpenAI",
		Region:       "global",
		BaseURL:      "https://api.openai.com/v1",
		DefaultModel: "gpt-5.2",
		Models:       []string{"gpt-5.2", "gpt-5.1", "gpt-4.1"},
		KeyHint:      "sk-",
		Note:         "官方 Chat Completions",
		DocsURL:      "https://platform.openai.com/docs/",
	},
	{
		ID:           "gemini",
		Name:         "Gemini",
		Region:       "global",
		BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai/",
		DefaultModel: "gemini-3.8-flash",
		Models:       []string{"gemini-3.8-flash", "gemini-3.8-pro", "gemini-3.5-flash"},
		KeyHint:      "AIza",
		Note:         "Google 的 OpenAI 兼容端点",
		DocsURL:      "https://ai.google.dev/gemini-api/docs/openai",
	},
	{
		ID:           "openrouter",
		Name:         "OpenRouter",
		Region:       "global",
		BaseURL:      "https://openrouter.ai/api/v1",
		DefaultModel: "openrouter/auto",
		Models:       []string{"openrouter/auto", "deepseek/deepseek-chat", "google/gemini-3.8-flash"},
		KeyHint:      "sk-or-",
		Note:         "一个 Key 路由多家模型",
		DocsURL:      "https://openrouter.ai/docs",
	},
	{
		ID:           "ollama",
		Name:         "本地 Ollama",
		Region:       "local",
		BaseURL:      "http://host.docker.internal:11434/v1",
		DefaultModel: "qwen3:8b",
		Models:       []string{"qwen3:8b", "deepseek-r1:7b", "llama3.2:3b"},
		Note:         "容器内通过 host.docker.internal 访问宿主机；本机直跑可改成 localhost",
		DocsURL:      "https://ollama.com/",
	},
	{
		ID:           "custom",
		Name:         "自定义接口",
		Region:       "custom",
		DefaultModel: "",
		Note:         "任意 OpenAI 兼容接口的 Base URL（如 https://api.example.com/v1）",
		Custom:       true,
	},
}

// ProviderPresets 返回一份不会被调用方修改的预置列表。
func ProviderPresets() []ProviderPreset {
	out := make([]ProviderPreset, 0, len(providerPresets))
	for _, p := range providerPresets {
		p.Models = append([]string(nil), p.Models...)
		out = append(out, p)
	}
	return out
}

// FindProviderPreset 按 ID 查找预设。
func FindProviderPreset(id string) (ProviderPreset, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, p := range providerPresets {
		if p.ID == id {
			p.Models = append([]string(nil), p.Models...)
			return p, true
		}
	}
	return ProviderPreset{}, false
}

// NormalizeBaseURL 统一去掉末尾斜杠与完整的 /chat/completions 后缀。
func NormalizeBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	raw = strings.TrimSuffix(raw, "/chat/completions")
	raw = strings.TrimRight(raw, "/")
	return raw
}

// InferProviderID 根据地址反推预置 ID；无法匹配时返回 custom。
func InferProviderID(baseURL string) string {
	base := NormalizeBaseURL(baseURL)
	if base == "" {
		return ""
	}
	for _, p := range providerPresets {
		if p.Custom {
			continue
		}
		if NormalizeBaseURL(p.BaseURL) == base {
			return p.ID
		}
	}
	return "custom"
}

// IsLocalBaseURL 判断是否指向本机/容器宿主机，这类端点通常不需要 Key。
func IsLocalBaseURL(baseURL string) bool {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "host.docker.internal":
		return true
	}
	return false
}
