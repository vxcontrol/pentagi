package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/schema"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// performerNotFound is the start of the executor's answer to a call of a tool that does not exist.
const performerNotFound = "function 'execute_task_and_return_summary' not found in available tools list"

func TestPerformer_CorrectorFailure_ReportsTheToolErrorUnlessTheFlowStopped(t *testing.T) {
	toolErr := errors.New("exit status 2: no such option --frobnicate")
	expired, cancelExpired := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancelExpired()
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	for _, tc := range []struct {
		name         string
		ctx          context.Context
		correctorErr error
		wantIs       error
	}{
		{
			name:         "a gateway failure of the corrector reports the tool's own error",
			ctx:          context.Background(),
			correctorErr: fmt.Errorf("failed to call simple chain: %w", errors.New("502: bad gateway")),
			wantIs:       toolErr,
		},
		{
			name:         "a flow past its deadline reports the deadline",
			ctx:          expired,
			correctorErr: fmt.Errorf("failed to call simple chain: %w", context.DeadlineExceeded),
			wantIs:       context.DeadlineExceeded,
		},
		{
			name:         "a cancelled flow reports the cancellation",
			ctx:          cancelled,
			correctorErr: fmt.Errorf("failed to call simple chain: %w", context.Canceled),
			wantIs:       context.Canceled,
		},
		{
			name:         "a corrector cancelled while the flow still runs reports the cancellation",
			ctx:          context.Background(),
			correctorErr: fmt.Errorf("failed to call simple chain: %w", context.Canceled),
			wantIs:       context.Canceled,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, correctorFailure(tc.ctx, tc.correctorErr, "terminal", toolErr), tc.wantIs)
		})
	}
}

func TestPerformer_AiResponsePart_KeepsTheReasoningWithOrWithoutText(t *testing.T) {
	signature := []byte("signature-over-omitted-thinking")

	for _, turn := range []struct {
		name      string
		content   string
		reasoning *reasoning.ContentReasoning
	}{
		{"redacted only", "", &reasoning.ContentReasoning{
			Blocks: []reasoning.Block{{Redacted: []byte("encrypted-thinking")}},
		}},
		{"signature only", "", &reasoning.ContentReasoning{Signature: signature}},
		{"omitted thinking beside the answer", "the visible answer", &reasoning.ContentReasoning{Signature: signature}},
		{"readable thinking beside the answer", "the visible answer", &reasoning.ContentReasoning{
			Content: "why I answered that", Signature: signature,
		}},
	} {
		t.Run(turn.name, func(t *testing.T) {
			part, ok := aiResponsePart(turn.content, turn.reasoning)
			require.True(t, ok, "the turn left nothing to store")

			blob, err := json.Marshal([]llms.MessageContent{{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{part}}})
			require.NoError(t, err)
			var stored []llms.MessageContent
			require.NoError(t, json.Unmarshal(blob, &stored))

			text, ok := stored[0].Parts[0].(llms.TextContent)
			require.True(t, ok, "the stored part is %T", stored[0].Parts[0])
			assert.Equal(t, turn.content, text.Text)
			assert.Equal(t, turn.reasoning.Sequence(), text.Reasoning.Sequence(), "stored: %s", blob)
		})
	}

	t.Run("a turn with neither text nor reasoning stores nothing", func(t *testing.T) {
		_, ok := aiResponsePart("", &reasoning.ContentReasoning{})
		assert.False(t, ok)
	})
}

