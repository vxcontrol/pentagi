package csum

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"pentagi/pkg/cast"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// SummarizerChecks contains text validation checks for text passed to summarizer
type SummarizerChecks struct {
	ExpectedStrings   []string // Strings that should be present in the text
	UnexpectedStrings []string // Strings that should not be present in the text
	ExpectedCallCount int      // Number of times the summarizer is expected to be called
}

// Helper function to create new text message
func newTextMsg(role llms.ChatMessageType, text string) *llms.MessageContent {
	return &llms.MessageContent{
		Role:  role,
		Parts: []llms.ContentPart{llms.TextContent{Text: text}},
	}
}

// Helper function to create a Chain AST for testing
func createTestChainAST(sections ...*cast.ChainSection) *cast.ChainAST {
	return &cast.ChainAST{Sections: sections}
}

// verifyASTConsistency performs comprehensive validation of the AST structure
// to ensure it remains valid after operations
func verifyASTConsistency(t *testing.T, ast *cast.ChainAST) {
	// Check that the AST is not nil
	assert.NotNil(t, ast, "AST should not be nil")

	// 1. Check headers in sections
	for i, section := range ast.Sections {
		if i == 0 {
			// First section can have system message, human message, or both
			if section.Header.SystemMessage == nil && section.Header.HumanMessage == nil {
				t.Errorf("First section header cannot have both system and human messages be nil")
			}
		} else {
			// Non-first sections should not have system messages
			assert.Nil(t, section.Header.SystemMessage,
				fmt.Sprintf("Section %d should not have system message", i))

			// Non-first sections should have human messages
			assert.NotNil(t, section.Header.HumanMessage,
				fmt.Sprintf("Section %d should have human message", i))
		}
	}

	// 2. Check body pairs in sections
	for i, section := range ast.Sections {
		if i < len(ast.Sections)-1 && len(section.Body) == 0 {
			t.Errorf("Section %d (not last) must have non-empty body pairs", i)
		}

		// Check each body pair
		for j, pair := range section.Body {
			switch pair.Type {
			case cast.RequestResponse, cast.Summarization:
				// Check that each tool call has a response
				toolCallCount := countToolCalls(pair.AIMessage)
				responseCount := countToolResponses(pair.ToolMessages)

				if toolCallCount > 0 && len(pair.ToolMessages) == 0 {
					t.Errorf("Section %d, BodyPair %d: RequestResponse has tool calls but no responses", i, j)
				}

				if toolCallCount != responseCount {
					t.Errorf("Section %d, BodyPair %d: Tool call count (%d) doesn't match response count (%d)",
						i, j, toolCallCount, responseCount)
				}
			case cast.Completion:
				// Completion pairs shouldn't have tool calls or tool messages
				if pair.AIMessage == nil {
					t.Errorf("Section %d, BodyPair %d: Completion pair has nil AIMessage", i, j)
				} else if hasToolCalls(pair.AIMessage) {
					t.Errorf("Section %d, BodyPair %d: Completion pair contains tool calls", i, j)
				}

				if len(pair.ToolMessages) > 0 {
					t.Errorf("Section %d, BodyPair %d: Completion pair has non-empty ToolMessages", i, j)
				}
			default:
				t.Errorf("Section %d, BodyPair %d: Unexpected pair type %d", i, j, pair.Type)
			}
		}
	}

	// 3. Check size calculation
	verifyASTSizes(t, ast)

	// 4. Check that the AST can be converted to messages and back
	messages := ast.Messages()
	newAST, err := cast.NewChainAST(messages, false)
	if err != nil {
		t.Errorf("Failed to create AST from messages: %v", err)
	} else {
		newMessages := newAST.Messages()

		// Convert both message lists to JSON for comparison
		origJSON, _ := json.Marshal(messages)
		newJSON, _ := json.Marshal(newMessages)

		if string(origJSON) != string(newJSON) {
			t.Errorf("Messages from new AST don't match original messages")
		}
	}
}

// verifyASTSizes validates that sizes are calculated correctly throughout the AST
func verifyASTSizes(t *testing.T, ast *cast.ChainAST) {
	// Check AST total size
	expectedTotalSize := 0
	for _, section := range ast.Sections {
		expectedTotalSize += section.Size()
	}
	assert.Equal(t, expectedTotalSize, ast.Size(), "AST size should equal sum of section sizes")

	// Check section sizes
	for i, section := range ast.Sections {
		expectedSectionSize := section.Header.Size()
		for _, pair := range section.Body {
			expectedSectionSize += pair.Size()
		}
		assert.Equal(t, expectedSectionSize, section.Size(),
			fmt.Sprintf("Section %d size should equal header size plus sum of body pair sizes", i))
	}
}

// Create a mock summarizer for testing with validation
type mockSummarizer struct {
	expectedMessages []llms.MessageContent
	returnText       string
	returnError      error
	called           bool
	callCount        int
	checksPerformed  bool
	checks           *SummarizerChecks
	receivedTexts    []string // Store all received texts for validation
}

func newMockSummarizer(returnText string, returnError error, checks *SummarizerChecks) *mockSummarizer {
	return &mockSummarizer{
		returnText:    returnText,
		returnError:   returnError,
		checks:        checks,
		receivedTexts: []string{},
	}
}

// Summarize implements the mock summarizer function with validation
func (m *mockSummarizer) Summarize(ctx context.Context, text string) (string, error) {
	m.called = true
	m.callCount++
	m.receivedTexts = append(m.receivedTexts, text)

	// Store basic check status - actual validation happens in ValidateChecks
	if m.checks != nil {
		m.checksPerformed = true
	}

	return m.returnText, m.returnError
}

// ValidateChecks validates that at least one received text contains each expected string
// and no received text contains any unexpected string
func (m *mockSummarizer) ValidateChecks(t *testing.T) {
	if m.checks == nil || !m.checksPerformed {
		return
	}

	// Check for expected strings - must be present in any text
	for _, expected := range m.checks.ExpectedStrings {
		found := false
		for _, text := range m.receivedTexts {
			if strings.Contains(text, expected) {
				found = true
				break
			}
		}
		assert.True(t, found, fmt.Sprintf("Expected string '%s' not found in any text passed to summarizer", expected))
	}

	// Check for unexpected strings - must not be present in any text
	for _, unexpected := range m.checks.UnexpectedStrings {
		for _, text := range m.receivedTexts {
			assert.False(t, strings.Contains(text, unexpected),
				fmt.Sprintf("Unexpected string '%s' found in text passed to summarizer", unexpected))
		}
	}

	// Check expected call count if provided
	if m.checks.ExpectedCallCount > 0 {
		assert.Equal(t, m.checks.ExpectedCallCount, m.callCount, "Summarizer call count doesn't match expected")
	}
}

// SummarizerHandler returns the Summarize function as a tools.SummarizeHandler
func (m *mockSummarizer) SummarizerHandler() tools.SummarizeHandler {
	return m.Summarize
}

// createMockSummarizeHandler creates a simple mock handler for testing
func createMockSummarizeHandler() tools.SummarizeHandler {
	return newMockSummarizer("Summarized content", nil, nil).SummarizerHandler()
}

// Helper to count summarized pairs in a section
func countSummarizedPairs(section *cast.ChainSection) int {
	count := 0
	for _, pair := range section.Body {
		if containsSummarizedContent(pair) {
			count++
		}
	}
	return count
}

