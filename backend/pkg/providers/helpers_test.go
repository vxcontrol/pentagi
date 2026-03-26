package providers

import (
	"encoding/json"
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"pentagi/pkg/cast"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

var (
	// basicChain represents a basic conversation with system, human, and AI messages
	basicChain = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hello, how are you?"}},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "I'm doing well! How can I help you today?"}},
		},
	}

	// chainStartingWithHuman represents a conversation without system message, starting with human
	chainStartingWithHuman = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hey there, can you help me?"}},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Of course! What do you need assistance with?"}},
		},
	}

	// chainWithOnlySystem represents a conversation with only a system message
	chainWithOnlySystem = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are an AI assistant that provides helpful information."}},
		},
	}

	// chainWithToolCall represents a conversation where AI called a tool
	chainWithToolCall = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather like?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "tool-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
	}

	// chainWithToolResponse represents a conversation with a tool call that has been responded to
	chainWithToolResponse = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather like?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "tool-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "tool-1",
					Name:       "get_weather",
					Content:    "The weather in New York is sunny with a high of 75°F.",
				},
			},
		},
	}

	// chainWithMultipleToolCalls represents a conversation with multiple tool calls
	chainWithMultipleToolCalls = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather and time in New York?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "tool-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
				llms.ToolCall{
					ID:   "tool-2",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_time",
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
	}

	// incompleteChainWithMultipleToolCalls represents a conversation with multiple tool calls where only one has a response
	incompleteChainWithMultipleToolCalls = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather and time in New York?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "tool-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
				llms.ToolCall{
					ID:   "tool-2",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_time",
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "tool-1",
					Name:       "get_weather",
					Content:    "The weather in New York is sunny with a high of 75°F.",
				},
			},
		},
	}
)

// Helper to clone a chain to avoid modifying test fixtures
func cloneChain(chain []llms.MessageContent) []llms.MessageContent {
	b, _ := json.Marshal(chain)
	var cloned []llms.MessageContent
	_ = json.Unmarshal(b, &cloned)
	return cloned
}

func newFlowProvider() *flowProvider {
	return &flowProvider{
		mx:              &sync.RWMutex{},
		callCounter:     &atomic.Int64{},
		maxGACallsLimit: maxGeneralAgentChainIterations,
		maxLACallsLimit: maxLimitedAgentChainIterations,
		buildMonitor: func() *executionMonitor {
			return &executionMonitor{
				enabled: false,
			}
		},
	}
}

func findUnrespondedToolCalls(chain []llms.MessageContent) ([]llms.ToolCall, error) {
	if len(chain) == 0 {
		return nil, nil
	}

	var lastAIMsg llms.MessageContent
	var lastAIMsgIdx int
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].Role == llms.ChatMessageTypeAI {
			lastAIMsg = chain[i]
			lastAIMsgIdx = i
			break
		}
	}

	if lastAIMsg.Role != llms.ChatMessageTypeAI {
		return nil, nil // No AI message found
	}

	var toolCalls []llms.ToolCall
	for _, part := range lastAIMsg.Parts {
		toolCall, ok := part.(llms.ToolCall)
		if !ok || toolCall.FunctionCall == nil {
			continue
		}
		toolCalls = append(toolCalls, toolCall)
	}

	if len(toolCalls) == 0 {
		return nil, nil // No tool calls in the AI message
	}

	respondedToolCalls := make(map[string]bool)
	for i := lastAIMsgIdx + 1; i < len(chain); i++ {
		if chain[i].Role != llms.ChatMessageTypeTool {
			continue
		}

		for _, part := range chain[i].Parts {
			toolResponse, ok := part.(llms.ToolCallResponse)
			if !ok {
				continue
			}
			respondedToolCalls[toolResponse.ToolCallID] = true
		}
	}

	var unrespondedToolCalls []llms.ToolCall
	for _, toolCall := range toolCalls {
		if !respondedToolCalls[toolCall.ID] {
			unrespondedToolCalls = append(unrespondedToolCalls, toolCall)
		}
	}

	return unrespondedToolCalls, nil
}

