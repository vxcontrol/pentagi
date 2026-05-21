package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
)

// mockQuerier satisfies database.Querier by embedding the interface;
// only the methods used by flow_manager are overridden via function fields.
type mockQuerier struct {
	database.Querier
	getFlowTasksFn        func(ctx context.Context, flowID int64) ([]database.Task, error)
	getFlowSubtasksFn     func(ctx context.Context, flowID int64) ([]database.Subtask, error)
	getFlowTaskSubtasksFn func(ctx context.Context, arg database.GetFlowTaskSubtasksParams) ([]database.Subtask, error)
	getTaskPlannedFn      func(ctx context.Context, taskID int64) ([]database.Subtask, error)
	getSubtaskMsgLogsFn   func(ctx context.Context, subtaskID sql.NullInt64) ([]database.Msglog, error)
}

func (m *mockQuerier) GetFlowTasks(ctx context.Context, flowID int64) ([]database.Task, error) {
	if m.getFlowTasksFn != nil {
		return m.getFlowTasksFn(ctx, flowID)
	}
	return nil, nil
}

func (m *mockQuerier) GetFlowSubtasks(ctx context.Context, flowID int64) ([]database.Subtask, error) {
	if m.getFlowSubtasksFn != nil {
		return m.getFlowSubtasksFn(ctx, flowID)
	}
	return nil, nil
}

func (m *mockQuerier) GetFlowTaskSubtasks(ctx context.Context, arg database.GetFlowTaskSubtasksParams) ([]database.Subtask, error) {
	if m.getFlowTaskSubtasksFn != nil {
		return m.getFlowTaskSubtasksFn(ctx, arg)
	}
	return nil, nil
}

func (m *mockQuerier) GetTaskPlannedSubtasks(ctx context.Context, taskID int64) ([]database.Subtask, error) {
	if m.getTaskPlannedFn != nil {
		return m.getTaskPlannedFn(ctx, taskID)
	}
	return nil, nil
}

func (m *mockQuerier) GetSubtaskMsgLogs(ctx context.Context, subtaskID sql.NullInt64) ([]database.Msglog, error) {
	if m.getSubtaskMsgLogsFn != nil {
		return m.getSubtaskMsgLogsFn(ctx, subtaskID)
	}
	return nil, nil
}

// ------------------------------------------------------------------
// Helpers

func makeTasks(statuses ...database.TaskStatus) []database.Task {
	tasks := make([]database.Task, len(statuses))
	for i, s := range statuses {
		tasks[i] = database.Task{ID: int64(i + 1), Status: s, Title: fmt.Sprintf("task-%d", i+1)}
	}
	return tasks
}

func makeSubtasks(statuses ...database.SubtaskStatus) []database.Subtask {
	subs := make([]database.Subtask, len(statuses))
	for i, s := range statuses {
		subs[i] = database.Subtask{ID: int64(i + 1), Status: s, Title: fmt.Sprintf("subtask-%d", i+1), TaskID: 1}
	}
	return subs
}

// ------------------------------------------------------------------
// Pure function tests

func TestInferFlowStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		statuses []database.TaskStatus
		want     string
	}{
		{
			name: "no tasks",
			want: "waiting (ready for next input)",
		},
		{
			name:     "running task takes priority",
			statuses: []database.TaskStatus{database.TaskStatusWaiting, database.TaskStatusRunning, database.TaskStatusFinished},
			want:     "running",
		},
		{
			name:     "waiting task, no running",
			statuses: []database.TaskStatus{database.TaskStatusFinished, database.TaskStatusWaiting},
			want:     "waiting (subtask asking for user input)",
		},
		{
			name:     "all finished",
			statuses: []database.TaskStatus{database.TaskStatusFinished, database.TaskStatusFinished},
			want:     "waiting (ready for next input)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tasks := makeTasks(tt.statuses...)
			got := inferFlowStatus(tasks)
			if got != tt.want {
				t.Errorf("inferFlowStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "shorter than max", input: "hello", maxLen: 10, want: "hello"},
		{name: "exactly max", input: "hello", maxLen: 5, want: "hello"},
		{name: "longer than max — appends ellipsis", input: "hello world", maxLen: 5, want: "hello..."},
		{name: "empty string", input: "", maxLen: 5, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncateText(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------
// flowStatusTool.Handle dispatch

func TestFlowStatusToolHandle(t *testing.T) {
	t.Parallel()

	db := &mockQuerier{
		getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
		getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, nil },
	}
	tool := NewFlowStatusTool(1, db, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		args    string
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name:    "invalid JSON returns error",
			args:    `{bad json}`,
			wantErr: true,
		},
		{
			name:    "unknown detail returns error",
			args:    `{"detail":"nonexistent","message":"x"}`,
			wantErr: true,
		},
		{
			name: "summary detail dispatches correctly",
			args: `{"detail":"summary","message":"x"}`,
			check: func(t *testing.T, result string) {
				t.Helper()
				if !strings.Contains(result, "Flow ID:") {
					t.Errorf("summary result should contain 'Flow ID:', got: %s", result)
				}
			},
		},
		{
			name: "tasks detail dispatches correctly",
			args: `{"detail":"tasks","message":"x"}`,
			check: func(t *testing.T, result string) {
				t.Helper()
				if result == "" {
					t.Error("tasks result should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := tool.Handle(ctx, "get_flow_status", json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// ------------------------------------------------------------------
// buildSummary

func TestBuildSummary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("db error on GetFlowTasks", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
			return nil, errors.New("db down")
		}}
		_, err := NewFlowStatusTool(1, db, nil).buildSummary(ctx, false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("db error on GetFlowSubtasks", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, errors.New("db down") },
		}
		_, err := NewFlowStatusTool(1, db, nil).buildSummary(ctx, false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty tasks emits no-tasks notice", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, nil },
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSummary(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No tasks yet") {
			t.Errorf("expected 'No tasks yet' in result, got: %s", result)
		}
	})

	t.Run("status counts are correct", func(t *testing.T) {
		t.Parallel()
		tasks := makeTasks(
			database.TaskStatusRunning,
			database.TaskStatusWaiting,
			database.TaskStatusFinished,
			database.TaskStatusFailed,
			database.TaskStatusCreated,
		)
		subs := makeSubtasks(
			database.SubtaskStatusRunning,
			database.SubtaskStatusFinished,
			database.SubtaskStatusFinished,
		)
		db := &mockQuerier{
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return tasks, nil },
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSummary(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "total: 5") {
			t.Errorf("expected 'total: 5' tasks in result, got: %s", result)
		}
		if !strings.Contains(result, "total: 3") {
			t.Errorf("expected 'total: 3' subtasks in result, got: %s", result)
		}
	})

	t.Run("active task identified and shown", func(t *testing.T) {
		t.Parallel()
		tasks := makeTasks(database.TaskStatusFinished, database.TaskStatusRunning)
		db := &mockQuerier{
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return tasks, nil },
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, nil },
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSummary(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "Active task:") {
			t.Errorf("expected 'Active task:' in result, got: %s", result)
		}
		if !strings.Contains(result, "task-2") {
			t.Errorf("expected active task 'task-2' in result, got: %s", result)
		}
	})

	t.Run("summarizer called when output exceeds summaryLimit", func(t *testing.T) {
		t.Parallel()
		largeTasks := make([]database.Task, 1)
		largeTasks[0] = database.Task{ID: 1, Status: database.TaskStatusRunning, Title: strings.Repeat("x", summaryLimit+1)}
		db := &mockQuerier{
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return largeTasks, nil },
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, nil },
		}
		summarizerCalled := false
		summarizer := func(_ context.Context, _ string) (string, error) {
			summarizerCalled = true
			return "summarized", nil
		}
		result, err := NewFlowStatusTool(1, db, summarizer).buildSummary(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !summarizerCalled {
			t.Error("expected summarizer to be called, but it was not")
		}
		if result != "summarized" {
			t.Errorf("expected summarizer output, got: %s", result)
		}
	})
}

// ------------------------------------------------------------------
// buildTasksList

func TestBuildTasksList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name       string
		tasks      []database.Task
		tasksErr   error
		verbose    bool
		wantErr    bool
		wantSubstr string
	}{
		{
			name:       "no tasks returns specific message",
			wantSubstr: "No tasks found",
		},
		{
			name:     "db error propagated",
			tasksErr: errors.New("db fail"),
			wantErr:  true,
		},
		{
			name:       "tasks listed with IDs and titles",
			tasks:      makeTasks(database.TaskStatusFinished, database.TaskStatusRunning),
			wantSubstr: "task-1",
		},
		{
			name:    "verbose: result field included when set",
			verbose: true,
			tasks: []database.Task{
				{ID: 1, Status: database.TaskStatusFinished, Title: "t1", Result: "my-result"},
			},
			wantSubstr: "my-result",
		},
		{
			name:    "non-verbose: result field NOT included",
			verbose: false,
			tasks: []database.Task{
				{ID: 1, Status: database.TaskStatusFinished, Title: "t1", Result: "secret-result"},
			},
			wantSubstr: "", // will be checked via absence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := &mockQuerier{getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
				return tt.tasks, tt.tasksErr
			}}
			result, err := NewFlowStatusTool(1, db, nil).buildTasksList(ctx, tt.verbose)
			if (err != nil) != tt.wantErr {
				t.Fatalf("buildTasksList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.wantSubstr != "" && !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
			// non-verbose check: result field must not appear
			if !tt.verbose && len(tt.tasks) > 0 && tt.tasks[0].Result != "" {
				if strings.Contains(result, tt.tasks[0].Result) {
					t.Errorf("non-verbose mode should not include result field, got: %s", result)
				}
			}
		})
	}
}

// ------------------------------------------------------------------
// buildSubtasksList

func TestBuildSubtasksList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("without taskID uses GetFlowSubtasks", func(t *testing.T) {
		t.Parallel()
		flowSubtasksCalled := false
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
				flowSubtasksCalled = true
				return makeSubtasks(database.SubtaskStatusFinished), nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSubtasksList(ctx, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !flowSubtasksCalled {
			t.Error("expected GetFlowSubtasks to be called")
		}
		if !strings.Contains(result, "All subtasks for flow") {
			t.Errorf("expected flow-level header, got: %s", result)
		}
	})

	t.Run("with valid taskID uses GetFlowTaskSubtasks", func(t *testing.T) {
		t.Parallel()
		taskSubtasksCalled := false
		taskID := int64(42)
		db := &mockQuerier{
			getFlowTaskSubtasksFn: func(_ context.Context, arg database.GetFlowTaskSubtasksParams) ([]database.Subtask, error) {
				taskSubtasksCalled = true
				if arg.TaskID != taskID {
					return nil, fmt.Errorf("unexpected taskID %d", arg.TaskID)
				}
				return makeSubtasks(database.SubtaskStatusRunning), nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSubtasksList(ctx, &taskID, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !taskSubtasksCalled {
			t.Error("expected GetFlowTaskSubtasks to be called")
		}
		if !strings.Contains(result, "Subtasks for task 42") {
			t.Errorf("expected task-level header, got: %s", result)
		}
	})

	t.Run("empty subtasks returns no-subtasks message", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return nil, nil },
		}
		result, err := NewFlowStatusTool(1, db, nil).buildSubtasksList(ctx, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No subtasks found") {
			t.Errorf("expected 'No subtasks found', got: %s", result)
		}
	})

	t.Run("db error propagated", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
				return nil, errors.New("db fail")
			},
		}
		_, err := NewFlowStatusTool(1, db, nil).buildSubtasksList(ctx, nil, false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// ------------------------------------------------------------------
// buildRunningInfo

func TestBuildRunningInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("no active subtasks — flow is idle", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
				return makeSubtasks(database.SubtaskStatusFinished, database.SubtaskStatusFailed), nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildRunningInfo(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No running or waiting subtask found") {
			t.Errorf("expected idle message, got: %s", result)
		}
	})

	t.Run("running subtask shows parent task chain", func(t *testing.T) {
		t.Parallel()
		tasks := []database.Task{
			{ID: 1, Status: database.TaskStatusRunning, Title: "main-task"},
		}
		subs := []database.Subtask{
			{ID: 10, Status: database.SubtaskStatusRunning, Title: "active-sub", TaskID: 1},
		}
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return tasks, nil },
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildRunningInfo(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "Active Task") {
			t.Errorf("expected 'Active Task' section, got: %s", result)
		}
		if !strings.Contains(result, "main-task") {
			t.Errorf("expected parent task title 'main-task', got: %s", result)
		}
		if !strings.Contains(result, "active-sub") {
			t.Errorf("expected subtask title 'active-sub', got: %s", result)
		}
	})

	t.Run("waiting subtask shows ask note and submit tool name", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{
			{ID: 5, Status: database.SubtaskStatusWaiting, Title: "ask-sub", TaskID: 2},
		}
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildRunningInfo(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "waiting for user input") {
			t.Errorf("expected waiting note, got: %s", result)
		}
		if !strings.Contains(result, SubmitFlowInputToolName) {
			t.Errorf("expected SubmitFlowInputToolName in result, got: %s", result)
		}
	})
	t.Run("verbose shows result and execution context", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{
			{ID: 7, Status: database.SubtaskStatusRunning, Title: "verbose-sub", TaskID: 1,
				Result: "partial-result", Context: "exec-context-data"},
		}
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildRunningInfo(ctx, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "partial-result") {
			t.Errorf("verbose: expected subtask result in output, got: %s", result)
		}
		if !strings.Contains(result, "exec-context-data") {
			t.Errorf("verbose: expected execution context in output, got: %s", result)
		}
	})

	t.Run("non-verbose hides execution context", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{
			{ID: 8, Status: database.SubtaskStatusRunning, Title: "quiet-sub", TaskID: 1,
				Context: "secret-context"},
		}
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
			getFlowTasksFn:    func(_ context.Context, _ int64) ([]database.Task, error) { return nil, nil },
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildRunningInfo(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result, "secret-context") {
			t.Errorf("non-verbose: execution context should be hidden, got: %s", result)
		}
	})
}

