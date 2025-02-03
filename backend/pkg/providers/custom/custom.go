package custom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type customProvider struct {
	llm      *openai.LLM
	model    string
	embedder *embeddings.EmbedderImpl
	options  map[provider.ProviderOptionsType][]llms.CallOption
}

func New(cfg *config.Config) (provider.Provider, error) {
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

	baseKey := cfg.LLMServerKey
	baseURL := cfg.LLMServerURL
	baseModel := cfg.LLMServerModel
	client, err := openai.New(
		openai.WithToken(baseKey),
		openai.WithModel(baseModel),
		openai.WithBaseURL(baseURL),
		openai.WithHTTPClient(httpClient),
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
		llms.WithTemperature(0.7),
		llms.WithTopP(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	// load provider options from config file if specified
	var options map[provider.ProviderOptionsType][]llms.CallOption
	if cfg.LLMServerConfig != "" {
		providerConfig, err := LoadConfig(cfg.LLMServerConfig, simple)
		if err != nil {
			return nil, fmt.Errorf("failed to load provider config: %w", err)
		}
		if providerConfig != nil {
			options = make(map[provider.ProviderOptionsType][]llms.CallOption)
			for _, optType := range []provider.ProviderOptionsType{
				provider.OptionsTypeSimple,
				provider.OptionsTypeSimpleJSON,
				provider.OptionsTypeAgent,
				provider.OptionsTypeGenerator,
				provider.OptionsTypeRefiner,
				provider.OptionsTypeAdviser,
				provider.OptionsTypeReflector,
				provider.OptionsTypeSearcher,
				provider.OptionsTypeEnricher,
				provider.OptionsTypeCoder,
				provider.OptionsTypeInstaller,
				provider.OptionsTypePentester,
			} {
				if opts := providerConfig.GetOptionsForType(optType); opts != nil {
					options[optType] = opts
				}
			}
		}
	}

	// if no config file or empty config, use default options
	if options == nil {
		options = map[provider.ProviderOptionsType][]llms.CallOption{
			provider.OptionsTypeSimple:     simple,
			provider.OptionsTypeSimpleJSON: simple,
			provider.OptionsTypeAgent:      simple,
			provider.OptionsTypeGenerator:  simple,
			provider.OptionsTypeRefiner:    simple,
			provider.OptionsTypeAdviser:    simple,
			provider.OptionsTypeReflector:  simple,
			provider.OptionsTypeSearcher:   simple,
			provider.OptionsTypeEnricher:   simple,
			provider.OptionsTypeCoder:      simple,
			provider.OptionsTypeInstaller:  simple,
			provider.OptionsTypePentester:  simple,
		}
	}

	return &customProvider{
		llm:      client,
		model:    baseModel,
		embedder: embedder,
		options:  options,
	}, nil
}

func (p *customProvider) Type() provider.ProviderType {
	return provider.ProviderCustom
}

func (p *customProvider) Model(opt provider.ProviderOptionsType) string {
	options, ok := p.options[opt]
	if !ok {
		return p.model
	}

	opts := llms.CallOptions{Model: p.model}
	for _, option := range options {
		option(&opts)
	}

	return opts.Model
}

func (p *customProvider) Embedder() *embeddings.EmbedderImpl {
	return p.embedder
}

func (p *customProvider) Call(
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

func (p *customProvider) CallEx(
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

func (p *customProvider) CallWithTools(
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

func (p *customProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64
	if value, ok := info["PromptTokens"]; ok {
		inputTokens = int64(value.(int))
	}

	if value, ok := info["CompletionTokens"]; ok {
		outputTokens = int64(value.(int))
	}

	return inputTokens, outputTokens
}
