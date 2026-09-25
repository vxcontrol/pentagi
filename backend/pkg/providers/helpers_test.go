package providers

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

const (
	// helpersFallback is what a tool call nobody answered is padded with.
	helpersFallback = "the call was not handled, please try again"
	helpersWeather  = "The weather in New York is sunny with a high of 75°F."
	// helpersNotFound is the start of the executor's answer to a call of a tool that does not exist.
	helpersNotFound = "function 'execute_task_and_return_summary' not found in available tools list"
)

func helpersText(role llms.ChatMessageType, text ...string) llms.MessageContent {
	msg := llms.MessageContent{Role: role}
	for _, part := range text {
		msg.Parts = append(msg.Parts, llms.TextContent{Text: part})
	}

	return msg
}

func helpersToolCall(id, name, args string) llms.ToolCall {
	return llms.ToolCall{ID: id, Type: "function", FunctionCall: &llms.FunctionCall{Name: name, Arguments: args}}
}

func helpersToolResult(id, name, content string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: id, Name: name, Content: content}},
	}
}

// helpersChain builds a fresh chain per call, so a row that rewrites it cannot leak into another.
func helpersChain(kind string, tail ...llms.MessageContent) []llms.MessageContent {
	system := helpersText(llms.ChatMessageTypeSystem, "You are a helpful assistant.")
	weatherCall := helpersToolCall("tool-1", "get_weather", `{"location": "New York"}`)
	timeCall := helpersToolCall("tool-2", "get_time", `{"location": "New York"}`)

	var chain []llms.MessageContent
	switch kind {
	case "basic":
		chain = []llms.MessageContent{
			system,
			helpersText(llms.ChatMessageTypeHuman, "Hello, how are you?"),
			helpersText(llms.ChatMessageTypeAI, "I'm doing well! How can I help you today?"),
		}
	case "one call":
		chain = []llms.MessageContent{
			system,
			helpersText(llms.ChatMessageTypeHuman, "What's the weather like?"),
			{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{weatherCall}},
		}
	case "two calls":
		chain = []llms.MessageContent{
			system,
			helpersText(llms.ChatMessageTypeHuman, "What's the weather and time in New York?"),
			{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{weatherCall, timeCall}},
		}
	default:
		panic("unknown chain " + kind)
	}

	return append(chain, tail...)
}

