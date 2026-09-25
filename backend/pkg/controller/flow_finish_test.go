package controller

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowFinish_FinishUnloadedFlow_ClosesTheFlowAndWhatItLeftOpen(t *testing.T) {
	t.Parallel()

	const flowID = int64(7)

	newQuerier := func() *finishFakeQuerier {
		return &finishFakeQuerier{
			flow: database.Flow{ID: flowID, UserID: 3, Status: database.FlowStatusRunning},
			tasks: []database.Task{
				{ID: 11, Status: database.TaskStatusRunning},
				{ID: 12, Status: database.TaskStatusFinished},
				{ID: 13, Status: database.TaskStatusFinished},
			},
			subtasks: map[int64][]database.Subtask{
				11: {{ID: 21, Status: database.SubtaskStatusRunning}},
				13: {{ID: 23, Status: database.SubtaskStatusRunning}},
			},
			assistants: []database.Assistant{
				{ID: 31, Status: database.AssistantStatusRunning},
				{ID: 32, Status: database.AssistantStatusFinished},
			},
			containers: []database.Container{
				{ID: 41, Status: database.ContainerStatusRunning, LocalID: sql.NullString{String: "cid-41", Valid: true}},
				{ID: 42, Status: database.ContainerStatusDeleted, LocalID: sql.NullString{String: "cid-42", Valid: true}},
			},
		}
	}
	finish := func(q *finishFakeQuerier, dockerErr error) (*cascadeFakePublisher, *finishFakeDocker, error) {
		fc, pub, dkr := newFinishController(q)
		dkr.err = dockerErr

		return pub, dkr, fc.FinishFlow(context.Background(), flowID)
	}
	finished := []database.UpdateFlowStatusParams{{Status: database.FlowStatusFinished, ID: flowID}}

	t.Run("a running flow is closed with every task, subtask and assistant it left open", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		pub, dkr, err := finish(q, nil)

		require.NoError(t, err, "a flow whose provider stopped resolving is still closable")
		assert.Equal(t, finished, q.flowStatus)
		assert.Equal(t, []database.UpdateTaskStatusParams{{Status: database.TaskStatusFinished, ID: 11}}, q.taskStatus,
			"a closed task is not rewritten")
		assert.Equal(t, []database.UpdateSubtaskStatusParams{
			{Status: database.SubtaskStatusFinished, ID: 21},
			{Status: database.SubtaskStatusFinished, ID: 23},
		}, q.subtaskStatus)
		require.Equal(t, []database.Task{
			{ID: 11, Status: database.TaskStatusFinished},
			{ID: 13, Status: database.TaskStatusFinished},
		}, pub.taskUpdated, "an open tab keeps a subtask running until it hears about it; a task with nothing closed is not republished")
		assert.Equal(t, []database.Subtask{{ID: 23, Status: database.SubtaskStatusFinished}}, pub.taskSubtasks[1])
		assert.Equal(t, []database.UpdateAssistantStatusParams{{Status: database.AssistantStatusFinished, ID: 31}},
			q.assistantStatus, "only the unfinished assistant is closed")
		require.Len(t, pub.assistantUpdated, 1)
		assert.Equal(t, int64(31), pub.assistantUpdated[0].ID)
		assert.Equal(t, []string{"cid-41"}, dkr.removed, "an already deleted container is not purged again")
		require.Len(t, pub.flowUpdated, 1, "the client learns the flow is finished")
		assert.Equal(t, database.FlowStatusFinished, pub.flowUpdated[0].Status)
	})

	t.Run("a container query that fails once is read again for the event", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		q.containersErr = errors.New("containers are unreadable")
		q.containersErrOnce = true
		pub, dkr, err := finish(q, nil)

		require.NoError(t, err)
		require.Len(t, pub.flowUpdated, 1, "the subscriber learns the flow is finished")
		assert.Equal(t, database.FlowStatusFinished, pub.flowUpdated[0].Status)
		assert.Len(t, pub.flowTerms[0], 2, "the event carries the terminals the flow still has")
		assert.Empty(t, dkr.removed, "the failed read left nothing to release")
	})

	t.Run("a container query that keeps failing publishes no event", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		q.containersErr = errors.New("containers are unreadable")
		pub, _, err := finish(q, nil)

		require.NoError(t, err)
		assert.Equal(t, finished, q.flowStatus, "the row is still closed")
		assert.Empty(t, pub.flowUpdated, "an event with no terminals would wipe the terminals an open tab shows")
	})

	t.Run("a failed assistant close does not keep the flow open", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		q.assistantStatusErr = errors.New("assistant row is locked")
		pub, _, err := finish(q, nil)

		require.NoError(t, err)
		assert.Equal(t, finished, q.flowStatus)
		assert.Len(t, pub.flowUpdated, 1)
		assert.Empty(t, pub.assistantUpdated)
	})

	t.Run("a failed purge does not keep the flow open", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		pub, _, err := finish(q, errors.New("docker is down"))

		require.NoError(t, err)
		assert.Equal(t, finished, q.flowStatus)
		assert.Len(t, pub.flowUpdated, 1)
	})

	t.Run("a missing row is reported as not found", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		q.flowErr = sql.ErrNoRows
		pub, _, err := finish(q, nil)

		assert.ErrorIs(t, err, ErrFlowNotFound)
		assert.Empty(t, q.flowStatus, "a flow that does not exist is not written to")
		assert.Empty(t, pub.flowUpdated)
	})

	t.Run("an already closed flow has its leftovers closed but is not written again", func(t *testing.T) {
		t.Parallel()

		q := newQuerier()
		q.flow.Status = database.FlowStatusFinished
		pub, dkr, err := finish(q, nil)

		require.NoError(t, err)
		assert.Empty(t, q.flowStatus)
		assert.Empty(t, pub.flowUpdated)
		assert.Equal(t, []database.UpdateTaskStatusParams{{Status: database.TaskStatusFinished, ID: 11}}, q.taskStatus,
			"a repeat is the only thing left that can close the task")
		assert.Equal(t, []database.UpdateAssistantStatusParams{{Status: database.AssistantStatusFinished, ID: 31}},
			q.assistantStatus, "and the assistant")
		assert.Equal(t, []string{"cid-41"}, dkr.removed, "and release the sandbox")
	})
}

