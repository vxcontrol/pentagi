package testdata

import (
	"testing"
)

func TestRegistryLoad(t *testing.T) {
	testYAML := `
- id: "test_basic"
  name: "Basic Test"
  type: "completion"
  group: "basic"
  prompt: "What is 2+2?"
  expected: "4"
  streaming: false

- id: "test_json"
  name: "JSON Test"
  type: "json"
  group: "json"
  messages:
    - role: "user"
      content: "Return JSON"
  expected:
    name: "test"
  streaming: false

- id: "test_tool"
  name: "Tool Test"
  type: "tool"
  group: "basic"
  messages:
    - role: "user"
      content: "Use echo function"
  tools:
    - name: "echo"
      description: "Echo function"
      parameters:
        type: "object"
        properties:
          message:
            type: "string"
        required: ["message"]
  expected:
    - function_name: "echo"
      arguments:
        message: "hello"
  streaming: false
`

	// test LoadRegistryFromYAML
	registry, err := LoadRegistryFromYAML([]byte(testYAML))
	if err != nil {
		t.Fatalf("Failed to load registry from YAML: %v", err)
	}

	if len(registry.definitions) != 3 {
		t.Fatalf("Expected 3 definitions, got %d", len(registry.definitions))
	}

	// test GetTestsByGroup
	basicTests := registry.GetTestsByGroup(TestGroupBasic)
	if len(basicTests) != 2 {
		t.Errorf("Expected 2 basic tests, got %d", len(basicTests))
	}

	jsonTests := registry.GetTestsByGroup(TestGroupJSON)
	if len(jsonTests) != 1 {
		t.Errorf("Expected 1 JSON test, got %d", len(jsonTests))
	}

	knowledgeTests := registry.GetTestsByGroup(TestGroupKnowledge)
	if len(knowledgeTests) != 0 {
		t.Errorf("Expected 0 knowledge tests, got %d", len(knowledgeTests))
	}

	// test GetTestsByType
	completionTests := registry.GetTestsByType(TestTypeCompletion)
	if len(completionTests) != 1 {
		t.Errorf("Expected 1 completion test, got %d", len(completionTests))
	}

	jsonTypeTests := registry.GetTestsByType(TestTypeJSON)
	if len(jsonTypeTests) != 1 {
		t.Errorf("Expected 1 JSON type test, got %d", len(jsonTypeTests))
	}

	toolTests := registry.GetTestsByType(TestTypeTool)
	if len(toolTests) != 1 {
		t.Errorf("Expected 1 tool test, got %d", len(toolTests))
	}

	// test GetAllTests
	allTests := registry.GetAllTests()
	if len(allTests) != 3 {
		t.Errorf("Expected 3 total tests, got %d", len(allTests))
	}
}

func TestTestSuiteCreation(t *testing.T) {
	testYAML := `
- id: "test1"
  name: "Test 1"
  type: "completion"
  group: "basic"
  prompt: "Test 1"
  expected: "result1"
  streaming: false

- id: "test2"
  name: "Test 2"
  type: "completion"
  group: "basic"
  prompt: "Test 2"
  expected: "result2"
  streaming: true

- id: "test3"
  name: "Test 3"
  type: "json"
  group: "advanced"
  messages:
    - role: "user"
      content: "Return JSON"
  expected:
    key: "value"
  streaming: false
`

	registry, err := LoadRegistryFromYAML([]byte(testYAML))
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// test GetTestSuite for basic group
	suite, err := registry.GetTestSuite(TestGroupBasic)
	if err != nil {
		t.Fatalf("Failed to get test suite: %v", err)
	}

	if suite.Group != TestGroupBasic {
		t.Errorf("Expected suite group 'basic', got %s", suite.Group)
	}
	if len(suite.Tests) != 2 {
		t.Fatalf("Expected 2 tests in basic suite, got %d", len(suite.Tests))
	}

	// verify test cases are properly created
	for i, testCase := range suite.Tests {
		if testCase.Group() != TestGroupBasic {
			t.Errorf("Test %d: expected group basic, got %s", i, testCase.Group())
		}
		if testCase.Type() != TestTypeCompletion {
			t.Errorf("Test %d: expected type completion, got %s", i, testCase.Type())
		}
	}

	// test streaming configuration
	if !suite.Tests[1].Streaming() {
		t.Errorf("Expected test2 to have streaming enabled")
	}
	if suite.Tests[0].Streaming() {
		t.Errorf("Expected test1 to have streaming disabled")
	}

	// test GetTestSuite for advanced group
	advancedSuite, err := registry.GetTestSuite(TestGroupAdvanced)
	if err != nil {
		t.Fatalf("Failed to get advanced test suite: %v", err)
	}

	if len(advancedSuite.Tests) != 1 {
		t.Fatalf("Expected 1 test in advanced suite, got %d", len(advancedSuite.Tests))
	}
	if advancedSuite.Tests[0].Type() != TestTypeJSON {
		t.Errorf("Expected JSON test in advanced suite, got %s", advancedSuite.Tests[0].Type())
	}

	// test empty group
	emptySuite, err := registry.GetTestSuite(TestGroupKnowledge)
	if err != nil {
		t.Fatalf("Failed to get empty test suite: %v", err)
	}
	if len(emptySuite.Tests) != 0 {
		t.Errorf("Expected 0 tests in knowledge suite, got %d", len(emptySuite.Tests))
	}
}

func TestRegistryErrors(t *testing.T) {
	// test invalid YAML
	invalidYAML := `
- id: "test1"
  name: "Test 1"
  type: "completion"
  group: "basic"
  prompt: "Test 1"
  expected: 123  # Invalid: completion tests need string expected
  streaming: false
`

	registry, err := LoadRegistryFromYAML([]byte(invalidYAML))
	if err != nil {
		t.Fatalf("Failed to load registry with invalid test: %v", err)
	}

	// should fail when creating test suite due to invalid test definition
	_, err = registry.GetTestSuite(TestGroupBasic)
	if err == nil {
		t.Errorf("Expected error when creating test suite with invalid completion test")
	}

	// test malformed YAML
	malformedYAML := `invalid yaml content {{{`
	_, err = LoadRegistryFromYAML([]byte(malformedYAML))
	if err == nil {
		t.Errorf("Expected error for malformed YAML")
	}

	// test unknown test type
	unknownTypeYAML := `
- id: "test1"
  name: "Test 1"
  type: "unknown_type"
  group: "basic"
  prompt: "Test 1"
  expected: "result1"
  streaming: false
`

	registry, err = LoadRegistryFromYAML([]byte(unknownTypeYAML))
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	_, err = registry.GetTestSuite(TestGroupBasic)
	if err == nil {
		t.Errorf("Expected error for unknown test type")
	}
}

func TestBuiltinRegistry(t *testing.T) {
	// test that builtin registry loads without error
	registry, err := LoadBuiltinRegistry()
	if err != nil {
		t.Fatalf("Failed to load builtin registry: %v", err)
	}

	// basic smoke test - should have some tests
	allTests := registry.GetAllTests()
	if len(allTests) == 0 {
		t.Errorf("Expected builtin registry to contain some tests")
	}

	// test that we can create test suites from builtin tests
	for _, group := range []TestGroup{TestGroupBasic, TestGroupAdvanced, TestGroupJSON, TestGroupKnowledge} {
		_, err := registry.GetTestSuite(group)
		if err != nil {
			t.Errorf("Failed to create test suite for group %s: %v", group, err)
		}
	}
}
