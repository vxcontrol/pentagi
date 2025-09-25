package main

import "time"

// TestResult represents the result of a single test for CLI compatibility
type TestResult struct {
	Name      string
	Type      string
	Success   bool
	Error     error
	Streaming bool
	Reasoning bool
	LatencyMs int64
	Response  string
	Expected  string
}

// AgentTestResult collects test results for each agent type for CLI compatibility
type AgentTestResult struct {
	AgentType       string
	ModelName       string
	Reasoning       bool
	BasicTests      []TestResult
	AdvancedTests   []TestResult
	TotalSuccess    int
	TotalTests      int
	AverageLatency  time.Duration
	SkippedAdvanced bool
	SkippedReason   string
}
