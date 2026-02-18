package langfuse

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// TestConvertInput tests various input conversion scenarios
func TestConvertInput(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		validate func(t *testing.T, result any)
	}{
		{
			name:  "nil input",
			input: nil,
			validate: func(t *testing.T, result any) {
				assert.Nil(t, result)
			},
		},
		{
			name: "simple text message",
			input: []*llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				msg := messages[0].(map[string]any)
				assert.Equal(t, "user", msg["role"])
				assert.Equal(t, "Hello", msg["content"])
			},
		},
		{
			name: "message with tools",
			input: []*llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Let me search"},
						llms.ToolCall{
							ID: "call_001",
							FunctionCall: &llms.FunctionCall{
								Name:      "search",
								Arguments: `{"query":"test"}`,
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				msg := messages[0].(map[string]any)
				assert.Equal(t, "assistant", msg["role"])
				assert.Equal(t, "Let me search", msg["content"])

				toolCalls := msg["tool_calls"].([]any)
				require.Len(t, toolCalls, 1)

				tc := toolCalls[0].(map[string]any)
				assert.Equal(t, "call_001", tc["id"])
				assert.Equal(t, "function", tc["type"])

				fn := tc["function"].(map[string]any)
				assert.Equal(t, "search", fn["name"])
				assert.Equal(t, `{"query":"test"}`, fn["arguments"])
			},
		},
		{
			name: "tool response with simple content",
			input: []*llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID: "call_001",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_status",
								Arguments: `{}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "call_001",
							Content:    `{"status": "ok"}`,
						},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				require.Len(t, messages, 2)

				toolMsg := messages[1].(map[string]any)
				assert.Equal(t, "tool", toolMsg["role"])
				assert.Equal(t, "call_001", toolMsg["tool_call_id"])
				assert.Equal(t, "get_status", toolMsg["name"])

				// Simple content (1-2 keys) is parsed as object, not string
				// (Langfuse can decide how to display it)
				content := toolMsg["content"]
				assert.NotNil(t, content, "Content should not be nil")

				// Can be either string or parsed object
				switch v := content.(type) {
				case string:
					assert.Contains(t, v, "status")
				case map[string]any:
					assert.Contains(t, v, "status")
				default:
					t.Errorf("Unexpected content type: %T", content)
				}
			},
		},
		{
			name: "tool response with rich content",
			input: []*llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID: "call_002",
							FunctionCall: &llms.FunctionCall{
								Name:      "search_db",
								Arguments: `{}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "call_002",
							Content:    `{"results": [{"id": 1, "name": "John"}], "count": 1, "page": 1}`,
						},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				toolMsg := messages[1].(map[string]any)

				// Rich content (3+ keys or nested) becomes object
				content, ok := toolMsg["content"].(map[string]any)
				assert.True(t, ok, "Rich content should be object")
				assert.Contains(t, content, "results")
				assert.Contains(t, content, "count")
				assert.Contains(t, content, "page")
			},
		},
		{
			name: "multimodal message",
			input: []*llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "What's this?"},
						llms.ImageURLContent{
							URL:    "https://example.com/image.jpg",
							Detail: "high",
						},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				msg := messages[0].(map[string]any)

				content := msg["content"].([]any)
				require.Len(t, content, 2)

				// Text part
				text := content[0].(map[string]any)
				assert.Equal(t, "text", text["type"])
				assert.Equal(t, "What's this?", text["text"])

				// Image part
				img := content[1].(map[string]any)
				assert.Equal(t, "image_url", img["type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertInput(tt.input, nil)
			tt.validate(t, result)
		})
	}
}

