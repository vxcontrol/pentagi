package custom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type customProvider struct {
	llm     *openai.LLM
	model   string
	options map[provider.ProviderOptionsType][]llms.CallOption
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
	opts := []openai.Option{
		openai.WithToken(baseKey),
		openai.WithModel(baseModel),
		openai.WithBaseURL(baseURL),
		openai.WithHTTPClient(httpClient),
	}
	if !cfg.LLMServerLegacyReasoning {
		opts = append(opts,
			openai.WithUsingReasoningMaxTokens(),
			openai.WithModernReasoningFormat(),
		)
	}
	client, err := openai.New(opts...)
	if err != nil {
		return nil, err
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
				provider.OptionsTypeAssistant,
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
			provider.OptionsTypeAssistant:  simple,
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
		llm:     client,
		model:   baseModel,
		options: options,
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

func (p *customProvider) CallWithTools(
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