func TestHelpers_UpdateMsgChainResult_PutsTheResultWhereTheChainExpectsIt(t *testing.T) {
	fp := newFlowProvider()

	const (
		help  = "I need help with my code."
		clock = "The current time in New York is 3:45 PM."
	)
	system := helpersText(llms.ChatMessageTypeSystem, "You are a helpful assistant.")

	tests := []struct {
		name     string
		chain    []llms.MessageContent
		toolName string
		input    string
		want     []llms.MessageContent
	}{
		{
			name:     "an empty chain becomes the human message",
			chain:    []llms.MessageContent{},
			toolName: "ask_user",
			input:    "Hello!",
			want:     []llms.MessageContent{helpersText(llms.ChatMessageTypeHuman, "Hello!")},
		},
		{
			name:     "a lone system message is followed by the human message",
			chain:    helpersChain("basic")[:1],
			toolName: "ask_user",
			input:    "Hello!",
			want:     []llms.MessageContent{system, helpersText(llms.ChatMessageTypeHuman, "Hello!")},
		},
		{
			name:     "a trailing human message takes the result as another part",
			chain:    helpersChain("basic")[:2],
			toolName: "ask_user",
			input:    " " + help,
			want: []llms.MessageContent{
				system,
				helpersText(llms.ChatMessageTypeHuman, "Hello, how are you?", " "+help),
			},
		},
		{
			name:     "a trailing answer is followed by the human message",
			chain:    helpersChain("basic"),
			toolName: "ask_user",
			input:    help,
			want:     helpersChain("basic", helpersText(llms.ChatMessageTypeHuman, help)),
		},
		{
			name:     "the pending call of the named tool gets the result",
			chain:    helpersChain("one call"),
			toolName: "get_weather",
			input:    helpersWeather,
			want:     helpersChain("one call", helpersToolResult("tool-1", "get_weather", helpersWeather)),
		},
		{
			name:     "a result for another tool pads the pending call and follows as a human message",
			chain:    helpersChain("one call"),
			toolName: "wrong_tool",
			input:    "This is a response to a wrong tool.",
			want: helpersChain("one call",
				helpersToolResult("tool-1", "get_weather", helpersFallback),
				helpersText(llms.ChatMessageTypeHuman, "This is a response to a wrong tool."),
			),
		},
		{
			name:     "an answered call has its result replaced",
			chain:    helpersChain("one call", helpersToolResult("tool-1", "get_weather", helpersWeather)),
			toolName: "get_weather",
			input:    "Updated: The weather in New York is rainy.",
			want: helpersChain("one call",
				helpersToolResult("tool-1", "get_weather", "Updated: The weather in New York is rainy.")),
		},
		{
			name:     "the first of two pending calls gets the result and the second is padded",
			chain:    helpersChain("two calls"),
			toolName: "get_weather",
			input:    helpersWeather,
			want: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersWeather),
				helpersToolResult("tool-2", "get_time", helpersFallback),
			),
		},
		{
			name:     "the second of two pending calls gets the result and the first is padded",
			chain:    helpersChain("two calls"),
			toolName: "get_time",
			input:    clock,
			want: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersFallback),
				helpersToolResult("tool-2", "get_time", clock),
			),
		},
		{
			name: "a result for no pending call follows the padded calls as a human message",
			chain: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersWeather)),
			toolName: "ask_user",
			input:    "I want to know more about the weather there.",
			want: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersWeather),
				helpersToolResult("tool-2", "get_time", helpersFallback),
				helpersText(llms.ChatMessageTypeHuman, "I want to know more about the weather there."),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fp.updateMsgChainResult(tt.chain, tt.toolName, tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHelpers_EnsureChainConsistency_AnswersEveryPendingToolCall(t *testing.T) {
	fp := newFlowProvider()

	tests := []struct {
		name  string
		chain []llms.MessageContent
		want  []llms.MessageContent
	}{
		{
			name:  "an empty chain stays empty",
			chain: []llms.MessageContent{},
			want:  []llms.MessageContent{},
		},
		{
			name:  "a chain with no pending call is left alone",
			chain: helpersChain("basic"),
			want:  helpersChain("basic"),
		},
		{
			name:  "a pending call is padded",
			chain: helpersChain("one call"),
			want:  helpersChain("one call", helpersToolResult("tool-1", "get_weather", helpersFallback)),
		},
		{
			name:  "every pending call is padded in order",
			chain: helpersChain("two calls"),
			want: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersFallback),
				helpersToolResult("tool-2", "get_time", helpersFallback),
			),
		},
		{
			name:  "only the call still pending is padded",
			chain: helpersChain("two calls", helpersToolResult("tool-1", "get_weather", helpersWeather)),
			want: helpersChain("two calls",
				helpersToolResult("tool-1", "get_weather", helpersWeather),
				helpersToolResult("tool-2", "get_time", helpersFallback),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fp.ensureChainConsistency(tt.chain)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// A row's cached answer is set after its first call, the way execToolCall stores what the tool said.
func TestHelpers_RepeatingDetector_FlagsTheThirdIdenticalCallInARow(t *testing.T) {
	search := helpersToolCall("test-id", "search", `{"query":"test"}`)

	tests := []struct {
		name       string
		calls      []llms.ToolCall
		cached     string
		want       []bool
		wantLen    int
		wantCached string
	}{
		{
			name:    "a call without a function is ignored",
			calls:   []llms.ToolCall{{ID: "test", Type: "function"}},
			want:    []bool{false},
			wantLen: 0,
		},
		{
			name:    "the first call is not a repeat",
			calls:   []llms.ToolCall{search},
			want:    []bool{false},
			wantLen: 1,
		},
		{
			name:       "two identical calls stay below the threshold and keep the cached answer",
			calls:      []llms.ToolCall{search, search},
			cached:     helpersNotFound,
			want:       []bool{false, false},
			wantLen:    2,
			wantCached: helpersNotFound,
		},
		{
			name:    "the third identical call is a repeat",
			calls:   []llms.ToolCall{search, search, search},
			want:    []bool{false, false, true},
			wantLen: 3,
		},
		{
			name:    "every identical call past the threshold is a repeat and counted",
			calls:   []llms.ToolCall{search, search, search, search, search, search, search},
			want:    []bool{false, false, true, true, true, true, true},
			wantLen: 7,
		},
		{
			name: "another tool with the same arguments starts a new run and clears the cached answer",
			calls: []llms.ToolCall{
				search, search, helpersToolCall("test-id", "browse", `{"query":"test"}`),
			},
			cached:  helpersNotFound,
			want:    []bool{false, false, false},
			wantLen: 1,
		},
		{
			name: "other arguments start a new run and clear the cached answer",
			calls: []llms.ToolCall{
				helpersToolCall("test-id", "search", `{"query":"test","limit":"10"}`),
				helpersToolCall("test-id", "search", `{"query":"test","limit":"10"}`),
				helpersToolCall("test-id", "search", `{"query":"different","limit":"20"}`),
			},
			cached:  helpersNotFound,
			want:    []bool{false, false, false},
			wantLen: 1,
		},
		{
			name: "the message field does not tell calls apart",
			calls: []llms.ToolCall{
				helpersToolCall("test-id", "search", `{"query":"test","message":"first attempt"}`),
				helpersToolCall("test-id", "search", `{"query":"test","message":"second attempt"}`),
				helpersToolCall("test-id", "search", `{"query":"test","message":"third attempt"}`),
			},
			want:    []bool{false, false, true},
			wantLen: 3,
		},
		{
			// Five keys, so unsorted map order would almost never make three calls agree.
			name: "key order does not tell calls apart",
			calls: []llms.ToolCall{
				helpersToolCall("test-id", "search", `{"a":"1","b":"2","c":"3","d":"4","e":"5"}`),
				helpersToolCall("test-id", "search", `{"e":"5","d":"4","c":"3","b":"2","a":"1"}`),
				helpersToolCall("test-id", "search", `{"c":"3","a":"1","e":"5","b":"2","d":"4"}`),
			},
			want:    []bool{false, false, true},
			wantLen: 3,
		},
		{
			name: "arguments that are not JSON are compared as sent",
			calls: []llms.ToolCall{
				helpersToolCall("test-id", "search", "not valid json"),
				helpersToolCall("test-id", "search", "not valid json"),
				helpersToolCall("test-id", "search", "other text"),
			},
			want:    []bool{false, false, false},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &repeatingDetector{}

			for i, call := range tt.calls {
				assert.Equal(t, tt.want[i], detector.detect(call), "call %d", i)
				if i == 0 && tt.cached != "" {
					detector.lastResponse = tt.cached
				}
			}

			assert.Len(t, detector.funcCalls, tt.wantLen)
			assert.Equal(t, tt.wantCached, detector.lastResponse)
		})
	}
}

func TestHelpers_RepeatingToolResponse_QuotesWhatTheToolAnswered(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 8192)

	tests := []struct {
		name     string
		funcName string
		repeats  int
		previous string
		want     string
		contains []string
		excludes []string
		quoted   int
	}{
		{
			name:     "the answer is quoted with the count and a remedy",
			funcName: "execute_task_and_return_summary",
			repeats:  4,
			previous: helpersNotFound,
			contains: []string{helpersNotFound, "4 times", "Alter the arguments"},
		},
		{
			name:     "with nothing to quote the notice names the tool",
			funcName: "web_search",
			repeats:  3,
			want:     "tool call 'web_search' is repeating, please try another tool",
		},
		{
			name:     "a long answer is quoted up to 4096 bytes",
			funcName: "terminal",
			repeats:  3,
			previous: long,
			excludes: []string{long},
			quoted:   4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := repeatingToolResponse(tt.funcName, tt.repeats, tt.previous)

			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}
			for _, banned := range tt.excludes {
				assert.NotContains(t, got, banned)
			}
			if tt.quoted > 0 {
				assert.Equal(t, tt.quoted, strings.Count(got, "x"), "bytes of the answer quoted back")
			}
		})
	}
}

