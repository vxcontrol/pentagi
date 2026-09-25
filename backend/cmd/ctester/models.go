package main

import "time"

// TestResult represents the result of a single test for CLI compatibility
type TestResult struct {
	Name       string
	Type       string
	Capability string
	Success    bool
	// Unsupported means the case could not exercise what it asks for on this
	// configuration (a typed capability refusal, or an output limit the library
	// changed before the vendor), and Error says why. It does not count against
	// TotalSuccess/TotalTests.
	Unsupported     bool
	ContentFiltered bool
	Error           error
	StopReason      string
	Streaming       bool
	Reasoning       bool
	LatencyMs       int64
	Response        string
	Expected        string
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
	TotalFiltered   int
	AverageLatency  time.Duration
	SkippedAdvanced bool
	SkippedReason   string
}
