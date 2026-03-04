package bedrock

import (
	"sort"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
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

// toolNames is a test helper that extracts tool names from a slice of llms.Tool.
func toolNames(tools []llms.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		if t.Function != nil {
			names = append(names, t.Function.Name)
		}
	}
	return names
}

// TestInferPropertyType verifies type inference for individual property values.
func TestInferPropertyType(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"nil value", nil, "null"},
		{"string", "hello", "string"},
		{"empty string", "", "string"},
		{"boolean true", true, "boolean"},
		{"boolean false", false, "boolean"},
		{"int", 42, "number"},
		{"int64", int64(42), "number"},
		{"float32", float32(3.14), "number"},
		{"float64", 3.14159, "number"},
		{"slice", []int{1, 2, 3}, "array"},
		{"empty slice", []string{}, "array"},
		{"map", map[string]string{"key": "value"}, "object"},
		{"empty map", map[string]any{}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferPropertyType(tt.value)
			if result != tt.expected {
				t.Errorf("inferPropertyType(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

// TestInferSchemaFromArguments verifies JSON schema inference from argument samples.
func TestInferSchemaFromArguments(t *testing.T) {
	t.Run("no samples returns empty schema", func(t *testing.T) {
		schema := inferSchemaFromArguments(nil)
		if schema["type"] != "object" {
			t.Errorf("expected type 'object', got %v", schema["type"])
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok || len(props) != 0 {
			t.Errorf("expected empty properties, got %v", schema["properties"])
		}
	})

	t.Run("single sample with multiple types", func(t *testing.T) {
		samples := []string{`{"name":"test","count":5,"active":true,"tags":["a","b"],"meta":{}}`}
		schema := inferSchemaFromArguments(samples)

		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("expected properties map, got %T", schema["properties"])
		}

		expectedTypes := map[string]string{
			"name":   "string",
			"count":  "number",
			"active": "boolean",
			"tags":   "array",
			"meta":   "object",
		}

		for key, expectedType := range expectedTypes {
			prop, exists := props[key]
			if !exists {
				t.Errorf("property %q not found in schema", key)
				continue
			}
			propMap, ok := prop.(map[string]any)
			if !ok {
				t.Errorf("property %q is not a map", key)
				continue
			}
			if propMap["type"] != expectedType {
				t.Errorf("property %q type = %v, want %v", key, propMap["type"], expectedType)
			}
		}
	})

	t.Run("multiple samples aggregate properties", func(t *testing.T) {
		samples := []string{
			`{"field1":"value1"}`,
			`{"field2":42}`,
			`{"field3":true}`,
		}
		schema := inferSchemaFromArguments(samples)

		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("expected properties map")
		}

		if len(props) != 3 {
			t.Errorf("expected 3 properties, got %d", len(props))
		}

		expectedTypes := map[string]string{
			"field1": "string",
			"field2": "number",
			"field3": "boolean",
		}

		for key, expectedType := range expectedTypes {
			prop := props[key].(map[string]any)
			if prop["type"] != expectedType {
				t.Errorf("property %q type = %v, want %v", key, prop["type"], expectedType)
			}
		}
	})

	t.Run("invalid JSON is ignored", func(t *testing.T) {
		samples := []string{
			`{invalid json}`,
			`{"valid":"field"}`,
			`not json at all`,
		}
		schema := inferSchemaFromArguments(samples)

		props := schema["properties"].(map[string]any)
		if len(props) != 1 {
			t.Errorf("expected 1 valid property, got %d", len(props))
		}
	})

	t.Run("empty string samples are skipped", func(t *testing.T) {
		samples := []string{"", "", `{"key":"value"}`}
		schema := inferSchemaFromArguments(samples)

		props := schema["properties"].(map[string]any)
		if len(props) != 1 {
			t.Errorf("expected 1 property, got %d", len(props))
		}
	})

	t.Run("duplicate keys use first occurrence", func(t *testing.T) {
		samples := []string{
			`{"field":"string_value"}`,
			`{"field":123}`, // Same field with different type - should be ignored
		}
		schema := inferSchemaFromArguments(samples)

		props := schema["properties"].(map[string]any)
		fieldProp := props["field"].(map[string]any)
		if fieldProp["type"] != "string" {
			t.Errorf("expected first occurrence type 'string', got %v", fieldProp["type"])
		}
	})
}

// TestCollectToolUsageFromChain verifies tool usage collection from message chains.
func TestCollectToolUsageFromChain(t *testing.T) {
	t.Run("empty chain returns empty map", func(t *testing.T) {
		result := collectToolUsageFromChain(nil)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}

		result = collectToolUsageFromChain([]llms.MessageContent{})
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("collects from ToolCall", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "search",
							Arguments: `{"query":"test"}`,
						},
					},
				},
			},
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(result))
		}
		if args, ok := result["search"]; !ok || len(args) != 1 {
			t.Errorf("expected search tool with 1 argument, got %v", result)
		}
	})

	t.Run("collects from ToolCallResponse", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: "c1",
						Name:       "execute",
						Content:    "done",
					},
				},
			},
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(result))
		}
		if _, ok := result["execute"]; !ok {
			t.Errorf("expected execute tool in result")
		}
	})

	t.Run("aggregates multiple calls to same tool", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "calc",
							Arguments: `{"op":"add","a":1,"b":2}`,
						},
					},
				},
			},
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c2",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "calc",
							Arguments: `{"op":"multiply","a":3,"b":4}`,
						},
					},
				},
			},
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool (deduplicated), got %d", len(result))
		}
		if args := result["calc"]; len(args) != 2 {
			t.Errorf("expected 2 argument samples for calc, got %d", len(args))
		}
	})

	t.Run("handles mixed ToolCall and ToolCallResponse", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:           "c1",
						Type:         "function",
						FunctionCall: &llms.FunctionCall{Name: "tool1", Arguments: `{}`},
					},
				},
			},
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "c1", Name: "tool1", Content: "ok"},
				},
			},
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "c2", Name: "tool2", Content: "ok"},
				},
			},
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 2 {
			t.Fatalf("expected 2 tools, got %d", len(result))
		}
		if _, ok := result["tool1"]; !ok {
			t.Error("expected tool1 in result")
		}
		if _, ok := result["tool2"]; !ok {
			t.Error("expected tool2 in result")
		}
	})

	t.Run("ignores ToolCall without FunctionCall", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{ID: "c1", Type: "function"}, // no FunctionCall
				},
			},
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 0 {
			t.Errorf("expected empty result for ToolCall without FunctionCall, got %v", result)
		}
	})

	t.Run("ignores non-tool parts", func(t *testing.T) {
		chain := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "hello"),
			llms.TextParts(llms.ChatMessageTypeAI, "hi"),
		}

		result := collectToolUsageFromChain(chain)
		if len(result) != 0 {
			t.Errorf("expected empty result for text-only chain, got %v", result)
		}
	})
}

