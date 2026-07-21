package testdata

import "time"

// TestResult represents the result of a single test execution
type TestResult struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       TestType       `json:"type"`
	Group      TestGroup      `json:"group"`
	Capability TestCapability `json:"capability,omitempty"`
	Success    bool           `json:"success"`
	// Unsupported reports that the provider/model proactively rejected the
	// requested capability with a typed SDK error (ErrStructuredOutputUnsupported,
	// ErrStructuredOutputConflict, reasoning.ErrReasoningOffUnsupported) rather
	// than the call failing for an unrelated reason. It is distinct from a hard
	// failure: the config is fine, this particular optional capability simply
	// isn't available on this model, which is itself a useful production
	// readiness signal rather than a bug in the tested configuration.
	Unsupported bool          `json:"unsupported,omitempty"`
	Error       error         `json:"error"`
	Streaming   bool          `json:"streaming"`
	Reasoning   bool          `json:"reasoning"`
	Latency     time.Duration `json:"latency"`
}
