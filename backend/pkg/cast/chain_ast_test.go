package cast

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestNewChainAST_EmptyChain(t *testing.T) {
	// Test with empty chain
	ast, err := NewChainAST(emptyChain, false)
	assert.NoError(t, err)
	assert.NotNil(t, ast)
	assert.Empty(t, ast.Sections)

	// Check that Messages() returns an empty chain
	chain := ast.Messages()
	assert.Empty(t, chain)

	// Check that Dump() also returns an empty chain (backward compatibility)
	dumpedChain := ast.Messages()
	assert.Empty(t, dumpedChain)

	// Check total size is 0
	assert.Equal(t, 0, ast.Size())
}

func TestNewChainAST_BasicChains(t *testing.T) {
	tests := []struct {
		name              string
		chain             []llms.MessageContent
		expectedErr       bool
		expectedSections  int
		expectedHeaders   int
		expectNonZeroSize bool
	}{
		{
			name:              "System only",
			chain:             systemOnlyChain,
			expectedErr:       false,
			expectedSections:  1,
			expectedHeaders:   1,
			expectNonZeroSize: true,
		},
		{
			name:              "Human only",
			chain:             humanOnlyChain,
			expectedErr:       false,
			expectedSections:  1,
			expectedHeaders:   1,
			expectNonZeroSize: true,
		},
		{
			name:              "System + Human",
			chain:             systemHumanChain,
			expectedErr:       false,
			expectedSections:  1,
			expectedHeaders:   2,
			expectNonZeroSize: true,
		},
		{
			name:              "System + Human + AI",
			chain:             basicConversationChain,
			expectedErr:       false,
			expectedSections:  1,
			expectedHeaders:   2,
			expectNonZeroSize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, false)

			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, ast)
			assert.Equal(t, tt.expectedSections, len(ast.Sections))

			// Verify headers
			if len(ast.Sections) > 0 {
				section := ast.Sections[0]
				hasSystem := section.Header.SystemMessage != nil
				hasHuman := section.Header.HumanMessage != nil

				headerCount := 0
				if hasSystem {
					headerCount++
				}
				if hasHuman {
					headerCount++
				}

				assert.Equal(t, tt.expectedHeaders, headerCount, "Header count doesn't match expected value")

				// Check header size tracking
				if hasSystem || hasHuman {
					assert.Greater(t, section.Header.Size(), 0, "Header size should be greater than 0")
				}

				// Check section size tracking
				if tt.expectNonZeroSize {
					assert.Greater(t, section.Size(), 0, "Section size should be greater than 0")
					assert.Greater(t, ast.Size(), 0, "Total size should be greater than 0")
				}

				// Get messages and verify length
				messages := ast.Messages()
				assert.Equal(t, len(tt.chain), len(messages), "Messages length doesn't match original")

				// Check that Dump() returns the same result (backward compatibility)
				dumpedChain := ast.Messages()
				assert.Equal(t, len(messages), len(dumpedChain), "Messages method results should be consistent")
			}
		})
	}
}

func TestNewChainAST_ToolCallChains(t *testing.T) {
	tests := []struct {
		name                  string
		chain                 []llms.MessageContent
		force                 bool
		expectedErr           bool
		expectedBodyPairs     int
		expectedToolCalls     int
		expectedToolResponses int
		expectAddedResponses  bool
	}{
		{
			name:                  "Chain with tool call, no response, without force",
			chain:                 chainWithTool,
			force:                 false,
			expectedErr:           true, // Should error because there are tool calls without responses
			expectedBodyPairs:     1,
			expectedToolCalls:     1,
			expectedToolResponses: 0,     // No responses expected because it should error
			expectAddedResponses:  false, // No responses should be added without force=true
		},
		{
			name:                  "Chain with tool call, no response, with force",
			chain:                 chainWithTool,
			force:                 true,
			expectedErr:           false,
			expectedBodyPairs:     1,
			expectedToolCalls:     1,
			expectedToolResponses: 1,
			expectAddedResponses:  true,
		},
		{
			name:                  "Chain with tool call and response",
			chain:                 chainWithSingleToolResponse,
			force:                 false,
			expectedErr:           false,
			expectedBodyPairs:     1,
			expectedToolCalls:     1,
			expectedToolResponses: 1,
			expectAddedResponses:  false,
		},
		{
			name:                  "Chain with multiple tool calls, no responses, with force",
			chain:                 chainWithMultipleTools,
			force:                 true,
			expectedErr:           false,
			expectedBodyPairs:     1,
			expectedToolCalls:     2,
			expectedToolResponses: 2,
			expectAddedResponses:  true,
		},
		{
			name:                  "Chain with missing tool response, with force",
			chain:                 chainWithMissingToolResponse,
			force:                 true,
			expectedErr:           false,
			expectedBodyPairs:     1,
			expectedToolCalls:     2,
			expectedToolResponses: 2,
			expectAddedResponses:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, tt.force)

			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, ast)
			assert.NotEmpty(t, ast.Sections)

			// Get the first section's body pairs to analyze
			section := ast.Sections[0]
			assert.Equal(t, tt.expectedBodyPairs, len(section.Body))

			if len(section.Body) > 0 {
				bodyPair := section.Body[0]

				if tt.expectedToolCalls > 0 {
					assert.Equal(t, RequestResponse, bodyPair.Type)

					// Count actual tool calls in the AI message
					toolCallCount := 0
					toolCallIDs := []string{}
					for _, part := range bodyPair.AIMessage.Parts {
						if toolCall, ok := part.(llms.ToolCall); ok {
							toolCallCount++
							toolCallIDs = append(toolCallIDs, toolCall.ID)
						}
					}
					assert.Equal(t, tt.expectedToolCalls, toolCallCount, "Tool call count doesn't match expected value")
					t.Logf("Tool call IDs: %v", toolCallIDs)

					// Check tool responses
					responseCount := 0
					responseIDs := []string{}
					for _, toolMsg := range bodyPair.ToolMessages {
						for _, part := range toolMsg.Parts {
							if resp, ok := part.(llms.ToolCallResponse); ok {
								responseCount++
								responseIDs = append(responseIDs, resp.ToolCallID)
							}
						}
					}
					assert.Equal(t, tt.expectedToolResponses, responseCount, "Tool response count doesn't match expected value")
					t.Logf("Tool response IDs: %v", responseIDs)

					// Verify matching between tool calls and responses
					toolCallsInfo := bodyPair.GetToolCallsInfo()
					t.Logf("Pending tool call IDs: %v", toolCallsInfo.PendingToolCallIDs)
					t.Logf("Unmatched tool call IDs: %v", toolCallsInfo.UnmatchedToolCallIDs)
					t.Logf("Completed tool calls: %v", toolCallsInfo.CompletedToolCalls)

					// If we expect all tools to have responses, verify that
					if tt.force {
						assert.Empty(t, toolCallsInfo.PendingToolCallIDs, "With force=true, there should be no pending tool calls")
					}
				} else {
					assert.Equal(t, Completion, bodyPair.Type)
				}
			}

			// Test dumping
			chain := ast.Messages()

			// Check chain length based on whether responses were added
			if tt.expectAddedResponses {
				// If we expect responses to be added, don't check exact equality
				t.Logf("Original chain length: %d, Dumped chain length: %d", len(tt.chain), len(chain))
			} else {
				assert.Equal(t, len(tt.chain), len(chain), "Dumped chain length doesn't match original without force changes")
			}

			// Debug output
			if t.Failed() {
				t.Logf("Original chain structure: \n%s", DumpChainStructure(tt.chain))
				t.Logf("AST structure: \n%s", ast.String())
				t.Logf("Dumped chain structure: \n%s", DumpChainStructure(chain))
			}
		})
	}
}

