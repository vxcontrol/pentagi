package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type ProviderType string

func (p ProviderType) String() string {
	return string(p)
}

// ReasoningProvider maps the provider type to the langchaingo reasoning.Provider
// consumed ONLY by capability introspection for the settings UI (CannotDisable /
// Supported hints via llms.ReasoningSupportFor / reasoning.ResolveOff) and by the
// save-time check against what the door refuses — it has no effect on the actual
// wire call, which each provider builds independently.
//
// Every OpenAI-compatible door, Custom included, maps to reasoning.ProviderOpenAI,
// so the hints and the save-time check follow the model name, not whatever backend
// Custom fronts behind LLM_SERVER_URL.
func (p ProviderType) ReasoningProvider() reasoning.Provider {
	switch p {
	case ProviderAnthropic:
		return reasoning.ProviderAnthropic
	case ProviderBedrock:
		return reasoning.ProviderBedrock
	case ProviderGemini:
		return reasoning.ProviderGoogleAI
	case ProviderOpenAI, ProviderDeepSeek, ProviderGLM, ProviderKimi, ProviderQwen, ProviderMiniMax,
		ProviderMistral, ProviderXAI, ProviderCustom:
		return reasoning.ProviderOpenAI
	case ProviderOllama:
		return reasoning.ProviderOllama
	default:
		return reasoning.ProviderUnknown
	}
}

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGemini    ProviderType = "gemini"
	ProviderBedrock   ProviderType = "bedrock"
	ProviderOllama    ProviderType = "ollama"
	ProviderCustom    ProviderType = "custom"
	ProviderDeepSeek  ProviderType = "deepseek"
	ProviderGLM       ProviderType = "glm"
	ProviderKimi      ProviderType = "kimi"
	ProviderQwen      ProviderType = "qwen"
	ProviderMiniMax   ProviderType = "minimax"
	ProviderMistral   ProviderType = "mistral"
	ProviderXAI       ProviderType = "xai"
)

// AllProviderTypes enumerates every supported provider type; keep it in sync with
// the consts above. The API-layer type whitelist validates against it.
var AllProviderTypes = ProvidersListTypes{
	ProviderOpenAI,
	ProviderAnthropic,
	ProviderGemini,
	ProviderBedrock,
	ProviderOllama,
	ProviderCustom,
	ProviderDeepSeek,
	ProviderGLM,
	ProviderKimi,
	ProviderQwen,
	ProviderMiniMax,
	ProviderMistral,
	ProviderXAI,
}

type ProviderName string

func (p ProviderName) String() string {
	return string(p)
}

const (
	DefaultProviderNameOpenAI    ProviderName = ProviderName(ProviderOpenAI)
	DefaultProviderNameAnthropic ProviderName = ProviderName(ProviderAnthropic)
	DefaultProviderNameGemini    ProviderName = ProviderName(ProviderGemini)
	DefaultProviderNameBedrock   ProviderName = ProviderName(ProviderBedrock)
	DefaultProviderNameOllama    ProviderName = ProviderName(ProviderOllama)
	DefaultProviderNameCustom    ProviderName = ProviderName(ProviderCustom)
	DefaultProviderNameDeepSeek  ProviderName = ProviderName(ProviderDeepSeek)
	DefaultProviderNameGLM       ProviderName = ProviderName(ProviderGLM)
	DefaultProviderNameKimi      ProviderName = ProviderName(ProviderKimi)
	DefaultProviderNameQwen      ProviderName = ProviderName(ProviderQwen)
	DefaultProviderNameMiniMax   ProviderName = ProviderName(ProviderMiniMax)
	DefaultProviderNameMistral   ProviderName = ProviderName(ProviderMistral)
	DefaultProviderNameXAI       ProviderName = ProviderName(ProviderXAI)
)

type Provider interface {
	Type() ProviderType
	Name() ProviderName
	Model(opt pconfig.ProviderOptionsType) string
	// ModelWithPrefix returns model name WITH provider prefix for LLM API calls and Langfuse logging
	ModelWithPrefix(opt pconfig.ProviderOptionsType) string
	GetUsage(info map[string]any) pconfig.CallUsage

	Call(ctx context.Context, opt pconfig.ProviderOptionsType, prompt string) (string, error)
	CallEx(
		ctx context.Context,
		opt pconfig.ProviderOptionsType,
		chain []llms.MessageContent,
		streamCb streaming.Callback,
	) (*llms.ContentResponse, error)
	CallWithTools(
		ctx context.Context,
		opt pconfig.ProviderOptionsType,
		chain []llms.MessageContent,
		tools []llms.Tool,
		streamCb streaming.Callback,
	) (*llms.ContentResponse, error)
	// A reasoning option in extra cannot replace the thinking of an agent
	// that uses adaptive thinking.
	CallWithExtraOptions(
		ctx context.Context,
		opt pconfig.ProviderOptionsType,
		chain []llms.MessageContent,
		tools []llms.Tool,
		streamCb streaming.Callback,
		extra ...llms.CallOption,
	) (*llms.ContentResponse, error)

	// Configuration access methods
	GetRawConfig() []byte
	GetProviderConfig() *pconfig.ProviderConfig

	// Pricing information methods
	GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo

	// Models information methods
	GetModels() pconfig.ModelsConfig

	// GetToolCallIDTemplate returns the pattern template for tool call IDs
	// This method is cached per provider instance using sync.Once
	GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error)
}

type (
	ProvidersListNames []ProviderName
	ProvidersListTypes []ProviderType
	Providers          map[ProviderName]Provider
	ProvidersConfig    map[ProviderType]*pconfig.ProviderConfig
)

func (pln ProvidersListNames) Contains(pname ProviderName) bool {
	for _, item := range pln {
		if item == pname {
			return true
		}
	}
	return false
}

func (plt ProvidersListTypes) Contains(ptype ProviderType) bool {
	for _, item := range plt {
		if item == ptype {
			return true
		}
	}
	return false
}

func (p Providers) Get(pname ProviderName) (Provider, error) {
	provider, ok := p[pname]
	if !ok {
		return nil, fmt.Errorf("provider not found by name '%s'", pname)
	}

	return provider, nil
}

func (p Providers) ListNames() ProvidersListNames {
	listNames := make([]ProviderName, 0, len(p))
	for pname := range p {
		listNames = append(listNames, pname)
	}

	sort.Slice(listNames, func(i, j int) bool {
		return strings.Compare(string(listNames[i]), string(listNames[j])) > 0
	})

	return listNames
}

func (p Providers) ListTypes() ProvidersListTypes {
	mapTypes := make(map[ProviderType]struct{})
	for _, provider := range p {
		mapTypes[provider.Type()] = struct{}{}
	}

	listTypes := make([]ProviderType, 0, len(mapTypes))
	for ptype := range mapTypes {
		listTypes = append(listTypes, ptype)
	}
	sort.Slice(listTypes, func(i, j int) bool {
		return strings.Compare(string(listTypes[i]), string(listTypes[j])) > 0
	})

	return listTypes
}
