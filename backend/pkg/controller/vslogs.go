package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type VectorStoreLogController interface {
	NewFlowVectorStoreLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowVectorStoreLogWorker, error)
	ListFlowsVectorStoreLog(ctx context.Context) ([]FlowVectorStoreLogWorker, error)
	GetFlowVectorStoreLog(ctx context.Context, flowID int64) (FlowVectorStoreLogWorker, error)
}

type vectorStoreLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowVectorStoreLogWorker
}

func NewVectorStoreLogController(db database.Querier) VectorStoreLogController {
	return &vectorStoreLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowVectorStoreLogWorker),
	}
}

func (vslc *vectorStoreLogController) NewFlowVectorStoreLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowVectorStoreLogWorker, error) {
	vslc.mx.Lock()
	defer vslc.mx.Unlock()

	flw := NewFlowVectorStoreLogWorker(vslc.db, flowID, pub)
	vslc.flows[flowID] = flw

	return flw, nil
}

func (tlc *vectorStoreLogController) ListFlowsVectorStoreLog(ctx context.Context) ([]FlowVectorStoreLogWorker, error) {
	tlc.mx.Lock()
	defer tlc.mx.Unlock()

	flows := make([]FlowVectorStoreLogWorker, 0, len(tlc.flows))
	for _, flw := range tlc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (vslc *vectorStoreLogController) GetFlowVectorStoreLog(
	ctx context.Context,
	flowID int64,
) (FlowVectorStoreLogWorker, error) {
	vslc.mx.Lock()
	defer vslc.mx.Unlock()

	flw, ok := vslc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
