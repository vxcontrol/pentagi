package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestProviders_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	provider := Provider{UserID: 1, Type: "openai", Name: "my-openai", Config: json.RawMessage(`{"api_key":"test"}`)}
	info := ProviderInfo{
		Name:         "my-provider",
		Type:         "openai",
		DefaultModel: "gpt-4o",
		Models:       []ModelInfo{{Name: "gpt-4o", AgentTypes: []string{"primary_agent"}}},
	}
	name := "updated"

	modelsCheckValid(t, []modelsValidCase{
		{"a supported provider type", ProviderType("openai"), ""},
		{"an empty provider type", ProviderType(""), "invalid ProviderType: "},
		{"an unsupported provider type", ProviderType("azure"), "invalid ProviderType: azure"},

		{"a complete provider", provider, ""},
		{"a provider without a name", modelsWith(provider, func(p *Provider) { p.Name = "" }), "Provider.Name:required"},
		{"a provider without a config", modelsWith(provider, func(p *Provider) { p.Config = nil }), "Provider.Config:required"},
		{"a provider of an unknown type", modelsWith(provider, func(p *Provider) { p.Type = "invalid" }), "Provider.Type:valid"},

		{"a provider request", CreateProvider{Config: json.RawMessage(`{"key":"val"}`)}, ""},
		{"a provider request without a config", CreateProvider{}, "CreateProvider.Config:required"},

		{"a provider rename", PatchProvider{Name: &name}, ""},
		{"an empty provider edit", PatchProvider{}, ""},

		{"complete provider info", info, ""},
		{"provider info without a name", modelsWith(info, func(p *ProviderInfo) { p.Name = "" }), "ProviderInfo.Name:required"},
		{"provider info of an unknown type", modelsWith(info, func(p *ProviderInfo) { p.Type = "invalid" }),
			"ProviderInfo.Type:valid"},
		{"provider info without a default model", modelsWith(info, func(p *ProviderInfo) { p.DefaultModel = "" }),
			"ProviderInfo.DefaultModel:required"},
		{"provider info without models", modelsWith(info, func(p *ProviderInfo) { p.Models = nil }), "ProviderInfo.Models:required"},
	})
}

func TestProviders_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "openai", ProviderType("openai").String())
	assert.Equal(t, "anthropic", ProviderType("anthropic").String())
}

func TestProviders_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "providers", (&Provider{}).TableName())
}