type budgetEatingDocker struct {
	docker.DockerClient

	removed atomic.Int64
}

func (d *budgetEatingDocker) RemoveContainer(ctx context.Context, _ string, _ int64) error {
	d.removed.Add(1)
	<-ctx.Done()

	return ctx.Err()
}

func TestFlowFinish_FinishUnloadedFlow_RecordsTheFlowBeforeTheSandboxSpendsTheBudget(t *testing.T) {
	t.Parallel()

	const flowID = int64(7)

	q := &finishFakeQuerier{
		flow: database.Flow{ID: flowID, Status: database.FlowStatusWaiting},
		containers: []database.Container{{
			ID:      1,
			FlowID:  flowID,
			Status:  database.ContainerStatusRunning,
			LocalID: sql.NullString{String: "abc", Valid: true},
		}},
	}
	dkr := &budgetEatingDocker{}
	fc, _, _ := newFinishController(q)
	fc.docker = dkr

	require.NoError(t, fc.finishFlowWithin(context.Background(), flowID, 50*time.Millisecond))

	require.Equal(t, int64(1), dkr.removed.Load(), "the sandbox teardown is still attempted")
	assert.Equal(t, []database.UpdateFlowStatusParams{{Status: database.FlowStatusFinished, ID: flowID}}, q.flowStatus,
		"a sandbox that outlives the budget does not keep the flow open")
}

