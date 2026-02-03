package custom

import (
	"encoding/json"
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
						"supported_parameters": ["max_tokens", "tools", "temperature"]
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
					"supported_parameters": ["reasoning", "max_tokens", "tools"],
					"pricing": {"prompt": "0.001", "completion": "0.002"}
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

func TestPatchProviderConfigWithProviderName(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		providerName string
		validate     func(*testing.T, []byte)
		description  string
	}{
		{
			name:         "empty provider name returns original data",
			providerName: "",
			configData: `{
				"simple": {"model": "gpt-4"},
				"assistant": {"model": "claude-3"}
			}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "gpt-4" {
					t.Errorf("Expected model 'gpt-4', got '%s'", config.Simple.Model)
				}
				if config.Assistant.Model != "claude-3" {
					t.Errorf("Expected model 'claude-3', got '%s'", config.Assistant.Model)
				}
			},
			description: "should return original data when provider name is empty",
		},
		{
			name:         "patch JSON config with multiple agents",
			providerName: "openrouter",
			configData: `{
				"simple": {"model": "gpt-4o", "temperature": 0.7},
				"assistant": {"model": "claude-3-sonnet", "max_tokens": 4096},
				"generator": {"model": "gemini-pro"},
				"coder": {"model": "deepseek-coder"}
			}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "openrouter/gpt-4o" {
					t.Errorf("Expected 'openrouter/gpt-4o', got '%s'", config.Simple.Model)
				}
				if config.Assistant.Model != "openrouter/claude-3-sonnet" {
					t.Errorf("Expected 'openrouter/claude-3-sonnet', got '%s'", config.Assistant.Model)
				}
				if config.Generator.Model != "openrouter/gemini-pro" {
					t.Errorf("Expected 'openrouter/gemini-pro', got '%s'", config.Generator.Model)
				}
				if config.Coder.Model != "openrouter/deepseek-coder" {
					t.Errorf("Expected 'openrouter/deepseek-coder', got '%s'", config.Coder.Model)
				}
				// Verify other parameters are preserved
				if config.Simple.Temperature != 0.7 {
					t.Errorf("Expected temperature 0.7, got %f", config.Simple.Temperature)
				}
				if config.Assistant.MaxTokens != 4096 {
					t.Errorf("Expected max_tokens 4096, got %d", config.Assistant.MaxTokens)
				}
			},
			description: "should patch all agent models with provider prefix",
		},
		{
			name:         "patch YAML config",
			providerName: "anthropic",
			configData: `
simple:
  model: claude-3-opus
  temperature: 0.5
refiner:
  model: claude-3-sonnet
  max_tokens: 2048
`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "anthropic/claude-3-opus" {
					t.Errorf("Expected 'anthropic/claude-3-opus', got '%s'", config.Simple.Model)
				}
				if config.Refiner.Model != "anthropic/claude-3-sonnet" {
					t.Errorf("Expected 'anthropic/claude-3-sonnet', got '%s'", config.Refiner.Model)
				}
			},
			description: "should parse YAML and patch model names",
		},
		{
			name:         "agents with empty model names should not be patched",
			providerName: "provider",
			configData: `{
				"simple": {"model": "", "temperature": 0.9},
				"assistant": {"model": "gpt-4"},
				"generator": {"temperature": 0.5}
			}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "" {
					t.Errorf("Expected empty model, got '%s'", config.Simple.Model)
				}
				if config.Assistant.Model != "provider/gpt-4" {
					t.Errorf("Expected 'provider/gpt-4', got '%s'", config.Assistant.Model)
				}
				if config.Generator.Model != "" {
					t.Errorf("Expected empty model, got '%s'", config.Generator.Model)
				}
			},
			description: "should not add prefix to empty model names",
		},
		{
			name:         "all agent types are patched",
			providerName: "test",
			configData: `{
				"simple": {"model": "m1"},
				"simple_json": {"model": "m2"},
				"primary_agent": {"model": "m3"},
				"assistant": {"model": "m4"},
				"generator": {"model": "m5"},
				"refiner": {"model": "m6"},
				"adviser": {"model": "m7"},
				"reflector": {"model": "m8"},
				"searcher": {"model": "m9"},
				"enricher": {"model": "m10"},
				"coder": {"model": "m11"},
				"installer": {"model": "m12"},
				"pentester": {"model": "m13"}
			}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				expectedModels := map[string]string{
					"simple":        "test/m1",
					"simple_json":   "test/m2",
					"primary_agent": "test/m3",
					"assistant":     "test/m4",
					"generator":     "test/m5",
					"refiner":       "test/m6",
					"adviser":       "test/m7",
					"reflector":     "test/m8",
					"searcher":      "test/m9",
					"enricher":      "test/m10",
					"coder":         "test/m11",
					"installer":     "test/m12",
					"pentester":     "test/m13",
				}

				if config.Simple.Model != expectedModels["simple"] {
					t.Errorf("Simple: expected '%s', got '%s'", expectedModels["simple"], config.Simple.Model)
				}
				if config.SimpleJSON.Model != expectedModels["simple_json"] {
					t.Errorf("SimpleJSON: expected '%s', got '%s'", expectedModels["simple_json"], config.SimpleJSON.Model)
				}
				if config.PrimaryAgent.Model != expectedModels["primary_agent"] {
					t.Errorf("PrimaryAgent: expected '%s', got '%s'", expectedModels["primary_agent"], config.PrimaryAgent.Model)
				}
				if config.Assistant.Model != expectedModels["assistant"] {
					t.Errorf("Assistant: expected '%s', got '%s'", expectedModels["assistant"], config.Assistant.Model)
				}
				if config.Generator.Model != expectedModels["generator"] {
					t.Errorf("Generator: expected '%s', got '%s'", expectedModels["generator"], config.Generator.Model)
				}
				if config.Refiner.Model != expectedModels["refiner"] {
					t.Errorf("Refiner: expected '%s', got '%s'", expectedModels["refiner"], config.Refiner.Model)
				}
				if config.Adviser.Model != expectedModels["adviser"] {
					t.Errorf("Adviser: expected '%s', got '%s'", expectedModels["adviser"], config.Adviser.Model)
				}
				if config.Reflector.Model != expectedModels["reflector"] {
					t.Errorf("Reflector: expected '%s', got '%s'", expectedModels["reflector"], config.Reflector.Model)
				}
				if config.Searcher.Model != expectedModels["searcher"] {
					t.Errorf("Searcher: expected '%s', got '%s'", expectedModels["searcher"], config.Searcher.Model)
				}
				if config.Enricher.Model != expectedModels["enricher"] {
					t.Errorf("Enricher: expected '%s', got '%s'", expectedModels["enricher"], config.Enricher.Model)
				}
				if config.Coder.Model != expectedModels["coder"] {
					t.Errorf("Coder: expected '%s', got '%s'", expectedModels["coder"], config.Coder.Model)
				}
				if config.Installer.Model != expectedModels["installer"] {
					t.Errorf("Installer: expected '%s', got '%s'", expectedModels["installer"], config.Installer.Model)
				}
				if config.Pentester.Model != expectedModels["pentester"] {
					t.Errorf("Pentester: expected '%s', got '%s'", expectedModels["pentester"], config.Pentester.Model)
				}
			},
			description: "should patch all 13 agent types",
		},
		{
			name:         "invalid JSON returns original data",
			providerName: "provider",
			configData:   `{"simple": {"model": "test", "invalid": }`,
			validate: func(t *testing.T, result []byte) {
				expected := `{"simple": {"model": "test", "invalid": }`
				if string(result) != expected {
					t.Errorf("Expected original data to be returned unchanged")
				}
			},
			description: "should return original data on unmarshal error",
		},
		{
			name:         "invalid YAML returns original data",
			providerName: "provider",
			configData: `
simple:
  model: test
  invalid: [unclosed
`,
			validate: func(t *testing.T, result []byte) {
				if len(result) == 0 {
					t.Error("Expected original data, got empty result")
				}
			},
			description: "should return original data when both YAML and JSON parsing fail",
		},
		{
			name:         "special characters in provider name",
			providerName: "my-provider.v2",
			configData:   `{"simple": {"model": "gpt-4"}}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "my-provider.v2/gpt-4" {
					t.Errorf("Expected 'my-provider.v2/gpt-4', got '%s'", config.Simple.Model)
				}
			},
			description: "should handle special characters in provider name",
		},
		{
			name:         "special characters in model name",
			providerName: "provider",
			configData:   `{"simple": {"model": "claude-3.5-sonnet@20241022"}}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "provider/claude-3.5-sonnet@20241022" {
					t.Errorf("Expected 'provider/claude-3.5-sonnet@20241022', got '%s'", config.Simple.Model)
				}
			},
			description: "should handle special characters in model name",
		},
		{
			name:         "model name already contains slash",
			providerName: "openrouter",
			configData:   `{"simple": {"model": "anthropic/claude-3-opus"}}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				// Function adds prefix regardless of existing slashes
				if config.Simple.Model != "openrouter/anthropic/claude-3-opus" {
					t.Errorf("Expected 'openrouter/anthropic/claude-3-opus', got '%s'", config.Simple.Model)
				}
			},
			description: "should add prefix even if model name contains slash",
		},
		{
			name:         "empty config object",
			providerName: "provider",
			configData:   `{}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				// All fields should be nil
				if config.Simple != nil && config.Simple.Model != "" {
					t.Error("Expected no simple config")
				}
			},
			description: "should handle empty config object",
		},
		{
			name:         "preserve non-model fields",
			providerName: "provider",
			configData: `{
				"simple": {
					"model": "gpt-4",
					"temperature": 0.8,
					"max_tokens": 1000,
					"top_p": 0.95,
					"frequency_penalty": 0.5
				}
			}`,
			validate: func(t *testing.T, result []byte) {
				var config pconfig.ProviderConfig
				if err := json.Unmarshal(result, &config); err != nil {
					t.Fatalf("Failed to unmarshal result: %v", err)
				}
				if config.Simple.Model != "provider/gpt-4" {
					t.Errorf("Expected 'provider/gpt-4', got '%s'", config.Simple.Model)
				}
				if config.Simple.Temperature != 0.8 {
					t.Errorf("Expected temperature 0.8, got %f", config.Simple.Temperature)
				}
				if config.Simple.MaxTokens != 1000 {
					t.Errorf("Expected max_tokens 1000, got %d", config.Simple.MaxTokens)
				}
				if config.Simple.TopP != 0.95 {
					t.Errorf("Expected top_p 0.95, got %f", config.Simple.TopP)
				}
				if config.Simple.FrequencyPenalty != 0.5 {
					t.Errorf("Expected frequency_penalty 0.5, got %f", config.Simple.FrequencyPenalty)
				}
			},
			description: "should preserve all non-model configuration fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := patchProviderConfigWithProviderName([]byte(tt.configData), tt.providerName)

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
