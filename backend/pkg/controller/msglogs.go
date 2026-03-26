package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type MsgLogController interface {
	NewFlowMsgLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowMsgLogWorker, error)
	ListFlowsMsgLog(ctx context.Context) ([]FlowMsgLogWorker, error)
	GetFlowMsgLog(ctx context.Context, flowID int64) (FlowMsgLogWorker, error)
}

type msgLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowMsgLogWorker
}

func NewMsgLogController(db database.Querier) MsgLogController {
	return &msgLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowMsgLogWorker),
	}
}

func (mlc *msgLogController) NewFlowMsgLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowMsgLogWorker, error) {
	mlc.mx.Lock()
	defer mlc.mx.Unlock()

	flw := NewFlowMsgLogWorker(mlc.db, flowID, pub)
	mlc.flows[flowID] = flw

	return flw, nil
}

func (mlc *msgLogController) ListFlowsMsgLog(ctx context.Context) ([]FlowMsgLogWorker, error) {
	mlc.mx.Lock()
	defer mlc.mx.Unlock()

	flows := make([]FlowMsgLogWorker, 0, len(mlc.flows))
	for _, flw := range mlc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (mlc *msgLogController) GetFlowMsgLog(ctx context.Context, flowID int64) (FlowMsgLogWorker, error) {
	mlc.mx.Lock()
	defer mlc.mx.Unlock()

	flw, ok := mlc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