// toString converts any value to a string
func toString(t *testing.T, st any) string {
	str, err := json.Marshal(st)
	assert.NoError(t, err, "Failed to marshal to string")
	return string(str)
}

// compareMessages compares two message slices by converting to JSON
func compareMessages(t *testing.T, expected, actual []llms.MessageContent) {
	expectedJSON, err := json.Marshal(expected)
	assert.NoError(t, err, "Failed to marshal expected messages")

	actualJSON, err := json.Marshal(actual)
	assert.NoError(t, err, "Failed to marshal actual messages")

	assert.Equal(t, string(expectedJSON), string(actualJSON), "Messages differ")
}

// countToolCalls counts the number of tool calls in a message
func countToolCalls(msg *llms.MessageContent) int {
	if msg == nil {
		return 0
	}

	count := 0
	for _, part := range msg.Parts {
		if _, isToolCall := part.(llms.ToolCall); isToolCall {
			count++
		}
	}
	return count
}

// countToolResponses counts the number of tool responses in a slice of messages
func countToolResponses(messages []*llms.MessageContent) int {
	count := 0
	for _, msg := range messages {
		if msg == nil {
			continue
		}

		for _, part := range msg.Parts {
			if _, isResponse := part.(llms.ToolCallResponse); isResponse {
				count++
			}
		}
	}
	return count
}

// hasToolCalls checks if a message contains tool calls
func hasToolCalls(msg *llms.MessageContent) bool {
	return countToolCalls(msg) > 0
}

// verifySummarizationPatterns checks that the summarized sections have proper content
func verifySummarizationPatterns(t *testing.T, ast *cast.ChainAST, summarizationType string, keepQASections int) {
	// Skip empty ASTs
	if len(ast.Sections) == 0 {
		return
	}

	switch summarizationType {
	case "section":
		// In section summarization, all sections except the last one should have exactly one Summarization body pair
		for i, section := range ast.Sections {
			if i < len(ast.Sections)-keepQASections {
				if len(section.Body) != 1 {
					t.Errorf("Section %d should have exactly one body pair after section summarization", i)
				} else if section.Body[0].Type != cast.Summarization && section.Body[0].Type != cast.Completion {
					t.Errorf("Section %d should have Summarization or Completion type body pair after section summarization", i)
				}
			}
		}
	case "lastSection":
		// Last section should have at least one summarized body pair
		if len(ast.Sections) > 0 {
			lastSection := ast.Sections[len(ast.Sections)-1]
			if len(lastSection.Body) > 0 {
				// At least one pair should be summarized
				summarizedCount := countSummarizedPairs(lastSection)
				assert.Greater(t, summarizedCount, 0, "Last section should have at least one summarized pair")
			}
		}
	case "qaPair":
		// First section should have summarized QA content
		if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
			assert.True(t, containsSummarizedContent(ast.Sections[0].Body[0]),
				"First section should contain QA summarized content")
		}
	}
}

// verifySizeReduction checks that summarization reduces size of the AST
func verifySizeReduction(t *testing.T, originalSize int, ast *cast.ChainAST) {
	// Only check if original size is significant
	if originalSize > 1000 {
		assert.Less(t, ast.Size(), originalSize, "Summarization should reduce the overall size")
	}
}

// TestSummarizeSections tests the summarizeSections function
func TestSummarizeSections(t *testing.T) {
	ctx := context.Background()
	// Test cases
	tests := []struct {
		name               string
		sections           []*cast.ChainSection
		summarizerChecks   *SummarizerChecks
		returnText         string
		returnError        error
		expectedNoChange   bool
		expectedErrorCheck func(error) bool
		keepQASections     int
	}{
		{
			// Test with empty chain (0 sections) - should return without changes
			name:             "Empty chain",
			sections:         []*cast.ChainSection{},
			returnText:       "Summarized content",
			expectedNoChange: true,
			keepQASections:   keepMinLastQASections,
		},
		{
			// Test with one section - should return without changes
			name: "One section only",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Human message"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("AI response"),
					},
				),
			},
			returnText:       "Summarized content",
			expectedNoChange: true,
			keepQASections:   keepMinLastQASections,
		},
		{
			// Test with multiple sections, but all non-last sections already have only one Completion body pair
			name: "Sections already correctly summarized",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 1")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(SummarizedContentPrefix + "Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromSummarization("Answer 2", cast.ToolCallIDTemplate),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
						cast.NewBodyPairFromCompletion("Answer 3 continued"),
					},
				),
			},
			returnText:       "Summarized content",
			expectedNoChange: true,
			keepQASections:   keepMinLastQASections,
		},
		{
			// Test with multiple sections, some with multiple pairs or RequestResponse pairs
			name: "Sections needing summarization",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1a"),
						cast.NewBodyPairFromCompletion("Answer 1b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						// Create a valid RequestResponse BodyPair with proper tool call and response
						func() *cast.BodyPair {
							aiMsg := &llms.MessageContent{
								Role: llms.ChatMessageTypeAI,
								Parts: []llms.ContentPart{
									llms.TextContent{Text: "Let me search"},
									llms.ToolCall{
										ID:   "search-tool-1",
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
										ToolCallID: "search-tool-1",
										Name:       "search",
										Content:    "Search results",
									},
								},
							}
							return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
						}(),
						cast.NewBodyPairFromCompletion("Based on the search, here's my answer"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Follow-up question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Final answer"),
					},
				),
			},
			summarizerChecks: &SummarizerChecks{
				// First call should be for first section
				ExpectedStrings: []string{"Answer 1a", "Answer 1b"},
				// Second call should be for second section with tool call
				UnexpectedStrings: []string{"Final answer"},
				ExpectedCallCount: 2,
			},
			returnText:       "Summarized content",
			expectedNoChange: false,
			keepQASections:   keepMinLastQASections,
		},
		{
			// Test with summarizer returning error
			name: "Summarizer error",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 1")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1a"),
						cast.NewBodyPairFromCompletion("Answer 1b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
			},
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings: []string{"Answer 1a", "Answer 1b"}, // Should be summarizing first section
			},
			returnText:  "Shouldn't be used due to error",
			returnError: fmt.Errorf("summarizer error"),
			expectedErrorCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "summary generation failed")
			},
			keepQASections: keepMinLastQASections,
		},
		{
			// Test with keepQASections=2 - should keep the last 2 sections unchanged
			name: "Keep last 2 QA sections",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 1")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1a"),
						cast.NewBodyPairFromCompletion("Answer 1b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2a"),
						cast.NewBodyPairFromCompletion("Answer 2b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
						cast.NewBodyPairFromCompletion("Answer 3 continued"),
					},
				),
			},
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Answer 1a", "Answer 1b"}, // Should summarize only the first section
				UnexpectedStrings: []string{},                         // No unexpected strings to check
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized content",
			expectedNoChange: false,
			keepQASections:   2, // Keep the last 2 sections
		},
		{
			// Test with keepQASections=3 - should not summarize any sections because there are only 3
			name: "Keep all 3 QA sections",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 1")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1a"),
						cast.NewBodyPairFromCompletion("Answer 1b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
					},
				),
			},
			returnText:       "Summarized content",
			expectedNoChange: true, // No changes expected as we're keeping all sections
			keepQASections:   3,    // Keep all 3 sections
		},
		{
			// Test with keepQASections being larger than the number of sections
			name: "keepQASections larger than number of sections",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 1")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2a"),
						cast.NewBodyPairFromCompletion("Answer 2b"),
					},
				),
			},
			returnText:       "Shouldn't be used",
			expectedNoChange: true, // No changes when keepQASections > section count
			keepQASections:   5,    // More than the number of sections
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test AST
			ast := createTestChainAST(tt.sections...)

			// Verify initial AST consistency
			verifyASTConsistency(t, ast)

			// Save original messages and AST for comparison
			originalMessages := ast.Messages()
			originalMessagesString := toString(t, originalMessages)
			originalSize := ast.Size()
			originalASTString := toString(t, ast)

			// Create mock summarizer
			mockSum := newMockSummarizer(tt.returnText, tt.returnError, tt.summarizerChecks)

			// Call the function with keepQASections parameter
			err := summarizeSections(ctx, ast, mockSum.SummarizerHandler(), tt.keepQASections, cast.ToolCallIDTemplate)

			// Check error if expected
			if tt.expectedErrorCheck != nil {
				assert.True(t, tt.expectedErrorCheck(err), "Error does not match expected check")
				return
			} else {
				assert.NoError(t, err)
			}

			// Verify AST consistency after operations
			verifyASTConsistency(t, ast)

			// Check changes
			if tt.expectedNoChange {
				// Messages and AST should be the same
				messages := ast.Messages()
				compareMessages(t, originalMessages, messages)
				assert.Equal(t, originalMessagesString, toString(t, messages),
					"Messages should not change")
				assert.Equal(t, originalASTString, toString(t, ast),
					"AST should not change")

				// Check if summarizer was called (it shouldn't have been if no changes needed)
				assert.False(t, mockSum.called, "Summarizer should not have been called")
			} else {
				// Check if sections were properly summarized
				for i := 0; i < len(ast.Sections)-tt.keepQASections; i++ {
					assert.Equal(t, 1, len(ast.Sections[i].Body),
						fmt.Sprintf("Section %d should have exactly one body pair", i))

					// The sections should now be of type Summarization, not Completion
					bodyType := ast.Sections[i].Body[0].Type
					assert.True(t, bodyType == cast.Summarization || bodyType == cast.Completion,
						fmt.Sprintf("Section %d should have Summarization or Completion type body pair after section summarization", i))
				}

				// Verify summarizer was called and checks performed
				assert.True(t, mockSum.called, "Summarizer should have been called")
				if tt.summarizerChecks != nil {
					// Validate all checks after all summarizer calls are completed
					mockSum.ValidateChecks(t)
				}

				// Verify summarization patterns
				verifySummarizationPatterns(t, ast, "section", tt.keepQASections)

				// Verify size reduction if applicable
				verifySizeReduction(t, originalSize, ast)
			}

			assert.Equal(t, len(ast.Sections), len(tt.sections), "Number of sections should be the same")

			// Last keepQASections should not be modified
			if len(ast.Sections) > 0 && len(ast.Sections) == len(tt.sections) {
				l := len(ast.Sections)
				for i := l - 1; i >= 0 && i >= l-tt.keepQASections; i-- {
					lastOriginal := tt.sections[i]
					lastCurrent := ast.Sections[i]

					assert.Equal(t, len(lastOriginal.Body), len(lastCurrent.Body),
						fmt.Sprintf("Section %d body pairs should not be modified due to keepQASections=%d",
							i, tt.keepQASections))
				}
			}
		})
	}
}

