package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"

	"github.com/google/uuid"
	"github.com/vxcontrol/langchaingo/llms"
)

const (
	maxRetries          = 5
	sampleCount         = 5
	testFunctionName    = "get_number"
	patternFunctionName = "submit_pattern"
)

var cacheTemplates sync.Map

// attemptRecord stores information about a failed pattern detection attempt
type attemptRecord struct {
	Template string
	Error    string
}

func lookupInCache(provider Provider) (string, bool) {
	if template, ok := cacheTemplates.Load(provider.Type()); ok {
		if template, ok := template.(string); ok {
			return template, true
		}
	}
	return "", false
}

func storeInCache(provider Provider, template string) {
	cacheTemplates.Store(provider.Type(), template)
}

// DetermineToolCallIDTemplate analyzes tool call ID format by collecting samples
// and using AI to detect the pattern, with fallback to heuristic analysis
func DetermineToolCallIDTemplate(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	prompter templates.Prompter,
) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	agent := observation.Agent(
		langfuse.WithAgentName("tool call ID template detector"),
		langfuse.WithAgentInput(map[string]any{
			"provider":   provider.Type(),
			"agent_type": string(opt),
		}),
	)
	ctx, _ = agent.Observation(ctx)
	wrapEndAgentSpan := func(template, status string, err error) (string, error) {
		if err != nil {
			agent.End(
				langfuse.WithAgentStatus(err.Error()),
				langfuse.WithAgentLevel(langfuse.ObservationLevelError),
			)
		} else {
			agent.End(
				langfuse.WithAgentOutput(template),
				langfuse.WithAgentStatus(status),
			)
		}

		return template, err
	}

	// Step 0: Check if template is already in cache
	if template, ok := lookupInCache(provider); ok {
		return wrapEndAgentSpan(template, "found in cache", nil)
	}

	// Step 1: Collect 5 sample tool call IDs in parallel
	samples, err := collectToolCallIDSamples(ctx, provider, opt, prompter)
	if err != nil {
		return wrapEndAgentSpan("", "", fmt.Errorf("failed to collect tool call ID samples: %w", err))
	}

	if len(samples) == 0 {
		return wrapEndAgentSpan("", "", fmt.Errorf("no tool call ID samples collected"))
	}

	// Step 2-4: Try to detect pattern using AI with retry logic
	var previousAttempts []attemptRecord
	for attempt := range maxRetries {
		template, newSample, err := detectPatternWithAI(ctx, provider, opt, prompter, samples, previousAttempts)
		if err != nil {
			// Record the failure - agent didn't call the function or other error occurred
			previousAttempts = append(previousAttempts, attemptRecord{
				Template: "<no template - agent failed to call function>",
				Error:    err.Error(),
			})

			// If AI detection completely fails, use fallback
			if attempt == maxRetries-1 {
				template = fallbackHeuristicDetection(samples)
				storeInCache(provider, template)
				return wrapEndAgentSpan(template, "partially detected", nil)
			}
			continue
		}

		// Add new sample from detector call
		allSamples := append(samples, newSample)

		// Validate template against all samples
		validationErr := templates.ValidatePattern(template, allSamples)
		if validationErr == nil {
			storeInCache(provider, template)
			return wrapEndAgentSpan(template, "validated", nil)
		}

		// Validation failed, record attempt and retry
		previousAttempts = append(previousAttempts, attemptRecord{
			Template: template,
			Error:    validationErr.Error(),
		})

		// Update samples to include the new one for next iteration
		samples = allSamples
	}

	// All retries exhausted, use fallback heuristic
	template := fallbackHeuristicDetection(samples)
	storeInCache(provider, template)
	return wrapEndAgentSpan(template, "fallback heuristic detection", nil)
}

// collectToolCallIDSamples collects tool call ID samples in parallel
func collectToolCallIDSamples(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	prompter templates.Prompter,
) ([]templates.PatternSample, error) {
	type sampleResult struct {
		samples []templates.PatternSample
		err     error
	}

	results := make(chan sampleResult, sampleCount)
	var wg sync.WaitGroup

	// Launch parallel goroutines to collect samples
	for range sampleCount {
		wg.Add(1)
		go func() {
			defer wg.Done()

			samples, err := runToolCallIDCollector(ctx, provider, opt, prompter)
			results <- sampleResult{samples: samples, err: err}
		}()
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results - use map to deduplicate by Value
	samplesMap := make(map[string]templates.PatternSample)
	var errs []error
	for result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
		} else {
			for _, sample := range result.samples {
				samplesMap[sample.Value] = sample
			}
		}
	}

	samples := make([]templates.PatternSample, 0, len(samplesMap))
	for _, sample := range samplesMap {
		samples = append(samples, sample)
	}

	// Sort by value for consistency
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Value < samples[j].Value
	})

	// Return error only if we got no samples at all
	if len(samples) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all sample collection attempts failed: %w", errors.Join(errs...))
	}

	return samples, nil
}

