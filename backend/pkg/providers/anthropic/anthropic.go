package anthropic

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const AnthropicAgentModel = "claude-sonnet-4-6"

const AnthropicToolCallIDTemplate = "toolu_{r:24:b}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(AnthropicAgentModel),
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

// Configured reports whether the environment names a credential the door can
// authenticate with: an API key, or a complete federation config. The token file
// is not opened here, so a projected volume may be mounted after startup.
func Configured(cfg *config.Config) bool {
	if cfg.AnthropicAPIKey != "" {
		return true
	}

	_, _, missing := federation(cfg)
	return len(missing) == 0
}

// CredentialsNotice explains why the environment's federation variables will
// not be used or will stop working, or returns an empty string when there is
// nothing to explain.
func CredentialsNotice(cfg *config.Config) string {
	if !federationRequested(cfg) {
		return ""
	}

	if cfg.AnthropicAPIKey != "" {
		return "anthropic: ANTHROPIC_API_KEY is set, so the federation variables are ignored; unset the key to federate"
	}

	_, identity, missing := federation(cfg)
	if len(missing) > 0 {
		return "anthropic provider disabled: federation is missing " + strings.Join(missing, ", ")
	}

	return identity.notice(clock())
}

func federationRequested(cfg *config.Config) bool {
	return cfg.AnthropicFederationRuleID != "" || cfg.AnthropicOrganizationID != "" ||
		cfg.AnthropicServiceAccountID != "" || cfg.AnthropicIdentityTokenFile != "" ||
		cfg.AnthropicIdentityToken != ""
}

// federation also returns the variables the config still lacks. The token file
// wins over a literal token: it is re-read on every exchange, so it follows a
// rotating projected token.
func federation(cfg *config.Config) (federationIdentity, identityToken, []string) {
	var missing []string
	for _, required := range []struct{ value, name string }{
		{cfg.AnthropicFederationRuleID, "ANTHROPIC_FEDERATION_RULE_ID"},
		{cfg.AnthropicOrganizationID, "ANTHROPIC_ORGANIZATION_ID"},
		{cfg.AnthropicServiceAccountID, "ANTHROPIC_SERVICE_ACCOUNT_ID"},
	} {
		if required.value == "" {
			missing = append(missing, required.name)
		}
	}

	var identity identityToken
	switch {
	case cfg.AnthropicIdentityTokenFile != "":
		identity = identityToken{file: cfg.AnthropicIdentityTokenFile}
	case strings.TrimSpace(cfg.AnthropicIdentityToken) != "":
		identity = literalIdentityToken(cfg.AnthropicIdentityToken)
	default:
		missing = append(missing, "ANTHROPIC_IDENTITY_TOKEN_FILE or ANTHROPIC_IDENTITY_TOKEN")
	}

	baseURL := strings.TrimSuffix(cfg.AnthropicServerURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return federationIdentity{
		ruleID:           cfg.AnthropicFederationRuleID,
		organizationID:   cfg.AnthropicOrganizationID,
		serviceAccountID: cfg.AnthropicServiceAccountID,
		workspaceID:      cfg.AnthropicWorkspaceID,
		baseURL:          baseURL,
	}, identity, missing
}

// Only a non-empty key shadows federation, here and in Configured and
// CredentialsNotice: docker-compose exports ANTHROPIC_API_KEY even when unset.
func credentials(cfg *config.Config, client *http.Client) ([]anthropic.Option, error) {
	if cfg.AnthropicAPIKey != "" {
		return []anthropic.Option{anthropic.WithToken(cfg.AnthropicAPIKey), anthropic.WithHTTPClient(client)}, nil
	}

	id, identity, missing := federation(cfg)
	if len(missing) > 0 {
		return nil, fmt.Errorf("anthropic: set ANTHROPIC_API_KEY or complete the federation config (missing %s)",
			strings.Join(missing, ", "))
	}

	// Not the library's WithFederation: it caches per client, so every door would present the identity
	// token again, and waits on the exchange deaf to the caller's context. The public WithToken and
	// WithHTTPClient suffice to put the shared bearer on every request; the placeholder key never leaves.
	return []anthropic.Option{
		anthropic.WithToken("federated"),
		anthropic.WithHTTPClient(federatedDoer{client: client, source: sharedSource(id, identity)}),
	}, nil
}

type anthropicProvider struct {
	llm            *anthropic.LLM
	models         pconfig.ModelsConfig
	providerName   provider.ProviderName
	providerConfig *pconfig.ProviderConfig
}

func New(
	cfg *config.Config,
	providerName provider.ProviderName,
	providerConfig *pconfig.ProviderConfig,
) (provider.Provider, error) {
	baseURL := cfg.AnthropicServerURL
	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	options, err := credentials(cfg, httpClient)
	if err != nil {
		return nil, err
	}

	client, err := anthropic.New(append(options,
		anthropic.WithModel(AnthropicAgentModel),
		anthropic.WithBaseURL(baseURL),
		// Enable prompt caching for cost optimization (90% savings on cached reads)
		anthropic.WithDefaultCacheStrategy(anthropic.CacheStrategy{
			CacheTools:    true,
			CacheSystem:   true,
			CacheMessages: true,
			TTL:           "5m",
		}),
	)...)
	if err != nil {
		return nil, err
	}

	return &anthropicProvider{
		llm:            client,
		models:         models,
		providerName:   providerName,
		providerConfig: providerConfig,
	}, nil
}

func (p *anthropicProvider) Type() provider.ProviderType {
	return provider.ProviderAnthropic
}

func (p *anthropicProvider) Name() provider.ProviderName {
	return p.providerName
}

func (p *anthropicProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *anthropicProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *anthropicProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *anthropicProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *anthropicProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := AnthropicAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *anthropicProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	// Anthropic provider doesn't need prefix support (passthrough mode in LiteLLM)
	return p.Model(opt)
}

func (p *anthropicProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *anthropicProvider) CallEx(
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

func (p *anthropicProvider) CallWithTools(
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

func (p *anthropicProvider) CallWithExtraOptions(
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

func (p *anthropicProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *anthropicProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, AnthropicToolCallIDTemplate)
}
