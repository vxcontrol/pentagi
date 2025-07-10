package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowTermLogWorker interface {
	PutMsg(ctx context.Context, msgType database.TermlogType, msg string, containerID int64) (int64, error)
	GetMsg(ctx context.Context, msgID int64) (database.Termlog, error)
	GetContainers(ctx context.Context) ([]database.Container, error)
}

type flowTermLogWorker struct {
	db         database.Querier
	mx         *sync.Mutex
	flowID     int64
	containers map[int64]struct{}
	pub        subscriptions.FlowPublisher
}

func NewFlowTermLogWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowTermLogWorker {
	return &flowTermLogWorker{
		db:         db,
		mx:         &sync.Mutex{},
		flowID:     flowID,
		containers: make(map[int64]struct{}),
		pub:        pub,
	}
}

func (tlw *flowTermLogWorker) PutMsg(
	ctx context.Context,
	msgType database.TermlogType,
	msg string,
	containerID int64,
) (int64, error) {
	tlw.mx.Lock()
	defer tlw.mx.Unlock()

	if _, ok := tlw.containers[containerID]; !ok {
		// try to update the container map
		containers, err := tlw.GetContainers(ctx)
		if err != nil {
			return 0, err
		}
		tlw.containers = make(map[int64]struct{})
		for _, container := range containers {
			tlw.containers[container.ID] = struct{}{}
		}
		if _, ok := tlw.containers[containerID]; !ok {
			return 0, fmt.Errorf("container not found")
		}
	}

	termLog, err := tlw.db.CreateTermLog(ctx, database.CreateTermLogParams{
		Type:        msgType,
		Text:        database.SanitizeUTF8(msg),
		ContainerID: containerID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create termlog: %w", err)
	}

	tlw.pub.TerminalLogAdded(ctx, termLog)

	return termLog.ID, nil
}

func (tlw *flowTermLogWorker) GetMsg(ctx context.Context, msgID int64) (database.Termlog, error) {
	msg, err := tlw.db.GetTermLog(ctx, msgID)
	if err != nil {
		return database.Termlog{}, fmt.Errorf("failed to get termlog: %w", err)
	}

	return msg, nil
}

func (tlw *flowTermLogWorker) GetContainers(ctx context.Context) ([]database.Container, error) {
	containers, err := tlw.db.GetFlowContainers(ctx, tlw.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers: %w", err)
	}

	return containers, nil
}
