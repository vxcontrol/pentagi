// Package openaicompat is the shared implementation for the OpenAI-compatible
// providers that differ only in constants and credentials (deepseek, glm, kimi,
// qwen, minimax). The native OpenAI provider is intentionally NOT built on this:
// it has no model-prefix support and no reasoning-preservation option.
package openaicompat

import (
	"context"
	"fmt"
	"net/http"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// Spec is the per-provider data the shared base needs.
type Spec struct {
	Type               provider.ProviderType
	Model              string // fallback agent model
	ToolCallIDTemplate string
	APIKey             string
	ServerURL          string
	Prefix             string // LiteLLM model-name prefix (empty disables prefixing)
	PreserveReasoning  bool   // adds openai.WithPreserveReasoningContent() for thinking models
}

// Provider implements provider.Provider over an OpenAI-compatible endpoint.
type Provider struct {
	llm                *openai.LLM
	models             pconfig.ModelsConfig
	providerName       provider.ProviderName
	providerConfig     *pconfig.ProviderConfig
	providerType       provider.ProviderType
	prefix             string
	agentModel         string
	toolCallIDTemplate string
}

// New builds an OpenAI-compatible provider from spec. The caller supplies the
// shared HTTP client and the already-loaded model catalog.
func New(
	spec Spec,
	httpClient *http.Client,
	models pconfig.ModelsConfig,
	providerName provider.ProviderName,
	providerConfig *pconfig.ProviderConfig,
) (provider.Provider, error) {
	if spec.APIKey == "" {
		return nil, fmt.Errorf("missing API key for provider %q", spec.Type)
	}

	opts := []openai.Option{
		openai.WithToken(spec.APIKey),
		openai.WithModel(spec.Model),
		openai.WithBaseURL(spec.ServerURL),
		openai.WithHTTPClient(httpClient),
	}
	// Do NOT add openai.WithModernReasoningFormat() to these shared opts: DeepSeek's
	// API requires the legacy top-level "reasoning_effort" string, which langchaingo
	// emits only while ModernReasoningFormat is off (the default). The modern
	// "reasoning":{} object form breaks DeepSeek thinking mode.
	if spec.PreserveReasoning {
		opts = append(opts, openai.WithPreserveReasoningContent())
	}

	client, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return &Provider{
		llm:                client,
		models:             models,
		providerName:       providerName,
		providerConfig:     providerConfig,
		providerType:       spec.Type,
		prefix:             spec.Prefix,
		agentModel:         spec.Model,
		toolCallIDTemplate: spec.ToolCallIDTemplate,
	}, nil
}

func (p *Provider) Type() provider.ProviderType {
	return p.providerType
}

func (p *Provider) Name() provider.ProviderName {
	return p.providerName
}

func (p *Provider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *Provider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *Provider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *Provider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *Provider) Model(opt pconfig.ProviderOptionsType) string {
	model := p.agentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *Provider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return provider.ApplyModelPrefix(p.Model(opt), p.prefix)
}

func (p *Provider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *Provider) CallEx(
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

func (p *Provider) CallWithTools(
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

// CallWithExtraOptions: extra is appended last, so it overrides the config.
// Shared by deepseek/glm/kimi/qwen/minimax.
func (p *Provider) CallWithExtraOptions(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
	extra ...llms.CallOption,
) (*llms.ContentResponse, error) {
	options := []llms.CallOption{llms.WithStreamingFunc(streamCb)}
	if len(tools) > 0 {
		options = append(options, llms.WithTools(tools))
	}
	options = append(options, p.providerConfig.GetOptionsForType(opt)...)
	options = append(options, extra...)

	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
}

func (p *Provider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *Provider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, p.toolCallIDTemplate)
}
