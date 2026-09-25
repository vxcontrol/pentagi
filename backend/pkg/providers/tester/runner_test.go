package tester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/cases"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func testerRegistry(t *testing.T, yaml string) *cases.TestRegistry {
	t.Helper()

	registry, err := cases.LoadRegistryFromYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return registry
}

func testerRunOne(t *testing.T, prv provider.Provider, yaml string) cases.TestResult {
	t.Helper()

	results, err := TestProvider(
		t.Context(), prv,
		WithCustomRegistry(testerRegistry(t, yaml)),
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(cases.TestGroupBasic),
		WithParallelWorkers(1),
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results.Simple) != 1 {
		t.Fatalf("want one result, got %d", len(results.Simple))
	}
	return results.Simple[0]
}

func openAIShapedAnswer(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
		`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`)
}

func TestRunner_GroupResults_FilesEachAgentsResultsUnderItsOwnField(t *testing.T) {
	fields := func(r ProviderTestResults) map[pconfig.ProviderOptionsType]AgentTestResults {
		return map[pconfig.ProviderOptionsType]AgentTestResults{
			pconfig.OptionsTypeSimple:       r.Simple,
			pconfig.OptionsTypeSimpleJSON:   r.SimpleJSON,
			pconfig.OptionsTypePrimaryAgent: r.PrimaryAgent,
			pconfig.OptionsTypeAssistant:    r.Assistant,
			pconfig.OptionsTypeGenerator:    r.Generator,
			pconfig.OptionsTypeRefiner:      r.Refiner,
			pconfig.OptionsTypeAdviser:      r.Adviser,
			pconfig.OptionsTypeReflector:    r.Reflector,
			pconfig.OptionsTypeSearcher:     r.Searcher,
			pconfig.OptionsTypeEnricher:     r.Enricher,
			pconfig.OptionsTypeCoder:        r.Coder,
			pconfig.OptionsTypeInstaller:    r.Installer,
			pconfig.OptionsTypePentester:    r.Pentester,
		}
	}

	prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	prv.SetDefaultResponse("Mock response")

	run := func(t *testing.T, opts ...TestOption) map[pconfig.ProviderOptionsType]AgentTestResults {
		t.Helper()
		results, err := TestProvider(t.Context(), prv, append([]TestOption{WithStreamingMode(false)}, opts...)...)
		if err != nil {
			t.Fatalf("TestProvider failed: %v", err)
		}
		return fields(results)
	}

	t.Run("a run naming no agent fills every field", func(t *testing.T) {
		for agent, got := range run(t) {
			if len(got) == 0 {
				t.Errorf("%s has no results", agent)
			}
		}
	})

	for agent := range fields(ProviderTestResults{}) {
		t.Run("a run naming only "+string(agent), func(t *testing.T) {
			for field, got := range run(t, WithAgentTypes(agent), WithParallelWorkers(2)) {
				if filled := len(got) > 0; filled != (field == agent) {
					t.Errorf("%s holds %d results, want results only under %s", field, len(got), agent)
				}
			}
		})
	}
}

func TestRunner_CollectTestRequests_RunsStreamingCasesOnlyInStreamingMode(t *testing.T) {
	const yaml = `
- id: "plain"
  name: "Plain"
  type: "completion"
  group: "basic"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
  streaming: false
- id: "streamed"
  name: "Streamed"
  type: "completion"
  group: "basic"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
  streaming: true
`

	for _, tt := range []struct {
		name      string
		streaming bool
		want      []string
	}{
		{"streaming on runs both", true, []string{"plain", "streamed"}},
		{"streaming off drops the streamed case", false, []string{"plain"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
			prv.SetStreamingDelay(time.Millisecond * 5)

			results, err := TestProvider(
				t.Context(), prv,
				WithCustomRegistry(testerRegistry(t, yaml)),
				WithAgentTypes(pconfig.OptionsTypeSimple),
				WithGroups(cases.TestGroupBasic),
				WithStreamingMode(tt.streaming),
			)
			if err != nil {
				t.Fatalf("TestProvider streaming failed: %v", err)
			}

			var ran []string
			for _, result := range results.Simple {
				ran = append(ran, result.ID)
				if result.Streaming && result.Latency <= 0 {
					t.Errorf("Expected non-zero latency for streaming test")
				}
				if result.Streaming != (result.ID == "streamed") {
					t.Errorf("%s reports streaming = %v", result.ID, result.Streaming)
				}
			}
			slices.Sort(ran)
			if !slices.Equal(ran, tt.want) {
				t.Errorf("ran %v, want %v", ran, tt.want)
			}
		})
	}
}