func TestBuildPlannedList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("with taskID uses GetTaskPlannedSubtasks", func(t *testing.T) {
		t.Parallel()
		plannedCalled := false
		taskID := int64(7)
		db := &mockQuerier{
			getTaskPlannedFn: func(_ context.Context, id int64) ([]database.Subtask, error) {
				plannedCalled = true
				if id != taskID {
					return nil, fmt.Errorf("unexpected taskID %d", id)
				}
				return makeSubtasks(database.SubtaskStatusCreated), nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildPlannedList(ctx, &taskID, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !plannedCalled {
			t.Error("expected GetTaskPlannedSubtasks to be called")
		}
		if !strings.Contains(result, "Planned subtasks for task 7") {
			t.Errorf("expected task-level header, got: %s", result)
		}
	})

	t.Run("without taskID filters created subtasks from flow", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{
			{ID: 1, Status: database.SubtaskStatusCreated, Title: "planned-sub", TaskID: 1},
			{ID: 2, Status: database.SubtaskStatusFinished, Title: "done-sub", TaskID: 1},
		}
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) { return subs, nil },
		}
		result, err := NewFlowStatusTool(1, db, nil).buildPlannedList(ctx, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "planned-sub") {
			t.Errorf("expected created subtask 'planned-sub', got: %s", result)
		}
		if strings.Contains(result, "done-sub") {
			t.Errorf("finished subtask 'done-sub' should not appear in planned list, got: %s", result)
		}
	})

	t.Run("all subtasks executed returns specific message", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
				return makeSubtasks(database.SubtaskStatusFinished, database.SubtaskStatusFailed), nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).buildPlannedList(ctx, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No planned") {
			t.Errorf("expected 'No planned' message, got: %s", result)
		}
	})
}