// runToolCallIDCollector collects a single tool call ID sample
func runToolCallIDCollector(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	prompter templates.Prompter,
) ([]templates.PatternSample, error) {
	// Generate random context to prevent caching
	randomContext := uuid.New().String()

	// Render collector prompt
	prompt, err := prompter.RenderTemplate(templates.PromptTypeToolCallIDCollector, map[string]any{
		"FunctionName":  testFunctionName,
		"RandomContext": randomContext,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render collector prompt: %w", err)
	}

	// Create test tool
	testTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        testFunctionName,
			Description: "Get a number value",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"value": map[string]any{
						"type":        "integer",
						"description": "The number value",
					},
				},
				"required": []string{"value"},
			},
		},
	}

	// Call LLM with tool
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
	}

	response, err := provider.CallWithTools(ctx, opt, messages, []llms.Tool{testTool}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	sampleMap := make(map[string]templates.PatternSample)

	// Extract tool call ID and function name
	for _, choice := range response.Choices {
		for _, toolCall := range choice.ToolCalls {
			if toolCall.ID != "" {
				functionName := ""
				if toolCall.FunctionCall != nil {
					functionName = toolCall.FunctionCall.Name
				}
				sampleMap[toolCall.ID] = templates.PatternSample{
					Value:        toolCall.ID,
					FunctionName: functionName,
				}
			}
		}
	}

	samples := make([]templates.PatternSample, 0, len(sampleMap))
	for _, sample := range sampleMap {
		samples = append(samples, sample)
	}

	// Sort by value for consistency
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Value < samples[j].Value
	})

	return samples, nil
}

// detectPatternWithAI uses AI to analyze samples and detect pattern template
func detectPatternWithAI(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	prompter templates.Prompter,
	samples []templates.PatternSample,
	previousAttempts []attemptRecord,
) (string, templates.PatternSample, error) {
	// Extract just the values for the prompt
	sampleValues := make([]string, len(samples))
	for i, s := range samples {
		sampleValues[i] = s.Value
	}

	// Render detector prompt
	prompt, err := prompter.RenderTemplate(templates.PromptTypeToolCallIDDetector, map[string]any{
		"FunctionName":     patternFunctionName,
		"Samples":          sampleValues,
		"PreviousAttempts": previousAttempts,
	})
	if err != nil {
		return "", templates.PatternSample{}, fmt.Errorf("failed to render detector prompt: %w", err)
	}

	// Create pattern submission tool
	patternTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        patternFunctionName,
			Description: "Submit the detected pattern template",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"template": map[string]any{
						"type":        "string",
						"description": "The pattern template in format like 'toolu_{r:24:b}' or 'call_{r:24:x}' or '{f}:{r:1:d}'",
					},
				},
				"required": []string{"template"},
			},
		},
	}

	// Call LLM with tool
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
	}

	response, err := provider.CallWithTools(ctx, opt, messages, []llms.Tool{patternTool}, nil)
	if err != nil {
		return "", templates.PatternSample{}, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Extract template and new tool call ID from response
	var detectedTemplate string
	var newSample templates.PatternSample

	for _, choice := range response.Choices {
		for _, toolCall := range choice.ToolCalls {
			if toolCall.ID != "" {
				newSample.Value = toolCall.ID
				if toolCall.FunctionCall != nil {
					newSample.FunctionName = toolCall.FunctionCall.Name
				}
			}
			if toolCall.FunctionCall != nil && toolCall.FunctionCall.Name == patternFunctionName {
				// Parse arguments to get template
				var args struct {
					Template string `json:"template"`
				}
				if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err == nil {
					detectedTemplate = args.Template
				}
			}
		}
	}

	if detectedTemplate == "" {
		return "", templates.PatternSample{}, fmt.Errorf("no template found in AI response")
	}
	if newSample.Value == "" {
		return "", templates.PatternSample{}, fmt.Errorf("no tool call ID found in AI response")
	}

	return detectedTemplate, newSample, nil
}

