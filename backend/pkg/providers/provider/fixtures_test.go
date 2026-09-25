package provider

import (
	"testing"

	"pentagi/pkg/providers/pconfig"

	"github.com/vxcontrol/langchaingo/llms"
)

// catalogProvider embeds a nil Provider so an unstubbed method panics instead of answering zero values.
type catalogProvider struct {
	Provider

	model  string
	prefix string
	models pconfig.ModelsConfig
	config *pconfig.ProviderConfig
}

func (c *catalogProvider) Type() ProviderType                                 { return ProviderType("test") }
func (c *catalogProvider) Model(pconfig.ProviderOptionsType) string           { return c.model }
func (c *catalogProvider) ModelWithPrefix(pconfig.ProviderOptionsType) string { return c.prefix }
func (c *catalogProvider) GetModels() pconfig.ModelsConfig                    { return c.models }
func (c *catalogProvider) GetProviderConfig() *pconfig.ProviderConfig         { return c.config }
func (c *catalogProvider) GetPriceInfo(pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return nil
}
func (c *catalogProvider) GetUsage(map[string]any) pconfig.CallUsage { return pconfig.CallUsage{} }

func limit(value int) *int { return &value }

func catalogue() pconfig.ModelsConfig {
	return pconfig.ModelsConfig{
		{Name: "glm-5", ContextWindow: limit(204800), MaxOutputTokens: limit(131072)},
		{Name: "gpt-4o", ContextWindow: limit(128000), MaxOutputTokens: limit(16384)},
		{Name: "no-limits"},
	}
}

func appliedMaxTokens(t *testing.T, options []llms.CallOption) *int {
	t.Helper()

	var applied llms.CallOptions
	for _, option := range options {
		option(&applied)
	}

	return applied.MaxTokens
}
