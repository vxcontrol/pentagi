package tester

import (
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/tester/cases"
)

func TestConfig_ApplyOptions_StartsFromTheDefaultsAndAppliesTheOptions(t *testing.T) {
	config := applyOptions(nil)
	if config == nil {
		t.Fatalf("Expected non-nil config")
	}

	if len(config.agentTypes) == 0 {
		t.Errorf("Expected default agent types")
	}
	if len(config.groups) == 0 {
		t.Errorf("Expected default groups")
	}
	if !config.streamingMode {
		t.Errorf("Expected streaming mode enabled by default")
	}
	if config.verbose {
		t.Errorf("Expected verbose mode disabled by default")
	}
	if config.parallelWorkers != 4 {
		t.Errorf("Expected 4 parallel workers by default, got %d", config.parallelWorkers)
	}

	config = applyOptions([]TestOption{
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(cases.TestGroupBasic),
		WithStreamingMode(false),
		WithVerbose(true),
		WithParallelWorkers(8),
	})

	if len(config.agentTypes) != 1 || config.agentTypes[0] != pconfig.OptionsTypeSimple {
		t.Errorf("Agent types not applied correctly")
	}
	if len(config.groups) != 1 || config.groups[0] != cases.TestGroupBasic {
		t.Errorf("Groups not applied correctly")
	}
	if config.streamingMode {
		t.Errorf("Streaming mode not disabled")
	}
	if !config.verbose {
		t.Errorf("Verbose mode not enabled")
	}
	if config.parallelWorkers != 8 {
		t.Errorf("Parallel workers not set correctly")
	}
}

func TestConfig_DefaultConfig_RunsEveryGroupTheCatalogueShips(t *testing.T) {
	registry, err := cases.LoadBuiltinRegistry()
	if err != nil {
		t.Fatalf("builtin registry: %v", err)
	}

	byDefault := make(map[cases.TestGroup]bool)
	for _, group := range defaultConfig().groups {
		byDefault[group] = true
	}

	for _, group := range []cases.TestGroup{
		cases.TestGroupBasic,
		cases.TestGroupAdvanced,
		cases.TestGroupJSON,
		cases.TestGroupKnowledge,
	} {
		shipped := len(registry.GetTestsByGroup(group))
		if shipped > 0 && !byDefault[group] {
			t.Errorf("the catalogue ships %d cases in group %q, and a run that names no group executes none of them",
				shipped, group)
		}
	}
}
