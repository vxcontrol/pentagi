package anthropic

import (
	"context"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	AnthropicAgentModel  = "claude-sonnet-4-20250514"
	AnthropicSimpleModel = "claude-3-5-haiku-20241022"
)

type anthropicProvider struct {
	llm     *anthropic.LLM
	options map[provider.ProviderOptionsType][]llms.CallOption
}

func New(cfg *config.Config) (provider.Provider, error) {
	baseURL := cfg.AnthropicServerURL
	httpClient := http.DefaultClient
	if cfg.ProxyURL != "" {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(cfg.ProxyURL)
				},
			},
		}
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

	simple := []llms.CallOption{
		llms.WithModel(AnthropicSimpleModel),
		llms.WithTemperature(0.5),
		llms.WithTopP(0.5),
		llms.WithN(1),
		llms.WithMaxTokens(3000),
	}

	creative := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	assistant := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(6000),
	}

	determine := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(0.2),
		llms.WithTopP(0.2),
		llms.WithN(1),
		llms.WithMaxTokens(6000),
	}

	adviser := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(1.0),
		llms.WithTopP(0.8),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	return &anthropicProvider{
		llm: client,
		options: map[provider.ProviderOptionsType][]llms.CallOption{
			provider.OptionsTypeSimple:     simple,
			provider.OptionsTypeSimpleJSON: append(simple, llms.WithJSONMode()),
			provider.OptionsTypeAgent:      creative,
			provider.OptionsTypeAssistant:  assistant,
			provider.OptionsTypeGenerator:  append(creative, llms.WithMaxTokens(8192)),
			provider.OptionsTypeRefiner:    append(creative, llms.WithMaxTokens(8192)),
			provider.OptionsTypeAdviser:    adviser,
			provider.OptionsTypeReflector:  adviser,
			provider.OptionsTypeSearcher:   creative,
			provider.OptionsTypeEnricher:   creative,
			provider.OptionsTypeCoder:      determine,
			provider.OptionsTypeInstaller:  determine,
			provider.OptionsTypePentester:  creative,
		},
	}, nil
}

func (p *anthropicProvider) Type() provider.ProviderType {
	return provider.ProviderAnthropic
}

func (p *anthropicProvider) Model(opt provider.ProviderOptionsType) string {
	options, ok := p.options[opt]
	if !ok {
		return AnthropicAgentModel
	}

	opts := llms.CallOptions{Model: AnthropicAgentModel}
	for _, option := range options {
		option(&opts)
	}

	return opts.Model
}

func (p *anthropicProvider) Call(
	ctx context.Context,
	opt provider.ProviderOptionsType,
	prompt string,
) (string, error) {
	options, ok := p.options[opt]
	if !ok {
		return "", provider.ErrInvalidProviderOptionsType
	}

	return provider.WrapGenerateFromSinglePrompt(ctx, p, opt, p.llm, prompt, options...)
}

func (p *anthropicProvider) CallEx(
	ctx context.Context,
	opt provider.ProviderOptionsType,
	chain []llms.MessageContent,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	options, ok := p.options[opt]
	if !ok {
		return nil, provider.ErrInvalidProviderOptionsType
	}

	return provider.WrapGenerateContent(
		ctx, p, opt, p.llm.GenerateContent, chain,
		append([]llms.CallOption{
			llms.WithStreamingFunc(streamCb),
		}, options...)...,
	)
}

func (p *anthropicProvider) CallWithTools(
	ctx context.Context,
	opt provider.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	options, ok := p.options[opt]
	if !ok {
		return nil, provider.ErrInvalidProviderOptionsType
	}

	return provider.WrapGenerateContent(
		ctx, p, opt, p.llm.GenerateContent, chain,
		append([]llms.CallOption{
			llms.WithTools(tools),
			llms.WithStreamingFunc(streamCb),
		}, options...)...,
	)
}

func (p *anthropicProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64
	if value, ok := info["InputTokens"]; ok {
		inputTokens = int64(value.(int))
	}

	if value, ok := info["OutputTokens"]; ok {
		outputTokens = int64(value.(int))
	}

	return inputTokens, outputTokens
}