// fallbackHeuristicDetection performs character-by-character analysis to build pattern
func fallbackHeuristicDetection(samples []templates.PatternSample) string {
	if len(samples) == 0 {
		return ""
	}

	// Extract values from samples
	values := make([]string, len(samples))
	for i, s := range samples {
		values[i] = s.Value
	}

	// Find minimum length
	minLen := len(values[0])
	for _, value := range values[1:] {
		if len(value) < minLen {
			minLen = len(value)
		}
	}

	var pattern strings.Builder
	pos := 0

	for pos < minLen {
		// Get character at position from all values
		chars := make([]byte, len(values))
		for i, value := range values {
			chars[i] = value[pos]
		}

		// Check if all characters are the same (literal)
		allSame := true
		firstChar := chars[0]
		for _, ch := range chars[1:] {
			if ch != firstChar {
				allSame = false
				break
			}
		}

		if allSame {
			// Collect all consecutive literal characters
			for pos < minLen {
				chars := make([]byte, len(values))
				for i, value := range values {
					chars[i] = value[pos]
				}

				allSame := true
				firstChar := chars[0]
				for _, ch := range chars[1:] {
					if ch != firstChar {
						allSame = false
						break
					}
				}

				if !allSame {
					break
				}

				pattern.WriteByte(firstChar)
				pos++
			}
		} else {
			// Random part - collect all consecutive random characters
			var allCharsInRandom [][]byte

			// Collect all random characters until we hit a literal
			for pos < minLen {
				chars := make([]byte, len(values))
				for i, value := range values {
					chars[i] = value[pos]
				}

				// Check if this position is literal
				allSame := true
				firstChar := chars[0]
				for _, ch := range chars[1:] {
					if ch != firstChar {
						allSame = false
						break
					}
				}

				if allSame {
					break
				}

				allCharsInRandom = append(allCharsInRandom, chars)
				pos++
			}

			// Determine charset for all collected random characters
			if len(allCharsInRandom) > 0 {
				charset := determineCommonCharset(allCharsInRandom)
				pattern.WriteString(fmt.Sprintf("{r:%d:%s}", len(allCharsInRandom), charset))
			}
		}
	}

	return pattern.String()
}

// determineCommonCharset finds charset that covers all character sets across positions
func determineCommonCharset(allCharsPerPosition [][]byte) string {
	hasDigit := false
	hasLower := false
	hasUpper := false

	// Check all positions
	for _, chars := range allCharsPerPosition {
		for _, ch := range chars {
			if ch >= '0' && ch <= '9' {
				hasDigit = true
			} else if ch >= 'a' && ch <= 'z' {
				hasLower = true
			} else if ch >= 'A' && ch <= 'Z' {
				hasUpper = true
			}
		}
	}

	// Determine minimal charset that covers all
	if hasDigit && !hasLower && !hasUpper {
		return "d" // digit
	}
	if !hasDigit && hasLower && !hasUpper {
		// Check if hex lowercase
		isHex := true
		for _, chars := range allCharsPerPosition {
			for _, ch := range chars {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
					isHex = false
					break
				}
			}
			if !isHex {
				break
			}
		}
		if isHex {
			return "h" // hex lowercase
		}
		return "l" // lower
	}
	if !hasDigit && !hasLower && hasUpper {
		// Check if hex uppercase
		isHex := true
		for _, chars := range allCharsPerPosition {
			for _, ch := range chars {
				if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
					isHex = false
					break
				}
			}
			if !isHex {
				break
			}
		}
		if isHex {
			return "H" // hex uppercase
		}
		return "u" // upper
	}
	if !hasDigit && hasLower && hasUpper {
		return "a" // alpha
	}
	if hasDigit && hasLower && !hasUpper {
		// Check if hex lowercase
		isHex := true
		for _, chars := range allCharsPerPosition {
			for _, ch := range chars {
				if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
					isHex = false
					break
				}
			}
			if !isHex {
				break
			}
		}
		if isHex {
			return "h" // hex lowercase
		}
		return "x" // alnum
	}
	if hasDigit && !hasLower && hasUpper {
		// Check if hex uppercase
		isHex := true
		for _, chars := range allCharsPerPosition {
			for _, ch := range chars {
				if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
					isHex = false
					break
				}
			}
			if !isHex {
				break
			}
		}
		if isHex {
			return "H" // hex uppercase
		}
		return "x" // alnum
	}

	// All three: base62
	return "b"
}

// determineMinimalCharset finds the minimal charset that covers all characters
func determineMinimalCharset(chars []byte) string {
	hasDigit := false
	hasLower := false
	hasUpper := false

	for _, ch := range chars {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		} else if ch >= 'a' && ch <= 'z' {
			hasLower = true
		} else if ch >= 'A' && ch <= 'Z' {
			hasUpper = true
		}
	}

	// Determine minimal charset
	if hasDigit && !hasLower && !hasUpper {
		return "d" // digit
	}
	if !hasDigit && hasLower && !hasUpper {
		return "l" // lower
	}
	if !hasDigit && !hasLower && hasUpper {
		return "u" // upper
	}
	if !hasDigit && hasLower && hasUpper {
		return "a" // alpha
	}

	// Check for hex (lowercase)
	isHexLower := true
	for _, ch := range chars {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			isHexLower = false
			break
		}
	}
	if isHexLower && !hasUpper {
		return "h" // hex lowercase
	}

	// Check for hex (uppercase)
	isHexUpper := true
	for _, ch := range chars {
		if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
			isHexUpper = false
			break
		}
	}
	if isHexUpper && !hasLower {
		return "H" // hex uppercase
	}

	// Alphanumeric or base62
	if hasDigit && hasLower && hasUpper {
		return "b" // base62
	}

	return "x" // alnum (fallback)
}