func TestRunner_CollectTestRequests_RunsOnlyTheRequestedGroups(t *testing.T) {
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetDefaultResponse("Group test response")

	tests := []struct {
		name      string
		agentType pconfig.ProviderOptionsType
		groups    []cases.TestGroup
	}{
		{"basic only", pconfig.OptionsTypeSimple, []cases.TestGroup{cases.TestGroupBasic}},
		{"advanced only", pconfig.OptionsTypeSimple, []cases.TestGroup{cases.TestGroupAdvanced}},
		{"json only", pconfig.OptionsTypeSimpleJSON, []cases.TestGroup{cases.TestGroupJSON}},
		{"knowledge only", pconfig.OptionsTypeSimple, []cases.TestGroup{cases.TestGroupKnowledge}},
		{"basic and advanced", pconfig.OptionsTypeSimple, []cases.TestGroup{cases.TestGroupBasic, cases.TestGroupAdvanced}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := TestProvider(
				t.Context(),
				mockProvider,
				WithAgentTypes(tt.agentType),
				WithGroups(tt.groups...),
			)
			if err != nil {
				t.Fatalf("TestProvider groups failed: %v", err)
			}

			agentResults := results.Simple
			if tt.agentType == pconfig.OptionsTypeSimpleJSON {
				agentResults = results.SimpleJSON
			}
			if len(agentResults) == 0 {
				t.Fatalf("no results for %s", tt.agentType)
			}

			for _, result := range agentResults {
				if !slices.Contains(tt.groups, result.Group) {
					t.Errorf("Result belongs to unexpected group: %s", result.Group)
				}
			}
		})
	}
}

func TestRunner_CollectTestRequests_BuildsARegistryCaseAfreshForEachAgent(t *testing.T) {
	prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	prv.SetSequentialResponses(
		&llms.ContentResponse{Choices: []*llms.ContentChoice{{
			Content:   "ok",
			Reasoning: &reasoning.ContentReasoning{Content: "the first agent thought about it"},
		}}},
		&llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "ok"}}},
	)

	results, err := TestProvider(
		t.Context(), prv,
		WithCustomRegistry(testerRegistry(t, `
- id: "streamed"
  name: "Streamed"
  type: "completion"
  group: "basic"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
  streaming: true
`)),
		WithAgentTypes(pconfig.OptionsTypeSimple, pconfig.OptionsTypePrimaryAgent),
		WithGroups(cases.TestGroupBasic),
		WithParallelWorkers(1),
	)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(results.Simple) != 1 || len(results.PrimaryAgent) != 1 {
		t.Fatalf("want one result per agent, got simple=%d primary=%d", len(results.Simple), len(results.PrimaryAgent))
	}
	if !results.Simple[0].Reasoning {
		t.Fatalf("the first agent's answer carried reasoning, so this half proves nothing")
	}
	if results.PrimaryAgent[0].Reasoning {
		t.Error("the second agent reports reasoning its own answer never carried: the case instance is shared")
	}
}

func TestRunner_CollectTestRequests_BuildsTheFileEditCaseAfreshForEachAgent(t *testing.T) {
	emptyRegistry := testerRegistry(t, "[]")
	config := &testConfig{
		agentTypes:     []pconfig.ProviderOptionsType{pconfig.OptionsTypePrimaryAgent, pconfig.OptionsTypeCoder, pconfig.OptionsTypePentester},
		groups:         []cases.TestGroup{cases.TestGroupAdvanced},
		customRegistry: emptyRegistry,
	}

	requests, err := collectTestRequests(emptyRegistry, mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model"), config, "")
	if err != nil {
		t.Fatalf("collectTestRequests: %v", err)
	}

	if len(requests) != len(config.agentTypes) {
		t.Fatalf("expected %d file_edit requests (one per agent type), got %d", len(config.agentTypes), len(requests))
	}

	seen := make(map[cases.TestCase]pconfig.ProviderOptionsType, len(requests))
	for _, req := range requests {
		if req.testCase.Type() != cases.TestTypeFileEdit {
			t.Fatalf("expected only file_edit requests, got type %q", req.testCase.Type())
		}
		if owner, dup := seen[req.testCase]; dup {
			t.Fatalf("agent types %q and %q were given the SAME fileEditTestCase instance - state from one run would leak into the other", owner, req.agentType)
		}
		seen[req.testCase] = req.agentType

		if _, ok := req.testCase.(cases.MultiTurnTestCase); !ok {
			t.Fatalf("file_edit test case for %q does not implement MultiTurnTestCase", req.agentType)
		}
	}
}

func TestRunner_IsTestCompatibleWithAgent_RunsToolCasesOnlyOnAgentsWhoseCallsCarryTools(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		openAIShapedAnswer(w)
	}))
	t.Cleanup(srv.Close)

	providerConfig, err := openai.BuildProviderConfig([]byte(`
adviser: {model: gpt-5.6-sol, reasoning: {effort: xhigh}}
reflector: {model: gpt-5.6-sol, reasoning: {effort: high}}
coder: {model: gpt-5.6-sol}
`))
	require.NoError(t, err)
	prv, err := openai.New(&config.Config{OpenAIServerURL: srv.URL, OpenAIKey: "k"}, provider.DefaultProviderNameOpenAI, providerConfig)
	require.NoError(t, err)

	results, err := TestProvider(t.Context(), prv,
		WithAgentTypes(pconfig.OptionsTypeAdviser, pconfig.OptionsTypeReflector, pconfig.OptionsTypeCoder),
		WithStreamingMode(false),
	)
	require.NoError(t, err)

	for agent, got := range map[pconfig.ProviderOptionsType]AgentTestResults{
		pconfig.OptionsTypeAdviser:   results.Adviser,
		pconfig.OptionsTypeReflector: results.Reflector,
	} {
		require.NotEmpty(t, got, agent)
		for _, result := range got {
			assert.NotContains(t, []cases.TestType{cases.TestTypeTool, cases.TestTypeFileEdit}, result.Type,
				"%s ran %q", agent, result.Name)
			var refused *reasoning.ErrEffortWithTools
			assert.False(t, errors.As(result.Error, &refused), "%s: %q failed on tools: %v", agent, result.Name, result.Error)
		}
	}

	ranTools := false
	for _, result := range results.Coder {
		ranTools = ranTools || result.Type == cases.TestTypeTool
	}
	assert.True(t, ranTools, "the coder calls with tools, so its tool cases still run")
}

