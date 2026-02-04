package csum

import (
	"context"
	"testing"

	"pentagi/pkg/cast"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// TestSummarizeOversizedBodyPairs_WithReasoning tests that oversized body pairs
// with reasoning signatures are properly summarized with fake signatures
func TestSummarizeOversizedBodyPairs_WithReasoning(t *testing.T) {
	// Create a section with an oversized body pair that contains reasoning
	oversizedContent := make([]byte, 20*1024) // 20KB
	for i := range oversizedContent {
		oversizedContent[i] = 'X'
	}

	// Create a body pair with reasoning signature
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   "call_test123",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_data",
					Arguments: `{"query": "test"}`,
				},
				Reasoning: &reasoning.ContentReasoning{
					Signature: []byte("original_gemini_signature_12345"),
				},
			},
		},
	}

	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "call_test123",
				Name:       "get_data",
				Content:    string(oversizedContent),
			},
		},
	}

	bodyPair := cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	assert.Greater(t, bodyPair.Size(), 16*1024, "Body pair should be oversized")

	// Create a section with this body pair followed by a normal pair
	section := cast.NewChainSection(
		cast.NewHeader(nil, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Test question"}},
		}),
		[]*cast.BodyPair{
			bodyPair,
			cast.NewBodyPairFromCompletion("This is a normal response"),
		},
	)

	// Create handler that returns a simple summary
	handler := func(ctx context.Context, text string) (string, error) {
		return "Summarized: got data", nil
	}

	// Summarize oversized pairs
	err := summarizeOversizedBodyPairs(
		context.Background(),
		section,
		handler,
		16*1024,
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	// Verify that the first pair was summarized
	assert.Equal(t, 2, len(section.Body), "Should still have 2 body pairs")

	// First pair should now be a summarization
	firstPair := section.Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type, "First pair should be Summarization type")

	// Check that the summarized pair has a fake reasoning signature
	foundSignature := false
	for _, part := range firstPair.AIMessage.Parts {
		if toolCall, ok := part.(llms.ToolCall); ok {
			if toolCall.FunctionCall != nil && toolCall.FunctionCall.Name == cast.SummarizationToolName {
				assert.NotNil(t, toolCall.Reasoning, "Summarized tool call should have reasoning")
				assert.Equal(t, []byte(cast.FakeReasoningSignatureGemini), toolCall.Reasoning.Signature,
					"Should have the fake Gemini signature")
				foundSignature = true
				t.Logf("Found fake signature: %s", toolCall.Reasoning.Signature)
				break
			}
		}
	}
	assert.True(t, foundSignature, "Should find a tool call with fake signature")

	// Second pair should remain unchanged
	assert.Equal(t, cast.Completion, section.Body[1].Type, "Second pair should remain Completion")
}

// TestSummarizeOversizedBodyPairs_WithoutReasoning tests that oversized body pairs
// without reasoning signatures are summarized without fake signatures
func TestSummarizeOversizedBodyPairs_WithoutReasoning(t *testing.T) {
	// Create a section with an oversized body pair WITHOUT reasoning
	oversizedContent := make([]byte, 20*1024) // 20KB
	for i := range oversizedContent {
		oversizedContent[i] = 'Y'
	}

	// Create a body pair without reasoning signature
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   "call_test456",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_info",
					Arguments: `{"query": "test"}`,
				},
				// No Reasoning field
			},
		},
	}

	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "call_test456",
				Name:       "get_info",
				Content:    string(oversizedContent),
			},
		},
	}

	bodyPair := cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	assert.Greater(t, bodyPair.Size(), 16*1024, "Body pair should be oversized")

	// Create a section with this body pair followed by a normal pair
	section := cast.NewChainSection(
		cast.NewHeader(nil, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Test question"}},
		}),
		[]*cast.BodyPair{
			bodyPair,
			cast.NewBodyPairFromCompletion("This is a normal response"),
		},
	)

	// Create handler that returns a simple summary
	handler := func(ctx context.Context, text string) (string, error) {
		return "Summarized: got info", nil
	}

	// Summarize oversized pairs
	err := summarizeOversizedBodyPairs(
		context.Background(),
		section,
		handler,
		16*1024,
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	// Verify that the first pair was summarized
	assert.Equal(t, 2, len(section.Body), "Should still have 2 body pairs")

	// First pair should now be a summarization
	firstPair := section.Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type, "First pair should be Summarization type")

	// Check that the summarized pair does NOT have a reasoning signature
	for _, part := range firstPair.AIMessage.Parts {
		if toolCall, ok := part.(llms.ToolCall); ok {
			if toolCall.FunctionCall != nil && toolCall.FunctionCall.Name == cast.SummarizationToolName {
				assert.Nil(t, toolCall.Reasoning, "Summarized tool call should NOT have reasoning when original didn't")
				t.Logf("Correctly created summarization without fake signature")
				break
			}
		}
	}
}

