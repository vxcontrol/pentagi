package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/providers"
	"pentagi/pkg/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generatorFakeQuerier fails its writes once the context is done, the way pgx does.
type generatorFakeQuerier struct {
	database.Querier

	taskStatus    []database.UpdateTaskStatusParams
	failedResults []string
	subtasks      []database.Subtask
	subtasksErr   error
	subtaskLimit  int
}

func (q *generatorFakeQuerier) CreateTask(ctx context.Context, arg database.CreateTaskParams) (database.Task, error) {
	if err := ctx.Err(); err != nil {
		return database.Task{}, err
	}

	return database.Task{ID: 5, Status: arg.Status, Title: arg.Title}, nil
}

func (q *generatorFakeQuerier) UpdateTaskStatus(
	ctx context.Context, arg database.UpdateTaskStatusParams,
) (database.Task, error) {
	if err := ctx.Err(); err != nil {
		return database.Task{}, err
	}
	q.taskStatus = append(q.taskStatus, arg)

	return database.Task{ID: arg.ID, Status: arg.Status}, nil
}

func (q *generatorFakeQuerier) UpdateTaskFailedResult(
	ctx context.Context, arg database.UpdateTaskFailedResultParams,
) (database.Task, error) {
	if err := ctx.Err(); err != nil {
		return database.Task{}, err
	}
	q.taskStatus = append(q.taskStatus, database.UpdateTaskStatusParams{Status: database.TaskStatusFailed, ID: arg.ID})
	q.failedResults = append(q.failedResults, arg.Result)

	return database.Task{ID: arg.ID, Status: database.TaskStatusFailed, Result: arg.Result}, nil
}

func (q *generatorFakeQuerier) CreateSubtask(
	ctx context.Context, arg database.CreateSubtaskParams,
) (database.Subtask, error) {
	if err := ctx.Err(); err != nil {
		return database.Subtask{}, err
	}
	if q.subtaskLimit > 0 && len(q.subtasks) == q.subtaskLimit {
		return database.Subtask{}, errors.New("subtask row rejected")
	}
	subtask := database.Subtask{
		ID: int64(len(q.subtasks) + 1), TaskID: arg.TaskID, Status: arg.Status,
		Title: arg.Title, Description: arg.Description,
	}
	q.subtasks = append(q.subtasks, subtask)

	return subtask, nil
}

func (q *generatorFakeQuerier) GetTaskSubtasks(context.Context, int64) ([]database.Subtask, error) {
	if q.subtasksErr != nil {
		return nil, q.subtasksErr
	}

	return q.subtasks, nil
}

type generatorFakeMsgLog struct {
	FlowMsgLogWorker

	err error
}

func (m generatorFakeMsgLog) PutTaskMsg(ctx context.Context, _ database.MsglogType, _ int64, _, _ string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	return 1, m.err
}

// generatorFakeProvider plans task 5's subtasks; with cancel set its user stops it mid-generation.
type generatorFakeProvider struct {
	providers.FlowProvider

	plan      []tools.SubtaskInfo
	err       error
	cancel    context.CancelFunc
	refineErr error
}

func (generatorFakeProvider) GetTaskTitle(context.Context, string) (string, error) {
	return "generated title", nil
}

func (p generatorFakeProvider) GenerateSubtasks(ctx context.Context, _ int64) ([]tools.SubtaskInfo, error) {
	if p.cancel != nil {
		p.cancel()

		return nil, ctx.Err()
	}

	return p.plan, p.err
}

func (p generatorFakeProvider) RefineSubtasks(context.Context, int64) ([]tools.SubtaskInfo, error) {
	return nil, p.refineErr
}

func TestTask_AnExplicitlyEmptyGeneratorPlanRunsTheOriginalRequest(t *testing.T) {
	q := &generatorFakeQuerier{}
	pub := &cascadeFakePublisher{}
	flowCtx := &FlowContext{
		DB: q, UserID: 3, FlowID: 7,
		Provider:  generatorFakeProvider{plan: []tools.SubtaskInfo{}},
		Publisher: pub, MsgLog: generatorFakeMsgLog{},
	}
	const request = "Check the service and report what you find"

	worker, err := NewTaskWorker(context.Background(), flowCtx, request, nil)

	require.NoError(t, err)
	require.NotNil(t, worker)
	require.Len(t, q.subtasks, 1)
	assert.Equal(t, database.SubtaskStatusCreated, q.subtasks[0].Status)
	assert.Equal(t, "generated title", q.subtasks[0].Title)
	assert.Equal(t, request, q.subtasks[0].Description)
	assert.Empty(t, q.failedResults)
	require.Len(t, pub.taskSubtasks, 1)
	assert.Equal(t, q.subtasks, pub.taskSubtasks[0], "the executable fallback is published as the task plan")
}

func TestTask_AnAbsentGeneratorPlanIsNotAnExplicitEmptyPlan(t *testing.T) {
	q := &generatorFakeQuerier{}
	flowCtx := &FlowContext{
		DB: q, UserID: 3, FlowID: 7,
		Provider:  generatorFakeProvider{plan: nil},
		Publisher: &cascadeFakePublisher{}, MsgLog: generatorFakeMsgLog{},
	}

	_, err := NewTaskWorker(context.Background(), flowCtx, "check the service", nil)

	require.ErrorContains(t, err, "no subtasks generated")
	assert.Empty(t, q.subtasks, "a missing tool result must not turn into an executable request")
}

