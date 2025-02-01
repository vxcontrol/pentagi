package providers

import (
	"context"
	"fmt"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"
)

type providerController struct {
	docker   docker.DockerClient
	publicIP string
	provider.Providers
}

type ProviderController interface {
	NewFlowProvider(
		ctx context.Context,
		db database.Querier,
		prvtype provider.ProviderType,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		flowID int64,
		input string,
	) (FlowProvider, error)
	LoadFlowProvider(
		db database.Querier,
		prvtype provider.ProviderType,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		flowID int64,
		image, language, title string,
	) (FlowProvider, error)
	Get(ptype provider.ProviderType) (provider.Provider, error)
	List() provider.ProvidersList
	ListStrings() []string
}

func NewProviderController(cfg *config.Config, docker docker.DockerClient) (ProviderController, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	providers := make(provider.Providers)

	if cfg.OpenAIKey != "" {
		provider, err := openai.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create openai provider: %w", err)
		}

		providers[provider.Type()] = provider
	}

	if cfg.LLMServerURL != "" && cfg.LLMServerModel != "" {
		provider, err := custom.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom provider: %w", err)
		}

		providers[provider.Type()] = provider
	}

	if cfg.AnthropicAPIKey != "" {
		provider, err := anthropic.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create anthropic provider: %w", err)
		}

		providers[provider.Type()] = provider
	}

	return &providerController{
		docker:    docker,
		publicIP:  cfg.DockerPublicIP,
		Providers: providers,
	}, nil
}

func (pc *providerController) NewFlowProvider(
	ctx context.Context,
	db database.Querier,
	prvtype provider.ProviderType,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	flowID int64,
	input string,
) (FlowProvider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.NewFlowProvider")
	defer span.End()

	prv, err := pc.Get(prvtype)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	imageTmpl, err := prompter.RenderTemplate(templates.PromptTypeImageChooser, map[string]any{
		"DefaultImage": pc.docker.GetDefaultImage(),
		"Input":        input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get primary docker image template: %w", err)
	}

	image, err := prv.Call(ctx, provider.OptionsTypeSimple, imageTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary docker image: %w", err)
	}
	image = strings.ToLower(strings.TrimSpace(image))

	languageTmpl, err := prompter.RenderTemplate(templates.PromptTypeLanguageChooser, map[string]any{
		"Input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get language template: %w", err)
	}

	language, err := prv.Call(ctx, provider.OptionsTypeSimple, languageTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}
	language = strings.TrimSpace(language)

	titleTmpl, err := prompter.RenderTemplate(templates.PromptTypeFlowDescriptor, map[string]any{
		"Input": input,
		"Lang":  language,
		"N":     15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title template: %w", err)
	}

	title, err := prv.Call(ctx, provider.OptionsTypeSimple, titleTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title: %w", err)
	}
	title = strings.TrimSpace(title)

	fp := &flowProvider{
		db:       db,
		flowID:   flowID,
		publicIP: pc.publicIP,
		image:    image,
		title:    title,
		language: language,
		prompter: prompter,
		executor: executor,
		Provider: prv,
	}

	return fp, nil
}

func (pc *providerController) LoadFlowProvider(
	db database.Querier,
	prvtype provider.ProviderType,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	flowID int64,
	image, language, title string,
) (FlowProvider, error) {
	prv, err := pc.Get(prvtype)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	fp := &flowProvider{
		db:       db,
		flowID:   flowID,
		publicIP: pc.publicIP,
		image:    image,
		title:    title,
		language: language,
		prompter: prompter,
		executor: executor,
		Provider: prv,
	}

	return fp, nil
}
