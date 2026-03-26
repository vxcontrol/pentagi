package tester

import (
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/tester/testdata"
)

// testConfig holds private configuration for test execution
type testConfig struct {
	agentTypes      []pconfig.ProviderOptionsType
	groups          []testdata.TestGroup
	streamingMode   bool
	verbose         bool
	parallelWorkers int
	customRegistry  *testdata.TestRegistry
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
func WithGroups(groups ...testdata.TestGroup) TestOption {
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
func WithCustomRegistry(registry *testdata.TestRegistry) TestOption {
	return func(c *testConfig) {
		c.customRegistry = registry
	}
}

// defaultConfig returns default test configuration
func defaultConfig() *testConfig {
	return &testConfig{
		agentTypes:      pconfig.AllAgentTypes,
		groups:          []testdata.TestGroup{testdata.TestGroupBasic, testdata.TestGroupAdvanced, testdata.TestGroupKnowledge},
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
