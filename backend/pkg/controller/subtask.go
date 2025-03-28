package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
)

type TaskUpdater interface {
	SetStatus(ctx context.Context, status database.TaskStatus) error
}

type SubtaskWorker interface {
	GetMsgChainID() int64
	GetSubtaskID() int64
	GetTaskID() int64
	GetFlowID() int64
	GetUserID() int64
	GetTitle() string
	GetDescription() string
	IsCompleted() bool
	IsWaiting() bool
	GetStatus(ctx context.Context) (database.SubtaskStatus, error)
	SetStatus(ctx context.Context, status database.SubtaskStatus) error
	GetResult(ctx context.Context) (string, error)
	SetResult(ctx context.Context, result string) error
	PutInput(ctx context.Context, input string) error
	Run(ctx context.Context) error
	Finish(ctx context.Context) error
}

type subtaskWorker struct {
	mx         *sync.RWMutex
	subtaskCtx *SubtaskContext
	updater    TaskUpdater
	completed  bool
	waiting    bool
}

func NewSubtaskWorker(
	ctx context.Context,
	taskCtx *TaskContext,
	id int64,
	title,
	description string,
	updater TaskUpdater,
) (SubtaskWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.NewSubtaskWorker")
	defer span.End()

	msgChainID, err := taskCtx.Provider.PrepareAgentChain(ctx, taskCtx.TaskID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare primary agent chain for subtask %d: %w", id, err)
	}

	return &subtaskWorker{
		mx: &sync.RWMutex{},
		subtaskCtx: &SubtaskContext{
			MsgChainID:         msgChainID,
			SubtaskID:          id,
			SubtaskTitle:       title,
			SubtaskDescription: description,
			TaskContext:        *taskCtx,
		},
		updater:   updater,
		completed: false,
		waiting:   false,
	}, nil
}

