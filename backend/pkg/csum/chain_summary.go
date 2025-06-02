package csum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"pentagi/pkg/cast"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
)

// Default configuration constants for the summarization algorithm
const (
	// preserveAllLastSectionPairs determines whether to keep all pairs in the last section
	preserveAllLastSectionPairs = true

	// maxLastSectionByteSize defines the maximum byte size for last section (50 KB)
	maxLastSectionByteSize = 50 * 1024

	// maxSingleBodyPairByteSize defines the maximum byte size for a single body pair (16 KB)
	maxSingleBodyPairByteSize = 16 * 1024

	// useQAPairSummarization determines whether to use QA pair summarization
	useQAPairSummarization = false

	// maxQAPairSections defines the maximum QA pair sections to preserve
	maxQAPairSections = 10

	// maxQAPairByteSize defines the maximum byte size for QA pair sections (64 KB)
	maxQAPairByteSize = 64 * 1024

	// summarizeHumanMessagesInQAPairs determines whether to summarize human messages in QA pairs
	summarizeHumanMessagesInQAPairs = false

	// lastSectionReservePercentage defines percentage of section size to reserve for future messages (25%)
	lastSectionReservePercentage = 25

	// keepMinLastQASections defines minimum number of QA sections to keep in the chain (1)
	keepMinLastQASections = 1

	// Default marker prefix for summarized content
	SummarizedContentPrefix = "**summarized content:**\n"
)

// SummarizerConfig defines the configuration for the summarizer
type SummarizerConfig struct {
	PreserveLast   bool
	UseQA          bool
	SummHumanInQA  bool
	LastSecBytes   int
	MaxBPBytes     int
	MaxQASections  int
	MaxQABytes     int
	KeepQASections int
}

// Summarizer is a wrapper around the summarizer configuration
type Summarizer interface {
	SummarizeChain(
		ctx context.Context,
		handler tools.SummarizeHandler,
		chain []llms.MessageContent,
	) ([]llms.MessageContent, error)
}

type summarizer struct {
	config SummarizerConfig
}

// NewSummarizer creates a new summarizer with the given configuration
func NewSummarizer(config SummarizerConfig) Summarizer {
	if config.PreserveLast {
		if config.LastSecBytes <= 0 {
			config.LastSecBytes = maxLastSectionByteSize
		}
	}

	if config.UseQA {
		if config.MaxQASections <= 0 {
			config.MaxQASections = maxQAPairSections
		}
		if config.MaxQABytes <= 0 {
			config.MaxQABytes = maxQAPairByteSize
		}
	}

	if config.MaxBPBytes <= 0 {
		config.MaxBPBytes = maxSingleBodyPairByteSize
	}

	if config.KeepQASections <= 0 {
		config.KeepQASections = keepMinLastQASections
	}

	return &summarizer{config: config}
}

// SummarizeChain takes a message chain and summarizes old messages to prevent context from growing too large
// Uses ChainAST with size tracking for efficient summarization decisions
func (s *summarizer) SummarizeChain(
	ctx context.Context,
	handler tools.SummarizeHandler,
	chain []llms.MessageContent,
) ([]llms.MessageContent, error) {
	// Skip summarization for empty chains
	if len(chain) == 0 {
		return chain, nil
	}

	// Parse chain into ChainAST with automatic size calculation
	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return chain, fmt.Errorf("failed to create ChainAST: %w", err)
	}

	// Apply different summarization strategies sequentially
	// Each function modifies the ast directly
	cfg := s.config

	// 0. All sections except last N should have exactly one Completion body pair
	err = summarizeSections(ctx, ast, handler, cfg.KeepQASections)
	if err != nil {
		return chain, fmt.Errorf("failed to summarize sections: %w", err)
	}

	// 1. Number of last sections rotation - manage active conversation size
	if cfg.PreserveLast {
		percent := lastSectionReservePercentage
		lastSectionIndexLeft := len(ast.Sections) - 1
		lastSectionIndexRight := len(ast.Sections) - cfg.KeepQASections
		for sdx := lastSectionIndexLeft; sdx >= lastSectionIndexRight && sdx >= 0; sdx-- {
			err = summarizeLastSection(ctx, ast, handler, sdx, cfg.LastSecBytes, cfg.MaxBPBytes, percent)
			if err != nil {
				return chain, fmt.Errorf("failed to summarize last section %d: %w", sdx, err)
			}
		}
	}

	// 2. QA-pair summarization - focus on question-answer sections
	if cfg.UseQA {
		err = summarizeQAPairs(ctx, ast, handler, cfg.MaxQASections, cfg.MaxQABytes, cfg.SummHumanInQA)
		if err != nil {
			return chain, fmt.Errorf("failed to summarize QA pairs: %w", err)
		}
	}

	return ast.Messages(), nil
}

