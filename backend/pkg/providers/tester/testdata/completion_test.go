package testdata

import (
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func TestCompletionTestCase(t *testing.T) {
	testYAML := `
- id: "test_basic"
  name: "Basic Math Test"
  type: "completion"
  group: "basic"
  prompt: "What is 2+2?"
  expected: "4"
  streaming: false

- id: "test_messages"
  name: "System User Test"
  type: "completion"  
  group: "basic"
  messages:
    - role: "system"
      content: "You are a math assistant"
    - role: "user"
      content: "Calculate 5 * 10"
  expected: "50"
  streaming: false
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(definitions) != 2 {
		t.Fatalf("Expected 2 definitions, got %d", len(definitions))
	}

	// test basic completion case
	basicDef := definitions[0]
	testCase, err := newCompletionTestCase(basicDef)
	if err != nil {
		t.Fatalf("Failed to create basic test case: %v", err)
	}

	if testCase.ID() != "test_basic" {
		t.Errorf("Expected ID 'test_basic', got %s", testCase.ID())
	}
	if testCase.Type() != TestTypeCompletion {
		t.Errorf("Expected type completion, got %s", testCase.Type())
	}
	if testCase.Prompt() != "What is 2+2?" {
		t.Errorf("Expected prompt 'What is 2+2?', got %s", testCase.Prompt())
	}
	if len(testCase.Messages()) != 0 {
		t.Errorf("Expected no messages for basic test, got %d", len(testCase.Messages()))
	}

	// test execution with correct response
	result := testCase.Execute("The answer is 4", time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for correct response, got failure: %v", result.Error)
	}
	if result.Latency != time.Millisecond*100 {
		t.Errorf("Expected latency 100ms, got %v", result.Latency)
	}

	// test execution with incorrect response
	result = testCase.Execute("The answer is 5", time.Millisecond*50)
	if result.Success {
		t.Errorf("Expected failure for incorrect response, got success")
	}

	// test messages case
	messagesDef := definitions[1]
	testCase, err = newCompletionTestCase(messagesDef)
	if err != nil {
		t.Fatalf("Failed to create messages test case: %v", err)
	}

	if len(testCase.Messages()) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(testCase.Messages()))
	}

	// test with ContentResponse
	response := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "The result is 50",
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*200)
	if !result.Success {
		t.Errorf("Expected success for ContentResponse, got failure: %v", result.Error)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		response string
		expected string
		want     bool
	}{
		{"4", "4", true},
		{"The answer is 4", "4", true},
		{"1, 2, 3, 4, 5", "1, 2, 3, 4, 5", true},
		{"1,2,3,4,5", "1, 2, 3, 4, 5", true},
		{"Response: 1, 2, 3, 4, 5", "1, 2, 3, 4, 5", true},
		{"HELLO WORLD", "hello world", true},
		{"wrong answer", "4", false},
		{"", "4", false},
	}

	for _, tt := range tests {
		got := containsString(tt.response, tt.expected)
		if got != tt.want {
			t.Errorf("containsString(%q, %q) = %v, want %v", tt.response, tt.expected, got, tt.want)
		}
	}
}
