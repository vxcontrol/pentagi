package csum

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"pentagi/pkg/cast"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

// astModifier is a function that modifies an AST for testing
type astModifier func(t *testing.T, ast *cast.ChainAST)

// astCheck is a function that verifies the AST state after summarization
type astCheck func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST)

// Comprehensive check for verifying that the AST has been properly summarized according to configuration
func checkSummarizationResults(config SummarizerConfig) astCheck {
	return func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
		// Basic integrity checks
		verifyASTConsistency(t, ast)

		// Section count checks based on QA summarization
		if config.UseQA && len(originalAST.Sections) > config.MaxQASections {
			// Should be at most maxQASections + 1 (for summary section)
			assert.LessOrEqual(t, len(ast.Sections), config.MaxQASections+1,
				"After QA summarization, section count should be within limits")

			// First section should contain QA summarized content
			if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
				assert.True(t, containsSummarizedContent(ast.Sections[0].Body[0]),
					"First section should contain QA summarized content")
			}
		}

		// Last section size checks based on preserveLast configuration
		if config.PreserveLast && len(ast.Sections) > 0 {
			lastSection := ast.Sections[len(ast.Sections)-1]

			// If original size was larger than the limit, verify it was reduced
			if originalLastSectionSize(originalAST) > config.LastSecBytes {
				assert.LessOrEqual(t, lastSection.Size(), config.LastSecBytes+500, // Allow more overhead
					"Last section size should be around the limit after summarization")

				// Check for summarized content in the last section
				hasSummary := false
				for _, pair := range lastSection.Body {
					if containsSummarizedContent(pair) {
						hasSummary = true
						break
					}
				}
				assert.True(t, hasSummary, "Last section should contain summarized content")
			}
		}

		// Individual body pair size checks
		if config.MaxBPBytes > 0 && len(ast.Sections) > 0 {
			// Check all sections for oversized body pairs
			for _, section := range ast.Sections {
				for _, pair := range section.Body {
					// Skip already summarized pairs
					if containsSummarizedContent(pair) {
						continue
					}

					// Verify that non-summarized Completion body pairs are within size limits
					if pair.Type == cast.Completion {
						assert.LessOrEqual(t, pair.Size(), config.MaxBPBytes+200, // Allow some overhead
							"Individual non-summarized body pairs should not exceed MaxBPBytes limit")
					}
				}
			}
		}

		// Message count checks - should not increase after summarization
		originalMsgs := originalAST.Messages()
		newMsgs := ast.Messages()
		assert.LessOrEqual(t, len(newMsgs), len(originalMsgs),
			"Message count should not increase after summarization")

		// Section checks - non-last sections should have exactly one Completion body pair
		for i := 0; i < len(ast.Sections)-1; i++ {
			section := ast.Sections[i]
			if i == 0 && config.UseQA && len(originalAST.Sections) > config.MaxQASections {
				// If QA summarization happened, first section is the summary
				assert.Equal(t, 1, len(section.Body),
					"First section after QA summarization should have exactly one body pair")

				// Check for either Completion or Summarization type
				// Both are valid after our changes to the summarizer code
				bodyPairType := section.Body[0].Type
				assert.True(t, bodyPairType == cast.Completion || bodyPairType == cast.Summarization,
					"First section body pair should be either Completion or Summarization type")
			} else if i > 0 || !config.UseQA || len(originalAST.Sections) <= config.MaxQASections {
				// Other non-last sections should have one body pair (Completion or Summarization)
				assert.Equal(t, 1, len(section.Body),
					fmt.Sprintf("Non-last section %d should have exactly one body pair", i))

				bodyPairType := section.Body[0].Type
				assert.True(t, bodyPairType == cast.Completion || bodyPairType == cast.Summarization,
					fmt.Sprintf("Non-last section %d body pair should be either Completion or Summarization type", i))
			}
		}
	}
}