func TestHelpers_ShouldInvokeMentor_FiresAtTheSameToolOrTheTotalThreshold(t *testing.T) {
	tests := []struct {
		name        string
		disabled    bool
		same, total int
		calls       []string
		want        []bool
	}{
		{
			name:  "the same tool reaching its limit",
			same:  5,
			total: 10,
			calls: []string{"tool1", "tool1", "tool1", "tool1", "tool1"},
			want:  []bool{false, false, false, false, true},
		},
		{
			name:  "distinct tools reaching the total limit",
			same:  5,
			total: 10,
			calls: []string{"tool1", "tool2", "tool3", "tool4", "tool5", "tool6", "tool7", "tool8", "tool9", "tool10"},
			want:  []bool{false, false, false, false, false, false, false, false, false, true},
		},
		{
			name:  "another tool restarts the same-tool count",
			same:  3,
			total: 10,
			calls: []string{"tool1", "tool1", "tool2", "tool1", "tool1"},
			want:  []bool{false, false, false, false, false},
		},
		{
			name:  "interleaved tools reaching the total limit",
			same:  5,
			total: 10,
			calls: []string{"tool1", "tool2", "tool1", "tool3", "tool1", "tool2", "tool3", "tool4", "tool5", "tool6"},
			want:  []bool{false, false, false, false, false, false, false, false, false, true},
		},
		{
			name:     "a disabled monitor never fires",
			disabled: true,
			same:     5,
			total:    10,
			calls:    []string{"tool1", "tool1", "tool1", "tool1", "tool1", "tool1"},
			want:     []bool{false, false, false, false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &executionMonitor{enabled: !tt.disabled, sameThreshold: tt.same, totalThreshold: tt.total}

			for i, call := range tt.calls {
				assert.Equal(t, tt.want[i], monitor.shouldInvokeMentor(helpersToolCall("", call, "")), "call %d (%s)", i, call)
			}
		})
	}
}