func TestPerformer_AiResponsePart_ReplaysATextlessThinkingTurnOnTheClaudeDoors(t *testing.T) {
	const (
		anthropicToolUse = `{"type":"tool_use","id":"toolu_1","name":"noop","input":{}}`
		bedrockToolUse   = `{"toolUse":{"toolUseId":"tooluse_1","name":"noop","input":{}}}`
	)

	for _, door := range []struct {
		name    string
		open    func(t *testing.T, content string) (provider.Provider, func() map[string]any)
		project func(block map[string]any) string
		shapes  []performerThinkingShape
	}{
		{"anthropic", performerOpenAnthropicDoor, performerAnthropicBlock, []performerThinkingShape{
			{"redacted only",
				`[{"type":"redacted_thinking","data":"encrypted-thinking"},` + anthropicToolUse + `]`,
				[]string{`redacted_thinking "encrypted-thinking"`, "tool_use noop"}},
			{"signature only",
				`[{"type":"thinking","thinking":"","signature":"sig"},` + anthropicToolUse + `]`,
				[]string{`thinking "" "sig"`, "tool_use noop"}},
			{"omitted thinking beside the answer",
				`[{"type":"thinking","thinking":"","signature":"sig"},{"type":"text","text":"Checking."},` + anthropicToolUse + `]`,
				[]string{`thinking "" "sig"`, `text "Checking."`, "tool_use noop"}},
		}},
		{"bedrock", performerOpenBedrockDoor, performerBedrockBlock, []performerThinkingShape{
			{"redacted only",
				`[{"reasoningContent":{"redactedContent":"ZW5jcnlwdGVkLXRoaW5raW5n"}},` + bedrockToolUse + `]`,
				[]string{`redactedContent "ZW5jcnlwdGVkLXRoaW5raW5n"`, "toolUse noop"}},
			{"signature only",
				`[{"reasoningContent":{"reasoningText":{"text":"","signature":"sig"}}},` + bedrockToolUse + `]`,
				[]string{`reasoningText "" "sig"`, "toolUse noop"}},
			{"omitted thinking beside the answer",
				`[{"reasoningContent":{"reasoningText":{"text":"","signature":"sig"}}},{"text":"Checking."},` + bedrockToolUse + `]`,
				[]string{`reasoningText "" "sig"`, `text "Checking."`, "toolUse noop"}},
		}},
	} {
		for _, shape := range door.shapes {
			t.Run(door.name+"/"+shape.name, func(t *testing.T) {
				prov, replayed := door.open(t, shape.content)
				performerStoreTurnAndReplay(t, prov)

				var got []string
				for _, block := range performerAssistantContent(t, replayed()) {
					if projected := door.project(block); projected != "" {
						got = append(got, projected)
					}
				}
				assert.Equal(t, shape.want, got)
			})
		}
	}
}

type performerThinkingShape struct {
	name    string
	content string
	want    []string
}

var performerNoopTools = []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{
	Name: "noop", Description: "does nothing",
	Parameters: map[string]any{"type": "object", "properties": map[string]any{}},
}}}

func performerStoreTurnAndReplay(t *testing.T, prov provider.Provider) {
	t.Helper()

	ctx := context.Background()
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "call noop")}
	resp, err := prov.CallWithTools(ctx, pconfig.OptionsTypeSimple, chain, performerNoopTools, nil)
	require.NoError(t, err)
	require.Len(t, resp.Choices, 1)
	choice := resp.Choices[0]

	turn := llms.MessageContent{Role: llms.ChatMessageTypeAI}
	if part, ok := aiResponsePart(choice.Content, choice.Reasoning); ok {
		turn.Parts = append(turn.Parts, part)
	}
	results := llms.MessageContent{Role: llms.ChatMessageTypeTool}
	for _, call := range choice.ToolCalls {
		turn.Parts = append(turn.Parts, call)
		results.Parts = append(results.Parts, llms.ToolCallResponse{
			ToolCallID: call.ID, Name: call.FunctionCall.Name, Content: "done",
		})
	}
	require.NotEmpty(t, results.Parts, "the first response carried no tool call")

	blob, err := json.Marshal(append(chain, turn, results))
	require.NoError(t, err)
	var stored []llms.MessageContent
	require.NoError(t, json.Unmarshal(blob, &stored))

	_, err = prov.CallWithTools(ctx, pconfig.OptionsTypeSimple, stored, performerNoopTools, nil)
	require.NoError(t, err)
}

func performerClaudeDoorServing(t *testing.T, start func(*httptest.Server), first, final string) (string, func() map[string]any) {
	t.Helper()

	var bodies []map[string]any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("the door sent a body that is not JSON: %v", err)
		}
		bodies = append(bodies, body)

		w.Header().Set("Content-Type", "application/json")
		if len(bodies) == 1 {
			fmt.Fprint(w, first)
			return
		}
		fmt.Fprint(w, final)
	})

	srv := httptest.NewUnstartedServer(handler)
	start(srv)
	t.Cleanup(srv.Close)

	return srv.URL, func() map[string]any {
		require.Len(t, bodies, 2)
		return bodies[1]
	}
}