// summarizeSections ensures all sections except the last N ones consist of a header
// and a single Completion-type body pair by summarizing multiple pairs if needed
func summarizeSections(
	ctx context.Context,
	ast *cast.ChainAST,
	handler tools.SummarizeHandler,
	keepQASections int,
) error {
	// Concurrent processing of sections summarization
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	ch := make(chan error, max(len(ast.Sections)-keepQASections, 0))
	defer close(ch)

	// Process all sections except the last N ones
	for i := 0; i < len(ast.Sections)-keepQASections; i++ {
		section := ast.Sections[i]

		// Skip if section already has just one of Summarization or Completion body pair
		if len(section.Body) == 1 && containsSummarizedContent(section.Body[0]) {
			continue
		}

		// Collect all messages from body pairs for summarization
		var messagesToSummarize []llms.MessageContent
		for _, pair := range section.Body {
			pairMessages := pair.Messages()
			messagesToSummarize = append(messagesToSummarize, pairMessages...)
		}

		// Skip if no messages to summarize
		if len(messagesToSummarize) == 0 {
			continue
		}

		// Add human message if it exists
		var humanMessages []llms.MessageContent
		if section.Header.HumanMessage != nil {
			humanMessages = append(humanMessages, *section.Header.HumanMessage)
		}

		wg.Add(1)
		go func(section *cast.ChainSection, i int) {
			defer wg.Done()

			// Generate summary
			summaryText, err := GenerateSummary(ctx, handler, humanMessages, messagesToSummarize)
			if err != nil {
				ch <- fmt.Errorf("section %d summary generation failed: %w", i, err)
				return
			}

			// Create a new Summarization body pair with the summary
			var summaryPair *cast.BodyPair
			switch t := determineTypeToSummarizedSection(section); t {
			case cast.Summarization:
				summaryPair = cast.NewBodyPairFromSummarization(summaryText)
			case cast.Completion:
				summaryPair = cast.NewBodyPairFromCompletion(SummarizedContentPrefix + summaryText)
			default:
				ch <- fmt.Errorf("invalid summarized section type: %d", t)
				return
			}

			mx.Lock()
			defer mx.Unlock()

			// Replace all body pairs with just the summary pair
			newSection := cast.NewChainSection(section.Header, []*cast.BodyPair{summaryPair})
			ast.Sections[i] = newSection
		}(section, i)
	}

	wg.Wait()

	// Check for any errors
	errs := make([]error, 0, len(ch))
	for edx := 0; edx < len(ch); edx++ {
		errs = append(errs, <-ch)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to summarize sections: %w", errors.Join(errs...))
	}

	return nil
}