func TestTask_RefinerFailureKeepsTheExistingPlanOnlyForModelFailures(t *testing.T) {
	modelFailure := fmt.Errorf("Moonshot refused the refinement: %w", providers.ErrAgentModelCall)
	storageFailure := errors.New("subtask read failed")
	for _, tc := range []struct {
		name        string
		plan        []database.Subtask
		providerErr error
		dbErr       error
		cancelled   bool
		wantErr     error
	}{
		{
			name: "model refusal keeps the next planned subtask",
			plan: []database.Subtask{{ID: 1, Status: database.SubtaskStatusFinished},
				{ID: 2, Status: database.SubtaskStatusCreated}},
			providerErr: modelFailure,
		},
		{
			name:        "model refusal without a remaining subtask still stops",
			plan:        []database.Subtask{{ID: 1, Status: database.SubtaskStatusFinished}},
			providerErr: modelFailure, wantErr: providers.ErrAgentModelCall,
		},
		{
			name:  "database failure still stops",
			dbErr: storageFailure, wantErr: storageFailure,
		},
		{
			name:        "cancelled model run still stops",
			plan:        []database.Subtask{{ID: 2, Status: database.SubtaskStatusCreated}},
			providerErr: modelFailure, cancelled: true, wantErr: providers.ErrAgentModelCall,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancelled {
				cancel()
			}
			q := &generatorFakeQuerier{subtasks: tc.plan, subtasksErr: tc.dbErr}
			stc := NewSubtaskController(&TaskContext{
				TaskID:      5,
				FlowContext: FlowContext{DB: q, Provider: generatorFakeProvider{refineErr: tc.providerErr}},
			})

			err := stc.RefineSubtasks(ctx)
			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.wantErr)
			}
			assert.Equal(t, tc.plan, q.subtasks, "refiner failure must not mutate the original plan")
		})
	}
}

func TestTask_NewTaskWorker_ClosesATaskThatCannotStart(t *testing.T) {
	newContext := func(q *generatorFakeQuerier, provider generatorFakeProvider) (*FlowContext, *cascadeFakePublisher) {
		pub := &cascadeFakePublisher{}

		return &FlowContext{
			DB:        q,
			UserID:    3,
			FlowID:    7,
			Provider:  provider,
			Publisher: pub,
			MsgLog:    generatorFakeMsgLog{},
		}, pub
	}
	failed := []database.UpdateTaskStatusParams{{Status: database.TaskStatusFailed, ID: 5}}

	t.Run("a refused generator marks the task failed and records why", func(t *testing.T) {
		refused := errors.New("vendor refused the prompt")
		q := &generatorFakeQuerier{}
		flowCtx, pub := newContext(q, generatorFakeProvider{err: refused})

		_, err := NewTaskWorker(context.Background(), flowCtx, "input", nil)

		require.ErrorIs(t, err, refused)
		assert.Equal(t, failed, q.taskStatus, "the task is not left in the created status")
		require.Len(t, q.failedResults, 1)
		assert.Contains(t, q.failedResults[0], "vendor refused the prompt")
		require.Len(t, pub.taskUpdated, 1)
		assert.Equal(t, database.TaskStatusFailed, pub.taskUpdated[0].Status)
		assert.Contains(t, pub.taskUpdated[0].Result, "vendor refused the prompt")
	})

	t.Run("a refused input log marks the task failed", func(t *testing.T) {
		unwritable := errors.New("msglog is unwritable")
		q := &generatorFakeQuerier{}
		flowCtx, pub := newContext(q, generatorFakeProvider{})
		flowCtx.MsgLog = generatorFakeMsgLog{err: unwritable}

		_, err := NewTaskWorker(context.Background(), flowCtx, "input", nil)

		require.ErrorIs(t, err, unwritable)
		assert.Equal(t, failed, q.taskStatus)
		assert.Len(t, pub.taskUpdated, 1)
	})

	t.Run("an unreadable subtask list marks the task failed", func(t *testing.T) {
		unreadable := errors.New("subtasks are unreadable")
		q := &generatorFakeQuerier{subtasksErr: unreadable}
		flowCtx, pub := newContext(q, generatorFakeProvider{plan: []tools.SubtaskInfo{{Title: "scan"}}})

		_, err := NewTaskWorker(context.Background(), flowCtx, "input", nil)

		require.ErrorIs(t, err, unreadable)
		assert.Equal(t, failed, q.taskStatus)
		assert.Len(t, pub.taskUpdated, 1)
	})

	t.Run("a generator that stops part way publishes the subtasks it committed", func(t *testing.T) {
		q := &generatorFakeQuerier{subtaskLimit: 2}
		flowCtx, pub := newContext(q, generatorFakeProvider{plan: []tools.SubtaskInfo{{Title: "a"}, {Title: "b"}, {Title: "c"}}})

		_, err := NewTaskWorker(context.Background(), flowCtx, "input", nil)

		require.ErrorContains(t, err, "subtask row rejected")
		require.Len(t, pub.taskSubtasks, 1)
		assert.Len(t, pub.taskSubtasks[0], 2, "the client sees the subtasks the database holds")
	})

	t.Run("a generator stopped by its user closes the task without failing it", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		q := &generatorFakeQuerier{}
		flowCtx, pub := newContext(q, generatorFakeProvider{cancel: cancel})

		_, err := NewTaskWorker(ctx, flowCtx, "input", nil)

		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, []database.UpdateTaskStatusParams{{Status: database.TaskStatusFinished, ID: 5}}, q.taskStatus,
			"a task left created never loads again, and stopping a flow is not a task failure")
		assert.Len(t, pub.taskUpdated, 1)
	})
}

func TestTask_Report_RefusesACompletedTaskWithASentinel(t *testing.T) {
	t.Parallel()

	tw := &taskWorker{
		mx:        &sync.RWMutex{},
		completed: true,
		taskCtx:   &TaskContext{TaskID: 42},
	}

	// Every other field is nil, so a guard that fired after touching one would panic.
	assert.ErrorIs(t, tw.Report(context.Background()), ErrTaskAlreadyCompleted)
}
