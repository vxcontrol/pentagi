package custom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
)

func TestConfigLoading(t *testing.T) {
	cfg := &config.Config{
		LLMServerKey:   "test-key",
		LLMServerURL:   "https://api.openai.com/v1",
		LLMServerModel: "gpt-4o-mini",
	}

	tests := []struct {
		name           string
		configPath     string
		expectError    bool
		checkRawConfig bool
	}{
		{
			name:           "config without file",
			configPath:     "",
			expectError:    false,
			checkRawConfig: true,
		},
		{
			name:        "config with invalid file path",
			configPath:  "/nonexistent/config.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCfg := *cfg
			testCfg.LLMServerConfig = tt.configPath

			providerConfig, err := DefaultProviderConfig(&testCfg)
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create provider config: %v", err)
			}

			prov, err := New(&testCfg, providerConfig)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if tt.checkRawConfig {
				rawConfig := prov.GetRawConfig()
				if len(rawConfig) == 0 {
					t.Fatal("Raw config should not be empty")
				}
			}

			providerConfig = prov.GetProviderConfig()
			if providerConfig == nil {
				t.Fatal("Provider config should not be nil")
			}

			for _, agentType := range pconfig.AllAgentTypes {
				options := providerConfig.GetOptionsForType(agentType)
				if len(options) == 0 {
					t.Errorf("Expected options for agent type %s, got none", agentType)
				}

				model := prov.Model(agentType)
				if model == "" {
					t.Errorf("Expected model for agent type %s, got empty string", agentType)
				}

				priceInfo := prov.GetPriceInfo(agentType)
				// custom provider may not have pricing info, that's acceptable
				_ = priceInfo
			}
		})
	}
}

func TestProviderType(t *testing.T) {
	cfg := &config.Config{
		LLMServerKey:   "test-key",
		LLMServerURL:   "https://api.openai.com/v1",
		LLMServerModel: "gpt-4o-mini",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	expectedType := provider.ProviderCustom
	if prov.Type() != expectedType {
		t.Errorf("Expected provider type %s, got %s", expectedType, prov.Type())
	}
}

func TestBuildProviderConfig(t *testing.T) {
	cfg := &config.Config{
		LLMServerModel: "test-model",
	}

	tests := []struct {
		name       string
		configData string
		expectErr  bool
	}{
		{
			name:       "empty config",
			configData: "{}",
			expectErr:  false,
		},
		{
			name:       "default empty config",
			configData: pconfig.EmptyProviderConfigRaw,
			expectErr:  false,
		},
		{
			name: "config with agent settings",
			configData: `{
				"simple": {
					"model": "custom-model",
					"temperature": 0.5
				}
			}`,
			expectErr: false,
		},
		{
			name:       "invalid json",
			configData: `{"simple": {"model": "test", "temperature": invalid}}`,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig, err := BuildProviderConfig(cfg, []byte(tt.configData))
			if tt.expectErr {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if providerConfig == nil {
				t.Fatal("Provider config should not be nil")
			}

			// check that default model is applied when config doesn't specify one
			options := providerConfig.GetOptionsForType(pconfig.OptionsTypeSimple)
			if len(options) == 0 {
				t.Fatal("Expected default options")
			}
		})
	}
}

// TestLoadModelsFromServer, TestLoadModelsFromServerTimeout, TestLoadModelsFromServerHeaders
// tests removed - these functions are now tested in provider/litellm_test.go with
// LoadModelsFromHTTP which provides the same functionality.

func TestProviderModelsIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"data": [
				{
					"id": "model-a",
					"description": "Basic model without special features"
				},
				{
					"id": "model-b",
					"created": 1686588896,
					"supported_parameters": ["reasoning", "max_tokens", "tools"],
					"pricing": {"prompt": "0.0001", "completion": "0.0002"}
				},
				{
					"id": "model-c",
					"created": 1686588896,
					"supported_parameters": ["reasoning", "max_tokens"],
					"pricing": {"prompt": "0.003", "completion": "0.004"}
				}
			]
		}`)
	}))
	defer server.Close()

	cfg := &config.Config{
		LLMServerKey:   "test-key",
		LLMServerURL:   server.URL,
		LLMServerModel: "model-a",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	models := prov.GetModels()
	if len(models) != 2 { // exclude model-c, it has no tools
		t.Errorf("Expected 2 models, got %d", len(models))
		return
	}

	// Verify first model with extended fields
	model1 := models[0]
	if model1.Name != "model-a" {
		t.Errorf("Expected first model name 'model-a', got '%s'", model1.Name)
	}
	if model1.Description == nil || *model1.Description != "Basic model without special features" {
		t.Error("Expected description to be set for first model")
	}

	// Verify second model with reasoning and automatic price conversion
	model2 := models[1]
	if model2.Name != "model-b" {
		t.Errorf("Expected second model name 'model-b', got '%s'", model2.Name)
	}
	if model2.Thinking == nil || !*model2.Thinking {
		t.Error("Expected second model to have reasoning capability")
	}
	if model2.ReleaseDate == nil {
		t.Error("Expected second model to have release date")
	}
	if model2.Price == nil {
		t.Error("Expected second model to have pricing")
	} else {
		// Test automatic price conversion: both prices < 0.001 triggers conversion to per-million-token
		// 0.0001 * 1000000 = 100.0
		if model2.Price.Input != 100.0 {
			t.Errorf("Expected input price 100.0 (after automatic conversion), got %f", model2.Price.Input)
		}
		// 0.0002 * 1000000 = 200.0
		if model2.Price.Output != 200.0 {
			t.Errorf("Expected output price 200.0 (after automatic conversion), got %f", model2.Price.Output)
		}
	}
}

// TestPatchProviderConfigWithProviderName test removed - config patching is no longer used.
// Prefix handling is now done at runtime via ModelWithPrefix() method.
