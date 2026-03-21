package glm

import (
	"context"
	"embed"
	"fmt"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const GLMAgentModel = "glm-4.7-flashx"

const GLMToolCallIDTemplate = "call_-{r:19:d}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(GLMAgentModel),
		llms.WithN(1),
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

type glmProvider struct {
	llm            *openai.LLM
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig
	providerPrefix string
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	if cfg.GLMAPIKey == "" {
		return nil, fmt.Errorf("missing GLM_API_KEY environment variable")
	}

	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := openai.New(
		openai.WithToken(cfg.GLMAPIKey),
		openai.WithModel(GLMAgentModel),
		openai.WithBaseURL(cfg.GLMServerURL),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	return &glmProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
		providerPrefix: cfg.GLMProvider,
	}, nil
}

func (p *glmProvider) Type() provider.ProviderType {
	return provider.ProviderGLM
}

func (p *glmProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *glmProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *glmProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *glmProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *glmProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := GLMAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *glmProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return provider.ApplyModelPrefix(p.Model(opt), p.providerPrefix)
}

func (p *glmProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *glmProvider) CallEx(
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

func (p *glmProvider) CallWithTools(
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

func (p *glmProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *glmProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, GLMToolCallIDTemplate)
}