// summarizeLastSection manages the size of the last (active) section
// by rotating older body pairs into a summary when the section exceeds size limits
func summarizeLastSection(
	ctx context.Context,
	ast *cast.ChainAST,
	handler tools.SummarizeHandler,
	numLastSection int,
	maxLastSectionBytes int,
	maxSingleBodyPairBytes int,
	reservePercent int,
) error {
	// Prevent out of bounds access
	if numLastSection >= len(ast.Sections) || numLastSection < 0 {
		return nil
	}

	lastSection := ast.Sections[numLastSection]

	// 1. First, handle oversized individual body pairs
	err := summarizeOversizedBodyPairs(ctx, lastSection, handler, maxSingleBodyPairBytes)
	if err != nil {
		return fmt.Errorf("failed to summarize oversized body pairs: %w", err)
	}

	// 2. If section is still under size limit, keep everything
	if lastSection.Size() <= maxLastSectionBytes {
		return nil
	}

	// 3. Determine which pairs to keep and which to summarize
	pairsToKeep, pairsToSummarize := determineLastSectionPairs(lastSection, maxLastSectionBytes, reservePercent)

	// 4. If we have pairs to summarize, create a summary
	if len(pairsToSummarize) > 0 {
		// Convert pairs to messages for summarization
		var messagesToSummarize []llms.MessageContent
		for _, pair := range pairsToSummarize {
			messagesToSummarize = append(messagesToSummarize, pair.Messages()...)
		}

		// Add human message if it exists
		var humanMessages []llms.MessageContent
		if lastSection.Header.HumanMessage != nil {
			humanMessages = append(humanMessages, *lastSection.Header.HumanMessage)
		}

		// Generate summary
		summaryText, err := GenerateSummary(ctx, handler, humanMessages, messagesToSummarize)
		if err != nil {
			// If summary generation fails, just keep the most recent messages
			lastSection.Body = pairsToKeep
			return fmt.Errorf("last section summary generation failed: %w", err)
		}

		// Create a new Summarization body pair with the summary
		var summaryPair *cast.BodyPair
		sectionToSummarize := cast.NewChainSection(lastSection.Header, pairsToSummarize)
		switch t := determineTypeToSummarizedSection(sectionToSummarize); t {
		case cast.Summarization:
			summaryPair = cast.NewBodyPairFromSummarization(summaryText)
		case cast.Completion:
			summaryPair = cast.NewBodyPairFromCompletion(SummarizedContentPrefix + summaryText)
		default:
			return fmt.Errorf("invalid summarized section type: %d", t)
		}

		// Replace the body with summary pair followed by kept pairs
		newBody := []*cast.BodyPair{summaryPair}
		newBody = append(newBody, pairsToKeep...)

		// Create a new section with the same header but new body pairs
		newSection := cast.NewChainSection(lastSection.Header, newBody)

		// Update the last section
		ast.Sections[numLastSection] = newSection
	}

	return nil
}

// determineTypeToSummarizedSection determines the type of each body pair to summarize
// based on the type of the body pairs in the section
// if all body pairs are Completion, return Completion, otherwise return Summarization
func determineTypeToSummarizedSection(section *cast.ChainSection) cast.BodyPairType {
	summarizedType := cast.Completion
	for _, pair := range section.Body {
		if pair.Type == cast.Summarization || pair.Type == cast.RequestResponse {
			summarizedType = cast.Summarization
			break
		}
	}

	return summarizedType
}

// determineTypeToSummarizedSections determines the type of each body pair to summarize
// based on the type of the body pairs in the sections to summarize
// if all sections are Completion, return Completion, otherwise return Summarization
func determineTypeToSummarizedSections(sections []*cast.ChainSection) cast.BodyPairType {
	summarizedType := cast.Completion
	for _, section := range sections {
		sectionType := determineTypeToSummarizedSection(section)
		if sectionType == cast.Summarization || sectionType == cast.RequestResponse {
			summarizedType = cast.Summarization
			break
		}
	}

	return summarizedType
}