// TestSummarizeSections_WithReasoning tests that section summarization of PREVIOUS turns
// does NOT add fake signatures, even if original sections contained reasoning.
// This is correct because Gemini only validates thought_signature in the CURRENT turn.
func TestSummarizeSections_WithReasoning(t *testing.T) {
	// Create sections with reasoning signatures
	sections := []*cast.ChainSection{
		// Section 1 with reasoning (previous turn - will be summarized)
		cast.NewChainSection(
			cast.NewHeader(nil, &llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: "Question 1"}},
			}),
			[]*cast.BodyPair{
				func() *cast.BodyPair {
					aiMsg := &llms.MessageContent{
						Role: llms.ChatMessageTypeAI,
						Parts: []llms.ContentPart{
							llms.ToolCall{
								ID:   "call_reasoning_1",
								Type: "function",
								FunctionCall: &llms.FunctionCall{
									Name:      "search",
									Arguments: `{"query": "test1"}`,
								},
								Reasoning: &reasoning.ContentReasoning{
									Signature: []byte("gemini_signature_abc123"),
								},
							},
						},
					}
					toolMsg := &llms.MessageContent{
						Role: llms.ChatMessageTypeTool,
						Parts: []llms.ContentPart{
							llms.ToolCallResponse{
								ToolCallID: "call_reasoning_1",
								Name:       "search",
								Content:    "Result 1",
							},
						},
					}
					return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
				}(),
			},
		),
		// Section 2 - this is the last section (current turn), should NOT be summarized
		cast.NewChainSection(
			cast.NewHeader(nil, &llms.MessageContent{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: "Question 2"}},
			}),
			[]*cast.BodyPair{
				cast.NewBodyPairFromCompletion("Answer 2"),
			},
		),
	}

	ast := &cast.ChainAST{Sections: sections}

	// Create handler
	handler := func(ctx context.Context, text string) (string, error) {
		return "Summary of section", nil
	}

	// Summarize sections (keep last 1 section = current turn)
	err := summarizeSections(
		context.Background(),
		ast,
		handler,
		1, // keep last 1 section (current turn)
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	// First section should be summarized
	assert.Equal(t, 1, len(ast.Sections[0].Body), "First section should have 1 body pair")
	firstPair := ast.Sections[0].Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type, "Should be Summarization type")

	// IMPORTANT: Check that there is NO fake signature
	// Previous turns don't need fake signatures - only current turn needs them
	for _, part := range firstPair.AIMessage.Parts {
		if toolCall, ok := part.(llms.ToolCall); ok {
			if toolCall.FunctionCall != nil && toolCall.FunctionCall.Name == cast.SummarizationToolName {
				assert.Nil(t, toolCall.Reasoning,
					"Previous turn should NOT have fake signature (Gemini only validates current turn)")
				t.Logf("Correctly created summarization WITHOUT fake signature for previous turn")
			}
		}
	}

	// Second section should remain unchanged (this is current turn)
	assert.Equal(t, 1, len(ast.Sections[1].Body), "Second section should remain unchanged")
	assert.Equal(t, cast.Completion, ast.Sections[1].Body[0].Type, "Should be Completion type")
}

