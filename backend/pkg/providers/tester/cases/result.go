package cases

import "time"

// TestResult represents the result of a single test execution
type TestResult struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       TestType       `json:"type"`
	Group      TestGroup      `json:"group"`
	Capability TestCapability `json:"capability,omitempty"`
	Success    bool           `json:"success"`
	// Unsupported reports that the case could not exercise what it asks for on
	// this configuration, and Error says why: either the provider/model refused
	// the capability with a typed error, or the library changed the output limit
	// the case sets before it reached the vendor. It is not a vendor gap in the
	// second case, and it is not a failure of the tested configuration in either.
	Unsupported     bool          `json:"unsupported,omitempty"`
	ContentFiltered bool          `json:"contentFiltered,omitempty"`
	Error           error         `json:"error"`
	StopReason      string        `json:"stopReason,omitempty"`
	Streaming       bool          `json:"streaming"`
	Reasoning       bool          `json:"reasoning"`
	Latency         time.Duration `json:"latency"`
}
