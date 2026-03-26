package testdata

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed tests.yml
var testsData embed.FS

// TestRegistry manages test definitions and creates test suites
type TestRegistry struct {
	definitions []TestDefinition
}

// LoadBuiltinRegistry loads test definitions from embedded tests.yml
func LoadBuiltinRegistry() (*TestRegistry, error) {
	data, err := testsData.ReadFile("tests.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read builtin tests: %w", err)
	}
	return LoadRegistryFromYAML(data)
}

// LoadRegistryFromYAML creates registry from YAML data
func LoadRegistryFromYAML(data []byte) (*TestRegistry, error) {
	var definitions []TestDefinition
	if err := yaml.Unmarshal(data, &definitions); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &TestRegistry{definitions: definitions}, nil
}

// GetTestSuite creates a test suite with stateful test cases for a specific group
func (r *TestRegistry) GetTestSuite(group TestGroup) (*TestSuite, error) {
	var testCases []TestCase
	for _, def := range r.definitions {
		if def.Group == group {
			testCase, err := r.createTestCase(def)
			if err != nil {
				return nil, fmt.Errorf("failed to create test case %s: %w", def.ID, err)
			}
			testCases = append(testCases, testCase)
		}
	}

	return &TestSuite{
		Group: group,
		Tests: testCases,
	}, nil
}

// GetTestsByGroup returns test definitions filtered by group
func (r *TestRegistry) GetTestsByGroup(group TestGroup) []TestDefinition {
	var filtered []TestDefinition
	for _, def := range r.definitions {
		if def.Group == group {
			filtered = append(filtered, def)
		}
	}
	return filtered
}

// GetTestsByType returns test definitions filtered by type
func (r *TestRegistry) GetTestsByType(testType TestType) []TestDefinition {
	var filtered []TestDefinition
	for _, def := range r.definitions {
		if def.Type == testType {
			filtered = append(filtered, def)
		}
	}
	return filtered
}

// GetAllTests returns all test definitions
func (r *TestRegistry) GetAllTests() []TestDefinition {
	return r.definitions
}

// createTestCase creates appropriate test case implementation based on type
func (r *TestRegistry) createTestCase(def TestDefinition) (TestCase, error) {
	switch def.Type {
	case TestTypeCompletion:
		return newCompletionTestCase(def)
	case TestTypeJSON:
		return newJSONTestCase(def)
	case TestTypeTool:
		return newToolTestCase(def)
	default:
		return nil, fmt.Errorf("unknown test type: %s", def.Type)
	}
}