const (
	firstTurnText  = "first turn"
	secondTurnText = "second turn marker"
)

type growingTestCase struct {
	handled bool
	final   string
}

func (c *growingTestCase) ID() string                            { return "growing" }
func (c *growingTestCase) Name() string                          { return "Growing" }
func (c *growingTestCase) Type() cases.TestType                  { return cases.TestTypeTool }
func (c *growingTestCase) Group() cases.TestGroup                { return cases.TestGroupAdvanced }
func (c *growingTestCase) Streaming() bool                       { return false }
func (c *growingTestCase) Prompt() string                        { return "" }
func (c *growingTestCase) StreamingCallback() streaming.Callback { return nil }
func (c *growingTestCase) Capability() cases.TestCapability      { return cases.CapabilityNone }
func (c *growingTestCase) ExtraOptions() []llms.CallOption {
	return []llms.CallOption{llms.WithTemperature(0.25)}
}
func (c *growingTestCase) ExpectRefusal() cases.RefusalKind { return cases.RefusalNone }
func (c *growingTestCase) ExpectTruncated() bool            { return false }

func (c *growingTestCase) Tools() []llms.Tool {
	return []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "noop"}}}
}

func (c *growingTestCase) Messages() []llms.MessageContent {
	text := firstTurnText
	if c.handled {
		text = secondTurnText
	}
	return []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, text)}
}

func (c *growingTestCase) HandleToolResponse(*llms.ContentResponse) bool {
	if c.handled {
		return false
	}
	c.handled = true
	return true
}

func (c *growingTestCase) Execute(response any, latency time.Duration) cases.TestResult {
	if resp, ok := response.(*llms.ContentResponse); ok && len(resp.Choices) > 0 {
		c.final = resp.Choices[0].Content
	}
	return cases.TestResult{ID: c.ID(), Success: true, Latency: latency}
}

func TestRunner_ExecuteTest_ResendsTheConversationAsItGrows(t *testing.T) {
	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetResponses([]mock.ResponseConfig{
		{Key: firstTurnText, Response: &llms.ContentResponse{Choices: []*llms.ContentChoice{{
			ToolCalls: []llms.ToolCall{{
				ID:           "call_1",
				Type:         "function",
				FunctionCall: &llms.FunctionCall{Name: "noop", Arguments: "{}"},
			}},
		}}}},
		{Key: secondTurnText, Response: "done"},
	})

	testCase := &growingTestCase{}
	result, err := executeTest(t.Context(), testRequest{
		agentType: pconfig.OptionsTypeSimple,
		testCase:  testCase,
		provider:  mockProvider,
	})
	if err != nil {
		t.Fatalf("executeTest: %v", err)
	}
	if !result.Success {
		t.Fatalf("case failed: %v", result.Error)
	}

	if !testCase.handled {
		t.Fatal("the tool call was never answered, so the second turn never happened")
	}
	if testCase.final != "done" {
		t.Errorf("final answer = %q, want %q: the second call did not carry the conversation as it stood by then",
			testCase.final, "done")
	}
}

type chainRecorder struct {
	*mock.Provider

	template string

	mu     sync.Mutex
	chains [][]llms.MessageContent
}

func (r *chainRecorder) GetToolCallIDTemplate(context.Context, templates.Prompter) (string, error) {
	return r.template, nil
}