func TestNewChainAST_MultipleHumanMessages(t *testing.T) {
	// Test with chain containing multiple human messages (sections)
	ast, err := NewChainAST(chainWithMultipleSections, false)
	assert.NoError(t, err)
	assert.NotNil(t, ast)
	assert.Equal(t, 2, len(ast.Sections), "Should have two sections")

	// First section should have system, human, and AI message
	assert.NotNil(t, ast.Sections[0].Header.SystemMessage)
	assert.NotNil(t, ast.Sections[0].Header.HumanMessage)
	assert.Equal(t, 1, len(ast.Sections[0].Body))
	assert.Equal(t, Completion, ast.Sections[0].Body[0].Type)

	// Second section should have human, and AI with tool call
	assert.NotNil(t, ast.Sections[1].Header.HumanMessage)
	assert.Equal(t, 1, len(ast.Sections[1].Body))
	assert.Equal(t, RequestResponse, ast.Sections[1].Body[0].Type)

	// The tool call should have a response
	toolMsg := ast.Sections[1].Body[0].ToolMessages
	assert.Equal(t, 1, len(toolMsg))

	// Dump and verify length
	chain := ast.Messages()
	assert.Equal(t, len(chainWithMultipleSections), len(chain))
}

func TestNewChainAST_ConsecutiveHumans(t *testing.T) {
	// Modify chainWithConsecutiveHumans for the test
	// One System + two Human in a row
	testChain := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "First human message"}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Second human message"}},
		},
	}

	// Test without force (should error)
	_, err := NewChainAST(testChain, false)
	assert.Error(t, err, "Should error with consecutive humans without force=true")

	// Test with force (should merge)
	ast, err := NewChainAST(testChain, true)
	assert.NoError(t, err)
	assert.NotNil(t, ast)

	// Check that we have only one section
	assert.Equal(t, 1, len(ast.Sections), "Should have one section after merging consecutive humans")

	// Verify the merged parts - human message should have 2 parts after merge
	humanMsg := ast.Sections[0].Header.HumanMessage
	assert.NotNil(t, humanMsg)
	assert.Equal(t, 2, len(humanMsg.Parts), "Human message should contain both parts after merge")
}

func TestNewChainAST_UnexpectedTool(t *testing.T) {
	// Test with unexpected tool message without force
	_, err := NewChainAST(chainWithUnexpectedTool, false)
	assert.Error(t, err, "Should error with unexpected tool message")

	// Test with force (should skip the invalid tool message)
	ast, err := NewChainAST(chainWithUnexpectedTool, true)
	assert.NoError(t, err, "Should not error with force=true")
	assert.NotNil(t, ast)

	// Check that all valid messages were processed
	assert.Equal(t, 1, len(ast.Sections), "Should have one section")

	// Verify section structure
	if len(ast.Sections) > 0 {
		section := ast.Sections[0]
		assert.NotNil(t, section.Header.SystemMessage)
		assert.NotNil(t, section.Header.HumanMessage)
		assert.Equal(t, 1, len(section.Body))

		// The unexpected tool message should have been skipped
		chain := ast.Messages()
		assert.True(t, len(chain) < len(chainWithUnexpectedTool),
			"Dumped chain should be shorter than original after skipping invalid messages")
	}
}

