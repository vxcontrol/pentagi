package providers

import (
	"testing"

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