func (r *chainRecorder) record(chain []llms.MessageContent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chains = append(r.chains, chain)
}

func (r *chainRecorder) CallEx(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent, streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	r.record(chain)
	return r.Provider.CallEx(ctx, opt, chain, streamCb)
}

func (r *chainRecorder) CallWithTools(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent, offered []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	r.record(chain)
	return r.Provider.CallWithTools(ctx, opt, chain, offered, streamCb)
}

func (r *chainRecorder) CallWithExtraOptions(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent, offered []llms.Tool,
	streamCb streaming.Callback, extra ...llms.CallOption,
) (*llms.ContentResponse, error) {
	r.record(chain)
	return r.Provider.CallWithExtraOptions(ctx, opt, chain, offered, streamCb, extra...)
}

func TestRunner_ExecuteTest_CallsAMultiTurnCaseAgainUntilItStops(t *testing.T) {
	read := fileToolCallResponse("call_1", map[string]string{"action": "read_file", "path": FileEditTestPath})
	edit := fileToolCallResponse("call_2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff})
	write := fileToolCallResponse("call_1", map[string]string{"action": "write_file", "path": FileEditTestPath, "content": "oops"})

	for _, tt := range []struct {
		name          string
		responses     []any
		wantPass      bool
		wantCalls     int
		wantLastChain int
	}{
		{"read then edit takes a second call carrying the answered read", []any{read, edit}, true, 2, 3},
		{"a wrong first call ends the exchange without a second call", []any{write, edit}, false, 1, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tc, err := newFileEditTestCase()
			if err != nil {
				t.Fatalf("newFileEditTestCase() error = %v", err)
			}

			recorder := &chainRecorder{
				Provider: mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model"),
			}
			recorder.SetSequentialResponses(tt.responses...)

			result, err := executeTest(t.Context(), testRequest{
				agentType: pconfig.OptionsTypePrimaryAgent,
				testCase:  tc,
				provider:  recorder,
			})
			if err != nil {
				t.Fatalf("executeTest() error = %v", err)
			}
			if result.Success != tt.wantPass {
				t.Fatalf("success = %v, want %v (error: %v)", result.Success, tt.wantPass, result.Error)
			}
			if result.Type != cases.TestTypeFileEdit {
				t.Errorf("result.Type = %q, want %q", result.Type, cases.TestTypeFileEdit)
			}
			if len(recorder.chains) != tt.wantCalls {
				t.Fatalf("the provider was called %d times, want %d", len(recorder.chains), tt.wantCalls)
			}
			if got := len(recorder.chains[len(recorder.chains)-1]); got != tt.wantLastChain {
				t.Errorf("the last call carried %d messages, want %d", got, tt.wantLastChain)
			}
		})
	}
}

const filteredToolCaseYAML = `
- id: "scan"
  name: "Scan"
  type: "tool"
  group: "basic"
  messages:
    - role: "user"
      content: "Call nmap for 192.168.1.1"
  tools:
    - name: "nmap"
      description: "Scans a host"
      parameters:
        type: "object"
        properties:
          target:
            type: "string"
        required: ["target"]
  expected:
    - function_name: "nmap"
      arguments:
        target: "192.168.1.1"
  streaming: false
`

func TestRunner_ExecuteTest_MarksAVendorContentFilterApartFromABreakage(t *testing.T) {
	answer := func(stopReason, content string) *llms.ContentResponse {
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: content, StopReason: stopReason}}}
	}

	for _, tc := range []struct {
		label        string
		response     any
		wantPass     bool
		wantFiltered bool
	}{
		{
			label: "a filter the vendor reports as an error",
			response: errors.New("API returned unexpected status code: 400: litellm.BadRequestError: DashscopeException - " +
				"Input data may contain inappropriate content. For details, see: " +
				"https://www.alibabacloud.com/help/en/model-studio/error-code#inappropriate-content."),
			wantFiltered: true,
		},
		{label: "a filter the vendor reports as a stop reason", response: answer("content_filter", ""), wantFiltered: true},
		{label: "a refusal in plain text with a natural stop", response: answer("STOP", "Sorry, I cannot fulfill your request.")},
		{label: "a transport failure", response: errors.New("connection reset by peer")},
		{
			label: "a passing answer the vendor stopped with the filter reason",
			response: &llms.ContentResponse{Choices: []*llms.ContentChoice{{
				StopReason: "content_filter",
				ToolCalls: []llms.ToolCall{{
					ID:           "call-1",
					FunctionCall: &llms.FunctionCall{Name: "nmap", Arguments: `{"target": "192.168.1.1"}`},
				}},
			}}},
			wantPass: true,
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
			prv.SetSequentialResponses(tc.response)
			result := testerRunOne(t, prv, filteredToolCaseYAML)

			if result.Success != tc.wantPass {
				t.Fatalf("success = %v, want %v (error: %v)", result.Success, tc.wantPass, result.Error)
			}
			if result.ContentFiltered != tc.wantFiltered {
				t.Errorf("ContentFiltered = %v, want %v (error: %v)", result.ContentFiltered, tc.wantFiltered, result.Error)
			}
			if tc.wantPass {
				return
			}
			if said := strings.Contains(result.Error.Error(), "content filter"); said != tc.wantFiltered {
				t.Errorf("the failure names the content filter = %v, want %v: %v", said, tc.wantFiltered, result.Error)
			}
			if vendorErr, isErr := tc.response.(error); isErr && !errors.Is(result.Error, vendorErr) {
				t.Errorf("the result lost the vendor's error: %v", result.Error)
			}
		})
	}
}

