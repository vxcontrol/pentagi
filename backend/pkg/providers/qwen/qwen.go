package qwen

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

const QwenAgentModel = "qwen-plus"

const QwenToolCallIDTemplate = "call_{r:24:h}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(QwenAgentModel),
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

type qwenProvider struct {
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
	if cfg.QwenAPIKey == "" {
		return nil, fmt.Errorf("missing QWEN_API_KEY environment variable")
	}

	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	// Alibaba Cloud DashScope OpenAI-compatible API. Thinking is controlled via
	// extra_body.enable_thinking (true/false) in the per-agent config — this is a
	// DashScope-specific parameter, not OpenAI standard. Qwen3.5/3.6/3.7 hybrid
	// models have thinking ENABLED by default, so utility agents must explicitly
	// set enable_thinking=false or reasoning_content will be returned inline as
	// part of content (corrupting outputs that expect short deterministic answers
	// like docker image selection or descriptors).
	// WithPreserveReasoningContent() is required for multi-turn with tool calls
	// when preserve_thinking=true is set in extra_body for qwen3.7-max/qwen3.6-plus
	// (other Qwen3 models do not support preserve_thinking).
	client, err := openai.New(
		openai.WithToken(cfg.QwenAPIKey),
		openai.WithModel(QwenAgentModel),
		openai.WithBaseURL(cfg.QwenServerURL),
		openai.WithHTTPClient(httpClient),
		openai.WithPreserveReasoningContent(),
	)
	if err != nil {
		return nil, err
	}

	return &qwenProvider{
		llm:            client,
		models:         models,
		providerName:   providerName,
		providerConfig: providerConfig,
		providerPrefix: cfg.QwenProvider,
	}, nil
}

func (p *qwenProvider) Type() provider.ProviderType {
	return provider.ProviderQwen
}

func (p *qwenProvider) Name() provider.ProviderName {
	return p.providerName
}

func (p *qwenProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *qwenProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *qwenProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *qwenProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *qwenProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := QwenAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *qwenProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return provider.ApplyModelPrefix(p.Model(opt), p.providerPrefix)
}

func (p *qwenProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *qwenProvider) CallEx(
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

func (p *qwenProvider) CallWithTools(
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

func (p *qwenProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *qwenProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, QwenToolCallIDTemplate)
}