// Tests that summarization reduces the size of the AST
func checkSizeReduction(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
	// Skip the check if original AST is empty
	if len(originalAST.Sections) == 0 {
		return
	}

	// Compare message counts rather than direct AST size
	// This is more reliable since AST size includes internal structures
	originalMsgs := originalAST.Messages()
	newMsgs := ast.Messages()

	// Should never increase message count
	assert.LessOrEqual(t, len(newMsgs), len(originalMsgs),
		"Summarization should not increase message count")

	// For larger message sets, we expect reduction
	if len(originalMsgs) > 10 {
		assert.Less(t, len(newMsgs), len(originalMsgs),
			"Summarization should reduce message count for larger chains")
	}
}

// Gets the size of the last section in an AST, or 0 if empty
func originalLastSectionSize(ast *cast.ChainAST) int {
	if len(ast.Sections) == 0 {
		return 0
	}
	return ast.Sections[len(ast.Sections)-1].Size()
}

// TestSummarizeChain verifies the combined chain summarization algorithm
// that integrates section summarization, last section rotation, and QA pair summarization.
// It tests various configurations and sequential modifications to ensure that
// the overall algorithm behaves correctly in real-world usage scenarios.
func TestSummarizeChain(t *testing.T) {
	ctx := context.Background()
	// Test cases for different summarization scenarios
	tests := []struct {
		name           string
		initialAST     *cast.ChainAST
		providerConfig SummarizerConfig
		modifiers      []astModifier
		checks         []astCheck
	}{
		{
			// Tests that last section rotation properly summarizes content when
			// the last section exceeds byte size limit
			name: "Last section rotation",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Initial question"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Initial response"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast: true,
				LastSecBytes: 500,  // Small enough to trigger summarization
				MaxBPBytes:   1000, // Larger than body pairs so only last section logic triggers
				UseQA:        false,
			},
			modifiers: []astModifier{
				// Add 5 body pairs, each with 200 bytes
				addBodyPairsToLastSection(5, 200),
			},
			checks: []astCheck{
				// After summarization, verify all aspects of the result
				checkSummarizedContent,
				checkLastSectionSize(500),
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// Check size reduction
					checkSizeReduction(t, ast, originalAST)

					// Verify that the last section structure follows expected pattern
					assert.Equal(t, 1, len(ast.Sections), "Should have one section")
					lastSection := ast.Sections[0]

					// First body pair should be the summary
					assert.True(t, containsSummarizedContent(lastSection.Body[0]),
						"First body pair should be a summary")

					// Verify the original header is preserved
					assert.NotNil(t, lastSection.Header.SystemMessage, "System message should be preserved")
					assert.NotNil(t, lastSection.Header.HumanMessage, "Human message should be preserved")
				},
				// Comprehensive check with all the provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast: true,
					LastSecBytes: 500,
					MaxBPBytes:   1000,
					UseQA:        false,
				}),
			},
		},
		{
			// Tests QA pair summarization when the number of sections exceeds the limit
			name: "QA pair summarization",
			initialAST: createTestChainAST(
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
					cast.NewHeader(
						nil,
						newTextMsg(llms.ChatMessageTypeHuman, "Question 2"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast:  false,
				UseQA:         true,
				MaxQASections: 2, // Small enough to trigger summarization
				MaxQABytes:    10000,
				MaxBPBytes:    1000, // Not relevant for this test
			},
			modifiers: []astModifier{
				// Add 3 new sections to exceed maxQASections
				addNewSection("Question 3", 1, 100),
				addNewSection("Question 4", 1, 100),
				addNewSection("Question 5", 1, 100),
			},
			checks: []astCheck{
				// After summarization, should have QA summarized content
				checkSummarizedContent,
				checkSectionCount(3),
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// First section should contain QA summary
					assert.Greater(t, len(ast.Sections), 0, "AST should have at least one section")
					if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
						assert.True(t, containsSummarizedContent(ast.Sections[0].Body[0]),
							"First section should contain QA summarized content")
					}

					// System message should be preserved in the first section
					if len(ast.Sections) > 0 {
						assert.NotNil(t, ast.Sections[0].Header.SystemMessage,
							"System message should be preserved in first section after QA summarization")
					}

					// Check size reduction
					checkSizeReduction(t, ast, originalAST)
				},
				// Comprehensive check with all the provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast:  false,
					UseQA:         true,
					MaxQASections: 2,
					MaxQABytes:    10000,
					MaxBPBytes:    1000,
				}),
			},
		},
		{
			// Tests combined summarization with sequential modifications
			// First last section grows, then new sections are added
			name: "Combined summarization with sequential modifications",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast:  true,
				LastSecBytes:  500,
				UseQA:         true,
				MaxQASections: 2,
				MaxQABytes:    10000,
				MaxBPBytes:    1000, // Not a limiting factor in this test
			},
			modifiers: []astModifier{
				// First add many body pairs to last section
				addBodyPairsToLastSection(5, 200),
				// Then add new sections
				addNewSection("Question 2", 1, 100),
				addNewSection("Question 3", 1, 100),
				addNewSection("Question 4", 1, 100),
			},
			checks: []astCheck{
				// After first modification, last section should be summarized
				checkSummarizedContent,
				checkLastSectionSize(500),

				// After adding sections, QA summarization should happen
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// First section should have summarized QA content
					assert.True(t, len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0,
						"First section should have body pairs")
					if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
						pair := ast.Sections[0].Body[0]
						// The pair was summarized once and contains the summarized content prefix
						assert.True(t, containsSummarizedContent(pair),
							"First section should contain QA summarized content")
					}

					// System message should be preserved in the first section
					if len(ast.Sections) > 0 {
						assert.NotNil(t, ast.Sections[0].Header.SystemMessage,
							"System message should be preserved in first section")
					}

					// Total sections should be limited
					assert.LessOrEqual(t, len(ast.Sections), 3, // 1 summary + maxQASections
						"Section count should be within limit after summarization")

					// Check size reduction
					checkSizeReduction(t, ast, originalAST)
				},

				// Comprehensive check with all provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast:  true,
					LastSecBytes:  500,
					UseQA:         true,
					MaxQASections: 2,
					MaxQABytes:    10000,
					MaxBPBytes:    1000,
				}),
			},
		},
		{
			// Tests how tool calls are handled before section summarization
			name: "Tool calls followed by section summarization",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Initial question"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Initial response"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast: true,
				LastSecBytes: 2000,
				MaxBPBytes:   5000,  // Larger than content to not trigger individual pair summarization
				UseQA:        false, // Testing only section summarization first
			},
			modifiers: []astModifier{
				// First add a tool call
				addToolCallToLastSection("search"),

				// Then add many body pairs to last section
				addBodyPairsToLastSection(6, 500), // Big enough to trigger summarization for sure
			},
			checks: []astCheck{
				// After tool call, no summarization needed yet
				checkSectionCount(1),

				// After adding many body pairs, last section should be summarized
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// Verify section count
					assert.Equal(t, 1, len(ast.Sections), "Should still have one section")

					// Verify last section has summarized content
					lastSection := ast.Sections[0]
					foundSummary := false
					for _, pair := range lastSection.Body {
						if containsSummarizedContent(pair) {
							foundSummary = true
							break
						}
					}
					assert.True(t, foundSummary, "Section should contain summarized content")

					// Last section size should be within limits with some tolerance
					assert.LessOrEqual(t, lastSection.Size(), 2500,
						"Last section size should be reasonably close to the limit")

					// Tool call and response should be preserved or summarized
					foundToolRef := false
					for _, pair := range lastSection.Body {
						if pair.Type == cast.RequestResponse || pair.Type == cast.Summarization {
							// Original tool call preserved
							foundToolRef = true
							break
						}

						if pair.Type == cast.Completion && pair.AIMessage != nil {
							for _, part := range pair.AIMessage.Parts {
								if textContent, ok := part.(llms.TextContent); ok {
									if strings.Contains(textContent.Text, "search") ||
										strings.Contains(textContent.Text, "tool") {
										// Reference to tool in summary
										foundToolRef = true
										break
									}
								}
							}
						}
					}
					assert.True(t, foundToolRef, "Reference to tool call should be preserved in some form")

					// Check size reduction
					checkSizeReduction(t, ast, originalAST)
				},

				// Comprehensive check with all provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast: true,
					LastSecBytes: 2000,
					MaxBPBytes:   5000,
					UseQA:        false,
				}),
			},
		},
		{
			// Tests QA summarization with many sections to verify the algorithm
			// correctly reduces the total number of sections
			name: "Sequential QA summarization after section growth",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 1"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast:  false, // Focus on QA summarization only
				UseQA:         true,
				MaxQASections: 2, // Very restrictive to ensure summarization
				MaxQABytes:    10000,
				MaxBPBytes:    5000, // Large enough to not impact this test
			},
			modifiers: []astModifier{
				// Add many sections to exceed the QA section limit
				addNewSection("Question 2", 1, 500),
				addNewSection("Question 3", 1, 500),
				addNewSection("Question 4", 1, 500),
				addNewSection("Question 5", 1, 500),
				addNewSection("Question 6", 1, 500),
				addNewSection("Question 7", 1, 500),
				addNewSection("Question 8", 1, 500),
			},
			checks: []astCheck{
				// After adding so many sections, verify the chain is summarized
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// The key check: after summarization, we should have fewer sections
					// than the original number of added sections (7) + initial section
					assert.Less(t, len(ast.Sections), 8,
						"QA summarization should reduce the total number of sections")

					// Also verify that the number of sections is within the maxQASections limit
					// plus potentially 1 for the summary section
					assert.LessOrEqual(t, len(ast.Sections), 3,
						"Section count should be within limits (summary + maxQASections)")

					// Check size reduction
					checkSizeReduction(t, ast, originalAST)

					// Verify system message preservation
					if len(ast.Sections) > 0 {
						assert.NotNil(t, ast.Sections[0].Header.SystemMessage,
							"System message should be preserved after QA summarization")
					}
				},

				// Comprehensive check with all provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast:  false,
					UseQA:         true,
					MaxQASections: 2,
					MaxQABytes:    10000,
					MaxBPBytes:    5000,
				}),
			},
		},
		{
			// Tests QA summarization triggered by byte size limit rather than section count
			name: "Byte size limit in QA pairs",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question 1"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion(strings.Repeat("A", 500)),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast:  false,
				UseQA:         true,
				MaxQASections: 10,   // Large enough to not trigger count limit
				MaxQABytes:    800,  // Smaller to ensure byte limit is triggered
				MaxBPBytes:    5000, // Not the limiting factor in this test
			},
			modifiers: []astModifier{
				// Add sections with large content
				addNewSection("Question 2", 1, 500),
				addNewSection("Question 3", 1, 500),
				addNewSection("Question 4", 1, 500),
				addNewSection("Question 5", 1, 500), // More sections to ensure we exceed the limits
			},
			checks: []astCheck{
				// Should trigger byte limit summarization
				checkSummarizedContent,
				checkTotalSize(1000), // maxQABytes + some overhead

				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// Verify that the size trigger rather than count trigger was used
					assert.Less(t, ast.Size(), originalAST.Size(),
						"Total size should be reduced after byte-triggered summarization")

					// Check for QA summarization pattern
					if len(ast.Sections) > 0 && len(ast.Sections[0].Body) > 0 {
						assert.True(t, containsSummarizedContent(ast.Sections[0].Body[0]),
							"First section should contain QA summarized content")
					}

					// System message should be preserved
					if len(ast.Sections) > 0 {
						assert.NotNil(t, ast.Sections[0].Header.SystemMessage,
							"System message should be preserved")
					}
				},

				// Comprehensive check with all provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast:  false,
					UseQA:         true,
					MaxQASections: 10,
					MaxQABytes:    800,
					MaxBPBytes:    5000,
				}),
			},
		},
		{
			// Tests oversized individual body pairs summarization
			name: "Oversized individual body pairs summarization",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question with potentially large responses"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Initial normal response"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast: true,
				LastSecBytes: 50 * 1024, // Large enough to not trigger full section summarization
				MaxBPBytes:   16 * 1024, // Default value for maxSingleBodyPairByteSize
				UseQA:        false,
			},
			modifiers: []astModifier{
				// Add one normal pair and one oversized pair (exceeding 16KB)
				addNormalAndOversizedBodyPairs(),
				// Add additional pairs to ensure size reduction for SummarizeChain return
				func(t *testing.T, ast *cast.ChainAST) {
					if len(ast.Sections) == 0 {
						return
					}
					lastSection := ast.Sections[0]
					// Add many pairs to ensure message count reduction
					for i := 0; i < 20; i++ {
						pair := cast.NewBodyPairFromCompletion(fmt.Sprintf("Additional pair %d", i))
						lastSection.AddBodyPair(pair)
					}
				},
			},
			checks: []astCheck{
				// Just verify the basic structure after processing
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// Verify section count
					assert.Equal(t, 1, len(ast.Sections), "Should still have one section")

					// Get the body pairs
					lastSection := ast.Sections[0]

					// Check which pair was the oversized one and verify it's summarized
					// The oversized pair should be at index 2 (after original and normal pair)
					assert.True(t, containsSummarizedContent(lastSection.Body[2]),
						"Should have summarized the oversized body pair")

					// We should have 3 body pairs in total (original + 2 added)
					assert.Equal(t, 3, len(lastSection.Body), "Should have 3 body pairs in total")

					// Basic AST check
					verifyASTConsistency(t, ast)
				},

				// Comprehensive check with all provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast: true,
					LastSecBytes: 50 * 1024,
					MaxBPBytes:   16 * 1024,
					UseQA:        false,
				}),
			},
		},
		{
			// Tests section summarization with keepQASections=2
			name: "Section summarization with keep last 2 QA sections",
			initialAST: createTestChainAST(
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
					cast.NewHeader(
						nil,
						newTextMsg(llms.ChatMessageTypeHuman, "Question 2"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 2a"),
						cast.NewBodyPairFromCompletion("Answer 2b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(
						nil,
						newTextMsg(llms.ChatMessageTypeHuman, "Question 3"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 3a"),
						cast.NewBodyPairFromCompletion("Answer 3b"),
					},
				),
				cast.NewChainSection(
					cast.NewHeader(
						nil,
						newTextMsg(llms.ChatMessageTypeHuman, "Question 4"),
					),
					[]*cast.BodyPair{
						cast.NewBodyPairFromCompletion("Answer 4a"),
						cast.NewBodyPairFromCompletion("Answer 4b"),
					},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast:   false,
				UseQA:          false, // Not testing QA summarization here
				KeepQASections: 2,     // Key configuration - keep last 2 sections
				MaxBPBytes:     5000,  // Not the limiting factor in this test
			},
			modifiers: []astModifier{
				// No modifiers needed as we've already set up the sections in initialAST
			},
			checks: []astCheck{
				// Verify the effect of keepQASections=2
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// We should have 4 sections total
					assert.Equal(t, 4, len(ast.Sections), "Should have 4 sections total")

					// All sections except last 2 should be summarized (have one body pair)
					for i := 0; i < len(ast.Sections)-2; i++ {
						section := ast.Sections[i]
						assert.Equal(t, 1, len(section.Body),
							fmt.Sprintf("Section %d should have exactly one body pair (summarized)", i))

						assert.True(t, containsSummarizedContent(section.Body[0]),
							fmt.Sprintf("Section %d should have summarized content", i))
					}

					// Last 2 sections should not be summarized (preserve original body pairs)
					// Section at index 2 (Third section)
					assert.Equal(t, 2, len(ast.Sections[2].Body),
						"Third section should have 2 body pairs (not summarized)")

					// Section at index 3 (Fourth section)
					assert.Equal(t, 2, len(ast.Sections[3].Body),
						"Fourth section should have 2 body pairs (not summarized)")

					// Check that the content of the last two sections is preserved
					// Third section should contain "Question 3" in its human message
					humanMsg := ast.Sections[2].Header.HumanMessage
					assert.Contains(t, humanMsg.Parts[0].(llms.TextContent).Text, "Question 3",
						"Third section should have original human message")

					// Fourth section should contain "Question 4" in its human message
					humanMsg = ast.Sections[3].Header.HumanMessage
					assert.Contains(t, humanMsg.Parts[0].(llms.TextContent).Text, "Question 4",
						"Fourth section should have original human message")
				},

				// Comprehensive check
				checkSummarizationResults(SummarizerConfig{
					PreserveLast:   false,
					UseQA:          false,
					KeepQASections: 2,
					MaxBPBytes:     5000,
				}),
			},
		},
		{
			// Test to verify that MaxBPBytes limitation works properly
			// Should summarize only the oversized pair while leaving other pairs intact
			name: "MaxBPBytes_specific_test",
			initialAST: createTestChainAST(
				cast.NewChainSection(
					cast.NewHeader(
						newTextMsg(llms.ChatMessageTypeSystem, "System message"),
						newTextMsg(llms.ChatMessageTypeHuman, "Question requiring various size responses"),
					),
					[]*cast.BodyPair{},
				),
			),
			providerConfig: SummarizerConfig{
				PreserveLast: true,
				LastSecBytes: 30 * 1024, // Very large to avoid triggering last section summarization
				MaxBPBytes:   1000,      // Small enough to trigger oversized pair summarization but not for normal pairs
				UseQA:        false,
			},
			modifiers: []astModifier{
				// Add specifically crafted body pairs:
				// 1. A normal pair that is just under the MaxBPBytes limit
				// 2. An oversized pair that exceeds the MaxBPBytes limit
				// 3. Another normal pair
				func(t *testing.T, ast *cast.ChainAST) {
					if len(ast.Sections) == 0 {
						t.Fatal("AST has no sections")
					}

					lastSection := ast.Sections[0]

					// Add a short response
					normalPair1 := cast.NewBodyPairFromCompletion("Short initial response")
					lastSection.AddBodyPair(normalPair1)

					// Add a response around 300 bytes
					normalPair2 := cast.NewBodyPairFromCompletion(strings.Repeat("A", 300))
					lastSection.AddBodyPair(normalPair2)

					// Add a body pair that's just under the MaxBPBytes limit
					underLimitPair := cast.NewBodyPairFromCompletion(strings.Repeat("B", 900)) // 900 bytes, under 1000 limit
					lastSection.AddBodyPair(underLimitPair)

					// Add an oversized pair significantly over the MaxBPBytes limit
					oversizedPair := cast.NewBodyPairFromCompletion(strings.Repeat("C", 2000)) // 2000 bytes, over 1000 limit
					lastSection.AddBodyPair(oversizedPair)

					// Add another normal pair well under the limit
					normalPair3 := cast.NewBodyPairFromCompletion("Another normal response")
					lastSection.AddBodyPair(normalPair3)

					// Create a smaller message set to ensure summarizeChain will return the modified chain
					// Add a large number of additional pairs to trigger message count reduction
					for i := 0; i < 10; i++ {
						additionalPair := cast.NewBodyPairFromCompletion(fmt.Sprintf("Additional message %d", i))
						lastSection.AddBodyPair(additionalPair)
					}
				},
			},
			checks: []astCheck{
				// Verify that only the oversized pair was summarized
				func(t *testing.T, ast *cast.ChainAST, originalAST *cast.ChainAST) {
					// Verify section count
					assert.Equal(t, 1, len(ast.Sections), "Should have one section")

					lastSection := ast.Sections[0]

					// Count summarized body pairs
					summarizedCount := 0
					for _, pair := range lastSection.Body {
						if containsSummarizedContent(pair) {
							summarizedCount++
						}
					}

					// Only one pair should be summarized (the oversized one)
					assert.Equal(t, 1, summarizedCount, "Only one body pair should be summarized")

					// Should have all the original body pairs
					assert.Equal(t, 15, len(lastSection.Body), "Should have all body pairs (5 original + 10 additional)")

					// Check size of non-summarized pairs
					for _, pair := range lastSection.Body {
						if !containsSummarizedContent(pair) {
							assert.LessOrEqual(t, pair.Size(), 1000+100, // MaxBPBytes + small overhead
								"Non-summarized pairs should be under MaxBPBytes limit")
						}
					}
				},

				// Comprehensive check with the provider's configuration
				checkSummarizationResults(SummarizerConfig{
					PreserveLast: true,
					LastSecBytes: 30 * 1024,
					MaxBPBytes:   1000,
					UseQA:        false,
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone AST for testing
			ast := cloneAST(tt.initialAST)

			// Create mock summarizer
			mockSum := newMockSummarizer("Summarized content", nil, nil)

			// Create flow provider with test configuration
			summarizer := NewSummarizer(tt.providerConfig)

			// Run through sequential modifications and checks
			for i, modifier := range tt.modifiers {
				// Apply modifier
				modifier(t, ast)

				// Verify AST consistency after modification
				verifyASTConsistency(t, ast)

				// Convert to messages - this is what's passed to SummarizeChain
				messages := ast.Messages()
				originalSize := len(messages)

				// Save the original AST for comparison in checks
				originalAST := cloneAST(ast)

				// Summarize chain
				newMessages, err := summarizer.SummarizeChain(ctx, mockSum.SummarizerHandler(), messages)
				assert.NoError(t, err, "Failed to summarize chain")

				// Convert back to AST for verification
				newAST, err := cast.NewChainAST(newMessages, false)
				assert.NoError(t, err, "Failed to create AST from summarized messages")

				// Verify new AST consistency
				verifyASTConsistency(t, newAST)

				// Run check for this iteration if available
				if i < len(tt.checks) {
					tt.checks[i](t, newAST, originalAST)
				}

				// Verify that summarization either reduced size or left it unchanged
				assert.LessOrEqual(t, len(newMessages), originalSize,
					"Summarization should not increase message count")

				// Update AST for next iteration
				ast = newAST
			}
		})
	}
}

// Clones an AST by serializing to messages and back
func cloneAST(ast *cast.ChainAST) *cast.ChainAST {
	messages := ast.Messages()
	newAST, _ := cast.NewChainAST(messages, false)
	return newAST
}

// Adds body pairs to the last section
func addBodyPairsToLastSection(count int, size int) astModifier {
	return func(t *testing.T, ast *cast.ChainAST) {
		if len(ast.Sections) == 0 {
			// Add a new section if none exists
			header := cast.NewHeader(
				newTextMsg(llms.ChatMessageTypeSystem, "System message"),
				newTextMsg(llms.ChatMessageTypeHuman, "Initial question"),
			)
			section := cast.NewChainSection(header, []*cast.BodyPair{})
			ast.AddSection(section)
		}

		lastSection := ast.Sections[len(ast.Sections)-1]
		for i := 0; i < count; i++ {
			text := strings.Repeat("A", size)
			bodyPair := cast.NewBodyPairFromCompletion(fmt.Sprintf("Response %d: %s", i, text))
			lastSection.AddBodyPair(bodyPair)
		}
	}
}

// Adds a new section to the AST
func addNewSection(human string, bodyPairCount int, bodyPairSize int) astModifier {
	return func(t *testing.T, ast *cast.ChainAST) {
		humanMsg := newTextMsg(llms.ChatMessageTypeHuman, human)
		header := cast.NewHeader(nil, humanMsg)
		bodyPairs := make([]*cast.BodyPair, 0, bodyPairCount)

		for i := 0; i < bodyPairCount; i++ {
			text := strings.Repeat("B", bodyPairSize)
			bodyPairs = append(bodyPairs, cast.NewBodyPairFromCompletion(fmt.Sprintf("Answer %d: %s", i, text)))
		}

		section := cast.NewChainSection(header, bodyPairs)
		ast.AddSection(section)
	}
}

// Adds a tool call to the last section
func addToolCallToLastSection(toolName string) astModifier {
	return func(t *testing.T, ast *cast.ChainAST) {
		if len(ast.Sections) == 0 {
			// Add a new section if none exists
			header := cast.NewHeader(
				newTextMsg(llms.ChatMessageTypeSystem, "System message"),
				newTextMsg(llms.ChatMessageTypeHuman, "Initial question"),
			)
			section := cast.NewChainSection(header, []*cast.BodyPair{})
			ast.AddSection(section)
		}

		lastSection := ast.Sections[len(ast.Sections)-1]

		// Create a RequestResponse pair with tool call
		aiMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Let me use a tool"},
				llms.ToolCall{
					ID:   toolName + "-id",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      toolName,
						Arguments: `{"query": "test"}`,
					},
				},
			},
		}
		toolMsg := &llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: toolName + "-id",
					Name:       toolName,
					Content:    "Tool response for " + toolName,
				},
			},
		}

		bodyPair := cast.NewBodyPair(aiMsg, []*llms.MessageContent{toolMsg})
		lastSection.AddBodyPair(bodyPair)
	}
}