func performerOpenAnthropicDoor(t *testing.T, content string) (provider.Provider, func() map[string]any) {
	t.Helper()

	url, replayed := performerClaudeDoorServing(t, (*httptest.Server).Start,
		`{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5","content":`+content+
			`,"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`,
		`{"id":"msg_2","type":"message","role":"assistant","model":"claude-sonnet-5",`+
			`"content":[{"type":"text","text":"done"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)

	providerConfig, err := anthropic.BuildProviderConfig([]byte(
		"simple:\n  model: claude-sonnet-5\n  reasoning:\n    mode: adaptive\n"))
	require.NoError(t, err)
	prov, err := anthropic.New(&config.Config{AnthropicAPIKey: "k", AnthropicServerURL: url},
		provider.DefaultProviderNameAnthropic, providerConfig)
	require.NoError(t, err)

	return prov, replayed
}

func performerOpenBedrockDoor(t *testing.T, content string) (provider.Provider, func() map[string]any) {
	t.Helper()

	url, replayed := performerClaudeDoorServing(t, (*httptest.Server).StartTLS,
		`{"output":{"message":{"role":"assistant","content":`+content+`}},`+
			`"stopReason":"tool_use","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`,
		`{"output":{"message":{"role":"assistant","content":[{"text":"done"}]}},`+
			`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`)

	providerConfig, err := bedrock.BuildProviderConfig([]byte(
		"simple:\n  model: us.anthropic.claude-sonnet-5\n  reasoning:\n    mode: adaptive\n"))
	require.NoError(t, err)
	prov, err := bedrock.New(&config.Config{
		BedrockRegion:       "us-east-1",
		BedrockBearerToken:  "t",
		BedrockServerURL:    url,
		ExternalSSLInsecure: true,
	}, provider.DefaultProviderNameBedrock, providerConfig)
	require.NoError(t, err)

	return prov, replayed
}

func performerAssistantContent(t *testing.T, body map[string]any) []map[string]any {
	t.Helper()

	messages, _ := body["messages"].([]any)
	for _, raw := range messages {
		message, _ := raw.(map[string]any)
		if message["role"] != "assistant" {
			continue
		}
		content, _ := message["content"].([]any)
		blocks := make([]map[string]any, 0, len(content))
		for _, block := range content {
			typed, _ := block.(map[string]any)
			blocks = append(blocks, typed)
		}
		return blocks
	}

	t.Fatalf("the replay carried no assistant turn: %v", body)
	return nil
}

func performerAnthropicBlock(block map[string]any) string {
	switch block["type"] {
	case "redacted_thinking":
		return fmt.Sprintf("redacted_thinking %q", block["data"])
	case "thinking":
		return fmt.Sprintf("thinking %q %q", block["thinking"], block["signature"])
	case "text":
		return fmt.Sprintf("text %q", block["text"])
	case "tool_use":
		return fmt.Sprintf("tool_use %v", block["name"])
	}
	return fmt.Sprintf("%v", block)
}

func performerBedrockBlock(block map[string]any) string {
	if thought, ok := block["reasoningContent"].(map[string]any); ok {
		if text, ok := thought["reasoningText"].(map[string]any); ok {
			return fmt.Sprintf("reasoningText %q %q", text["text"], text["signature"])
		}
		return fmt.Sprintf("redactedContent %q", thought["redactedContent"])
	}
	if call, ok := block["toolUse"].(map[string]any); ok {
		return fmt.Sprintf("toolUse %v", call["name"])
	}
	if text, ok := block["text"]; ok {
		return fmt.Sprintf("text %q", text)
	}
	if _, ok := block["cachePoint"]; ok {
		return ""
	}
	return fmt.Sprintf("%v", block)
}

func TestPerformer_PerformReflector_ChargesTheAgentThatAnswered(t *testing.T) {
	const (
		originInput  = 3.0
		originOutput = 6.0
		priceGap     = 100.0
	)

	answering := mock.NewProvider(provider.ProviderOpenAI, "openai", "origin-model")
	answering.SetPriceInfo(pconfig.OptionsTypeSimple, pconfig.PriceInfo{
		Input:  originInput,
		Output: originOutput,
	})
	answering.SetPriceInfo(pconfig.OptionsTypeReflector, pconfig.PriceInfo{
		Input:  originInput * priceGap,
		Output: originOutput * priceGap,
	})

	db := &usageCapturingQuerier{}

	fp := newFlowProvider()
	fp.db = db
	fp.Provider = answering
	fp.prompter = templates.NewDefaultPrompter()

	_, err := fp.performReflector(
		context.Background(),
		pconfig.OptionsTypeSimple,
		42,
		nil, nil,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "run the scan")},
		oneToolExecutor{},
		"run the scan",
		"I am not sure which tool to call.",
		"",
		1,
	)
	require.NoError(t, err)

	require.Len(t, db.updates, 1, "the reflector must account the answering call exactly once")

	usage := answering.GetUsage(nil)
	assert.InDelta(t, float64(usage.Input)*originInput/1e6, db.updates[0].UsageCostIn, 1e-12)
	assert.InDelta(t, float64(usage.Output)*originOutput/1e6, db.updates[0].UsageCostOut, 1e-12)
}

