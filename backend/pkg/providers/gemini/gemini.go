package gemini

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

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

// apiKeyTransport is a custom HTTP transport wrapper that handles API key injection
// and URL rewriting for Google Gemini API requests.
//
// This is needed because:
// 1. The Google AI library doesn't properly add the API key when using WithHTTPClient
// 2. Some versions of the library don't respect the WithEndpoint option
// 3. LiteLLM passthrough requires both query parameter and Authorization header
type apiKeyTransport struct {
	wrapped   http.RoundTripper // underlying transport (may include proxy settings)
	serverURL *url.URL          // custom server URL to replace the default Gemini endpoint
	apiKey    string            // API key to inject into requests
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// clone the request to avoid modifying the original one
	newReq := req.Clone(req.Context())

	// preserve original query parameters before URL manipulation
	q := newReq.URL.Query()

	// replace base URL if custom server URL is configured
	// this handles cases where the library doesn't respect WithEndpoint
	if t.serverURL != nil {
		// JoinPath creates a new URL with serverURL's scheme/host and appends the original path
		// note: JoinPath doesn't copy query parameters, so we restore them below
		newReq.URL = t.serverURL.JoinPath(newReq.URL.Path)
		newReq.URL.RawQuery = q.Encode()
		if newReq.URL.Path != "" && !strings.HasPrefix(newReq.URL.Path, "/") {
			newReq.URL.Path = "/" + newReq.URL.Path
		}
		// update Host header to match the new URL
		newReq.Host = newReq.URL.Host
	}

	// add API key as query parameter if not already present
	// Google Gemini API expects: ?key=YOUR_API_KEY
	if q.Get("key") == "" && t.apiKey != "" {
		q.Set("key", t.apiKey)
		newReq.URL.RawQuery = q.Encode()
	}

	// add Authorization header for LiteLLM passthrough compatibility
	// LiteLLM can accept either query parameter or Bearer token depending on configuration
	if t.apiKey != "" && newReq.Host != defaultGeminiHost {
		newReq.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	return t.wrapped.RoundTrip(newReq)
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
		googleai.WithEndpoint(cfg.GeminiServerURL),
		googleai.WithDefaultModel(GeminiAgentModel),
	}

	parsedURL, err := url.Parse(cfg.GeminiServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Gemini server URL: %w", err)
	}

	// always use custom transport to ensure API key injection and URL rewriting
	var baseTransport http.RoundTripper = http.DefaultTransport
	if cfg.ProxyURL != "" {
		if _, err := url.Parse(cfg.ProxyURL); err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}

		baseTransport = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(cfg.ProxyURL)
			},
		}
	}

	opts = append(opts, googleai.WithHTTPClient(&http.Client{
		Transport: &apiKeyTransport{
			wrapped:   baseTransport,
			serverURL: parsedURL,
			apiKey:    cfg.GeminiAPIKey,
		},
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

func (p *geminiProvider) GetUsage(info map[string]any) (int64, int64) {
	var inputTokens, outputTokens int64

	if value, ok := info["input_tokens"]; ok {
		switch v := value.(type) {
		case int:
			inputTokens = int64(v)
		case int32:
			inputTokens = int64(v)
		case int64:
			inputTokens = v
		case float64:
			inputTokens = int64(v)
		}
	}

	if value, ok := info["output_tokens"]; ok {
		switch v := value.(type) {
		case int:
			outputTokens = int64(v)
		case int32:
			outputTokens = int64(v)
		case int64:
			outputTokens = v
		case float64:
			outputTokens = int64(v)
		}
	}

	return inputTokens, outputTokens
}
