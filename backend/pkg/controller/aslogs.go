package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type AssistantLogController interface {
	NewFlowAssistantLog(
		ctx context.Context, flowID int64, assistantID int64, pub subscriptions.FlowPublisher,
	) (FlowAssistantLogWorker, error)
	ListFlowsAssistantLog(ctx context.Context, flowID int64) ([]FlowAssistantLogWorker, error)
	GetFlowAssistantLog(ctx context.Context, flowID int64, assistantID int64) (FlowAssistantLogWorker, error)
}

type assistantLogController struct {
	db    database.Querier
	mx    *sync.Mutex
	flows map[int64]map[int64]FlowAssistantLogWorker
}

func NewAssistantLogController(db database.Querier) AssistantLogController {
	return &assistantLogController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: make(map[int64]map[int64]FlowAssistantLogWorker),
	}
}

func (aslc *assistantLogController) NewFlowAssistantLog(
	ctx context.Context, flowID, assistantID int64, pub subscriptions.FlowPublisher,
) (FlowAssistantLogWorker, error) {
	aslc.mx.Lock()
	defer aslc.mx.Unlock()

	flw := NewFlowAssistantLogWorker(aslc.db, flowID, assistantID, pub)
	if _, ok := aslc.flows[flowID]; !ok {
		aslc.flows[flowID] = make(map[int64]FlowAssistantLogWorker)
	}
	aslc.flows[flowID][assistantID] = flw

	return flw, nil
}

func (aslc *assistantLogController) ListFlowsAssistantLog(
	ctx context.Context, flowID int64,
) ([]FlowAssistantLogWorker, error) {
	aslc.mx.Lock()
	defer aslc.mx.Unlock()

	if _, ok := aslc.flows[flowID]; !ok {
		return []FlowAssistantLogWorker{}, nil
	}

	flows := make([]FlowAssistantLogWorker, 0, len(aslc.flows[flowID]))
	for _, flw := range aslc.flows[flowID] {
		flows = append(flows, flw)
	}

	return flows, nil
}

func (aslc *assistantLogController) GetFlowAssistantLog(
	ctx context.Context, flowID, assistantID int64,
) (FlowAssistantLogWorker, error) {
	aslc.mx.Lock()
	defer aslc.mx.Unlock()

	flw, ok := aslc.flows[flowID]
	if !ok {
		return nil, fmt.Errorf("flow not found")
	}

	aslw, ok := flw[assistantID]
	if !ok {
		return nil, fmt.Errorf("assistant not found")
	}

	return aslw, nil
}
