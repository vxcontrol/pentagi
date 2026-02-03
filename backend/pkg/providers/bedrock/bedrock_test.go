package bedrock

import (
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func TestConfigLoading(t *testing.T) {
	cfg := &config.Config{
		BedrockRegion:    "us-east-1",
		BedrockAccessKey: "test-key",
		BedrockSecretKey: "test-key",
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
		BedrockRegion:    "us-east-1",
		BedrockAccessKey: "test-key",
		BedrockSecretKey: "test-key",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderBedrock {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderBedrock, prov.Type())
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

func TestBedrockSpecificFeatures(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	// Test that we have current Bedrock models
	expectedModels := []string{
		"us.anthropic.claude-sonnet-4-20250514-v1:0",
		"us.anthropic.claude-3-5-haiku-20241022-v1:0",
		"us.amazon.nova-premier-v1:0",
		"us.amazon.nova-pro-v1:0",
		"us.amazon.nova-lite-v1:0",
	}
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
	if BedrockAgentModel != bedrock.ModelAnthropicClaudeSonnet4 {
		t.Errorf("Expected default agent model to be %s, got %s",
			bedrock.ModelAnthropicClaudeSonnet4, BedrockAgentModel)
	}
}

func TestGetUsage(t *testing.T) {
	cfg := &config.Config{
		BedrockRegion:    "us-east-1",
		BedrockAccessKey: "test-key",
		BedrockSecretKey: "test-key",
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
		"PromptTokens":     int32(100),
		"CompletionTokens": int32(50),
	}

	usage := prov.GetUsage(usageInfo)
	if usage.Input != 100 {
		t.Errorf("Expected input tokens 100, got %d", usage.Input)
	}
	if usage.Output != 50 {
		t.Errorf("Expected output tokens 50, got %d", usage.Output)
	}

	// Test with missing usage info
	emptyInfo := map[string]any{}
	usage = prov.GetUsage(emptyInfo)
	if !usage.IsZero() {
		t.Errorf("Expected zero tokens with empty usage info, got %s", usage.String())
	}
}