func TestUpdateMsgChainResult(t *testing.T) {
	provider := newFlowProvider()

	tests := []struct {
		name          string
		chain         []llms.MessageContent
		toolName      string
		input         string
		expectedChain []llms.MessageContent
		expectError   bool
	}{
		{
			name:        "Empty chain",
			chain:       []llms.MessageContent{},
			toolName:    "ask_user",
			input:       "Hello!",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Hello!"}},
				},
			},
		},
		{
			name:        "System message as last message",
			chain:       cloneChain(basicChain)[:1], // Just the system message
			toolName:    "ask_user",
			input:       "Hello!",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Hello!"}},
				},
			},
		},
		{
			name:        "Human message as last message",
			chain:       cloneChain(basicChain)[:2], // System + Human
			toolName:    "ask_user",
			input:       " I need help with my code.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello, how are you?"},
						llms.TextContent{Text: " I need help with my code."},
					},
				},
			},
		},
		{
			name:        "AI message as last message",
			chain:       cloneChain(basicChain), // Full basic chain
			toolName:    "ask_user",
			input:       "I need help with my code.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Hello, how are you?"}},
				},
				{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextContent{Text: "I'm doing well! How can I help you today?"}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "I need help with my code."}},
				},
			},
		},
		{
			name:        "Tool call without response",
			chain:       cloneChain(chainWithToolCall),
			toolName:    "get_weather",
			input:       "The weather in New York is sunny with a high of 75°F.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather like?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    "The weather in New York is sunny with a high of 75°F.",
						},
					},
				},
			},
		},
		{
			name:        "Tool call with wrong tool name",
			chain:       cloneChain(chainWithToolCall),
			toolName:    "wrong_tool",
			input:       "This is a response to a wrong tool.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather like?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    cast.FallbackResponseContent,
						},
					},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "This is a response to a wrong tool."}},
				},
			},
		},
		{
			name:        "Update existing tool response",
			chain:       cloneChain(chainWithToolResponse),
			toolName:    "get_weather",
			input:       "Updated: The weather in New York is rainy.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather like?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    "Updated: The weather in New York is rainy.",
						},
					},
				},
			},
		},
		{
			name:        "Multiple tool calls - respond to first tool",
			chain:       cloneChain(chainWithMultipleToolCalls),
			toolName:    "get_weather",
			input:       "The weather in New York is sunny with a high of 75°F.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather and time in New York?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
						llms.ToolCall{
							ID:   "tool-2",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_time",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    "The weather in New York is sunny with a high of 75°F.",
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-2",
							Name:       "get_time",
							Content:    cast.FallbackResponseContent,
						},
					},
				},
			},
		},
		{
			name:        "Multiple tool calls - respond to second tool",
			chain:       cloneChain(chainWithMultipleToolCalls),
			toolName:    "get_time",
			input:       "The current time in New York is 3:45 PM.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather and time in New York?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
						llms.ToolCall{
							ID:   "tool-2",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_time",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    cast.FallbackResponseContent,
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-2",
							Name:       "get_time",
							Content:    "The current time in New York is 3:45 PM.",
						},
					},
				},
			},
		},
		{
			name:        "Tool message as last message without matching tool",
			chain:       cloneChain(incompleteChainWithMultipleToolCalls),
			toolName:    "ask_user",
			input:       "I want to know more about the weather there.",
			expectError: false,
			expectedChain: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "What's the weather and time in New York?"}},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "tool-1",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
						llms.ToolCall{
							ID:   "tool-2",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_time",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-1",
							Name:       "get_weather",
							Content:    "The weather in New York is sunny with a high of 75°F.",
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "tool-2",
							Name:       "get_time",
							Content:    cast.FallbackResponseContent,
						},
					},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "I want to know more about the weather there."}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultChain, err := provider.updateMsgChainResult(tt.chain, tt.toolName, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedChain, resultChain)
			}
		})
	}
}