func LoadSubtaskWorker(
	ctx context.Context,
	subtask database.Subtask,
	taskCtx *TaskContext,
	updater TaskUpdater,
) (SubtaskWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.LoadSubtaskWorker")
	defer span.End()

	var completed, waiting bool
	switch subtask.Status {
	case database.SubtaskStatusFinished, database.SubtaskStatusFailed:
		completed = true
	case database.SubtaskStatusWaiting:
		waiting = true
	case database.SubtaskStatusRunning:
		var err error
		// if subtask is running, it means that it was not finished by previous run
		// so we need to set it to created and continue from the beginning
		subtask, err = taskCtx.DB.UpdateSubtaskStatus(ctx, database.UpdateSubtaskStatusParams{
			Status: database.SubtaskStatusCreated,
			ID:     subtask.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update subtask %d status to created: %w", subtask.ID, err)
		}
	case database.SubtaskStatusCreated:
		return nil, fmt.Errorf("subtask %d has created yet: %w", subtask.ID, ErrNothingToLoad)
	default:
		return nil, fmt.Errorf("unexpected subtask status: %s", subtask.Status)
	}

	msgChains, err := taskCtx.DB.GetSubtaskPrimaryMsgChains(ctx, database.Int64ToNullInt64(&subtask.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask primary msg chains for subtask %d: %w", subtask.ID, err)
	}

	if len(msgChains) == 0 {
		return nil, fmt.Errorf("subtask %d has no msg chains: %w", subtask.ID, ErrNothingToLoad)
	}

	return &subtaskWorker{
		mx: &sync.RWMutex{},
		subtaskCtx: &SubtaskContext{
			MsgChainID:         msgChains[0].ID,
			SubtaskID:          subtask.ID,
			SubtaskTitle:       subtask.Title,
			SubtaskDescription: subtask.Description,
			TaskContext:        *taskCtx,
		},
		updater:   updater,
		completed: completed,
		waiting:   waiting,
	}, nil
}

func (stw *subtaskWorker) GetMsgChainID() int64 {
	return stw.subtaskCtx.MsgChainID
}

func (stw *subtaskWorker) GetSubtaskID() int64 {
	return stw.subtaskCtx.SubtaskID
}

func (stw *subtaskWorker) GetTaskID() int64 {
	return stw.subtaskCtx.TaskID
}

func (stw *subtaskWorker) GetFlowID() int64 {
	return stw.subtaskCtx.FlowID
}

func (stw *subtaskWorker) GetUserID() int64 {
	return stw.subtaskCtx.UserID
}

func (stw *subtaskWorker) GetTitle() string {
	return stw.subtaskCtx.SubtaskTitle
}

func (stw *subtaskWorker) GetDescription() string {
	return stw.subtaskCtx.SubtaskDescription
}

func (stw *subtaskWorker) IsCompleted() bool {
	stw.mx.RLock()
	defer stw.mx.RUnlock()

	return stw.completed
}

func (stw *subtaskWorker) IsWaiting() bool {
	stw.mx.RLock()
	defer stw.mx.RUnlock()

	return stw.waiting
}

func (stw *subtaskWorker) GetStatus(ctx context.Context) (database.SubtaskStatus, error) {
	subtask, err := stw.subtaskCtx.DB.GetSubtask(ctx, stw.subtaskCtx.SubtaskID)
	if err != nil {
		return database.SubtaskStatusFailed, err
	}

	return subtask.Status, nil
}

func (stw *subtaskWorker) SetStatus(ctx context.Context, status database.SubtaskStatus) error {
	_, err := stw.subtaskCtx.DB.UpdateSubtaskStatus(ctx, database.UpdateSubtaskStatusParams{
		Status: status,
		ID:     stw.subtaskCtx.SubtaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to set subtask %d status: %w", stw.subtaskCtx.SubtaskID, err)
	}

	stw.mx.Lock()
	defer stw.mx.Unlock()

	switch status {
	case database.SubtaskStatusRunning:
		stw.completed = false
		stw.waiting = false
		err = stw.updater.SetStatus(ctx, database.TaskStatusRunning)
	case database.SubtaskStatusWaiting:
		stw.completed = false
		stw.waiting = true
		err = stw.updater.SetStatus(ctx, database.TaskStatusWaiting)
	case database.SubtaskStatusFinished, database.SubtaskStatusFailed:
		stw.completed = true
		stw.waiting = false
		// statuses Finished and Failed will be produced by stack from Run function call
	default:
		// status Created is not possible to set by this call
		return fmt.Errorf("unsupported subtask status: %s", status)
	}
	if err != nil {
		return fmt.Errorf("failed to set task status in back propagation: %w", err)
	}

	return nil
}

func (stw *subtaskWorker) GetResult(ctx context.Context) (string, error) {
	subtask, err := stw.subtaskCtx.DB.GetSubtask(ctx, stw.subtaskCtx.SubtaskID)
	if err != nil {
		return "", err
	}

	return subtask.Result, nil
}

func (stw *subtaskWorker) SetResult(ctx context.Context, result string) error {
	_, err := stw.subtaskCtx.DB.UpdateSubtaskResult(ctx, database.UpdateSubtaskResultParams{
		Result: result,
		ID:     stw.subtaskCtx.SubtaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to set subtask %d result: %w", stw.subtaskCtx.SubtaskID, err)
	}

	return nil
}

func (stw *subtaskWorker) PutInput(ctx context.Context, input string) error {
	if stw.IsCompleted() {
		return fmt.Errorf("subtask has already completed")
	}

	if !stw.IsWaiting() {
		return fmt.Errorf("subtask is not waiting, run first")
	}

	err := stw.subtaskCtx.Provider.PutInputToAgentChain(ctx, stw.subtaskCtx.MsgChainID, input)
	if err != nil {
		return fmt.Errorf("failed to put input for subtask %d: %w", stw.subtaskCtx.SubtaskID, err)
	}

	_, err = stw.subtaskCtx.MsgLog.PutSubtaskMsg(
		ctx,
		database.MsglogTypeInput,
		stw.subtaskCtx.TaskID,
		stw.subtaskCtx.SubtaskID,
		input,
	)
	if err != nil {
		return fmt.Errorf("failed to put input for subtask %d: %w", stw.subtaskCtx.SubtaskID, err)
	}

	stw.mx.Lock()
	defer stw.mx.Unlock()

	stw.waiting = false

	return nil
}

func (stw *subtaskWorker) Run(ctx context.Context) error {
	if stw.IsCompleted() {
		return fmt.Errorf("subtask has already completed")
	}

	if stw.IsWaiting() {
		return fmt.Errorf("subtask is waiting, put input first")
	}

	if err := stw.SetStatus(ctx, database.SubtaskStatusRunning); err != nil {
		return err
	}

	var (
		taskID     = stw.subtaskCtx.TaskID
		subtaskID  = stw.subtaskCtx.SubtaskID
		msgChainID = stw.subtaskCtx.MsgChainID
	)

	performResult, err := stw.subtaskCtx.Provider.PerformAgentChain(ctx, taskID, subtaskID, msgChainID)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			_ = stw.SetStatus(ctx, database.SubtaskStatusWaiting)
		}
		return fmt.Errorf("failed to perform agent chain for subtask %d: %w", subtaskID, err)
	}

	switch performResult {
	case providers.PerformResultWaiting:
		if err := stw.SetStatus(ctx, database.SubtaskStatusWaiting); err != nil {
			return err
		}
	case providers.PerformResultDone:
		if err := stw.SetStatus(ctx, database.SubtaskStatusFinished); err != nil {
			return fmt.Errorf("failed to set subtask %d status to finished: %w", subtaskID, err)
		}
	case providers.PerformResultError:
		if err := stw.SetStatus(ctx, database.SubtaskStatusFailed); err != nil {
			return fmt.Errorf("failed to set subtask %d status to failed: %w", subtaskID, err)
		}
	default:
		return fmt.Errorf("unknown perform result: %d", performResult)
	}

	return nil
}

func (stw *subtaskWorker) Finish(ctx context.Context) error {
	if stw.IsCompleted() {
		return fmt.Errorf("subtask has already completed")
	}

	if err := stw.SetStatus(ctx, database.SubtaskStatusFinished); err != nil {
		return err
	}

	return nil
}
