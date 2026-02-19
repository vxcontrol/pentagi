package bedrock

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	bconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const BedrockAgentModel = bedrock.ModelAnthropicClaudeSonnet4

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(BedrockAgentModel),
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

type bedrockProvider struct {
	llm            *bedrock.LLM
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig

	toolCallIDTemplate     string
	toolCallIDTemplateOnce sync.Once
	toolCallIDTemplateErr  error
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	opts := []func(*bconfig.LoadOptions) error{
		bconfig.WithRegion(cfg.BedrockRegion),
		bconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.BedrockAccessKey,
			cfg.BedrockSecretKey,
			cfg.BedrockSessionToken,
		)),
	}

	if cfg.BedrockServerURL != "" {
		opts = append(opts, bconfig.WithBaseEndpoint(cfg.BedrockServerURL))
	}

	if cfg.ProxyURL != "" {
		opts = append(opts, bconfig.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(cfg.ProxyURL)
				},
			},
		}))
	}

	bcfg, err := bconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	bclient := bedrockruntime.NewFromConfig(bcfg)

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := bedrock.New(
		bedrock.WithClient(bclient),
		bedrock.WithModel(BedrockAgentModel),
		bedrock.WithConverseAPI(),
	)
	if err != nil {
		return nil, err
	}

	return &bedrockProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
	}, nil
}

func (p *bedrockProvider) Type() provider.ProviderType {
	return provider.ProviderBedrock
}

func (p *bedrockProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *bedrockProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *bedrockProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *bedrockProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *bedrockProvider) Model(opt pconfig.ProviderOptionsType) string {
	opts := llms.CallOptions{Model: BedrockAgentModel}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.Model
}

func (p *bedrockProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *bedrockProvider) CallEx(
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

func (p *bedrockProvider) CallWithTools(
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

func (p *bedrockProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *bedrockProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter)
}
