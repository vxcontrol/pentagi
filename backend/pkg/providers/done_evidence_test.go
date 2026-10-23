package providers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

const (
	doneTaskID    = int64(3)
	doneSubtaskID = int64(4)
)

// doneQuerier serves the primary agent chain of subtask 4 and the tool calls already logged for it.
type doneQuerier struct {
	database.Querier

	toolcalls  []database.Toolcall
	toolsErr   error
	toolsCalls int
	results    []string
}

func (q *doneQuerier) GetMsgChain(context.Context, int64) (database.Msgchain, error) {
	return database.Msgchain{ID: 1, Chain: json.RawMessage(`[]`)}, nil
}

func (q *doneQuerier) GetSubtask(context.Context, int64) (database.Subtask, error) {
	return database.Subtask{ID: doneSubtaskID, TaskID: doneTaskID, Title: "Print the sandbox hostname"}, nil
}

func (q *doneQuerier) GetTask(context.Context, int64) (database.Task, error) {
	return database.Task{ID: doneTaskID, Title: "Print the sandbox hostname", Input: "print the hostname"}, nil
}

func (q *doneQuerier) GetFlowTasks(ctx context.Context, _ int64) ([]database.Task, error) {
	task, err := q.GetTask(ctx, doneTaskID)
	return []database.Task{task}, err
}

func (q *doneQuerier) GetFlowSubtasks(ctx context.Context, _ int64) ([]database.Subtask, error) {
	subtask, err := q.GetSubtask(ctx, doneSubtaskID)
	return []database.Subtask{subtask}, err
}

func (q *doneQuerier) GetSubtaskToolcalls(ctx context.Context, subtaskID sql.NullInt64) ([]database.Toolcall, error) {
	q.toolsCalls++
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if q.toolsErr != nil {
		return nil, q.toolsErr
	}
	var calls []database.Toolcall
	for _, tc := range q.toolcalls {
		if tc.SubtaskID == subtaskID {
			calls = append(calls, tc)
		}
	}
	return calls, nil
}

func (q *doneQuerier) UpdateSubtaskResult(_ context.Context, arg database.UpdateSubtaskResultParams) (database.Subtask, error) {
	q.results = append(q.results, arg.Result)
	return database.Subtask{ID: arg.ID, Result: arg.Result}, nil
}

var errStopAfterConfig = errors.New("stop after the primary executor config is built")

// configCapturingExecutor keeps the primary executor config so a test can call its barrier directly.
type configCapturingExecutor struct {
	tools.FlowToolsExecutor
	cfg tools.PrimaryExecutorConfig
}

func (e *configCapturingExecutor) GetPrimaryExecutor(cfg tools.PrimaryExecutorConfig) (tools.ContextToolsExecutor, error) {
	e.cfg = cfg
	return nil, errStopAfterConfig
}

func primaryBarrier(t *testing.T, q *doneQuerier) func(context.Context, string, json.RawMessage) (string, error) {
	t.Helper()

	fp := newFlowProvider()
	fp.Provider = mock.NewProvider(provider.ProviderMistral, "mistral", "ministral-8b")
	fp.db = q
	fp.prompter = templates.NewDefaultPrompter()
	exec := &configCapturingExecutor{}
	fp.executor = exec

	_, err := fp.PerformAgentChain(t.Context(), doneTaskID, doneSubtaskID, 1)
	require.ErrorIs(t, err, errStopAfterConfig)
	require.NotNil(t, exec.cfg.Barrier)

	return exec.cfg.Barrier
}

func doneArgs(success bool, result string) json.RawMessage {
	args, _ := json.Marshal(map[string]any{"success": success, "result": result, "message": "closing"})
	return args
}

func subtaskCall(subtaskID int64, name string) database.Toolcall {
	return database.Toolcall{
		Name:      name,
		Status:    database.ToolcallStatusFinished,
		SubtaskID: sql.NullInt64{Int64: subtaskID, Valid: true},
	}
}