// summarizeOversizedBodyPairs handles individual body pairs that exceed the maximum size
// by summarizing them in place, before the main pair selection logic runs
func summarizeOversizedBodyPairs(
	ctx context.Context,
	section *cast.ChainSection,
	handler tools.SummarizeHandler,
	maxBodyPairBytes int,
) error {
	if len(section.Body) == 0 {
		return nil
	}

	// Concurrent processing of body pairs summarization
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}

	// Map of body pairs that have been summarized
	bodyPairsSummarized := make(map[int]*cast.BodyPair)

	// Process each body pair
	for i, pair := range section.Body {
		// Skip pairs that are already summarized content or under the size limit
		if pair.Size() <= maxBodyPairBytes || containsSummarizedContent(pair) {
			continue
		}

		// Convert to messages
		pairMessages := pair.Messages()
		if len(pairMessages) == 0 {
			continue
		}

		// Add human message if it exists
		var humanMessages []llms.MessageContent
		if section.Header.HumanMessage != nil {
			humanMessages = append(humanMessages, *section.Header.HumanMessage)
		}

		wg.Add(1)
		go func(pair *cast.BodyPair, i int) {
			defer wg.Done()

			// Generate summary
			summaryText, err := GenerateSummary(ctx, handler, humanMessages, pairMessages)
			if err != nil {
				return // It's should collected next step in summarizeLastSection function
			}

			mx.Lock()
			defer mx.Unlock()

			// Create a new Summarization or Completion body pair with the summary
			// If the pair is a Completion, we need to create a new Completion pair
			// If the pair is a RequestResponse, we need to create a new Summarization pair
			if pair.Type == cast.RequestResponse {
				bodyPairsSummarized[i] = cast.NewBodyPairFromSummarization(summaryText)
			} else {
				bodyPairsSummarized[i] = cast.NewBodyPairFromCompletion(SummarizedContentPrefix + summaryText)
			}
		}(pair, i)
	}

	wg.Wait()

	// If any pairs were summarized, create a new section with the updated body
	// This ensures proper size calculation
	if len(bodyPairsSummarized) > 0 {
		for i, pair := range bodyPairsSummarized {
			section.Body[i] = pair
		}
		newSection := cast.NewChainSection(section.Header, section.Body)
		*section = *newSection
	}

	return nil
}

// containsSummarizedContent checks if a body pair contains summarized content
// Local helper function to avoid naming conflicts with test utilities
func containsSummarizedContent(pair *cast.BodyPair) bool {
	if pair == nil {
		return false
	}

	switch pair.Type {
	case cast.Summarization:
		return true
	case cast.RequestResponse:
		return false
	case cast.Completion:
		if pair.AIMessage == nil || len(pair.AIMessage.Parts) == 0 {
			return false
		}

		textContent, ok := pair.AIMessage.Parts[0].(llms.TextContent)
		if !ok {
			return false
		}

		if strings.HasPrefix(textContent.Text, SummarizedContentPrefix) {
			return true
		}

		return false
	default:
		return false
	}
}

// summarizeQAPairs handles QA pair summarization strategy
// focusing on summarizing older question-answer sections as needed
func summarizeQAPairs(
	ctx context.Context,
	ast *cast.ChainAST,
	handler tools.SummarizeHandler,
	maxQASections int,
	maxQABytes int,
	summarizeHuman bool,
) error {
	// Skip if limits aren't exceeded
	if !exceedsQASectionLimits(ast, maxQASections, maxQABytes) {
		return nil
	}

	// Identify sections to summarize
	humanMessages, aiMessages := prepareQASectionsForSummarization(ast, maxQASections, maxQABytes)
	if len(humanMessages) == 0 && len(aiMessages) == 0 {
		return nil
	}

	// Generate human message summary if it exists and needed
	var humanMsg *llms.MessageContent
	if len(humanMessages) > 0 {
		if summarizeHuman {
			humanSummary, err := GenerateSummary(ctx, handler, humanMessages, nil)
			if err != nil {
				return fmt.Errorf("QA (human) summary generation failed: %w", err)
			}
			msg := llms.TextParts(llms.ChatMessageTypeHuman, humanSummary)
			humanMsg = &msg
		} else {
			humanMsg = &llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
			}
			for _, msg := range humanMessages {
				humanMsg.Parts = append(humanMsg.Parts, msg.Parts...)
			}
		}
	}

	// Generate summary
	aiSummary, err := GenerateSummary(ctx, handler, humanMessages, aiMessages)
	if err != nil {
		return fmt.Errorf("QA (ai) summary generation failed: %w", err)
	}

	// Create a new AST with summary + recent sections
	sectionsToKeep := determineRecentSectionsToKeep(ast, maxQASections, maxQABytes)

	// Create a summarization body pair with the generated summary
	var summaryPair *cast.BodyPair
	sectionsToSummarize := ast.Sections[:len(ast.Sections)-sectionsToKeep]
	switch t := determineTypeToSummarizedSections(sectionsToSummarize); t {
	case cast.Summarization:
		summaryPair = cast.NewBodyPairFromSummarization(aiSummary)
	case cast.Completion:
		summaryPair = cast.NewBodyPairFromCompletion(SummarizedContentPrefix + aiSummary)
	default:
		return fmt.Errorf("invalid summarized section type: %d", t)
	}

	// Create a new AST
	newAST := &cast.ChainAST{
		Sections: make([]*cast.ChainSection, 0, sectionsToKeep+1), // +1 for summary section
	}

	// Add the summary section (with system message if it exists)
	var systemMsg *llms.MessageContent
	if len(ast.Sections) > 0 && ast.Sections[0].Header.SystemMessage != nil {
		systemMsg = ast.Sections[0].Header.SystemMessage
	}

	summaryHeader := cast.NewHeader(systemMsg, humanMsg)
	summarySection := cast.NewChainSection(summaryHeader, []*cast.BodyPair{summaryPair})
	newAST.AddSection(summarySection)

	// Add the most recent sections that should be kept
	totalSections := len(ast.Sections)
	if sectionsToKeep > 0 && totalSections > 0 {
		for i := totalSections - sectionsToKeep; i < totalSections; i++ {
			// Copy the section but ensure no system message (already added in summary section)
			section := ast.Sections[i]
			newHeader := cast.NewHeader(nil, section.Header.HumanMessage)
			newSection := cast.NewChainSection(newHeader, section.Body)
			newAST.AddSection(newSection)
		}
	}

	// Replace the original AST with the new one
	ast.Sections = newAST.Sections

	return nil
}

