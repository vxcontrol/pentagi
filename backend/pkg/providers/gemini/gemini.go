package gemini

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const GeminiAgentModel = "gemini-2.5-flash"

const defaultGeminiHost = "generativelanguage.googleapis.com"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(GeminiAgentModel),
		llms.WithTemperature(1.0),
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

type geminiProvider struct {
	llm            *googleai.GoogleAI
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	opts := []googleai.Option{
		googleai.WithRest(),
		googleai.WithAPIKey(cfg.GeminiAPIKey),
		googleai.WithDefaultModel(GeminiAgentModel),
	}

	if _, err := url.Parse(cfg.GeminiServerURL); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini server URL: %w", err)
	}

	// always use custom transport to ensure API key injection and URL rewriting
	customTransport := &httputil.ApiKeyTransport{
		Transport: http.DefaultTransport,
		APIKey:    cfg.GeminiAPIKey,
		BaseURL:   cfg.GeminiServerURL,
		ProxyURL:  cfg.ProxyURL,
	}

	opts = append(opts, googleai.WithHTTPClient(&http.Client{
		Transport: customTransport,
	}))

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := googleai.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return &geminiProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
	}, nil
}

func (p *geminiProvider) Type() provider.ProviderType {
	return provider.ProviderGemini
}

func (p *geminiProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *geminiProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *geminiProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *geminiProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *geminiProvider) Model(opt pconfig.ProviderOptionsType) string {
	opts := llms.CallOptions{Model: GeminiAgentModel}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.Model
}

func (p *geminiProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *geminiProvider) CallEx(
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

func (p *geminiProvider) CallWithTools(
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

func (p *geminiProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *geminiProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter)
}