// TestRestoreMissedToolsFromChain verifies the main function that merges
// declared tools with inferred tools from the chain.
func TestRestoreMissedToolsFromChain(t *testing.T) {
	t.Run("empty chain returns original tools unchanged", func(t *testing.T) {
		declaredTools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "existing_tool",
					Description: "Already declared",
					Parameters:  map[string]any{"type": "object"},
				},
			},
		}

		result := restoreMissedToolsFromChain(nil, declaredTools)
		if len(result) != len(declaredTools) {
			t.Errorf("expected %d tools, got %d", len(declaredTools), len(result))
		}
	})

	t.Run("chain with no tool usage returns original tools", func(t *testing.T) {
		chain := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Hello"),
			llms.TextParts(llms.ChatMessageTypeAI, "Hi"),
		}
		declaredTools := []llms.Tool{
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool1"}},
		}

		result := restoreMissedToolsFromChain(chain, declaredTools)
		if len(result) != 1 {
			t.Errorf("expected 1 tool, got %d", len(result))
		}
	})

	t.Run("adds new tools from chain", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "new_tool",
							Arguments: `{"param":"value"}`,
						},
					},
				},
			},
		}

		result := restoreMissedToolsFromChain(chain, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(result))
		}
		if result[0].Function.Name != "new_tool" {
			t.Errorf("expected tool name 'new_tool', got %q", result[0].Function.Name)
		}

		// Verify inferred schema has the parameter
		schema, ok := result[0].Function.Parameters.(map[string]any)
		if !ok {
			t.Fatalf("expected Parameters to be map[string]any")
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("expected properties in schema")
		}
		if _, exists := props["param"]; !exists {
			t.Error("expected 'param' property in inferred schema")
		}
	})

	t.Run("does not overwrite existing tool declarations", func(t *testing.T) {
		declaredTools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "search",
					Description: "Custom search description",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query":  map[string]any{"type": "string"},
							"limit":  map[string]any{"type": "number"},
							"custom": map[string]any{"type": "boolean"},
						},
					},
				},
			},
		}

		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "search",
							Arguments: `{"query":"test"}`, // Different schema
						},
					},
				},
			},
		}

		result := restoreMissedToolsFromChain(chain, declaredTools)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool (not duplicated), got %d", len(result))
		}

		// Verify the declared tool was preserved exactly
		if result[0].Function.Description != "Custom search description" {
			t.Errorf("declared tool description was overwritten")
		}

		schema, ok := result[0].Function.Parameters.(map[string]any)
		if !ok {
			t.Fatalf("expected Parameters to be map[string]any")
		}
		props := schema["properties"].(map[string]any)
		if _, exists := props["custom"]; !exists {
			t.Error("declared tool schema was overwritten - 'custom' field missing")
		}
	})

	t.Run("merges declared and inferred tools", func(t *testing.T) {
		declaredTools := []llms.Tool{
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool_a"}},
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool_b"}},
		}

		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:           "c1",
						Type:         "function",
						FunctionCall: &llms.FunctionCall{Name: "tool_b", Arguments: `{}`},
					},
					llms.ToolCall{
						ID:           "c2",
						Type:         "function",
						FunctionCall: &llms.FunctionCall{Name: "tool_c", Arguments: `{}`},
					},
				},
			},
		}

		result := restoreMissedToolsFromChain(chain, declaredTools)

		// Should have tool_a, tool_b (declared), and tool_c (inferred)
		if len(result) != 3 {
			t.Fatalf("expected 3 tools, got %d (%v)", len(result), toolNames(result))
		}

		names := toolNames(result)
		sort.Strings(names)
		expected := []string{"tool_a", "tool_b", "tool_c"}
		for i, name := range expected {
			if names[i] != name {
				t.Errorf("expected tool[%d] = %q, got %q", i, name, names[i])
			}
		}
	})

	t.Run("handles complex schema inference", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "complex_tool",
							Arguments: `{"str":"text","num":42,"bool":true,"arr":[1,2,3],"obj":{"nested":"value"}}`,
						},
					},
				},
			},
		}

		result := restoreMissedToolsFromChain(chain, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(result))
		}

		schema := result[0].Function.Parameters.(map[string]any)
		props := schema["properties"].(map[string]any)

		expectedTypes := map[string]string{
			"str":  "string",
			"num":  "number",
			"bool": "boolean",
			"arr":  "array",
			"obj":  "object",
		}

		for key, expectedType := range expectedTypes {
			prop, exists := props[key]
			if !exists {
				t.Errorf("expected property %q in schema", key)
				continue
			}
			propMap := prop.(map[string]any)
			if propMap["type"] != expectedType {
				t.Errorf("property %q type = %v, want %v", key, propMap["type"], expectedType)
			}
		}
	})
}

