package tester

import (
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/tester/cases"
)

// testConfig holds private configuration for test execution
type testConfig struct {
	agentTypes      []pconfig.ProviderOptionsType
	groups          []cases.TestGroup
	streamingMode   bool
	verbose         bool
	parallelWorkers int
	customRegistry  *cases.TestRegistry
}

// TestOption configures test execution
type TestOption func(*testConfig)

// WithAgentTypes filters tests to specific agent types
func WithAgentTypes(types ...pconfig.ProviderOptionsType) TestOption {
	return func(c *testConfig) {
		c.agentTypes = types
	}
}

// WithGroups filters tests to specific groups
func WithGroups(groups ...cases.TestGroup) TestOption {
	return func(c *testConfig) {
		c.groups = groups
	}
}

// WithStreamingMode enables/disables streaming tests
func WithStreamingMode(enabled bool) TestOption {
	return func(c *testConfig) {
		c.streamingMode = enabled
	}
}

// WithVerbose enables verbose output during testing
func WithVerbose(enabled bool) TestOption {
	return func(c *testConfig) {
		c.verbose = enabled
	}
}

// WithParallelWorkers sets the number of parallel workers
func WithParallelWorkers(workers int) TestOption {
	return func(c *testConfig) {
		if workers > 0 {
			c.parallelWorkers = workers
		}
	}
}

// WithCustomRegistry sets a custom test registry
func WithCustomRegistry(registry *cases.TestRegistry) TestOption {
	return func(c *testConfig) {
		c.customRegistry = registry
	}
}

// defaultConfig returns default test configuration
func defaultConfig() *testConfig {
	return &testConfig{
		agentTypes:      pconfig.AllAgentTypes,
		groups:          []cases.TestGroup{cases.TestGroupBasic, cases.TestGroupAdvanced, cases.TestGroupJSON, cases.TestGroupKnowledge},
		streamingMode:   true,
		verbose:         false,
		parallelWorkers: 4,
	}
}

// applyOptions applies test options to configuration
func applyOptions(opts []TestOption) *testConfig {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}
