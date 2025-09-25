package testdata

import "time"

// TestResult represents the result of a single test execution
type TestResult struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      TestType      `json:"type"`
	Group     TestGroup     `json:"group"`
	Success   bool          `json:"success"`
	Error     error         `json:"error"`
	Streaming bool          `json:"streaming"`
	Reasoning bool          `json:"reasoning"`
	Latency   time.Duration `json:"latency"`
}
