package main

import "time"

// TestResult represents the result of a single test for CLI compatibility
type TestResult struct {
	Name       string
	Type       string
	Capability string
	Success    bool
	// Unsupported means the provider/model proactively rejected the requested
	// capability (e.g. structured output, or disabling reasoning) with a typed
	// SDK error rather than the call failing for an unrelated reason. It does
	// not count against TotalSuccess/TotalTests: an optional capability being
	// unavailable on a given model is informative, not a defect in the config.
	Unsupported bool
	Error       error
	Streaming   bool
	Reasoning   bool
	LatencyMs   int64
	Response    string
	Expected    string
}

// AgentTestResult collects test results for each agent type for CLI compatibility
type AgentTestResult struct {
	AgentType     string
	ModelName     string
	Reasoning     bool
	BasicTests    []TestResult
	AdvancedTests []TestResult
	// CapabilityTests holds results for capability-gated tests (adaptive
	// thinking, reasoning off, structured output) kept separate from
	// Basic/AdvancedTests so the report can call out "does this optional
	// capability work" without it affecting the primary pass/fail tally.
	CapabilityTests []TestResult
	TotalSuccess    int
	TotalTests      int
	AverageLatency  time.Duration
	SkippedAdvanced bool
	SkippedReason   string
}