func TestFindUnrespondedToolCalls(t *testing.T) {
	tests := []struct {
		name          string
		chain         []llms.MessageContent
		expectedCalls int
		expectedNames []string
		expectedHasAI bool
	}{
		{
			name:          "Empty chain",
			chain:         []llms.MessageContent{},
			expectedCalls: 0,
			expectedHasAI: false,
		},
		{
			name:          "Chain without AI message",
			chain:         cloneChain(basicChain)[:2], // System + Human
			expectedCalls: 0,
			expectedHasAI: false,
		},
		{
			name:          "Chain with AI message but no tool calls",
			chain:         cloneChain(basicChain),
			expectedCalls: 0,
			expectedHasAI: true,
		},
		{
			name:          "Chain with tool call but no response",
			chain:         cloneChain(chainWithToolCall),
			expectedCalls: 1,
			expectedNames: []string{"get_weather"},
			expectedHasAI: true,
		},
		{
			name:          "Chain with tool call and response",
			chain:         cloneChain(chainWithToolResponse),
			expectedCalls: 0,
			expectedHasAI: true,
		},
		{
			name:          "Chain with multiple tool calls and no responses",
			chain:         cloneChain(chainWithMultipleToolCalls),
			expectedCalls: 2,
			expectedNames: []string{"get_weather", "get_time"},
			expectedHasAI: true,
		},
		{
			name:          "Chain with multiple tool calls and one response",
			chain:         cloneChain(incompleteChainWithMultipleToolCalls),
			expectedCalls: 1,
			expectedNames: []string{"get_time"},
			expectedHasAI: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolCalls, err := findUnrespondedToolCalls(tt.chain)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCalls, len(toolCalls))

			if tt.expectedCalls > 0 {
				foundNames := make([]string, 0, len(toolCalls))
				for _, call := range toolCalls {
					if call.FunctionCall != nil {
						foundNames = append(foundNames, call.FunctionCall.Name)
					}
				}

				for _, expectedName := range tt.expectedNames {
					found := slices.Contains(foundNames, expectedName)
					assert.True(t, found, "Expected to find tool call named '%s'", expectedName)
				}
			}
		})
	}
}

func TestEnsureChainConsistency(t *testing.T) {
	provider := newFlowProvider()

	tests := []struct {
		name             string
		chain            []llms.MessageContent
		expectedAdded    int
		expectConsistent bool
	}{
		{
			name:             "Empty chain",
			chain:            []llms.MessageContent{},
			expectedAdded:    0,
			expectConsistent: true,
		},
		{
			name:             "Already consistent chain",
			chain:            cloneChain(basicChain),
			expectedAdded:    0,
			expectConsistent: true,
		},
		{
			name:             "Chain with tool call but no response",
			chain:            cloneChain(chainWithToolCall),
			expectedAdded:    1,
			expectConsistent: false,
		},
		{
			name:             "Chain with multiple tool calls and no responses",
			chain:            cloneChain(chainWithMultipleToolCalls),
			expectedAdded:    2,
			expectConsistent: false,
		},
		{
			name:             "Chain with multiple tool calls and one response",
			chain:            cloneChain(incompleteChainWithMultipleToolCalls),
			expectedAdded:    1,
			expectConsistent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLen := len(tt.chain)
			resultChain, err := provider.ensureChainConsistency(tt.chain)

			assert.NoError(t, err)
			assert.Equal(t, originalLen+tt.expectedAdded, len(resultChain))

			// Verify all tool calls have responses
			if tt.expectedAdded > 0 {
				// Find all tool calls
				toolCalls, err := findUnrespondedToolCalls(resultChain)
				assert.NoError(t, err)
				assert.Empty(t, toolCalls, "There should be no unresponded tool calls after ensuring consistency")

				// Check the last messages are tool responses with the default content
				if !tt.expectConsistent {
					for i := range tt.expectedAdded {
						idx := originalLen + i
						assert.Equal(t, llms.ChatMessageTypeTool, resultChain[idx].Role)

						for _, part := range resultChain[idx].Parts {
							if resp, ok := part.(llms.ToolCallResponse); ok {
								assert.Equal(t, cast.FallbackResponseContent, resp.Content)
							}
						}
					}
				}
			}
		})
	}
}