func TestHelpers_Reset_StartsTheCountsAgain(t *testing.T) {
	monitor := &executionMonitor{enabled: true, sameThreshold: 3, totalThreshold: 5}
	search := helpersToolCall("", "search", "")

	monitor.shouldInvokeMentor(search)
	monitor.shouldInvokeMentor(search)
	require.True(t, monitor.shouldInvokeMentor(search), "the third identical call reaches the threshold")

	monitor.reset()

	assert.Zero(t, monitor.sameToolCount)
	assert.Zero(t, monitor.totalCallCount)
	assert.Empty(t, monitor.lastToolName)
	assert.False(t, monitor.shouldInvokeMentor(search), "the first call after a reset must not fire")
}

func TestHelpers_WrapToolCallIDTemplateError_TellsTheOperatorTheModelCannotCallTools(t *testing.T) {
	t.Parallel()

	// Wrapped as the tool call ID sampling in provider/agents.go wraps an Ollama refusal.
	ollamaErr := fmt.Errorf(
		"all sample collection attempts failed: %w",
		fmt.Errorf("failed to call LLM: %w",
			errors.New("400 Bad Request: registry.ollama.ai/library/gemma3:27b-it-q4_K_M does not support tools")),
	)
	genericErr := errors.New("connection refused")

	cases := []struct {
		name           string
		input          error
		wantNil        bool
		wantContains   []string
		wantUnwrapsTo  error
		wantNotContain []string
	}{
		{
			name:    "no error passes through",
			input:   nil,
			wantNil: true,
		},
		{
			name:  "a model without tools gets an actionable hint",
			input: ollamaErr,
			wantContains: []string{
				"failed to determine tool call ID template",
				"does not support tool/function calling",
				"PentAGI requires",
				"flows and assistant sessions",
				"metadata advertises tool/function calling",
				"does not support tools",
			},
			wantNotContain: []string{
				// PentAGI does not verify Ollama capabilities, so no model tag is named as tool-capable.
				"llama3.1",
				"qwen2.5",
				"mistral-nemo",
				// The flow and the assistant paths share the helper.
				"flow execution",
				"flow creation",
			},
			wantUnwrapsTo: ollamaErr,
		},
		{
			name:  "any other error keeps the plain wrap",
			input: genericErr,
			wantContains: []string{
				"failed to determine tool call ID template",
				"connection refused",
			},
			wantNotContain: []string{
				"does not support tool/function calling",
				"metadata advertises tool/function calling",
			},
			wantUnwrapsTo: genericErr,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := wrapToolCallIDTemplateError(tc.input)
			if tc.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)

			msg := got.Error()
			for _, want := range tc.wantContains {
				assert.Contains(t, msg, want)
			}
			for _, banned := range tc.wantNotContain {
				assert.NotContains(t, msg, banned)
			}
			assert.ErrorIs(t, got, tc.wantUnwrapsTo)
		})
	}
}

func TestHelpers_ToolCallIDTemplateOrDefault_NeverYieldsAnEmptyTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		detected string
		want     string
	}{
		{name: "an empty detection falls back to the default", detected: "", want: "call_{r:24:x}"},
		{name: "a detected pattern is kept", detected: "toolu_{r:24:b}", want: "toolu_{r:24:b}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toolCallIDTemplateOrDefault(tt.detected)

			assert.Equal(t, tt.want, got)
			assert.NotEmpty(t, templates.GenerateFromPattern(got, tools.AdviceToolName), "the template must yield an ID")
		})
	}
}
