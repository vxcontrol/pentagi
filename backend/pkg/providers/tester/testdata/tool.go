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

type testCaseTool struct {
	def TestDefinition

	// state for streaming and response collection
	mu        sync.Mutex
	content   strings.Builder
	reasoning strings.Builder
	messages  []llms.MessageContent
	tools     []llms.Tool
	expected  []ExpectedToolCall
}

func newToolTestCase(def TestDefinition) (TestCase, error) {
	// parse expected tool calls
	expectedInterface, ok := def.Expected.([]any)
	if !ok {
		return nil, fmt.Errorf("tool test expected must be array of tool calls")
	}

	var expected []ExpectedToolCall
	for _, exp := range expectedInterface {
		expMap, ok := exp.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool call expected must be object")
		}

		functionName, ok := expMap["function_name"].(string)
		if !ok {
			return nil, fmt.Errorf("function_name must be string")
		}

		arguments, ok := expMap["arguments"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("arguments must be object")
		}

		expected = append(expected, ExpectedToolCall{
			FunctionName: functionName,
			Arguments:    arguments,
		})
	}

	// convert MessagesData to llms.MessageContent
	messages, err := def.Messages.ToMessageContent()
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %v", err)
	}

	// convert ToolData to llms.Tool
	var tools []llms.Tool
	for _, toolData := range def.Tools {
		tool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        toolData.Name,
				Description: toolData.Description,
				Parameters:  toolData.Parameters,
			},
		}
		tools = append(tools, tool)
	}

	return &testCaseTool{
		def:      def,
		expected: expected,
		messages: messages,
		tools:    tools,
	}, nil
}

func (t *testCaseTool) ID() string                      { return t.def.ID }
func (t *testCaseTool) Name() string                    { return t.def.Name }
func (t *testCaseTool) Type() TestType                  { return t.def.Type }
func (t *testCaseTool) Group() TestGroup                { return t.def.Group }
func (t *testCaseTool) Streaming() bool                 { return t.def.Streaming }
func (t *testCaseTool) Prompt() string                  { return "" }
func (t *testCaseTool) Messages() []llms.MessageContent { return t.messages }
func (t *testCaseTool) Tools() []llms.Tool              { return t.tools }

func (t *testCaseTool) StreamingCallback() streaming.Callback {
	if !t.def.Streaming {
		return nil
	}

	return func(ctx context.Context, chunk streaming.Chunk) error {
		t.mu.Lock()
		defer t.mu.Unlock()

		t.content.WriteString(chunk.Content)
		if !chunk.Reasoning.IsEmpty() {
			t.reasoning.WriteString(chunk.Reasoning.Content)
		}
		return nil
	}
}