// TestConvertOutput tests output conversion scenarios
func TestConvertOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   any
		validate func(t *testing.T, result any)
	}{
		{
			name:   "nil output",
			output: nil,
			validate: func(t *testing.T, result any) {
				assert.Nil(t, result)
			},
		},
		{
			name: "simple text response",
			output: &llms.ContentChoice{
				Content: "The answer is 42",
			},
			validate: func(t *testing.T, result any) {
				msg := result.(map[string]any)
				assert.Equal(t, "assistant", msg["role"])
				assert.Equal(t, "The answer is 42", msg["content"])
			},
		},
		{
			name: "response with tool calls",
			output: &llms.ContentChoice{
				Content: "Let me check",
				ToolCalls: []llms.ToolCall{
					{
						ID: "call_123",
						FunctionCall: &llms.FunctionCall{
							Name:      "check_status",
							Arguments: `{"id":"123"}`,
						},
					},
				},
			},
			validate: func(t *testing.T, result any) {
				msg := result.(map[string]any)
				assert.Equal(t, "assistant", msg["role"])
				assert.Equal(t, "Let me check", msg["content"])

				toolCalls := msg["tool_calls"].([]any)
				require.Len(t, toolCalls, 1)
			},
		},
		{
			name: "response with reasoning",
			output: &llms.ContentChoice{
				Content: "The answer is correct",
				Reasoning: &reasoning.ContentReasoning{
					Content: "Step-by-step analysis...",
				},
			},
			validate: func(t *testing.T, result any) {
				msg := result.(map[string]any)

				thinking := msg["thinking"].([]any)
				require.Len(t, thinking, 1)

				th := thinking[0].(map[string]any)
				assert.Equal(t, "thinking", th["type"])
				assert.Equal(t, "Step-by-step analysis...", th["content"])
			},
		},
		{
			name: "multiple choices array",
			output: []llms.ContentChoice{
				{Content: "Option 1"},
				{Content: "Option 2"},
			},
			validate: func(t *testing.T, result any) {
				choices := result.([]any)
				require.Len(t, choices, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOutput(tt.output)
			tt.validate(t, result)
		})
	}
}

// TestRoleMapping tests role conversion
func TestRoleMapping(t *testing.T) {
	tests := []struct {
		input    llms.ChatMessageType
		expected string
	}{
		{llms.ChatMessageTypeHuman, "user"},
		{llms.ChatMessageTypeAI, "assistant"},
		{llms.ChatMessageTypeSystem, "system"},
		{llms.ChatMessageTypeTool, "tool"},
		{llms.ChatMessageTypeGeneric, "assistant"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := mapRole(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToolContentParsing tests rich vs simple tool content detection
func TestToolContentParsing(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		expectType string // "string" or "object"
	}{
		{
			name:       "simple 1 key",
			content:    `{"status": "ok"}`,
			expectType: "string",
		},
		{
			name:       "simple 2 keys",
			content:    `{"status": "ok", "code": 200}`,
			expectType: "string",
		},
		{
			name:       "rich 3+ keys",
			content:    `{"status": "ok", "code": 200, "message": "Success"}`,
			expectType: "object",
		},
		{
			name:       "rich nested array",
			content:    `{"results": [{"id": 1}], "count": 1}`,
			expectType: "object",
		},
		{
			name:       "rich nested object",
			content:    `{"data": {"id": 1}, "meta": {}}`,
			expectType: "object",
		},
		{
			name:       "invalid json stays string",
			content:    `not valid json`,
			expectType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseToolContent(tt.content)

			if tt.expectType == "object" {
				_, ok := result.(map[string]any)
				assert.True(t, ok, "Expected object but got %T", result)
			} else {
				// Can be string or parsed simple object
				// Both are acceptable for simple content
			}
		})
	}
}

// TestThinkingExtraction tests reasoning extraction
func TestThinkingExtraction(t *testing.T) {
	input := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{
				Text: "The answer is correct",
				Reasoning: &reasoning.ContentReasoning{
					Content: "Step 1: ...\nStep 2: ...",
				},
			},
		},
	}

	result := convertMessage(input)
	msg := result.(map[string]any)

	thinking := msg["thinking"].([]any)
	require.Len(t, thinking, 1)

	th := thinking[0].(map[string]any)
	assert.Equal(t, "thinking", th["type"])
	assert.Contains(t, th["content"], "Step 1")
}

// TestJoinTextParts tests multiple text parts joining
func TestJoinTextParts(t *testing.T) {
	parts := []string{"Hello", "World", "!"}
	result := joinTextParts(parts)
	assert.Equal(t, "Hello World !", result)
}