// exceedsQASectionLimits checks if QA sections exceed the configured limits
func exceedsQASectionLimits(ast *cast.ChainAST, maxSections int, maxBytes int) bool {
	return len(ast.Sections) > maxSections || ast.Size() > maxBytes
}

// prepareQASectionsForSummarization prepares QA sections for summarization
// returns human and ai messages separately for better control over the summarization process
func prepareQASectionsForSummarization(
	ast *cast.ChainAST,
	maxSections int,
	maxBytes int,
) ([]llms.MessageContent, []llms.MessageContent) {
	totalSections := len(ast.Sections)
	if totalSections == 0 {
		return nil, nil
	}

	// Calculate how many recent sections to keep
	sectionsToKeep := determineRecentSectionsToKeep(ast, maxSections, maxBytes)

	// Select oldest sections for summarization
	sectionsToSummarize := ast.Sections[:totalSections-sectionsToKeep]
	if len(sectionsToSummarize) == 0 {
		return nil, nil
	}
	if len(sectionsToSummarize) == 1 && len(sectionsToSummarize[0].Body) == 1 &&
		sectionsToSummarize[0].Body[0].Type == cast.Summarization {
		return nil, nil
	}

	// Convert selected sections to messages for summarization
	humanMessages := convertSectionsHeadersToMessages(sectionsToSummarize)
	aiMessages := convertSectionsPairsToMessages(sectionsToSummarize)

	return humanMessages, aiMessages
}

// determineRecentSectionsToKeep determines how many recent sections to preserve
func determineRecentSectionsToKeep(ast *cast.ChainAST, maxSections int, maxBytes int) int {
	totalSections := len(ast.Sections)
	keepCount := 0
	currentSize := 0

	// Reserve buffer space to ensure we don't exceed max bytes
	const bufferSpace = 1000
	effectiveMaxBytes := maxBytes - bufferSpace

	// Start from most recent sections (end of array) and work backwards
	for i := totalSections - 1; i >= 0; i-- {
		// Stop if we've reached max sections to keep
		if keepCount >= maxSections {
			break
		}

		sectionSize := ast.Sections[i].Size()

		// Stop if adding this section would exceed byte limit
		if currentSize+sectionSize > effectiveMaxBytes {
			break
		}

		currentSize += sectionSize
		keepCount++
	}

	return keepCount
}

