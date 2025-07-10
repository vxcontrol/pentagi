package custom

import (
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
				// Custom provider may not have pricing info, that's acceptable
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

			// Check that default model is applied when config doesn't specify one
			options := providerConfig.GetOptionsForType(pconfig.OptionsTypeSimple)
			if len(options) == 0 {
				t.Fatal("Expected default options")
			}
		})
	}
}
