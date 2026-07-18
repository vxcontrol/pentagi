package controller

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
)

type fallbackQuerier struct {
	database.Querier

	flow       database.Flow
	flowExists bool
	containers []database.Container
}

func (q *fallbackQuerier) GetFlow(ctx context.Context, id int64) (database.Flow, error) {
	if !q.flowExists || q.flow.ID != id {
		return database.Flow{}, sql.ErrNoRows
	}

	return q.flow, nil
}

func (q *fallbackQuerier) GetFlowContainers(ctx context.Context, flowID int64) ([]database.Container, error) {
	containers := make([]database.Container, 0, len(q.containers))
	for _, container := range q.containers {
		if container.FlowID == flowID {
			containers = append(containers, container)
		}
	}

	return containers, nil
}

func (q *fallbackQuerier) UpdateFlowStatus(ctx context.Context, arg database.UpdateFlowStatusParams) (database.Flow, error) {
	if !q.flowExists || q.flow.ID != arg.ID {
		return database.Flow{}, sql.ErrNoRows
	}

	q.flow.Status = arg.Status

	return q.flow, nil
}

func (q *fallbackQuerier) UpdateContainerStatus(
	ctx context.Context,
	arg database.UpdateContainerStatusParams,
) (database.Container, error) {
	for idx := range q.containers {
		if q.containers[idx].ID == arg.ID {
			q.containers[idx].Status = arg.Status
			return q.containers[idx], nil
		}
	}

	return database.Container{}, sql.ErrNoRows
}

type fallbackDockerClient struct {
	docker.DockerClient

	db      *fallbackQuerier
	removed []string
}

func (c *fallbackDockerClient) RemoveContainer(ctx context.Context, containerID string, dbID int64) error {
	c.removed = append(c.removed, containerID)
	_, err := c.db.UpdateContainerStatus(ctx, database.UpdateContainerStatusParams{
		ID:     dbID,
		Status: database.ContainerStatusDeleted,
	})

	return err
}

func TestFlowControllerFinishFlowFallbackRemovesContainerAndMarksFinished(t *testing.T) {
	ctx := context.Background()
	db := &fallbackQuerier{
		flowExists: true,
		flow: database.Flow{
			ID:     42,
			UserID: 7,
			Status: database.FlowStatusRunning,
		},
		containers: []database.Container{
			{
				ID:      11,
				FlowID:  42,
				Status:  database.ContainerStatusRunning,
				LocalID: database.StringToNullString("docker-11"),
			},
		},
	}
	dockerClient := &fallbackDockerClient{db: db}
	fc := &flowController{
		db:     db,
		mx:     &sync.Mutex{},
		flows:  map[int64]FlowWorker{},
		docker: dockerClient,
	}

	if err := fc.FinishFlow(ctx, 42); err != nil {
		t.Fatalf("FinishFlow returned error: %v", err)
	}

	if db.flow.Status != database.FlowStatusFinished {
		t.Fatalf("expected flow status %q, got %q", database.FlowStatusFinished, db.flow.Status)
	}
	if db.containers[0].Status != database.ContainerStatusDeleted {
		t.Fatalf("expected container status %q, got %q", database.ContainerStatusDeleted, db.containers[0].Status)
	}
	if !reflect.DeepEqual(dockerClient.removed, []string{"docker-11"}) {
		t.Fatalf("expected docker removal for docker-11, got %#v", dockerClient.removed)
	}
}

func TestFlowControllerFinishFlowFallbackMissingFlow(t *testing.T) {
	ctx := context.Background()
	db := &fallbackQuerier{}
	fc := &flowController{
		db:    db,
		mx:    &sync.Mutex{},
		flows: map[int64]FlowWorker{},
	}

	if err := fc.FinishFlow(ctx, 404); !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("expected ErrFlowNotFound, got %v", err)
	}
}

func TestFlowControllerStopFlowFallbackRemovesContainerAndMarksWaiting(t *testing.T) {
	ctx := context.Background()
	db := &fallbackQuerier{
		flowExists: true,
		flow: database.Flow{
			ID:     42,
			UserID: 7,
			Status: database.FlowStatusRunning,
		},
		containers: []database.Container{
			{
				ID:      11,
				FlowID:  42,
				Status:  database.ContainerStatusRunning,
				LocalID: database.StringToNullString("docker-11"),
			},
			{
				ID:      12,
				FlowID:  42,
				Status:  database.ContainerStatusStopped,
				LocalID: database.StringToNullString("docker-12"),
			},
		},
	}
	dockerClient := &fallbackDockerClient{db: db}
	fc := &flowController{
		db:     db,
		mx:     &sync.Mutex{},
		flows:  map[int64]FlowWorker{},
		docker: dockerClient,
	}

	if err := fc.StopFlow(ctx, 42); err != nil {
		t.Fatalf("StopFlow returned error: %v", err)
	}

	if db.flow.Status != database.FlowStatusWaiting {
		t.Fatalf("expected flow status %q, got %q", database.FlowStatusWaiting, db.flow.Status)
	}
	if db.containers[0].Status != database.ContainerStatusDeleted {
		t.Fatalf("expected running container status %q, got %q", database.ContainerStatusDeleted, db.containers[0].Status)
	}
	if db.containers[1].Status != database.ContainerStatusStopped {
		t.Fatalf("expected stopped container status %q, got %q", database.ContainerStatusStopped, db.containers[1].Status)
	}
	if !reflect.DeepEqual(dockerClient.removed, []string{"docker-11"}) {
		t.Fatalf("expected docker removal for docker-11, got %#v", dockerClient.removed)
	}
}

func TestFlowControllerFallbackMarksContainerWithoutLocalIDDeleted(t *testing.T) {
	ctx := context.Background()
	db := &fallbackQuerier{
		flowExists: true,
		flow: database.Flow{
			ID:     42,
			UserID: 7,
			Status: database.FlowStatusRunning,
		},
		containers: []database.Container{
			{
				ID:     11,
				FlowID: 42,
				Status: database.ContainerStatusRunning,
			},
		},
	}
	dockerClient := &fallbackDockerClient{db: db}
	fc := &flowController{
		db:     db,
		mx:     &sync.Mutex{},
		flows:  map[int64]FlowWorker{},
		docker: dockerClient,
	}

	if err := fc.FinishFlow(ctx, 42); err != nil {
		t.Fatalf("FinishFlow returned error: %v", err)
	}

	if db.containers[0].Status != database.ContainerStatusDeleted {
		t.Fatalf("expected container status %q, got %q", database.ContainerStatusDeleted, db.containers[0].Status)
	}
	if len(dockerClient.removed) != 0 {
		t.Fatalf("expected no docker removals, got %#v", dockerClient.removed)
	}
}