// convertSectionsHeadersToMessages extracts human messages from sections for summarization
func convertSectionsHeadersToMessages(sections []*cast.ChainSection) []llms.MessageContent {
	if len(sections) == 0 {
		return nil
	}

	var messages []llms.MessageContent

	for _, section := range sections {
		// Add human message if it exists
		if section.Header.HumanMessage != nil {
			messages = append(messages, *section.Header.HumanMessage)
		}
	}

	return messages
}

// convertSectionsPairsToMessages extracts ai messages from sections for summarization
func convertSectionsPairsToMessages(sections []*cast.ChainSection) []llms.MessageContent {
	if len(sections) == 0 {
		return nil
	}

	var messages []llms.MessageContent

	for _, section := range sections {
		// Get all messages from each body pair using the Messages() method
		for _, pair := range section.Body {
			pairMessages := pair.Messages()
			messages = append(messages, pairMessages...)
		}
	}

	return messages
}

// determineLastSectionPairs splits the last section's pairs into those to keep and those to summarize
func determineLastSectionPairs(
	section *cast.ChainSection,
	maxBytes int,
	reservePercent int,
) ([]*cast.BodyPair, []*cast.BodyPair) {
	var pairsToKeep []*cast.BodyPair
	var pairsToSummarize []*cast.BodyPair

	// Start with header size as the base size
	currentSize := section.Header.Size()

	// Calculate threshold with reserve some percentage of maxBytes
	// This should result in less frequent summaries
	threshold := maxBytes * (100 - reservePercent) / 100

	// To ensure we have at least some pairs, if there are any
	if len(section.Body) > 0 {
		// Always keep the last (most recent) pair
		pairsToKeep = make([]*cast.BodyPair, 0, len(section.Body))
		lastPair := section.Body[len(section.Body)-1]
		pairsToKeep = append(pairsToKeep, lastPair)
		currentSize += lastPair.Size()
		summarizeSize := 0

		// Process pairs in reverse order (newest to oldest), starting from the second-to-last
		borderFound := false
		for i := len(section.Body) - 2; i >= 0; i-- {
			pair := section.Body[i]
			pairSize := pair.Size()

			// If adding this pair would fit within our threshold, keep it
			if currentSize+pairSize <= threshold && !borderFound {
				pairsToKeep = append(pairsToKeep, pair)
				currentSize += pairSize
			} else {
				pairsToSummarize = append(pairsToSummarize, pair)
				summarizeSize += pairSize
				borderFound = true
			}
		}

		// Reverse slices to get them in original order (oldest first)
		slices.Reverse(pairsToSummarize)
		slices.Reverse(pairsToKeep)

		if currentSize+summarizeSize <= maxBytes {
			pairsToKeep = append(pairsToSummarize, pairsToKeep...)
			pairsToSummarize = nil
		}
	}

	// Prevent double summarization of the last pair
	if len(pairsToSummarize) == 1 && pairsToSummarize[0].Type == cast.Summarization {
		pairsToKeep = append(pairsToSummarize, pairsToKeep...)
		pairsToSummarize = nil
	}

	return pairsToKeep, pairsToSummarize
}

// GenerateSummary generates a summary of the provided messages
func GenerateSummary(
	ctx context.Context,
	handler tools.SummarizeHandler,
	humanMessages []llms.MessageContent,
	aiMessages []llms.MessageContent,
) (string, error) {
	if handler == nil {
		return "", fmt.Errorf("summarizer handler cannot be nil")
	}

	if len(humanMessages) == 0 && len(aiMessages) == 0 {
		return "", fmt.Errorf("cannot summarize empty message list")
	}

	// Convert messages to text format optimized for summarization
	text := messagesToPrompt(humanMessages, aiMessages)

	// Generate the summary using provided summarizer handler
	summary, err := handler(ctx, text)
	if err != nil {
		return "", fmt.Errorf("summarization failed: %w", err)
	}

	return summary, nil
}

