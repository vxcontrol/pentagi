package ollama

import (
	"context"
	"embed"
	"net/http"
	"net/url"
	"os"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/ollama/ollama/api"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/ollama"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml
var configFS embed.FS

const OllamaAgentModel = "llama3.1:8b"

func BuildProviderConfig(cfg *config.Config, configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithN(1),
		llms.WithMaxTokens(4000),
		llms.WithModel(OllamaAgentModel),
	}

	providerConfig, err := pconfig.LoadConfigData(configData, defaultOptions)
	if err != nil {
		return nil, err
	}

	return providerConfig, nil
}

func DefaultProviderConfig(cfg *config.Config) (*pconfig.ProviderConfig, error) {
	var (
		configData []byte
		err        error
	)

	if cfg.OllamaServerConfig == "" {
		configData, err = configFS.ReadFile("config.yml")
	} else {
		configData, err = os.ReadFile(cfg.OllamaServerConfig)
	}
	if err != nil {
		return nil, err
	}

	return BuildProviderConfig(cfg, configData)
}

func loadModelsFromServer(serverURL string, httpClient *http.Client) pconfig.ModelsConfig {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil
	}

	client := api.NewClient(parsedURL, httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := client.List(ctx)
	if err != nil {
		return nil
	}

	var models pconfig.ModelsConfig
	for _, model := range response.Models {
		modelConfig := pconfig.ModelConfig{
			Name:  model.Name,
			Price: nil, // ollama is free local inference, no pricing
		}
		models = append(models, modelConfig)
	}

	return models
}

type ollamaProvider struct {
	llm            *ollama.LLM
	model          string
	models         pconfig.ModelsConfig
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

	model := OllamaAgentModel
	serverURL := cfg.OllamaServerURL

	client, err := ollama.New(
		ollama.WithServerURL(serverURL),
		ollama.WithHTTPClient(httpClient),
		ollama.WithModel(model),
		ollama.WithPullModel(),
	)
	if err != nil {
		return nil, err
	}

	models := loadModelsFromServer(serverURL, httpClient)

	return &ollamaProvider{
		llm:            client,
		model:          model,
		models:         models,
		providerConfig: providerConfig,
	}, nil
}

func (p *ollamaProvider) Type() provider.ProviderType {
	return provider.ProviderOllama
}

func (p *ollamaProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *ollamaProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *ollamaProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *ollamaProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *ollamaProvider) Model(opt pconfig.ProviderOptionsType) string {
	opts := llms.CallOptions{Model: p.model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.Model
}

func (p *ollamaProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *ollamaProvider) CallEx(
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

func (p *ollamaProvider) CallWithTools(
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

func (p *ollamaProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64
	if value, ok := info["PromptTokens"]; ok {
		inputTokens = int64(value.(int))
	}

	if value, ok := info["CompletionTokens"]; ok {
		outputTokens = int64(value.(int))
	}

	return inputTokens, outputTokens
}
