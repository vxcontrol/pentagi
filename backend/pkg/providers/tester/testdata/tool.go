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
	expectedInterface, ok := def.Expected.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tool test expected must be array of tool calls")
	}

	var expected []ExpectedToolCall
	for _, exp := range expectedInterface {
		expMap, ok := exp.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tool call expected must be object")
		}

		functionName, ok := expMap["function_name"].(string)
		if !ok {
			return nil, fmt.Errorf("function_name must be string")
		}

		arguments, ok := expMap["arguments"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("arguments must be object")
		}

		expected = append(expected, ExpectedToolCall{
			FunctionName: functionName,
			Arguments:    arguments,
		})
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
		t.reasoning.WriteString(chunk.ReasoningContent)
		return nil
	}
}

func (t *testCaseTool) Execute(response interface{}, latency time.Duration) TestResult {
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
		if len(choice.ReasoningContent) > 0 {
			result.Reasoning = true
		}

		toolCalls = append(toolCalls, choice.ToolCalls...)
	}

	if len(toolCalls) != len(t.expected) {
		result.Success = false
		result.Error = fmt.Errorf("expected %d tool calls, got %d", len(t.expected), len(toolCalls))
		return result
	}

	// validate each tool call
	for i, expected := range t.expected {
		if i >= len(toolCalls) {
			result.Success = false
			result.Error = fmt.Errorf("missing tool call %d", i)
			return result
		}

		call := toolCalls[i]
		if call.FunctionCall == nil {
			result.Success = false
			result.Error = fmt.Errorf("tool call %d has no function call", i)
			return result
		}

		if call.FunctionCall.Name != expected.FunctionName {
			result.Success = false
			result.Error = fmt.Errorf("expected function %s, got %s", expected.FunctionName, call.FunctionCall.Name)
			return result
		}

		// parse and validate arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(call.FunctionCall.Arguments), &args); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("invalid function arguments: %v", err)
			return result
		}

		// check required arguments
		for key, expectedVal := range expected.Arguments {
			actualVal, exists := args[key]
			if !exists {
				result.Success = false
				result.Error = fmt.Errorf("function %s missing argument: %s", expected.FunctionName, key)
				return result
			}

			if !validateArgumentValue(key, actualVal, expectedVal, expected.FunctionName) {
				result.Success = false
				result.Error = fmt.Errorf("function %s arg %s: expected %v, got %v", expected.FunctionName, key, expectedVal, actualVal)
				return result
			}
		}
	}

	result.Success = true
	return result
}

// validateArgumentValue performs flexible validation for function arguments using JSON comparison
func validateArgumentValue(key string, actual, expected interface{}, functionName string) bool {
	// convert both values to JSON and compare
	actualBytes, err1 := json.Marshal(actual)
	expectedBytes, err2 := json.Marshal(expected)

	if err1 != nil || err2 != nil {
		// if JSON marshaling fails, fall back to direct comparison
		return jsonValuesEqual(actual, expected)
	}

	// direct JSON comparison
	if string(actualBytes) == string(expectedBytes) {
		return true
	}

	// for strings, allow more flexible comparison
	actualStr, actualIsStr := actual.(string)
	expectedStr, expectedIsStr := expected.(string)

	if actualIsStr && expectedIsStr {
		actualStr = strings.TrimSpace(actualStr)
		expectedStr = strings.TrimSpace(expectedStr)

		// case-insensitive comparison
		if strings.EqualFold(actualStr, expectedStr) {
			return true
		}

		// allow partial matches for longer strings (more than 10 chars)
		if len(expectedStr) > 10 {
			return strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedStr)) ||
				strings.Contains(strings.ToLower(expectedStr), strings.ToLower(actualStr))
		}
	}

	return false
}