// messagesToPrompt converts a slice of messages to a text representation
func messagesToPrompt(humanMessages []llms.MessageContent, aiMessages []llms.MessageContent) string {
	var buffer strings.Builder

	humanMessagesText := humanMessagesToText(humanMessages)
	aiMessagesText := aiMessagesToText(aiMessages)

	// case 0: no messages
	if len(humanMessages) == 0 && len(aiMessages) == 0 {
		return "nothing to summarize"
	}

	// case 1: use human messages as a context for ai messages
	if len(humanMessages) > 0 && len(aiMessages) > 0 {
		instructions := getSummarizationInstructions(1)
		buffer.WriteString(fmt.Sprintf("<instructions>%s</instructions>\n\n", instructions))
		buffer.WriteString(humanMessagesText)
		buffer.WriteString(aiMessagesText)
	}

	// case 2: use ai messages as a content to summarize without context
	if len(aiMessages) > 0 && len(humanMessages) == 0 {
		instructions := getSummarizationInstructions(2)
		buffer.WriteString(fmt.Sprintf("<instructions>%s</instructions>\n\n", instructions))
		buffer.WriteString(aiMessagesText)
	}

	// case 3: use human messages as a instructions to summarize them
	if len(humanMessages) > 0 && len(aiMessages) == 0 {
		instructions := getSummarizationInstructions(3)
		buffer.WriteString(fmt.Sprintf("<instructions>%s</instructions>\n\n", instructions))
		buffer.WriteString(humanMessagesText)
	}

	return buffer.String()
}

// getSummarizationInstructions returns the summarization instructions for the given case
func getSummarizationInstructions(sumCase int) string {
	switch sumCase {
	case 1:
		return fmt.Sprintf(`
SUMMARIZATION TASK: Create a concise summary of AI responses while preserving essential information from the conversation context.

DATA STRUCTURE:
- <tasks> contains user queries that provide critical context for understanding AI responses
- <messages> contains AI responses that need to be summarized

HANDLING PREVIOUSLY SUMMARIZED CONTENT:
When you encounter a sequence of messages where:
1. A message contains <tool_call name="%s"> 
2. Followed by a message with role="tool" containing execution history

This pattern is a crucial signal - it means you're looking at ALREADY summarized information. When you see this:
1. MUST treat this summarized content as HIGH PRIORITY
2. Extract and PRESERVE the key technical details (commands, parameters, errors, results)
3. Integrate this information into your new summary without duplicating
4. Understand that this summary already represents multiple previous interactions and essential technical details

KEY REQUIREMENTS:
1. Preserve ALL technical details: function names, parameters, file paths, URLs, versions, numerical values
2. Maintain complete code examples that demonstrate implementation
3. Keep intact any step-by-step instructions or procedures
4. Ensure the summary directly addresses the user queries found in <tasks>
5. Organize information in a logical flow that matches the problem-solution structure
6. NEVER include context in the summary, just the summarized content, use context only to understand the <messages>
`, cast.SummarizationToolName)

	case 2:
		return fmt.Sprintf(`
SUMMARIZATION TASK: Distill standalone AI responses into a comprehensive yet concise summary.

DATA STRUCTURE:
- <messages> contains AI responses that need to be summarized without user context

HANDLING PREVIOUSLY SUMMARIZED CONTENT:
When you encounter a sequence of messages where:
1. A message contains <tool_call name="%s"> 
2. Followed by a message with role="tool" containing execution history

This pattern is a crucial signal - it means you're looking at ALREADY summarized information. When you see this:
1. MUST treat this summarized content as HIGH PRIORITY
2. Extract and PRESERVE the key technical details (commands, parameters, errors, results)
3. Integrate this information into your new summary without duplicating
4. Understand that this summary already represents multiple previous interactions and essential technical details

KEY REQUIREMENTS:
1. Ensure the summary is self-contained and provides complete context
2. Preserve ALL technical details: function names, parameters, file paths, URLs, versions, numerical values
3. Maintain complete code examples that demonstrate implementation
4. Identify and prioritize main conclusions, recommendations, and technical explanations
5. Organize information in a logical, sequential structure
`, cast.SummarizationToolName)

	case 3:
		return `
SUMMARIZATION TASK: Extract key requirements and context from user queries.

DATA STRUCTURE:
- <tasks> contains user messages that need to be summarized

KEY REQUIREMENTS:
1. Identify primary goals, questions, and objectives expressed by the user
2. Preserve ALL technical specifications: function names, parameters, file paths, URLs, versions
3. Maintain all constraints, requirements, and success criteria mentioned
4. Capture the complete problem context and any background information provided
5. Organize requirements in order of stated or implied priority
6. USE directive forms and imperative mood for better translate original text
`
	default:
		return ""
	}
}