func TestRunner_ExecuteTest_KeepsTheStopReasonOfAPassingAnswer(t *testing.T) {
	prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	prv.SetSequentialResponses(&llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "ok", StopReason: "stop"}}})

	result := testerRunOne(t, prv, `
- id: "plain"
  name: "Plain"
  type: "completion"
  group: "basic"
  messages:
    - role: "user"
      content: "hi"
  expected: "ok"
  streaming: false
`)
	if !result.Success {
		t.Fatalf("the case did not pass: %+v", result)
	}
	if result.StopReason != "stop" {
		t.Errorf("StopReason = %q, want %q: a passing result must still say why the model stopped", result.StopReason, "stop")
	}
}

type stubTestCase struct {
	want cases.RefusalKind
}

func (c stubTestCase) ID() string                            { return "stub" }
func (c stubTestCase) Name() string                          { return "Stub" }
func (c stubTestCase) Type() cases.TestType                  { return cases.TestTypeJSON }
func (c stubTestCase) Group() cases.TestGroup                { return cases.TestGroupJSON }
func (c stubTestCase) Streaming() bool                       { return false }
func (c stubTestCase) Prompt() string                        { return "" }
func (c stubTestCase) Messages() []llms.MessageContent       { return nil }
func (c stubTestCase) Tools() []llms.Tool                    { return nil }
func (c stubTestCase) StreamingCallback() streaming.Callback { return nil }
func (c stubTestCase) ExtraOptions() []llms.CallOption       { return nil }
func (c stubTestCase) ExpectRefusal() cases.RefusalKind      { return c.want }
func (c stubTestCase) ExpectTruncated() bool                 { return false }

func (c stubTestCase) Capability() cases.TestCapability {
	return cases.CapabilityStructuredOutput
}

func (c stubTestCase) Execute(any, time.Duration) cases.TestResult {
	panic("Execute must not run for a call judged as a refusal")
}

func TestRunner_JudgeRefusal_JudgesACallAgainstTheRefusalItExpects(t *testing.T) {
	for _, tt := range []struct {
		name            string
		want            cases.RefusalKind
		err             error
		wantSuccess     bool
		wantUnsupported bool
		wantSays        string
	}{
		{
			name:            "no expectation and reasoning that cannot be turned off",
			err:             &reasoning.ErrReasoningOffUnsupported{Model: "o3"},
			wantUnsupported: true,
		},
		{
			name: "no expectation and conflicting structured output",
			err: &llms.ErrStructuredOutputConflict{
				Provider: "openai",
				Detail:   "per-call WithStructuredOutput conflicts with client-level WithResponseFormat",
			},
			wantUnsupported: true,
		},
		{name: "no expectation and a wire failure", err: errors.New("connection reset")},
		{
			name:        "the named refusal",
			want:        cases.RefusalStructuredOutputConfig,
			err:         fmt.Errorf("%w: schema is empty", llms.ErrStructuredOutputConfig),
			wantSuccess: true,
		},
		{name: "an answer where a refusal was named", want: cases.RefusalStructuredOutputConfig, wantSays: "got an answer"},
		{
			name: "a refusal of another kind",
			want: cases.RefusalStructuredOutputConfig,
			err:  &llms.ErrStructuredOutputUnsupported{Model: "gpt-4"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := judgeRefusal(stubTestCase{want: tt.want}, tt.want, tt.err, time.Second)

			if result.Success != tt.wantSuccess || result.Unsupported != tt.wantUnsupported {
				t.Errorf("success = %v, unsupported = %v, want %v and %v (error: %v)",
					result.Success, result.Unsupported, tt.wantSuccess, tt.wantUnsupported, result.Error)
			}
			if tt.wantSays != "" && !strings.Contains(fmt.Sprint(result.Error), tt.wantSays) {
				t.Errorf("the failure does not say %q: %v", tt.wantSays, result.Error)
			}
		})
	}
}

