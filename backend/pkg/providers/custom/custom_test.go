package custom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestLoadModelsFromServer(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantModels int
		validate   func(*testing.T, pconfig.ModelsConfig)
	}{
		{
			name:       "basic models response",
			statusCode: http.StatusOK,
			response: `{
				"data": [
					{"id": "model-1"},
					{"id": "model-2"}
				]
			}`,
			wantModels: 2,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				// Both models use ID as name
				if models[0].Name != "model-1" {
					t.Errorf("Expected name 'model-1', got '%s'", models[0].Name)
				}
				if models[1].Name != "model-2" {
					t.Errorf("Expected name 'model-2', got '%s'", models[1].Name)
				}
			},
		},
		{
			name:       "models with all extended fields",
			statusCode: http.StatusOK,
			response: `{
				"data": [
					{
						"id": "anthropic/claude-3.7-sonnet",
						"created": 1740422110,
						"description": "Claude 3.7 Sonnet is an advanced large language model with improved reasoning capabilities.",
						"supported_parameters": ["max_tokens", "temperature", "reasoning", "tools"],
						"pricing": {
							"prompt": "0.000003",
							"completion": "0.000015"
						}
					}
				]
			}`,
			wantModels: 1,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				model := models[0]

				// Check ID is used as name
				if model.Name != "anthropic/claude-3.7-sonnet" {
					t.Errorf("Expected name 'anthropic/claude-3.7-sonnet', got '%s'", model.Name)
				}

				// Check description
				if model.Description == nil {
					t.Error("Expected description to be set")
				} else if *model.Description != "Claude 3.7 Sonnet is an advanced large language model with improved reasoning capabilities." {
					t.Errorf("Unexpected description: %s", *model.Description)
				}

				// Check release date from timestamp
				if model.ReleaseDate == nil {
					t.Error("Expected release date to be set")
				} else {
					expected := time.Unix(1740422110, 0).UTC()
					if !model.ReleaseDate.Equal(expected) {
						t.Errorf("Expected release date %v, got %v", expected, *model.ReleaseDate)
					}
				}

				// Check thinking capability from supported_parameters
				if model.Thinking == nil {
					t.Error("Expected thinking to be set")
				} else if !*model.Thinking {
					t.Error("Expected thinking to be true")
				}

				// Check pricing conversion
				if model.Price == nil {
					t.Error("Expected pricing to be set")
				} else {
					// 0.000003 * 1000000 = 3.0
					if model.Price.Input != 3.0 {
						t.Errorf("Expected input price 3.0, got %f", model.Price.Input)
					}
					// 0.000015 * 1000000 = 15.0
					if model.Price.Output != 15.0 {
						t.Errorf("Expected output price 15.0, got %f", model.Price.Output)
					}
				}
			},
		},
		{
			name:       "models with partial fields and edge cases",
			statusCode: http.StatusOK,
			response: `{
				"data": [
					{
						"id": "no-reasoning-model",
						"description": "Model without reasoning support",
						"supported_parameters": ["max_tokens", "temperature"]
					},
					{
						"id": "invalid-timestamp-model", 
						"created": 0,
						"description": ""
					},
					{
						"id": "pricing-only-model",
						"pricing": {
							"prompt": "2.5",
							"completion": "10.0"
						}
					}
				]
			}`,
			wantModels: 3,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				// First model: no reasoning, has description
				model1 := models[0]
				if model1.Thinking != nil && *model1.Thinking {
					t.Error("Expected thinking to be false for non-reasoning model")
				}
				if model1.Description == nil || *model1.Description != "Model without reasoning support" {
					t.Error("Expected description to be set correctly")
				}

				// Second model: invalid timestamp, empty description
				model2 := models[1]
				if model2.Thinking != nil {
					t.Error("Expected thinking to be nil for model when supported_parameters is not present")
				}
				if model2.ReleaseDate != nil {
					t.Error("Expected release date to be nil for invalid timestamp")
				}
				if model2.Description != nil {
					t.Error("Expected description to be nil for empty string")
				}

				// Third model: pricing only, no conversion needed
				model3 := models[2]
				if model3.Price == nil {
					t.Error("Expected pricing to be set")
				} else {
					// Already in per-million format (>= 0.01)
					if model3.Price.Input != 2.5 {
						t.Errorf("Expected input price 2.5, got %f", model3.Price.Input)
					}
					if model3.Price.Output != 10.0 {
						t.Errorf("Expected output price 10.0, got %f", model3.Price.Output)
					}
				}
			},
		},
		{
			name:       "fallback parsing when main structure fails",
			statusCode: http.StatusOK,
			response: `{
				"data": [
					{
						"id": "fallback-model-1",
						"unsupported_field": {"complex": "structure"},
						"created": "invalid_type_not_int",
						"description": 12345
					},
					{
						"id": "fallback-model-2",
						"supported_parameters": "invalid_type_not_array"
					}
				]
			}`,
			wantModels: 2,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				// Verify fallback worked and basic names are set
				if models[0].Name != "fallback-model-1" {
					t.Errorf("Expected name 'fallback-model-1', got '%s'", models[0].Name)
				}
				if models[1].Name != "fallback-model-2" {
					t.Errorf("Expected name 'fallback-model-2', got '%s'", models[1].Name)
				}

				// Verify extended fields are not set in fallback mode
				if models[0].Description != nil {
					t.Error("Expected description to be nil in fallback mode")
				}
				if models[0].ReleaseDate != nil {
					t.Error("Expected release date to be nil in fallback mode")
				}
				if models[0].Thinking != nil {
					t.Error("Expected thinking to be nil in fallback mode")
				}
				if models[0].Price != nil {
					t.Error("Expected price to be nil in fallback mode")
				}
			},
		},
		{
			name:       "error responses",
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "internal server error"}`,
			wantModels: 0,
		},
		{
			name:       "invalid json response",
			statusCode: http.StatusOK,
			response:   `{"data": [invalid json}`,
			wantModels: 0,
		},
		{
			name:       "empty models list",
			statusCode: http.StatusOK,
			response:   `{"data": []}`,
			wantModels: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/models" {
					t.Errorf("Expected /models path, got %s", r.URL.Path)
				}
				if r.Method != "GET" {
					t.Errorf("Expected GET method, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.response)
			}))
			defer server.Close()

			client := &http.Client{Timeout: 5 * time.Second}
			models := loadModelsFromServer(server.URL, "test-key", client)

			if len(models) != tt.wantModels {
				t.Errorf("Expected %d models, got %d", tt.wantModels, len(models))
				return
			}

			if tt.validate != nil {
				tt.validate(t, models)
			}
		})
	}
}

func TestLoadModelsFromServerTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(4 * time.Second) // longer than 3s timeout
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data": []}`)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	models := loadModelsFromServer(server.URL, "test-key", client)

	if len(models) != 0 {
		t.Errorf("Expected 0 models due to timeout, got %d", len(models))
	}
}