type performerToolRecordingProvider struct {
	*mock.Provider

	withTools    map[pconfig.ProviderOptionsType]bool
	withoutTools map[pconfig.ProviderOptionsType]bool
}

func (p *performerToolRecordingProvider) record(opt pconfig.ProviderOptionsType, tools []llms.Tool) {
	if len(tools) > 0 {
		p.withTools[opt] = true
	} else {
		p.withoutTools[opt] = true
	}
}

func (p *performerToolRecordingProvider) CallEx(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent, streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	p.record(opt, nil)

	return p.Provider.CallEx(ctx, opt, chain, streamCb)
}

func (p *performerToolRecordingProvider) CallWithTools(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent,
	tools []llms.Tool, streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	p.record(opt, tools)

	return p.Provider.CallWithTools(ctx, opt, chain, tools, streamCb)
}

// The reflector and the adviser send no tools; the agent the reflector answers for keeps its own.
func TestPerformer_ToolFreeAgentsAreCalledWithoutTools(t *testing.T) {
	recorder := &performerToolRecordingProvider{
		Provider:     mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-5.6-sol"),
		withTools:    map[pconfig.ProviderOptionsType]bool{},
		withoutTools: map[pconfig.ProviderOptionsType]bool{},
	}

	fp := newFlowProvider()
	fp.db = &usageCapturingQuerier{}
	fp.Provider = recorder
	fp.prompter = templates.NewDefaultPrompter()

	_, err := fp.performReflector(
		context.Background(),
		pconfig.OptionsTypeCoder,
		42,
		nil, nil,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "run the scan")},
		oneToolExecutor{},
		"run the scan",
		"I am not sure which tool to call.",
		"",
		1,
	)
	require.NoError(t, err)

	msgChainType := database.MsgchainTypeAdviser
	_, err = fp.performSimpleChain(context.Background(), nil, nil, agentByChain[msgChainType], msgChainType, "system", "question")
	require.NoError(t, err)

	require.True(t, recorder.withTools[pconfig.OptionsTypeCoder], "the answering agent calls with tools")
	require.True(t, recorder.withoutTools[pconfig.OptionsTypeReflector], "the reflector was called")
	require.True(t, recorder.withoutTools[pconfig.OptionsTypeAdviser], "the adviser was called")

	for opt := range recorder.withTools {
		assert.True(t, opt.UsesTools(), "%s sent tools", opt)
	}
	for opt := range recorder.withoutTools {
		if !recorder.withTools[opt] {
			assert.False(t, opt.UsesTools(), "%s ran with no tools", opt)
		}
	}
}

// performerFailingProvider fails every call; with cancel set it cancels the chain first, the way a stopped flow does.
type performerFailingProvider struct {
	provider.Provider

	calls  int
	err    error
	cancel context.CancelFunc
}

type performerExecutionContextQuerier struct {
	database.Querier
	err error
}

func (q performerExecutionContextQuerier) GetFlowTasks(context.Context, int64) ([]database.Task, error) {
	return nil, q.err
}

func TestPerformer_OnlyModelFailuresAreMarkedAsOptionalRefinerFailures(t *testing.T) {
	refusal := &reasoning.ErrReasoningOffUnsupported{Model: "refiner-test-model"}
	storageFailure := errors.New("cannot read the task plan")

	t.Run("the model refused the call", func(t *testing.T) {
		fp := newFlowProvider()
		fp.db = performerExecutionContextQuerier{}
		fp.Provider = &performerFailingProvider{err: refusal}

		err := fp.performAgentChain(context.Background(), pconfig.OptionsTypeRefiner, 42,
			nil, nil, []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "revise the plan")},
			oneToolExecutor{}, nil)

		require.ErrorIs(t, err, refusal)
		assert.ErrorIs(t, err, ErrAgentModelCall)
	})

	t.Run("a storage read failed before the model", func(t *testing.T) {
		fp := newFlowProvider()
		fp.db = performerExecutionContextQuerier{err: storageFailure}
		fp.Provider = &performerFailingProvider{err: refusal}

		err := fp.performAgentChain(context.Background(), pconfig.OptionsTypeRefiner, 42,
			nil, nil, []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "revise the plan")},
			oneToolExecutor{}, nil)

		require.ErrorIs(t, err, storageFailure)
		assert.NotErrorIs(t, err, ErrAgentModelCall)
	})

	t.Run("a reflector that cannot obtain a tool call", func(t *testing.T) {
		fp := newFlowProvider()
		fp.Provider = &performerFailingProvider{err: refusal}

		_, err := fp.performReflector(context.Background(), pconfig.OptionsTypeRefiner, 42,
			nil, nil, nil, oneToolExecutor{}, "revise the plan", "still no tool call", "", maxReflectorCallsPerChain+1)

		assert.ErrorIs(t, err, ErrAgentModelCall)
	})

	t.Run("a reflector model response followed by a usage write failure", func(t *testing.T) {
		fp := newFlowProvider()
		fp.db = &performerUsageFailingQuerier{err: storageFailure}
		fp.Provider = mock.NewProvider(provider.ProviderOpenAI, "openai", "refiner-test-model")
		fp.prompter = templates.NewDefaultPrompter()

		_, err := fp.performReflector(context.Background(), pconfig.OptionsTypeRefiner, 42,
			nil, nil,
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "revise the plan")},
			oneToolExecutor{}, "revise the plan", "choose a tool", "", 1)

		require.ErrorIs(t, err, storageFailure)
		assert.NotErrorIs(t, err, ErrAgentModelCall,
			"a successful model response must not disguise a failed DB write as an optional refiner error")
	})
}

