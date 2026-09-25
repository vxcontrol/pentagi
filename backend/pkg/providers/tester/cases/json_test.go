package cases

import (
	"testing"
	"time"
)

func TestJSON_Execute_RequiresEveryExpectedField(t *testing.T) {
	definitions := testerDefinitions(t, `
- id: "test_object"
  name: "JSON Object Test"
  type: "json"
  group: "json"
  messages:
    - role: "system"
      content: "Respond with JSON only"
    - role: "user"
      content: "Create person info"
  expected:
    name: "John Doe"
    age: 30
  streaming: false
`)

	testCase, err := newJSONTestCase(definitions[0])
	if err != nil {
		t.Fatalf("Failed to create JSON object test case: %v", err)
	}

	if testCase.ID() != "test_object" {
		t.Errorf("Expected ID 'test_object', got %s", testCase.ID())
	}
	if testCase.Type() != TestTypeJSON {
		t.Errorf("Expected type json, got %s", testCase.Type())
	}
	if len(testCase.Messages()) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(testCase.Messages()))
	}

	result := testCase.Execute(`{"name": "John Doe", "age": 30, "city": "New York"}`, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for valid JSON, got failure: %v", result.Error)
	}

	result = testCase.Execute(`{"name": "John Doe"}`, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for missing required field, got success")
	}
}
