package custom

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type modelsResponse struct {
	Data []modelInfo `json:"data"`
}

type modelInfo struct {
	ID                  string       `json:"id"`
	Created             *int64       `json:"created,omitempty"`
	Description         string       `json:"description,omitempty"`
	SupportedParameters []string     `json:"supported_parameters,omitempty"`
	Pricing             *pricingInfo `json:"pricing,omitempty"`
}

type fallbackModelInfo struct {
	ID string `json:"id"`
}

type fallbackModelsResponse struct {
	Data []fallbackModelInfo `json:"data"`
}

type pricingInfo struct {
	Prompt     string `json:"prompt,omitempty"`
	Completion string `json:"completion,omitempty"`
}

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

	models := loadModelsFromServer(baseURL, baseKey, httpClient)

	return &customProvider{
		llm:            client,
		model:          baseModel,
		models:         models,
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

func (p *customProvider) GetModels() pconfig.ModelsConfig {
	return p.models
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

func loadModelsFromServer(baseURL, baseKey string, client *http.Client) pconfig.ModelsConfig {
	modelsURL := strings.TrimRight(baseURL, "/") + "/models"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	if baseKey != "" {
		req.Header.Set("Authorization", "Bearer "+baseKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	// Read the response body once
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// Try to parse with full structure first
	var response modelsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// Fallback to simplified structure if main parsing fails
		var fallbackResponse fallbackModelsResponse
		if err := json.Unmarshal(body, &fallbackResponse); err != nil {
			return nil
		}

		// Convert fallback models to ModelsConfig
		var models pconfig.ModelsConfig
		for _, model := range fallbackResponse.Data {
			modelConfig := pconfig.ModelConfig{
				Name: model.ID,
			}
			models = append(models, modelConfig)
		}
		return models
	}

	var models pconfig.ModelsConfig
	for _, model := range response.Data {
		modelConfig := pconfig.ModelConfig{
			Name: model.ID,
		}

		// Parse description if available
		if model.Description != "" {
			modelConfig.Description = &model.Description
		}

		// Parse created timestamp to release_date if available
		if model.Created != nil && *model.Created > 0 {
			releaseDate := time.Unix(*model.Created, 0).UTC()
			modelConfig.ReleaseDate = &releaseDate
		}

		// Check for reasoning support in supported_parameters
		if len(model.SupportedParameters) > 0 {
			thinking := slices.Contains(model.SupportedParameters, "reasoning")
			modelConfig.Thinking = &thinking
		}

		// Check for tool support
		if len(model.SupportedParameters) > 0 {
			// Skip models if we sure they don't support tools and structured outputs
			hasTools := slices.Contains(model.SupportedParameters, "tools")
			hasStructuredOutputs := slices.Contains(model.SupportedParameters, "structured_outputs")
			if !hasTools && !hasStructuredOutputs {
				continue
			}
		}

		// Parse pricing if available
		if model.Pricing != nil {
			if input, err := strconv.ParseFloat(model.Pricing.Prompt, 64); err == nil {
				if output, err := strconv.ParseFloat(model.Pricing.Completion, 64); err == nil {
					// convert per-token prices to per-million-token if needed
					if input < 0.01 {
						input *= 1000000
					}
					if output < 0.01 {
						output *= 1000000
					}

					modelConfig.Price = &pconfig.PriceInfo{
						Input:  input,
						Output: output,
					}
				}
			}
		}

		models = append(models, modelConfig)
	}

	return models
}