func TestRunner_JudgeRefusal_PassesTheStructuredOutputCasesOnTheOutcomeTheyExpect(t *testing.T) {
	for _, tt := range []struct {
		name            string
		configure       func(p *mock.Provider)
		caseID          string
		wantUnsupported bool
		wantSays        string
	}{
		{
			name:      "a non-object schema the SDK refuses",
			configure: func(p *mock.Provider) { p.SetDefaultResponse(`{"name": "Alice Smith", "age": 28, "city": "Seattle"}`) },
			caseID:    "structured_output_invalid_schema",
		},
		{
			name: "a schema the model cannot take",
			configure: func(p *mock.Provider) {
				p.SetResponses([]mock.ResponseConfig{{
					Key: "name='Alice Smith', age=28, city='Seattle'",
					Response: &llms.ErrStructuredOutputUnsupported{
						Provider: "openai",
						Model:    "deepseek/deepseek-flash",
						Reason:   "the vendor's chat completions response_format takes only text and json_object",
					},
				}})
			},
			caseID:          "structured_output_with_schema",
			wantUnsupported: true,
			wantSays:        "json_object",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "deepseek/deepseek-flash")
			tt.configure(mockProvider)

			results, err := TestProvider(
				t.Context(),
				mockProvider,
				WithAgentTypes(pconfig.OptionsTypeSimpleJSON),
				WithGroups(cases.TestGroupJSON),
				WithStreamingMode(false),
			)
			if err != nil {
				t.Fatalf("TestProvider failed: %v", err)
			}

			var found bool
			for _, result := range results.SimpleJSON {
				if result.ID != tt.caseID {
					continue
				}
				found = true
				if !result.Success {
					t.Errorf("the refusal is the outcome this case expects, got a failure: %v", result.Error)
				}
				if result.Unsupported != tt.wantUnsupported {
					t.Errorf("unsupported = %v, want %v: only a model limitation is excused", result.Unsupported, tt.wantUnsupported)
				}
				if tt.wantSays != "" && (result.Error == nil || !strings.Contains(result.Error.Error(), tt.wantSays)) {
					t.Errorf("the result must keep the reason the library gave, got: %v", result.Error)
				}
			}
			if !found {
				t.Fatalf("%s did not run for the simple_json agent", tt.caseID)
			}
		})
	}
}

const truncationCaseYAML = `
- id: "capped"
  name: "Capped"
  type: "completion"
  group: "basic"
  params:
    max_tokens: 16
  messages:
    - role: "user"
      content: "write a long essay"
  expected: "a closing line the capped answer never reaches"
  expect_truncated: true
  streaming: false
`

func TestRunner_JudgeTruncation_PassesACappedAnswerOnlyWhenItStoppedAtTheLimit(t *testing.T) {
	raised := llms.Warning{
		Kind: llms.WarningClamp, Option: "WithMaxTokens", Model: "claude-haiku-4-5",
		Asked: "16", Sent: "8192", Reason: "the answer limit was raised to leave room for the thinking budget",
	}

	for _, tc := range []struct {
		label           string
		stopReason      string
		warnings        []llms.Warning
		wantPass        bool
		wantUnsupported bool
		wantSays        string
	}{
		{label: "the openai spelling", stopReason: "length", wantPass: true},
		{label: "the anthropic spelling", stopReason: "max_tokens", wantPass: true},
		{label: "finished on its own", stopReason: "stop", wantSays: `stop reason "stop"`},
		{label: "no stop reason at all", wantSays: "carries no stop reason"},
		{
			label: "raised and finished below the raised limit", stopReason: "stop", warnings: []llms.Warning{raised},
			wantPass: true, wantUnsupported: true, wantSays: "8192",
		},
		{label: "raised and still stopped at the limit", stopReason: "length", warnings: []llms.Warning{raised}, wantPass: true},
		{
			label: "another option changed on the way", stopReason: "stop",
			warnings: []llms.Warning{{Kind: llms.WarningDrop, Option: "WithTopK", Asked: "40"}},
			wantSays: `stop reason "stop"`,
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			prv := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
			prv.SetSequentialResponses(&llms.ContentResponse{
				Choices:  []*llms.ContentChoice{{Content: "The sea is", StopReason: tc.stopReason}},
				Warnings: tc.warnings,
			})
			result := testerRunOne(t, prv, truncationCaseYAML)

			if result.Success != tc.wantPass || result.Unsupported != tc.wantUnsupported {
				t.Fatalf("success = %v, unsupported = %v, want %v and %v (error: %v)",
					result.Success, result.Unsupported, tc.wantPass, tc.wantUnsupported, result.Error)
			}
			if result.StopReason != tc.stopReason {
				t.Errorf("StopReason = %q, want %q: the result must carry what the door said", result.StopReason, tc.stopReason)
			}
			if tc.wantSays != "" && (result.Error == nil || !strings.Contains(result.Error.Error(), tc.wantSays)) {
				t.Errorf("the result does not say %q: %v", tc.wantSays, result.Error)
			}
		})
	}
}