func TestLoadModelsFromServerHeaders(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		expectAuth string
	}{
		{
			name:       "with API key",
			apiKey:     "test-key-123",
			expectAuth: "Bearer test-key-123",
		},
		{
			name:       "without API key",
			apiKey:     "",
			expectAuth: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := r.Header.Get("Authorization")
				if auth != tt.expectAuth {
					t.Errorf("Expected Authorization '%s', got '%s'", tt.expectAuth, auth)
				}

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, `{"data": [{"id": "test-model"}]}`)
			}))
			defer server.Close()

			client := &http.Client{}
			models := loadModelsFromServer(server.URL, tt.apiKey, client)

			if len(models) != 1 {
				t.Errorf("Expected 1 model, got %d", len(models))
			}
		})
	}
}

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
					"supported_parameters": ["reasoning", "max_tokens"],
					"pricing": {"prompt": "0.001", "completion": "0.002"}
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
	if len(models) != 2 {
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

	// Verify second model with reasoning and pricing
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
		// 0.001 * 1000000 = 1000.0 (automatic conversion)
		if model2.Price.Input != 1000.0 {
			t.Errorf("Expected input price 1000.0, got %f", model2.Price.Input)
		}
		// 0.002 * 1000000 = 2000.0 (automatic conversion)
		if model2.Price.Output != 2000.0 {
			t.Errorf("Expected output price 2000.0, got %f", model2.Price.Output)
		}
	}
}
