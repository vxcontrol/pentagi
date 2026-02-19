package testdata

import (
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"

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
				Content: "",
				Reasoning: &reasoning.ContentReasoning{
					Content: "Let me think about this...",
				},
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

func TestValidateArgumentValue(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		// JSON fast path
		{"json_exact_match", 42, 42, true},
		{"json_string_match", "test", "test", true},

		// Numeric tests
		{"int_string_match", "42", 42, true},
		{"int_int_match", 42, 42, true},
		{"float_to_int", 42.7, 42, true},
		{"float_to_int_wrong", 43.7, 42, false},
		{"int_invalid_type", []int{1}, 42, false},

		// Float tests
		{"float_exact", 3.14159, 3.14159, true},
		{"float_precision", 3.141592653, 3.14159, true},
		{"int_to_float_prefix", 3, 3.14159, true},
		{"string_to_float_prefix", "3.14", 3.14159, true},
		{"float_invalid_type", map[string]int{}, 3.14, false},

		// String tests
		{"string_exact", "Hello", "hello", true},
		{"string_trimspace", " Hello ", "hello", true},
		{"string_long_contains", "This is a long test message", "test message", true},
		{"string_long_reverse", "test", "This is a long test message", true},
		{"string_short_nomatch", "hello", "world", false},
		{"int_to_string", 42, "42", true},
		{"float_to_string", 3.14, "3.14", true},

		// Boolean tests
		{"bool_exact", true, true, true},
		{"bool_false", false, false, true},
		{"string_true", "true", true, true},
		{"string_false", "false", false, true},
		{"string_quoted", "'true'", true, true},
		{"bool_invalid_type", 1, true, false},

		// Slice tests
		{"slice_to_slice_match", []any{1, 2, 3}, []any{1, 2}, true},
		{"slice_to_slice_nomatch", []any{1, 2}, []any{1, 2, 3}, false},
		{"simple_to_slice_match", "hello", []any{"hello", "world"}, true},
		{"simple_to_slice_nomatch", "test", []any{"hello", "world"}, false},
		{"slice_invalid_type", map[string]int{}, []any{1, 2}, false},
		{"slice_to_slice_map_match", []map[string]any{{"key": "value"}, {"key": "value2"}},
			[]map[string]any{{"key": "value"}, {"key": "value2"}}, true},
		{"slice_to_slice_map_nomatch", []map[string]any{{"key": "value"}, {"key": "value2"}},
			[]map[string]any{{"key": "value"}, {"key": "value2"}, {"key": "value3"}}, false},

		// Map tests
		{"map_exact_match", map[string]any{"key": "value"}, map[string]any{"key": "value"}, true},
		{"map_missing_key", map[string]any{}, map[string]any{"key": "value"}, false},
		{"map_wrong_value", map[string]any{"key": "wrong"}, map[string]any{"key": "value"}, false},
		{"map_nested", map[string]any{"key": map[string]any{"nested": "value"}},
			map[string]any{"key": map[string]any{"nested": "value"}}, true},
		{"map_invalid_type", "not_a_map", map[string]any{"key": "value"}, false},
		{"map_slice_match_value", []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "value"}, true},
		{"map_slice_nomatch_value", []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "wrong"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgumentValue("", tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("validateArgumentValue(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"string_match", "42", 42, true},
		{"string_nomatch", "43", 42, false},
		{"string_spaces", " 42 ", 42, true},
		{"int_match", 42, 42, true},
		{"uint_match", uint(42), 42, true},
		{"float_truncate", 42.7, 42, true},
		{"float_truncate_fail", 43.7, 42, false},
		{"invalid_type", []int{}, 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareNumeric(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareNumeric(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareFloat(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"exact_match", 3.14159, 3.14159, true},
		{"precision_match", 3.141592653, 3.14159, true},
		{"int_prefix", 3, 3.14159, true},
		{"string_prefix", "3.14", 3.14159, true},
		{"string_contains", "value: 3.14000 found", 3.14, true},
		{"no_prefix", 4, 3.14159, false},
		{"invalid_type", []int{}, 3.14, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareFloat(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareFloat(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareString(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"exact_match", "Hello", "hello", true},
		{"spaces_trimmed", " Hello ", "hello", true},
		{"long_contains", "This is a very long test message", "test message", true},
		{"long_reverse", "test", "This is a very long test message", true},
		{"short_nomatch", "hello", "world", false},
		{"int_match", 42, "42", true},
		{"float_match", 3.14, "3.14", true},
		{"invalid_type", []int{}, "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareString(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareString(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareBool(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"true_match", true, true, true},
		{"false_match", false, false, true},
		{"true_nomatch", true, false, false},
		{"string_true", "true", true, true},
		{"string_false", "false", false, true},
		{"string_quoted", "'true'", true, true},
		{"string_spaced", " true ", true, true},
		{"string_wrong", "yes", true, false},
		{"invalid_type", 1, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareBool(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareBool(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareSlice(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected []any
		want     bool
	}{
		{"slice_all_match", []any{1, 2, 3}, []any{1, 2}, true},
		{"slice_partial_nomatch", []any{1, 2}, []any{1, 2, 3}, false},
		{"slice_empty_expected", []any{1, 2, 3}, []any{}, true},
		{"simple_in_slice", "hello", []any{"hello", "world"}, true},
		{"simple_not_in_slice", "test", []any{"hello", "world"}, false},
		{"int_in_slice", 42, []any{41, 42, 43}, true},
		{"invalid_type", map[string]int{}, []any{1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareSlice(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareSlice(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestCompareMap(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"exact_match", map[string]any{"key": "value"}, map[string]any{"key": "value"}, true},
		{"missing_key", map[string]any{}, map[string]any{"key": "value"}, false},
		{"wrong_value", map[string]any{"key": "wrong"}, map[string]any{"key": "value"}, false},
		{"extra_keys_ok", map[string]any{"key": "value", "extra": "ok"}, map[string]any{"key": "value"}, true},
		{"nested_match", map[string]any{"key": map[string]any{"nested": "value"}},
			map[string]any{"key": map[string]any{"nested": "value"}}, true},
		{"not_a_map", "string", map[string]any{"key": "value"}, false},
		{"map_slice_match_value", []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "value"}, true},
		{"map_slice_nomatch_value", []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "wrong"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareMap(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compareMap(%v, %v) = %v, want %v, error: %v",
					tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

// Test enhanced tool call validation
func TestToolCallEnhancedValidation(t *testing.T) {
	// Test case with order-independent function calls
	t.Run("order_independent_calls", func(t *testing.T) {
		def := TestDefinition{
			ID:   "test_order",
			Type: TestTypeTool,
			Tools: []ToolData{
				{
					Name:        "search",
					Description: "Search function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{"type": "string"},
						},
						"required": []string{"query"},
					},
				},
				{
					Name:        "echo",
					Description: "Echo function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
						"required": []string{"message"},
					},
				},
			},
			Expected: []any{
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "hello"},
				},
				map[string]any{
					"function_name": "search",
					"arguments":     map[string]any{"query": "test"},
				},
			},
		}

		testCase, err := newToolTestCase(def)
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}

		// Create response with functions in different order
		response := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					ToolCalls: []llms.ToolCall{
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "search",
								Arguments: `{"query": "test"}`,
							},
						},
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
			t.Errorf("Expected success for order-independent calls, got failure: %v", result.Error)
		}
	})

	// Test case with extra function calls from LLM
	t.Run("extra_function_calls", func(t *testing.T) {
		def := TestDefinition{
			ID:   "test_extra",
			Type: TestTypeTool,
			Tools: []ToolData{
				{
					Name:        "echo",
					Description: "Echo function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
						"required": []string{"message"},
					},
				},
			},
			Expected: []any{
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "hello"},
				},
			},
		}

		testCase, err := newToolTestCase(def)
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}

		// Create response with extra function calls
		response := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					ToolCalls: []llms.ToolCall{
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "echo",
								Arguments: `{"message": "hello"}`,
							},
						},
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "search",
								Arguments: `{"query": "additional"}`,
							},
						},
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "echo",
								Arguments: `{"message": "extra"}`,
							},
						},
					},
				},
			},
		}

		result := testCase.Execute(response, time.Millisecond*100)
		if !result.Success {
			t.Errorf("Expected success with extra function calls, got failure: %v", result.Error)
		}
	})

	// Test case with missing expected function call
	t.Run("missing_expected_call", func(t *testing.T) {
		def := TestDefinition{
			ID:   "test_missing",
			Type: TestTypeTool,
			Tools: []ToolData{
				{
					Name:        "echo",
					Description: "Echo function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
						"required": []string{"message"},
					},
				},
			},
			Expected: []any{
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "hello"},
				},
				map[string]any{
					"function_name": "search",
					"arguments":     map[string]any{"query": "test"},
				},
			},
		}

		testCase, err := newToolTestCase(def)
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}

		// Create response with only one function call (missing search)
		response := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
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
		if result.Success {
			t.Errorf("Expected failure for missing expected function call, got success")
		}
		if !strings.Contains(result.Error.Error(), "search") {
			t.Errorf("Expected error about missing 'search' function, got: %v", result.Error)
		}
	})

	// Test case with no function calls at all
	t.Run("no_function_calls", func(t *testing.T) {
		def := TestDefinition{
			ID:   "test_none",
			Type: TestTypeTool,
			Tools: []ToolData{
				{
					Name:        "echo",
					Description: "Echo function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
						"required": []string{"message"},
					},
				},
			},
			Expected: []any{
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "hello"},
				},
			},
		}

		testCase, err := newToolTestCase(def)
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}

		// Create response with no function calls
		response := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content:   "I'll help you with that.",
					ToolCalls: []llms.ToolCall{},
				},
			},
		}

		result := testCase.Execute(response, time.Millisecond*100)
		if result.Success {
			t.Errorf("Expected failure for no function calls, got success")
		}
		if !strings.Contains(result.Error.Error(), "no tool calls found") {
			t.Errorf("Expected error about no tool calls, got: %v", result.Error)
		}
	})

	// Test case with function calls across multiple choices
	t.Run("multiple_choices", func(t *testing.T) {
		def := TestDefinition{
			ID:   "test_choices",
			Type: TestTypeTool,
			Tools: []ToolData{
				{
					Name:        "echo",
					Description: "Echo function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
						"required": []string{"message"},
					},
				},
			},
			Expected: []any{
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "hello"},
				},
				map[string]any{
					"function_name": "echo",
					"arguments":     map[string]any{"message": "world"},
				},
			},
		}

		testCase, err := newToolTestCase(def)
		if err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}

		// Create response with function calls distributed across choices
		response := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					ToolCalls: []llms.ToolCall{
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "echo",
								Arguments: `{"message": "hello"}`,
							},
						},
					},
				},
				{
					ToolCalls: []llms.ToolCall{
						{
							FunctionCall: &llms.FunctionCall{
								Name:      "echo",
								Arguments: `{"message": "world"}`,
							},
						},
					},
				},
			},
		}

		result := testCase.Execute(response, time.Millisecond*100)
		if !result.Success {
			t.Errorf("Expected success with multiple choices, got failure: %v", result.Error)
		}
	})
}
