package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"pentagi/pkg/database"
)

type NewSubtaskInfo struct {
	Title       string
	Description string
}

type SubtaskController interface {
	LoadSubtasks(ctx context.Context, taskID int64, updater TaskUpdater) error
	GenerateSubtasks(ctx context.Context) error
	RefineSubtasks(ctx context.Context) error
	PopSubtask(ctx context.Context, updater TaskUpdater) (SubtaskWorker, error)
	ListSubtasks(ctx context.Context) []SubtaskWorker
	GetSubtask(ctx context.Context, subtaskID int64) (SubtaskWorker, error)
}

type subtaskController struct {
	mx       *sync.Mutex
	taskCtx  *TaskContext
	subtasks map[int64]SubtaskWorker
}

func NewSubtaskController(taskCtx *TaskContext) SubtaskController {
	return &subtaskController{
		mx:       &sync.Mutex{},
		taskCtx:  taskCtx,
		subtasks: make(map[int64]SubtaskWorker),
	}
}

func (stc *subtaskController) LoadSubtasks(ctx context.Context, taskID int64, updater TaskUpdater) error {
	stc.mx.Lock()
	defer stc.mx.Unlock()

	subtasks, err := stc.taskCtx.DB.GetTaskSubtasks(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get subtasks for task %d: %w", taskID, err)
	}

	if len(subtasks) == 0 {
		return fmt.Errorf("no subtasks found for task %d: %w", taskID, ErrNothingToLoad)
	}

	for _, subtask := range subtasks {
		st, err := LoadSubtaskWorker(ctx, subtask, stc.taskCtx, updater)
		if err != nil {
			if errors.Is(err, ErrNothingToLoad) {
				continue
			}

			return fmt.Errorf("failed to create subtask worker: %w", err)
		}

		stc.subtasks[subtask.ID] = st
	}

	return nil
}

func (stc *subtaskController) GenerateSubtasks(ctx context.Context) error {
	plan, err := stc.taskCtx.Provider.GenerateSubtasks(ctx, stc.taskCtx.TaskID)
	if err != nil {
		return fmt.Errorf("failed to generate subtasks for task %d: %w", stc.taskCtx.TaskID, err)
	}

	if len(plan) == 0 {
		return fmt.Errorf("no subtasks generated for task %d", stc.taskCtx.TaskID)
	}

	// TODO: change it to insert subtasks in transaction
	for _, info := range plan {
		_, err := stc.taskCtx.DB.CreateSubtask(ctx, database.CreateSubtaskParams{
			Status:      database.SubtaskStatusCreated,
			TaskID:      stc.taskCtx.TaskID,
			Title:       info.Title,
			Description: info.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to create subtask for task %d: %w", stc.taskCtx.TaskID, err)
		}
	}

	return nil
}

func (stc *subtaskController) RefineSubtasks(ctx context.Context) error {
	subtasks, err := stc.taskCtx.DB.GetTaskSubtasks(ctx, stc.taskCtx.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task %d subtasks: %w", stc.taskCtx.TaskID, err)
	}

	plan, err := stc.taskCtx.Provider.RefineSubtasks(ctx, stc.taskCtx.TaskID)
	if err != nil {
		return fmt.Errorf("failed to refine subtasks for task %d: %w", stc.taskCtx.TaskID, err)
	}

	if len(plan) == 0 {
		return nil // no subtasks refined
	}

	subtaskIDs := make([]int64, 0, len(subtasks))
	for _, subtask := range subtasks {
		if subtask.Status == database.SubtaskStatusCreated {
			subtaskIDs = append(subtaskIDs, subtask.ID)
		}
	}

	err = stc.taskCtx.DB.DeleteSubtasks(ctx, subtaskIDs)
	if err != nil {
		return fmt.Errorf("failed to delete subtasks for task %d: %w", stc.taskCtx.TaskID, err)
	}

	// TODO: change it to insert subtasks in transaction and union it with delete ones
	for _, info := range plan {
		_, err := stc.taskCtx.DB.CreateSubtask(ctx, database.CreateSubtaskParams{
			Status:      database.SubtaskStatusCreated,
			TaskID:      stc.taskCtx.TaskID,
			Title:       info.Title,
			Description: info.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to create subtask for task %d: %w", stc.taskCtx.TaskID, err)
		}
	}

	return nil
}

func (stc *subtaskController) PopSubtask(ctx context.Context, updater TaskUpdater) (SubtaskWorker, error) {
	stc.mx.Lock()
	defer stc.mx.Unlock()

	subtasks, err := stc.taskCtx.DB.GetTaskPlannedSubtasks(ctx, stc.taskCtx.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task planned subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		return nil, nil
	}

	stdb := subtasks[0]
	if st, ok := stc.subtasks[stdb.ID]; ok {
		return st, nil
	}

	st, err := NewSubtaskWorker(ctx, stc.taskCtx, stdb.ID, stdb.Title, stdb.Description, updater)
	if err != nil {
		return nil, fmt.Errorf("failed to create subtask worker: %w", err)
	}

	stc.subtasks[stdb.ID] = st

	return st, nil
}

func (stc *subtaskController) ListSubtasks(ctx context.Context) []SubtaskWorker {
	stc.mx.Lock()
	defer stc.mx.Unlock()

	subtasks := make([]SubtaskWorker, 0)
	for _, subtask := range stc.subtasks {
		subtasks = append(subtasks, subtask)
	}

	return subtasks
}

func (stc *subtaskController) GetSubtask(ctx context.Context, subtaskID int64) (SubtaskWorker, error) {
	stc.mx.Lock()
	defer stc.mx.Unlock()

	subtask, ok := stc.subtasks[subtaskID]
	if !ok {
		return nil, fmt.Errorf("subtask not found")
	}

	return subtask, nil
}
