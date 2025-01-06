package controller

import (
	"context"
	"fmt"
	"sync"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/tools"
)

type FlowUpdater interface {
	SetStatus(ctx context.Context, status database.FlowStatus) error
}

type TaskWorker interface {
	GetTaskID() int64
	GetFlowID() int64
	GetUserID() int64
	GetTitle() string
	IsCompleted() bool
	IsWaiting() bool
	GetStatus(ctx context.Context) (database.TaskStatus, error)
	SetStatus(ctx context.Context, status database.TaskStatus) error
	GetResult(ctx context.Context) (string, error)
	SetResult(ctx context.Context, result string) error
	PutInput(ctx context.Context, input string) error
	Run(ctx context.Context) error
}

type taskWorker struct {
	mx        *sync.RWMutex
	stc       SubtaskController
	taskCtx   *TaskContext
	updater   FlowUpdater
	completed bool
	waiting   bool
}

func NewTaskWorker(
	ctx context.Context,
	flowCtx *FlowContext,
	input string,
	updater FlowUpdater,
) (TaskWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.NewTaskWorker")
	defer span.End()

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypePrimaryAgent)

	title, err := flowCtx.Provider.GetTaskTitle(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get task title: %w", err)
	}

	task, err := flowCtx.DB.CreateTask(ctx, database.CreateTaskParams{
		Status: database.TaskStatusCreated,
		Title:  title,
		Input:  input,
		FlowID: flowCtx.FlowID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task in DB: %w", err)
	}

	flowCtx.Publisher.TaskCreated(ctx, task, []database.Subtask{})

	taskCtx := &TaskContext{
		FlowContext: *flowCtx,
		TaskID:      task.ID,
		TaskTitle:   title,
		TaskInput:   input,
	}
	stc := NewSubtaskController(taskCtx)

	_, err = taskCtx.MsgLog.PutTaskMsg(
		ctx,
		database.MsglogTypeInput,
		taskCtx.TaskID,
		input,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to put input for task %d: %w", taskCtx.TaskID, err)
	}

	err = stc.GenerateSubtasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate subtasks: %w", err)
	}

	subtasks, err := flowCtx.DB.GetTaskSubtasks(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks for task %d: %w", task.ID, err)
	}

	flowCtx.Publisher.TaskUpdated(ctx, task, subtasks)

	return &taskWorker{
		mx:        &sync.RWMutex{},
		stc:       stc,
		taskCtx:   taskCtx,
		updater:   updater,
		completed: false,
		waiting:   false,
	}, nil
}

func LoadTaskWorker(
	ctx context.Context,
	task database.Task,
	flowCtx *FlowContext,
	updater FlowUpdater,
) (TaskWorker, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "controller.LoadTaskWorker")
	defer span.End()

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypePrimaryAgent)
	taskCtx := &TaskContext{
		FlowContext: *flowCtx,
		TaskID:      task.ID,
		TaskTitle:   task.Title,
		TaskInput:   task.Input,
	}

	stc := NewSubtaskController(taskCtx)
	var completed, waiting bool
	switch task.Status {
	case database.TaskStatusFinished, database.TaskStatusFailed:
		completed = true
	case database.TaskStatusWaiting:
		waiting = true
	case database.TaskStatusRunning:
	case database.TaskStatusCreated:
		return nil, fmt.Errorf("task %d has created yet: loading aborted: %w", task.ID, ErrNothingToLoad)
	}

	tw := &taskWorker{
		mx:        &sync.RWMutex{},
		stc:       stc,
		taskCtx:   taskCtx,
		updater:   updater,
		completed: completed,
		waiting:   waiting,
	}

	if err := tw.stc.LoadSubtasks(ctx, task.ID, tw); err != nil {
		return nil, fmt.Errorf("failed to load subtasks for task %d: %w", task.ID, err)
	}

	return tw, nil
}

func (tw *taskWorker) GetTaskID() int64 {
	return tw.taskCtx.TaskID
}

func (tw *taskWorker) GetFlowID() int64 {
	return tw.taskCtx.FlowID
}

func (tw *taskWorker) GetUserID() int64 {
	return tw.taskCtx.UserID
}

func (tw *taskWorker) GetTitle() string {
	return tw.taskCtx.TaskTitle
}

func (tw *taskWorker) IsCompleted() bool {
	tw.mx.RLock()
	defer tw.mx.RUnlock()

	return tw.completed
}

