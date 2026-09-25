package cases

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestTool_NewToolTestCase_BuildsTheCaseFromItsDefinition(t *testing.T) {
	definitions := testerDefinitions(t, `
- id: "test_echo"
  name: "Echo Function Test"
  type: "tool"
  group: "basic"
  capability: "adaptive_thinking"
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
`)

	testCase, err := newToolTestCase(definitions[0])
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
	if got := testCase.Capability(); got != CapabilityAdaptiveThinking {
		t.Errorf("Capability() = %q, want %q: the gate would run the case against an agent that never asks for it",
			got, CapabilityAdaptiveThinking)
	}
}

func TestTool_Execute_PassesOnlyWhenEveryExpectedCallWasMade(t *testing.T) {
	expect := func(name string, args map[string]any) any {
		return map[string]any{"function_name": name, "arguments": args}
	}
	call := func(name, args string) llms.ToolCall {
		return llms.ToolCall{FunctionCall: &llms.FunctionCall{Name: name, Arguments: args}}
	}
	echoHello := expect("echo", map[string]any{"message": "hello"})

	for _, tc := range []struct {
		name          string
		expected      []any
		choices       [][]llms.ToolCall
		thinking      string
		wantPass      bool
		wantReasoning bool
		wantSays      string
	}{
		{
			name:     "the expected call",
			expected: []any{echoHello},
			choices:  [][]llms.ToolCall{{call("echo", `{"message": "hello"}`)}},
			wantPass: true,
		},
		{
			name:          "the expected call made after thinking",
			expected:      []any{echoHello},
			choices:       [][]llms.ToolCall{{call("echo", `{"message": "hello"}`)}},
			thinking:      "Let me think about this...",
			wantPass:      true,
			wantReasoning: true,
		},
		{
			name:     "a call to another function",
			expected: []any{echoHello},
			choices:  [][]llms.ToolCall{{call("wrong_function", `{"message": "hello"}`)}},
			wantSays: "expected function echo not found",
		},
		{
			name:     "the expected function with another argument",
			expected: []any{echoHello},
			choices:  [][]llms.ToolCall{{call("echo", `{"wrong_arg": "hello"}`)}},
			wantSays: "argument message not found",
		},
		{
			name:     "two expected calls made in the other order",
			expected: []any{echoHello, expect("search", map[string]any{"query": "test"})},
			choices:  [][]llms.ToolCall{{call("search", `{"query": "test"}`), call("echo", `{"message": "hello"}`)}},
			wantPass: true,
		},
		{
			name:     "extra calls beside the expected one",
			expected: []any{echoHello},
			choices: [][]llms.ToolCall{{
				call("echo", `{"message": "hello"}`),
				call("search", `{"query": "additional"}`),
				call("echo", `{"message": "extra"}`),
			}},
			wantPass: true,
		},
		{
			name:     "one of two expected calls missing",
			expected: []any{echoHello, expect("search", map[string]any{"query": "test"})},
			choices:  [][]llms.ToolCall{{call("echo", `{"message": "hello"}`)}},
			wantSays: "search",
		},
		{
			name:     "expected calls spread over two choices",
			expected: []any{echoHello, expect("echo", map[string]any{"message": "world"})},
			choices:  [][]llms.ToolCall{{call("echo", `{"message": "hello"}`)}, {call("echo", `{"message": "world"}`)}},
			wantPass: true,
		},
		{
			name:     "a tool call carrying no function",
			expected: []any{expect("noop", map[string]any{})},
			choices:  [][]llms.ToolCall{{{ID: "call-1", Type: "function"}}},
			wantSays: "call-1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testCase, err := newToolTestCase(TestDefinition{ID: "c", Type: TestTypeTool, Expected: tc.expected})
			if err != nil {
				t.Fatalf("Failed to create test case: %v", err)
			}

			response := &llms.ContentResponse{}
			for _, calls := range tc.choices {
				response.Choices = append(response.Choices, &llms.ContentChoice{ToolCalls: calls})
			}
			if tc.thinking != "" {
				response.Choices[0].Reasoning = &reasoning.ContentReasoning{Content: tc.thinking}
			}

			result := testCase.Execute(response, time.Millisecond*100)
			if result.Success != tc.wantPass {
				t.Fatalf("success = %v, want %v (error: %v)", result.Success, tc.wantPass, result.Error)
			}
			if result.Latency != time.Millisecond*100 {
				t.Errorf("Expected latency 100ms, got %v", result.Latency)
			}
			if result.Reasoning != tc.wantReasoning {
				t.Errorf("reasoning = %v, want %v", result.Reasoning, tc.wantReasoning)
			}
			if tc.wantSays != "" && (result.Error == nil || !strings.Contains(result.Error.Error(), tc.wantSays)) {
				t.Errorf("the failure does not mention %q: %v", tc.wantSays, result.Error)
			}
		})
	}
}

