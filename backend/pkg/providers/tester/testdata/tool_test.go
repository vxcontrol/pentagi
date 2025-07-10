package testdata

import (
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func TestToolTestCase(t *testing.T) {
	testYAML := `
- id: "test_echo"
  name: "Echo Function Test"
  type: "tool"
  group: "basic"
  messages:
    - role: "system"
      content: "Use tools only"
    - role: "user"
      content: "Call echo with message hello"
  tools:
    - name: "echo"
      description: "Echoes back the input"
      parameters:
        type: "object"
        properties:
          message:
            type: "string"
            description: "Message to echo"
        required: ["message"]
  expected:
    - function_name: "echo"
      arguments:
        message: "hello"
  streaming: false
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(definitions))
	}

	// test tool case
	toolDef := definitions[0]
	testCase, err := newToolTestCase(toolDef)
	if err != nil {
		t.Fatalf("Failed to create tool test case: %v", err)
	}

	if testCase.ID() != "test_echo" {
		t.Errorf("Expected ID 'test_echo', got %s", testCase.ID())
	}
	if testCase.Type() != TestTypeTool {
		t.Errorf("Expected type tool, got %s", testCase.Type())
	}
	if len(testCase.Messages()) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(testCase.Messages()))
	}
	if len(testCase.Tools()) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(testCase.Tools()))
	}

	// test execution with correct function call
	response := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"message": "hello"}`,
						},
					},
				},
			},
		},
	}
	result := testCase.Execute(response, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for correct function call, got failure: %v", result.Error)
	}
	if result.Latency != time.Millisecond*100 {
		t.Errorf("Expected latency 100ms, got %v", result.Latency)
	}

	// test execution with wrong function name
	response = &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "wrong_function",
							Arguments: `{"message": "hello"}`,
						},
					},
				},
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for wrong function name, got success")
	}

	// test execution with wrong arguments
	response = &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"wrong_arg": "hello"}`,
						},
					},
				},
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for wrong arguments, got success")
	}

	// test execution with no tool calls
	response = &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:   "I cannot call functions",
				ToolCalls: nil,
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for no tool calls, got success")
	}

	// test execution with reasoning content
	response = &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:          "",
				ReasoningContent: "Let me think about this...",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"message": "hello"}`,
						},
					},
				},
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for function call with reasoning, got failure: %v", result.Error)
	}
	if !result.Reasoning {
		t.Errorf("Expected reasoning to be detected, got false")
	}
}

func TestToolTestCaseMultipleFunctions(t *testing.T) {
	testYAML := `
- id: "test_multiple"
  name: "Multiple Function Test"
  type: "tool"
  group: "advanced"
  messages:
    - role: "user"
      content: "Call both functions"
  tools:
    - name: "echo"
      description: "Echoes back the input"
      parameters:
        type: "object"
        properties:
          message:
            type: "string"
        required: ["message"]
    - name: "count"
      description: "Counts to a number"
      parameters:
        type: "object"
        properties:
          number:
            type: "integer"
        required: ["number"]
  expected:
    - function_name: "echo"
      arguments:
        message: "test"
    - function_name: "count"
      arguments:
        number: 5
  streaming: false
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	testCase, err := newToolTestCase(definitions[0])
	if err != nil {
		t.Fatalf("Failed to create tool test case: %v", err)
	}

	// test execution with correct multiple function calls
	response := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"message": "test"}`,
						},
					},
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "count",
							Arguments: `{"number": 5}`,
						},
					},
				},
			},
		},
	}
	result := testCase.Execute(response, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for multiple function calls, got failure: %v", result.Error)
	}

	// test execution with wrong number of function calls
	response = &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "",
				ToolCalls: []llms.ToolCall{
					{
						FunctionCall: &llms.FunctionCall{
							Name:      "echo",
							Arguments: `{"message": "test"}`,
						},
					},
				},
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for wrong number of function calls, got success")
	}
}