func (tw *taskWorker) IsWaiting() bool {
	tw.mx.RLock()
	defer tw.mx.RUnlock()

	return tw.waiting
}

func (tw *taskWorker) GetStatus(ctx context.Context) (database.TaskStatus, error) {
	task, err := tw.taskCtx.DB.GetTask(ctx, tw.taskCtx.TaskID)
	if err != nil {
		return database.TaskStatusFailed, err
	}

	return task.Status, nil
}

// this function is exclusively change task internal properties "completed" and "waiting"
func (tw *taskWorker) SetStatus(ctx context.Context, status database.TaskStatus) error {
	task, err := tw.taskCtx.DB.UpdateTaskStatus(ctx, database.UpdateTaskStatusParams{
		Status: status,
		ID:     tw.taskCtx.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to set task %d status: %w", tw.taskCtx.TaskID, err)
	}

	subtasks, err := tw.taskCtx.DB.GetTaskSubtasks(ctx, tw.taskCtx.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task %d subtasks: %w", tw.taskCtx.TaskID, err)
	}

	tw.taskCtx.Publisher.TaskUpdated(ctx, task, subtasks)

	tw.mx.Lock()
	defer tw.mx.Unlock()

	switch status {
	case database.TaskStatusRunning:
		tw.completed = false
		tw.waiting = false
		err = tw.updater.SetStatus(ctx, database.FlowStatusRunning)
	case database.TaskStatusWaiting:
		tw.completed = false
		tw.waiting = true
		err = tw.updater.SetStatus(ctx, database.FlowStatusWaiting)
	case database.TaskStatusFinished, database.TaskStatusFailed:
		tw.completed = true
		tw.waiting = false
		// the last task was done, set flow status to Waiting new user input
		err = tw.updater.SetStatus(ctx, database.FlowStatusWaiting)
	default:
		// status Created is not possible to set by this call
		return fmt.Errorf("unsupported task status: %s", status)
	}
	if err != nil {
		return fmt.Errorf("failed to set flow status in back propagation: %w", err)
	}

	return nil
}

func (tw *taskWorker) GetResult(ctx context.Context) (string, error) {
	task, err := tw.taskCtx.DB.GetTask(ctx, tw.taskCtx.TaskID)
	if err != nil {
		return "", err
	}

	return task.Result, nil
}

func (tw *taskWorker) SetResult(ctx context.Context, result string) error {
	_, err := tw.taskCtx.DB.UpdateTaskResult(ctx, database.UpdateTaskResultParams{
		Result: result,
		ID:     tw.taskCtx.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to set task %d result: %w", tw.taskCtx.TaskID, err)
	}

	return nil
}

func (tw *taskWorker) PutInput(ctx context.Context, input string) error {
	if !tw.IsWaiting() {
		return fmt.Errorf("task is not waiting")
	}

	for _, st := range tw.stc.ListSubtasks(ctx) {
		if !st.IsCompleted() && st.IsWaiting() {
			if err := st.PutInput(ctx, input); err != nil {
				return fmt.Errorf("failed to put input to subtask %d: %w", st.GetSubtaskID(), err)
			} else {
				break
			}
		}
	}

	return nil
}

func (tw *taskWorker) Run(ctx context.Context) error {
	ctx = tools.PutAgentContext(ctx, database.MsgchainTypePrimaryAgent)

	for len(tw.stc.ListSubtasks(ctx)) < providers.TasksNumberLimit+3 {
		st, err := tw.stc.PopSubtask(ctx, tw)
		if err != nil {
			return err
		}

		// empty queue for subtasks means that task is done
		if st == nil {
			break
		}

		if err := st.Run(ctx); err != nil {
			return err
		}

		// pass through if task is waiting from back status propagation
		if tw.IsWaiting() {
			return nil
		} // otherwise subtask is done

		if err := tw.stc.RefineSubtasks(ctx); err != nil {
			return err
		}
	}

	jobResult, err := tw.taskCtx.Provider.GetTaskResult(ctx, tw.taskCtx.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task %d result: %w", tw.taskCtx.TaskID, err)
	}

	var taskStatus database.TaskStatus
	if jobResult.Success {
		taskStatus = database.TaskStatusFinished
	} else {
		taskStatus = database.TaskStatusFailed
	}

	if err := tw.SetResult(ctx, jobResult.Result); err != nil {
		return err
	}

	if err := tw.SetStatus(ctx, taskStatus); err != nil {
		return err
	}

	return nil
}