const outputLimitUnderABudgetYAML = `
- id: "capped"
  name: "Capped"
  type: "completion"
  group: "basic"
  params:
    max_tokens: 1000
  messages:
    - role: "user"
      content: "write a long essay"
  expected: ""
  expect_truncated: true
  streaming: false
`

type wireCapture struct {
	mu   sync.Mutex
	body map[string]any
}

func (c *wireCapture) last() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.body
}

func (c *wireCapture) record(body map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.body = body
}

func driveCase(t *testing.T, model, providerYAML, caseYAML string) (ProviderTestResults, map[string]any) {
	t.Helper()

	capture := &wireCapture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			fmt.Fprintf(w, `{"data":[{"id":%q,"supported_parameters":["tools"]}]}`, model)
			return
		}

		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		capture.record(body)

		openAIShapedAnswer(w)
	}))
	t.Cleanup(srv.Close)

	cfg := &config.Config{LLMServerURL: srv.URL, LLMServerKey: "k", LLMServerModel: model}
	providerConfig, err := custom.BuildProviderConfig(cfg, []byte(providerYAML))
	if err != nil {
		t.Fatalf("custom provider config: %v", err)
	}
	prv, err := custom.New(cfg, provider.DefaultProviderNameCustom, providerConfig, nil)
	if err != nil {
		t.Fatalf("custom provider: %v", err)
	}

	results, err := TestProvider(
		t.Context(),
		prv,
		WithCustomRegistry(testerRegistry(t, caseYAML)),
		WithAgentTypes(pconfig.OptionsTypeSimple),
		WithGroups(cases.TestGroupBasic),
		WithStreamingMode(false),
	)
	if err != nil {
		t.Fatalf("TestProvider: %v", err)
	}

	body := capture.last()
	if body == nil {
		t.Fatal("no request reached the wire")
	}
	return results, body
}

func TestRunner_JudgeTruncation_ExcusesTheCapARealDoorRaised(t *testing.T) {
	results, body := driveCase(t, "claude-haiku-4-5",
		"simple:\n  model: claude-haiku-4-5\n  reasoning:\n    mode: budget\n    max_tokens: 2048\n",
		outputLimitUnderABudgetYAML)

	sent, _ := body["max_completion_tokens"].(float64)
	if sent <= 1000 {
		t.Fatalf("the door sent max_completion_tokens %v; this case needs the library to raise the 1000 the case asks for", sent)
	}

	result := results.Simple[0]
	if !result.Success || !result.Unsupported {
		t.Fatalf("success = %v, unsupported = %v, want both: the cap the case asked for never reached the vendor (error: %v)",
			result.Success, result.Unsupported, result.Error)
	}
	if result.Error == nil || !strings.Contains(result.Error.Error(), strconv.Itoa(int(sent))) {
		t.Errorf("the result must name the limit that reached the vendor, %v: %v", sent, result.Error)
	}
}

func TestRunner_ProviderToolCallIDTemplate_GivesReplayedToolCallsTheProviderTemplate(t *testing.T) {
	const template = "{r:9:x}"

	recorder := &chainRecorder{
		Provider: mock.NewProvider(provider.ProviderMistral, provider.DefaultProviderNameMistral, "mistral-large-latest"),
		template: template,
	}

	_, err := TestProvider(
		t.Context(),
		recorder,
		WithAgentTypes(pconfig.OptionsTypeSimple, pconfig.OptionsTypeSimpleJSON),
		WithGroups(cases.TestGroupAdvanced),
		WithStreamingMode(false),
		WithParallelWorkers(1),
	)
	if err != nil {
		t.Fatalf("TestProvider: %v", err)
	}

	var replayed int
	for _, chain := range recorder.chains {
		called := make(map[string]bool)
		for _, message := range chain {
			for _, part := range message.Parts {
				switch part := part.(type) {
				case llms.ToolCall:
					if part.FunctionCall == nil || part.FunctionCall.Name == tools.FileToolName {
						continue
					}
					replayed++
					sample := []templates.PatternSample{{Value: part.ID, FunctionName: part.FunctionCall.Name}}
					if err := templates.ValidatePattern(template, sample); err != nil {
						t.Errorf("replayed tool call %s carries id %q, not one the provider template %q makes: %v",
							part.FunctionCall.Name, part.ID, template, err)
					}
					called[part.ID] = true
				case llms.ToolCallResponse:
					if part.Name != tools.FileToolName && !called[part.ToolCallID] {
						t.Errorf("tool response %s answers id %q, which no earlier call in the chain carries",
							part.Name, part.ToolCallID)
					}
				}
			}
		}
	}

	if replayed == 0 {
		t.Fatal("no case replayed a tool call, so nothing was checked")
	}
}