// makeToolCall is a helper to create a ToolCall with the given function name and arguments.
func makeToolCall(name, args string) llms.ToolCall {
	return llms.ToolCall{
		ID:   "test-id",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      name,
			Arguments: args,
		},
	}
}

// maxSoftDetectionsBeforeAbort mirrors the constant in performer.go.
// Keep in sync with performer.go:maxSoftDetectionsBeforeAbort.
const testMaxSoftDetectionsBeforeAbort = 4

func TestRepeatingDetector(t *testing.T) {
	tests := []struct {
		name             string
		calls            []llms.ToolCall
		expectedDetected []bool // expected detect() return for each call
		expectedLen      int    // expected len(funcCalls) after all calls
	}{
		{
			name: "nil function call returns false",
			calls: []llms.ToolCall{
				{ID: "test", Type: "function", FunctionCall: nil},
			},
			expectedDetected: []bool{false},
			expectedLen:      0,
		},
		{
			name: "first call returns false",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test"}`),
			},
			expectedDetected: []bool{false},
			expectedLen:      1,
		},
		{
			name: "two identical calls below threshold",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test"}`),
				makeToolCall("search", `{"query":"test"}`),
			},
			expectedDetected: []bool{false, false},
			expectedLen:      2,
		},
		{
			name: "three identical calls triggers detection",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test"}`),
				makeToolCall("search", `{"query":"test"}`),
				makeToolCall("search", `{"query":"test"}`),
			},
			expectedDetected: []bool{false, false, true},
			expectedLen:      3,
		},
		{
			name: "different call resets funcCalls",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test"}`),
				makeToolCall("search", `{"query":"test"}`),
				makeToolCall("browse", `{"url":"http://example.com"}`),
			},
			expectedDetected: []bool{false, false, false},
			expectedLen:      1,
		},
		{
			name: "same name different args resets funcCalls",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test","limit":"10"}`),
				makeToolCall("search", `{"query":"test","limit":"10"}`),
				makeToolCall("search", `{"query":"different","limit":"20"}`),
			},
			expectedDetected: []bool{false, false, false},
			expectedLen:      1,
		},
		{
			name: "six identical calls still below escalation threshold",
			calls: func() []llms.ToolCall {
				tc := makeToolCall("search", `{"query":"test"}`)
				return []llms.ToolCall{tc, tc, tc, tc, tc, tc}
			}(),
			expectedDetected: []bool{false, false, true, true, true, true},
			expectedLen:      6,
		},
		{
			name: "seven identical calls reaches escalation threshold",
			calls: func() []llms.ToolCall {
				tc := makeToolCall("search", `{"query":"test"}`)
				return []llms.ToolCall{tc, tc, tc, tc, tc, tc, tc}
			}(),
			expectedDetected: []bool{false, false, true, true, true, true, true},
			expectedLen:      7,
		},
		{
			name: "message field stripped treats calls as identical",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test","message":"first attempt"}`),
				makeToolCall("search", `{"query":"test","message":"second attempt"}`),
				makeToolCall("search", `{"query":"test","message":"third attempt"}`),
			},
			expectedDetected: []bool{false, false, true},
			expectedLen:      3,
		},
		{
			name: "different JSON key order treated as identical",
			calls: []llms.ToolCall{
				makeToolCall("search", `{"query":"test","limit":"10"}`),
				makeToolCall("search", `{"limit":"10","query":"test"}`),
				makeToolCall("search", `{"query":"test","limit":"10"}`),
			},
			expectedDetected: []bool{false, false, true},
			expectedLen:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &repeatingDetector{}

			for i, call := range tt.calls {
				detected := detector.detect(call)
				assert.Equal(t, tt.expectedDetected[i], detected,
					"call %d: expected detect=%v, got %v", i, tt.expectedDetected[i], detected)
			}

			assert.Equal(t, tt.expectedLen, len(detector.funcCalls),
				"expected funcCalls length %d, got %d", tt.expectedLen, len(detector.funcCalls))
		})
	}
}

