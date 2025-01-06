package provider

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

type ProviderType string

func (p ProviderType) String() string {
	return string(p)
}

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderCustom    ProviderType = "custom"
	ProviderAnthropic ProviderType = "anthropic"
)

type ProviderOptionsType string

const (
	OptionsTypeAgent      ProviderOptionsType = "agent"
	OptionsTypeSimple     ProviderOptionsType = "simple"
	OptionsTypeSimpleJSON ProviderOptionsType = "simple_json"
	OptionsTypeAdviser    ProviderOptionsType = "adviser"
	OptionsTypeGenerator  ProviderOptionsType = "generator"
	OptionsTypeRefiner    ProviderOptionsType = "refiner"
	OptionsTypeSearcher   ProviderOptionsType = "searcher"
	OptionsTypeEnricher   ProviderOptionsType = "enricher"
	OptionsTypeCoder      ProviderOptionsType = "coder"
	OptionsTypeInstaller  ProviderOptionsType = "installer"
	OptionsTypePentester  ProviderOptionsType = "pentester"
	OptionsTypeReflector  ProviderOptionsType = "reflector"
)

var ErrInvalidProviderOptionsType = fmt.Errorf("provider options type not found")

type Provider interface {
	Type() ProviderType
	Model(opt ProviderOptionsType) string
	Embedder() *embeddings.EmbedderImpl
	GetUsage(info map[string]any) (int64, int64)

	Call(ctx context.Context, opt ProviderOptionsType, prompt string) (string, error)
	CallEx(
		ctx context.Context,
		opt ProviderOptionsType,
		chain []llms.MessageContent,
	) (*llms.ContentResponse, error)
	CallWithTools(
		ctx context.Context,
		opt ProviderOptionsType,
		chain []llms.MessageContent,
		tools []llms.Tool,
	) (*llms.ContentResponse, error)
}

type ProvidersList []ProviderType

type Providers map[ProviderType]Provider

func (p Providers) Get(ptype ProviderType) (Provider, error) {
	provider, ok := p[ptype]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	return provider, nil
}

func (p Providers) List() ProvidersList {
	providers := make([]ProviderType, 0, len(p))
	for provider := range p {
		providers = append(providers, provider)
	}

	return providers
}

func (p Providers) ListStrings() []string {
	providers := make([]string, 0, len(p))
	for provider := range p {
		providers = append(providers, string(provider))
	}

	return providers
}