// TestSummarizeLastSection tests the summarizeLastSection function
func TestSummarizeLastSection(t *testing.T) {
	ctx := context.Background()

	// Test cases
	tests := []struct {
		name                 string
		sections             []*cast.ChainSection
		maxBytes             int
		maxBodyPairBytes     int
		reservePercent       int
		summarizerChecks     *SummarizerChecks
		returnText           string
		returnError          error
		expectedNoChange     bool
		expectedErrorCheck   func(error) bool
		expectedSummaryCheck func(*cast.ChainAST) bool
		skipSizeCheck        bool
	}{
		{
			// Test with empty chain - should return nil
			name:             "Empty chain",
			sections:         []*cast.ChainSection{},
			maxBytes:         1000,
			maxBodyPairBytes: 16 * 1024,
			returnText:       "Summarized content",
			expectedNoChange: true,
			reservePercent:   25, // Default
			skipSizeCheck:    false,
		},
		{
			// Test with section under size limit - should not trigger summarization
			name: "Section under size limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Test response"),
					},
				),
			},
			maxBytes:         1000, // Larger than the section size
			maxBodyPairBytes: 16 * 1024,
			returnText:       "Summarized content",
			expectedNoChange: true,
			reservePercent:   25, // Default
			skipSizeCheck:    false,
		},
		{
			// Test with section over size limit - should summarize oldest pairs
			name: "Section over size limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100) + "Response 1"), // Larger response
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100) + "Response 2"), // Larger response
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 100) + "Response 3"), // Larger response
						cast.NewBodyPairFromCompletion("Response 4"),                            // Small response that will be kept
					},
				),
			},
			maxBytes:         200, // Small enough to trigger summarization
			maxBodyPairBytes: 16 * 1024,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Response 1", "Response 2"},
				UnexpectedStrings: []string{"Response 4"}, // Last response should be kept
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized first responses",
			expectedNoChange: false,
			reservePercent:   25, // Default
			skipSizeCheck:    false,
		},
		{
			// Test with RequestResponse pairs when section exceeds limit
			// Should preserve tool calls in the summary
			name: "Section with RequestResponse pairs over limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						// Create a RequestResponse pair with tool call
						cast.NewBodyPair(
							&llms.MessageContent{
								Role: llms.ChatMessageTypeAI,
								Parts: []llms.ContentPart{
									llms.ToolCall{
										ID:   "test-id",
										Type: "function",
										FunctionCall: &llms.FunctionCall{
											Name:      "test_func",
											Arguments: `{"query": "test"}`,
										},
									},
								},
							},
							[]*llms.MessageContent{
								{
									Role: llms.ChatMessageTypeTool,
									Parts: []llms.ContentPart{
										llms.ToolCallResponse{
											ToolCallID: "test-id",
											Name:       "test_func",
											Content:    "Tool response",
										},
									},
								},
							},
						),
						// Add normal Completion pairs
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100) + "Response 1"), // Larger response
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100) + "Response 2"), // Larger response
						cast.NewBodyPairFromCompletion("Response 3"),                            // Small response that will be kept
					},
				),
			},
			maxBytes:         200, // Small enough to trigger summarization
			maxBodyPairBytes: 16 * 1024,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"test_func", "Tool response"},
				UnexpectedStrings: []string{"Response 3"}, // Last response should be kept
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized with tool calls",
			expectedNoChange: false,
			reservePercent:   25, // Default
			skipSizeCheck:    false,
		},
		{
			// Test with summarizer returning error
			name: "Summarizer error",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100) + "Response 1"), // Larger response
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100) + "Response 2"), // Larger response
						cast.NewBodyPairFromCompletion("Response 3"),                            // Small response
					},
				),
			},
			maxBytes:         200, // Small enough to trigger summarization
			maxBodyPairBytes: 16 * 1024,
			returnText:       "Won't be used due to error",
			returnError:      fmt.Errorf("summarizer error"),
			expectedErrorCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "last section summary generation failed")
			},
			reservePercent: 25, // Default
			skipSizeCheck:  false,
		},
		{
			// Test edge case - very large header, no body pairs
			name: "Large header, empty body",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, strings.Repeat("S", 150)), // Large system message
						newTextMsg(llms.ChatMessageTypeHuman, strings.Repeat("H", 150)),  // Large human message
					),
					[]*cast.BodyPair{},
				),
			},
			maxBytes:         200, // Smaller than header
			maxBodyPairBytes: 16 * 1024,
			returnText:       "Summarized content",
			expectedNoChange: true, // No body pairs to summarize
			reservePercent:   25,   // Default
			skipSizeCheck:    false,
		},
		{
			// Test for summarizing oversized individual body pairs before main summarization
			name: "Oversized individual body pairs",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question with large response")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Normal size answer"),
						cast.NewBodyPairFromCompletion(strings.Repeat("X", 20*1024)), // 20KB answer, exceeds maxBodyPairBytes
						cast.NewBodyPairFromCompletion("Another normal size answer"),
					},
				),
			},
			maxBytes:         50 * 1024, // Large enough to not trigger full section summarization
			maxBodyPairBytes: 16 * 1024, // Set to trigger only the oversized body pair
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"XXX"},            // Should contain text from the oversized answer
				UnexpectedStrings: []string{"Another normal"}, // Should not contain text from normal answers
				ExpectedCallCount: 1,                          // Called once for the single oversized pair
			},
			returnText:       "Summarized large response",
			expectedNoChange: false, // Should change the oversized pair only
			reservePercent:   25,    // Default
			skipSizeCheck:    false,
		},
		{
			// Test with lastSectionReservePercentage=0 (no reserve buffer)
			name: "No reserve buffer",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100) + "Response 1"),
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100) + "Response 2"),
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 100) + "Response 3"),
						cast.NewBodyPairFromCompletion(strings.Repeat("D", 100) + "Response 4"),
					},
				),
			},
			maxBytes:         200, // Reduced to ensure it triggers summarization
			maxBodyPairBytes: 16 * 1024,
			reservePercent:   0, // No reserve - should only summarize minimum needed
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Response 1"},
				UnexpectedStrings: []string{"Response 4"}, // Last response should be kept
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized first response",
			expectedNoChange: false,
			skipSizeCheck:    false,
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]
				// With 0% reserve, we should keep most messages and summarize fewer
				return len(lastSection.Body) == 2 && // 1 summary + 1 kept message (the last one)
					(lastSection.Body[0].Type == cast.Summarization || lastSection.Body[0].Type == cast.Completion)
			},
		},
		{
			// Test with lastSectionReservePercentage=50 (large reserve buffer)
			name: "Large reserve buffer",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100) + "Response 1"),
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100) + "Response 2"),
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 100) + "Response 3"),
						cast.NewBodyPairFromCompletion(strings.Repeat("D", 100) + "Response 4"),
					},
				),
			},
			maxBytes:         200, // Reduced to ensure it triggers summarization
			maxBodyPairBytes: 16 * 1024,
			reservePercent:   50, // Half reserved - should summarize more aggressively
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Response 1", "Response 2", "Response 3"},
				UnexpectedStrings: []string{"Response 4"}, // Last response should be kept
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized first three responses",
			expectedNoChange: false,
			skipSizeCheck:    false,
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]
				// With 50% reserve, we should have primarily summary and few kept messages
				return len(lastSection.Body) == 2 && // 1 summary + 1 kept message (the last one)
					(lastSection.Body[0].Type == cast.Summarization || lastSection.Body[0].Type == cast.Completion)
			},
		},
		{
			// Test with reservePercent = 100% (maximum reserve)
			name: "Maximum reserve buffer",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question with multiple responses")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 50) + "First response"),
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 50) + "Second response"),
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 50) + "Third response"),
						cast.NewBodyPairFromCompletion(strings.Repeat("D", 50) + "Fourth response"),
						cast.NewBodyPairFromCompletion("Fifth response - this should be the only one kept"),
					},
				),
			},
			maxBytes:         300, // Set this so section will exceed it and trigger summarization
			maxBodyPairBytes: 16 * 1024,
			reservePercent:   100, // Maximum reserve - should summarize everything except the last message
			summarizerChecks: &SummarizerChecks{
				// Should summarize all earlier responses
				ExpectedStrings: []string{"First", "Second", "Third", "Fourth"},
				// Should not summarize the last response
				UnexpectedStrings: []string{"Fifth response"},
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized all but the last response",
			expectedNoChange: false,
			skipSizeCheck:    false,
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]

				// With 100% reserve, there should be exactly 2 body parts:
				// 1. The summary of all previous messages
				// 2. Only the very last message
				if len(lastSection.Body) != 2 {
					return false
				}

				// Check first part is a summary
				if !containsSummarizedContent(lastSection.Body[0]) {
					return false
				}

				// Check second part is the last message
				content, ok := lastSection.Body[1].AIMessage.Parts[0].(llms.TextContent)
				return ok && strings.Contains(content.Text, "Fifth response")
			},
		},
		{
			// Test with already summarized content exceeding maxBodyPairBytes
			name: "Already summarized large content should not be re-summarized",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question")),
					[]*cast.BodyPair{
						// Create a pair with already summarized but large content
						func() *cast.BodyPair {
							return cast.NewBodyPairFromSummarization(strings.Repeat("S", 20*1024), cast.ToolCallIDTemplate)
						}(),
						cast.NewBodyPairFromCompletion("Normal response"),
					},
				),
			},
			maxBytes:         10 * 1024, // Small enough to potentially trigger summarization
			maxBodyPairBytes: 16 * 1024, // The summarized content exceeds this
			returnText:       "This should not be used",
			expectedNoChange: true, // No change should occur due to the content already being summarized
			reservePercent:   25,   // Default
			skipSizeCheck:    true, // Skip size check as already summarized content may exceed the limit
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]
				// Check the content directly
				if len(lastSection.Body) != 2 {
					return false
				}

				// Check the first pair for summarized content prefix
				if lastSection.Body[0].AIMessage == nil || len(lastSection.Body[0].AIMessage.Parts) == 0 {
					return false
				}

				return containsSummarizedContent(lastSection.Body[0])
			},
		},
		{
			// Test where total content exceeds maxBytes but single pairs don't exceed maxBodyPairBytes
			name: "Many small pairs exceeding section limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Test question with many small responses")),
					func() []*cast.BodyPair {
						// Create 10 small body pairs that collectively exceed the limit
						pairs := make([]*cast.BodyPair, 20) // Increase to 20 pairs
						for i := 0; i < 20; i++ {
							// Make each response slightly larger
							pairs[i] = cast.NewBodyPairFromCompletion(fmt.Sprintf("%s Small response %d", strings.Repeat("X", 20), i))
						}
						return pairs
					}(),
				),
			},
			maxBytes:         100, // Reduced to ensure triggering summarization
			maxBodyPairBytes: 16 * 1024,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"X Small response 0", "X Small response 1"},
				UnexpectedStrings: []string{"X Small response 19"}, // Last response should be kept
				ExpectedCallCount: 1,
			},
			returnText:       "Summarized small responses",
			expectedNoChange: false,
			reservePercent:   25,   // Default
			skipSizeCheck:    true, // Skip size check as the size may vary depending on summarization
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]
				// Should have summarized early messages but kept later ones
				return containsSummarizedContent(lastSection.Body[0]) &&
					strings.Contains(toString(t, lastSection.Body[len(lastSection.Body)-1]), "X Small response 19")
			},
		},
		{
			// Test where the summarizer returns a large summary
			name: "Large summary returned",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question with large summary")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 50) + "Response 1"),
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 50) + "Response 2"),
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 50) + "Response 3"),
					},
				),
			},
			maxBytes:         200, // Small size to trigger summarization
			maxBodyPairBytes: 16 * 1024,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Response"}, // Just check for any response content
				ExpectedCallCount: 1,
			},
			returnText:       strings.Repeat("X", 300) + "Very large summary", // Summary larger than original content
			expectedNoChange: false,
			reservePercent:   25,   // Default
			skipSizeCheck:    true, // Skip size check because the summarizer returns a very large result
			expectedSummaryCheck: func(ast *cast.ChainAST) bool {
				if len(ast.Sections) == 0 {
					return false
				}
				lastSection := ast.Sections[len(ast.Sections)-1]
				// Should have the large summary at the beginning
				return len(lastSection.Body) > 0 &&
					containsSummarizedContent(lastSection.Body[0]) &&
					strings.Contains(toString(t, lastSection.Body[0]), "Very large summary")
			},
		},
		{
			// Test with exactly one body pair that is not oversized - no summarization needed
			name: "Single body pair under size limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Simple question")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Single response"),
					},
				),
			},
			maxBytes:         5000, // Much larger than content
			maxBodyPairBytes: 16 * 1024,
			returnText:       "Shouldn't be used",
			expectedNoChange: true,
			reservePercent:   25, // Default
			skipSizeCheck:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test AST
			ast := createTestChainAST(tt.sections...)

			// Verify initial AST consistency
			verifyASTConsistency(t, ast)

			// Get original messages and AST for comparison
			originalMessages := ast.Messages()
			originalMessagesString := toString(t, originalMessages)
			originalSize := ast.Size()
			originalASTString := toString(t, ast)

			// Create mock summarizer
			mockSum := newMockSummarizer(tt.returnText, tt.returnError, tt.summarizerChecks)

			// Call summarizeLastSection with the correct arguments, including reserve percent
			var err error
			err = summarizeLastSection(ctx, ast, mockSum.SummarizerHandler(),
				len(ast.Sections)-1, tt.maxBytes, tt.maxBodyPairBytes, tt.reservePercent, cast.ToolCallIDTemplate)

			// Check error if expected
			if tt.expectedErrorCheck != nil {
				assert.True(t, tt.expectedErrorCheck(err), "Error does not match expected check")
				return
			} else {
				assert.NoError(t, err)
			}

			// Verify AST consistency after operations
			verifyASTConsistency(t, ast)

			// Skip further checks if empty chain
			if len(ast.Sections) == 0 {
				return
			}

			// Get the last section after processing
			lastSection := ast.Sections[len(ast.Sections)-1]

			if tt.expectedNoChange {
				// Messages and AST should be the same
				messages := ast.Messages()
				compareMessages(t, originalMessages, messages)
				assert.Equal(t, originalMessagesString, toString(t, messages),
					"Messages should not change")
				assert.Equal(t, originalASTString, toString(t, ast),
					"AST should not change")

				// Check if summarizer was called (it shouldn't have been if no changes needed)
				assert.False(t, mockSum.called, "Summarizer should not have been called")
			} else {
				// There should be body pairs after processing
				assert.Greater(t, len(lastSection.Body), 0, "Last section should have body pairs")

				// Check if the summarizer was called
				assert.True(t, mockSum.called, "Summarizer should have been called")

				// At least one body pair should have summarized content
				summarizedCount := countSummarizedPairs(lastSection)
				assert.Greater(t, summarizedCount, 0, "At least one body pair should contain summarized content")

				// Last section size should be within limits, except for tests with large summaries
				// where we know the limit might be exceeded
				if !tt.skipSizeCheck {
					// Use a more flexible check with buffer for summarization overhead
					// The summarization might add some overhead, but generally should be close to the limit
					// Allow up to 100% overhead since summarization tool responses can be larger than original content
					maxAllowedSize := tt.maxBytes + summarizedCount*250 // 250 is the average size of a tool call
					assert.LessOrEqual(t, lastSection.Size(), maxAllowedSize,
						"Last section size should be within a reasonable range of the specified limit")
				}

				// Verify summarization patterns
				verifySummarizationPatterns(t, ast, "lastSection", 1)

				// Verify that summarizer checks were performed
				if tt.summarizerChecks != nil {
					// Validate all checks after all summarizer calls are completed
					mockSum.ValidateChecks(t)
				}

				// Verify size reduction if applicable
				if tt.returnError == nil {
					verifySizeReduction(t, originalSize, ast)
				}
			}

			// Run additional structure checks if provided
			if tt.expectedSummaryCheck != nil {
				assert.True(t, tt.expectedSummaryCheck(ast), "AST structure does not match expected")
			}

			// If this was the oversized body pair test, check that only the oversized pair was summarized
			if tt.name == "Oversized individual body pairs" && !tt.expectedNoChange {
				lastSection := ast.Sections[len(ast.Sections)-1]

				// Check the first pair is unchanged
				assert.Contains(t, toString(t, lastSection.Body[0]), "Normal size answer",
					"First normal-sized pair should be unchanged")

				// Check the second (oversized) pair was summarized
				assert.True(t, lastSection.Body[1].Type == cast.Summarization || lastSection.Body[1].Type == cast.Completion,
					"Oversized pair should be summarized as Summarization or Completion")

				// Check the third pair is unchanged
				assert.Contains(t, toString(t, lastSection.Body[2]), "Another normal size answer",
					"Last normal-sized pair should be unchanged")
			}
		})
	}
}

