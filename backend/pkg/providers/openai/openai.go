package openai

import (
	"context"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	OpenAIAgentModel     = "o4-mini"
	OpenAIAssistantModel = "o4-mini"
	OpenAISimpleModel    = "gpt-4.1-mini"
	OpenAIAviserModel    = "o4-mini"
	OpenAIGeneratorModel = "o3"
	OpenAIRefinerModel   = "gpt-4.1"
	OpenAISearcherModel  = "gpt-4.1-mini"
	OpenAIEnricherModel  = "gpt-4.1-mini"
	OpenAICoderModel     = "gpt-4.1"
	OpenAIInstallerModel = "gpt-4.1"
	OpenAIPentesterModel = "o4-mini"
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
		llms.WithMaxTokens(3000),
	}

	agent := []llms.CallOption{
		llms.WithModel(OpenAIAgentModel),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
		llms.WithReasoning(llms.ReasoningLow, 0),
	}

	assistant := []llms.CallOption{
		llms.WithModel(OpenAIAssistantModel),
		llms.WithN(1),
		llms.WithMaxTokens(6000),
		llms.WithReasoning(llms.ReasoningMedium, 0),
	}

	creative := []llms.CallOption{
		llms.WithTemperature(0.7),
		llms.WithTopP(0.8),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	determine := []llms.CallOption{
		llms.WithTemperature(0.2),
		llms.WithTopP(0.1),
		llms.WithN(1),
		llms.WithMaxTokens(6000),
	}

	adviser := []llms.CallOption{
		llms.WithModel(OpenAIAviserModel),
		llms.WithN(1),
		llms.WithMaxTokens(8192),
		llms.WithReasoning(llms.ReasoningMedium, 0),
	}

	pentester := []llms.CallOption{
		llms.WithModel(OpenAIPentesterModel),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
		llms.WithReasoning(llms.ReasoningLow, 0),
	}

	return &openaiProvider{
		llm: client,
		options: map[provider.ProviderOptionsType][]llms.CallOption{
			provider.OptionsTypeSimple:     simple,
			provider.OptionsTypeSimpleJSON: append(simple, llms.WithJSONMode()),
			provider.OptionsTypeAgent:      agent,
			provider.OptionsTypeAssistant:  assistant,
			provider.OptionsTypeGenerator:  append(adviser, llms.WithModel(OpenAIGeneratorModel)),
			provider.OptionsTypeRefiner:    append(creative, llms.WithModel(OpenAIRefinerModel)),
			provider.OptionsTypeAdviser:    append(adviser, llms.WithMaxTokens(4000)),
			provider.OptionsTypeReflector:  append(adviser, llms.WithMaxTokens(3000)),
			provider.OptionsTypeSearcher:   append(creative, llms.WithModel(OpenAISearcherModel)),
			provider.OptionsTypeEnricher:   append(creative, llms.WithModel(OpenAIEnricherModel)),
			provider.OptionsTypeCoder:      append(determine, llms.WithModel(OpenAICoderModel)),
			provider.OptionsTypeInstaller:  append(determine, llms.WithModel(OpenAIInstallerModel)),
			provider.OptionsTypePentester:  pentester,
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

func (p *openaiProvider) CallWithTools(
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
