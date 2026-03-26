package cast

import (
	"fmt"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
)

// Basic test fixtures - represent standard message chains in different configurations
var (
	// Empty chain
	emptyChain = []llms.MessageContent{}

	// Chain with only system message
	systemOnlyChain = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
	}

	// Chain with system and human messages
	systemHumanChain = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hello, how are you?"}},
		},
	}

	// Chain with human message only
	humanOnlyChain = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hello, can you help me?"}},
		},
	}

	// Chain with basic conversation (System, Human, AI)
	basicConversationChain = []llms.MessageContent{
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

	// Chain with tool call
	chainWithTool = []llms.MessageContent{
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

	// Chain with tool call and response
	chainWithSingleToolResponse = []llms.MessageContent{
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

	// Chain with multiple tool calls
	chainWithMultipleTools = []llms.MessageContent{
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

	// Chain with multiple tool calls and responses
	chainWithMultipleToolResponses = []llms.MessageContent{
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
					Content:    "The current time in New York is 3:45 PM.",
				},
			},
		},
	}

	// Chain with multiple sections (multiple human messages)
	chainWithMultipleSections = []llms.MessageContent{
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

	// Chain with error: consecutive human messages
	chainWithConsecutiveHumans = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hello, how are you?"}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Can you help me with something?"}},
		},
	}

	// Chain with error: missing tool response
	chainWithMissingToolResponse = []llms.MessageContent{
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

	// Chain with error: unexpected tool message (without preceding AI with tool call)
	chainWithUnexpectedTool = []llms.MessageContent{
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

	// Chain with summarization as the only body pair in a section
	chainWithSummarization = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Can you summarize the previous conversation?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "summary-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      SummarizationToolName,
						Arguments: SummarizationToolArgs,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "summary-1",
					Name:       SummarizationToolName,
					Content:    "This is a summary of the previous conversation about the weather in New York.",
				},
			},
		},
	}

	// Chain with summarization at the beginning followed by other body pairs
	chainWithSummarizationAndOtherPairs = []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Can you summarize and then tell me about the weather?"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "summary-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      SummarizationToolName,
						Arguments: SummarizationToolArgs,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "summary-1",
					Name:       SummarizationToolName,
					Content:    "This is a summary of the previous conversation.",
				},
			},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Now I'll check the weather for you."}},
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
)

// ChainConfig represents configuration options for generating a chain
type ChainConfig struct {
	// Whether to include a system message at the start
	IncludeSystem bool
	// Number of sections to include (each section has a human message)
	Sections int
	// For each section, how many body pairs to include
	BodyPairsPerSection []int
	// For each body pair, whether it should be a tool call or simple completion
	ToolsForBodyPairs []bool
	// For each tool body pair, how many tool calls to include
	ToolCallsPerBodyPair []int
	// Whether all tool calls should have responses
	IncludeAllToolResponses bool
}

// DefaultChainConfig returns a default chain configuration
func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		IncludeSystem:           true,
		Sections:                1,
		BodyPairsPerSection:     []int{1},
		ToolsForBodyPairs:       []bool{false},
		ToolCallsPerBodyPair:    []int{0},
		IncludeAllToolResponses: true,
	}
}