func TestRepeatingDetectorEscalationThreshold(t *testing.T) {
	// This test validates the escalation math used in performer.go:
	// len(detector.funcCalls) >= RepeatingToolCallThreshold + maxSoftDetectionsBeforeAbort
	// With threshold=3 and maxSoftDetections=4, abort triggers at len >= 7

	detector := &repeatingDetector{}
	tc := makeToolCall("search", `{"query":"test"}`)

	for i := 0; i < 7; i++ {
		detector.detect(tc)
	}

	assert.Equal(t, 7, len(detector.funcCalls))
	assert.True(t, len(detector.funcCalls) >= RepeatingToolCallThreshold+testMaxSoftDetectionsBeforeAbort,
		"7 calls should reach escalation threshold: %d >= %d+%d",
		len(detector.funcCalls), RepeatingToolCallThreshold, testMaxSoftDetectionsBeforeAbort)

	// Verify 6 calls is below threshold
	detector2 := &repeatingDetector{}
	for i := 0; i < 6; i++ {
		detector2.detect(tc)
	}

	assert.Equal(t, 6, len(detector2.funcCalls))
	assert.False(t, len(detector2.funcCalls) >= RepeatingToolCallThreshold+testMaxSoftDetectionsBeforeAbort,
		"6 calls should NOT reach escalation threshold: %d < %d+%d",
		len(detector2.funcCalls), RepeatingToolCallThreshold, 4)
}

func TestClearCallArguments(t *testing.T) {
	tests := []struct {
		name         string
		input        llms.FunctionCall
		expectedName string
		expectedArgs string
	}{
		{
			name: "strips message field",
			input: llms.FunctionCall{
				Name:      "search",
				Arguments: `{"cmd":"ls","message":"please run this"}`,
			},
			expectedName: "search",
			expectedArgs: "cmd: ls\n",
		},
		{
			name: "sorts keys alphabetically",
			input: llms.FunctionCall{
				Name:      "execute",
				Arguments: `{"z_param":"1","a_param":"2","m_param":"3"}`,
			},
			expectedName: "execute",
			expectedArgs: "a_param: 2\nm_param: 3\nz_param: 1\n",
		},
		{
			name: "invalid JSON returns original",
			input: llms.FunctionCall{
				Name:      "search",
				Arguments: "not valid json",
			},
			expectedName: "search",
			expectedArgs: "not valid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &repeatingDetector{}
			result := detector.clearCallArguments(&tt.input)

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.expectedArgs, result.Arguments)
		})
	}
}