func TestPerformer_AnEmptyGeneratorPlanMustBeExplicit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    string
		wantErr bool
	}{
		{"an explicit empty plan", `{"subtasks":[],"message":"one step"}`, false},
		{"a quoted empty plan", `{"subtasks":"[]","message":"one step"}`, false},
		{"a missing plan", `{"message":"one step"}`, true},
		{"a null plan", `{"subtasks":null,"message":"one step"}`, true},
		{"malformed arguments", `{`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list, err := decodeGeneratorSubtaskList(json.RawMessage(tc.args))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, list.Subtasks, "only an actual empty array may become an executable fallback")
			assert.Empty(t, list.Subtasks)
		})
	}
}

func (p *performerFailingProvider) Type() provider.ProviderType { return provider.ProviderOpenAI }

func (p *performerFailingProvider) CallWithTools(
	_ context.Context,
	_ pconfig.ProviderOptionsType,
	_ []llms.MessageContent,
	_ []llms.Tool,
	_ streaming.Callback,
) (*llms.ContentResponse, error) {
	p.calls++
	if p.cancel != nil {
		p.cancel()
	}

	return nil, p.err
}

// The provider hides a cancellation behind its own message; the loop must still return context.Canceled.
func TestPerformer_CallWithRetries_StopsAskingWhenRetryingCannotHelp(t *testing.T) {
	refusal := &reasoning.ErrReasoningOffUnsupported{Model: "claude-fable-5-1"}

	tests := []struct {
		name    string
		err     error
		cancels bool
		wantIs  error
	}{
		{name: "a refusal the library decided locally", err: refusal, wantIs: refusal},
		{name: "a chain cancelled while the model answered", err: errors.New("request cancelled"), cancels: true, wantIs: context.Canceled},
	}

	hook := logtest.NewGlobal()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook.Reset()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			prv := &performerFailingProvider{err: tt.err}
			if tt.cancels {
				prv.cancel = cancel
			}
			fp := newFlowProvider()
			fp.Provider = prv

			start := time.Now()
			_, err := fp.callWithRetries(
				ctx,
				pconfig.OptionsTypeSimple,
				42,
				nil, nil,
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "run the scan")},
				oneToolExecutor{},
				"run the scan",
			)
			elapsed := time.Since(start)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantIs)
			assert.Equal(t, 1, prv.calls, "the model must not be asked again")
			assert.Less(t, elapsed, delayBetweenRetries, "returning must not wait out the backoff")
			for _, entry := range hook.AllEntries() {
				assert.NotContains(t, entry.Message, "will retry",
					"the log must not promise a retry the loop is not going to make")
			}
		})
	}
}

// performerExecutor fails every call with err, or answers it as a call of a tool that does not exist.
type performerExecutor struct {
	tools.ContextToolsExecutor

	err   error
	calls int
}

// errPerformerNoSchema makes a call that reaches argument correction fail on its own error instead of a nil dereference.
var errPerformerNoSchema = errors.New("performerExecutor has no tool schemas")

func (e *performerExecutor) GetToolSchema(string) (*schema.Schema, error) {
	return nil, errPerformerNoSchema
}

func (e *performerExecutor) Execute(
	context.Context, int64, string, string, string, string, json.RawMessage,
) (string, error) {
	e.calls++
	if e.err != nil {
		return "", e.err
	}

	return performerNotFound, nil
}

func performerToolCallResult() *callResult {
	return &callResult{funcCalls: []llms.ToolCall{{
		ID:           "call_1",
		Type:         "function",
		FunctionCall: &llms.FunctionCall{Name: "execute_task_and_return_summary", Arguments: `{"question":"scan"}`},
	}}}
}

