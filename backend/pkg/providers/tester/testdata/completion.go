package testdata

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type testCaseCompletion struct {
	def TestDefinition

	// state for streaming and response collection
	mu        sync.Mutex
	content   strings.Builder
	reasoning strings.Builder
	expected  string
	messages  []llms.MessageContent
}

func newCompletionTestCase(def TestDefinition) (TestCase, error) {
	expected, ok := def.Expected.(string)
	if !ok {
		return nil, fmt.Errorf("completion test expected must be string")
	}

	// convert MessagesData to llms.MessageContent
	messages, err := def.Messages.ToMessageContent()
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %v", err)
	}

	return &testCaseCompletion{
		def:      def,
		expected: expected,
		messages: messages,
	}, nil
}

func (t *testCaseCompletion) ID() string                      { return t.def.ID }
func (t *testCaseCompletion) Name() string                    { return t.def.Name }
func (t *testCaseCompletion) Type() TestType                  { return t.def.Type }
func (t *testCaseCompletion) Group() TestGroup                { return t.def.Group }
func (t *testCaseCompletion) Streaming() bool                 { return t.def.Streaming }
func (t *testCaseCompletion) Prompt() string                  { return t.def.Prompt }
func (t *testCaseCompletion) Messages() []llms.MessageContent { return t.messages }
func (t *testCaseCompletion) Tools() []llms.Tool              { return nil }

func (t *testCaseCompletion) StreamingCallback() streaming.Callback {
	if !t.def.Streaming {
		return nil
	}

	return func(ctx context.Context, chunk streaming.Chunk) error {
		t.mu.Lock()
		defer t.mu.Unlock()

		t.content.WriteString(chunk.Content)
		t.reasoning.WriteString(chunk.ReasoningContent)
		return nil
	}
}

func (t *testCaseCompletion) Execute(response any, latency time.Duration) TestResult {
	result := TestResult{
		ID:        t.def.ID,
		Name:      t.def.Name,
		Type:      t.def.Type,
		Group:     t.def.Group,
		Streaming: t.def.Streaming,
		Latency:   latency,
	}

	var responseStr string
	var hasReasoning bool

	// handle different response types
	switch resp := response.(type) {
	case string:
		// direct string response from p.Call()
		responseStr = resp
	case *llms.ContentResponse:
		// response from p.CallEx() with messages
		if len(resp.Choices) == 0 {
			result.Success = false
			result.Error = fmt.Errorf("empty response from model")
			return result
		}

		choice := resp.Choices[0]
		responseStr = choice.Content

		// check for reasoning content
		if len(choice.ReasoningContent) > 0 {
			hasReasoning = true
		}
		if reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"]; ok {
			if tokens, ok := reasoningTokens.(int); ok && tokens > 0 {
				hasReasoning = true
			}
		}
	default:
		result.Success = false
		result.Error = fmt.Errorf("expected string or *llms.ContentResponse, got %T", response)
		return result
	}

	// check for streaming reasoning content
	if t.reasoning.Len() > 0 {
		hasReasoning = true
	}
	result.Reasoning = hasReasoning

	// validate response contains expected text using enhanced matching logic
	responseStr = strings.TrimSpace(responseStr)
	expected := strings.TrimSpace(t.expected)

	success := containsString(responseStr, expected)

	result.Success = success
	if !success {
		result.Error = fmt.Errorf("expected text '%s' not found", t.expected)
	}

	return result
}

// containsString implements enhanced string matching logic with combinatorial modifiers.
func containsString(response, expected string) bool {
	if len(response) == 0 {
		return false
	}

	// direct equality check first
	if response == expected {
		return true
	}

	// apply all possible combinations of modifiers and test each one
	return tryAllModifierCombinations(response, expected, 0, []stringModifier{})
}

type stringModifier func(string) string

// available modifiers - order may matter, so we preserve it for future extensibility
var availableModifiers = []stringModifier{
	normalizeCase,     // convert to lowercase
	removeWhitespace,  // remove all whitespace characters
	removeMarkdown,    // remove markdown formatting
	removePunctuation, // remove punctuation marks
	removeQuotes,      // remove various quote characters
	normalizeNumbers,  // normalize number sequences
}