// GenerateChain generates a message chain based on the provided configuration
func GenerateChain(config ChainConfig) []llms.MessageContent {
	var chain []llms.MessageContent

	// Add system message if requested
	if config.IncludeSystem {
		chain = append(chain, llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		})
	}

	toolCallId := 1

	// Generate each section
	for section := 0; section < config.Sections; section++ {
		// Add human message for this section
		chain = append(chain, llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: fmt.Sprintf("Question %d", section+1)}},
		})

		// Generate body pairs for this section
		bodyPairsCount := 1
		if section < len(config.BodyPairsPerSection) {
			bodyPairsCount = config.BodyPairsPerSection[section]
		}

		for pair := 0; pair < bodyPairsCount; pair++ {
			useTool := false
			if pair < len(config.ToolsForBodyPairs) {
				useTool = config.ToolsForBodyPairs[pair]
			}

			if useTool {
				// Create AI message with tool calls
				toolCallsCount := 1
				if pair < len(config.ToolCallsPerBodyPair) {
					toolCallsCount = config.ToolCallsPerBodyPair[pair]
				}

				var toolCallParts []llms.ContentPart
				var toolIds []string

				for t := 0; t < toolCallsCount; t++ {
					toolId := fmt.Sprintf("tool-%d", toolCallId)
					toolIds = append(toolIds, toolId)
					toolCallId++

					toolCallParts = append(toolCallParts, llms.ToolCall{
						ID:   toolId,
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      fmt.Sprintf("get_data_%d", t+1),
							Arguments: fmt.Sprintf(`{"query": "Test query %d"}`, t+1),
						},
					})
				}

				chain = append(chain, llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: toolCallParts,
				})

				// Add tool responses if requested
				if config.IncludeAllToolResponses {
					for _, toolId := range toolIds {
						toolName := ""
						for _, part := range chain[len(chain)-1].Parts {
							if tc, ok := part.(llms.ToolCall); ok && tc.ID == toolId && tc.FunctionCall != nil {
								toolName = tc.FunctionCall.Name
								break
							}
						}

						chain = append(chain, llms.MessageContent{
							Role: llms.ChatMessageTypeTool,
							Parts: []llms.ContentPart{
								llms.ToolCallResponse{
									ToolCallID: toolId,
									Name:       toolName,
									Content:    fmt.Sprintf("Response for %s", toolId),
								},
							},
						})
					}
				}
			} else {
				// Simple AI response without tool calls
				chain = append(chain, llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextContent{Text: fmt.Sprintf("Response to question %d", section+1)}},
				})
			}
		}
	}

	return chain
}

// GenerateComplexChain generates a more complex chain with multiple sections and tool calls
func GenerateComplexChain(numSections, numToolCalls, numMissingResponses int) []llms.MessageContent {
	config := ChainConfig{
		IncludeSystem: true,
		Sections:      numSections,
	}

	bodyPairs := make([]int, numSections)
	toolsForPairs := make([]bool, numSections)
	toolCallsPerPair := make([]int, numSections)

	for i := 0; i < numSections; i++ {
		bodyPairs[i] = 1
		toolsForPairs[i] = i%2 == 0 // Alternate between tool calls and simple responses
		if toolsForPairs[i] {
			toolCallsPerPair[i] = numToolCalls
		} else {
			toolCallsPerPair[i] = 0
		}
	}

	config.BodyPairsPerSection = bodyPairs
	config.ToolsForBodyPairs = toolsForPairs
	config.ToolCallsPerBodyPair = toolCallsPerPair
	config.IncludeAllToolResponses = numMissingResponses == 0

	chain := GenerateChain(config)

	// If we want missing responses, remove some of them
	if numMissingResponses > 0 {
		var newChain []llms.MessageContent
		missingCount := 0

		for i := 0; i < len(chain); i++ {
			if chain[i].Role == llms.ChatMessageTypeTool && missingCount < numMissingResponses {
				missingCount++
				continue
			}
			newChain = append(newChain, chain[i])
		}

		chain = newChain
	}

	return chain
}

// DumpChainStructure returns a string representation of the chain structure for debugging
func DumpChainStructure(chain []llms.MessageContent) string {
	var b strings.Builder
	b.WriteString("Chain Structure:\n")

	for i, msg := range chain {
		b.WriteString(fmt.Sprintf("[%d] Role: %s\n", i, msg.Role))

		for j, part := range msg.Parts {
			switch v := part.(type) {
			case llms.TextContent:
				b.WriteString(fmt.Sprintf("  [%d] TextContent: %s\n", j, truncateString(v.Text, 30)))
			case llms.ToolCall:
				if v.FunctionCall != nil {
					b.WriteString(fmt.Sprintf("  [%d] ToolCall: ID=%s, Function=%s\n", j, v.ID, v.FunctionCall.Name))
				} else {
					b.WriteString(fmt.Sprintf("  [%d] ToolCall: ID=%s (no function call)\n", j, v.ID))
				}
			case llms.ToolCallResponse:
				b.WriteString(fmt.Sprintf("  [%d] ToolCallResponse: ID=%s, Name=%s\n", j, v.ToolCallID, v.Name))
			default:
				b.WriteString(fmt.Sprintf("  [%d] Unknown part type: %T\n", j, part))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// Helper function to truncate a string for display purposes
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