func TestPerformer_ExecToolCall_AnswersARepeatWithWhatTheToolSaidUntilItAborts(t *testing.T) {
	fp := newFlowProvider()
	fp.Provider = mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-5.6-sol")

	executor := &performerExecutor{}
	detector := &repeatingDetector{}
	result := performerToolCallResult()
	call := func() (string, error) {
		return fp.execToolCall(
			context.Background(), pconfig.OptionsTypePentester, 1, 0, result,
			&executionMonitor{}, detector, executor, nil, nil, nil,
		)
	}

	for n := 1; n <= 6; n++ {
		answer, err := call()
		require.NoError(t, err, "call %d", n)

		if n <= 2 {
			assert.Equal(t, performerNotFound, answer, "call %d reaches the tool", n)
			continue
		}
		assert.Contains(t, answer, performerNotFound, "call %d must quote what the tool said", n)
		assert.Contains(t, answer, fmt.Sprintf("%d times", n), "call %d must count the repeats", n)
	}
	assert.Equal(t, 2, executor.calls, "a repeat must not reach the tool")

	_, err := call()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repeated 7 times consecutively, aborting chain")
}

func TestPerformer_ExecToolCall_ReturnsAnErrorNoCorrectionCanFixAtOnce(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "a call cancelled with the flow", err: fmt.Errorf("terminal: %w", context.Canceled)},
		{name: "a call the flow state forbids", err: fmt.Errorf("no plan yet: %w", tools.ErrFlowStateGuard)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := newFlowProvider()
			fp.Provider = mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-5.6-sol")
			executor := &performerExecutor{err: tt.err}

			_, err := fp.execToolCall(
				context.Background(), pconfig.OptionsTypePentester, 1, 0, performerToolCallResult(),
				&executionMonitor{}, &repeatingDetector{}, executor, nil, nil, nil,
			)

			assert.ErrorIs(t, err, tt.err)
			assert.Equal(t, 1, executor.calls, "the call must not be tried again")
		})
	}
}

// usageCapturingQuerier creates chains and records usage writes; any other Querier method panics.
type usageCapturingQuerier struct {
	database.Querier

	updates []database.UpdateMsgChainUsageParams
}

type performerUsageFailingQuerier struct {
	usageCapturingQuerier
	err error
}

func (q performerUsageFailingQuerier) UpdateMsgChainUsage(
	context.Context, database.UpdateMsgChainUsageParams,
) (database.Msgchain, error) {
	return database.Msgchain{}, q.err
}

func (q *usageCapturingQuerier) CreateMsgChain(
	_ context.Context, _ database.CreateMsgChainParams,
) (database.Msgchain, error) {
	return database.Msgchain{}, nil
}

func (q *usageCapturingQuerier) UpdateMsgChainUsage(
	_ context.Context, arg database.UpdateMsgChainUsageParams,
) (database.Msgchain, error) {
	q.updates = append(q.updates, arg)

	return database.Msgchain{}, nil
}

type oneToolExecutor struct {
	tools.ContextToolsExecutor
}

func (oneToolExecutor) Tools() []llms.Tool {
	return []llms.Tool{{Function: &llms.FunctionDefinition{Name: "report"}}}
}

func (oneToolExecutor) GetBarrierToolNames() []string { return []string{"report"} }

func (oneToolExecutor) GetBarrierTools() []tools.FunctionInfo { return nil }

// performerLocalRefusal ends performAgentChain at its first agent call: isLocalRefusal stops the retries.
var performerLocalRefusal = &reasoning.ErrReasoningOffUnsupported{Model: "small-model"}

// performerWindowProvider answers the summariser and refuses the agent's call, recording what that call carried.
type performerWindowProvider struct {
	*mock.Provider

	summaries       atomic.Int64
	summariesAtCall int64
	sentBytes       int
}

func (p *performerWindowProvider) CallEx(
	ctx context.Context, opt pconfig.ProviderOptionsType, chain []llms.MessageContent, streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	p.summaries.Add(1)

	return p.Provider.CallEx(ctx, opt, chain, streamCb)
}

func (p *performerWindowProvider) CallWithTools(
	_ context.Context, _ pconfig.ProviderOptionsType, chain []llms.MessageContent, _ []llms.Tool, _ streaming.Callback,
) (*llms.ContentResponse, error) {
	p.summariesAtCall = p.summaries.Load()
	p.sentBytes += windowCompactionBytes(chain)

	return nil, performerLocalRefusal
}

