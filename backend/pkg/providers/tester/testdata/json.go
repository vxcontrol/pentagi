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
	mu            sync.Mutex
	content       strings.Builder
	reasoning     strings.Builder
	messages      []llms.MessageContent
	expected      map[string]interface{}
	isArray       bool
	expectFailure bool
}

func newJSONTestCase(def TestDefinition) (TestCase, error) {
	// for array tests, expected can be empty or nil
	var expected map[string]interface{}
	if !def.IsArray && def.Expected != nil {
		exp, ok := def.Expected.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("JSON test expected must be map[string]interface{}")
		}
		expected = exp
	}

	// convert MessageData to llms.MessageContent
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

	return &testCaseJSON{
		def:           def,
		expected:      expected,
		messages:      messages,
		isArray:       def.IsArray,
		expectFailure: def.ExpectFailure,
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

func (t *testCaseJSON) Execute(response interface{}, latency time.Duration) TestResult {
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

	// check for array tests
	if t.isArray {
		var parsed []interface{}
		if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
			if t.expectFailure {
				result.Success = true
				return result
			}
			result.Success = false
			result.Error = fmt.Errorf("invalid JSON array response: %v", err)
			return result
		}
		result.Success = true
		return result
	}

	// parse JSON object
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		if t.expectFailure {
			result.Success = true
			return result
		}
		result.Success = false
		result.Error = fmt.Errorf("invalid JSON response: %v", err)
		return result
	}

	// for negative tests, if we got valid JSON when expecting failure
	if t.expectFailure {
		result.Success = false
		result.Error = fmt.Errorf("expected JSON parsing to fail, but got valid JSON")
		return result
	}

	// validate expected values
	for key, expectedVal := range t.expected {
		actualVal, exists := parsed[key]
		if !exists {
			result.Success = false
			result.Error = fmt.Errorf("missing required key: %s", key)
			return result
		}

		if !jsonValuesEqual(actualVal, expectedVal) {
			result.Success = false
			result.Error = fmt.Errorf("key %s: expected %v, got %v", key, expectedVal, actualVal)
			return result
		}
	}

	result.Success = true
	return result
}

// jsonValuesEqual compares JSON values with type conversion
func jsonValuesEqual(actual, expected interface{}) bool {
	// handle numeric type conversions (JSON unmarshaling can produce float64)
	switch exp := expected.(type) {
	case int:
		if act, ok := actual.(float64); ok {
			return act == float64(exp)
		}
		if act, ok := actual.(int); ok {
			return act == exp
		}
	case float64:
		if act, ok := actual.(float64); ok {
			return act == exp
		}
		if act, ok := actual.(int); ok {
			return float64(act) == exp
		}
	default:
		return actual == expected
	}
	return actual == expected
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