func TestFlowFinish_MissingFlowError_AnswersNotFoundOnlyForAFlowTheDatabaseLacks(t *testing.T) {
	t.Parallel()

	stored := database.Flow{ID: 7, Status: database.FlowStatusRunning}
	unreadable := errors.New("connection reset")
	hungUp, cancel := context.WithCancel(context.Background())
	cancel()
	expired, cancelExpired := context.WithDeadline(context.Background(), time.Unix(0, 0))
	cancelExpired()

	for _, tc := range []struct {
		name    string
		q       *finishFakeQuerier
		ctx     context.Context
		wantErr error
	}{
		{name: "a flow the database holds is not loaded", q: &finishFakeQuerier{flow: stored}, ctx: context.Background(), wantErr: ErrFlowNotLoaded},
		{name: "a flow the database lacks is not found", q: &finishFakeQuerier{flowErr: sql.ErrNoRows}, ctx: context.Background(), wantErr: ErrFlowNotFound},
		{name: "a flow that cannot be read is reported as it failed", q: &finishFakeQuerier{flowErr: unreadable}, ctx: context.Background(), wantErr: unreadable},
		{name: "a caller that hung up gets its own cancellation", q: &finishFakeQuerier{flow: stored}, ctx: hungUp, wantErr: context.Canceled},
		{name: "a caller whose deadline ran out gets its own deadline", q: &finishFakeQuerier{flow: stored}, ctx: expired, wantErr: context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fc, _, _ := newFinishController(tc.q)
			for call, err := range map[string]error{
				"stop":   fc.StopFlow(tc.ctx, 7),
				"rename": fc.RenameFlow(tc.ctx, 7, "title"),
				"report": fc.ReportFlow(tc.ctx, 7, 0),
			} {
				assert.ErrorIs(t, err, tc.wantErr, call)
				if !errors.Is(tc.wantErr, ErrFlowNotFound) {
					assert.NotErrorIs(t, err, ErrFlowNotFound, "%s answered a flow the database may hold as missing", call)
				}
				if !errors.Is(tc.wantErr, ErrFlowNotFound) && !errors.Is(tc.wantErr, ErrFlowNotLoaded) {
					assert.NotErrorIs(t, err, ErrFlowNotLoaded, "%s answered a failed read as an unloaded flow", call)
				}
			}
		})
	}
}

func TestFlowFinish_FinishFlowLeftovers_ClosesWhatTheWorkerDidNotHold(t *testing.T) {
	t.Parallel()

	const flowID = int64(35)
	unreadable := errors.New("connection reset")

	for _, tc := range []struct {
		name    string
		q       *finishFakeQuerier
		wantErr error
	}{
		{
			name: "a task and an assistant the worker never held are closed",
			q: &finishFakeQuerier{
				flow:       database.Flow{ID: flowID, UserID: 5, Status: database.FlowStatusWaiting},
				tasks:      []database.Task{{ID: 59, Status: database.TaskStatusCreated}},
				assistants: []database.Assistant{{ID: 31, Status: database.AssistantStatusWaiting}},
			},
		},
		{name: "a row gone under the worker is reported as not found", q: &finishFakeQuerier{flowErr: sql.ErrNoRows}, wantErr: ErrFlowNotFound},
		{name: "a row that cannot be read is reported as it failed", q: &finishFakeQuerier{flowErr: unreadable}, wantErr: unreadable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fc, pub, _ := newFinishController(tc.q)
			worker := &finishRecordingWorker{noopFlowWorker: noopFlowWorker{flowID: flowID}}
			fc.register(flowID, worker)

			err := fc.FinishFlow(context.Background(), flowID)

			assert.True(t, worker.finished.Load(), "the loaded path still goes through the worker")
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, []database.UpdateTaskStatusParams{{Status: database.TaskStatusFinished, ID: 59}}, tc.q.taskStatus,
				"a task the worker never restored is closed too")
			assert.Len(t, pub.taskUpdated, 1, "the client learns the task is closed")
			assert.Equal(t, []database.UpdateAssistantStatusParams{{Status: database.AssistantStatusFinished, ID: 31}},
				tc.q.assistantStatus, "an assistant whose provider stopped resolving never reaches the worker")
			assert.Len(t, pub.assistantUpdated, 1, "an open tab keeps showing it active until it hears otherwise")
		})
	}
}
