package controller

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
)

type FlowTermLogWorker interface {
	PutMsg(
		ctx context.Context,
		msgType database.TermlogType,
		msg string,
		containerID int64,
		taskID, subtaskID *int64,
	) (int64, error)
	GetMsg(ctx context.Context, msgID int64) (database.Termlog, error)
	GetContainers(ctx context.Context) ([]database.Container, error)
	ContainerNotRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error
	ContainerRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error
}

type flowTermLogWorker struct {
	db         database.Querier
	mx         *sync.Mutex
	flowID     int64
	containers map[int64]struct{}
	pub        subscriptions.FlowPublisher

	stateMX sync.Mutex
	failed  map[int64]struct{}
}

func NewFlowTermLogWorker(db database.Querier, flowID int64, pub subscriptions.FlowPublisher) FlowTermLogWorker {
	return &flowTermLogWorker{
		db:         db,
		mx:         &sync.Mutex{},
		flowID:     flowID,
		containers: make(map[int64]struct{}),
		pub:        pub,
		failed:     make(map[int64]struct{}),
	}
}

func (tlw *flowTermLogWorker) PutMsg(
	ctx context.Context,
	msgType database.TermlogType,
	msg string,
	containerID int64,
	taskID, subtaskID *int64,
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
		FlowID:      tlw.flowID,
		TaskID:      database.Int64ToNullInt64(taskID),
		SubtaskID:   database.Int64ToNullInt64(subtaskID),
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

func (tlw *flowTermLogWorker) ContainerNotRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error {
	tlw.stateMX.Lock()
	defer tlw.stateMX.Unlock()

	if _, isFailed := tlw.failed[containerID]; isFailed {
		return nil
	}

	containers, idx, err := tlw.containerWithStatus(ctx, containerID, database.ContainerStatusRunning)
	if err != nil || idx < 0 {
		return err
	}

	containers[idx], err = tlw.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		Status: database.ContainerStatusFailed,
		ID:     containerID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark container %d failed: %w", containerID, err)
	}
	tlw.failed[containerID] = struct{}{}

	return tlw.publishContainerChange(ctx, containers, containerID,
		database.TermlogTypeStderr, "The sandbox container stopped.\r\n", taskID, subtaskID)
}

func (tlw *flowTermLogWorker) ContainerRunning(ctx context.Context, containerID int64, taskID, subtaskID *int64) error {
	tlw.stateMX.Lock()
	defer tlw.stateMX.Unlock()

	if _, isFailed := tlw.failed[containerID]; !isFailed {
		return nil
	}

	containers, idx, err := tlw.containerWithStatus(ctx, containerID, database.ContainerStatusFailed)
	if err != nil {
		return err
	}
	if idx < 0 {
		delete(tlw.failed, containerID)
		return nil
	}

	containers[idx], err = tlw.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		Status: database.ContainerStatusRunning,
		ID:     containerID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark container %d running: %w", containerID, err)
	}
	delete(tlw.failed, containerID)

	return tlw.publishContainerChange(ctx, containers, containerID,
		database.TermlogTypeStdout, "The sandbox container is running again.\r\n", taskID, subtaskID)
}

func (tlw *flowTermLogWorker) containerWithStatus(
	ctx context.Context, containerID int64, status database.ContainerStatus,
) ([]database.Container, int, error) {
	containers, err := tlw.GetContainers(ctx)
	if err != nil {
		return nil, -1, err
	}

	idx := slices.IndexFunc(containers, func(container database.Container) bool {
		return container.ID == containerID && container.Status == status
	})

	return containers, idx, nil
}

func (tlw *flowTermLogWorker) publishContainerChange(
	ctx context.Context,
	containers []database.Container,
	containerID int64,
	msgType database.TermlogType,
	msg string,
	taskID, subtaskID *int64,
) error {
	flow, err := tlw.db.GetFlow(ctx, tlw.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow: %w", err)
	}

	tlw.pub.FlowUpdated(ctx, flow, containers)

	_, err = tlw.PutMsg(ctx, msgType, msg, containerID, taskID, subtaskID)

	return err
}
