package gemini

import (
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

func TestConfigLoading(t *testing.T) {
	cfg := &config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
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
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderGemini {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderGemini, prov.Type())
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

		if model.Price.Input != 0 || model.Price.Output != 0 { // exclude totally free models
			if model.Price.Input <= 0 {
				t.Errorf("Model %s should have positive input price", model.Name)
			}

			if model.Price.Output <= 0 {
				t.Errorf("Model %s should have positive output price", model.Name)
			}
		}
	}
}

func TestGeminiSpecificFeatures(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	// Test that we have current Gemini models
	expectedModels := []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.0-flash"}
	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range models {
			if model.Name == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found in models list", expectedModel)
		}
	}

	// Test default agent model
	if GeminiAgentModel != "gemini-2.5-flash" {
		t.Errorf("Expected default agent model to be gemini-2.5-flash, got %s", GeminiAgentModel)
	}
}

func TestGetUsage(t *testing.T) {
	cfg := &config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Test usage parsing with Google AI format
	usageInfo := map[string]any{
		"input_tokens":  int32(100),
		"output_tokens": int32(50),
	}

	inputTokens, outputTokens := prov.GetUsage(usageInfo)
	if inputTokens != 100 {
		t.Errorf("Expected input tokens 100, got %d", inputTokens)
	}
	if outputTokens != 50 {
		t.Errorf("Expected output tokens 50, got %d", outputTokens)
	}

	// Test with missing usage info
	emptyInfo := map[string]any{}
	inputTokens, outputTokens = prov.GetUsage(emptyInfo)
	if inputTokens != 0 || outputTokens != 0 {
		t.Errorf("Expected zero tokens with empty usage info, got input: %d, output: %d", inputTokens, outputTokens)
	}
}
