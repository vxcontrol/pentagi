package openai

import (
	"context"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	OpenAIAgentModel     = "gpt-4o"
	OpenAISimpleModel    = "gpt-4o-mini"
	OpenAIAviserModel    = "o1-preview"
	OpenAIGeneratorModel = "gpt-4"
	OpenAIRefinerModel   = "gpt-4o"
	OpenAISearcherModel  = "gpt-4o"
	OpenAIEnricherModel  = "gpt-4"
	OpenAICoderModel     = "gpt-4o"
	OpenAIInstallerModel = "gpt-4o"
	OpenAIPentesterModel = "gpt-4o"
)

type openaiProvider struct {
	llm     *openai.LLM
	options map[provider.ProviderOptionsType][]llms.CallOption
}

func New(cfg *config.Config) (provider.Provider, error) {
	baseURL := cfg.OpenAIServerURL
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

	client, err := openai.New(
		openai.WithToken(cfg.OpenAIKey),
		openai.WithModel(OpenAIAgentModel),
		openai.WithBaseURL(baseURL),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	simple := []llms.CallOption{
		llms.WithModel(OpenAISimpleModel),
		llms.WithTemperature(0.5),
		llms.WithTopP(0.5),
		llms.WithN(1),
	}

	creative := []llms.CallOption{
		llms.WithModel(OpenAIAgentModel),
		llms.WithTemperature(0.7),
		llms.WithTopP(0.8),
		llms.WithN(1),
	}

	determine := []llms.CallOption{
		llms.WithModel(OpenAIAviserModel),
		llms.WithTemperature(0.2),
		llms.WithTopP(0.1),
		llms.WithN(1),
	}

	adviser := []llms.CallOption{
		llms.WithModel(OpenAIAviserModel),
		llms.WithTemperature(1.0),
		llms.WithN(1),
	}

	return &openaiProvider{
		llm: client,
		options: map[provider.ProviderOptionsType][]llms.CallOption{
			provider.OptionsTypeSimple:     simple,
			provider.OptionsTypeSimpleJSON: append(simple, llms.WithJSONMode()),
			provider.OptionsTypeAgent:      append(creative, llms.WithModel(OpenAIAgentModel)),
			provider.OptionsTypeGenerator:  append(creative, llms.WithModel(OpenAIGeneratorModel)),
			provider.OptionsTypeRefiner:    append(creative, llms.WithModel(OpenAIRefinerModel)),
			provider.OptionsTypeAdviser:    adviser,
			provider.OptionsTypeReflector:  adviser,
			provider.OptionsTypeSearcher:   append(creative, llms.WithModel(OpenAISearcherModel)),
			provider.OptionsTypeEnricher:   append(creative, llms.WithModel(OpenAIEnricherModel)),
			provider.OptionsTypeCoder:      append(determine, llms.WithModel(OpenAICoderModel)),
			provider.OptionsTypeInstaller:  append(determine, llms.WithModel(OpenAIInstallerModel)),
			provider.OptionsTypePentester:  append(creative, llms.WithModel(OpenAIPentesterModel)),
		},
	}, nil
}

func (p *openaiProvider) Type() provider.ProviderType {
	return provider.ProviderOpenAI
}

func (p *openaiProvider) Model(opt provider.ProviderOptionsType) string {
	options, ok := p.options[opt]
	if !ok {
		return OpenAIAgentModel
	}

	opts := llms.CallOptions{Model: OpenAIAgentModel}
	for _, option := range options {
		option(&opts)
	}

	return opts.Model
}

func (p *openaiProvider) Call(
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

func (p *openaiProvider) CallEx(
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

func (p *openaiProvider) CallWithTools(
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

func (p *openaiProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64
	if value, ok := info["PromptTokens"]; ok {
		inputTokens = int64(value.(int))
	}

	if value, ok := info["CompletionTokens"]; ok {
		outputTokens = int64(value.(int))
	}

	return inputTokens, outputTokens
}