func TestDone_PrimaryBarrier(t *testing.T) {
	const phantom = "The sandbox hostname is `debian`."

	tests := []struct {
		name         string
		toolcalls    []database.Toolcall
		toolsErr     error
		success      bool
		calls        int
		wantDeclined int
		wantStored   int
		wantLookups  int
		why          string
	}{
		{
			name:         "a success with nothing delegated is sent back once",
			toolcalls:    []database.Toolcall{subtaskCall(doneSubtaskID, tools.FinalyToolName)},
			success:      true,
			calls:        1,
			wantDeclined: 1,
			wantLookups:  1,
			why:          "no specialist produced the reported output",
		},
		{
			name:         "the same success sent again is accepted",
			toolcalls:    []database.Toolcall{subtaskCall(doneSubtaskID, tools.FinalyToolName)},
			success:      true,
			calls:        2,
			wantDeclined: 1,
			wantStored:   1,
			wantLookups:  2,
			why:          "an agent that insists the subtask needed no work is believed after one rejection",
		},
		{
			name: "a success after a delegation is accepted",
			toolcalls: []database.Toolcall{
				subtaskCall(doneSubtaskID, tools.MaintenanceToolName),
				subtaskCall(doneSubtaskID, tools.FinalyToolName),
			},
			success:     true,
			calls:       1,
			wantStored:  1,
			wantLookups: 1,
		},
		{
			name: "any specialist counts as a delegation",
			toolcalls: []database.Toolcall{
				subtaskCall(doneSubtaskID, tools.SearchToolName),
			},
			success:     true,
			calls:       1,
			wantStored:  1,
			wantLookups: 1,
		},
		{
			name:        "a failure with nothing delegated is accepted",
			success:     false,
			calls:       1,
			wantStored:  1,
			wantLookups: 0,
			why:         "a blocked or impossible subtask must be able to end without work",
		},
		{
			name: "a delegation in another subtask does not count",
			toolcalls: []database.Toolcall{
				subtaskCall(doneSubtaskID-1, tools.MaintenanceToolName),
				subtaskCall(doneSubtaskID, tools.FinalyToolName),
			},
			success:      true,
			calls:        1,
			wantDeclined: 1,
			wantLookups:  1,
		},
		{
			name:        "a failed lookup lets the result through",
			toolsErr:    errors.New("connection refused"),
			success:     true,
			calls:       1,
			wantStored:  1,
			wantLookups: 1,
			why:         "the check must never be the reason a subtask cannot finish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &doneQuerier{toolcalls: tt.toolcalls, toolsErr: tt.toolsErr}
			barrier := primaryBarrier(t, q)

			declined := 0
			for range tt.calls {
				_, err := barrier(t.Context(), tools.FinalyToolName, doneArgs(tt.success, phantom))
				var reason *tools.BarrierDeclinedError
				if errors.As(err, &reason) {
					declined++
					assert.Contains(t, reason.Reason, "Delegate the work")
					continue
				}
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantDeclined, declined, tt.why)
			assert.Len(t, q.results, tt.wantStored, tt.why)
			assert.Equal(t, tt.wantLookups, q.toolsCalls)
		})
	}
}

func TestDone_PrimaryBarrier_CanceledContextIsAnErrorNotARejection(t *testing.T) {
	q := &doneQuerier{}
	barrier := primaryBarrier(t, q)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := barrier(ctx, tools.FinalyToolName, doneArgs(true, "hostname is debian"))

	require.ErrorIs(t, err, context.Canceled)
	assert.False(t, errors.As(err, new(*tools.BarrierDeclinedError)), "a stopped flow must not be asked to delegate")
	assert.Empty(t, q.results)
}

// decliningBarrierExecutor declines the first call of its barrier and accepts the next one.
type decliningBarrierExecutor struct {
	performerBarrierExecutor

	calls int
}

func (e *decliningBarrierExecutor) Execute(
	context.Context, int64, string, string, string, string, json.RawMessage,
) (string, error) {
	e.calls++
	if e.calls == 1 {
		return "", fmt.Errorf("failed to execute handler: %w", &tools.BarrierDeclinedError{Reason: "delegate first"})
	}
	return "ok", nil
}

func TestDone_ADeclinedBarrierKeepsTheChainRunning(t *testing.T) {
	doneCall := func(id string) *llms.ContentResponse {
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{ToolCalls: []llms.ToolCall{{
			ID: id, Type: "function", FunctionCall: &llms.FunctionCall{Name: tools.FinalyToolName, Arguments: "{}"},
		}}}}}
	}
	prv := mock.NewProvider(provider.ProviderOpenAI, "openai", "gpt-5.6-sol")
	prv.SetSequentialResponses(doneCall("call_1"), doneCall("call_2"))

	fp := newFlowProvider()
	fp.flowID = 42
	fp.db = &performerScopeQuerier{}
	fp.Provider = prv
	fp.prompter = templates.NewDefaultPrompter()

	executor := &decliningBarrierExecutor{performerBarrierExecutor: performerBarrierExecutor{barrier: tools.FinalyToolName}}
	err := fp.performAgentChain(
		context.Background(), pconfig.OptionsTypePrimaryAgent, 1, nil, nil, nil, executor, nil,
	)

	require.NoError(t, err)
	assert.Equal(t, 2, executor.calls, "a declined barrier must not end the chain, nor be retried with fixed arguments")
}
