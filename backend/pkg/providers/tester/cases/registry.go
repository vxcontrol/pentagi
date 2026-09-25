package cases

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"slices"

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
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&definitions); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	seen := make(map[string]bool, len(definitions))
	for _, def := range definitions {
		if seen[def.ID] {
			return nil, fmt.Errorf("test case %s is defined twice", def.ID)
		}
		seen[def.ID] = true
		if err := validateDefinition(def); err != nil {
			return nil, fmt.Errorf("test case %s: %w", def.ID, err)
		}
	}

	return &TestRegistry{definitions: definitions}, nil
}

type SuiteOption func(*suiteOptions)

type suiteOptions struct {
	toolCallIDTemplate string
}

func WithToolCallIDTemplate(template string) SuiteOption {
	return func(options *suiteOptions) {
		options.toolCallIDTemplate = template
	}
}

// GetTestSuite creates a test suite with stateful test cases for a specific group
func (r *TestRegistry) GetTestSuite(group TestGroup, opts ...SuiteOption) (*TestSuite, error) {
	var options suiteOptions
	for _, opt := range opts {
		opt(&options)
	}

	var testCases []TestCase
	for _, def := range r.definitions {
		if def.Group == group {
			if options.toolCallIDTemplate != "" {
				def.Messages = def.Messages.withToolCallIDs(options.toolCallIDTemplate)
			}
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

func (r *TestRegistry) ReplaysToolCalls(groups ...TestGroup) bool {
	for _, def := range r.definitions {
		if !slices.Contains(groups, def.Group) {
			continue
		}
		for _, msg := range def.Messages {
			if len(msg.ToolCalls) > 0 || msg.ToolCallID != "" {
				return true
			}
		}
	}
	return false
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
	if err := validateDefinition(def); err != nil {
		return nil, err
	}

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