// TestExtractToolsFromOptions verifies tool extraction from CallOptions.
func TestExtractToolsFromOptions(t *testing.T) {
	t.Run("empty options returns nil", func(t *testing.T) {
		result := extractToolsFromOptions(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}

		result = extractToolsFromOptions([]llms.CallOption{})
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("extracts tools from WithTools option", func(t *testing.T) {
		tools := []llms.Tool{
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool1"}},
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool2"}},
		}

		options := []llms.CallOption{
			llms.WithTools(tools),
		}

		result := extractToolsFromOptions(options)
		if len(result) != 2 {
			t.Errorf("expected 2 tools, got %d", len(result))
		}
	})

	t.Run("extracts tools from multiple options", func(t *testing.T) {
		tools := []llms.Tool{
			{Type: "function", Function: &llms.FunctionDefinition{Name: "tool1"}},
		}

		options := []llms.CallOption{
			llms.WithModel("test-model"),
			llms.WithTemperature(0.7),
			llms.WithTools(tools),
			llms.WithMaxTokens(100),
		}

		result := extractToolsFromOptions(options)
		if len(result) != 1 {
			t.Errorf("expected 1 tool, got %d", len(result))
		}
	})
}

// TestAuthenticationStrategies verifies all supported authentication methods.
func TestAuthenticationStrategies(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	t.Run("static credentials authentication", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:    "us-east-1",
			BedrockAccessKey: "test-access-key",
			BedrockSecretKey: "test-secret-key",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with static credentials: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
		if prov.Type() != provider.ProviderBedrock {
			t.Errorf("Expected provider type Bedrock, got %v", prov.Type())
		}
	})

	t.Run("static credentials with session token", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:       "us-west-2",
			BedrockAccessKey:    "test-access-key",
			BedrockSecretKey:    "test-secret-key",
			BedrockSessionToken: "test-session-token",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with session token: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("bearer token authentication", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:      "eu-west-1",
			BedrockBearerToken: "test-bearer-token-value",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with bearer token: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("default AWS authentication", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:      "ap-southeast-1",
			BedrockDefaultAuth: true,
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with default auth: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("bearer token takes precedence over static credentials", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:      "us-east-1",
			BedrockBearerToken: "bearer-token",
			BedrockAccessKey:   "access-key",
			BedrockSecretKey:   "secret-key",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("default auth takes precedence over all", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:      "us-east-1",
			BedrockDefaultAuth: true,
			BedrockBearerToken: "bearer-token",
			BedrockAccessKey:   "access-key",
			BedrockSecretKey:   "secret-key",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("custom server URL with authentication", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:    "us-east-1",
			BedrockServerURL: "https://custom-bedrock-endpoint.example.com",
			BedrockAccessKey: "test-key",
			BedrockSecretKey: "test-secret",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with custom server URL: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})

	t.Run("proxy configuration", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:    "us-east-1",
			BedrockAccessKey: "test-key",
			BedrockSecretKey: "test-secret",
			ProxyURL:         "http://proxy.example.com:8080",
		}

		prov, err := New(cfg, providerConfig)
		if err != nil {
			t.Fatalf("Failed to create provider with proxy: %v", err)
		}
		if prov == nil {
			t.Fatal("Expected provider to be created")
		}
	})
}

// TestAuthenticationErrors verifies error handling for invalid configurations.
func TestAuthenticationErrors(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	t.Run("no authentication method configured", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion: "us-east-1",
			// No auth credentials set
		}

		_, err := New(cfg, providerConfig)
		if err == nil {
			t.Error("Expected error when no authentication method is configured")
		}
		if err != nil && err.Error() != "no valid authentication method configured for Bedrock" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("only access key without secret key", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:    "us-east-1",
			BedrockAccessKey: "test-key",
			// BedrockSecretKey not set
		}

		_, err := New(cfg, providerConfig)
		if err == nil {
			t.Error("Expected error when only access key is provided")
		}
	})

	t.Run("only secret key without access key", func(t *testing.T) {
		cfg := &config.Config{
			BedrockRegion:    "us-east-1",
			BedrockSecretKey: "test-secret",
			// BedrockAccessKey not set
		}

		_, err := New(cfg, providerConfig)
		if err == nil {
			t.Error("Expected error when only secret key is provided")
		}
	})
}