func TestTool_AnsweredInstead_KeepsWhatTheModelSaidInsteadOfCalling(t *testing.T) {
	definitions := testerDefinitions(t, `
- id: "refusal"
  name: "Refusal Instead Of A Tool Call"
  type: "tool"
  group: "basic"
  messages:
    - role: "user"
      content: "Scan 192.168.1.1"
  tools:
    - name: "scan"
      description: "Scans a host"
      parameters:
        type: "object"
        properties:
          host:
            type: "string"
        required: ["host"]
  expected:
    - function_name: "scan"
      arguments:
        host: "192.168.1.1"
  streaming: false
`)

	testCase, err := newToolTestCase(definitions[0])
	if err != nil {
		t.Fatalf("case: %v", err)
	}

	for _, tc := range []struct {
		label      string
		choice     llms.ContentChoice
		wantInText []string
	}{
		{
			label:      "refusal in plain text",
			choice:     llms.ContentChoice{Content: "Sorry, I cannot fulfill your request to perform a network scan.", StopReason: "STOP"},
			wantInText: []string{"STOP", "cannot fulfill"},
		},
		{
			label:      "vendor broke the call itself",
			choice:     llms.ContentChoice{StopReason: "MALFORMED_FUNCTION_CALL"},
			wantInText: []string{"MALFORMED_FUNCTION_CALL"},
		},
		{
			label:      "an answer with no stop reason",
			choice:     llms.ContentChoice{Content: "I'll help you with that.", ToolCalls: []llms.ToolCall{}},
			wantInText: []string{"no tool calls found", "I'll help you with that."},
		},
		{
			label:      "nothing to report",
			choice:     llms.ContentChoice{},
			wantInText: nil,
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			choice := tc.choice
			result := testCase.Execute(&llms.ContentResponse{Choices: []*llms.ContentChoice{&choice}}, time.Second)

			if result.Success {
				t.Fatal("a response with no tool calls must not pass")
			}
			for _, want := range tc.wantInText {
				if !strings.Contains(result.Error.Error(), want) {
					t.Errorf("error %q does not carry %q", result.Error, want)
				}
			}
			if tc.wantInText == nil && result.Error.Error() != "no tool calls found, expected at least 1" {
				t.Errorf("with nothing to add the message must stay bare, got %q", result.Error)
			}
		})
	}
}

func TestTool_FindMatchingToolCall_NamesTheArgumentThatDidNotMatch(t *testing.T) {
	subject := &testCaseTool{}
	expected := ExpectedToolCall{
		FunctionName: "set_temperature",
		Arguments:    map[string]any{"value": "42"},
	}

	calls := []llms.ToolCall{
		{
			ID:           "call-1",
			FunctionCall: &llms.FunctionCall{Arguments: `{"value": "7"}`, Name: "set_temperature"},
		},
	}

	err := subject.findMatchingToolCall(calls, expected)
	if err == nil {
		t.Fatal("a call with the wrong argument was accepted")
	}

	if strings.Contains(err.Error(), "not found in tool calls") {
		t.Errorf("the model did call %s and the argument is what disagreed, but the report says the call is missing: %v",
			expected.FunctionName, err)
	}
	if !strings.Contains(err.Error(), "value") {
		t.Errorf("the report does not name the argument that disagreed: %v", err)
	}
}

