package main

import (
	"time"

	"pentagi/pkg/tools"

	"github.com/tmc/langchaingo/llms"
)

// SimpleCompletionTest represents a test for simple text completion
type SimpleCompletionTest struct {
	Prompt   string
	Expected string
}

// SystemUserPromptTest represents a test with separate system and user prompts
type SystemUserPromptTest struct {
	SystemPrompt string
	UserPrompt   string
	Expected     string
}

// JSONCompletionTest represents a test for structured JSON response
type JSONCompletionTest struct {
	Prompt         string
	Schema         map[string]interface{}
	RequiredFields []string
}

// SimpleJSONCompletionTest represents a test for simple JSON response without function calling
type SimpleJSONCompletionTest struct {
	Prompt         string
	RequiredFields []string
	IsArray        bool // Indicates if the response should be a JSON array
	ExpectFailure  bool // For negative testing scenarios
}

// FunctionCallTest represents a test for function calling capability
type FunctionCallTest struct {
	Name         string
	Prompt       string
	Functions    []llms.Tool
	ExpectedTool string
	ValidateFunc func(args map[string]interface{}) bool
}

// TestFileAction represents a structure for testing file operations
type TestFileAction struct {
	Action  tools.CodeAction `json:"action"`
	Path    string           `json:"path"`
	Content string           `json:"content"`
	Message string           `json:"message"`
}

// TestSearchAction represents a structure for testing search operations
type TestSearchAction struct {
	Query      string      `json:"query"`
	MaxResults tools.Int64 `json:"max_results"`
	Message    string      `json:"message"`
}

// TestAskAdvice represents a structure for testing mentor advice requests
type TestAskAdvice struct {
	Question string `json:"question"`
	Code     string `json:"code"`
	Output   string `json:"output"`
	Message  string `json:"message"`
}

// TestResult represents the result of a single test
type TestResult struct {
	Name      string
	Type      string
	Success   bool
	Error     error
	LatencyMs int64
	Response  string
	Expected  string
}

// AgentTestResult collects test results for each agent
type AgentTestResult struct {
	AgentType       string
	ModelName       string
	BasicTests      []TestResult
	AdvancedTests   []TestResult
	TotalSuccess    int
	TotalTests      int
	AverageLatency  time.Duration
	SkippedAdvanced bool
	SkippedReason   string
}

// TestSuite represents a complete test suite
type TestSuite struct {
	Name  string
	Basic bool
	Tests []interface{}
}

// TestConfig holds the configuration for test execution
type TestConfig struct {
	AgentTypes  []string
	TestGroups  []string
	ConfigPath  string
	ReportPath  string
	VerboseMode bool
}