func (t *testCaseTool) Execute(response any, latency time.Duration) TestResult {
	result := TestResult{
		ID:        t.def.ID,
		Name:      t.def.Name,
		Type:      t.def.Type,
		Group:     t.def.Group,
		Streaming: t.def.Streaming,
		Latency:   latency,
	}

	contentResponse, ok := response.(*llms.ContentResponse)
	if !ok {
		result.Success = false
		result.Error = fmt.Errorf("expected *llms.ContentResponse for tool test, got %T", response)
		return result
	}

	// check for reasoning content
	if t.reasoning.Len() > 0 {
		result.Reasoning = true
	}

	// extract tool calls from response
	if len(contentResponse.Choices) == 0 {
		result.Success = false
		result.Error = fmt.Errorf("no choices in response")
		return result
	}

	var toolCalls []llms.ToolCall
	for _, choice := range contentResponse.Choices {
		// check for reasoning tokens
		if reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"]; ok {
			if tokens, ok := reasoningTokens.(int); ok && tokens > 0 {
				result.Reasoning = true
			}
		}
		if !choice.Reasoning.IsEmpty() {
			result.Reasoning = true
		}

		toolCalls = append(toolCalls, choice.ToolCalls...)
	}

	// ensure at least one tool call was made
	if len(toolCalls) == 0 {
		result.Success = false
		result.Error = fmt.Errorf("no tool calls found, expected at least %d", len(t.expected))
		return result
	}

	// validate that each expected function call has a matching tool call
	if err := t.validateExpectedToolCalls(toolCalls); err != nil {
		result.Success = false
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

// validateExpectedToolCalls checks that each expected function call has at least one matching tool call
func (t *testCaseTool) validateExpectedToolCalls(toolCalls []llms.ToolCall) error {
	for _, expected := range t.expected {
		if err := t.findMatchingToolCall(toolCalls, expected); err != nil {
			return fmt.Errorf("expected function '%s' not found in tool calls: %w", expected.FunctionName, err)
		}
	}
	return nil
}

// findMatchingToolCall searches for a tool call that matches the expected function call
func (t *testCaseTool) findMatchingToolCall(toolCalls []llms.ToolCall, expected ExpectedToolCall) error {
	var lastErr error
	for _, call := range toolCalls {
		if call.FunctionCall == nil {
			return fmt.Errorf("tool call %s has no function call", call.FunctionCall.Name)
		}

		if call.FunctionCall.Name != expected.FunctionName {
			continue
		}

		// parse and validate arguments
		var args map[string]any
		if err := json.Unmarshal([]byte(call.FunctionCall.Arguments), &args); err != nil {
			return fmt.Errorf("invalid JSON in tool call %s: %v", call.FunctionCall.Name, err)
		}

		// check if all required arguments match
		if lastErr = t.validateFunctionArguments(args, expected); lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("expected function %s not found in tool calls", expected.FunctionName)
}

// validateFunctionArguments checks if all expected arguments match the actual arguments
func (t *testCaseTool) validateFunctionArguments(args map[string]any, expected ExpectedToolCall) error {
	for key, expectedVal := range expected.Arguments {
		actualVal, exists := args[key]
		if !exists {
			return fmt.Errorf("argument %s not found in tool call", key)
		}

		if err := validateArgumentValue(key, actualVal, expectedVal); err != nil {
			return err
		}
	}
	return nil
}

// validateArgumentValue performs flexible validation for function arguments using type-specific comparison
func validateArgumentValue(key string, actual, expected any) error {
	// fast path: JSON comparison first
	actualBytes, err1 := json.Marshal(actual)
	expectedBytes, err2 := json.Marshal(expected)
	if err1 == nil && err2 == nil && string(actualBytes) == string(expectedBytes) {
		return nil
	}

	var err error
	switch expected := expected.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		err = compareNumeric(actual, expected)
	case float32, float64:
		err = compareFloat(actual, expected)
	case string:
		err = compareString(actual, expected)
	case bool:
		err = compareBool(actual, expected)
	case []any:
		err = compareSlice(actual, expected)
	case map[string]any:
		err = compareMap(actual, expected)
	default:
		err = fmt.Errorf("unsupported type: %T", expected)
	}

	if err != nil {
		return fmt.Errorf("invalid argument '%s': %w", key, err)
	}

	return nil
}

func compareNumeric(actual, expected any) error {
	expectedStr := fmt.Sprintf("%v", expected)

	switch actual := actual.(type) {
	case string:
		if strings.TrimSpace(actual) != expectedStr {
			return fmt.Errorf("expected %s, got %s", expectedStr, actual)
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if fmt.Sprintf("%v", actual) != expectedStr {
			return fmt.Errorf("expected %s, got %v", expectedStr, actual)
		}
	case float32:
		if fmt.Sprintf("%d", int(actual)) != expectedStr {
			return fmt.Errorf("expected %s, got %d", expectedStr, int(actual))
		}
	case float64:
		if fmt.Sprintf("%d", int(actual)) != expectedStr {
			return fmt.Errorf("expected %s, got %d", expectedStr, int(actual))
		}
	default:
		return fmt.Errorf("unsupported type for numeric comparison: %T", actual)
	}

	return nil
}

func compareFloat(actual, expected any) error {
	expectedStr := fmt.Sprintf("%.5f", expected)

	switch actual := actual.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		actualStr := fmt.Sprintf("%v", actual)
		if !strings.HasPrefix(expectedStr, actualStr) && !strings.HasPrefix(actualStr, expectedStr) {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	case float32, float64:
		actualStr := fmt.Sprintf("%.5f", actual)
		if actualStr != expectedStr {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	case string:
		actualStr := strings.TrimSpace(actual)
		if !strings.Contains(actualStr, expectedStr) && !strings.Contains(expectedStr, actualStr) {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	default:
		return fmt.Errorf("unsupported type for float comparison: %T", actual)
	}

	return nil
}

func compareString(actual, expected any) error {
	expectedStr := strings.ToLower(expected.(string))

	switch actual := actual.(type) {
	case string:
		actualStr := strings.ToLower(strings.TrimSpace(actual))
		if actualStr == expectedStr {
			return nil
		}
		if len(expectedStr) > 10 || len(actualStr) > 10 {
			if strings.Contains(actualStr, expectedStr) || strings.Contains(expectedStr, actualStr) {
				return nil
			}
		}
		return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		actualStr := strings.ToLower(fmt.Sprintf("%v", actual))
		if actualStr != expectedStr {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	case float32, float64:
		actualStr := strings.ToLower(fmt.Sprintf("%v", actual))
		if actualStr != expectedStr {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	default:
		return fmt.Errorf("unsupported type for string comparison: %T", actual)
	}

	return nil
}

func compareBool(actual, expected any) error {
	expectedBool := expected.(bool)

	switch actual := actual.(type) {
	case bool:
		if actual != expectedBool {
			return fmt.Errorf("expected %t, got %t", expectedBool, actual)
		}
	case string:
		actualStr := strings.Trim(strings.ToLower(actual), "' \"\n\r\t")
		expectedStr := fmt.Sprintf("%t", expectedBool)
		if actualStr != expectedStr {
			return fmt.Errorf("expected %s, got %s", expectedStr, actualStr)
		}
	default:
		return fmt.Errorf("unsupported type for bool comparison: %T", actual)
	}

	return nil
}

func compareSlice(actual any, expected []any) error {
	switch actual := actual.(type) {
	case []any:
		// each element in expected must match at least one element in actual
		for _, exp := range expected {
			found := false
			var lastErr error
			for _, act := range actual {
				if lastErr = validateArgumentValue("", act, exp); lastErr == nil {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("expected %v, got %v: %w", expected, actual, lastErr)
			}
		}
		return nil
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// actual simple type must match at least one element in expected
		var lastErr error
		for _, exp := range expected {
			if lastErr = validateArgumentValue("", actual, exp); lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("expected %v, got %v: %w", expected, actual, lastErr)
	default:
		return fmt.Errorf("unsupported type for slice comparison: %T", actual)
	}
}

func compareMap(actual, expected any) error {
	var lastErr error
	if actualSlice, ok := actual.([]any); ok {
		for _, actualMap := range actualSlice {
			if lastErr = compareMap(actualMap, expected); lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("expected %v, got %v: %w", expected, actual, lastErr)
	}

	if actualSlice, ok := actual.([]map[string]any); ok {
		for _, actualMap := range actualSlice {
			if lastErr = compareMap(actualMap, expected); lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("expected %v, got %v: %w", expected, actual, lastErr)
	}

	actualMap, ok := actual.(map[string]any)
	if !ok {
		return fmt.Errorf("expected map, got %T", actual)
	}

	expectedMap := expected.(map[string]any)

	// exact key match required
	for key, expectedVal := range expectedMap {
		actualVal, exists := actualMap[key]
		if !exists {
			return fmt.Errorf("expected key %s not found in actual map", key)
		}
		if err := validateArgumentValue(key, actualVal, expectedVal); err != nil {
			return err
		}
	}

	return nil
}