// performerChainQuerier serves a flow with no tasks and counts the rewrites of the agent's chain.
type performerChainQuerier struct {
	usageCapturingQuerier

	chainWrites int
}

func (q *performerChainQuerier) GetFlowTasks(ctx context.Context, _ int64) ([]database.Task, error) {
	return nil, ctx.Err()
}

func (q *performerChainQuerier) UpdateMsgChain(
	ctx context.Context, _ database.UpdateMsgChainParams,
) (database.Msgchain, error) {
	if err := ctx.Err(); err != nil {
		return database.Msgchain{}, err
	}
	q.chainWrites++

	return database.Msgchain{}, nil
}

type performerSchemaExecutor struct {
	tools.ContextToolsExecutor

	schemas []llms.Tool
}

func (e performerSchemaExecutor) Tools() []llms.Tool { return e.schemas }

func TestPerformer_PerformAgentChain_CompactsAChainItsToolSchemasPushOutOfTheWindow(t *testing.T) {
	const chainBytes = 24016

	tests := []struct {
		name          string
		description   int
		wantCompacted bool
	}{
		{name: "a chain that fits beside its tool schemas is sent as it is", description: 100},
		{name: "a chain that fits only without its tool schemas is summarised first", description: 20000, wantCompacted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := 8000
			prv := &performerWindowProvider{Provider: mock.NewProvider(provider.ProviderOpenAI, "openai", "small-model")}
			prv.SetModels(pconfig.ModelsConfig{{Name: "small-model", ContextWindow: &window}})
			prv.SetDefaultResponse("short summary")

			db := &performerChainQuerier{}
			fp := newFlowProvider()
			fp.db = db
			fp.Provider = prv
			fp.prompter = templates.NewDefaultPrompter()
			fp.tcIDTemplate = "toolu_{r:24:b}"

			chain := windowCompactionChain(3, 4000)
			executor := performerSchemaExecutor{schemas: windowCompactionTools(tt.description)}

			err := fp.performAgentChain(
				context.Background(), pconfig.OptionsTypePentester, 42, nil, nil, chain, executor,
				csum.NewSummarizer(csum.SummarizerConfig{}),
			)
			require.ErrorIs(t, err, performerLocalRefusal)

			if !tt.wantCompacted {
				assert.Zero(t, prv.summariesAtCall)
				assert.Equal(t, chainBytes, prv.sentBytes)
				assert.Zero(t, db.chainWrites)
				return
			}
			assert.Positive(t, prv.summariesAtCall, "the summariser runs before the agent is called")
			assert.Less(t, prv.sentBytes, chainBytes)
			assert.Equal(t, 1, db.chainWrites, "the compacted chain replaces the stored one")
		})
	}
}

// performerScopeQuerier also serves the subtask reads an agent with a task or a subtask makes for its execution context.
type performerScopeQuerier struct {
	performerChainQuerier
}

func (q *performerScopeQuerier) GetFlowSubtasks(ctx context.Context, _ int64) ([]database.Subtask, error) {
	return nil, ctx.Err()
}

func (q *performerScopeQuerier) GetSubtask(ctx context.Context, _ int64) (database.Subtask, error) {
	return database.Subtask{}, ctx.Err()
}

// performerBarrierExecutor answers "ok" to every call, and a call of its barrier ends the chain.
type performerBarrierExecutor struct {
	tools.ContextToolsExecutor

	barrier string
}

func (e performerBarrierExecutor) Tools() []llms.Tool {
	return []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: e.barrier}}}
}

func (e performerBarrierExecutor) Execute(
	context.Context, int64, string, string, string, string, json.RawMessage,
) (string, error) {
	return "ok", nil
}

func (e performerBarrierExecutor) IsBarrierFunction(name string) bool { return name == e.barrier }

type performerGraphitiEpisode struct {
	name    string
	source  string
	content string
}

