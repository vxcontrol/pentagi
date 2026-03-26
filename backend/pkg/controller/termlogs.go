package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type TermLogController interface {
	NewFlowTermLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowTermLogWorker, error)
	ListFlowsTermLog(ctx context.Context) ([]FlowTermLogWorker, error)
	GetFlowTermLog(ctx context.Context, flowID int64) (FlowTermLogWorker, error)
	GetFlowContainers(ctx context.Context, flowID int64) ([]database.Container, error)
}

type termLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowTermLogWorker
}

func NewTermLogController(db database.Querier) TermLogController {
	return &termLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowTermLogWorker),
	}
}

func (tlc *termLogController) NewFlowTermLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowTermLogWorker, error) {
	tlc.mx.Lock()
	defer tlc.mx.Unlock()

	flw := NewFlowTermLogWorker(tlc.db, flowID, pub)
	tlc.flows[flowID] = flw

	return flw, nil
}

func (tlc *termLogController) ListFlowsTermLog(ctx context.Context) ([]FlowTermLogWorker, error) {
	tlc.mx.Lock()
	defer tlc.mx.Unlock()

	flows := make([]FlowTermLogWorker, 0, len(tlc.flows))
	for _, flw := range tlc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (tlc *termLogController) GetFlowTermLog(ctx context.Context, flowID int64) (FlowTermLogWorker, error) {
	tlc.mx.Lock()
	defer tlc.mx.Unlock()

	flw, ok := tlc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}

func (tlc *termLogController) GetFlowContainers(ctx context.Context, flowID int64) ([]database.Container, error) {
	tlc.mx.Lock()
	defer tlc.mx.Unlock()

	flw, ok := tlc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw.GetContainers(ctx)
}
