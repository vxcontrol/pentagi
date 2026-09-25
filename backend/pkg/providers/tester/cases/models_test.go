package cases

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestModels_ToMessageContent_ReplaysThinkingToolCallsAndTheirResponses(t *testing.T) {
	registry, err := LoadRegistryFromYAML([]byte(`
- id: "memory_test_completion"
  name: "Memory Test with Extended Messages"
  type: "completion"
  group: "advanced"
  messages:
    - role: "system"
      content: "You are helpful"
    - role: "user"
      content: "Remember my name is Alice"
    - role: "assistant"
      content: "I'll remember that your name is Alice"
      reasoning: "counted on my fingers"
      reasoning_signature: "sig-from-an-earlier-turn"
    - role: "user"
      content: "What is my name?"
  expected: "Alice"
  streaming: false

- id: "memory_test_tool"
  name: "Memory Test with Tool Calls"
  type: "tool"
  group: "advanced"
  messages:
    - role: "system"
      content: "You are a helpful assistant"
    - role: "user"
      content: "Get weather for London"
    - role: "assistant"
      content: "I'll get the weather for London"
      tool_calls:
        - id: "call_1"
          type: "function"
          function:
            name: "get_weather"
            arguments:
              location: "London"
    - role: "tool"
      tool_call_id: "call_1"
      name: "get_weather"
      content: "Weather in London is cloudy, 15°C"
    - role: "user"
      content: "Now get weather for Paris"
  tools:
    - name: "get_weather"
      description: "Gets current weather for a location"
      parameters:
        type: "object"
        properties:
          location:
            type: "string"
            description: "City name"
        required: ["location"]
  expected:
    - function_name: "get_weather"
      arguments:
        location: "Paris"
  streaming: false
`))
	if err != nil {
		t.Fatalf("Failed to load registry from YAML: %v", err)
	}

	completionTests := registry.GetTestsByType(TestTypeCompletion)
	toolTests := registry.GetTestsByType(TestTypeTool)
	if len(completionTests) != 1 || len(toolTests) != 1 {
		t.Fatalf("want one completion and one tool definition, got %d and %d", len(completionTests), len(toolTests))
	}

	completionCase, err := registry.createTestCase(completionTests[0])
	if err != nil {
		t.Fatalf("Failed to create completion test case: %v", err)
	}
	if completionCase.Type() != TestTypeCompletion {
		t.Errorf("Expected completion test type, got %s", completionCase.Type())
	}
	messages := completionCase.Messages()
	if len(messages) != 4 {
		t.Fatalf("Expected 4 messages, got %d", len(messages))
	}
	var (
		said     string
		thinking *reasoning.ContentReasoning
	)
	for _, part := range messages[2].Parts {
		if text, ok := part.(llms.TextContent); ok {
			said, thinking = text.Text, text.Reasoning
		}
	}
	if thinking == nil || thinking.Content != "counted on my fingers" || string(thinking.Signature) != "sig-from-an-earlier-turn" {
		t.Errorf("replayed thinking = %+v, want the text and signature from the case: a vendor rejects a turn without its signature", thinking)
	}
	if said != "I'll remember that your name is Alice" {
		t.Errorf("the assistant's own text = %q, want it replayed beside its thinking", said)
	}

	toolCase, err := registry.createTestCase(toolTests[0])
	if err != nil {
		t.Fatalf("Failed to create tool test case: %v", err)
	}
	if toolCase.Type() != TestTypeTool {
		t.Errorf("Expected tool test type, got %s", toolCase.Type())
	}

	toolMessages := toolCase.Messages()
	if len(toolMessages) != 5 {
		t.Fatalf("Expected 5 messages, got %d", len(toolMessages))
	}

	var toolCallPart *llms.ToolCall
	for _, part := range toolMessages[2].Parts {
		if tc, ok := part.(llms.ToolCall); ok {
			toolCallPart = &tc
			break
		}
	}
	if toolCallPart == nil {
		t.Error("Expected tool call in assistant message parts")
	} else {
		if toolCallPart.ID != "call_1" {
			t.Errorf("Expected tool call ID 'call_1', got %s", toolCallPart.ID)
		}
		if toolCallPart.FunctionCall.Name != "get_weather" {
			t.Errorf("Expected function name 'get_weather', got %s", toolCallPart.FunctionCall.Name)
		}
		if toolCallPart.FunctionCall.Arguments != `{"location":"London"}` {
			t.Errorf("Unexpected function call arguments, got %s", toolCallPart.FunctionCall.Arguments)
		}
	}

	var toolResponsePart *llms.ToolCallResponse
	for _, part := range toolMessages[3].Parts {
		if tr, ok := part.(llms.ToolCallResponse); ok {
			toolResponsePart = &tr
			break
		}
	}
	if toolResponsePart == nil {
		t.Error("Expected tool response in tool message parts")
	} else {
		if toolResponsePart.ToolCallID != "call_1" {
			t.Errorf("Expected tool call ID 'call_1', got %s", toolResponsePart.ToolCallID)
		}
		if toolResponsePart.Name != "get_weather" {
			t.Errorf("Expected tool name 'get_weather', got %s", toolResponsePart.Name)
		}
		if toolResponsePart.Content != "Weather in London is cloudy, 15°C" {
			t.Errorf("Unexpected tool response content, got %s", toolResponsePart.Content)
		}
	}
}
