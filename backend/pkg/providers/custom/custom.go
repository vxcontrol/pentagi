package custom

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func BuildProviderConfig(cfg *config.Config, configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithTemperature(0.7),
		llms.WithTopP(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	if cfg.LLMServerModel != "" {
		defaultOptions = append(defaultOptions, llms.WithModel(cfg.LLMServerModel))
	}

	providerConfig, err := pconfig.LoadConfigData(configData, defaultOptions)
	if err != nil {
		return nil, err
	}

	return providerConfig, nil
}

func DefaultProviderConfig(cfg *config.Config) (*pconfig.ProviderConfig, error) {
	if cfg.LLMServerConfig == "" {
		return BuildProviderConfig(cfg, []byte("{}"))
	}

	configData, err := os.ReadFile(cfg.LLMServerConfig)
	if err != nil {
		return nil, err
	}

	return BuildProviderConfig(cfg, configData)
}

type customProvider struct {
	llm            *openai.LLM
	model          string
	providerConfig *pconfig.ProviderConfig
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
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

	return &customProvider{
		llm:            client,
		model:          baseModel,
		providerConfig: providerConfig,
	}, nil
}

func (p *customProvider) Type() provider.ProviderType {
	return provider.ProviderCustom
}

func (p *customProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *customProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *customProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *customProvider) Model(opt pconfig.ProviderOptionsType) string {
	opts := llms.CallOptions{Model: p.model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.Model
}

func (p *customProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *customProvider) CallEx(
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

func (p *customProvider) CallWithTools(
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
