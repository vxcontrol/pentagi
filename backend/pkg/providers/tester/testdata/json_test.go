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
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(definitions))
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
}

func TestJSONValueValidation(t *testing.T) {
	tests := []struct {
		name     string
		actual   any
		expected any
		want     bool
	}{
		// Basic exact matches
		{"string_exact", "test", "test", true},
		{"int_exact", 123, 123, true},
		{"bool_exact", true, true, true},

		// JSON unmarshaling type conversions
		{"float_to_int", 123.0, 123, true},     // JSON unmarshaling produces float64
		{"int_to_float", 123, 123.0, true},     // int to float64 conversion
		{"string_int", "123", 123, true},       // string to int conversion
		{"string_float", "123.5", 123.5, true}, // string to float conversion
		{"string_bool", "true", true, true},    // string to bool conversion

		// Case insensitive string matching
		{"string_case", "TEST", "test", true},
		{"string_case_mixed", "Test", "TEST", true},

		// Failures
		{"string_different", "test", "other", false},
		{"int_different", 123, 456, false},
		{"bool_different", true, false, false},
		{"type_mismatch", "test", 123, false},

		// JSON-specific scenarios
		{"json_string_number", "42", 42, true},
		{"json_string_float", "3.14", 3.14, true},
		{"json_bool_string", "false", false, true},
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