func TestExecutionMonitorDetector_ShouldInvokeAdviser(t *testing.T) {
	tests := []struct {
		name      string
		threshold struct{ same, total int }
		calls     []string
		expected  []bool
	}{
		{
			name:      "trigger on same tool limit",
			threshold: struct{ same, total int }{5, 10},
			calls:     []string{"tool1", "tool1", "tool1", "tool1", "tool1"},
			expected:  []bool{false, false, false, false, true},
		},
		{
			name:      "trigger on total tool limit",
			threshold: struct{ same, total int }{5, 10},
			calls:     []string{"tool1", "tool2", "tool3", "tool4", "tool5", "tool6", "tool7", "tool8", "tool9", "tool10"},
			expected:  []bool{false, false, false, false, false, false, false, false, false, true},
		},
		{
			name:      "reset after different tool",
			threshold: struct{ same, total int }{3, 10},
			calls:     []string{"tool1", "tool1", "tool2", "tool1", "tool1"},
			expected:  []bool{false, false, false, false, false},
		},
		{
			name:      "mixed tools reaching total limit",
			threshold: struct{ same, total int }{5, 10},
			calls:     []string{"tool1", "tool2", "tool1", "tool3", "tool1", "tool2", "tool3", "tool4", "tool5", "tool6"},
			expected:  []bool{false, false, false, false, false, false, false, false, false, true},
		},
		{
			name:      "disabled detector",
			threshold: struct{ same, total int }{5, 10},
			calls:     []string{"tool1", "tool1", "tool1", "tool1", "tool1", "tool1"},
			expected:  []bool{false, false, false, false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emd := &executionMonitor{
				enabled:        tt.name != "disabled detector",
				sameThreshold:  tt.threshold.same,
				totalThreshold: tt.threshold.total,
			}

			for i, call := range tt.calls {
				result := emd.shouldInvokeMentor(mockToolCall(call))
				if result != tt.expected[i] {
					t.Errorf("call %d (%s): expected %v, got %v", i, call, tt.expected[i], result)
				}
			}
		})
	}
}

func TestExecutionMonitorDetector_Reset(t *testing.T) {
	emd := &executionMonitor{
		enabled:        true,
		sameThreshold:  5,
		totalThreshold: 10,
		sameToolCount:  3,
		totalCallCount: 7,
		lastToolName:   "tool1",
	}

	emd.reset()

	if emd.sameToolCount != 0 {
		t.Errorf("expected sameToolCount to be 0 after reset, got %d", emd.sameToolCount)
	}
	if emd.totalCallCount != 0 {
		t.Errorf("expected totalCallCount to be 0 after reset, got %d", emd.totalCallCount)
	}
	if emd.lastToolName != "" {
		t.Errorf("expected lastToolName to be empty after reset, got %s", emd.lastToolName)
	}
}

func TestExecutionMonitorDetector_SameToolSequence(t *testing.T) {
	emd := &executionMonitor{
		enabled:        true,
		sameThreshold:  3,
		totalThreshold: 100,
	}

	// First 2 calls should not trigger
	if emd.shouldInvokeMentor(mockToolCall("search")) {
		t.Error("first call should not trigger adviser")
	}
	if emd.shouldInvokeMentor(mockToolCall("search")) {
		t.Error("second call should not trigger adviser")
	}

	// Third call should trigger on same tool threshold
	if !emd.shouldInvokeMentor(mockToolCall("search")) {
		t.Error("third identical call should trigger adviser")
	}

	// After reset, same tool should not trigger immediately
	emd.reset()
	if emd.shouldInvokeMentor(mockToolCall("search")) {
		t.Error("first call after reset should not trigger adviser")
	}
}

func TestExecutionMonitorDetector_TotalCallsSequence(t *testing.T) {
	emd := &executionMonitor{
		enabled:        true,
		sameThreshold:  100,
		totalThreshold: 5,
	}

	tools := []string{"tool1", "tool2", "tool3", "tool4", "tool5"}

	// First 4 calls should not trigger
	for i := 0; i < 4; i++ {
		if emd.shouldInvokeMentor(mockToolCall(tools[i])) {
			t.Errorf("call %d should not trigger adviser", i)
		}
	}

	// Fifth call should trigger on total threshold
	if !emd.shouldInvokeMentor(mockToolCall(tools[4])) {
		t.Error("fifth call should trigger adviser on total threshold")
	}

	// After reset, counter should restart
	emd.reset()
	if emd.totalCallCount != 0 {
		t.Error("total count should be 0 after reset")
	}
}

func mockToolCall(name string) llms.ToolCall {
	return llms.ToolCall{FunctionCall: &llms.FunctionCall{Name: name}}
}
