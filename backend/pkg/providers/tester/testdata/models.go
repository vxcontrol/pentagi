package testdata

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type TestType string

const (
	TestTypeCompletion TestType = "completion"
	TestTypeJSON       TestType = "json"
	TestTypeTool       TestType = "tool"
	// TestTypeFileEdit is a MultiTurnTestCase: it isn't built from tests.yml
	// (see tester.newFileEditTestCase), only used to label its TestResult.
	TestTypeFileEdit TestType = "file_edit"
)

type TestGroup string

const (
	TestGroupBasic     TestGroup = "basic"
	TestGroupAdvanced  TestGroup = "advanced"
	TestGroupJSON      TestGroup = "json"
	TestGroupKnowledge TestGroup = "knowledge"
)

// TestCapability names a wire capability a test case exercises. A test with a
// non-empty Capability only runs against an agent whose ACTUAL loaded config
// would make this exact call in a real PentAGI flow (see
// tester.capabilitySupported, which mirrors pconfig.AgentConfig.BuildOptions /
// ProviderConfig.UsesAdaptiveThinking exactly) — ctester answers "will this
// configuration work inside PentAGI", so testing a CallOption combination the
// app's runtime would never actually send for this config would produce a
// pass or fail that says nothing about a flow that will ever really happen.
type TestCapability string

const (
	// CapabilityNone is the default: the test runs unconditionally for every
	// agent/model, exactly like the pre-existing basic/advanced/knowledge tests.
	CapabilityNone TestCapability = ""
	// CapabilityAdaptiveThinking exercises llms.WithAdaptiveReasoning, gated to
	// agents whose config actually triggers it in production: explicit
	// `reasoning: {mode: adaptive}`, or an adaptive-only model (Claude Opus
	// 4.7/4.8), which PrepareAdaptiveCallOptions forces on regardless of config.
	CapabilityAdaptiveThinking TestCapability = "adaptive_thinking"
	// CapabilityReasoningOff exercises llms.WithReasoningDisabled, gated to
	// agents whose config explicitly sets `reasoning: {mode: off}` — the only
	// case where AgentConfig.BuildOptions actually emits it.
	CapabilityReasoningOff TestCapability = "reasoning_off"
	// CapabilityStructuredOutput exercises llms.WithStructuredOutput with the
	// test's Schema. Deliberately run ahead of AgentConfig.BuildOptions
	// actually wiring it in: that integration is landing soon for the
	// simple_json agent, so this validates the LLM/backend can honor
	// schema-constrained output before the real call path depends on it. Only
	// ever scheduled for OptionsTypeSimpleJSON (see isTestCompatibleWithAgent).
	CapabilityStructuredOutput TestCapability = "structured_output"
)

// MessagesData represents a collection of message data with conversion capabilities
type MessagesData []MessageData

// ToMessageContent converts MessagesData to llms.MessageContent array with tool call support
func (md MessagesData) ToMessageContent() ([]llms.MessageContent, error) {
	var messages []llms.MessageContent

	for _, msg := range md {
		var msgType llms.ChatMessageType
		switch strings.ToLower(msg.Role) {
		case "system":
			msgType = llms.ChatMessageTypeSystem
		case "user", "human":
			msgType = llms.ChatMessageTypeHuman
		case "assistant", "ai":
			msgType = llms.ChatMessageTypeAI
		case "tool":
			msgType = llms.ChatMessageTypeTool
		default:
			return nil, fmt.Errorf("unknown message role: %s", msg.Role)
		}

		if msgType == llms.ChatMessageTypeTool {
			// tool response message
			messages = append(messages, llms.MessageContent{
				Role: msgType,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: msg.ToolCallID,
						Name:       msg.Name,
						Content:    msg.Content,
					},
				},
			})
		} else if len(msg.ToolCalls) > 0 {
			// assistant message with tool calls
			var parts []llms.ContentPart
			if msg.Content != "" {
				parts = append(parts, llms.TextContent{Text: msg.Content})
			}

			for _, tc := range msg.ToolCalls {
				argsBytes, err := json.Marshal(tc.Function.Arguments)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call arguments: %v", err)
				}

				parts = append(parts, llms.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					FunctionCall: &llms.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: string(argsBytes),
					},
				})
			}

			messages = append(messages, llms.MessageContent{
				Role:  msgType,
				Parts: parts,
			})
		} else {
			// regular text message
			messages = append(messages, llms.TextParts(msgType, msg.Content))
		}
	}

	return messages, nil
}