// TestSummarizeLastSection_WithReasoning tests that summarization of the CURRENT turn
// (last section) DOES add fake signatures when original content contained reasoning.
// This is critical for Gemini API compatibility.
func TestSummarizeLastSection_WithReasoning(t *testing.T) {
	// Create a large body pair with reasoning in the last section (current turn)
	oversizedContent := make([]byte, 30*1024) // 30KB to trigger summarization
	for i := range oversizedContent {
		oversizedContent[i] = 'Z'
	}

	// Create body pair with reasoning signature
	bodyPairWithReasoning := func() *cast.BodyPair {
		aiMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   "call_current_turn",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "analyze",
						Arguments: `{"data": "large dataset"}`,
					},
					Reasoning: &reasoning.ContentReasoning{
						Signature: []byte("gemini_current_turn_signature_xyz"),
					},
				},
			},
		}
		toolMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "call_current_turn",
					Name:       "analyze",
					Content:    string(oversizedContent),
				},
			},
		}
		return cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	}()

	// Create the last section (current turn) with two pairs
	lastSection := cast.NewChainSection(
		cast.NewHeader(nil, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Analyze this data"}},
		}),
		[]*cast.BodyPair{
			bodyPairWithReasoning,                            // Will be summarized (oversized)
			cast.NewBodyPairFromCompletion("Final response"), // Will be kept (last pair)
		},
	)

	ast := &cast.ChainAST{Sections: []*cast.ChainSection{lastSection}}

	// Create handler
	handler := func(ctx context.Context, text string) (string, error) {
		return "Summarized analysis result", nil
	}

	// Summarize the last section (index 0 because it's the only section)
	err := summarizeLastSection(
		context.Background(),
		ast,
		handler,
		0,       // last section index
		50*1024, // max last section bytes
		16*1024, // max single body pair bytes (will trigger oversized pair summarization)
		25,      // reserve percent
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	// Check that the section now has a summarized pair with fake signature
	lastSectionAfter := ast.Sections[0]

	// Should have at least 2 pairs: summarized + final response
	assert.GreaterOrEqual(t, len(lastSectionAfter.Body), 2, "Should have summarized pair + kept pairs")

	// First pair should be the summarization with fake signature
	firstPair := lastSectionAfter.Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type, "First pair should be Summarization")

	// CRITICAL: Check that fake signature WAS added for current turn
	foundFakeSignature := false
	for _, part := range firstPair.AIMessage.Parts {
		if toolCall, ok := part.(llms.ToolCall); ok {
			if toolCall.FunctionCall != nil && toolCall.FunctionCall.Name == cast.SummarizationToolName {
				assert.NotNil(t, toolCall.Reasoning,
					"Current turn summarization MUST have fake signature for Gemini compatibility")
				assert.Equal(t, []byte(cast.FakeReasoningSignatureGemini), toolCall.Reasoning.Signature,
					"Should have Gemini fake signature")
				foundFakeSignature = true
				t.Logf("✓ Correctly added fake signature for current turn: %s", toolCall.Reasoning.Signature)
			}
		}
	}
	assert.True(t, foundFakeSignature, "Must find fake signature in current turn summarization")

	// Last pair should be preserved (never summarized)
	lastPair := lastSectionAfter.Body[len(lastSectionAfter.Body)-1]
	assert.Equal(t, cast.Completion, lastPair.Type, "Last pair should remain Completion")
}