// ------------------------------------------------------------------
// waitForTaskReady — context cancellation

func TestWaitForTaskReadyContextCancelled(t *testing.T) {
	t.Parallel()

	db := &mockQuerier{
		getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
			// Always return no running task so the loop keeps spinning.
			return makeTasks(database.TaskStatusCreated), nil
		},
	}
	tool := NewSubmitFlowInputTool(1, db, nil)
	tool.pollInterval = 1 * time.Millisecond
	tool.pollTimeout = 1 * time.Hour // would hang without ctx cancellation

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately so the first ticker tick triggers ctx.Done().
	cancel()

	_, err := tool.waitForTaskReady(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("expected 'context cancelled' in error, got: %v", err)
	}
}

// ------------------------------------------------------------------
// appendSubtaskMsgLogs

func TestAppendSubtaskMsgLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("db error returns error", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, errors.New("db down")
			},
		}
		_, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 5, 10)
		if err == nil {
			t.Fatal("expected error on DB failure, got nil")
		}
	})

	t.Run("no logs returns specific message", func(t *testing.T) {
		t.Parallel()
		db := &mockQuerier{
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return nil, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No agent messages") {
			t.Errorf("expected 'No agent messages', got: %s", result)
		}
	})

	t.Run("fewer logs than limit — all shown", func(t *testing.T) {
		t.Parallel()
		logs := []database.Msglog{
			{ID: 1, Message: "msg-alpha"},
			{ID: 2, Message: "msg-beta"},
		}
		db := &mockQuerier{
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return logs, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "msg-alpha") || !strings.Contains(result, "msg-beta") {
			t.Errorf("expected all messages in result, got: %s", result)
		}
	})

	t.Run("more logs than limit — only tail shown", func(t *testing.T) {
		t.Parallel()
		logs := make([]database.Msglog, 5)
		for i := range logs {
			logs[i] = database.Msglog{ID: int64(i + 1), Message: fmt.Sprintf("msg-%d", i+1)}
		}
		db := &mockQuerier{
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return logs, nil
			},
		}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Only last 2 (msg-4, msg-5) should appear; msg-1 through msg-3 should not.
		for _, absent := range []string{"msg-1", "msg-2", "msg-3"} {
			if strings.Contains(result, absent) {
				t.Errorf("old message %q should have been dropped by limit, got: %s", absent, result)
			}
		}
		for _, present := range []string{"msg-4", "msg-5"} {
			if !strings.Contains(result, present) {
				t.Errorf("recent message %q should appear in result, got: %s", present, result)
			}
		}
	})

	t.Run("summarizer called when logs exceed msgLogsLimit", func(t *testing.T) {
		t.Parallel()
		// Each message is capped to 1024 chars in the output; to exceed msgLogsLimit (16 KB)
		// we need enough log entries so their combined formatted size exceeds the limit.
		numLogs := msgLogsLimit/1024 + 2
		logs := make([]database.Msglog, numLogs)
		for i := range logs {
			logs[i] = database.Msglog{ID: int64(i + 1), Message: strings.Repeat("x", 1024)}
		}
		db := &mockQuerier{
			getSubtaskMsgLogsFn: func(_ context.Context, _ sql.NullInt64) ([]database.Msglog, error) {
				return logs, nil
			},
		}
		summarizerCalled := false
		summarizer := func(_ context.Context, _ string) (string, error) {
			summarizerCalled = true
			return "summarized-logs", nil
		}
		result, err := NewFlowStatusTool(1, db, summarizer).appendSubtaskMsgLogs(ctx, 1, numLogs+10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !summarizerCalled {
			t.Error("expected summarizer to be called for oversized logs output")
		}
		if result != "summarized-logs" {
			t.Errorf("expected summarizer output, got: %s", result)
		}
	})
}

