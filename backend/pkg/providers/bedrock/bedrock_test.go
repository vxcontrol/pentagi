package bedrock

import (
	"fmt"
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

// TestBuildMinimalToolsFromChain verifies the helper that reconstructs minimal
// tool definitions from a conversation chain.  This is the foundation of the
// fix for the AWS Bedrock Converse API ValidationException that occurs when
// messages contain toolUse / toolResult blocks but no toolConfig is provided.
func TestBuildMinimalToolsFromChain(t *testing.T) {
	t.Run("empty chain returns nil", func(t *testing.T) {
		result := buildMinimalToolsFromChain(nil)
		if result != nil {
			t.Errorf("expected nil for empty chain, got %v", result)
		}

		result = buildMinimalToolsFromChain([]llms.MessageContent{})
		if result != nil {
			t.Errorf("expected nil for empty chain slice, got %v", result)
		}
	})

	t.Run("chain with only text messages returns nil", func(t *testing.T) {
		chain := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, "You are helpful."),
			llms.TextParts(llms.ChatMessageTypeHuman, "What is 2+2?"),
			llms.TextParts(llms.ChatMessageTypeAI, "4"),
		}
		result := buildMinimalToolsFromChain(chain)
		if result != nil {
			t.Errorf("expected nil for text-only chain, got %v", result)
		}
	})

	t.Run("chain with ToolCall returns tool definition", func(t *testing.T) {
		chain := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather?"),
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "call_abc",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location":"NYC"}`,
						},
					},
				},
			},
		}

		result := buildMinimalToolsFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool definition, got %d", len(result))
		}
		if result[0].Function == nil {
			t.Fatal("tool Function must not be nil")
		}
		if result[0].Function.Name != "get_weather" {
			t.Errorf("expected tool name 'get_weather', got '%s'", result[0].Function.Name)
		}
		if result[0].Type != "function" {
			t.Errorf("expected type 'function', got '%s'", result[0].Type)
		}
		if result[0].Function.Parameters == nil {
			t.Error("expected non-nil Parameters (placeholder schema)")
		}
	})

	t.Run("chain with ToolCallResponse returns tool definition", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: "call_abc",
						Name:       "calculate",
						Content:    "42",
					},
				},
			},
		}

		result := buildMinimalToolsFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool definition, got %d", len(result))
		}
		if result[0].Function.Name != "calculate" {
			t.Errorf("expected tool name 'calculate', got '%s'", result[0].Function.Name)
		}
	})

	t.Run("chain with both ToolCall and ToolCallResponse deduplicates names", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "call_1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "search",
							Arguments: `{"query":"go"}`,
						},
					},
				},
			},
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: "call_1",
						Name:       "search", // same name — should deduplicate
						Content:    "results",
					},
				},
			},
		}

		result := buildMinimalToolsFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 deduplicated tool definition, got %d (%v)", len(result), toolNames(result))
		}
		if result[0].Function.Name != "search" {
			t.Errorf("expected tool name 'search', got '%s'", result[0].Function.Name)
		}
	})

	t.Run("chain with multiple distinct tool names returns all", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{Name: "search"},
					},
				},
			},
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "c1", Name: "search", Content: "r"},
				},
			},
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c2",
						Type: "function",
						FunctionCall: &llms.FunctionCall{Name: "execute_command"},
					},
				},
			},
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "c2", Name: "execute_command", Content: "ok"},
				},
			},
		}

		result := buildMinimalToolsFromChain(chain)
		if len(result) != 2 {
			t.Fatalf("expected 2 tool definitions, got %d (%v)", len(result), toolNames(result))
		}

		names := toolNames(result)
		sort.Strings(names)
		expected := []string{"execute_command", "search"}
		for i, n := range expected {
			if names[i] != n {
				t.Errorf("expected name[%d]=%s, got %s", i, n, names[i])
			}
		}
	})

	t.Run("ToolCall without FunctionCall is skipped", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{ID: "c1", Type: "function"},
					// FunctionCall is nil — should be ignored
				},
			},
		}
		result := buildMinimalToolsFromChain(chain)
		if result != nil {
			t.Errorf("expected nil when ToolCall has nil FunctionCall, got %v", result)
		}
	})

	t.Run("generated tool has valid placeholder schema", func(t *testing.T) {
		chain := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.ToolCall{
						ID:   "c1",
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      "my_tool",
							Arguments: `{}`,
						},
					},
				},
			},
		}

		result := buildMinimalToolsFromChain(chain)
		if len(result) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(result))
		}
		tool := result[0]

		// Verify placeholder description follows the expected format
		expectedDesc := fmt.Sprintf("Tool: %s", "my_tool")
		if tool.Function.Description != expectedDesc {
			t.Errorf("expected description %q, got %q", expectedDesc, tool.Function.Description)
		}

		// Verify placeholder schema has the required object type
		schema, ok := tool.Function.Parameters.(map[string]any)
		if !ok {
			t.Fatalf("expected Parameters to be map[string]any, got %T", tool.Function.Parameters)
		}
		if schema["type"] != "object" {
			t.Errorf("expected schema type 'object', got %v", schema["type"])
		}
		props, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("expected properties to be map[string]any, got %T", schema["properties"])
		}
		if len(props) != 0 {
			t.Errorf("expected empty properties map, got %v", props)
		}
	})
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
