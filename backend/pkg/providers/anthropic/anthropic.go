package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	AnthropicAgentModel  = "claude-3-5-sonnet-20241022"
	AnthropicSimpleModel = "claude-3-5-haiku-20241022"
	AnthropicAviserModel = "claude-3-opus-20240229"
)

type anthropicProvider struct {
	llm      *anthropic.LLM
	embedder *embeddings.EmbedderImpl
	options  map[provider.ProviderOptionsType][]llms.CallOption
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

	oclient, err := openai.New(
		openai.WithToken(cfg.OpenAIKey),
		openai.WithModel("gpt-4o"),
		openai.WithBaseURL(cfg.OpenAIServerURL),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	embedder, err := embeddings.NewEmbedder(oclient, embeddings.WithStripNewLines(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	simple := []llms.CallOption{
		llms.WithModel(AnthropicSimpleModel),
		llms.WithTemperature(0.5),
		llms.WithTopP(0.5),
		llms.WithN(1),
	}

	creative := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(0.7),
		llms.WithTopP(0.9),
		llms.WithN(1),
	}

	determine := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
		llms.WithTemperature(0.2),
		llms.WithTopP(0.2),
		llms.WithN(1),
	}

	adviser := []llms.CallOption{
		llms.WithModel(AnthropicAviserModel),
		llms.WithTemperature(0.5),
		llms.WithTopP(0.8),
		llms.WithN(1),
	}

	return &anthropicProvider{
		llm:      client,
		embedder: embedder,
		options: map[provider.ProviderOptionsType][]llms.CallOption{
			provider.OptionsTypeSimple:     simple,
			provider.OptionsTypeSimpleJSON: append(simple, llms.WithJSONMode()),
			provider.OptionsTypeAgent:      creative,
			provider.OptionsTypeGenerator:  creative,
			provider.OptionsTypeRefiner:    creative,
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

func (p *anthropicProvider) Embedder() *embeddings.EmbedderImpl {
	return p.embedder
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
) (*llms.ContentResponse, error) {
	options, ok := p.options[opt]
	if !ok {
		return nil, provider.ErrInvalidProviderOptionsType
	}

	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
}

func (p *anthropicProvider) CallWithTools(
	ctx context.Context,
	opt provider.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
) (*llms.ContentResponse, error) {
	options, ok := p.options[opt]
	if !ok {
		return nil, provider.ErrInvalidProviderOptionsType
	}

	options = append(options, llms.WithTools(tools))
	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
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