// ------------------------------------------------------------------
// Text helper methods

func TestGetTextHelpers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tool := NewFlowStatusTool(1, &mockQuerier{}, nil)

	tests := []struct {
		name     string
		call     func(string) (string, error)
		limit    int
		twoLimit int
	}{
		{
			name:     "getInputText",
			call:     func(s string) (string, error) { return tool.getInputText(ctx, s) },
			limit:    inputLimit,
			twoLimit: 2 * inputLimit,
		},
		{
			name:     "getDescriptionText",
			call:     func(s string) (string, error) { return tool.getDescriptionText(ctx, s) },
			limit:    descriptionLimit,
			twoLimit: 2 * descriptionLimit,
		},
		{
			name:     "getResultText",
			call:     func(s string) (string, error) { return tool.getResultText(ctx, s) },
			limit:    resultLimit,
			twoLimit: 2 * resultLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/empty returns empty", func(t *testing.T) {
			t.Parallel()
			got, err := tt.call("")
			if err != nil || got != "" {
				t.Errorf("empty input: got (%q, %v), want (\"\", nil)", got, err)
			}
		})

		t.Run(tt.name+"/short text truncated to limit", func(t *testing.T) {
			t.Parallel()
			text := strings.Repeat("a", tt.limit+10)
			got, err := tt.call(text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) > tt.limit+3 { // +3 for "..."
				t.Errorf("expected text to be truncated to ~%d chars, got len=%d", tt.limit, len(got))
			}
		})

		t.Run(tt.name+"/oversized text triggers summarizer", func(t *testing.T) {
			t.Parallel()
			summarizerCalled := false
			toolWithSummarizer := NewFlowStatusTool(1, &mockQuerier{}, func(_ context.Context, _ string) (string, error) {
				summarizerCalled = true
				return "summarized", nil
			})
			text := strings.Repeat("b", tt.twoLimit+1)
			// Call the correct method on the tool with summarizer
			var got string
			var err error
			switch tt.name {
			case "getInputText":
				got, err = toolWithSummarizer.getInputText(ctx, text)
			case "getDescriptionText":
				got, err = toolWithSummarizer.getDescriptionText(ctx, text)
			case "getResultText":
				got, err = toolWithSummarizer.getResultText(ctx, text)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !summarizerCalled {
				t.Error("expected summarizer to be called for oversized text")
			}
			if got != "summarized" {
				t.Errorf("expected summarizer output, got: %s", got)
			}
		})
	}
}

