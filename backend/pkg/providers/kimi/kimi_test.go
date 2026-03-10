package kimi

import (
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

func TestConfigLoading(t *testing.T) {
	cfg := &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	rawConfig := prov.GetRawConfig()
	if len(rawConfig) == 0 {
		t.Fatal("Raw config should not be empty")
	}

	providerConfig = prov.GetProviderConfig()
	if providerConfig == nil {
		t.Fatal("Provider config should not be nil")
	}

	for _, agentType := range pconfig.AllAgentTypes {
		model := prov.Model(agentType)
		if model == "" {
			t.Errorf("Agent type %v should have a model assigned", agentType)
		}
	}

	for _, agentType := range pconfig.AllAgentTypes {
		priceInfo := prov.GetPriceInfo(agentType)
		if priceInfo == nil {
			t.Errorf("Agent type %v should have price information", agentType)
		} else {
			if priceInfo.Input <= 0 || priceInfo.Output <= 0 {
				t.Errorf("Agent type %v should have positive input (%f) and output (%f) prices",
					agentType, priceInfo.Input, priceInfo.Output)
			}
		}
	}
}

func TestProviderType(t *testing.T) {
	cfg := &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderKimi {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderKimi, prov.Type())
	}
}

func TestModelsLoading(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Models list should not be empty")
	}

	for _, model := range models {
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}

		if model.Price == nil {
			t.Errorf("Model %s should have price information", model.Name)
			continue
		}

		if model.Price.Input <= 0 {
			t.Errorf("Model %s should have positive input price", model.Name)
		}

		if model.Price.Output <= 0 {
			t.Errorf("Model %s should have positive output price", model.Name)
		}
	}
}

func TestModelWithPrefix(t *testing.T) {
	cfg := &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
		KimiProvider:   "moonshot",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, agentType := range pconfig.AllAgentTypes {
		modelWithPrefix := prov.ModelWithPrefix(agentType)
		model := prov.Model(agentType)

		expected := "moonshot/" + model
		if modelWithPrefix != expected {
			t.Errorf("Agent type %v: expected prefixed model %q, got %q", agentType, expected, modelWithPrefix)
		}
	}
}

func TestModelWithoutPrefix(t *testing.T) {
	cfg := &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	for _, agentType := range pconfig.AllAgentTypes {
		modelWithPrefix := prov.ModelWithPrefix(agentType)
		model := prov.Model(agentType)

		if modelWithPrefix != model {
			t.Errorf("Agent type %v: without prefix, ModelWithPrefix (%q) should equal Model (%q)",
				agentType, modelWithPrefix, model)
		}
	}
}

func TestMissingAPIKey(t *testing.T) {
	cfg := &config.Config{
		KimiServerURL: "https://api.moonshot.ai/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(cfg, providerConfig)
	if err == nil {
		t.Fatal("Expected error when API key is missing")
	}
}

func TestGetUsage(t *testing.T) {
	cfg := &config.Config{
		KimiAPIKey:    "test-key",
		KimiServerURL: "https://api.moonshot.ai/v1",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	usage := prov.GetUsage(map[string]any{
		"PromptTokens":     100,
		"CompletionTokens": 50,
	})
	if usage.Input != 100 || usage.Output != 50 {
		t.Errorf("Expected usage input=100 output=50, got input=%d output=%d", usage.Input, usage.Output)
	}
}
