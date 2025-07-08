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

	// convert MessageData to llms.MessageContent if messages exist
	var messages []llms.MessageContent
	for _, msg := range def.Messages {
		var msgType llms.ChatMessageType
		switch strings.ToLower(msg.Role) {
		case "system":
			msgType = llms.ChatMessageTypeSystem
		case "user", "human":
			msgType = llms.ChatMessageTypeHuman
		case "assistant", "ai":
			msgType = llms.ChatMessageTypeAI
		default:
			return nil, fmt.Errorf("unknown message role: %s", msg.Role)
		}
		messages = append(messages, llms.TextParts(msgType, msg.Content))
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

func (t *testCaseCompletion) Execute(response interface{}, latency time.Duration) TestResult {
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

// containsString implements enhanced string matching logic from original ctester.
func containsString(response, expected string) bool {
	if len(response) == 0 {
		return false
	}

	// direct equality check
	if response == expected {
		return true
	}

	// special case for number sequences - normalize by removing spaces
	if expected == "1, 2, 3, 4, 5" || expected == "1,2,3,4,5" {
		normalizedResponse := strings.ReplaceAll(response, " ", "")
		normalizedExpected := "1,2,3,4,5"

		if strings.Contains(normalizedResponse, normalizedExpected) {
			return true
		}

		// also try with spaces
		if strings.Contains(response, "1, 2, 3, 4, 5") {
			return true
		}
	}

	// check if response contains expected text
	if strings.Contains(response, expected) {
		return true
	}

	// check if expected contains response (reverse check)
	if strings.Contains(expected, response) {
		return true
	}

	// case-insensitive matching
	if strings.Contains(strings.ToLower(response), strings.ToLower(expected)) {
		return true
	}

	return false
}
