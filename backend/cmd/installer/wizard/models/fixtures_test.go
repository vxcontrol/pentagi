package models

import (
	"strings"
	"testing"
)

const providerFormPrefix = "llm_provider_form§"

func contains(t *testing.T, document, want, why string) {
	t.Helper()
	if !strings.Contains(document, want) {
		t.Errorf("%s\nexpected to find %q in:\n%s", why, want, document)
	}
}

func absent(t *testing.T, document, unwanted, why string) {
	t.Helper()
	if strings.Contains(document, unwanted) {
		t.Errorf("%s\nexpected NOT to find %q in:\n%s", why, unwanted, document)
	}
}

func offeredProviders(t *testing.T) []LLMProviderID {
	t.Helper()

	var offered []LLMProviderID
	for _, item := range (&LLMProvidersHandler{}).LoadItems() {
		id, ok := strings.CutPrefix(string(item.ID), providerFormPrefix)
		if !ok {
			t.Fatalf("menu item %q is not a provider form", item.ID)
		}
		offered = append(offered, LLMProviderID(id))
	}

	return offered
}