// performerGraphitiServing is an enabled graphiti client whose server records every message posted to it.
func performerGraphitiServing(t *testing.T) (*graphiti.Client, func() []performerGraphitiEpisode) {
	t.Helper()

	var (
		mx       sync.Mutex
		episodes []performerGraphitiEpisode
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthcheck":
			_, _ = io.WriteString(w, `{"status":"healthy"}`)
		case "/messages":
			var req graphiti.AddMessagesRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mx.Lock()
			for _, msg := range req.Messages {
				episodes = append(episodes, performerGraphitiEpisode{
					name: msg.Name, source: msg.SourceDescription, content: msg.Content,
				})
			}
			mx.Unlock()
			_, _ = io.WriteString(w, `{"message":"accepted","success":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := graphiti.NewClient(srv.URL, 10*time.Second, true)
	require.NoError(t, err)

	return client, func() []performerGraphitiEpisode {
		mx.Lock()
		defer mx.Unlock()

		return slices.Clone(episodes)
	}
}

func TestPerformer_PerformAgentChain_WritesAMissingScopeToGraphitiAsADash(t *testing.T) {
	taskID, subtaskID := int64(7), int64(11)

	tests := []struct {
		name      string
		agent     pconfig.ProviderOptionsType
		taskID    *int64
		subtaskID *int64
		barrier   string
		want      []performerGraphitiEpisode
	}{
		{
			name:  "an assistant turn has neither a task nor a subtask",
			agent: pconfig.OptionsTypeAssistant,
			want: []performerGraphitiEpisode{{
				name:    "agent_response",
				source:  "PentAGI assistant agent execution in flow 42",
				content: "Agent: assistant\nResponse: hello\nContext: Task -, Subtask -\n\n",
			}},
		},
		{
			name:    "an agent the assistant started has neither a task nor a subtask",
			agent:   pconfig.OptionsTypeSearcher,
			barrier: "search_result",
			want: []performerGraphitiEpisode{
				{
					name:    "agent_response",
					source:  "PentAGI searcher agent execution in flow 42",
					content: "Agent: searcher\nResponse: hello\nContext: Task -, Subtask -\n\n",
				},
				{
					name:   "tool_execution_search_result",
					source: "PentAGI tool execution in flow 42",
					content: "Tool: search_result\n" +
						"Description: Send the complex search result as a answer for the user question to the user\n" +
						"Barrier Function: true\nArguments: {}\nInvoked by: searcher Agent\nStatus: success\nResult: ok\n" +
						"Context: Task -, Subtask -\n\n",
				},
			},
		},
		{
			name:    "a generator run has a task but no subtask",
			agent:   pconfig.OptionsTypeGenerator,
			taskID:  &taskID,
			barrier: "subtask_list",
			want: []performerGraphitiEpisode{
				{
					name:    "agent_response",
					source:  "PentAGI generator agent execution in flow 42, task 7",
					content: "Agent: generator\nResponse: hello\nContext: Task 7, Subtask -\n\n",
				},
				{
					name:   "tool_execution_subtask_list",
					source: "PentAGI tool execution in flow 42, task 7",
					content: "Tool: subtask_list\nDescription: Send new generated subtask list to the user\n" +
						"Barrier Function: true\nArguments: {}\nInvoked by: generator Agent\nStatus: success\nResult: ok\n" +
						"Context: Task 7, Subtask -\n\n",
				},
			},
		},
		{
			name:      "an agent working on a subtask names both",
			agent:     pconfig.OptionsTypePentester,
			taskID:    &taskID,
			subtaskID: &subtaskID,
			barrier:   "hack_result",
			want: []performerGraphitiEpisode{
				{
					name:    "agent_response",
					source:  "PentAGI pentester agent execution in flow 42, task 7, subtask 11",
					content: "Agent: pentester\nResponse: hello\nContext: Task 7, Subtask 11\n\n",
				},
				{
					name:   "tool_execution_hack_result",
					source: "PentAGI tool execution in flow 42, task 7, subtask 11",
					content: "Tool: hack_result\nDescription: Send the penetration test result to the user with detailed report\n" +
						"Barrier Function: true\nArguments: {}\nInvoked by: pentester Agent\nStatus: success\nResult: ok\n" +
						"Context: Task 7, Subtask 11\n\n",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			choice := &llms.ContentChoice{Content: "hello"}
			if tt.barrier != "" {
				choice.ToolCalls = []llms.ToolCall{{
					ID:           "call_1",
					Type:         "function",
					FunctionCall: &llms.FunctionCall{Name: tt.barrier, Arguments: "{}"},
				}}
			}
			prv := mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-5.6-sol")
			prv.SetSequentialResponses(&llms.ContentResponse{Choices: []*llms.ContentChoice{choice}})

			client, episodes := performerGraphitiServing(t)
			fp := newFlowProvider()
			fp.flowID = 42
			fp.db = &performerScopeQuerier{}
			fp.Provider = prv
			fp.prompter = templates.NewDefaultPrompter()
			fp.graphitiClient = client

			err := fp.performAgentChain(
				context.Background(), tt.agent, 1, tt.taskID, tt.subtaskID, nil,
				performerBarrierExecutor{barrier: tt.barrier}, nil,
			)
			require.NoError(t, err)

			assert.Equal(t, tt.want, episodes())
		})
	}
}