func TestAddToolResponse(t *testing.T) {
	// Create a chain with one tool call and immediately add a response
	// to meet the requirement force=false
	toolCallID := "test-tool-1"
	toolCallName := "get_weather"

	completedChain := []llms.MessageContent{
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
					ID:   toolCallID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      toolCallName,
						Arguments: `{"location": "New York"}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: toolCallID,
					Name:       toolCallName,
					Content:    "Initial response",
				},
			},
		},
	}

	// Create a chain that already has a response for the tool call
	ast, err := NewChainAST(completedChain, false)
	assert.NoError(t, err)
	assert.NotNil(t, ast)

	// Add an updated response
	updatedContent := "The weather in New York is sunny."
	err = ast.AddToolResponse(toolCallID, toolCallName, updatedContent)
	assert.NoError(t, err)

	// Verify the response was added or updated
	responses := ast.FindToolCallResponses(toolCallID)
	assert.Equal(t, 1, len(responses), "Should have exactly one tool response")
	assert.Equal(t, updatedContent, responses[0].Content, "Response content should match the updated content")
	assert.Equal(t, toolCallName, responses[0].Name, "Tool name should match")

	// Test with invalid tool call ID
	err = ast.AddToolResponse("invalid-id", "invalid-name", "content")
	assert.Error(t, err, "Should error with invalid tool call ID")
}

func TestAppendHumanMessage(t *testing.T) {
	tests := []struct {
		name             string
		chain            []llms.MessageContent
		content          string
		expectedSections int
		expectedHeaders  int
	}{
		{
			name:             "Empty chain",
			chain:            emptyChain,
			content:          "Hello",
			expectedSections: 1,
			expectedHeaders:  1,
		},
		{
			name:             "Chain with system only",
			chain:            systemOnlyChain,
			content:          "Hello",
			expectedSections: 1,
			expectedHeaders:  2,
		},
		{
			name:             "Chain with existing conversation",
			chain:            basicConversationChain,
			content:          "Tell me more",
			expectedSections: 2,
			expectedHeaders:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, false)
			assert.NoError(t, err)

			// Append the human message
			ast.AppendHumanMessage(tt.content)

			// Check the results - total sections
			assert.Equal(t, tt.expectedSections, len(ast.Sections),
				"Section count doesn't match expected. AST structure: %s", ast.String())

			// Check the appended message
			lastSection := ast.Sections[len(ast.Sections)-1]
			assert.NotNil(t, lastSection.Header.HumanMessage)

			// Count headers to verify the human message was added
			headerCount := 0
			for _, section := range ast.Sections {
				if section.Header.SystemMessage != nil {
					headerCount++
				}
				if section.Header.HumanMessage != nil {
					headerCount++
				}
			}
			assert.Equal(t, tt.expectedHeaders, headerCount, "Total header count doesn't match expected")

			// Check the content of the appended message
			var textFound bool
			for _, part := range lastSection.Header.HumanMessage.Parts {
				if textContent, ok := part.(llms.TextContent); ok {
					if textContent.Text == tt.content {
						textFound = true
						break
					}
				}
			}
			assert.True(t, textFound, "Appended human message content not found")

			// Dump and check chain length
			chain := ast.Messages()

			// For empty chain, adding a human message adds one message
			// For system-only chain, adding a human message adds one message
			// For existing conversation, adding a human message adds one message
			expectedLength := len(tt.chain) + 1
			assert.Equal(t, expectedLength, len(chain),
				"Dumped chain length mismatch after appending human message")
		})
	}
}

func TestGeneratedChains(t *testing.T) {
	tests := []struct {
		name              string
		config            ChainConfig
		force             bool
		expectedSections  int
		expectedBodyPairs int
	}{
		{
			name:              "Default config",
			config:            DefaultChainConfig(),
			force:             false,
			expectedSections:  1,
			expectedBodyPairs: 1,
		},
		{
			name: "Multiple sections",
			config: ChainConfig{
				IncludeSystem:           true,
				Sections:                3,
				BodyPairsPerSection:     []int{1, 2, 1},
				ToolsForBodyPairs:       []bool{false, true, false},
				ToolCallsPerBodyPair:    []int{0, 2, 0},
				IncludeAllToolResponses: true,
			},
			force:             false,
			expectedSections:  3,
			expectedBodyPairs: 4, // 1 + 2 + 1
		},
		{
			name: "Missing tool responses",
			config: ChainConfig{
				IncludeSystem:           true,
				Sections:                1,
				BodyPairsPerSection:     []int{1},
				ToolsForBodyPairs:       []bool{true},
				ToolCallsPerBodyPair:    []int{2},
				IncludeAllToolResponses: false,
			},
			force:             true,
			expectedSections:  1,
			expectedBodyPairs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a chain using the config
			chain := GenerateChain(tt.config)

			// Create AST from the generated chain
			ast, err := NewChainAST(chain, tt.force)
			assert.NoError(t, err)

			// Verify section count
			assert.Equal(t, tt.expectedSections, len(ast.Sections))

			// Count total body pairs
			totalBodyPairs := 0
			for _, section := range ast.Sections {
				totalBodyPairs += len(section.Body)
			}
			assert.Equal(t, tt.expectedBodyPairs, totalBodyPairs)

			// Dump the chain
			dumpedChain := ast.Messages()

			// Without force and all responses, lengths should match
			if !tt.force && tt.config.IncludeAllToolResponses {
				assert.Equal(t, len(chain), len(dumpedChain))
			}

			// With force and missing responses, dumped chain might be longer
			if tt.force && !tt.config.IncludeAllToolResponses {
				assert.True(t, len(dumpedChain) >= len(chain))
			}

			// Debug output
			if t.Failed() {
				t.Logf("Generated chain structure: \n%s", DumpChainStructure(chain))
				t.Logf("AST structure: \n%s", ast.String())
				t.Logf("Dumped chain structure: \n%s", DumpChainStructure(dumpedChain))
			}
		})
	}
}

func TestComplexGeneratedChains(t *testing.T) {
	// Generate complex chains with various configurations
	chains := []struct {
		name         string
		sections     int
		toolCalls    int
		missingResps int
	}{
		{
			name:         "Small chain, all responses",
			sections:     2,
			toolCalls:    1,
			missingResps: 0,
		},
		{
			name:         "Medium chain, some missing responses",
			sections:     3,
			toolCalls:    2,
			missingResps: 2,
		},
		{
			name:         "Large chain, many missing responses",
			sections:     5,
			toolCalls:    3,
			missingResps: 7,
		},
	}

	for _, tc := range chains {
		t.Run(tc.name, func(t *testing.T) {
			chain := GenerateComplexChain(tc.sections, tc.toolCalls, tc.missingResps)

			t.Logf("Generated chain length: %d", len(chain))
			t.Logf("Generated chain structure: \n%s", DumpChainStructure(chain))

			// Parse with force = true
			ast, err := NewChainAST(chain, true)
			assert.NoError(t, err, "Should parse complex chain without error")

			// Dump and verify all tool calls have responses
			dumpedChain := ast.Messages()

			// If we had missing responses and force=true, dumped chain should be longer
			if tc.missingResps > 0 {
				assert.True(t, len(dumpedChain) >= len(chain),
					fmt.Sprintf("Dumped chain should be at least as long as original when fixing missing responses"))
			}

			// Check if all tool calls have responses
			newAst, err := NewChainAST(dumpedChain, false)
			assert.NoError(t, err)

			// Verify all tool calls have responses
			for _, section := range newAst.Sections {
				for _, bodyPair := range section.Body {
					if bodyPair.Type == RequestResponse {
						// Count tool calls
						toolCalls := 0
						toolCallIDs := make(map[string]bool)

						for _, part := range bodyPair.AIMessage.Parts {
							if toolCall, ok := part.(llms.ToolCall); ok && toolCall.FunctionCall != nil {
								toolCalls++
								toolCallIDs[toolCall.ID] = true
							}
						}

						// Count tool responses
						responses := 0
						respondedIDs := make(map[string]bool)

						for _, toolMsg := range bodyPair.ToolMessages {
							for _, part := range toolMsg.Parts {
								if resp, ok := part.(llms.ToolCallResponse); ok {
									responses++
									respondedIDs[resp.ToolCallID] = true
								}
							}
						}

						// Verify every tool call has a response
						assert.Equal(t, toolCalls, responses, "Each tool call should have exactly one response")

						for id := range toolCallIDs {
							assert.True(t, respondedIDs[id], "Tool call ID %s should have a response", id)
						}
					}
				}
			}
		})
	}
}

func TestMessages(t *testing.T) {
	// Test that all components correctly implement Messages()

	// Create a test chain with different message types
	chain := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "System message"}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Human message"}},
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
					Content:    "The weather in New York is sunny.",
				},
			},
		},
	}

	ast, err := NewChainAST(chain, false)
	assert.NoError(t, err)

	// Test Header.Messages()
	headerMsgs := ast.Sections[0].Header.Messages()
	assert.Equal(t, 2, len(headerMsgs), "Header should return system and human messages")
	assert.Equal(t, llms.ChatMessageTypeSystem, headerMsgs[0].Role)
	assert.Equal(t, llms.ChatMessageTypeHuman, headerMsgs[1].Role)

	// Test BodyPair.Messages()
	bodyPairMsgs := ast.Sections[0].Body[0].Messages()
	assert.Equal(t, 2, len(bodyPairMsgs), "BodyPair should return AI and tool messages")
	assert.Equal(t, llms.ChatMessageTypeAI, bodyPairMsgs[0].Role)
	assert.Equal(t, llms.ChatMessageTypeTool, bodyPairMsgs[1].Role)

	// Test ChainSection.Messages()
	sectionMsgs := ast.Sections[0].Messages()
	assert.Equal(t, 4, len(sectionMsgs), "Section should return all messages in order")

	// Test ChainAST.Messages()
	allMsgs := ast.Messages()
	assert.Equal(t, len(chain), len(allMsgs), "AST should return all messages")

	// Check order preservation
	for i, msg := range chain {
		assert.Equal(t, msg.Role, allMsgs[i].Role, "Role mismatch at position %d", i)
	}
}

func TestConstructors(t *testing.T) {
	// Test all the constructors

	// Test NewHeader
	sysMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: "System message"}},
	}
	humanMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "Human message"}},
	}

	header := NewHeader(sysMsg, humanMsg)
	assert.NotNil(t, header)
	assert.Equal(t, sysMsg, header.SystemMessage)
	assert.Equal(t, humanMsg, header.HumanMessage)
	assert.Greater(t, header.Size(), 0, "Header size should be calculated")

	// Test NewHeader with nil messages
	headerWithNilSystem := NewHeader(nil, humanMsg)
	assert.NotNil(t, headerWithNilSystem)
	assert.Nil(t, headerWithNilSystem.SystemMessage)
	assert.Equal(t, humanMsg, headerWithNilSystem.HumanMessage)
	assert.Greater(t, headerWithNilSystem.Size(), 0)

	headerWithNilHuman := NewHeader(sysMsg, nil)
	assert.NotNil(t, headerWithNilHuman)
	assert.Equal(t, sysMsg, headerWithNilHuman.SystemMessage)
	assert.Nil(t, headerWithNilHuman.HumanMessage)
	assert.Greater(t, headerWithNilHuman.Size(), 0)

	// Test NewBodyPair for Completion type
	aiMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{llms.TextContent{Text: "AI message"}},
	}

	completionPair := NewBodyPair(aiMsg, nil)
	assert.NotNil(t, completionPair)
	assert.Equal(t, Completion, completionPair.Type)
	assert.Equal(t, aiMsg, completionPair.AIMessage)
	assert.Empty(t, completionPair.ToolMessages)
	assert.Greater(t, completionPair.Size(), 0, "BodyPair size should be calculated")

	// Test NewBodyPair for RequestResponse type
	aiMsgWithTool := &llms.MessageContent{
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
	}
	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "tool-1",
				Name:       "get_weather",
				Content:    "The weather in New York is sunny.",
			},
		},
	}

	requestResponsePair := NewBodyPair(aiMsgWithTool, []*llms.MessageContent{toolMsg})
	assert.NotNil(t, requestResponsePair)
	assert.Equal(t, RequestResponse, requestResponsePair.Type)
	assert.Equal(t, aiMsgWithTool, requestResponsePair.AIMessage)
	assert.Equal(t, 1, len(requestResponsePair.ToolMessages))
	assert.Greater(t, requestResponsePair.Size(), 0, "BodyPair size should be calculated")

	// Test NewBodyPairFromMessages
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "AI message"}},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "tool-1",
					Name:       "get_weather",
					Content:    "The weather in New York is sunny.",
				},
			},
		},
	}

	bodyPair, err := NewBodyPairFromMessages(messages)
	assert.NoError(t, err)
	assert.NotNil(t, bodyPair)
	assert.Equal(t, Completion, bodyPair.Type) // No tool calls, so it's a Completion
	assert.Equal(t, 1, len(bodyPair.ToolMessages))

	// Test error case for NewBodyPairFromMessages
	invalidMessages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman, // First message should be AI
			Parts: []llms.ContentPart{llms.TextContent{Text: "Human message"}},
		},
	}

	_, err = NewBodyPairFromMessages(invalidMessages)
	assert.Error(t, err)

	emptyMessages := []llms.MessageContent{}
	_, err = NewBodyPairFromMessages(emptyMessages)
	assert.Error(t, err)

	// Test NewChainSection
	section := NewChainSection(header, []*BodyPair{completionPair, requestResponsePair})
	assert.NotNil(t, section)
	assert.Equal(t, header, section.Header)
	assert.Equal(t, 2, len(section.Body))
	assert.Equal(t, header.Size()+completionPair.Size()+requestResponsePair.Size(),
		section.Size(), "Section size should be sum of header and body pair sizes")

	// Test NewBodyPairFromCompletion
	text := "This is a completion response"
	pair := NewBodyPairFromCompletion(text)
	assert.NotNil(t, pair)
	assert.Equal(t, Completion, pair.Type)
	assert.NotNil(t, pair.AIMessage)
	assert.Equal(t, llms.ChatMessageTypeAI, pair.AIMessage.Role)

	// Extract text from the message
	textContent, ok := pair.AIMessage.Parts[0].(llms.TextContent)
	assert.True(t, ok)
	assert.Equal(t, text, textContent.Text)

	// Test HasToolCalls
	assert.True(t, HasToolCalls(aiMsgWithTool))
	assert.False(t, HasToolCalls(aiMsg))
	assert.False(t, HasToolCalls(nil))
}

func TestSizeTracking(t *testing.T) {
	// Test size calculation and tracking

	// Test CalculateMessageSize with different content types
	textMsg := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "Hello world"},
		},
	}
	textSize := CalculateMessageSize(&textMsg)
	assert.Equal(t, len("Hello world"), textSize)

	// Test with image URL
	imageMsg := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.ImageURLContent{URL: "https://example.com/image.jpg"},
		},
	}
	imageSize := CalculateMessageSize(&imageMsg)
	assert.Equal(t, len("https://example.com/image.jpg"), imageSize)

	// Test with tool call
	toolCallMsg := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   "call-1",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "test_function",
					Arguments: `{"param1": "value1"}`,
				},
			},
		},
	}
	toolCallSize := CalculateMessageSize(&toolCallMsg)
	expectedSize := len("call-1") + len("function") + len("test_function") + len(`{"param1": "value1"}`)
	assert.Equal(t, expectedSize, toolCallSize)

	// Test with tool response
	toolResponseMsg := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "call-1",
				Name:       "test_function",
				Content:    "Response content",
			},
		},
	}
	toolResponseSize := CalculateMessageSize(&toolResponseMsg)
	expectedResponseSize := len("call-1") + len("test_function") + len("Response content")
	assert.Equal(t, expectedResponseSize, toolResponseSize)

	// Test size changes when modifying AST

	// Create a basic AST
	ast := &ChainAST{Sections: []*ChainSection{}}
	assert.Equal(t, 0, ast.Size())

	// Add a section with system message
	sysMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: "System message"}},
	}
	header := NewHeader(sysMsg, nil)
	section := NewChainSection(header, []*BodyPair{})
	ast.AddSection(section)

	initialSize := ast.Size()
	assert.Equal(t, CalculateMessageSize(sysMsg), initialSize)

	// Add a human message and verify size increases
	humanContent := "Human message"
	ast.AppendHumanMessage(humanContent)

	expectedIncrease := len(humanContent)
	assert.Equal(t, initialSize+expectedIncrease, ast.Size())

	// Add a body pair and verify size increases
	aiMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{llms.TextContent{Text: "AI response"}},
	}
	bodyPair := NewBodyPair(aiMsg, nil)
	section.AddBodyPair(bodyPair)

	expectedBodyPairSize := CalculateMessageSize(aiMsg)
	assert.Equal(t, initialSize+expectedIncrease+expectedBodyPairSize, ast.Size())
}

func TestAddSectionAndBodyPair(t *testing.T) {
	// Test adding sections and body pairs

	// Create empty AST
	ast := &ChainAST{Sections: []*ChainSection{}}

	// Create section 1
	header1 := NewHeader(nil, &llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "Question 1"}},
	})
	section1 := NewChainSection(header1, []*BodyPair{})

	// Add section 1
	ast.AddSection(section1)
	assert.Equal(t, 1, len(ast.Sections))

	// Add body pair to section 1
	bodyPair1 := NewBodyPairFromCompletion("Answer 1")
	section1.AddBodyPair(bodyPair1)
	assert.Equal(t, 1, len(section1.Body))

	// Create and add section 2
	header2 := NewHeader(nil, &llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "Question 2"}},
	})
	section2 := NewChainSection(header2, []*BodyPair{})
	ast.AddSection(section2)
	assert.Equal(t, 2, len(ast.Sections))

	// Add body pair with tool call to section 2
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   "tool-1",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "test"}`,
				},
			},
		},
	}
	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "tool-1",
				Name:       "search",
				Content:    "Search results",
			},
		},
	}
	bodyPair2 := NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	section2.AddBodyPair(bodyPair2)
	assert.Equal(t, 1, len(section2.Body))
	assert.Equal(t, RequestResponse, section2.Body[0].Type)

	// Check that Messages() returns all messages in correct order
	messages := ast.Messages()
	assert.Equal(t, 5, len(messages)) // 2 human + 1 AI + 1 Tool + 1 AI

	// Order should be: human, AI, human, AI, tool
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[0].Role)
	assert.Equal(t, llms.ChatMessageTypeAI, messages[1].Role)
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[2].Role)
	assert.Equal(t, llms.ChatMessageTypeAI, messages[3].Role)
	assert.Equal(t, llms.ChatMessageTypeTool, messages[4].Role)
}

