package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowAgentLogWorker interface {
	PutLog(
		ctx context.Context,
		initiator database.MsgchainType,
		executor database.MsgchainType,
		task string,
		result string,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
	GetLog(ctx context.Context, msgID int64) (database.Agentlog, error)
}

type flowAgentLogWorker struct {
	db     database.Querier
	mx     *sync.Mutex
	flowID int64
	pub    subscriptions.FlowPublisher
}

func NewFlowAgentLogWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowAgentLogWorker {
	return &flowAgentLogWorker{
		db:     db,
		mx:     &sync.Mutex{},
		flowID: flowID,
		pub:    pub,
	}
}

func (flw *flowAgentLogWorker) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	task string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	flw.mx.Lock()
	defer flw.mx.Unlock()

	flLog, err := flw.db.CreateAgentLog(ctx, database.CreateAgentLogParams{
		Initiator: initiator,
		Executor:  executor,
		Task:      task,
		Result:    result,
		FlowID:    flw.flowID,
		TaskID:    database.Int64ToNullInt64(taskID),
		SubtaskID: database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create search log: %w", err)
	}

	flw.pub.AgentLogAdded(ctx, flLog)

	return flLog.ID, nil
}

func (flw *flowAgentLogWorker) GetLog(ctx context.Context, msgID int64) (database.Agentlog, error) {
	msg, err := flw.db.GetFlowAgentLog(ctx, database.GetFlowAgentLogParams{
		ID:     msgID,
		FlowID: flw.flowID,
	})
	if err != nil {
		return database.Agentlog{}, fmt.Errorf("failed to get agent log: %w", err)
	}

	return msg, nil
}
