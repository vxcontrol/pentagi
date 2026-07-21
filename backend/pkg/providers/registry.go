package providers

import (
	"pentagi/pkg/config"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/deepseek"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/glm"
	"pentagi/pkg/providers/kimi"
	"pentagi/pkg/providers/minimax"
	"pentagi/pkg/providers/ollama"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/qwen"
)

// registryEntry describes how to construct one provider type. It absorbs the
// small per-provider variance — which constructors take *config.Config and the
// credential gating — so the controller can wire every provider in one loop.
type registryEntry struct {
	Type        provider.ProviderType
	Name        provider.ProviderName
	Enabled     func(*config.Config) bool
	NewConfig   func(*config.Config) (*pconfig.ProviderConfig, error)
	New         func(*config.Config, provider.ProviderName, *pconfig.ProviderConfig) (provider.Provider, error)
	BuildConfig func(*config.Config, []byte) (*pconfig.ProviderConfig, error)
}

// ignoreConfig adapts a no-argument default-config loader to the registry signature.
func ignoreConfig(
	fn func() (*pconfig.ProviderConfig, error),
) func(*config.Config) (*pconfig.ProviderConfig, error) {
	return func(*config.Config) (*pconfig.ProviderConfig, error) { return fn() }
}

// fromData adapts a []byte-only config builder to the registry signature.
func fromData(
	fn func([]byte) (*pconfig.ProviderConfig, error),
) func(*config.Config, []byte) (*pconfig.ProviderConfig, error) {
	return func(_ *config.Config, data []byte) (*pconfig.ProviderConfig, error) { return fn(data) }
}

var providerRegistry = []registryEntry{
	{
		Type:        provider.ProviderOpenAI,
		Name:        provider.DefaultProviderNameOpenAI,
		Enabled:     func(c *config.Config) bool { return c.OpenAIKey != "" },
		NewConfig:   ignoreConfig(openai.DefaultProviderConfig),
		New:         openai.New,
		BuildConfig: fromData(openai.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderAnthropic,
		Name:        provider.DefaultProviderNameAnthropic,
		Enabled:     func(c *config.Config) bool { return c.AnthropicAPIKey != "" },
		NewConfig:   ignoreConfig(anthropic.DefaultProviderConfig),
		New:         anthropic.New,
		BuildConfig: fromData(anthropic.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderGemini,
		Name:        provider.DefaultProviderNameGemini,
		Enabled:     func(c *config.Config) bool { return c.GeminiAPIKey != "" },
		NewConfig:   ignoreConfig(gemini.DefaultProviderConfig),
		New:         gemini.New,
		BuildConfig: fromData(gemini.BuildProviderConfig),
	},
	{
		Type: provider.ProviderBedrock,
		Name: provider.DefaultProviderNameBedrock,
		// Bedrock accepts default AWS SDK auth, a bearer token, or static credentials.
		Enabled: func(c *config.Config) bool {
			return c.BedrockDefaultAuth || c.BedrockBearerToken != "" ||
				(c.BedrockAccessKey != "" && c.BedrockSecretKey != "")
		},
		NewConfig:   bedrock.DefaultProviderConfig,
		New:         bedrock.New,
		BuildConfig: fromData(bedrock.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderOllama,
		Name:        provider.DefaultProviderNameOllama,
		Enabled:     func(c *config.Config) bool { return c.OllamaServerURL != "" },
		NewConfig:   ollama.DefaultProviderConfig,
		New:         ollama.New,
		BuildConfig: ollama.BuildProviderConfig,
	},
	{
		Type: provider.ProviderCustom,
		Name: provider.DefaultProviderNameCustom,
		Enabled: func(c *config.Config) bool {
			return c.LLMServerURL != "" && (c.LLMServerModel != "" || c.LLMServerConfig != "")
		},
		NewConfig:   custom.DefaultProviderConfig,
		New:         custom.New,
		BuildConfig: custom.BuildProviderConfig,
	},
	{
		Type:        provider.ProviderDeepSeek,
		Name:        provider.DefaultProviderNameDeepSeek,
		Enabled:     func(c *config.Config) bool { return c.DeepSeekAPIKey != "" },
		NewConfig:   ignoreConfig(deepseek.DefaultProviderConfig),
		New:         deepseek.New,
		BuildConfig: fromData(deepseek.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderGLM,
		Name:        provider.DefaultProviderNameGLM,
		Enabled:     func(c *config.Config) bool { return c.GLMAPIKey != "" },
		NewConfig:   ignoreConfig(glm.DefaultProviderConfig),
		New:         glm.New,
		BuildConfig: fromData(glm.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderKimi,
		Name:        provider.DefaultProviderNameKimi,
		Enabled:     func(c *config.Config) bool { return c.KimiAPIKey != "" },
		NewConfig:   ignoreConfig(kimi.DefaultProviderConfig),
		New:         kimi.New,
		BuildConfig: fromData(kimi.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderQwen,
		Name:        provider.DefaultProviderNameQwen,
		Enabled:     func(c *config.Config) bool { return c.QwenAPIKey != "" },
		NewConfig:   ignoreConfig(qwen.DefaultProviderConfig),
		New:         qwen.New,
		BuildConfig: fromData(qwen.BuildProviderConfig),
	},
	{
		Type:        provider.ProviderMiniMax,
		Name:        provider.DefaultProviderNameMiniMax,
		Enabled:     func(c *config.Config) bool { return c.MiniMaxAPIKey != "" },
		NewConfig:   ignoreConfig(minimax.DefaultProviderConfig),
		New:         minimax.New,
		BuildConfig: fromData(minimax.BuildProviderConfig),
	},
}

func entryForType(t provider.ProviderType) (registryEntry, bool) {
	for _, e := range providerRegistry {
		if e.Type == t {
			return e, true
		}
	}
	return registryEntry{}, false
}