func TestAppendHumanMessageComplex(t *testing.T) {
	// Test complex scenarios with AppendHumanMessage

	// Test case 1: Empty AST
	ast1 := &ChainAST{Sections: []*ChainSection{}}
	ast1.AppendHumanMessage("First message")

	assert.Equal(t, 1, len(ast1.Sections))
	assert.NotNil(t, ast1.Sections[0].Header.HumanMessage)
	assert.Equal(t, "First message", extractText(ast1.Sections[0].Header.HumanMessage))

	// Test case 2: AST with system message only
	ast2 := &ChainAST{Sections: []*ChainSection{}}
	sysMsg := &llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{llms.TextContent{Text: "System prompt"}},
	}
	header := NewHeader(sysMsg, nil)
	section := NewChainSection(header, []*BodyPair{})
	ast2.AddSection(section)

	ast2.AppendHumanMessage("Human question")

	assert.Equal(t, 1, len(ast2.Sections))
	assert.NotNil(t, ast2.Sections[0].Header.SystemMessage)
	assert.NotNil(t, ast2.Sections[0].Header.HumanMessage)
	assert.Equal(t, "Human question", extractText(ast2.Sections[0].Header.HumanMessage))

	// Test case 3: AST with system+human but no body pairs
	ast3 := &ChainAST{Sections: []*ChainSection{}}
	header3 := NewHeader(
		&llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "System"}},
		},
		&llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Initial"}},
		},
	)
	section3 := NewChainSection(header3, []*BodyPair{})
	ast3.AddSection(section3)

	// Should append to existing human message
	ast3.AppendHumanMessage("Additional")

	assert.Equal(t, 1, len(ast3.Sections))
	humanMsg := ast3.Sections[0].Header.HumanMessage
	assert.NotNil(t, humanMsg)

	// Check that both parts are present in the correct order
	assert.Equal(t, 2, len(humanMsg.Parts))
	textPart1, ok1 := humanMsg.Parts[0].(llms.TextContent)
	textPart2, ok2 := humanMsg.Parts[1].(llms.TextContent)
	assert.True(t, ok1 && ok2)
	assert.Equal(t, "Initial", textPart1.Text)
	assert.Equal(t, "Additional", textPart2.Text)

	// Test case 4: AST with complete section (system+human+body pairs)
	ast4 := &ChainAST{Sections: []*ChainSection{}}
	header4 := NewHeader(
		&llms.MessageContent{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "System"}},
		},
		&llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Question"}},
		},
	)
	bodyPair4 := NewBodyPairFromCompletion("Answer")
	section4 := NewChainSection(header4, []*BodyPair{bodyPair4})
	ast4.AddSection(section4)

	// Should create new section
	ast4.AppendHumanMessage("Follow-up")

	assert.Equal(t, 2, len(ast4.Sections))
	assert.Nil(t, ast4.Sections[1].Header.SystemMessage)
	assert.NotNil(t, ast4.Sections[1].Header.HumanMessage)
	assert.Equal(t, "Follow-up", extractText(ast4.Sections[1].Header.HumanMessage))
}