// TestSummarizeQAPairs tests the summarizeQAPairs function
func TestSummarizeQAPairs(t *testing.T) {
	ctx := context.Background()
	// Test cases
	tests := []struct {
		name                string
		sections            []*cast.ChainSection
		maxSections         int
		maxBytes            int
		summarizeHuman      bool
		summarizerChecks    *SummarizerChecks
		returnText          string
		returnError         error
		expectedNoChange    bool
		expectedErrorCheck  func(error) bool
		expectedQAPairCheck func(*cast.ChainAST) bool
	}{
		{
			// Test with empty chain - should return without changes
			name:             "Empty chain",
			sections:         []*cast.ChainSection{},
			maxSections:      5,
			maxBytes:         1000,
			summarizeHuman:   false,
			returnText:       "Summarized QA content",
			expectedNoChange: true,
		},
		{
			// Test with QA sections under count limit - should return without changes
			name: "Under QA section count limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
			},
			maxSections:      5,    // Limit higher than current sections
			maxBytes:         1000, // Limit higher than current size
			summarizeHuman:   false,
			returnText:       "Summarized QA content",
			expectedNoChange: true,
		},
		{
			// Test with QA sections over count limit - should summarize oldest sections
			name: "Over QA section count limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 4")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 4"),
					},
				),
			},
			maxSections:    2,    // Limit lower than current sections
			maxBytes:       1000, // Limit higher than current size
			summarizeHuman: false,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Answer 1", "Answer 2"}, // Should summarize older sections
				ExpectedCallCount: 1,                                // One call to summarize older sections
			},
			returnText:       "Summarized QA content",
			expectedNoChange: false,
			expectedQAPairCheck: func(ast *cast.ChainAST) bool {
				// Just check that we have a summary section and some sections
				return len(ast.Sections) > 0 && containsSummarizedContent(ast.Sections[0].Body[0])
			},
		},
		{
			// Test with QA sections over byte limit - should summarize oldest sections
			name: "Over QA byte limit",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 200)), // Large answer
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 200)), // Large answer
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Short answer 3"),
					},
				),
			},
			maxSections:    10,  // Limit higher than current sections
			maxBytes:       400, // Limit lower than total size
			summarizeHuman: false,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"AAA"}, // Should include content from first section
				ExpectedCallCount: 1,               // One call to summarize over-sized sections
			},
			returnText:       "Summarized QA content",
			expectedNoChange: false,
			expectedQAPairCheck: func(ast *cast.ChainAST) bool {
				// Just check that we have a summary section and some sections
				return len(ast.Sections) > 0 && containsSummarizedContent(ast.Sections[0].Body[0])
			},
		},
		{
			// Test with both limits exceeded
			name: "Both limits exceeded",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 100)),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("B", 100)),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("C", 100)),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 4")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("D", 100)),
					},
				),
			},
			maxSections:    2,   // Limit lower than current sections
			maxBytes:       300, // Limit lower than total size
			summarizeHuman: false,
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"AAA", "BBB"}, // Should include content from first sections
				ExpectedCallCount: 1,                      // One call to summarize excess sections
			},
			returnText:       "Summarized QA content",
			expectedNoChange: false,
			expectedQAPairCheck: func(ast *cast.ChainAST) bool {
				// Should have summary section with system message, plus last section only
				return len(ast.Sections) <= 3 && // At most 3 sections: summary + up to 2 kept sections
					containsSummarizedContent(ast.Sections[0].Body[0])
			},
		},
		{
			// Test with summarizeHuman = true vs false
			name: "Summarize humans test",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
					},
				),
			},
			maxSections:    1, // Force summarization of first two sections
			maxBytes:       1000,
			summarizeHuman: true, // Test with human summarization enabled
			summarizerChecks: &SummarizerChecks{
				ExpectedStrings:   []string{"Question 1", "Question 2"}, // Should include human messages
				ExpectedCallCount: 2,                                    // Calls to summarize sections (human and ai)
			},
			returnText:       "Summarized QA content with humans",
			expectedNoChange: false,
		},
		{
			// Test with summarizer returning error
			name: "Summarizer error",
			sections: []*cast.ChainSection{
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 2")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(nil, newTextMsg(llms.ChatMessageTypeHuman, "Question 3")),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3"),
					},
				),
			},
			maxSections: 1, // Force summarization to trigger error
			maxBytes:    1000,
			returnText:  "Won't be used due to error",
			returnError: fmt.Errorf("summarizer error"),
			expectedErrorCheck: func(err error) bool {
				return err != nil && strings.Contains(err.Error(), "QA (ai) summary generation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test AST
			ast := createTestChainAST(tt.sections...)

			// Verify initial AST consistency
			verifyASTConsistency(t, ast)

			// Record initial state for comparison
			originalSectionCount := len(ast.Sections)
			originalMessages := ast.Messages()
			originalMessagesString := toString(t, originalMessages)
			originalSize := ast.Size()
			originalASTString := toString(t, ast)

			// Create mock summarizer
			mockSum := newMockSummarizer(tt.returnText, tt.returnError, tt.summarizerChecks)

			// Call the function
			err := summarizeQAPairs(ctx, ast, mockSum.SummarizerHandler(),
				tt.maxSections, tt.maxBytes, tt.summarizeHuman, cast.ToolCallIDTemplate)

			// Check error if expected
			if tt.expectedErrorCheck != nil {
				assert.True(t, tt.expectedErrorCheck(err), "Error does not match expected check")
				return
			} else {
				assert.NoError(t, err)
			}

			// Verify AST consistency after operations
			verifyASTConsistency(t, ast)

			// Check for no change if expected
			if tt.expectedNoChange {
				assert.Equal(t, originalSectionCount, len(ast.Sections),
					"Section count should not change")

				// Messages and AST should be the same
				messages := ast.Messages()
				compareMessages(t, originalMessages, messages)
				assert.Equal(t, originalMessagesString, toString(t, messages),
					"Messages should not change")
				assert.Equal(t, originalASTString, toString(t, ast),
					"AST should not change")

				// Check if summarizer was called (it shouldn't have been if no changes needed)
				assert.False(t, mockSum.called, "Summarizer should not have been called")
			} else {
				// Verify summarizer was called and checks performed
				assert.True(t, mockSum.called, "Summarizer should have been called")
				if tt.summarizerChecks != nil {
					// Validate all checks after all summarizer calls are completed
					mockSum.ValidateChecks(t)
				}

				// Check if the resulting structure matches expected for QA summarization
				if tt.expectedQAPairCheck != nil {
					assert.True(t, tt.expectedQAPairCheck(ast),
						"Chain structure does not match expectations after QA summarization")
				}

				// First section should contain QA summarized content
				assert.Greater(t, len(ast.Sections), 0, "Should have at least one section")
				if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
					assert.True(t, containsSummarizedContent(ast.Sections[0].Body[0]),
						"First section should contain QA summarized content")
				}

				// Result should have sections under limits
				assert.LessOrEqual(t, len(ast.Sections), tt.maxSections+1, // +1 for summary section
					"Section count should be within limit after summarization")

				// Approximate size check - rebuilding would be more precise
				totalSize := 0
				for _, section := range ast.Sections {
					totalSize += section.Size()
				}
				assert.LessOrEqual(t, totalSize, tt.maxBytes+200, // Allow some overhead
					"Total size should be approximately within limits")

				// Verify summarization patterns
				verifySummarizationPatterns(t, ast, "qaPair", 1)

				// Verify size reduction if applicable
				verifySizeReduction(t, originalSize, ast)
			}
		})
	}
}