// humanMessagesToText converts a slice of human messages to a text representation
func humanMessagesToText(humanMessages []llms.MessageContent) string {
	var buffer strings.Builder

	buffer.WriteString("<tasks>\n")
	for mdx, msg := range humanMessages {
		if msg.Role != llms.ChatMessageTypeHuman {
			continue
		}
		buffer.WriteString(fmt.Sprintf("<task id=\"%d\">\n", mdx))
		for _, part := range msg.Parts {
			switch v := part.(type) {
			case llms.TextContent:
				buffer.WriteString(fmt.Sprintf("%s\n", v.Text))
			case llms.ImageURLContent:
				buffer.WriteString(fmt.Sprintf("<image url=\"%s\">\n", v.URL))
				if v.Detail != "" {
					buffer.WriteString(fmt.Sprintf("%s\n", v.Detail))
				}
				buffer.WriteString("</image>\n")
			case llms.BinaryContent:
				buffer.WriteString(fmt.Sprintf("<binary mime=\"%s\">\n", v.MIMEType))
				if v.Data != nil {
					data := hex.EncodeToString(v.Data[:min(len(v.Data), 100)])
					buffer.WriteString(fmt.Sprintf("first 100 bytes in hex: %s\n", data))
				}
				buffer.WriteString("</binary>\n")
			}
		}
		buffer.WriteString("</task>\n")
	}
	buffer.WriteString("</tasks>\n")

	return buffer.String()
}

// aiMessagesToText converts a slice of ai messages to a text representation
func aiMessagesToText(aiMessages []llms.MessageContent) string {
	var buffer strings.Builder

	buffer.WriteString("<messages>\n")
	for mdx, msg := range aiMessages {
		buffer.WriteString(fmt.Sprintf("<message id=\"%d\" role=\"%s\">\n", mdx, msg.Role))
		for pdx, part := range msg.Parts {
			partNum := fmt.Sprintf("part=\"%d\"", pdx)
			switch v := part.(type) {
			case llms.TextContent:
				buffer.WriteString(fmt.Sprintf("<content %s>\n", partNum))
				buffer.WriteString(fmt.Sprintf("%s\n", v.Text))
				buffer.WriteString("</content>\n")
			case llms.ToolCall:
				if v.FunctionCall != nil {
					buffer.WriteString(fmt.Sprintf("<tool_call name=\"%s\" %s>\n", v.FunctionCall.Name, partNum))
					buffer.WriteString(fmt.Sprintf("%s\n", v.FunctionCall.Arguments))
					buffer.WriteString("</tool_call>\n")
				}
			case llms.ToolCallResponse:
				buffer.WriteString(fmt.Sprintf("<tool_call_response name=\"%s\" %s>\n", v.Name, partNum))
				buffer.WriteString(fmt.Sprintf("%s\n", v.Content))
				buffer.WriteString("</tool_call_response>\n")
			case llms.ImageURLContent:
				buffer.WriteString(fmt.Sprintf("<image url=\"%s\" %s>\n", v.URL, partNum))
				if v.Detail != "" {
					buffer.WriteString(fmt.Sprintf("%s\n", v.Detail))
				}
				buffer.WriteString("</image>\n")
			case llms.BinaryContent:
				buffer.WriteString(fmt.Sprintf("<binary mime=\"%s\" %s>\n", v.MIMEType, partNum))
				if v.Data != nil {
					data := hex.EncodeToString(v.Data[:min(len(v.Data), 100)])
					buffer.WriteString(fmt.Sprintf("first 100 bytes in hex: %s\n", data))
				}
				buffer.WriteString("</binary>\n")
			}
		}
		buffer.WriteString("</message>\n")
	}
	buffer.WriteString("</messages>")

	return buffer.String()
}
