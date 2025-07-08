package testdata

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestJSONTestCase(t *testing.T) {
	testYAML := `
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

- id: "test_array"
  name: "JSON Array Test"
  type: "json"
  group: "json"
  messages:
    - role: "user"
      content: "Create JSON array"
  is_array: true
  streaming: false

- id: "test_failure"
  name: "JSON Failure Test"
  type: "json"
  group: "json"
  messages:
    - role: "user"
      content: "Return invalid JSON"
  expect_failure: true
  streaming: false
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(definitions) != 3 {
		t.Fatalf("Expected 3 definitions, got %d", len(definitions))
	}

	// test JSON object case
	objectDef := definitions[0]
	testCase, err := newJSONTestCase(objectDef)
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

	// test execution with valid JSON
	validJSON := `{"name": "John Doe", "age": 30, "city": "New York"}`
	result := testCase.Execute(validJSON, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for valid JSON, got failure: %v", result.Error)
	}

	// test execution with missing field
	invalidJSON := `{"name": "John Doe"}`
	result = testCase.Execute(invalidJSON, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for missing required field, got success")
	}

	// test JSON array case
	arrayDef := definitions[1]
	testCase, err = newJSONTestCase(arrayDef)
	if err != nil {
		t.Fatalf("Failed to create JSON array test case: %v", err)
	}

	// test execution with valid array
	validArray := `[{"name": "red", "hex": "#FF0000"}, {"name": "blue", "hex": "#0000FF"}]`
	result = testCase.Execute(validArray, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for valid JSON array, got failure: %v", result.Error)
	}

	// test execution with invalid array
	invalidArray := `{"not": "array"}`
	result = testCase.Execute(invalidArray, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure for invalid JSON array, got success")
	}

	// test failure case
	failureDef := definitions[2]
	testCase, err = newJSONTestCase(failureDef)
	if err != nil {
		t.Fatalf("Failed to create JSON failure test case: %v", err)
	}

	// test execution with invalid JSON (should succeed as we expect failure)
	invalidJSON = `{invalid json`
	result = testCase.Execute(invalidJSON, time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for expected failure case, got failure: %v", result.Error)
	}

	// test execution with valid JSON (should fail as we expect failure)
	validJSON = `{"valid": "json"}`
	result = testCase.Execute(validJSON, time.Millisecond*100)
	if result.Success {
		t.Errorf("Expected failure when valid JSON provided for expect_failure test, got success")
	}
}

func TestJSONValuesEqual(t *testing.T) {
	tests := []struct {
		actual   interface{}
		expected interface{}
		want     bool
	}{
		{"test", "test", true},
		{123, 123, true},
		{123.0, 123, true}, // float64 to int conversion
		{123, 123.0, true}, // int to float64 conversion
		{true, true, true},
		{"test", "other", false},
		{123, 456, false},
		{true, false, false},
	}

	for _, tt := range tests {
		got := jsonValuesEqual(tt.actual, tt.expected)
		if got != tt.want {
			t.Errorf("jsonValuesEqual(%v, %v) = %v, want %v", tt.actual, tt.expected, got, tt.want)
		}
	}
}
