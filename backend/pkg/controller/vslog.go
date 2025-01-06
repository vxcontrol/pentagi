package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowVectorStoreLogWorker interface {
	PutLog(
		ctx context.Context,
		initiator database.MsgchainType,
		executor database.MsgchainType,
		filter json.RawMessage,
		query string,
		action database.VecstoreActionType,
		result string,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
	GetLog(ctx context.Context, msgID int64) (database.Vecstorelog, error)
}

type flowVectorStoreLogWorker struct {
	db         database.Querier
	mx         *sync.Mutex
	flowID     int64
	containers map[int64]struct{}
	pub        subscriptions.FlowPublisher
}

func NewFlowVectorStoreLogWorker(
	db database.Querier,
	flowID int64,
	pub subscriptions.FlowPublisher,
) FlowVectorStoreLogWorker {
	return &flowVectorStoreLogWorker{
		db:     db,
		mx:     &sync.Mutex{},
		flowID: flowID,
		pub:    pub,
	}
}

func (vslw *flowVectorStoreLogWorker) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	filter json.RawMessage,
	query string,
	action database.VecstoreActionType,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	vslw.mx.Lock()
	defer vslw.mx.Unlock()

	vsLog, err := vslw.db.CreateVectorStoreLog(ctx, database.CreateVectorStoreLogParams{
		Initiator: initiator,
		Executor:  executor,
		Filter:    filter,
		Query:     query,
		Action:    action,
		Result:    result,
		FlowID:    vslw.flowID,
		TaskID:    database.Int64ToNullInt64(taskID),
		SubtaskID: database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create vector store log: %w", err)
	}

	vslw.pub.VectorStoreLogAdded(ctx, vsLog)

	return vsLog.ID, nil
}

func (vslw *flowVectorStoreLogWorker) GetLog(ctx context.Context, msgID int64) (database.Vecstorelog, error) {
	msg, err := vslw.db.GetFlowVectorStoreLog(ctx, database.GetFlowVectorStoreLogParams{
		ID:     msgID,
		FlowID: vslw.flowID,
	})
	if err != nil {
		return database.Vecstorelog{}, fmt.Errorf("failed to get vector store log: %w", err)
	}

	return msg, nil
}
