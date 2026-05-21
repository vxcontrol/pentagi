package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type ToolCallLogController interface {
	NewFlowToolCallLog(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowToolCallLogWorker, error)
	ListFlowsToolCallLog(ctx context.Context) ([]FlowToolCallLogWorker, error)
	GetFlowToolCallLog(ctx context.Context, flowID int64) (FlowToolCallLogWorker, error)
}

type toolCallLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowToolCallLogWorker
}

func NewToolCallLogController(db database.Querier) ToolCallLogController {
	return &toolCallLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowToolCallLogWorker),
	}
}

func (c *toolCallLogController) NewFlowToolCallLog(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowToolCallLogWorker, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	flw := NewFlowToolCallLogWorker(c.db, flowID, pub)
	c.flows[flowID] = flw

	return flw, nil
}

func (c *toolCallLogController) ListFlowsToolCallLog(ctx context.Context) ([]FlowToolCallLogWorker, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	flows := make([]FlowToolCallLogWorker, 0, len(c.flows))
	for _, flw := range c.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (c *toolCallLogController) GetFlowToolCallLog(
	ctx context.Context,
	flowID int64,
) (FlowToolCallLogWorker, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	flw, ok := c.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
