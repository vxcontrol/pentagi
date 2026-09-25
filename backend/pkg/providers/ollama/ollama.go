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
	"pentagi/pkg/templates"

	"github.com/ollama/ollama/api"
	ollamamodel "github.com/ollama/ollama/types/model"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/ollama"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
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
		llms.WithMaxTokens(32768),
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

func loadAvailableModelsFromServer(client *api.Client, serverURL string) (pconfig.ModelsConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPICallTimeout)
	defer cancel()

	response, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	cloud := isCloudServer(serverURL)

	var models pconfig.ModelsConfig
	for _, model := range response.Models {
		modelConfig := pconfig.ModelConfig{
			Name:     model.Name,
			Thinking: thinkingCapability(model.Name, model.Capabilities),
		}
		if cloud {
			modelConfig.Price = cloudPriceFor(model.Name)
		}
		models = append(models, modelConfig)
	}

	return models, nil
}

func thinkingCapability(name string, declared []ollamamodel.Capability) *bool {
	var thinking bool
	if len(declared) > 0 {
		thinking = slices.Contains(declared, ollamamodel.CapabilityThinking)
	} else {
		thinking = reasoning.IsReasoningModel(name)
	}
	if !thinking {
		return nil
	}

	return &thinking
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

func catalogFromConfig(baseModel, serverURL string, providerConfig *pconfig.ProviderConfig) pconfig.ModelsConfig {
	cloud := isCloudServer(serverURL)

	names := getConfigModelsList(baseModel, providerConfig)
	models := make(pconfig.ModelsConfig, 0, len(names))
	for _, name := range names {
		modelConfig := pconfig.ModelConfig{
			Name:     name,
			Thinking: thinkingCapability(name, nil),
		}
		if cloud {
			modelConfig.Price = cloudPriceFor(name)
		}
		models = append(models, modelConfig)
	}

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
	providerName   provider.ProviderName
	providerConfig *pconfig.ProviderConfig
}

func New(
	cfg *config.Config,
	providerName provider.ProviderName,
	providerConfig *pconfig.ProviderConfig,
) (provider.Provider, error) {
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

	options := []ollama.Option{
		ollama.WithServerURL(serverURL),
		ollama.WithHTTPClient(httpClient),
		ollama.WithModel(baseModel),
	}

	// Add API key for Ollama Cloud support
	if cfg.OllamaServerAPIKey != "" {
		options = append(options, ollama.WithAPIKey(cfg.OllamaServerAPIKey))
	}
	if isCloudServer(serverURL) {
		options = append(options, ollama.WithCloudStructuredOutputFallback())
	}

	client, err := ollama.New(options...)
	if err != nil {
		return nil, err
	}

	var availableModels pconfig.ModelsConfig
	if cfg.OllamaServerLoadModelsEnabled {
		availableModels, err = loadAvailableModelsFromServer(apiClient, serverURL)
		if err != nil {
			return nil, err
		}
	} else {
		availableModels = catalogFromConfig(baseModel, serverURL, providerConfig)
	}

	return &ollamaProvider{
		llm:            client,
		model:          baseModel,
		models:         availableModels,
		providerName:   providerName,
		providerConfig: providerConfig,
	}, nil
}

func (p *ollamaProvider) Type() provider.ProviderType {
	return provider.ProviderOllama
}

func (p *ollamaProvider) Name() provider.ProviderName {
	return p.providerName
}

func (p *ollamaProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *ollamaProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *ollamaProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	if price := p.providerConfig.GetPriceInfoForType(opt); price != nil {
		return price
	}

	model := p.model
	if agent := p.providerConfig.AgentConfigForType(opt); agent != nil && agent.Model != "" {
		model = agent.Model
	}

	for _, available := range p.models {
		if available.Name == model {
			return available.Price
		}
	}

	return nil
}

func (p *ollamaProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *ollamaProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := p.model
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *ollamaProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	// Ollama provider doesn't need prefix support (passthrough mode in LiteLLM)
	return p.Model(opt)
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

func (p *ollamaProvider) CallWithExtraOptions(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
	extra ...llms.CallOption,
) (*llms.ContentResponse, error) {
	options := []llms.CallOption{llms.WithStreamingFunc(streamCb)}
	if len(tools) > 0 {
		options = append(options, llms.WithTools(tools))
	}
	options = append(options, p.providerConfig.GetOptionsForType(opt)...)
	options = append(options, extra...)

	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
}

func (p *ollamaProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *ollamaProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, "")
}