func TestAddToolResponseComplex(t *testing.T) {
	// Test complex scenarios with AddToolResponse

	// Create an AST with multiple tool calls
	chain := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "System prompt"}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Tell me about the weather and news"}},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "weather-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "New York"}`,
					},
				},
				llms.ToolCall{
					ID:   "news-1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_news",
						Arguments: `{"topic": "technology"}`,
					},
				},
			},
		},
	}

	// Using force=true because the original chain does not contain responses to tool calls
	ast, err := NewChainAST(chain, true)
	assert.NoError(t, err)

	// Test case 1: Add response to first tool call
	weatherResponse := "Sunny and 75°F in New York"
	err = ast.AddToolResponse("weather-1", "get_weather", weatherResponse)
	assert.NoError(t, err)

	// Verify the response was added
	responses := ast.FindToolCallResponses("weather-1")
	assert.Equal(t, 1, len(responses))
	assert.Equal(t, weatherResponse, responses[0].Content)

	// Test case 2: Add response to second tool call
	newsResponse := "Latest tech news: AI advances"
	err = ast.AddToolResponse("news-1", "get_news", newsResponse)
	assert.NoError(t, err)

	// Verify the response was added
	responses = ast.FindToolCallResponses("news-1")
	assert.Equal(t, 1, len(responses))
	assert.Equal(t, newsResponse, responses[0].Content)

	// Test case 3: Update existing response
	updatedWeatherResponse := "Partly cloudy and 72°F in New York"
	err = ast.AddToolResponse("weather-1", "get_weather", updatedWeatherResponse)
	assert.NoError(t, err)

	// Verify the response was updated
	responses = ast.FindToolCallResponses("weather-1")
	assert.Equal(t, 1, len(responses))
	assert.Equal(t, updatedWeatherResponse, responses[0].Content)

	// Test case 4: Invalid tool call ID
	err = ast.AddToolResponse("invalid-id", "invalid-function", "Response")
	assert.Error(t, err)
}

// Helper function to extract text from a message
func extractText(msg *llms.MessageContent) string {
	if msg == nil {
		return ""
	}

	var result strings.Builder
	for _, part := range msg.Parts {
		if textContent, ok := part.(llms.TextContent); ok {
			result.WriteString(textContent.Text)
		}
	}

	return result.String()
}

func TestNewChainAST_Summarization(t *testing.T) {
	tests := []struct {
		name                string
		chain               []llms.MessageContent
		force               bool
		expectedErr         bool
		expectedSections    int
		expectedBodyPairs   int
		expectedBodyPairIdx int
		expectedType        BodyPairType
	}{
		{
			name:                "Chain with summarization as the only body pair",
			chain:               chainWithSummarization,
			force:               false,
			expectedErr:         false,
			expectedSections:    1,
			expectedBodyPairs:   1,
			expectedBodyPairIdx: 0,
			expectedType:        Summarization,
		},
		{
			name:                "Chain with summarization followed by other pairs",
			chain:               chainWithSummarizationAndOtherPairs,
			force:               false,
			expectedErr:         false,
			expectedSections:    1,
			expectedBodyPairs:   3, // Summarization + text + tool call
			expectedBodyPairIdx: 0,
			expectedType:        Summarization,
		},
		// Test for missing response with force=true
		{
			name: "Chain with summarization missing tool response but force=true",
			chain: []llms.MessageContent{
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
							ID:   "summary-missing",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      SummarizationToolName,
								Arguments: SummarizationToolArgs,
							},
						},
					},
				},
				// No tool response
			},
			force:               true,
			expectedErr:         false,
			expectedSections:    1,
			expectedBodyPairs:   1,
			expectedBodyPairIdx: 0,
			expectedType:        Summarization,
		},
		// Test for missing response with force=false
		{
			name: "Chain with summarization missing tool response and force=false",
			chain: []llms.MessageContent{
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
							ID:   "summary-missing",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      SummarizationToolName,
								Arguments: SummarizationToolArgs,
							},
						},
					},
				},
				// No tool response
			},
			force:               false,
			expectedErr:         true,
			expectedSections:    0,
			expectedBodyPairs:   0,
			expectedBodyPairIdx: 0,
			expectedType:        Summarization,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing chain with %d messages", len(tt.chain))

			ast, err := NewChainAST(tt.chain, tt.force)

			if tt.expectedErr {
				assert.Error(t, err)
				t.Logf("Got expected error: %v", err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, ast)
			assert.Equal(t, tt.expectedSections, len(ast.Sections), "Section count doesn't match expected")

			if tt.expectedSections == 0 {
				return
			}

			section := ast.Sections[0]
			assert.Equal(t, tt.expectedBodyPairs, len(section.Body), "Body pair count doesn't match expected")

			if len(section.Body) <= tt.expectedBodyPairIdx {
				t.Fatalf("Not enough body pairs: got %d, index %d requested",
					len(section.Body), tt.expectedBodyPairIdx)
				return
			}

			// Check that the specified body pair is of the expected type
			bodyPair := section.Body[tt.expectedBodyPairIdx]
			assert.Equal(t, tt.expectedType, bodyPair.Type, "Body pair type doesn't match expected")

			// Log the structure of the AST for easier debugging
			t.Logf("AST Structure: %s", ast.String())

			// Specifically for summarization, check that:
			// 1. The function call name is SummarizationToolName
			// 2. The first tool message response is for this call
			if tt.expectedType == Summarization {
				found := false
				var toolCallID string
				for i, part := range bodyPair.AIMessage.Parts {
					if toolCall, ok := part.(llms.ToolCall); ok &&
						toolCall.FunctionCall != nil &&
						toolCall.FunctionCall.Name == SummarizationToolName {
						found = true
						toolCallID = toolCall.ID
						t.Logf("Found summarization tool call at index %d with ID %s", i, toolCallID)
						break
					}
				}
				assert.True(t, found, "Summarization tool call not found in body pair")

				// Check that we have a matching tool response
				if len(bodyPair.ToolMessages) > 0 {
					foundResponse := false
					for i, tool := range bodyPair.ToolMessages {
						for j, part := range tool.Parts {
							if resp, ok := part.(llms.ToolCallResponse); ok &&
								resp.ToolCallID == toolCallID &&
								resp.Name == SummarizationToolName {
								foundResponse = true
								t.Logf("Found matching tool response at tool message %d, part %d", i, j)
								break
							}
						}
						if foundResponse {
							break
						}
					}
					assert.True(t, foundResponse, "Matching tool response not found for summarization tool call")
				} else if tt.force {
					// If force=true, even with no original tool response, a response should be added
					assert.NotEmpty(t, bodyPair.ToolMessages,
						"With force=true, a tool response should be automatically added")
				}

				// Check that the body pair is valid
				assert.True(t, bodyPair.IsValid(), "Body pair should be valid")

				// Check that GetToolCallsInfo returns expected results
				toolCallsInfo := bodyPair.GetToolCallsInfo()
				assert.Empty(t, toolCallsInfo.PendingToolCallIDs, "Should have no pending tool calls")
				assert.Empty(t, toolCallsInfo.UnmatchedToolCallIDs, "Should have no unmatched tool calls")

				// For each completed tool call, verify it has the right name
				for id, pair := range toolCallsInfo.CompletedToolCalls {
					t.Logf("Completed tool call: ID=%s, Name=%s", id, pair.ToolCall.FunctionCall.Name)
					assert.Equal(t, SummarizationToolName, pair.ToolCall.FunctionCall.Name,
						"Completed tool call should be a summarization call")
				}
			}

			// Test dumping
			chain := ast.Messages()

			// If force=true with missing responses, the dumped chain should be longer
			if tt.force && len(tt.chain) < len(chain) {
				t.Logf("Force=true added responses: original length %d, dumped length %d",
					len(tt.chain), len(chain))
			} else {
				assert.Equal(t, len(tt.chain), len(chain),
					"Dumped chain length should match original")
			}

			// Verify the dumped chain can be parsed again without error
			_, err = NewChainAST(chain, false)
			assert.NoError(t, err, "Re-parsing the dumped chain should not error")
		})
	}
}

func TestBodyPairConstructors(t *testing.T) {
	// Test cases for NewBodyPair
	t.Run("NewBodyPair", func(t *testing.T) {
		// Test creating a Completion body pair
		aiMsgCompletion := &llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Simple text response"}},
		}

		completionPair := NewBodyPair(aiMsgCompletion, nil)
		assert.NotNil(t, completionPair)
		assert.Equal(t, Completion, completionPair.Type)
		assert.Equal(t, aiMsgCompletion, completionPair.AIMessage)
		assert.Empty(t, completionPair.ToolMessages)
		assert.True(t, completionPair.IsValid())
		assert.Greater(t, completionPair.Size(), 0)

		messages := completionPair.Messages()
		assert.Equal(t, 1, len(messages))
		assert.Equal(t, llms.ChatMessageTypeAI, messages[0].Role)

		// Log details for better debugging
		t.Logf("Completion pair size: %d bytes", completionPair.Size())

		// Test creating a RequestResponse body pair
		aiMsgToolCall := &llms.MessageContent{
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
		}

		toolMsg := []*llms.MessageContent{
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

		requestResponsePair := NewBodyPair(aiMsgToolCall, toolMsg)
		assert.NotNil(t, requestResponsePair)
		assert.Equal(t, RequestResponse, requestResponsePair.Type)
		assert.Equal(t, aiMsgToolCall, requestResponsePair.AIMessage)
		assert.Equal(t, toolMsg, requestResponsePair.ToolMessages)
		assert.True(t, requestResponsePair.IsValid())
		assert.Greater(t, requestResponsePair.Size(), 0)

		messages = requestResponsePair.Messages()
		assert.Equal(t, 2, len(messages))
		assert.Equal(t, llms.ChatMessageTypeAI, messages[0].Role)
		assert.Equal(t, llms.ChatMessageTypeTool, messages[1].Role)

		t.Logf("RequestResponse pair size: %d bytes", requestResponsePair.Size())

		// Test creating a Summarization body pair
		aiMsgSummarization := &llms.MessageContent{
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
		}

		toolMsgSummarization := []*llms.MessageContent{
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: "summary-1",
						Name:       SummarizationToolName,
						Content:    "This is a summary of the conversation.",
					},
				},
			},
		}

		summarizationPair := NewBodyPair(aiMsgSummarization, toolMsgSummarization)
		assert.NotNil(t, summarizationPair)
		assert.Equal(t, Summarization, summarizationPair.Type)
		assert.Equal(t, aiMsgSummarization, summarizationPair.AIMessage)
		assert.Equal(t, toolMsgSummarization, summarizationPair.ToolMessages)
		assert.True(t, summarizationPair.IsValid())
		assert.Greater(t, summarizationPair.Size(), 0)

		messages = summarizationPair.Messages()
		assert.Equal(t, 2, len(messages))
		assert.Equal(t, llms.ChatMessageTypeAI, messages[0].Role)
		assert.Equal(t, llms.ChatMessageTypeTool, messages[1].Role)

		t.Logf("Summarization pair size: %d bytes", summarizationPair.Size())

		// Test Completion with multiple text parts
		aiMsgMultiParts := &llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "First part of the response."},
				llms.TextContent{Text: "Second part of the response."},
			},
		}

		multiPartsPair := NewBodyPair(aiMsgMultiParts, nil)
		assert.NotNil(t, multiPartsPair)
		assert.Equal(t, Completion, multiPartsPair.Type)
		assert.Equal(t, 2, len(multiPartsPair.AIMessage.Parts))
		assert.True(t, multiPartsPair.IsValid())

		// Negative case: ToolCall without FunctionCall
		aiMsgInvalidToolCall := &llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "invalid-1",
					Type: "function",
					// FunctionCall is nil
				},
			},
		}

		invalidToolCallPair := NewBodyPair(aiMsgInvalidToolCall, nil)
		assert.NotNil(t, invalidToolCallPair)
		assert.Equal(t, Completion, invalidToolCallPair.Type) // Should default to Completion

		// Verify the invalid tool call was removed
		foundToolCall := false
		for _, part := range invalidToolCallPair.AIMessage.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				foundToolCall = true
				break
			}
		}
		assert.False(t, foundToolCall, "Invalid tool call should be removed")
	})

	// Test cases for NewBodyPairFromMessages
	t.Run("NewBodyPairFromMessages", func(t *testing.T) {
		// Positive case: Valid AI + Tool messages
		messages := []llms.MessageContent{
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

		bodyPair, err := NewBodyPairFromMessages(messages)
		assert.NoError(t, err)
		assert.NotNil(t, bodyPair)
		assert.Equal(t, RequestResponse, bodyPair.Type)
		assert.Equal(t, 1, len(bodyPair.ToolMessages))
		assert.True(t, bodyPair.IsValid())

		// Check GetToolCallsInfo
		toolCallsInfo := bodyPair.GetToolCallsInfo()
		assert.Empty(t, toolCallsInfo.PendingToolCallIDs, "Should have no pending tool calls")
		assert.Empty(t, toolCallsInfo.UnmatchedToolCallIDs, "Should have no unmatched tool calls")
		assert.Equal(t, 1, len(toolCallsInfo.CompletedToolCalls), "Should have one completed tool call")

		// Positive case: AI with multiple tool calls and their responses
		multiToolMessages := []llms.MessageContent{
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

		multiToolPair, err := NewBodyPairFromMessages(multiToolMessages)
		assert.NoError(t, err)
		assert.NotNil(t, multiToolPair)
		assert.Equal(t, RequestResponse, multiToolPair.Type)
		assert.Equal(t, 2, len(multiToolPair.ToolMessages))
		assert.True(t, multiToolPair.IsValid())

		// Positive case: AI completion (no tool calls)
		completionMessages := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextContent{Text: "Simple text response"}},
			},
		}

		completionPair, err := NewBodyPairFromMessages(completionMessages)
		assert.NoError(t, err)
		assert.NotNil(t, completionPair)
		assert.Equal(t, Completion, completionPair.Type)
		assert.Empty(t, completionPair.ToolMessages)
		assert.True(t, completionPair.IsValid())

		// Negative case: Empty messages
		_, err = NewBodyPairFromMessages([]llms.MessageContent{})
		assert.Error(t, err)
		t.Logf("Got expected error for empty messages: %v", err)

		// Negative case: First message not AI
		invalidMessages := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: "This should be an AI message"}},
			},
		}

		_, err = NewBodyPairFromMessages(invalidMessages)
		assert.Error(t, err)
		t.Logf("Got expected error for non-AI first message: %v", err)

		// Negative case: Non-tool message after AI
		invalidMessages = []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextContent{Text: "AI response"}},
			},
			{
				Role:  llms.ChatMessageTypeHuman, // Should be Tool
				Parts: []llms.ContentPart{llms.TextContent{Text: "This should be a tool message"}},
			},
		}

		_, err = NewBodyPairFromMessages(invalidMessages)
		assert.Error(t, err)
		t.Logf("Got expected error for non-tool message after AI: %v", err)
	})

	// Test cases for NewBodyPairFromSummarization
	t.Run("NewBodyPairFromSummarization", func(t *testing.T) {
		summarizationText := "This is a summary of the conversation about the weather in New York."

		bodyPair := NewBodyPairFromSummarization(summarizationText)
		assert.NotNil(t, bodyPair)
		assert.Equal(t, Summarization, bodyPair.Type)

		// Check AI message has correct tool call
		foundToolCall := false
		var toolCallID string
		for _, part := range bodyPair.AIMessage.Parts {
			if toolCall, ok := part.(llms.ToolCall); ok &&
				toolCall.FunctionCall != nil &&
				toolCall.FunctionCall.Name == SummarizationToolName {
				foundToolCall = true
				toolCallID = toolCall.ID
				assert.Equal(t, SummarizationToolArgs, toolCall.FunctionCall.Arguments)
				t.Logf("Found summarization tool call with ID %s", toolCallID)
				break
			}
		}
		assert.True(t, foundToolCall, "Summarization tool call not found")

		// Check tool message has correct response
		assert.Equal(t, 1, len(bodyPair.ToolMessages))
		foundResponse := false
		for _, part := range bodyPair.ToolMessages[0].Parts {
			if resp, ok := part.(llms.ToolCallResponse); ok {
				foundResponse = true
				assert.Equal(t, toolCallID, resp.ToolCallID)
				assert.Equal(t, SummarizationToolName, resp.Name)
				assert.Equal(t, summarizationText, resp.Content)
				t.Logf("Found summarization tool response with content: %s", resp.Content)
				break
			}
		}
		assert.True(t, foundResponse, "Summarization tool response not found")

		// Check validity and messages
		assert.True(t, bodyPair.IsValid())
		messages := bodyPair.Messages()
		assert.Equal(t, 2, len(messages))

		// Check GetToolCallsInfo
		toolCallsInfo := bodyPair.GetToolCallsInfo()
		assert.Empty(t, toolCallsInfo.PendingToolCallIDs)
		assert.Empty(t, toolCallsInfo.UnmatchedToolCallIDs)
		assert.Equal(t, 1, len(toolCallsInfo.CompletedToolCalls))

		// Test with empty text
		emptyTextPair := NewBodyPairFromSummarization("")
		assert.NotNil(t, emptyTextPair)
		assert.Equal(t, Summarization, emptyTextPair.Type)
		assert.True(t, emptyTextPair.IsValid())

		// Test the generated ID format
		foundValidID := false
		for _, part := range emptyTextPair.AIMessage.Parts {
			if toolCall, ok := part.(llms.ToolCall); ok {
				assert.True(t, strings.HasPrefix(toolCall.ID, "call_"),
					"Tool call ID should start with 'call_'")
				assert.Equal(t, 29, len(toolCall.ID),
					"Tool call ID should be 29 characters (call_ + 24 random chars)")
				foundValidID = true
				break
			}
		}
		assert.True(t, foundValidID, "Should find a valid tool call ID")
	})
}