func TestTool_ValidateArgumentValue_ComparesByTheExpectedType(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		{"identical json", 42, 42, true},
		{"a float and an int of the same value", 42.0, 42, true},
		{"a string holding the expected int", "42", 42, true},
		{"a float equal to the fifth decimal", 3.141592653, 3.14159, true},
		{"a string in another case", "Hello", "hello", true},
		{"a string holding the expected bool", "true", true, true},
		{"a list holding every expected item", []any{1, 2, 3}, []any{1, 2}, true},
		{"a map missing the expected key", map[string]any{}, map[string]any{"key": "value"}, false},
		{"an expected type no comparison knows",
			[]map[string]any{{"key": "value"}, {"key": "value2"}},
			[]map[string]any{{"key": "value"}, {"key": "value2"}, {"key": "value3"}}, false},
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

// Subtests are keyed by unit: the word before the colon names the compare function.
func TestTool_Compare_AcceptsOnlyAnEquivalentValue(t *testing.T) {
	slice := func(actual, expected any) error { return compareSlice(actual, expected.([]any)) }

	tests := []struct {
		name     string
		compare  func(actual, expected any) error
		actual   any
		expected any
		want     bool
	}{
		{"numeric: a string holding the number", compareNumeric, "42", 42, true},
		{"numeric: a string holding another number", compareNumeric, "43", 42, false},
		{"numeric: a string padded with spaces", compareNumeric, " 42 ", 42, true},
		{"numeric: the same int", compareNumeric, 42, 42, true},
		{"numeric: an unsigned int", compareNumeric, uint(42), 42, true},
		{"numeric: an integral float", compareNumeric, 42.0, 42, true},
		{"numeric: a string holding an integral float", compareNumeric, "42.0", 42, true},
		{"numeric: a float with a fraction", compareNumeric, 42.7, 42, false},
		{"numeric: another float", compareNumeric, 43.7, 42, false},
		{"numeric: another int", compareNumeric, 123, 456, false},
		{"numeric: a string that is no number", compareNumeric, "test", 123, false},
		{"numeric: a type that is no number", compareNumeric, []int{}, 42, false},

		{"float: the same float", compareFloat, 3.14159, 3.14159, true},
		{"float: equal to the fifth decimal", compareFloat, 3.141592653, 3.14159, true},
		{"float: an int equal to an integral float", compareFloat, 3, 3.0, true},
		{"float: an int below the float", compareFloat, 3, 3.14159, false},
		{"float: another int", compareFloat, 4, 3.14159, false},
		{"float: a string holding the float", compareFloat, "123.5", 123.5, true},
		{"float: a string cut short", compareFloat, "3.14", 3.14159, false},
		{"float: a string containing the float", compareFloat, "value: 3.14000 found", 3.14, true},
		{"float: a type that is no number", compareFloat, []int{}, 3.14, false},

		{"string: another case", compareString, "Hello", "hello", true},
		{"string: an expected value in upper case", compareString, "Test", "TEST", true},
		{"string: padded with spaces", compareString, " Hello ", "hello", true},
		{"string: a long string containing the expected", compareString, "This is a very long test message", "test message", true},
		{"string: a short string inside the long expected", compareString, "test", "This is a very long test message", true},
		{"string: two short different words", compareString, "hello", "world", false},
		{"string: the expected words with others between", compareString, "Golang Go programming language, overview", "golang programming language", true},
		{"string: an address range for one address", compareString, "192.168.1.1-100", "192.168.1.100", false},
		{"string: an address list for one address", compareString, "192.168.1.1, 10.0.0.100", "192.168.1.100", false},
		{"string: an int", compareString, 42, "42", true},
		{"string: a float", compareString, 3.14, "3.14", true},
		{"string: a type that is no string", compareString, []int{}, "test", false},

		{"bool: the same true", compareBool, true, true, true},
		{"bool: the same false", compareBool, false, false, true},
		{"bool: the opposite", compareBool, true, false, false},
		{"bool: a string true", compareBool, "true", true, true},
		{"bool: a string false", compareBool, "false", false, true},
		{"bool: a quoted string", compareBool, "'true'", true, true},
		{"bool: a padded string", compareBool, " true ", true, true},
		{"bool: a string that is no bool", compareBool, "yes", true, false},
		{"bool: a number", compareBool, 1, true, false},

		{"slice: a list holding every expected item", slice, []any{1, 2, 3}, []any{1, 2}, true},
		{"slice: a list missing an expected item", slice, []any{1, 2}, []any{1, 2, 3}, false},
		{"slice: nothing expected", slice, []any{1, 2, 3}, []any{}, true},
		{"slice: a value among the expected", slice, "hello", []any{"hello", "world"}, true},
		{"slice: a value not among the expected", slice, "test", []any{"hello", "world"}, false},
		{"slice: an int among the expected", slice, 42, []any{41, 42, 43}, true},
		{"slice: a type that is no list", slice, map[string]int{}, []any{1, 2}, false},

		{"map: the same map", compareMap, map[string]any{"key": "value"}, map[string]any{"key": "value"}, true},
		{"map: a missing key", compareMap, map[string]any{}, map[string]any{"key": "value"}, false},
		{"map: a wrong value", compareMap, map[string]any{"key": "wrong"}, map[string]any{"key": "value"}, false},
		{"map: extra keys", compareMap, map[string]any{"key": "value", "extra": "ok"}, map[string]any{"key": "value"}, true},
		{"map: a nested map", compareMap, map[string]any{"key": map[string]any{"nested": "value"}},
			map[string]any{"key": map[string]any{"nested": "value"}}, true},
		{"map: a type that is no map", compareMap, "string", map[string]any{"key": "value"}, false},
		{"map: a list of maps holding a match", compareMap, []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "value"}, true},
		{"map: a list of maps without a match", compareMap, []map[string]any{{"key": "value"}, {"key": "value2"}},
			map[string]any{"key": "wrong"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.compare(tt.actual, tt.expected)
			if succeed := err == nil; succeed != tt.want {
				t.Errorf("compare(%v, %v) = %v, want %v, error: %v", tt.actual, tt.expected, succeed, tt.want, err)
			}
		})
	}
}