func TestLastBodyPairNeverSummarized(t *testing.T) {
	// This test ensures that the last body pair is NEVER summarized
	// to preserve reasoning signatures (critical for Gemini's thought_signature)

	ctx := context.Background()
	handler := createMockSummarizeHandler()

	// Create a section with multiple large body pairs
	// All pairs are oversized, but the last one should NOT be summarized
	largePair1 := createLargeBodyPair(20*1024, "Large response 1")
	largePair2 := createLargeBodyPair(20*1024, "Large response 2")
	largePair3 := createLargeBodyPair(20*1024, "Large response 3 - LAST")

	header := cast.NewHeader(
		newTextMsg(llms.ChatMessageTypeSystem, "System"),
		newTextMsg(llms.ChatMessageTypeHuman, "Question"),
	)
	section := cast.NewChainSection(header, []*cast.BodyPair{largePair1, largePair2, largePair3})

	// Verify initial state
	assert.Equal(t, 3, len(section.Body))
	initialLastPair := section.Body[2]
	assert.NotNil(t, initialLastPair)

	// Test 1: summarizeOversizedBodyPairs should NOT summarize the last pair
	err := summarizeOversizedBodyPairs(ctx, section, handler, 16*1024, cast.ToolCallIDTemplate)
	assert.NoError(t, err)

	// Verify the last pair was NOT summarized
	assert.Equal(t, 3, len(section.Body))
	lastPair := section.Body[2]
	assert.Equal(t, cast.RequestResponse, lastPair.Type, "Last pair type should remain RequestResponse")
	assert.False(t, containsSummarizedContent(lastPair), "Last pair should NOT be summarized")

	// Verify the first two pairs WERE summarized
	assert.True(t, containsSummarizedContent(section.Body[0]) || section.Body[0].Type == cast.Summarization,
		"First pair should be summarized")
	assert.True(t, containsSummarizedContent(section.Body[1]) || section.Body[1].Type == cast.Summarization,
		"Second pair should be summarized")

	// Test 2: Create AST and test summarizeLastSection
	ast := &cast.ChainAST{Sections: []*cast.ChainSection{section}}

	err = summarizeLastSection(ctx, ast, handler, 0, 30*1024, 16*1024, 25, cast.ToolCallIDTemplate)
	assert.NoError(t, err)

	// The last pair should still be preserved (not summarized)
	finalSection := ast.Sections[0]
	finalLastPair := finalSection.Body[len(finalSection.Body)-1]
	assert.Equal(t, cast.RequestResponse, finalLastPair.Type,
		"Last pair should remain RequestResponse after summarizeLastSection")
	assert.False(t, containsSummarizedContent(finalLastPair),
		"Last pair should NOT be summarized even after summarizeLastSection")
}