// ------------------------------------------------------------------
// stopFlowTool.Handle

func TestStopFlowToolHandle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name          string
		args          string
		tasks         []database.Task
		tasksErr      error
		handler       func(context.Context, string) error
		newTasks      []database.Task
		newTaskErr    error
		wantErr       bool
		wantSubstr    string
		wantErrSubstr string
	}{
		{
			name:    "invalid JSON returns error",
			args:    `{bad}`,
			wantErr: true,
		},
		{
			name:       "no running tasks — already waiting",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusFinished, database.TaskStatusWaiting),
			wantSubstr: "already in 'waiting' state",
		},
		{
			name:       "running task — handler succeeds — confirms stopped",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusRunning),
			handler:    func(_ context.Context, _ string) error { return nil },
			newTasks:   makeTasks(database.TaskStatusFinished),
			wantSubstr: "stopped successfully",
		},
		{
			name:    "running task — handler error — propagated",
			args:    `{"reason":"test","message":"x"}`,
			tasks:   makeTasks(database.TaskStatusRunning),
			handler: func(_ context.Context, _ string) error { return errors.New("stop failed") },
			wantErr: true,
		},
		{
			name:  "running task — handler deadline exceeded — timeout message",
			args:  `{"reason":"test","message":"x"}`,
			tasks: makeTasks(database.TaskStatusRunning),
			handler: func(_ context.Context, _ string) error {
				return context.DeadlineExceeded
			},
			wantErr:       true,
			wantErrSubstr: "timed out",
		},
		{
			name:  "running task — handler context cancelled — timeout message",
			args:  `{"reason":"test","message":"x"}`,
			tasks: makeTasks(database.TaskStatusRunning),
			handler: func(_ context.Context, _ string) error {
				return context.Canceled
			},
			wantErr:       true,
			wantErrSubstr: "timed out",
		},
		{
			name:       "running task — after stop task still running — warns",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusRunning),
			handler:    func(_ context.Context, _ string) error { return nil },
			newTasks:   makeTasks(database.TaskStatusRunning),
			wantSubstr: "has not reached 'waiting'",
		},
		{
			name:       "stop succeeds but re-query fails — partial message returned",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusRunning),
			handler:    func(_ context.Context, _ string) error { return nil },
			newTaskErr: errors.New("db fail"),
			wantSubstr: "Could not verify new status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			callCount := 0
			db := &mockQuerier{
				getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
					callCount++
					if callCount == 1 {
						return tt.tasks, tt.tasksErr
					}
					return tt.newTasks, tt.newTaskErr
				},
			}
			tool := NewStopFlowTool(1, db, tt.handler)
			result, err := tool.Handle(ctx, "stop_flow", json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrSubstr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr)) {
					t.Errorf("expected %q in error, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if tt.wantSubstr != "" && !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
		})
	}
}

// ------------------------------------------------------------------
// submitFlowInputTool.Handle

