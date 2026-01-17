package anthropic

import (
	"context"
	"embed"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const AnthropicAgentModel = "claude-sonnet-4-20250514"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	providerConfig, err := pconfig.LoadConfigData(configData, defaultOptions)
	if err != nil {
		return nil, err
	}

	return providerConfig, nil
}

func DefaultProviderConfig() (*pconfig.ProviderConfig, error) {
	configData, err := configFS.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}

	return BuildProviderConfig(configData)
}

func DefaultModels() (pconfig.ModelsConfig, error) {
	configData, err := configFS.ReadFile("models.yml")
	if err != nil {
		return nil, err
	}

	return pconfig.LoadModelsConfigData(configData)
}

type anthropicProvider struct {
	llm            *anthropic.LLM
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	baseURL := cfg.AnthropicServerURL
	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := anthropic.New(
		anthropic.WithToken(cfg.AnthropicAPIKey),
		anthropic.WithModel(AnthropicAgentModel),
		anthropic.WithBaseURL(baseURL),
		anthropic.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	return &anthropicProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
	}, nil
}

func (p *anthropicProvider) Type() provider.ProviderType {
	return provider.ProviderAnthropic
}

func (p *anthropicProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *anthropicProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *anthropicProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *anthropicProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *anthropicProvider) Model(opt pconfig.ProviderOptionsType) string {
	opts := llms.CallOptions{Model: AnthropicAgentModel}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.Model
}

func (p *anthropicProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *anthropicProvider) CallEx(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	return provider.WrapGenerateContent(
		ctx, p, opt, p.llm.GenerateContent, chain,
		append([]llms.CallOption{
			llms.WithStreamingFunc(streamCb),
		}, p.providerConfig.GetOptionsForType(opt)...)...,
	)
}

func (p *anthropicProvider) CallWithTools(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	return provider.WrapGenerateContent(
		ctx, p, opt, p.llm.GenerateContent, chain,
		append([]llms.CallOption{
			llms.WithTools(tools),
			llms.WithStreamingFunc(streamCb),
		}, p.providerConfig.GetOptionsForType(opt)...)...,
	)
}

func (p *anthropicProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64

	if value, ok := info["InputTokens"]; ok {
		switch v := value.(type) {
		case int:
			inputTokens = int64(v)
		case int64:
			inputTokens = v
		case float64:
			inputTokens = int64(v)
		}
	}

	if value, ok := info["OutputTokens"]; ok {
		switch v := value.(type) {
		case int:
			outputTokens = int64(v)
		case int64:
			outputTokens = v
		case float64:
			outputTokens = int64(v)
		}
	}

	return inputTokens, outputTokens
}
