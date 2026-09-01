package providers

import (
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
)

// TestProviderRegistryMatchesAllProviderTypes keeps provider.AllProviderTypes (the
// API whitelist) and providerRegistry (construction wiring) in sync. Drift fails
// silently: a type in only AllProviderTypes is accepted then errors "unknown
// provider type" at construction; a type in only providerRegistry is rejected 422
// despite working. Keep the sets equal.
func TestProviderRegistryMatchesAllProviderTypes(t *testing.T) {
	registryTypes := make(map[provider.ProviderType]struct{}, len(providerRegistry))
	for _, e := range providerRegistry {
		if _, dup := registryTypes[e.Type]; dup {
			t.Errorf("duplicate providerRegistry entry for %q", e.Type)
		}
		registryTypes[e.Type] = struct{}{}
		assert.Equal(t, provider.ProviderName(e.Type), e.Name,
			"registry Name for %q must be the canonical provider name", e.Type)
	}

	allTypes := make(map[provider.ProviderType]struct{}, len(provider.AllProviderTypes))
	for _, pt := range provider.AllProviderTypes {
		allTypes[pt] = struct{}{}
	}

	for pt := range allTypes {
		_, ok := registryTypes[pt]
		assert.Truef(t, ok, "%q is in AllProviderTypes but missing from providerRegistry", pt)
	}
	for pt := range registryTypes {
		_, ok := allTypes[pt]
		assert.Truef(t, ok, "%q is in providerRegistry but missing from AllProviderTypes", pt)
	}
}

// TestEnabledDefaultProviderTypes pins that it reports exactly the types whose
// registry gate is satisfied — the same gate NewProviderController uses to
// decide what to construct — without requiring a database or Docker client.
func TestEnabledDefaultProviderTypes(t *testing.T) {
	t.Run("nothing configured", func(t *testing.T) {
		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{}))
	})

	t.Run("only credentialed types are reported", func(t *testing.T) {
		cfg := &config.Config{
			OpenAIKey:       "key",
			AnthropicAPIKey: "key",
			KimiAPIKey:      "key",
		}

		types := EnabledDefaultProviderTypes(cfg)

		assert.ElementsMatch(t, []string{
			string(provider.ProviderOpenAI),
			string(provider.ProviderAnthropic),
			string(provider.ProviderKimi),
		}, types)
	})

	t.Run("custom needs both a URL and a model or config path", func(t *testing.T) {
		assert.Empty(t, EnabledDefaultProviderTypes(&config.Config{LLMServerURL: "http://llm"}))

		types := EnabledDefaultProviderTypes(&config.Config{
			LLMServerURL:   "http://llm",
			LLMServerModel: "some-model",
		})
		assert.Equal(t, []string{string(provider.ProviderCustom)}, types)
	})
}