func TestSubmitFlowInputToolHandle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name          string
		args          string
		tasks         []database.Task
		subtasks      []database.Subtask
		handler       func(context.Context, string) error
		newTasks      []database.Task
		wantErr       bool
		wantSubstr    string
		wantErrSubstr string
	}{
		{
			name:    "invalid JSON returns error",
			args:    `{bad}`,
			wantErr: true,
		},
		{
			name:    "empty input returns error",
			args:    `{"input":"","message":"x"}`,
			wantErr: true,
		},
		{
			name:       "running task blocks submission — returns soft guidance, not error",
			args:       `{"input":"hello","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusRunning),
			wantSubstr: StopFlowToolName,
		},
		{
			name:       "waiting subtask — delivered as ask answer",
			args:       `{"input":"my answer","message":"x"}`,
			tasks:      makeTasks(database.TaskStatusWaiting),
			subtasks:   makeSubtasks(database.SubtaskStatusWaiting),
			handler:    func(_ context.Context, _ string) error { return nil },
			wantSubstr: "answer to the waiting subtask",
		},
		{
			name:    "idle flow — handler error propagated",
			args:    `{"input":"do something","message":"x"}`,
			handler: func(_ context.Context, _ string) error { return errors.New("submit failed") },
			wantErr: true,
		},
		{
			name: "idle flow — stale message chain after restart — actionable error",
			args: `{"input":"do something","message":"x"}`,
			handler: func(_ context.Context, _ string) error {
				return fmt.Errorf("failed to put flow input: failed to get message chain 63247: sql: no rows in result set")
			},
			wantErr:       true,
			wantErrSubstr: "no longer available in the database",
		},
		{
			name: "idle flow — handler deadline exceeded — timeout message",
			args: `{"input":"do something","message":"x"}`,
			handler: func(_ context.Context, _ string) error {
				return context.DeadlineExceeded
			},
			wantErr:       true,
			wantErrSubstr: "timed out",
		},
		{
			name: "idle flow — handler context cancelled — timeout message",
			args: `{"input":"do something","message":"x"}`,
			handler: func(_ context.Context, _ string) error {
				return context.Canceled
			},
			wantErr:       true,
			wantErrSubstr: "timed out",
		},
		{
			name:       "idle flow — new task now running",
			args:       `{"input":"do something","message":"x"}`,
			handler:    func(_ context.Context, _ string) error { return nil },
			newTasks:   []database.Task{{ID: 5, Status: database.TaskStatusRunning, Title: "new-task"}},
			wantSubstr: "new-task",
		},
		{
			name:       "idle flow — no running task after submit",
			args:       `{"input":"do something","message":"x"}`,
			handler:    func(_ context.Context, _ string) error { return nil },
			newTasks:   makeTasks(database.TaskStatusCreated),
			wantSubstr: "generator may still be working",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			taskCallCount := 0
			db := &mockQuerier{
				getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
					taskCallCount++
					if taskCallCount == 1 {
						return tt.tasks, nil
					}
					return tt.newTasks, nil
				},
				getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
					return tt.subtasks, nil
				},
			}
			tool := NewSubmitFlowInputTool(1, db, tt.handler)
			// Shorten polling so tests don't hang.
			tool.pollInterval = 1 * time.Millisecond
			tool.pollTimeout = 20 * time.Millisecond

			result, err := tool.Handle(ctx, "submit_flow_input", json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrSubstr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr)) {
					t.Errorf("expected %q in error, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if tt.wantSubstr != "" && !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
		})
	}
}

// ------------------------------------------------------------------
// patchFlowSubtasksTool.Handle

func TestPatchFlowSubtasksToolHandle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	validOp := SubtaskOperation{Op: SubtaskOpAdd, Title: "new sub", Description: "do thing"}
	invalidOp := SubtaskOperation{Op: SubtaskOpRemove} // remove without ID

	tests := []struct {
		name        string
		args        string
		tasks       []database.Task
		planned     []database.Subtask
		allSubtasks []database.Subtask
		plannedErr  error
		handler     func(context.Context, int64, SubtaskPatch) error
		newPlanned  []database.Subtask
		wantErr     bool
		wantSubstr  string
	}{
		{
			name:    "invalid JSON returns error",
			args:    `{bad}`,
			wantErr: true,
		},
		{
			name:    "task_id zero returns error",
			args:    `{"task_id":0,"operations":[],"message":"x"}`,
			wantErr: true,
		},
		{
			name:    "running task blocks patching",
			args:    `{"task_id":1,"operations":[],"message":"x"}`,
			tasks:   makeTasks(database.TaskStatusRunning),
			wantErr: true,
		},
		{
			name:    "task not in flow returns error",
			args:    `{"task_id":99,"operations":[],"message":"x"}`,
			tasks:   makeTasks(database.TaskStatusFinished),
			wantErr: true,
		},
		{
			name:       "empty operations returns unchanged message",
			args:       `{"task_id":1,"operations":[],"message":"x"}`,
			tasks:      makeTasks(database.TaskStatusFinished),
			planned:    makeSubtasks(database.SubtaskStatusCreated),
			wantSubstr: "unchanged",
		},
		{
			name:    "invalid patch operation returns validation error",
			args:    buildPatchArgs(1, invalidOp),
			tasks:   makeTasks(database.TaskStatusFinished),
			planned: makeSubtasks(database.SubtaskStatusCreated),
			wantErr: true,
		},
		{
			name:       "valid patch applied — reports count and new IDs",
			args:       buildPatchArgs(1, validOp),
			tasks:      makeTasks(database.TaskStatusFinished),
			planned:    makeSubtasks(database.SubtaskStatusCreated),
			handler:    func(_ context.Context, _ int64, _ SubtaskPatch) error { return nil },
			newPlanned: []database.Subtask{{ID: 100, Status: database.SubtaskStatusCreated, Title: "new sub", TaskID: 1}},
			wantSubstr: "1 operation(s) applied",
		},
		{
			name:        "no planned subtasks + operations + waiting subtask — descriptive error",
			args:        buildPatchArgs(1, validOp),
			tasks:       makeTasks(database.TaskStatusFinished),
			planned:     nil,
			allSubtasks: []database.Subtask{{ID: 2, Status: database.SubtaskStatusWaiting, Title: "ask-sub", TaskID: 1}},
			wantErr:     true,
		},
		{
			name:        "no planned subtasks + operations + running subtask — descriptive error",
			args:        buildPatchArgs(1, validOp),
			tasks:       makeTasks(database.TaskStatusFinished),
			planned:     nil,
			allSubtasks: []database.Subtask{{ID: 3, Status: database.SubtaskStatusRunning, Title: "run-sub", TaskID: 1}},
			wantErr:     true,
		},
		{
			name:        "no planned subtasks + operations + no active subtasks — generic error",
			args:        buildPatchArgs(1, validOp),
			tasks:       makeTasks(database.TaskStatusFinished),
			planned:     nil,
			allSubtasks: makeSubtasks(database.SubtaskStatusFinished),
			wantErr:     true,
		},
		{
			name:    "handler error propagated",
			args:    buildPatchArgs(1, validOp),
			tasks:   makeTasks(database.TaskStatusFinished),
			planned: makeSubtasks(database.SubtaskStatusCreated),
			handler: func(_ context.Context, _ int64, _ SubtaskPatch) error { return errors.New("patch failed") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plannedCallCount := 0
			db := &mockQuerier{
				getFlowTasksFn: func(_ context.Context, _ int64) ([]database.Task, error) {
					return tt.tasks, nil
				},
				getTaskPlannedFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
					plannedCallCount++
					if plannedCallCount == 1 {
						return tt.planned, tt.plannedErr
					}
					return tt.newPlanned, nil
				},
				getFlowSubtasksFn: func(_ context.Context, _ int64) ([]database.Subtask, error) {
					return tt.allSubtasks, nil
				},
			}
			tool := NewPatchFlowSubtasksTool(1, db, tt.handler)
			result, err := tool.Handle(ctx, "patch_flow_subtasks", json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantSubstr != "" && !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
		})
	}
}

func buildPatchArgs(taskID int64, ops ...SubtaskOperation) string {
	opsJSON, _ := json.Marshal(ops)
	return fmt.Sprintf(`{"task_id":%d,"operations":%s,"message":"x"}`, taskID, opsJSON)
}

// ------------------------------------------------------------------
// SubtaskPatch.Validate (logic coverage)

func TestSubtaskPatchValidate(t *testing.T) {
	t.Parallel()

	id := int64(1)

	tests := []struct {
		name    string
		ops     []SubtaskOperation
		wantErr bool
	}{
		{
			name:    "valid add operation",
			ops:     []SubtaskOperation{{Op: SubtaskOpAdd, Title: "t", Description: "d"}},
			wantErr: false,
		},
		{
			name:    "add missing title",
			ops:     []SubtaskOperation{{Op: SubtaskOpAdd, Description: "d"}},
			wantErr: true,
		},
		{
			name:    "add missing description",
			ops:     []SubtaskOperation{{Op: SubtaskOpAdd, Title: "t"}},
			wantErr: true,
		},
		{
			name:    "remove with id",
			ops:     []SubtaskOperation{{Op: SubtaskOpRemove, ID: &id}},
			wantErr: false,
		},
		{
			name:    "remove without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpRemove}},
			wantErr: true,
		},
		{
			name:    "modify with id and title",
			ops:     []SubtaskOperation{{Op: SubtaskOpModify, ID: &id, Title: "new title"}},
			wantErr: false,
		},
		{
			name:    "modify without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpModify, Title: "t"}},
			wantErr: true,
		},
		{
			name:    "modify without title or description",
			ops:     []SubtaskOperation{{Op: SubtaskOpModify, ID: &id}},
			wantErr: true,
		},
		{
			name:    "reorder with id",
			ops:     []SubtaskOperation{{Op: SubtaskOpReorder, ID: &id}},
			wantErr: false,
		},
		{
			name:    "reorder without id",
			ops:     []SubtaskOperation{{Op: SubtaskOpReorder}},
			wantErr: true,
		},
		{
			name:    "unknown operation type",
			ops:     []SubtaskOperation{{Op: "unknown"}},
			wantErr: true,
		},
		{
			name:    "empty operations list is valid",
			ops:     []SubtaskOperation{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			patch := SubtaskPatch{Operations: tt.ops}
			err := patch.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
