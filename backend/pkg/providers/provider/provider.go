package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pentagi/pkg/providers/pconfig"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type ProviderType string

func (p ProviderType) String() string {
	return string(p)
}

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGemini    ProviderType = "gemini"
	ProviderBedrock   ProviderType = "bedrock"
	ProviderOllama    ProviderType = "ollama"
	ProviderCustom    ProviderType = "custom"
)

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
)

type Provider interface {
	Type() ProviderType
	Model(opt pconfig.ProviderOptionsType) string
	GetUsage(info map[string]any) (int64, int64)

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

	// Configuration access methods
	GetRawConfig() []byte
	GetProviderConfig() *pconfig.ProviderConfig

	// Pricing information methods
	GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo

	// Models information methods
	GetModels() pconfig.ModelsConfig
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
