package kimi

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

// KimiAgentModel is the fallback model used when no agent-specific configuration exists.
// kimi-k2.5 is chosen as cost-effective default ($0.60/$3.00 input/output vs $0.95/$4.00 for k2.6).
// All legacy kimi-k2-* models (turbo-preview, 0905-preview, 0711-preview, thinking, thinking-turbo)
// were deprecated by Moonshot on 2026-05-25 and must not be used.
const KimiAgentModel = "kimi-k2.5"

const KimiToolCallIDTemplate = "{f}:{r:1:d}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(KimiAgentModel),
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

type kimiProvider struct {
	llm            *openai.LLM
	models         pconfig.ModelsConfig
	providerName   provider.ProviderName
	providerConfig *pconfig.ProviderConfig
	providerPrefix string
}

func New(
	cfg *config.Config,
	providerName provider.ProviderName,
	providerConfig *pconfig.ProviderConfig,
) (provider.Provider, error) {
	if cfg.KimiAPIKey == "" {
		return nil, fmt.Errorf("missing KIMI_API_KEY environment variable")
	}

	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	// Kimi K2.6/K2.5 OpenAI-compatible API requires legacy string form for reasoning
	// (no WithModernReasoningFormat). Thinking mode is controlled via
	// extra_body.thinking.type ("enabled"/"disabled") in the per-agent config, with
	// thinking.keep="all" required for multi-turn tool calls to preserve historical
	// reasoning_content (otherwise Moonshot returns "thinking is enabled but
	// reasoning_content is missing in assistant tool call message").
	// WithPreserveReasoningContent() keeps reasoning_content in TextContent parts
	// before ToolCall in assistant messages — Kimi rejects requests without it.
	client, err := openai.New(
		openai.WithToken(cfg.KimiAPIKey),
		openai.WithModel(KimiAgentModel),
		openai.WithBaseURL(cfg.KimiServerURL),
		openai.WithHTTPClient(httpClient),
		openai.WithPreserveReasoningContent(),
	)
	if err != nil {
		return nil, err
	}

	return &kimiProvider{
		llm:            client,
		models:         models,
		providerName:   providerName,
		providerConfig: providerConfig,
		providerPrefix: cfg.KimiProvider,
	}, nil
}

func (p *kimiProvider) Type() provider.ProviderType {
	return provider.ProviderKimi
}

func (p *kimiProvider) Name() provider.ProviderName {
	return p.providerName
}

func (p *kimiProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *kimiProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *kimiProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *kimiProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *kimiProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := KimiAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *kimiProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return provider.ApplyModelPrefix(p.Model(opt), p.providerPrefix)
}

func (p *kimiProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *kimiProvider) CallEx(
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

func (p *kimiProvider) CallWithTools(
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

func (p *kimiProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *kimiProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, KimiToolCallIDTemplate)
}
