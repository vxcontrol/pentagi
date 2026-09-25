package cases

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

// Subtests are keyed by unit: GetTestsByGroup, GetTestsByType and GetAllTests.
func TestRegistry_FiltersTheLoadedDefinitionsByGroupAndType(t *testing.T) {
	registry, err := LoadRegistryFromYAML([]byte(`
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
`))
	if err != nil {
		t.Fatalf("Failed to load registry from YAML: %v", err)
	}

	if len(registry.definitions) != 3 {
		t.Fatalf("Expected 3 definitions, got %d", len(registry.definitions))
	}

	t.Run("by group", func(t *testing.T) {
		for group, want := range map[TestGroup]int{TestGroupBasic: 2, TestGroupJSON: 1, TestGroupKnowledge: 0} {
			if got := len(registry.GetTestsByGroup(group)); got != want {
				t.Errorf("GetTestsByGroup(%s) = %d definitions, want %d", group, got, want)
			}
		}
	})
	t.Run("by type", func(t *testing.T) {
		for testType, want := range map[TestType]int{TestTypeCompletion: 1, TestTypeJSON: 1, TestTypeTool: 1} {
			if got := len(registry.GetTestsByType(testType)); got != want {
				t.Errorf("GetTestsByType(%s) = %d definitions, want %d", testType, got, want)
			}
		}
	})
	t.Run("all of them", func(t *testing.T) {
		if got := len(registry.GetAllTests()); got != 3 {
			t.Errorf("Expected 3 total tests, got %d", got)
		}
	})
}

func TestRegistry_GetTestSuite_BuildsTheCasesOfOneGroup(t *testing.T) {
	registry, err := LoadRegistryFromYAML([]byte(`
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
`))
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

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
	for i, testCase := range suite.Tests {
		if testCase.Group() != TestGroupBasic {
			t.Errorf("Test %d: expected group basic, got %s", i, testCase.Group())
		}
		if testCase.Type() != TestTypeCompletion {
			t.Errorf("Test %d: expected type completion, got %s", i, testCase.Type())
		}
	}
	if !suite.Tests[1].Streaming() {
		t.Errorf("Expected test2 to have streaming enabled")
	}
	if suite.Tests[0].Streaming() {
		t.Errorf("Expected test1 to have streaming disabled")
	}

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

	emptySuite, err := registry.GetTestSuite(TestGroupKnowledge)
	if err != nil {
		t.Fatalf("Failed to get empty test suite: %v", err)
	}
	if len(emptySuite.Tests) != 0 {
		t.Errorf("Expected 0 tests in knowledge suite, got %d", len(emptySuite.Tests))
	}
}

func TestRegistry_GetTestSuite_RefusesACaseItCannotBuild(t *testing.T) {
	registry, err := LoadRegistryFromYAML([]byte(`
- id: "test1"
  name: "Test 1"
  type: "completion"
  group: "basic"
  prompt: "Test 1"
  expected: 123
  streaming: false
`))
	if err != nil {
		t.Fatalf("the definition is well formed, so the load must accept it: %v", err)
	}

	_, err = registry.GetTestSuite(TestGroupBasic)
	if err == nil || !strings.Contains(err.Error(), "completion test expected must be string") {
		t.Errorf("a completion case expecting a number must be refused when its suite is built, got: %v", err)
	}
}

func TestRegistry_LoadRegistryFromYAML_RefusesTheWholeCatalogueForOneBadCase(t *testing.T) {
	tests := []struct{ name, yaml, want string }{
		{
			name: "a misspelled key",
			yaml: `
- id: "p"
  name: "P"
  type: "completion"
  group: "basic"
  prompt: "hi"
  expected: "ok"
  expect_refusals: "structured_output_config"
`,
			want: "field expect_refusals not found",
		},
		{
			name: "two cases under one id",
			yaml: `
- id: "p"
  name: "P"
  type: "completion"
  group: "basic"
  prompt: "hi"
  expected: "ok"
- id: "p"
  name: "P again"
  type: "completion"
  group: "basic"
  prompt: "hi"
  expected: "ok"
`,
			want: "defined twice",
		},
		{name: "malformed yaml", yaml: `invalid yaml content {{{`, want: "failed to parse YAML"},
		{
			name: "a refused case in a group nobody runs",
			yaml: `
- id: "fine"
  name: "Fine"
  type: "completion"
  group: "basic"
  prompt: "hi"
  expected: "ok"
- id: "bad"
  name: "Bad"
  type: "completion"
  group: "knowledge"
  params:
    model: "some-other-model"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
`,
			want: "test case bad: params must not set a model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadRegistryFromYAML([]byte(tt.yaml))
			if err == nil {
				t.Fatalf("%s must be refused at load", tt.name)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not mention %q", err, tt.want)
			}
		})
	}
}

func TestRegistry_LoadBuiltinRegistry_BuildsASuiteForEveryGroup(t *testing.T) {
	registry, err := LoadBuiltinRegistry()
	if err != nil {
		t.Fatalf("Failed to load builtin registry: %v", err)
	}
	if len(registry.GetAllTests()) == 0 {
		t.Errorf("Expected builtin registry to contain some tests")
	}
	for _, group := range []TestGroup{TestGroupBasic, TestGroupAdvanced, TestGroupJSON, TestGroupKnowledge} {
		if _, err := registry.GetTestSuite(group); err != nil {
			t.Errorf("Failed to create test suite for group %s: %v", group, err)
		}
	}
}

const anthropicBudgetFloor = 1024

func TestRegistry_LoadBuiltinRegistry_LeavesEveryCappedCaseRoomForTheThinkingBudget(t *testing.T) {
	registry, err := LoadBuiltinRegistry()
	if err != nil {
		t.Fatalf("loading the builtin registry: %v", err)
	}

	reasoning := &llms.ReasoningConfig{Effort: llms.ReasoningLow}

	for _, def := range registry.definitions {
		if def.Params == nil || def.Params.MaxTokens <= 0 {
			continue
		}

		if budget := reasoning.GetTokens(def.Params.MaxTokens); budget < anthropicBudgetFloor {
			t.Errorf(
				"case %s caps output at %d, which leaves a thinking budget of %d: every agent with budget thinking fails this case on the vendor, not on the answer",
				def.ID, def.Params.MaxTokens, budget,
			)
		}
	}
}
