package testdata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type testCaseJSON struct {
	def TestDefinition

	// state for streaming and response collection
	mu        sync.Mutex
	content   strings.Builder
	reasoning strings.Builder
	messages  []llms.MessageContent
	expected  map[string]any
}

func newJSONTestCase(def TestDefinition) (TestCase, error) {
	// for array tests, expected can be empty or nil
	var expected map[string]any
	if def.Expected != nil {
		exp, ok := def.Expected.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("JSON test expected must be map[string]any")
		}
		expected = exp
	}

	// convert MessagesData to llms.MessageContent
	messages, err := def.Messages.ToMessageContent()
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %v", err)
	}

	return &testCaseJSON{
		def:      def,
		expected: expected,
		messages: messages,
	}, nil
}

func (t *testCaseJSON) ID() string                      { return t.def.ID }
func (t *testCaseJSON) Name() string                    { return t.def.Name }
func (t *testCaseJSON) Type() TestType                  { return t.def.Type }
func (t *testCaseJSON) Group() TestGroup                { return t.def.Group }
func (t *testCaseJSON) Streaming() bool                 { return t.def.Streaming }
func (t *testCaseJSON) Prompt() string                  { return "" }
func (t *testCaseJSON) Messages() []llms.MessageContent { return t.messages }
func (t *testCaseJSON) Tools() []llms.Tool              { return nil }

func (t *testCaseJSON) StreamingCallback() streaming.Callback {
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

func (t *testCaseJSON) Execute(response any, latency time.Duration) TestResult {
	result := TestResult{
		ID:        t.def.ID,
		Name:      t.def.Name,
		Type:      t.def.Type,
		Group:     t.def.Group,
		Streaming: t.def.Streaming,
		Latency:   latency,
	}

	// handle different response types
	var jsonContent string
	switch resp := response.(type) {
	case string:
		jsonContent = resp
	case *llms.ContentResponse:
		if len(resp.Choices) == 0 {
			result.Success = false
			result.Error = fmt.Errorf("no choices in response")
			return result
		}

		// check for reasoning content
		choice := resp.Choices[0]
		if len(choice.ReasoningContent) > 0 {
			result.Reasoning = true
		}
		if reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"]; ok {
			if tokens, ok := reasoningTokens.(int); ok && tokens > 0 {
				result.Reasoning = true
			}
		}

		jsonContent = choice.Content
	default:
		result.Success = false
		result.Error = fmt.Errorf("unexpected response type for JSON test: %T", response)
		return result
	}

	// extract JSON from response (handle code blocks and extra text)
	jsonContent = extractJSON(jsonContent)
	jsonBytes := []byte(jsonContent)

	// parse JSON object
	var parsed any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("invalid JSON response: %v", err)
		return result
	}

	// validate expected values
	if err := validateArgumentValue("", parsed, t.expected); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("got %#v, expected %#v: %w", parsed, t.expected, err)
		return result
	}

	result.Success = true
	return result
}

// extractJSON extracts JSON content from text that may contain code blocks or extra text
func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	// first, try to find JSON in code blocks
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json")
		if start != -1 {
			start += 7 // len("```json")
			end := strings.Index(content[start:], "```")
			if end != -1 {
				return strings.TrimSpace(content[start : start+end])
			}
		}
	}

	// try generic code blocks
	if strings.Contains(content, "```") {
		start := strings.Index(content, "```")
		if start != -1 {
			start += 3
			end := strings.Index(content[start:], "```")
			if end != -1 {
				candidate := strings.TrimSpace(content[start : start+end])
				// check if it looks like JSON
				if strings.HasPrefix(candidate, "{") || strings.HasPrefix(candidate, "[") {
					return candidate
				}
			}
		}
	}

	// try to parse as a valid JSON and return one
	var raw any
	if err := json.Unmarshal([]byte(content), &raw); err == nil {
		return content
	}

	// try to find JSON array boundaries first (higher priority)
	if strings.Contains(content, "[") {
		start := strings.Index(content, "[")
		end := strings.LastIndex(content, "]")
		if start != -1 && end != -1 && end > start {
			return strings.TrimSpace(content[start : end+1])
		}
	}

	// try to find JSON object boundaries
	if strings.Contains(content, "{") {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start != -1 && end != -1 && end > start {
			return strings.TrimSpace(content[start : end+1])
		}
	}

	// return as-is if no extraction patterns match
	return content
}
