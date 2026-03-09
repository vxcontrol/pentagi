package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pentagi/pkg/providers/pconfig"
)

func TestApplyModelPrefix(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		prefix    string
		expected  string
	}{
		{
			name:      "with prefix",
			modelName: "deepseek-chat",
			prefix:    "deepseek",
			expected:  "deepseek/deepseek-chat",
		},
		{
			name:      "without prefix (empty string)",
			modelName: "deepseek-chat",
			prefix:    "",
			expected:  "deepseek-chat",
		},
		{
			name:      "model already has different prefix",
			modelName: "anthropic/claude-3",
			prefix:    "openrouter",
			expected:  "openrouter/anthropic/claude-3",
		},
		{
			name:      "complex model name with special chars",
			modelName: "claude-3.5-sonnet@20241022",
			prefix:    "provider",
			expected:  "provider/claude-3.5-sonnet@20241022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyModelPrefix(tt.modelName, tt.prefix)
			if result != tt.expected {
				t.Errorf("ApplyModelPrefix(%q, %q) = %q, want %q",
					tt.modelName, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestRemoveModelPrefix(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		prefix    string
		expected  string
	}{
		{
			name:      "model with matching prefix",
			modelName: "deepseek/deepseek-chat",
			prefix:    "deepseek",
			expected:  "deepseek-chat",
		},
		{
			name:      "model without prefix",
			modelName: "deepseek-chat",
			prefix:    "deepseek",
			expected:  "deepseek-chat",
		},
		{
			name:      "empty prefix",
			modelName: "deepseek/deepseek-chat",
			prefix:    "",
			expected:  "deepseek/deepseek-chat",
		},
		{
			name:      "model with different prefix",
			modelName: "openrouter/deepseek-chat",
			prefix:    "deepseek",
			expected:  "openrouter/deepseek-chat",
		},
		{
			name:      "model with nested prefixes",
			modelName: "openrouter/anthropic/claude-3",
			prefix:    "openrouter",
			expected:  "anthropic/claude-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveModelPrefix(tt.modelName, tt.prefix)
			if result != tt.expected {
				t.Errorf("RemoveModelPrefix(%q, %q) = %q, want %q",
					tt.modelName, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestLoadModelsFromYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlData    string
		prefix      string
		expectError bool
		validate    func(*testing.T, pconfig.ModelsConfig)
	}{
		{
			name: "basic models without prefix",
			yamlData: `
- name: deepseek-chat
  description: DeepSeek chat model
  thinking: false
  price:
    input: 0.28
    output: 0.42

- name: deepseek-reasoner
  description: DeepSeek reasoning model
  thinking: true
  price:
    input: 0.28
    output: 0.42
`,
			prefix:      "",
			expectError: false,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				if len(models) != 2 {
					t.Fatalf("Expected 2 models, got %d", len(models))
				}
				if models[0].Name != "deepseek-chat" {
					t.Errorf("Expected first model name 'deepseek-chat', got %q", models[0].Name)
				}
				if models[1].Name != "deepseek-reasoner" {
					t.Errorf("Expected second model name 'deepseek-reasoner', got %q", models[1].Name)
				}
				if models[0].Thinking != nil && *models[0].Thinking {
					t.Error("Expected first model thinking=false")
				}
				if models[1].Thinking == nil || !*models[1].Thinking {
					t.Error("Expected second model thinking=true")
				}
			},
		},
		{
			name: "models with all metadata fields",
			yamlData: `
- name: gpt-4o
  description: GPT-4 Optimized
  release_date: 2024-05-13
  thinking: false
  price:
    input: 5.0
    output: 15.0
`,
			prefix:      "",
			expectError: false,
			validate: func(t *testing.T, models pconfig.ModelsConfig) {
				if len(models) != 1 {
					t.Fatalf("Expected 1 model, got %d", len(models))
				}
				model := models[0]
				if model.Name != "gpt-4o" {
					t.Errorf("Expected model name 'gpt-4o', got %q", model.Name)
				}
				if model.Description == nil || *model.Description != "GPT-4 Optimized" {
					t.Error("Expected description 'GPT-4 Optimized'")
				}
				if model.ReleaseDate == nil {
					t.Error("Expected release_date to be set")
				}
				if model.Price == nil {
					t.Error("Expected price to be set")
				} else {
					if model.Price.Input != 5.0 {
						t.Errorf("Expected input price 5.0, got %f", model.Price.Input)
					}
					if model.Price.Output != 15.0 {
						t.Errorf("Expected output price 15.0, got %f", model.Price.Output)
					}
				}
			},
		},
		{
			name:        "invalid YAML",
			yamlData:    `invalid: [unclosed`,
			prefix:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models, err := pconfig.LoadModelsConfigData([]byte(tt.yamlData))
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, models)
			}
		})
	}
}

func TestLoadModelsFromHTTP_WithoutPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("Expected /models path, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		response := `{
			"data": [
				{
					"id": "model-a",
					"description": "Model A description",
					"supported_parameters": ["tools", "max_tokens"]
				},
				{
					"id": "model-b",
					"created": 1686588896,
					"description": "Model B description",
					"supported_parameters": ["reasoning", "tools"],
					"pricing": {
						"prompt": "0.0001",
						"completion": "0.0005"
					}
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	models, err := LoadModelsFromHTTP(server.URL, "test-key", client, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	// Verify first model
	if models[0].Name != "model-a" {
		t.Errorf("Expected first model name 'model-a', got %q", models[0].Name)
	}
	if models[0].Description == nil || *models[0].Description != "Model A description" {
		t.Error("Expected description for first model")
	}

	// Verify second model with all metadata
	if models[1].Name != "model-b" {
		t.Errorf("Expected second model name 'model-b', got %q", models[1].Name)
	}
	if models[1].Thinking == nil || !*models[1].Thinking {
		t.Error("Expected thinking capability for second model")
	}
	if models[1].Price == nil {
		t.Error("Expected pricing for second model")
	} else {
		// 0.0001 * 1000000 = 100.0
		if models[1].Price.Input != 100.0 {
			t.Errorf("Expected input price 100.0, got %f", models[1].Price.Input)
		}
		// 0.0005 * 1000000 = 500.0
		if models[1].Price.Output != 500.0 {
			t.Errorf("Expected output price 500.0, got %f", models[1].Price.Output)
		}
	}
}

func TestLoadModelsFromHTTP_WithPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate LiteLLM proxy returning models from multiple providers
		response := `{
			"data": [
				{
					"id": "deepseek/deepseek-chat",
					"description": "DeepSeek chat model",
					"supported_parameters": ["tools", "max_tokens"],
					"pricing": {
						"prompt": "0.28",
						"completion": "0.42"
					}
				},
				{
					"id": "deepseek/deepseek-reasoner",
					"description": "DeepSeek reasoning model",
					"supported_parameters": ["reasoning", "tools"],
					"pricing": {
						"prompt": "0.28",
						"completion": "0.42"
					}
				},
				{
					"id": "openai/gpt-4",
					"description": "GPT-4 model",
					"supported_parameters": ["tools"],
					"pricing": {
						"prompt": "30.0",
						"completion": "60.0"
					}
				},
				{
					"id": "anthropic/claude-3-opus",
					"description": "Claude 3 Opus",
					"supported_parameters": ["tools"],
					"pricing": {
						"prompt": "15.0",
						"completion": "75.0"
					}
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	models, err := LoadModelsFromHTTP(server.URL, "test-key", client, "deepseek")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only include deepseek models, with prefix stripped
	if len(models) != 2 {
		t.Fatalf("Expected 2 deepseek models, got %d", len(models))
	}

	// Verify model names have prefix stripped
	if models[0].Name != "deepseek-chat" {
		t.Errorf("Expected model name 'deepseek-chat' (without prefix), got %q", models[0].Name)
	}
	if models[1].Name != "deepseek-reasoner" {
		t.Errorf("Expected model name 'deepseek-reasoner' (without prefix), got %q", models[1].Name)
	}

	// Verify metadata is preserved
	if models[0].Description == nil || *models[0].Description != "DeepSeek chat model" {
		t.Error("Expected description for first model")
	}
	if models[1].Thinking == nil || !*models[1].Thinking {
		t.Error("Expected reasoning capability for second model")
	}

	// Verify pricing (should be in per-million-token format, not modified)
	if models[0].Price == nil {
		t.Error("Expected pricing for first model")
	} else {
		if models[0].Price.Input != 0.28 {
			t.Errorf("Expected input price 0.28, got %f", models[0].Price.Input)
		}
	}
}

func TestLoadModelsFromHTTP_FallbackParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simplified response format
		response := `{
			"data": [
				{"id": "model-1"},
				{"id": "model-2"}
			]
		}`
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	models, err := LoadModelsFromHTTP(server.URL, "", client, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	if models[0].Name != "model-1" {
		t.Errorf("Expected model name 'model-1', got %q", models[0].Name)
	}
	if models[1].Name != "model-2" {
		t.Errorf("Expected model name 'model-2', got %q", models[1].Name)
	}
}

func TestLoadModelsFromHTTP_SkipModelsWithoutTools(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": [
				{
					"id": "model-with-tools",
					"supported_parameters": ["tools", "max_tokens"]
				},
				{
					"id": "model-without-tools",
					"supported_parameters": ["max_tokens", "temperature"]
				},
				{
					"id": "model-with-structured-outputs",
					"supported_parameters": ["structured_outputs"]
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	models, err := LoadModelsFromHTTP(server.URL, "", client, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should only include models with tools or structured_outputs
	if len(models) != 2 {
		t.Fatalf("Expected 2 models (with tools/structured_outputs), got %d", len(models))
	}

	if models[0].Name != "model-with-tools" {
		t.Errorf("Expected first model 'model-with-tools', got %q", models[0].Name)
	}
	if models[1].Name != "model-with-structured-outputs" {
		t.Errorf("Expected second model 'model-with-structured-outputs', got %q", models[1].Name)
	}
}

func TestLoadModelsFromHTTP_Errors(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *httptest.Server
		expectError bool
	}{
		{
			name: "HTTP error status",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, `{"error": "internal error"}`)
				}))
			},
			expectError: true,
		},
		{
			name: "invalid JSON response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					fmt.Fprint(w, `{invalid json}`)
				}))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			client := &http.Client{Timeout: 5 * time.Second}
			_, err := LoadModelsFromHTTP(server.URL, "", client, "")
			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// TestEndToEndProviderSimulation simulates complete provider lifecycle with prefix handling
func TestEndToEndProviderSimulation(t *testing.T) {
	// Setup mock HTTP server simulating LiteLLM proxy
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": [
				{
					"id": "moonshot/kimi-k2-turbo",
					"description": "Kimi K2 Turbo",
					"supported_parameters": ["tools", "reasoning"],
					"pricing": {
						"prompt": "0.0001",
						"completion": "0.0002"
					}
				},
				{
					"id": "moonshot/kimi-k2.5",
					"description": "Kimi K2.5",
					"supported_parameters": ["tools"],
					"pricing": {
						"prompt": "0.00015",
						"completion": "0.0003"
					}
				},
				{
					"id": "openai/gpt-4o",
					"supported_parameters": ["tools"]
				}
			]
		}`
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	}))
	defer server.Close()

	// Step 1: Load models from HTTP with LiteLLM prefix
	client := &http.Client{Timeout: 5 * time.Second}
	providerPrefix := "moonshot"
	models, err := LoadModelsFromHTTP(server.URL, "test-key", client, providerPrefix)
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	// Verify only moonshot models loaded, with prefix stripped
	if len(models) != 2 {
		t.Fatalf("Expected 2 moonshot models, got %d", len(models))
	}
	if models[0].Name != "kimi-k2-turbo" {
		t.Errorf("Expected 'kimi-k2-turbo' (without prefix), got %q", models[0].Name)
	}
	if models[1].Name != "kimi-k2.5" {
		t.Errorf("Expected 'kimi-k2.5' (without prefix), got %q", models[1].Name)
	}

	// Step 2: Simulate Model() call - should return without prefix
	modelWithoutPrefix := models[0].Name
	if modelWithoutPrefix != "kimi-k2-turbo" {
		t.Errorf("Model() should return 'kimi-k2-turbo', got %q", modelWithoutPrefix)
	}

	// Step 3: Simulate ModelWithPrefix() call - should return with prefix
	modelWithPrefix := ApplyModelPrefix(modelWithoutPrefix, providerPrefix)
	if modelWithPrefix != "moonshot/kimi-k2-turbo" {
		t.Errorf("ModelWithPrefix() should return 'moonshot/kimi-k2-turbo', got %q", modelWithPrefix)
	}

	// Step 4: Verify round-trip consistency
	stripped := RemoveModelPrefix(modelWithPrefix, providerPrefix)
	if stripped != modelWithoutPrefix {
		t.Errorf("Round-trip failed: %q -> %q -> %q", modelWithoutPrefix, modelWithPrefix, stripped)
	}

	// Step 5: Verify metadata preservation
	if models[0].Price == nil {
		t.Error("Expected pricing information to be preserved")
	} else {
		// 0.0001 * 1000000 = 100.0
		if models[0].Price.Input != 100.0 {
			t.Errorf("Expected input price 100.0, got %f", models[0].Price.Input)
		}
	}
	if models[0].Thinking == nil || !*models[0].Thinking {
		t.Error("Expected reasoning capability to be preserved")
	}
}
