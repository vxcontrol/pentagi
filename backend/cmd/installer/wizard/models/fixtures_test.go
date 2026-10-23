package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
)

const providerFormPrefix = "llm_provider_form§"

// newTestController is a controller over an empty env file that sees result as the last check.
func newTestController(t *testing.T, result checker.CheckResult) controller.Controller {
	t.Helper()
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	return controller.NewController(st, files.NewFiles(), result)
}

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