type templatedVendor struct {
	template string

	mu     sync.Mutex
	bodies []map[string]any
}

func (v *templatedVendor) takeReplayedIDs() []templates.PatternSample {
	v.mu.Lock()
	defer v.mu.Unlock()

	var replayed []templates.PatternSample
	for _, body := range v.bodies {
		messages, _ := body["messages"].([]any)
		for _, message := range messages {
			calls, _ := message.(map[string]any)["tool_calls"].([]any)
			for _, call := range calls {
				call, _ := call.(map[string]any)
				function, _ := call["function"].(map[string]any)
				id, _ := call["id"].(string)
				name, _ := function["name"].(string)
				replayed = append(replayed, templates.PatternSample{Value: id, FunctionName: name})
			}
		}
	}
	v.bodies = nil
	return replayed
}

func (v *templatedVendor) serve(w http.ResponseWriter, r *http.Request, model string) {
	if r.URL.Path == "/models" {
		fmt.Fprintf(w, `{"data":[{"id":%q,"supported_parameters":["tools"]}]}`, model)
		return
	}

	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	v.mu.Lock()
	v.bodies = append(v.bodies, body)
	v.mu.Unlock()

	offered := make(map[string]bool)
	tools, _ := body["tools"].([]any)
	for _, tool := range tools {
		function, _ := tool.(map[string]any)["function"].(map[string]any)
		name, _ := function["name"].(string)
		offered[name] = true
	}

	var name, arguments string
	switch {
	case offered["submit_pattern"]:
		name = "submit_pattern"
		encoded, _ := json.Marshal(map[string]string{"template": v.template})
		arguments = string(encoded)
	case offered["get_number"]:
		name, arguments = "get_number", `{"value": 7}`
	default:
		openAIShapedAnswer(w)
		return
	}

	call, _ := json.Marshal(map[string]any{
		"id":       templates.GenerateFromPattern(v.template, name),
		"type":     "function",
		"function": map[string]string{"name": name, "arguments": arguments},
	})
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id":"x","object":"chat.completion","choices":[{"index":0,`+
		`"message":{"role":"assistant","content":"","tool_calls":[%s]},"finish_reason":"tool_calls"}]}`, call)
}

func standUpTemplatedCustomDoor(t *testing.T, name provider.ProviderName, template string) (provider.Provider, *templatedVendor) {
	t.Helper()

	const model = "vendor-model"
	vendor := &templatedVendor{template: template}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vendor.serve(w, r, model)
	}))
	t.Cleanup(srv.Close)

	cfg := &config.Config{LLMServerURL: srv.URL, LLMServerKey: "k", LLMServerModel: model}
	providerConfig, err := custom.DefaultProviderConfig(cfg)
	if err != nil {
		t.Fatalf("custom provider config: %v", err)
	}
	prv, err := custom.New(cfg, name, providerConfig, nil)
	if err != nil {
		t.Fatalf("custom provider: %v", err)
	}
	return prv, vendor
}

func TestRunner_ProviderToolCallIDTemplate_KeepsATestOutOfTheFlowsTemplateCache(t *testing.T) {
	const (
		flowTemplate   = "call_{r:24:x}"
		testedTemplate = "{f}:{r:2:d}"
	)

	flowProvider, _ := standUpTemplatedCustomDoor(t, provider.DefaultProviderNameCustom, flowTemplate)
	testedProvider, testedVendor := standUpTemplatedCustomDoor(t, "test-provider", testedTemplate)

	runTest := func() {
		t.Helper()

		_, err := TestProvider(
			t.Context(),
			testedProvider,
			WithAgentTypes(pconfig.OptionsTypeSimple),
			WithGroups(cases.TestGroupAdvanced),
			WithStreamingMode(false),
		)
		if err != nil {
			t.Fatalf("TestProvider: %v", err)
		}

		replayed := testedVendor.takeReplayedIDs()
		if len(replayed) == 0 {
			t.Fatal("no case replayed a tool call, so nothing was checked")
		}
		if err := templates.ValidatePattern(testedTemplate, replayed); err != nil {
			t.Errorf("the tested provider's cases replayed ids not drawn from its own template %q: %v",
				testedTemplate, err)
		}
	}

	runTest()

	template, err := flowProvider.GetToolCallIDTemplate(t.Context(), templates.NewDefaultPrompter())
	if err != nil {
		t.Fatalf("flow template: %v", err)
	}
	if template != flowTemplate {
		t.Errorf("after a Test of another provider of the same type, a flow got the template %q, want its own %q",
			template, flowTemplate)
	}

	runTest()
}
