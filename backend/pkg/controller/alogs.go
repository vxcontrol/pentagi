package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type AgentLogController interface {
	NewFlowAgentLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowAgentLogWorker, error)
	ListFlowsAgentLog(ctx context.Context) ([]FlowAgentLogWorker, error)
	GetFlowAgentLog(ctx context.Context, flowID int64) (FlowAgentLogWorker, error)
}

type agentLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowAgentLogWorker
}

func NewAgentLogController(db database.Querier) AgentLogController {
	return &agentLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowAgentLogWorker),
	}
}

func (alc *agentLogController) NewFlowAgentLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowAgentLogWorker, error) {
	alc.mx.Lock()
	defer alc.mx.Unlock()

	flw := NewFlowAgentLogWorker(alc.db, flowID, pub)
	alc.flows[flowID] = flw

	return flw, nil
}

func (alc *agentLogController) ListFlowsAgentLog(ctx context.Context) ([]FlowAgentLogWorker, error) {
	alc.mx.Lock()
	defer alc.mx.Unlock()

	flows := make([]FlowAgentLogWorker, 0, len(alc.flows))
	for _, flw := range alc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (alc *agentLogController) GetFlowAgentLog(
	ctx context.Context,
	flowID int64,
) (FlowAgentLogWorker, error) {
	alc.mx.Lock()
	defer alc.mx.Unlock()

	flw, ok := alc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
