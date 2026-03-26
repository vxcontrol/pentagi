package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type SearchLogController interface {
	NewFlowSearchLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowSearchLogWorker, error)
	ListFlowsSearchLog(ctx context.Context) ([]FlowSearchLogWorker, error)
	GetFlowSearchLog(ctx context.Context, flowID int64) (FlowSearchLogWorker, error)
}

type searchLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowSearchLogWorker
}

func NewSearchLogController(db database.Querier) SearchLogController {
	return &searchLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowSearchLogWorker),
	}
}

func (slc *searchLogController) NewFlowSearchLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowSearchLogWorker, error) {
	slc.mx.Lock()
	defer slc.mx.Unlock()

	flw := NewFlowSearchLogWorker(slc.db, flowID, pub)
	slc.flows[flowID] = flw

	return flw, nil
}

func (slc *searchLogController) ListFlowsSearchLog(ctx context.Context) ([]FlowSearchLogWorker, error) {
	slc.mx.Lock()
	defer slc.mx.Unlock()

	flows := make([]FlowSearchLogWorker, 0, len(slc.flows))
	for _, flw := range slc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (slc *searchLogController) GetFlowSearchLog(
	ctx context.Context,
	flowID int64,
) (FlowSearchLogWorker, error) {
	slc.mx.Lock()
	defer slc.mx.Unlock()

	flw, ok := slc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