func TestTool_Execute_JudgesTheCatalogueCasesByTheFactsTheyAskFor(t *testing.T) {
	for _, tc := range []struct {
		label    string
		caseID   string
		function string
		args     map[string]any
		wantPass bool
	}{
		{"open ports listed together", "pentest_memory_tool_call", "generate_report",
			map[string]any{"target": "192.168.1.100", "findings": []string{"Open ports: 22, 80, 443", "SQL injection in the login form"}}, true},
		{"one line per service", "pentest_memory_tool_call", "generate_report",
			map[string]any{"target": "192.168.1.100", "findings": []string{"22/tcp open (SSH)", "80/tcp open (HTTP)", "443/tcp open (HTTPS)", "SQL injection in login form"}}, true},
		{"port named in the singular", "pentest_memory_tool_call", "generate_report",
			map[string]any{"target": "192.168.1.100", "findings": []string{"Port 22 open", "Port 80 open", "Port 443 open", "SQL injection found"}}, true},
		{"a report without a single port", "pentest_memory_tool_call", "generate_report",
			map[string]any{"target": "192.168.1.100", "findings": []string{"SQL injection vulnerability in the login form"}}, false},

		{"the exact query", "search_function", "search", map[string]any{"query": "Golang programming language"}, true},
		{"words added after the query", "search_function", "search", map[string]any{"query": "Golang programming language overview features"}, true},
		{"a word added inside the query", "search_function", "search", map[string]any{"query": "Golang Go programming language overview"}, true},
		{"lower case with words added", "search_function", "search", map[string]any{"query": "golang go programming language overview features"}, true},
		{"another language", "search_function", "search", map[string]any{"query": "Rust programming language overview"}, false},
		{"the words out of order", "search_function", "search", map[string]any{"query": "programming language Golang"}, false},
	} {
		t.Run(tc.label, func(t *testing.T) {
			testCase, err := newToolTestCase(builtinDefinition(t, tc.caseID))
			if err != nil {
				t.Fatalf("case: %v", err)
			}
			arguments, err := json.Marshal(tc.args)
			if err != nil {
				t.Fatalf("arguments: %v", err)
			}

			result := testCase.Execute(&llms.ContentResponse{Choices: []*llms.ContentChoice{{ToolCalls: []llms.ToolCall{{
				ID:           "call-1",
				FunctionCall: &llms.FunctionCall{Name: tc.function, Arguments: string(arguments)},
			}}}}}, time.Second)
			if result.Success != tc.wantPass {
				t.Errorf("success = %v, want %v for arguments %v (error: %v)", result.Success, tc.wantPass, tc.args, result.Error)
			}
		})
	}
}