func TestLastBodyPairWithReasoning(t *testing.T) {
	// Test that last body pair with reasoning signatures is preserved
	// This covers both Gemini and Anthropic reasoning patterns

	tests := []struct {
		name        string
		createPair  func() *cast.BodyPair
		description string
		validate    func(t *testing.T, pair *cast.BodyPair)
	}{
		{
			name: "Gemini pattern - reasoning in ToolCall",
			createPair: func() *cast.BodyPair {
				// Gemini stores reasoning directly in ToolCall.Reasoning
				toolCallWithReasoning := llms.ToolCall{
					ID:   "fcall_test123gemini",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "execute_task_and_return_summary",
						Arguments: `{"question": "test"}`,
					},
					Reasoning: &reasoning.ContentReasoning{
						Content:   "Thinking about the task from Gemini",
						Signature: []byte("gemini_thought_signature_data"),
					},
				}

				aiMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Let me execute this"},
						toolCallWithReasoning,
					},
				}

				// Simulate large tool response (common scenario that triggers summarization)
				largeResponse := strings.Repeat("Data row: extensive output\n", 2000) // ~50KB

				toolMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "fcall_test123gemini",
							Name:       "execute_task_and_return_summary",
							Content:    largeResponse,
						},
					},
				}

				return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
			},
			description: "Gemini: reasoning in ToolCall with large response",
			validate: func(t *testing.T, pair *cast.BodyPair) {
				// Verify reasoning is still present in ToolCall
				for _, part := range pair.AIMessage.Parts {
					if toolCall, ok := part.(llms.ToolCall); ok && toolCall.FunctionCall != nil {
						assert.NotNil(t, toolCall.Reasoning, "Gemini ToolCall reasoning should be preserved")
						if toolCall.Reasoning != nil {
							assert.Equal(t, "Thinking about the task from Gemini", toolCall.Reasoning.Content)
							assert.Equal(t, []byte("gemini_thought_signature_data"), toolCall.Reasoning.Signature)
						}
					}
				}
			},
		},
		{
			name: "Anthropic pattern - reasoning in separate TextContent",
			createPair: func() *cast.BodyPair {
				// Anthropic stores reasoning in a separate TextContent BEFORE the ToolCall
				aiMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						// Reasoning comes first in a TextContent part
						llms.TextContent{
							Text: "", // Empty text, only reasoning
							Reasoning: &reasoning.ContentReasoning{
								Content:   "The data isn't reflected. Let me try examining send.php more carefully.",
								Signature: []byte("anthropic_crypto_signature_base64"),
							},
						},
						// Then the actual tool call WITHOUT reasoning in it
						llms.ToolCall{
							ID:   "toolu_011qigRrFEuu5dHKE78v3CuN",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "terminal",
								Arguments: `{"cwd":"/work","input":"curl -s http://example.com"}`,
							},
							// No Reasoning field here for Anthropic
						},
					},
				}

				// Large tool response
				largeResponse := strings.Repeat("Response data: ", 3000) // ~45KB

				toolMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "toolu_011qigRrFEuu5dHKE78v3CuN",
							Name:       "terminal",
							Content:    largeResponse,
						},
					},
				}

				return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
			},
			description: "Anthropic: reasoning in TextContent with large response",
			validate: func(t *testing.T, pair *cast.BodyPair) {
				// Verify reasoning is still present in TextContent
				foundReasoning := false
				for _, part := range pair.AIMessage.Parts {
					if textContent, ok := part.(llms.TextContent); ok && textContent.Reasoning != nil {
						foundReasoning = true
						assert.Equal(t, "The data isn't reflected. Let me try examining send.php more carefully.",
							textContent.Reasoning.Content)
						assert.Equal(t, []byte("anthropic_crypto_signature_base64"),
							textContent.Reasoning.Signature)
					}
				}
				assert.True(t, foundReasoning, "Anthropic TextContent reasoning should be preserved")

				// Verify tool call exists without reasoning
				foundToolCall := false
				for _, part := range pair.AIMessage.Parts {
					if toolCall, ok := part.(llms.ToolCall); ok && toolCall.FunctionCall != nil {
						foundToolCall = true
						assert.Nil(t, toolCall.Reasoning, "Anthropic ToolCall should not have reasoning")
					}
				}
				assert.True(t, foundToolCall, "Tool call should be present")
			},
		},
		{
			name: "Mixed pattern - both Gemini and Anthropic styles",
			createPair: func() *cast.BodyPair {
				// Some providers might use both patterns
				aiMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{
							Text: "Analyzing the situation",
							Reasoning: &reasoning.ContentReasoning{
								Content:   "Top-level reasoning",
								Signature: []byte("top_level_signature"),
							},
						},
						llms.ToolCall{
							ID:   "call_mixed123",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "analyze",
								Arguments: `{"data": "test"}`,
							},
							Reasoning: &reasoning.ContentReasoning{
								Content:   "Per-tool reasoning",
								Signature: []byte("tool_level_signature"),
							},
						},
					},
				}

				// Very large response to trigger size limits
				veryLargeResponse := strings.Repeat("Analysis result line\n", 5000) // ~100KB

				toolMsg := &llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "call_mixed123",
							Name:       "analyze",
							Content:    veryLargeResponse,
						},
					},
				}

				return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
			},
			description: "Mixed: both TextContent and ToolCall reasoning with very large response",
			validate: func(t *testing.T, pair *cast.BodyPair) {
				// Verify both reasoning types are preserved
				foundTextReasoning := false
				foundToolReasoning := false

				for _, part := range pair.AIMessage.Parts {
					if textContent, ok := part.(llms.TextContent); ok && textContent.Reasoning != nil {
						foundTextReasoning = true
						assert.NotNil(t, textContent.Reasoning.Signature)
					}
					if toolCall, ok := part.(llms.ToolCall); ok && toolCall.FunctionCall != nil && toolCall.Reasoning != nil {
						foundToolReasoning = true
						assert.NotNil(t, toolCall.Reasoning.Signature)
					}
				}

				assert.True(t, foundTextReasoning, "TextContent reasoning should be preserved")
				assert.True(t, foundToolReasoning, "ToolCall reasoning should be preserved")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			handler := createMockSummarizeHandler()

			t.Logf("Testing: %s", tt.description)

			lastPair := tt.createPair()

			// Create section with this as the ONLY pair (making it the last one)
			header := cast.NewHeader(
				newTextMsg(llms.ChatMessageTypeSystem, "System"),
				newTextMsg(llms.ChatMessageTypeHuman, "Question"),
			)
			section := cast.NewChainSection(header, []*cast.BodyPair{lastPair})

			initialSize := lastPair.Size()
			t.Logf("Initial last pair size: %d bytes", initialSize)

			// Test that the last pair is NOT summarized even if it's very large
			err := summarizeOversizedBodyPairs(ctx, section, handler, 16*1024, cast.ToolCallIDTemplate)
			assert.NoError(t, err)

			// Verify the pair was NOT summarized (because it's the last one)
			assert.Equal(t, 1, len(section.Body), "Should still have exactly one body pair")
			preservedPair := section.Body[0]

			// The type should remain the same (RequestResponse or Summarization)
			assert.Equal(t, lastPair.Type, preservedPair.Type,
				"Last pair type should not change when it's preserved")

			// If the original pair was already Summarization type, that's OK
			// What matters is that it wasn't RE-summarized (size should be the same)
			// For other types, it should not be converted to summarized content
			if lastPair.Type != cast.Summarization {
				assert.False(t, containsSummarizedContent(preservedPair),
					"Last pair should NOT be converted to summarized content")
			}

			// Verify the pair size is still the same (not summarized)
			assert.Equal(t, initialSize, preservedPair.Size(),
				"Last pair size should remain unchanged when preserved")

			// Run custom validation for this test case
			tt.validate(t, preservedPair)

			t.Logf("✓ Last pair preserved with size: %d bytes", preservedPair.Size())
		})
	}
}

