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
)

type TestGroup string

const (
	TestGroupBasic     TestGroup = "basic"
	TestGroupAdvanced  TestGroup = "advanced"
	TestGroupJSON      TestGroup = "json"
	TestGroupKnowledge TestGroup = "knowledge"
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
	ID        string       `yaml:"id"`
	Name      string       `yaml:"name"`
	Type      TestType     `yaml:"type"`
	Group     TestGroup    `yaml:"group"`
	Prompt    string       `yaml:"prompt,omitempty"`
	Messages  MessagesData `yaml:"messages,omitempty"`
	Tools     []ToolData   `yaml:"tools,omitempty"`
	Expected  any          `yaml:"expected"`
	Streaming bool         `yaml:"streaming"`
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

	// result validation and state management
	Execute(response any, latency time.Duration) TestResult
}

// TestSuite contains stateful test cases for execution
type TestSuite struct {
	Group TestGroup
	Tests []TestCase
}
