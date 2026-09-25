package langfuse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestConverter_ConvertInput_TranslatesAChainToOpenAIMessages(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		validate func(t *testing.T, result any)
	}{
		{
			name:     "a nil input stays nil",
			input:    nil,
			validate: func(t *testing.T, result any) { assert.Nil(t, result) },
		},
		{
			name:     "a value of another type passes through",
			input:    "plain string",
			validate: func(t *testing.T, result any) { assert.Equal(t, "plain string", result) },
		},
		{
			name:     "an empty chain",
			input:    []*llms.MessageContent{},
			validate: func(t *testing.T, result any) { assert.Equal(t, []any{}, result) },
		},
		{
			name: "simple text message",
			input: []*llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Hello"}}},
			},
			validate: func(t *testing.T, result any) {
				msg := result.([]any)[0].(map[string]any)
				assert.Equal(t, "user", msg["role"])
				assert.Equal(t, "Hello", msg["content"])
			},
		},
		{
			name: "several text parts are joined with spaces",
			input: []*llms.MessageContent{{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextContent{Text: "Hello"}, llms.TextContent{Text: "World"}, llms.TextContent{Text: "!"}},
			}},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, "Hello World !", result.([]any)[0].(map[string]any)["content"])
			},
		},
		{
			name: "message with tools",
			input: []*llms.MessageContent{{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Let me search"},
					llms.ToolCall{ID: "call_001", FunctionCall: &llms.FunctionCall{Name: "search", Arguments: `{"query":"test"}`}},
				},
			}},
			validate: func(t *testing.T, result any) {
				msg := result.([]any)[0].(map[string]any)
				assert.Equal(t, "assistant", msg["role"])
				assert.Equal(t, "Let me search", msg["content"])
				assert.Equal(t, []any{map[string]any{
					"id":       "call_001",
					"type":     "function",
					"function": map[string]any{"name": "search", "arguments": `{"query":"test"}`},
				}}, msg["tool_calls"])
			},
		},
		{
			name: "tool response with simple content",
			input: []*llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.ToolCall{ID: "call_001", FunctionCall: &llms.FunctionCall{Name: "get_status", Arguments: `{}`}}},
				},
				{
					Role:  llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: "call_001", Content: `{"status": "ok"}`}},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				require.Len(t, messages, 2)
				assert.Equal(t, map[string]any{
					"role":         "tool",
					"tool_call_id": "call_001",
					"name":         "get_status",
					"content":      map[string]any{"status": "ok"},
				}, messages[1])
			},
		},
		{
			name: "tool response with rich content",
			input: []*llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.ToolCall{ID: "call_002", FunctionCall: &llms.FunctionCall{Name: "search_db", Arguments: `{}`}}},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{
						ToolCallID: "call_002",
						Content:    `{"results": [{"id": 1, "name": "John"}], "count": 1, "page": 1}`,
					}},
				},
			},
			validate: func(t *testing.T, result any) {
				content, ok := result.([]any)[1].(map[string]any)["content"].(map[string]any)
				require.True(t, ok, "rich content should be an object")
				assert.Contains(t, content, "results")
				assert.Contains(t, content, "count")
				assert.Contains(t, content, "page")
			},
		},
		{
			name: "multimodal message",
			input: []*llms.MessageContent{{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "What's this?"},
					llms.ImageURLContent{URL: "https://example.com/image.jpg", Detail: "high"},
				},
			}},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, []any{
					map[string]any{"type": "text", "text": "What's this?"},
					map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.com/image.jpg", "detail": "high"}},
				}, result.([]any)[0].(map[string]any)["content"])
			},
		},
		{
			name: "a whole conversation keeps every role, the tool name and the thinking",
			input: []*llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: "You are a security analyst."}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "Check CVE-2024-1234"}}},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "I'll search for that vulnerability."},
						llms.ToolCall{ID: "call_001", FunctionCall: &llms.FunctionCall{Name: "search_cve", Arguments: `{"cve_id":"CVE-2024-1234"}`}},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{
						ToolCallID: "call_001",
						Content:    `{"severity":"high","description":"SQL injection","cvss_score":8.5,"exploit_available":true}`,
					}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextContent{
						Text:      "This is a high-severity SQL injection vulnerability with CVSS 8.5. Exploit is available.",
						Reasoning: &reasoning.ContentReasoning{Content: "The high CVSS score combined with exploit availability makes this critical."},
					}},
				},
			},
			validate: func(t *testing.T, result any) {
				messages := result.([]any)
				require.Len(t, messages, 5)

				roles := make([]any, len(messages))
				for i, message := range messages {
					roles[i] = message.(map[string]any)["role"]
				}
				assert.Equal(t, []any{"system", "user", "assistant", "tool", "assistant"}, roles)
				assert.Equal(t, []any{map[string]any{
					"id":       "call_001",
					"type":     "function",
					"function": map[string]any{"name": "search_cve", "arguments": `{"cve_id":"CVE-2024-1234"}`},
				}}, messages[2].(map[string]any)["tool_calls"])

				toolMsg := messages[3].(map[string]any)
				assert.Equal(t, "search_cve", toolMsg["name"])
				assert.Equal(t, "high", toolMsg["content"].(map[string]any)["severity"])

				assert.Equal(t, []any{map[string]any{
					"type":    "thinking",
					"content": "The high CVSS score combined with exploit availability makes this critical.",
				}}, messages[4].(map[string]any)["thinking"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, convertInput(tt.input, nil))
		})
	}
}