func TestLastBodyPairWithLargeResponse_MultiPair(t *testing.T) {
	// Test the scenario where:
	// 1. We have multiple body pairs in a section
	// 2. The last pair has a large tool response with reasoning
	// 3. Previous pairs can be summarized, but the last one must be preserved

	ctx := context.Background()
	handler := createMockSummarizeHandler()

	// Create first pair (normal size, can be summarized)
	normalPair1 := createLargeBodyPair(18*1024, "Normal pair 1")

	// Create second pair (oversized, can be summarized)
	largePair2 := createLargeBodyPair(25*1024, "Large pair 2")

	// Create last pair with Anthropic-style reasoning and large response (should NOT be summarized)
	anthropicStyleLastPair := func() *cast.BodyPair {
		aiMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				// Anthropic pattern: reasoning in separate TextContent
				llms.TextContent{
					Text: "",
					Reasoning: &reasoning.ContentReasoning{
						Content:   "Let me try a different approach. Maybe the SQL injection is in one of the POST parameters.",
						Signature: []byte("anthropic_signature_RXVJQ0NrWUlEeGdDS2tCdjU2enZVOGNJaER0U0pKM2ZSRlJFeU5y"),
					},
				},
				llms.ToolCall{
					ID:   "toolu_01QG5rJ5q3uoYNRB483Mp5tX",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "pentester",
						Arguments: `{"message":"Delegating to pentester","question":"I need help with SQL injection"}`,
					},
					// No reasoning in ToolCall for Anthropic
				},
			},
		}

		// Large tool response (50KB+)
		largeResponse := strings.Repeat("SQL injection test result: parameter X shows no delay\n", 1000)

		toolMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "toolu_01QG5rJ5q3uoYNRB483Mp5tX",
					Name:       "pentester",
					Content:    largeResponse,
				},
			},
		}

		return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	}()

	header := cast.NewHeader(
		newTextMsg(llms.ChatMessageTypeSystem, "System"),
		newTextMsg(llms.ChatMessageTypeHuman, "Find SQL injection"),
	)
	section := cast.NewChainSection(header, []*cast.BodyPair{normalPair1, largePair2, anthropicStyleLastPair})

	// Verify initial state
	assert.Equal(t, 3, len(section.Body))
	initialLastPairSize := section.Body[2].Size()
	t.Logf("Initial last pair size: %d bytes", initialLastPairSize)

	// Test summarizeOversizedBodyPairs
	err := summarizeOversizedBodyPairs(ctx, section, handler, 16*1024, cast.ToolCallIDTemplate)
	assert.NoError(t, err)

	// Verify results
	assert.Equal(t, 3, len(section.Body), "Should still have 3 body pairs")

	// First two pairs should be summarized (they're oversized and not last)
	assert.True(t, containsSummarizedContent(section.Body[0]) || section.Body[0].Type == cast.Summarization,
		"First pair should be summarized (oversized and not last)")
	assert.True(t, containsSummarizedContent(section.Body[1]) || section.Body[1].Type == cast.Summarization,
		"Second pair should be summarized (oversized and not last)")

	// CRITICAL: Last pair should NOT be summarized
	lastPair := section.Body[2]
	assert.Equal(t, cast.RequestResponse, lastPair.Type,
		"Last pair type should remain RequestResponse")
	assert.False(t, containsSummarizedContent(lastPair),
		"Last pair should NOT be summarized even though it's large")
	assert.Equal(t, initialLastPairSize, lastPair.Size(),
		"Last pair size should remain unchanged")

	// Verify Anthropic reasoning signature is preserved
	foundAnthropicReasoning := false
	for _, part := range lastPair.AIMessage.Parts {
		if textContent, ok := part.(llms.TextContent); ok && textContent.Reasoning != nil {
			foundAnthropicReasoning = true
			assert.Contains(t, textContent.Reasoning.Content, "SQL injection")
			assert.NotEmpty(t, textContent.Reasoning.Signature,
				"Anthropic signature should be preserved in last pair")
		}
	}
	assert.True(t, foundAnthropicReasoning,
		"Anthropic-style reasoning should be preserved in last pair")

	t.Log("✓ Last pair with Anthropic reasoning and large response preserved correctly")
}

// Helper to create a large body pair for testing
func createLargeBodyPair(size int, content string) *cast.BodyPair {
	// Create content of specified size
	largeContent := strings.Repeat("x", size)

	toolCallID := templates.GenerateFromPattern(cast.ToolCallIDTemplate)
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   toolCallID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "test_function",
					Arguments: fmt.Sprintf(`{"data": "%s"}`, content),
				},
			},
		},
	}

	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCallID,
				Name:       "test_function",
				Content:    largeContent,
			},
		},
	}

	return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
}