// Verifies section count
func checkSectionCount(expected int) astCheck {
	return func(t *testing.T, ast *cast.ChainAST, _ *cast.ChainAST) {
		assert.Equal(t, expected, len(ast.Sections),
			"AST should have the expected number of sections")
	}
}

// Verifies total size
func checkTotalSize(maxSize int) astCheck {
	return func(t *testing.T, ast *cast.ChainAST, _ *cast.ChainAST) {
		assert.LessOrEqual(t, ast.Size(), maxSize,
			"AST size should be less than or equal to the maximum size")
	}
}

// Verifies last section size
func checkLastSectionSize(maxSize int) astCheck {
	return func(t *testing.T, ast *cast.ChainAST, _ *cast.ChainAST) {
		if len(ast.Sections) == 0 {
			assert.Fail(t, "AST has no sections")
			return
		}

		lastSection := ast.Sections[len(ast.Sections)-1]
		assert.LessOrEqual(t, lastSection.Size(), maxSize,
			"Last section size should be less than or equal to the maximum size")
	}
}

// Checks for summarized content anywhere in the AST
func checkSummarizedContent(t *testing.T, ast *cast.ChainAST, _ *cast.ChainAST) {
	found := false
	for _, section := range ast.Sections {
		for _, pair := range section.Body {
			if containsSummarizedContent(pair) {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	assert.True(t, found, "AST should contain summarized content")
}

// Adds a normal body pair and one oversized body pair to test individual pair summarization
func addNormalAndOversizedBodyPairs() astModifier {
	return func(t *testing.T, ast *cast.ChainAST) {
		if len(ast.Sections) == 0 {
			// Add a new section if none exists
			header := cast.NewHeader(
				newTextMsg(llms.ChatMessageTypeSystem, "System message"),
				newTextMsg(llms.ChatMessageTypeHuman, "Initial question"),
			)
			section := cast.NewChainSection(header, []*cast.BodyPair{})
			ast.AddSection(section)
		}

		lastSection := ast.Sections[len(ast.Sections)-1]

		// Add a normal body pair first
		normalPair := cast.NewBodyPairFromCompletion("Another normal response that is well within size limits")
		lastSection.AddBodyPair(normalPair)

		// Add an oversized body pair (exceeding 16KB)
		oversizedText := strings.Repeat("X", 17*1024) // 17KB, which exceeds the 16KB limit
		oversizedPair := cast.NewBodyPairFromCompletion(
			fmt.Sprintf("This is an oversized response that should trigger individual pair summarization: %s", oversizedText),
		)
		lastSection.AddBodyPair(oversizedPair)
	}
}