func TestConverter_ConvertOutput_TranslatesChoicesAndMessages(t *testing.T) {
	tests := []struct {
		name     string
		output   any
		validate func(t *testing.T, result any)
	}{
		{
			name:     "a nil output stays nil",
			output:   nil,
			validate: func(t *testing.T, result any) { assert.Nil(t, result) },
		},
		{
			name:   "simple text response",
			output: &llms.ContentChoice{Content: "The answer is 42"},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, map[string]any{"role": "assistant", "content": "The answer is 42"}, result)
			},
		},
		{
			name: "response with tool calls",
			output: &llms.ContentChoice{
				Content:   "Let me check",
				ToolCalls: []llms.ToolCall{{ID: "call_123", FunctionCall: &llms.FunctionCall{Name: "check_status", Arguments: `{"id":"123"}`}}},
			},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, map[string]any{
					"role":    "assistant",
					"content": "Let me check",
					"tool_calls": []any{map[string]any{
						"id":       "call_123",
						"type":     "function",
						"function": map[string]any{"name": "check_status", "arguments": `{"id":"123"}`},
					}},
				}, result)
			},
		},
		{
			name: "response with reasoning",
			output: &llms.ContentChoice{
				Content:   "The answer is correct",
				Reasoning: &reasoning.ContentReasoning{Content: "Step-by-step analysis..."},
			},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, []any{map[string]any{"type": "thinking", "content": "Step-by-step analysis..."}}, result.(map[string]any)["thinking"])
			},
		},
		{
			name: "a message keeps its thinking",
			output: &llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextContent{
					Text:      "The answer is correct",
					Reasoning: &reasoning.ContentReasoning{Content: "Step 1: ...\nStep 2: ..."},
				}},
			},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, []any{map[string]any{"type": "thinking", "content": "Step 1: ...\nStep 2: ..."}}, result.(map[string]any)["thinking"])
			},
		},
		{
			name:   "multiple choices array",
			output: []llms.ContentChoice{{Content: "Option 1"}, {Content: "Option 2"}},
			validate: func(t *testing.T, result any) {
				assert.Equal(t, []any{
					map[string]any{"role": "assistant", "content": "Option 1"},
					map[string]any{"role": "assistant", "content": "Option 2"},
				}, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, convertOutput(tt.output))
		})
	}
}

func TestConverter_MapRole_FollowsOpenAIRoleNames(t *testing.T) {
	tests := []struct {
		name     string
		input    llms.ChatMessageType
		expected string
	}{
		{"human becomes user", llms.ChatMessageTypeHuman, "user"},
		{"ai becomes assistant", llms.ChatMessageTypeAI, "assistant"},
		{"system stays system", llms.ChatMessageTypeSystem, "system"},
		{"tool stays tool", llms.ChatMessageTypeTool, "tool"},
		{"generic falls back to assistant", llms.ChatMessageTypeGeneric, "assistant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapRole(tt.input))
		})
	}
}

func TestConverter_ParseToolContent_ParsesJSONAndKeepsTheRestAsText(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    any
	}{
		{name: "a one-key object", content: `{"status": "ok"}`, want: map[string]any{"status": "ok"}},
		{
			name:    "a three-key object",
			content: `{"status": "ok", "code": 200, "message": "Success"}`,
			want:    map[string]any{"status": "ok", "code": float64(200), "message": "Success"},
		},
		{
			name:    "a nested array",
			content: `{"results": [{"id": 1}], "count": 1}`,
			want:    map[string]any{"results": []any{map[string]any{"id": float64(1)}}, "count": float64(1)},
		},
		{name: "invalid json stays a string", content: `not valid json`, want: "not valid json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseToolContent(tt.content))
		})
	}
}

func TestConverter_ConvertToolCallToOpenAI_DropsCallsWithoutAFunction(t *testing.T) {
	assert.Nil(t, convertToolCallToOpenAI(&llms.ToolCall{ID: "call_001"}))

	result := convertToolCallToOpenAI(&llms.ToolCall{
		ID:           "call_001",
		FunctionCall: &llms.FunctionCall{Name: "test", Arguments: "invalid json{"},
	})
	require.NotNil(t, result)
	assert.Equal(t, "invalid json{", result.(map[string]any)["function"].(map[string]any)["arguments"], "arguments that are not json are kept verbatim")
}

func BenchmarkConvertInput(b *testing.B) {
	input := []*llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Test message"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Response"},
				llms.ToolCall{ID: "call_001", FunctionCall: &llms.FunctionCall{Name: "test", Arguments: `{"key":"value"}`}},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertInput(input, nil)
	}
}