// TestDefinition represents immutable test configuration from YAML
type TestDefinition struct {
	ID         string         `yaml:"id"`
	Name       string         `yaml:"name"`
	Type       TestType       `yaml:"type"`
	Group      TestGroup      `yaml:"group"`
	Prompt     string         `yaml:"prompt,omitempty"`
	Messages   MessagesData   `yaml:"messages,omitempty"`
	Tools      []ToolData     `yaml:"tools,omitempty"`
	Expected   any            `yaml:"expected"`
	Streaming  bool           `yaml:"streaming"`
	Capability TestCapability `yaml:"capability,omitempty"`

	// RequireReasoning, when non-nil, asserts the presence (true) or absence
	// (false) of reasoning content in the response — used by
	// CapabilityAdaptiveThinking (must reason) and CapabilityReasoningOff
	// (must NOT reason) tests to verify the capability actually took effect on
	// the wire, not just that the call happened not to error.
	RequireReasoning *bool `yaml:"require_reasoning,omitempty"`

	// Schema and SchemaName configure a CapabilityStructuredOutput json test:
	// Schema is the raw JSON Schema document (any YAML-decoded value, marshaled
	// to JSON when building the test case) passed verbatim to
	// llms.WithStructuredOutput.
	Schema     any    `yaml:"schema,omitempty"`
	SchemaName string `yaml:"schema_name,omitempty"`
}

type MessageData struct {
	Role       string         `yaml:"role"`
	Content    string         `yaml:"content"`
	ToolCalls  []ToolCallData `yaml:"tool_calls,omitempty"`
	ToolCallID string         `yaml:"tool_call_id,omitempty"`
	Name       string         `yaml:"name,omitempty"`
}

type ToolCallData struct {
	ID       string           `yaml:"id"`
	Type     string           `yaml:"type"`
	Function FunctionCallData `yaml:"function"`
}

type FunctionCallData struct {
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments"`
}

type ToolData struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Parameters  any    `yaml:"parameters"`
}

type ExpectedToolCall struct {
	FunctionName string         `yaml:"function_name"`
	Arguments    map[string]any `yaml:"arguments"`
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

	// Capability reports the wire behavior this test is meant to exercise
	// (CapabilityNone for every pre-existing test type). The runner uses it
	// to gate whether this agent's ACTUAL config would make this exact call
	// in production before spending a request on it (see
	// tester.capabilitySupported): Capability is informational/gating only,
	// it does not by itself force anything onto the wire.
	Capability() TestCapability
	// ExtraOptions returns extra llms.CallOption values to layer on top of
	// the agent's configured options. Nil for CapabilityAdaptiveThinking/
	// CapabilityReasoningOff: those are gated to configs that already produce
	// that behavior via the plain CallWithTools/CallEx path, so forcing it
	// here would be redundant at best and, for a config that doesn't
	// naturally produce it, would test a call PentAGI's runtime never makes.
	// Non-nil only for CapabilityStructuredOutput, which forces
	// llms.WithStructuredOutput ahead of AgentConfig.BuildOptions wiring it in
	// for real — see CapabilityStructuredOutput's doc comment.
	ExtraOptions() []llms.CallOption

	// result validation and state management
	Execute(response any, latency time.Duration) TestResult
}

// MultiTurnTestCase is an optional extension of TestCase for scenarios that
// need more than one round-trip to the provider before Execute can judge the
// outcome - e.g. a tool call whose result must be answered before the model
// makes its next call. The runner (tester.executeTest) detects it via a type
// assertion; every TestCase that doesn't implement it keeps going through
// the plain single-call path unmodified.
type MultiTurnTestCase interface {
	TestCase

	// HandleToolResponse receives the latest provider response. If it
	// recognizes something it needs to answer, it records the exchange
	// internally (so the next Messages() call reflects it) and returns
	// true, asking the runner for another round. Returning false ends the
	// exchange: the runner calls Execute with this same response, exactly
	// as it would for a plain TestCase.
	HandleToolResponse(resp *llms.ContentResponse) bool
}

// TestSuite contains stateful test cases for execution
type TestSuite struct {
	Group TestGroup
	Tests []TestCase
}