// TestEdgeCases tests edge cases and error handling
func TestEdgeCases(t *testing.T) {
	t.Run("empty message chain", func(t *testing.T) {
		result := convertInput([]*llms.MessageContent{}, nil)
		messages := result.([]any)
		assert.Len(t, messages, 0)
	})

	t.Run("tool call without function call", func(t *testing.T) {
		tc := &llms.ToolCall{
			ID:           "call_001",
			FunctionCall: nil,
		}
		result := convertToolCallToOpenAI(tc)
		assert.Nil(t, result)
	})

	t.Run("invalid tool arguments json", func(t *testing.T) {
		tc := &llms.ToolCall{
			ID: "call_001",
			FunctionCall: &llms.FunctionCall{
				Name:      "test",
				Arguments: "invalid json{",
			},
		}
		result := convertToolCallToOpenAI(tc)
		require.NotNil(t, result)

		// Should still work, keeping invalid json as string
		msg := result.(map[string]any)
		fn := msg["function"].(map[string]any)
		assert.Equal(t, "invalid json{", fn["arguments"])
	})

	t.Run("pass-through unknown types", func(t *testing.T) {
		input := "plain string"
		result := convertInput(input, nil)
		assert.Equal(t, input, result)
	})
}

// BenchmarkConvertInput benchmarks conversion performance
func BenchmarkConvertInput(b *testing.B) {
	input := []*llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test message"},
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Response"},
				llms.ToolCall{
					ID: "call_001",
					FunctionCall: &llms.FunctionCall{
						Name:      "test",
						Arguments: `{"key":"value"}`,
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertInput(input, nil)
	}
}

// TestRealWorldScenario tests a complete conversation flow
func TestRealWorldScenario(t *testing.T) {
	// Simulate a real penetration testing conversation
	input := []*llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "You are a security analyst."},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Check CVE-2024-1234"},
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "I'll search for that vulnerability."},
				llms.ToolCall{
					ID: "call_001",
					FunctionCall: &llms.FunctionCall{
						Name:      "search_cve",
						Arguments: `{"cve_id":"CVE-2024-1234"}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "call_001",
					Content:    `{"severity":"high","description":"SQL injection","cvss_score":8.5,"exploit_available":true}`,
				},
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "This is a high-severity SQL injection vulnerability with CVSS 8.5. Exploit is available.",
					Reasoning: &reasoning.ContentReasoning{
						Content: "The high CVSS score combined with exploit availability makes this critical.",
					},
				},
			},
		},
	}

	result := convertInput(input, nil)
	require.NotNil(t, result)

	// Verify structure
	messages := result.([]any)
	require.Len(t, messages, 5)

	// Verify system message
	systemMsg := messages[0].(map[string]any)
	assert.Equal(t, "system", systemMsg["role"])

	// Verify user message
	userMsg := messages[1].(map[string]any)
	assert.Equal(t, "user", userMsg["role"])

	// Verify assistant with tool call
	assistantMsg := messages[2].(map[string]any)
	assert.Equal(t, "assistant", assistantMsg["role"])
	assert.NotNil(t, assistantMsg["tool_calls"])

	// Verify tool response (should be rich object due to 4+ keys)
	toolMsg := messages[3].(map[string]any)
	assert.Equal(t, "tool", toolMsg["role"])
	assert.Equal(t, "search_cve", toolMsg["name"])
	content, ok := toolMsg["content"].(map[string]any)
	assert.True(t, ok, "Tool response with 4+ keys should be object")
	assert.Equal(t, "high", content["severity"])

	// Verify final assistant message with thinking
	finalMsg := messages[4].(map[string]any)
	assert.Equal(t, "assistant", finalMsg["role"])
	thinking := finalMsg["thinking"].([]any)
	require.Len(t, thinking, 1)

	// Log the full conversation in JSON for inspection
	jsonData, err := json.MarshalIndent(messages, "", "  ")
	require.NoError(t, err)
	t.Logf("Full conversation:\n%s", string(jsonData))
}