// TestSummarizeOversizedBodyPairs_WithReasoningMessage tests that oversized body pairs
// with reasoning TextContent (like Kimi/Moonshot) preserve the reasoning message
func TestSummarizeOversizedBodyPairs_WithReasoningMessage(t *testing.T) {
	// Create a section with an oversized body pair that contains reasoning in TextContent
	oversizedContent := make([]byte, 20*1024) // 20KB
	for i := range oversizedContent {
		oversizedContent[i] = 'K'
	}

	// Create a body pair with reasoning in TextContent (Kimi pattern)
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{
				Text: "Let me analyze the wp-abilities plugin...",
				Reasoning: &reasoning.ContentReasoning{
					Content: "The wp-abilities plugin seems to be the main target here. Need to find vulnerabilities.",
				},
			},
			llms.ToolCall{
				ID:   "call_kimi_test",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "wp-abilities CVE"}`,
				},
			},
		},
	}

	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "call_kimi_test",
				Name:       "search",
				Content:    string(oversizedContent),
			},
		},
	}

	bodyPair := cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
	assert.Greater(t, bodyPair.Size(), 16*1024, "Body pair should be oversized")

	// Create a section with this body pair followed by a normal pair
	section := cast.NewChainSection(
		cast.NewHeader(nil, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Find vulnerabilities"}},
		}),
		[]*cast.BodyPair{
			bodyPair,
			cast.NewBodyPairFromCompletion("This is the final response"),
		},
	)

	// Create handler that returns a simple summary
	handler := func(ctx context.Context, text string) (string, error) {
		return "Summarized: found vulnerability info", nil
	}

	// Summarize oversized pairs
	err := summarizeOversizedBodyPairs(
		context.Background(),
		section,
		handler,
		16*1024,
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	// Verify that the first pair was summarized
	assert.Equal(t, 2, len(section.Body), "Should still have 2 body pairs")

	// First pair should now be a summarization
	firstPair := section.Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type, "First pair should be Summarization type")

	// CRITICAL: Check that the summarized pair has BOTH:
	// 1. Reasoning TextContent (for Kimi compatibility)
	// 2. ToolCall with fake signature (for Gemini compatibility)
	assert.GreaterOrEqual(t, len(firstPair.AIMessage.Parts), 2,
		"Should have at least 2 parts: reasoning TextContent + ToolCall")

	// First part should be the reasoning TextContent
	firstPart, ok := firstPair.AIMessage.Parts[0].(llms.TextContent)
	assert.True(t, ok, "First part should be TextContent with reasoning")
	assert.Equal(t, "Let me analyze the wp-abilities plugin...", firstPart.Text)
	assert.NotNil(t, firstPart.Reasoning, "Should preserve original reasoning")
	assert.Equal(t, "The wp-abilities plugin seems to be the main target here. Need to find vulnerabilities.",
		firstPart.Reasoning.Content)
	t.Logf("✓ Preserved reasoning TextContent: %s", firstPart.Reasoning.Content)

	// Second part should be the ToolCall (without fake signature in this case, since no ToolCall.Reasoning in original)
	secondPart, ok := firstPair.AIMessage.Parts[1].(llms.ToolCall)
	assert.True(t, ok, "Second part should be ToolCall")
	assert.Equal(t, cast.SummarizationToolName, secondPart.FunctionCall.Name)
	// Original didn't have ToolCall.Reasoning, so no fake signature needed
	assert.Nil(t, secondPart.Reasoning, "No fake signature needed - original had no ToolCall.Reasoning")
	t.Logf("✓ Created ToolCall without fake signature (original had no ToolCall.Reasoning)")

	// Second pair should remain unchanged
	assert.Equal(t, cast.Completion, section.Body[1].Type, "Second pair should remain Completion")
}

// TestSummarizeOversizedBodyPairs_KimiPattern tests the full Kimi pattern:
// reasoning TextContent + ToolCall with ToolCall.Reasoning
func TestSummarizeOversizedBodyPairs_KimiPattern(t *testing.T) {
	// Create oversized content
	oversizedContent := make([]byte, 20*1024) // 20KB
	for i := range oversizedContent {
		oversizedContent[i] = 'M'
	}

	// Create a body pair with BOTH reasoning patterns (Kimi style)
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.TextContent{
				Text: "Analyzing the vulnerability...",
				Reasoning: &reasoning.ContentReasoning{
					Content: "This appears to be a privilege escalation issue.",
				},
			},
			llms.ToolCall{
				ID:   "call_kimi_full",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "exploit",
					Arguments: `{"target": "plugin"}`,
				},
				Reasoning: &reasoning.ContentReasoning{
					Signature: []byte("kimi_toolcall_signature_abc"),
				},
			},
		},
	}

	toolMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: "call_kimi_full",
				Name:       "exploit",
				Content:    string(oversizedContent),
			},
		},
	}

	bodyPair := cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})

	section := cast.NewChainSection(
		cast.NewHeader(nil, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Exploit the plugin"}},
		}),
		[]*cast.BodyPair{
			bodyPair,
			cast.NewBodyPairFromCompletion("Final response"),
		},
	)

	handler := func(ctx context.Context, text string) (string, error) {
		return "Summarized: exploitation attempt", nil
	}

	err := summarizeOversizedBodyPairs(
		context.Background(),
		section,
		handler,
		16*1024,
		cast.ToolCallIDTemplate,
	)
	assert.NoError(t, err)

	firstPair := section.Body[0]
	assert.Equal(t, cast.Summarization, firstPair.Type)

	// Should have reasoning TextContent + ToolCall
	assert.GreaterOrEqual(t, len(firstPair.AIMessage.Parts), 2)

	// Check reasoning TextContent
	textPart, ok := firstPair.AIMessage.Parts[0].(llms.TextContent)
	assert.True(t, ok, "First part should be reasoning TextContent")
	assert.NotNil(t, textPart.Reasoning)
	assert.Equal(t, "This appears to be a privilege escalation issue.", textPart.Reasoning.Content)
	t.Logf("✓ Preserved reasoning TextContent (Kimi requirement)")

	// Check ToolCall with fake signature
	toolCallPart, ok := firstPair.AIMessage.Parts[1].(llms.ToolCall)
	assert.True(t, ok, "Second part should be ToolCall")
	assert.NotNil(t, toolCallPart.Reasoning, "Should have fake signature (original had ToolCall.Reasoning)")
	assert.Equal(t, []byte(cast.FakeReasoningSignatureGemini), toolCallPart.Reasoning.Signature)
	t.Logf("✓ Added fake signature to ToolCall (Gemini requirement)")
}
