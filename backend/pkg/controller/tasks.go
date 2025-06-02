package controller

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type TaskController interface {
	CreateTask(ctx context.Context, input string, updater FlowUpdater) (TaskWorker, error)
	LoadTasks(ctx context.Context, flowID int64, updater FlowUpdater) error
	ListTasks(ctx context.Context) []TaskWorker
	GetTask(ctx context.Context, taskID int64) (TaskWorker, error)
}

type taskController struct {
	mx      *sync.Mutex
	tasks   map[int64]TaskWorker
	updater FlowUpdater
	flowCtx *FlowContext
}

func NewTaskController(flowCtx *FlowContext) TaskController {
	return &taskController{
		mx:      &sync.Mutex{},
		tasks:   make(map[int64]TaskWorker),
		flowCtx: flowCtx,
	}
}

func (tc *taskController) LoadTasks(
	ctx context.Context,
	flowID int64,
	updater FlowUpdater,
) error {
	tc.mx.Lock()
	defer tc.mx.Unlock()

	tasks, err := tc.flowCtx.DB.GetFlowTasks(ctx, flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow tasks: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found for flow %d: %w", flowID, ErrNothingToLoad)
	}

	for _, task := range tasks {
		tw, err := LoadTaskWorker(ctx, task, tc.flowCtx, updater)
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}

			return fmt.Errorf("failed to load task worker: %w", err)
		}

		tc.tasks[task.ID] = tw
	}

	return nil
}

func (tc *taskController) CreateTask(
	ctx context.Context,
	input string,
	updater FlowUpdater,
) (TaskWorker, error) {
	tc.mx.Lock()
	defer tc.mx.Unlock()

	tw, err := NewTaskWorker(ctx, tc.flowCtx, input, updater)
	if err != nil {
		return nil, fmt.Errorf("failed to create task worker: %w", err)
	}

	tc.tasks[tw.GetTaskID()] = tw

	return tw, nil
}

func (tc *taskController) ListTasks(ctx context.Context) []TaskWorker {
	tc.mx.Lock()
	defer tc.mx.Unlock()

	tasks := make([]TaskWorker, 0)
	for _, task := range tc.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetTaskID() < tasks[j].GetTaskID()
	})

	return tasks
}

func (tc *taskController) GetTask(ctx context.Context, taskID int64) (TaskWorker, error) {
	tc.mx.Lock()
	defer tc.mx.Unlock()

	task, ok := tc.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %d not found", taskID)
	}

	return task, nil
}
