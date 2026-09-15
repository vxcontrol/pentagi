package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func TestSearchEnginesFormParallelSaveAndReset(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("DUCKDUCKGO_ENABLED=false\nTAVILY_API_KEY=existing-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	c := controller.NewController(s, nil, checker.CheckResult{})
	m := NewSearchEnginesFormModel(c, styles.New(), window.New())
	m.BuildForm()
	before := c.GetSearchEnginesConfig()
	if before.ParallelSearchEnabled.Default != "false" || before.ConfiguredCount != 1 {
		t.Fatalf("unexpected initial config: %+v", before)
	}
	found := false
	for i := range m.fields {
		if m.fields[i].Key == "parallel_search_enabled" {
			m.fields[i].Input.SetValue("true")
			found = true
		}
	}
	if !found {
		t.Fatal("Parallel toggle is not loaded by the actual form")
	}
	if err := m.HandleSave(); err != nil {
		t.Fatal(err)
	}
	after := c.GetSearchEnginesConfig()
	if after.ParallelSearchEnabled.Value != "true" || after.ConfiguredCount != 2 || after.TavilyAPIKey.Value != "existing-key" || after.DuckDuckGoEnabled.Value != "false" {
		t.Fatalf("saved config changed another provider or lost Parallel: %+v", after)
	}
	if !strings.Contains(m.GetCurrentConfiguration(), "Parallel Search MCP") {
		t.Fatal("Parallel is missing from the configuration summary")
	}
	if !c.GetApplyChangesConfig().HasCritical {
		t.Fatal("enabling Parallel must be classified as a restart change")
	}
	if s.GetEulaConsent() {
		t.Fatal("configuring search must not accept the EULA")
	}
	// Reset discards pending form changes through the installer's native state.
	m.HandleReset()
	reset := c.GetSearchEnginesConfig().ParallelSearchEnabled
	if reset.Value != "false" && reset.Value != "" {
		t.Fatalf("reset kept the pending Parallel opt-in: %+v", reset)
	}
	for i := range m.fields {
		if m.fields[i].Key == "parallel_search_enabled" {
			m.fields[i].Input.SetValue("true")
		}
	}
	if err := m.HandleSave(); err != nil {
		t.Fatal(err)
	}
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	reloaded, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if value, _ := reloaded.GetVar("PARALLEL_SEARCH_ENABLED"); value.Value != "true" {
		t.Fatalf("Parallel toggle was not saved to the native env file: %+v", value)
	}
}

func TestSearchEnginesFormParallelRejectsInvalidBoolean(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("PARALLEL_SEARCH_ENABLED=false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	c := controller.NewController(s, nil, checker.CheckResult{})
	m := NewSearchEnginesFormModel(c, styles.New(), window.New())
	m.BuildForm()
	for i := range m.fields {
		if m.fields[i].Key == "parallel_search_enabled" {
			m.fields[i].Input.SetValue("invalid")
		}
	}
	if err := m.HandleSave(); err == nil {
		t.Fatal("invalid boolean was accepted")
	}
	if c.GetSearchEnginesConfig().ParallelSearchEnabled.Value != "false" {
		t.Fatal("failed validation changed the saved toggle")
	}
}
