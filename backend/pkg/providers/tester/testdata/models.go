package testdata

import (
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type TestType string

const (
	TestTypeCompletion TestType = "completion"
	TestTypeJSON       TestType = "json"
	TestTypeTool       TestType = "tool"
)

type TestGroup string

const (
	TestGroupBasic     TestGroup = "basic"
	TestGroupAdvanced  TestGroup = "advanced"
	TestGroupJSON      TestGroup = "json"
	TestGroupKnowledge TestGroup = "knowledge"
)

// TestDefinition represents immutable test configuration from YAML
type TestDefinition struct {
	ID            string        `yaml:"id"`
	Name          string        `yaml:"name"`
	Type          TestType      `yaml:"type"`
	Group         TestGroup     `yaml:"group"`
	Prompt        string        `yaml:"prompt,omitempty"`
	Messages      []MessageData `yaml:"messages,omitempty"`
	Tools         []ToolData    `yaml:"tools,omitempty"`
	Expected      interface{}   `yaml:"expected"`
	Streaming     bool          `yaml:"streaming"`
	IsArray       bool          `yaml:"is_array,omitempty"`       // for JSON array tests
	ExpectFailure bool          `yaml:"expect_failure,omitempty"` // for negative tests
}

type MessageData struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

type ToolData struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Parameters  interface{} `yaml:"parameters"`
}

type ExpectedToolCall struct {
	FunctionName string                 `yaml:"function_name"`
	Arguments    map[string]interface{} `yaml:"arguments"`
}

// TestCase represents a stateful test execution instance
type TestCase interface {
	ID() string
	Name() string
	Type() TestType
	Group() TestGroup
	Streaming() bool

	// LLM execution data
	Prompt() string
	Messages() []llms.MessageContent
	Tools() []llms.Tool
	StreamingCallback() streaming.Callback

	// result validation and state management
	Execute(response interface{}, latency time.Duration) TestResult
}

// TestSuite contains stateful test cases for execution
type TestSuite struct {
	Group TestGroup
	Tests []TestCase
}
