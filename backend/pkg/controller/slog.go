package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowSearchLogWorker interface {
	PutLog(
		ctx context.Context,
		initiator database.MsgchainType,
		executor database.MsgchainType,
		engine database.SearchengineType,
		query string,
		result string,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
	GetLog(ctx context.Context, msgID int64) (database.Searchlog, error)
}

type flowSearchLogWorker struct {
	db         database.Querier
	mx         *sync.Mutex
	flowID     int64
	containers map[int64]struct{}
	pub        subscriptions.FlowPublisher
}

func NewFlowSearchLogWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowSearchLogWorker {
	return &flowSearchLogWorker{
		db:     db,
		mx:     &sync.Mutex{},
		flowID: flowID,
		pub:    pub,
	}
}

func (slw *flowSearchLogWorker) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	slw.mx.Lock()
	defer slw.mx.Unlock()

	slLog, err := slw.db.CreateSearchLog(ctx, database.CreateSearchLogParams{
		Initiator: initiator,
		Executor:  executor,
		Engine:    engine,
		Query:     query,
		Result:    result,
		FlowID:    slw.flowID,
		TaskID:    database.Int64ToNullInt64(taskID),
		SubtaskID: database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create search log: %w", err)
	}

	slw.pub.SearchLogAdded(ctx, slLog)

	return slLog.ID, nil
}

func (slw *flowSearchLogWorker) GetLog(ctx context.Context, msgID int64) (database.Searchlog, error) {
	msg, err := slw.db.GetFlowSearchLog(ctx, database.GetFlowSearchLogParams{
		ID:     msgID,
		FlowID: slw.flowID,
	})
	if err != nil {
		return database.Searchlog{}, fmt.Errorf("failed to get search log: %w", err)
	}

	return msg, nil
}
