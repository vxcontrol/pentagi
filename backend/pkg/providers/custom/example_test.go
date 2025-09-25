package custom

import (
	"os"
	"path/filepath"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
)

func TestCustomProviderUsageModes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "custom_provider_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")
	configContent := `{
		"simple": {
			"model": "gpt-3.5-turbo",
			"temperature": 0.2,
			"max_tokens": 2000
		},
		"agent": {
			"model": "gpt-4",
			"temperature": 0.8
		}
	}`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	tests := []struct {
		name               string
		setupConfig        func() *config.Config
		expectError        bool
		expectedModel      string
		expectConfigLoaded bool
	}{
		{
			name: "mode 1: config from environment variables only",
			setupConfig: func() *config.Config {
				return &config.Config{
					LLMServerKey:   "test-key",
					LLMServerURL:   "https://api.openai.com/v1",
					LLMServerModel: "gpt-4o-mini",
				}
			},
			expectError:        false,
			expectedModel:      "gpt-4o-mini",
			expectConfigLoaded: true,
		},
		{
			name: "mode 2: config from file overrides environment",
			setupConfig: func() *config.Config {
				return &config.Config{
					LLMServerKey:    "test-key",
					LLMServerURL:    "https://api.openai.com/v1",
					LLMServerModel:  "gpt-4o-mini",
					LLMServerConfig: configFile,
				}
			},
			expectError:        false,
			expectedModel:      "gpt-3.5-turbo",
			expectConfigLoaded: true,
		},
		{
			name: "mode 3: minimal config without model",
			setupConfig: func() *config.Config {
				return &config.Config{
					LLMServerKey: "test-key",
					LLMServerURL: "https://api.openai.com/v1",
				}
			},
			expectError:        false,
			expectedModel:      "",
			expectConfigLoaded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			providerConfig, err := DefaultProviderConfig(cfg)
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create provider config: %v", err)
			}

			prov, err := New(cfg, providerConfig)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if tt.expectConfigLoaded {
				rawConfig := prov.GetRawConfig()
				if len(rawConfig) == 0 {
					t.Error("Expected raw config to be loaded")
				}

				providerCfg := prov.GetProviderConfig()
				if providerCfg == nil {
					t.Fatal("Expected provider config to be available")
				}

				for _, agentType := range []pconfig.ProviderOptionsType{
					pconfig.OptionsTypeSimple,
					pconfig.OptionsTypePrimaryAgent,
				} {
					options := providerCfg.GetOptionsForType(agentType)
					if len(options) == 0 {
						t.Errorf("Expected options for agent type %s", agentType)
					}

					model := prov.Model(agentType)
					if tt.expectedModel != "" && model != tt.expectedModel {
						// For simple type, check if it matches expected model
						if agentType == pconfig.OptionsTypeSimple {
							if model != tt.expectedModel {
								t.Errorf("Expected model %s for simple type, got %s", tt.expectedModel, model)
							}
						}
					}
				}
			}
		})
	}
}

func TestCustomProviderConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		description string
	}{
		{
			name: "valid minimal config",
			config: &config.Config{
				LLMServerKey: "test-key",
				LLMServerURL: "https://api.openai.com/v1",
			},
			expectError: false,
			description: "Should work with minimal required fields",
		},
		{
			name: "config with all fields",
			config: &config.Config{
				LLMServerKey:             "test-key",
				LLMServerURL:             "https://api.openai.com/v1",
				LLMServerModel:           "gpt-4",
				LLMServerLegacyReasoning: false,
				ProxyURL:                 "http://proxy:8080",
			},
			expectError: false,
			description: "Should work with all optional fields set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerConfig, err := DefaultProviderConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for %s but got none", tt.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", tt.description, err)
			}

			prov, err := New(tt.config, providerConfig)
			if err != nil {
				t.Fatalf("Failed to create provider for %s: %v", tt.description, err)
			}

			if prov.Type() != "custom" {
				t.Errorf("Expected provider type 'custom', got %s", prov.Type())
			}
		})
	}
}
