package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowToolCallLogWorker interface {
	PutLog(
		ctx context.Context,
		callID string,
		name string,
		args json.RawMessage,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
	UpdateLogSuccess(ctx context.Context, id int64, result string, durationSeconds float64) error
	UpdateLogFailed(ctx context.Context, id int64, result string, durationSeconds float64) error
	GetLog(ctx context.Context, id int64) (database.Toolcall, error)
}

type flowToolCallLogWorker struct {
	db     database.Querier
	mx     *sync.Mutex
	flowID int64
	pub    subscriptions.FlowPublisher
}

func NewFlowToolCallLogWorker(
	db database.Querier,
	flowID int64,
	pub subscriptions.FlowPublisher,
) FlowToolCallLogWorker {
	return &flowToolCallLogWorker{
		db:     db,
		mx:     &sync.Mutex{},
		flowID: flowID,
		pub:    pub,
	}
}

func (w *flowToolCallLogWorker) PutLog(
	ctx context.Context,
	callID string,
	name string,
	args json.RawMessage,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	w.mx.Lock()
	defer w.mx.Unlock()

	tc, err := w.db.CreateToolcall(ctx, database.CreateToolcallParams{
		CallID:    callID,
		Status:    database.ToolcallStatusRunning,
		Name:      name,
		Args:      args,
		FlowID:    w.flowID,
		TaskID:    database.Int64ToNullInt64(taskID),
		SubtaskID: database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create tool call log: %w", err)
	}

	w.pub.ToolCallLogAdded(ctx, tc)

	return tc.ID, nil
}

func (w *flowToolCallLogWorker) UpdateLogSuccess(
	ctx context.Context,
	id int64,
	result string,
	durationSeconds float64,
) error {
	tc, err := w.db.UpdateToolcallFinishedResult(ctx, database.UpdateToolcallFinishedResultParams{
		Result:          result,
		DurationSeconds: durationSeconds,
		ID:              id,
	})
	if err != nil {
		return fmt.Errorf("failed to update tool call log result: %w", err)
	}

	w.pub.ToolCallLogUpdated(ctx, tc)

	return nil
}

func (w *flowToolCallLogWorker) UpdateLogFailed(
	ctx context.Context,
	id int64,
	result string,
	durationSeconds float64,
) error {
	tc, err := w.db.UpdateToolcallFailedResult(ctx, database.UpdateToolcallFailedResultParams{
		Result:          result,
		DurationSeconds: durationSeconds,
		ID:              id,
	})
	if err != nil {
		return fmt.Errorf("failed to update tool call log failed result: %w", err)
	}

	w.pub.ToolCallLogUpdated(ctx, tc)

	return nil
}

func (w *flowToolCallLogWorker) GetLog(ctx context.Context, id int64) (database.Toolcall, error) {
	tc, err := w.db.GetFlowToolcall(ctx, database.GetFlowToolcallParams{
		ID:     id,
		FlowID: w.flowID,
	})
	if err != nil {
		return database.Toolcall{}, fmt.Errorf("failed to get tool call log: %w", err)
	}

	return tc, nil
}
