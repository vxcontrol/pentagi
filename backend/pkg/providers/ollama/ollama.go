package ollama

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"

	"github.com/ollama/ollama/api"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/ollama"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml
var configFS embed.FS

const (
	defaultPullTimeout    = 10 * time.Minute
	defaultAPICallTimeout = 10 * time.Second
)

func BuildProviderConfig(cfg *config.Config, configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithN(1),
		llms.WithMaxTokens(8192),
		llms.WithModel(cfg.OllamaServerModel),
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

func newOllamaClient(serverURL string, httpClient *http.Client) (*api.Client, error) {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ollama server URL: %w", err)
	}

	return api.NewClient(parsedURL, httpClient), nil
}

func loadAvailableModelsFromServer(client *api.Client) (pconfig.ModelsConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPICallTimeout)
	defer cancel()

	response, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	var models pconfig.ModelsConfig
	for _, model := range response.Models {
		modelConfig := pconfig.ModelConfig{
			Name:  model.Name,
			Price: nil, // ollama is free local inference, no pricing
		}
		models = append(models, modelConfig)
	}

	return models, nil
}

func getConfigModelsList(baseModel string, providerConfig *pconfig.ProviderConfig) []string {
	models := []string{baseModel}
	modelsMap := make(map[string]bool)
	modelsMap[baseModel] = true

	configModels := providerConfig.GetModelsMap()

	for _, model := range configModels {
		if !modelsMap[model] {
			models = append(models, model)
			modelsMap[model] = true
		}
	}

	slices.Sort(models)

	return models
}

func ensureModelsAvailable(ctx context.Context, client *api.Client, models []string) error {
	errs := make(chan error, len(models))
	pullProgress := func(api.ProgressResponse) error { return nil }

	for _, model := range models {
		go func(model string) {
			// fast path: if the model already exists locally, skip pulling
			showCtx, cancelShow := context.WithTimeout(ctx, defaultAPICallTimeout)
			defer cancelShow()

			if _, err := client.Show(showCtx, &api.ShowRequest{Model: model}); err == nil {
				// model exists locally, no need to pull
				errs <- nil
				return
			}

			// model doesn't exist, pull it from registry
			errs <- client.Pull(ctx, &api.PullRequest{Model: model}, pullProgress)
		}(model)
	}

	for range len(models) {
		if err := <-errs; err != nil {
			return err
		}
	}

	return nil
}

type ollamaProvider struct {
	llm            *ollama.LLM
	model          string
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	baseModel := cfg.OllamaServerModel
	serverURL := cfg.OllamaServerURL
	timeout := time.Duration(cfg.OllamaServerPullModelsTimeout) * time.Second
	if timeout <= 0 {
		timeout = defaultPullTimeout
	}

	apiClient, err := newOllamaClient(serverURL, httpClient)
	if err != nil {
		return nil, err
	}

	if cfg.OllamaServerPullModelsEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		configModels := getConfigModelsList(baseModel, providerConfig)
		err = ensureModelsAvailable(ctx, apiClient, configModels)
		if err != nil {
			return nil, err
		}
	}

	client, err := ollama.New(
		ollama.WithServerURL(serverURL),
		ollama.WithHTTPClient(httpClient),
		ollama.WithModel(baseModel),
	)
	if err != nil {
		return nil, err
	}

	availableModels := pconfig.ModelsConfig{
		{
			Name: baseModel,
		},
	}
	if cfg.OllamaServerLoadModelsEnabled {
		availableModels, err = loadAvailableModelsFromServer(apiClient)
		if err != nil {
			return nil, err
		}
	}

	return &ollamaProvider{
		llm:            client,
		model:          baseModel,
		models:         availableModels,
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
		switch v := value.(type) {
		case int:
			inputTokens = int64(v)
		case int64:
			inputTokens = v
		case float64:
			inputTokens = int64(v)
		}
	}

	if value, ok := info["CompletionTokens"]; ok {
		switch v := value.(type) {
		case int:
			outputTokens = int64(v)
		case int64:
			outputTokens = v
		case float64:
			outputTokens = int64(v)
		}
	}

	return inputTokens, outputTokens
}
