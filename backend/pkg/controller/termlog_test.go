package controller

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/database/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sandboxQuerier fails its writes once the context is done, the way pgx does.
type sandboxQuerier struct {
	database.Querier

	containers    []database.Container
	containersErr error
	updateErr     error
	reads         int
	updates       []database.UpdateContainerStatusParams
	termLogs      []database.CreateTermLogParams
}

func (q *sandboxQuerier) GetFlowContainers(context.Context, int64) ([]database.Container, error) {
	q.reads++

	return append([]database.Container(nil), q.containers...), q.containersErr
}

func (q *sandboxQuerier) UpdateContainerStatus(
	ctx context.Context, arg database.UpdateContainerStatusParams,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
	if q.updateErr != nil {
		return database.Container{}, q.updateErr
	}
	q.updates = append(q.updates, arg)
	for i, container := range q.containers {
		if container.ID == arg.ID {
			q.containers[i].Status = arg.Status
			container.Status = arg.Status

			return container, nil
		}
	}

	return database.Container{}, errors.New("no such container")
}

func (q *sandboxQuerier) GetFlow(_ context.Context, flowID int64) (database.Flow, error) {
	return database.Flow{ID: flowID}, nil
}

func (q *sandboxQuerier) CreateTermLog(ctx context.Context, arg database.CreateTermLogParams) (database.Termlog, error) {
	if err := ctx.Err(); err != nil {
		return database.Termlog{}, err
	}
	q.termLogs = append(q.termLogs, arg)

	return database.Termlog{ID: int64(len(q.termLogs)), Type: arg.Type, Text: arg.Text}, nil
}

func TestTermlog_ContainerNotRunning_MarksARunningContainerFailedOnce(t *testing.T) {
	db := &sandboxQuerier{containers: []database.Container{
		{ID: 7, FlowID: 42, Status: database.ContainerStatusRunning},
		{ID: 3, FlowID: 42, Status: database.ContainerStatusDeleted},
	}}
	pub := &cascadeFakePublisher{}
	worker := NewFlowTermLogWorker(db, 42, pub)
	taskID, subtaskID := int64(11), int64(12)

	require.NoError(t, worker.ContainerNotRunning(t.Context(), 7, &taskID, &subtaskID))
	readsToMark := db.reads
	for range 2 {
		require.NoError(t, worker.ContainerNotRunning(t.Context(), 7, nil, nil))
	}

	assert.Equal(t, []database.UpdateContainerStatusParams{{Status: database.ContainerStatusFailed, ID: 7}}, db.updates)
	assert.Equal(t, readsToMark, db.reads, "every later tool call finds the sandbox stopped without asking the database")
	require.Len(t, pub.flowUpdated, 1)
	assert.Equal(t, int64(42), pub.flowUpdated[0].ID)
	terminals := converter.ConvertContainers(pub.flowTerms[0])
	require.Len(t, terminals, 2, "the whole list goes out, or the client drops the other terminals")
	for _, terminal := range terminals {
		assert.False(t, terminal.Connected, "terminal %d", terminal.ID)
	}
	require.Len(t, db.termLogs, 1)
	assert.Equal(t, database.TermlogTypeStderr, db.termLogs[0].Type)
	assert.Equal(t, int64(7), db.termLogs[0].ContainerID)
	assert.Equal(t, sql.NullInt64{Int64: 12, Valid: true}, db.termLogs[0].SubtaskID,
		"the Terminal tab filtered to the subtask shows why its command stopped")
}

func TestTermlog_ContainerNotRunning_WritesNothingForAContainerItCannotMark(t *testing.T) {
	readErr := errors.New("connection refused")

	for _, tc := range []struct {
		name        string
		containerID int64
		readErr     error
	}{
		{name: "a container already released", containerID: 7},
		{name: "a container already stopped", containerID: 8},
		{name: "a container already marked failed", containerID: 9},
		{name: "a container the flow does not have", containerID: 99},
		{name: "a running container whose list cannot be read", containerID: 10, readErr: readErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &sandboxQuerier{containersErr: tc.readErr, containers: []database.Container{
				{ID: 7, FlowID: 42, Status: database.ContainerStatusDeleted},
				{ID: 8, FlowID: 42, Status: database.ContainerStatusStopped},
				{ID: 9, FlowID: 42, Status: database.ContainerStatusFailed},
				{ID: 10, FlowID: 42, Status: database.ContainerStatusRunning},
			}}
			pub := &cascadeFakePublisher{}

			err := NewFlowTermLogWorker(db, 42, pub).ContainerNotRunning(t.Context(), tc.containerID, nil, nil)

			if tc.readErr != nil {
				assert.ErrorIs(t, err, tc.readErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Empty(t, db.updates)
			assert.Empty(t, pub.flowUpdated)
			assert.Empty(t, db.termLogs)
		})
	}
}

func TestTermlog_ContainerRunning_MarksAFailedContainerRunningOnceAWriteLands(t *testing.T) {
	db := &sandboxQuerier{containers: []database.Container{{ID: 7, FlowID: 42, Status: database.ContainerStatusRunning}}}
	pub := &cascadeFakePublisher{}
	worker := NewFlowTermLogWorker(db, 42, pub)

	require.NoError(t, worker.ContainerRunning(t.Context(), 7, nil, nil))
	assert.Zero(t, db.reads, "a sandbox nobody marked failed costs nothing to confirm")

	require.NoError(t, worker.ContainerNotRunning(t.Context(), 7, nil, nil))
	updateErr := errors.New("connection reset")
	db.updateErr = updateErr
	require.ErrorIs(t, worker.ContainerRunning(t.Context(), 7, nil, nil), updateErr)
	db.updateErr = nil
	require.NoError(t, worker.ContainerRunning(t.Context(), 7, nil, nil), "a mark that did not land is tried again")

	assert.Equal(t, []database.UpdateContainerStatusParams{
		{Status: database.ContainerStatusFailed, ID: 7},
		{Status: database.ContainerStatusRunning, ID: 7},
	}, db.updates)
	assert.Equal(t, database.ContainerStatusRunning, db.containers[0].Status)
	require.Len(t, pub.flowTerms, 2)
	assert.True(t, converter.ConvertContainers(pub.flowTerms[1])[0].Connected)
	require.Len(t, db.termLogs, 2)
	assert.Equal(t, database.TermlogTypeStdout, db.termLogs[1].Type)
}
