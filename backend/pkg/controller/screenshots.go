package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type ScreenshotController interface {
	NewFlowScreenshot(ctx context.Context, flowID int64, pub subscriptions.FlowPublisher) (FlowScreenshotWorker, error)
	ListFlowsScreenshot(ctx context.Context) ([]FlowScreenshotWorker, error)
	GetFlowScreenshot(ctx context.Context, flowID int64) (FlowScreenshotWorker, error)
}

type screenshotController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]FlowScreenshotWorker
}

func NewScreenshotController(db database.Querier) ScreenshotController {
	return &screenshotController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]FlowScreenshotWorker),
	}
}

func (sc *screenshotController) NewFlowScreenshot(
	ctx context.Context,
	flowID int64,
	pub subscriptions.FlowPublisher,
) (FlowScreenshotWorker, error) {
	sc.mx.Lock()
	defer sc.mx.Unlock()

	flw := NewFlowScreenshotWorker(sc.db, flowID, pub)
	sc.flows[flowID] = flw

	return flw, nil
}

func (sc *screenshotController) ListFlowsScreenshot(ctx context.Context) ([]FlowScreenshotWorker, error) {
	sc.mx.Lock()
	defer sc.mx.Unlock()

	flows := make([]FlowScreenshotWorker, 0, len(sc.flows))
	for _, flw := range sc.flows {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (sc *screenshotController) GetFlowScreenshot(ctx context.Context, flowID int64) (FlowScreenshotWorker, error) {
	sc.mx.Lock()
	defer sc.mx.Unlock()

	flw, ok := sc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	return flw, nil
}
