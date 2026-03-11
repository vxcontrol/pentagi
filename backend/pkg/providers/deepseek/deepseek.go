package deepseek

import (
	"context"
	"embed"
	"fmt"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const DeepSeekAgentModel = "deepseek-chat"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(DeepSeekAgentModel),
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

type deepseekProvider struct {
	llm            *openai.LLM
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig
	providerPrefix string
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	if cfg.DeepSeekAPIKey == "" {
		return nil, fmt.Errorf("missing DEEPSEEK_API_KEY environment variable")
	}

	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := openai.New(
		openai.WithToken(cfg.DeepSeekAPIKey),
		openai.WithModel(DeepSeekAgentModel),
		openai.WithBaseURL(cfg.DeepSeekServerURL),
		openai.WithHTTPClient(httpClient),
		openai.WithPreserveReasoningContent(),
	)
	if err != nil {
		return nil, err
	}

	return &deepseekProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
		providerPrefix: cfg.DeepSeekProvider,
	}, nil
}

func (p *deepseekProvider) Type() provider.ProviderType {
	return provider.ProviderDeepSeek
}

func (p *deepseekProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *deepseekProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *deepseekProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *deepseekProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *deepseekProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := DeepSeekAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *deepseekProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return provider.ApplyModelPrefix(p.Model(opt), p.providerPrefix)
}

func (p *deepseekProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *deepseekProvider) CallEx(
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

func (p *deepseekProvider) CallWithTools(
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

func (p *deepseekProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *deepseekProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter)
}