// tryAllModifierCombinations recursively tries all possible combinations of modifiers
func tryAllModifierCombinations(response, expected string, startIdx int, currentModifiers []stringModifier) bool {
	// test current combination
	if testWithModifiers(response, expected, currentModifiers) {
		return true
	}

	// try adding each remaining modifier
	for i := startIdx; i < len(availableModifiers); i++ {
		newModifiers := append(currentModifiers, availableModifiers[i])
		if tryAllModifierCombinations(response, expected, i+1, newModifiers) {
			return true
		}
	}

	return false
}

// testWithModifiers applies the given modifiers and tests for match
func testWithModifiers(response, expected string, modifiers []stringModifier) bool {
	modifiedResponse := applyModifiers(response, modifiers)
	modifiedExpected := applyModifiers(expected, modifiers)

	// bidirectional contains check
	return contains(modifiedResponse, modifiedExpected) || contains(modifiedExpected, modifiedResponse)
}

// applyModifiers applies all modifiers in sequence to the input string
// NOTE: Order of application may matter for future modifiers, so we preserve sequence
func applyModifiers(input string, modifiers []stringModifier) string {
	result := input
	for _, modifier := range modifiers {
		result = modifier(result)
	}
	return result
}

// contains checks if haystack contains needle
func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// Modifier implementations

func normalizeCase(s string) string {
	return strings.ToLower(s)
}

func removeWhitespace(s string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"\n", "",
		"\r", "",
		"\t", "",
		"\u00A0", "", // non-breaking space
	)
	return replacer.Replace(s)
}

func removeMarkdown(s string) string {
	// remove common markdown formatting in specific order to avoid conflicts
	result := s

	// remove code blocks first
	result = strings.ReplaceAll(result, "```", "")

	// remove bold/italic (order matters: ** before *)
	result = strings.ReplaceAll(result, "**", "")
	result = strings.ReplaceAll(result, "__", "")
	result = strings.ReplaceAll(result, "*", "")
	result = strings.ReplaceAll(result, "_", "")

	// remove other formatting
	result = strings.ReplaceAll(result, "~~", "") // strikethrough
	result = strings.ReplaceAll(result, "`", "")  // inline code
	result = strings.ReplaceAll(result, "#", "")  // headers
	result = strings.ReplaceAll(result, ">", "")  // blockquotes

	// remove links [text](url)
	result = strings.ReplaceAll(result, "[", "")
	result = strings.ReplaceAll(result, "]", "")
	result = strings.ReplaceAll(result, "(", "")
	result = strings.ReplaceAll(result, ")", "")

	// remove list markers
	result = strings.ReplaceAll(result, "- ", "")
	result = strings.ReplaceAll(result, "+ ", "")

	return result
}

func removePunctuation(s string) string {
	// remove common punctuation but preserve alphanumeric
	replacer := strings.NewReplacer(
		".", "",
		",", "",
		"!", "",
		"?", "",
		";", "",
		":", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"/", "",
		"\\", "",
		"|", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"^", "",
		"&", "",
		"=", "",
		"+", "",
		"-", "",
	)
	return replacer.Replace(s)
}

func removeQuotes(s string) string {
	replacer := strings.NewReplacer(
		"\"", "", // double quotes
		"'", "", // single quotes
		"`", "", // backticks
		"\\\"", "\"", // smart quotes
		"\\'", "'", // smart single quotes
	)
	return replacer.Replace(s)
}

func normalizeNumbers(s string) string {
	// normalize common number sequence patterns
	replacer := strings.NewReplacer(
		"1, 2, 3, 4, 5", "1,2,3,4,5",
		"1 2 3 4 5", "1,2,3,4,5",
		"1-2-3-4-5", "1,2,3,4,5",
		"1.2.3.4.5", "1,2,3,4,5",
	)
	return replacer.Replace(s)
}
