package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"pentagi/pkg/database"
)

// flowManagerDB stands in for the five Querier reads of the flow_manager tools; a read sequence repeats its last entry.
type flowManagerDB struct {
	database.Querier

	taskReads   [][]database.Task
	taskErrs    []error
	subtasks    []database.Subtask
	subtasksErr error
	taskSubs    []database.Subtask
	taskSubsErr error
	planReads   [][]database.Subtask
	planErrs    []error
	logs        []database.Msglog
	logsErr     error

	taskN        int
	planN        int
	filterTaskID int64
	planTaskID   int64
	logSubtaskID sql.NullInt64
}

func flowManagerSeqErr(errs []error, i int) error {
	if i < len(errs) {
		return errs[i]
	}
	return nil
}

func (m *flowManagerDB) GetFlowTasks(context.Context, int64) ([]database.Task, error) {
	i := m.taskN
	m.taskN++
	if err := flowManagerSeqErr(m.taskErrs, i); err != nil {
		return nil, err
	}
	if len(m.taskReads) == 0 {
		return nil, nil
	}
	if i >= len(m.taskReads) {
		i = len(m.taskReads) - 1
	}
	return m.taskReads[i], nil
}

func (m *flowManagerDB) GetFlowSubtasks(context.Context, int64) ([]database.Subtask, error) {
	return m.subtasks, m.subtasksErr
}

func (m *flowManagerDB) GetFlowTaskSubtasks(_ context.Context, arg database.GetFlowTaskSubtasksParams) ([]database.Subtask, error) {
	m.filterTaskID = arg.TaskID
	return m.taskSubs, m.taskSubsErr
}

func (m *flowManagerDB) GetTaskPlannedSubtasks(_ context.Context, taskID int64) ([]database.Subtask, error) {
	m.planTaskID = taskID
	i := m.planN
	m.planN++
	if err := flowManagerSeqErr(m.planErrs, i); err != nil {
		return nil, err
	}
	if len(m.planReads) == 0 {
		return nil, nil
	}
	if i >= len(m.planReads) {
		i = len(m.planReads) - 1
	}
	return m.planReads[i], nil
}

func (m *flowManagerDB) GetSubtaskMsgLogs(_ context.Context, subtaskID sql.NullInt64) ([]database.Msglog, error) {
	m.logSubtaskID = subtaskID
	return m.logs, m.logsErr
}

func flowManagerTasks(statuses ...database.TaskStatus) []database.Task {
	tasks := make([]database.Task, len(statuses))
	for i, s := range statuses {
		tasks[i] = database.Task{ID: int64(i + 1), Status: s, Title: fmt.Sprintf("task-%d", i+1)}
	}
	return tasks
}

func flowManagerSubtasks(statuses ...database.SubtaskStatus) []database.Subtask {
	subs := make([]database.Subtask, len(statuses))
	for i, s := range statuses {
		subs[i] = database.Subtask{ID: int64(i + 1), Status: s, Title: fmt.Sprintf("subtask-%d", i+1), TaskID: 1}
	}
	return subs
}

func flowManagerStatusArgs(detail string, verbose bool, taskID *int64) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, `{"detail":%q,"message":"x"`, detail)
	if verbose {
		sb.WriteString(`,"verbose":true`)
	}
	if taskID != nil {
		fmt.Fprintf(sb, `,"task_id":%d`, *taskID)
	}
	sb.WriteString("}")
	return sb.String()
}

func flowManagerBuildPatchArgs(taskID int64, ops ...SubtaskOperation) string {
	opsJSON, _ := json.Marshal(ops)
	return fmt.Sprintf(`{"task_id":%d,"operations":%s,"message":"x"}`, taskID, opsJSON)
}

func TestFlowManager_StateGuard_IsRecognizableButKeepsTheOriginalMessage(t *testing.T) {
	t.Parallel()

	inner := errors.New("task is running")
	err := stateGuard(inner)

	if !errors.Is(err, ErrFlowStateGuard) {
		t.Error("expected errors.Is(err, ErrFlowStateGuard) to be true")
	}
	if err.Error() != "task is running" {
		t.Errorf("guard must not alter the message, got: %q", err.Error())
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped != inner {
		t.Errorf("Unwrap should yield the original error, got: %v", unwrapped)
	}
	if errors.Is(unwrapped, ErrFlowStateGuard) {
		t.Error("the unwrapped error must no longer be a state guard")
	}
}

func TestFlowManager_InferFlowStatus_ReportsTheMostActiveTaskState(t *testing.T) {
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
			name:     "running task takes priority over waiting",
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
			got := inferFlowStatus(flowManagerTasks(tt.statuses...))
			if got != tt.want {
				t.Errorf("inferFlowStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFlowManager_TruncateText_CutsBeyondTheLimitAndMarksTheCut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "shorter than max", input: "hello", maxLen: 10, want: "hello"},
		{name: "exactly max", input: "hello", maxLen: 5, want: "hello"},
		{name: "longer than max appends ellipsis", input: "hello world", maxLen: 5, want: "hello..."},
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

func TestFlowManager_FlowStatusTool_DispatchesEveryDetailLevel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name          string
		args          string
		wantErrSubstr string
		wantSubstr    string
	}{
		{name: "malformed arguments are rejected", args: `{bad json}`, wantErrSubstr: "failed to parse get_flow_status args"},
		{name: "an unknown detail level is rejected", args: `{"detail":"nonexistent","message":"x"}`, wantErrSubstr: `unknown detail level "nonexistent"`},
		{name: "summary counts the flow", args: flowManagerStatusArgs("summary", false, nil), wantSubstr: "Flow ID: 1"},
		{name: "tasks lists the tasks", args: flowManagerStatusArgs("tasks", false, nil), wantSubstr: "Tasks for flow 1"},
		{name: "subtasks lists the whole flow", args: flowManagerStatusArgs("subtasks", false, nil), wantSubstr: "All subtasks for flow 1"},
		{name: "running shows the active subtask", args: flowManagerStatusArgs("running", false, nil), wantSubstr: "=== Active Subtask ==="},
		{name: "planned lists the unstarted subtasks", args: flowManagerStatusArgs("planned", false, nil), wantSubstr: "All planned subtasks for flow 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// The fake counts its reads, so each parallel subtest needs its own.
			db := &flowManagerDB{
				taskReads: [][]database.Task{flowManagerTasks(database.TaskStatusRunning)},
				subtasks:  flowManagerSubtasks(database.SubtaskStatusRunning, database.SubtaskStatusCreated),
			}
			result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(tt.args))
			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected an error containing %q, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
		})
	}
}

func TestFlowManager_FlowStatusTool_ReportsAFailedReadAsAnError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbDown := errors.New("db down")
	taskID := int64(7)
	running := []database.Subtask{{ID: 4, Status: database.SubtaskStatusRunning, Title: "s", TaskID: 1}}

	tests := []struct {
		name string
		db   *flowManagerDB
		args string
	}{
		{
			name: "summary with a failed task read",
			db:   &flowManagerDB{taskErrs: []error{dbDown}},
			args: flowManagerStatusArgs("summary", false, nil),
		},
		{
			name: "summary with a failed subtask read",
			db:   &flowManagerDB{taskReads: [][]database.Task{nil}, subtasksErr: dbDown},
			args: flowManagerStatusArgs("summary", false, nil),
		},
		{
			name: "task list with a failed task read",
			db:   &flowManagerDB{taskErrs: []error{dbDown}},
			args: flowManagerStatusArgs("tasks", false, nil),
		},
		{
			name: "flow subtask list with a failed read",
			db:   &flowManagerDB{subtasksErr: dbDown},
			args: flowManagerStatusArgs("subtasks", false, nil),
		},
		{
			name: "one task's subtask list with a failed read",
			db:   &flowManagerDB{taskSubsErr: dbDown},
			args: flowManagerStatusArgs("subtasks", false, &taskID),
		},
		{
			name: "running chain with a failed subtask read",
			db:   &flowManagerDB{subtasksErr: dbDown},
			args: flowManagerStatusArgs("running", false, nil),
		},
		{
			name: "running chain with a failed parent task read",
			db:   &flowManagerDB{subtasks: running, taskErrs: []error{dbDown}},
			args: flowManagerStatusArgs("running", false, nil),
		},
		{
			name: "running chain with a failed agent message read",
			db:   &flowManagerDB{subtasks: running, logsErr: dbDown},
			args: flowManagerStatusArgs("running", false, nil),
		},
		{
			name: "flow planned list with a failed subtask read",
			db:   &flowManagerDB{subtasksErr: dbDown},
			args: flowManagerStatusArgs("planned", false, nil),
		},
		{
			name: "one task's planned list with a failed plan read",
			db:   &flowManagerDB{planErrs: []error{dbDown}},
			args: flowManagerStatusArgs("planned", false, &taskID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := NewFlowStatusTool(1, tt.db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(tt.args))
			if !errors.Is(err, dbDown) {
				t.Fatalf("expected the read failure to be returned, got err=%v result=%q", err, result)
			}
		})
	}
}

func TestFlowManager_FlowStatusTool_SummaryCountsWorkAndNamesWhatIsActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty flow emits the no-tasks notice", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{taskReads: [][]database.Task{nil}}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("summary", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No tasks yet") {
			t.Errorf("expected 'No tasks yet', got: %s", result)
		}
	})

	t.Run("counts tasks and subtasks and names the active pair", func(t *testing.T) {
		t.Parallel()
		tasks := flowManagerTasks(
			database.TaskStatusFinished,
			database.TaskStatusFailed,
			database.TaskStatusCreated,
			database.TaskStatusWaiting,
			database.TaskStatusRunning,
		)
		subs := flowManagerSubtasks(
			database.SubtaskStatusRunning,
			database.SubtaskStatusFinished,
			database.SubtaskStatusFinished,
		)
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}, subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("summary", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "total: 5") {
			t.Errorf("expected 5 tasks counted, got: %s", result)
		}
		if !strings.Contains(result, "total: 3") {
			t.Errorf("expected 3 subtasks counted, got: %s", result)
		}
		if !strings.Contains(result, "Active task:") || !strings.Contains(result, "task-5") {
			t.Errorf("expected running task-5 named active, got: %s", result)
		}
		if !strings.Contains(result, "Active subtask:") || !strings.Contains(result, "subtask-1") {
			t.Errorf("expected running subtask-1 named active, got: %s", result)
		}
	})

	// The created task and subtask after the waiting ones must not be named active instead.
	t.Run("a waiting task and subtask are named active", func(t *testing.T) {
		t.Parallel()
		tasks := flowManagerTasks(database.TaskStatusFinished, database.TaskStatusWaiting, database.TaskStatusCreated)
		subs := flowManagerSubtasks(database.SubtaskStatusFinished, database.SubtaskStatusWaiting, database.SubtaskStatusCreated)
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}, subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("summary", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, want := range []string{
			"\nActive task:    ID=2      | waiting  | task-2\n",
			"\nActive subtask: ID=2      | waiting  | subtask-2\n",
		} {
			if !strings.Contains(result, want) {
				t.Errorf("expected %q in the summary, got: %s", want, result)
			}
		}
	})

	t.Run("verbose adds the active task input and subtask description", func(t *testing.T) {
		t.Parallel()
		tasks := []database.Task{{ID: 1, Status: database.TaskStatusRunning, Title: "t", Input: "task-input-text"}}
		subs := []database.Subtask{{ID: 1, Status: database.SubtaskStatusRunning, Title: "s", TaskID: 1, Description: "subtask-desc-text"}}
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}, subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("summary", true, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "task-input-text") {
			t.Errorf("verbose summary should include the active task input, got: %s", result)
		}
		if !strings.Contains(result, "subtask-desc-text") {
			t.Errorf("verbose summary should include the active subtask description, got: %s", result)
		}
	})
}

func TestFlowManager_FlowStatusTool_SummarizesAnOversizedAnswer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	summarizerDown := errors.New("summarizer down")

	bigTask := func(limit int) []database.Task {
		return []database.Task{{ID: 1, Status: database.TaskStatusRunning, Title: strings.Repeat("x", limit+1)}}
	}
	bigSubs := func(limit int, status database.SubtaskStatus) []database.Subtask {
		return []database.Subtask{{ID: 1, Status: status, Title: strings.Repeat("y", limit+1), TaskID: 1, Description: "d"}}
	}

	tests := []struct {
		name string
		db   func() *flowManagerDB
		args string
	}{
		{
			name: "flow summary",
			db:   func() *flowManagerDB { return &flowManagerDB{taskReads: [][]database.Task{bigTask(summaryLimit)}} },
			args: flowManagerStatusArgs("summary", false, nil),
		},
		{
			name: "task list",
			db:   func() *flowManagerDB { return &flowManagerDB{taskReads: [][]database.Task{bigTask(taskListLimit)}} },
			args: flowManagerStatusArgs("tasks", false, nil),
		},
		{
			name: "subtask list",
			db: func() *flowManagerDB {
				return &flowManagerDB{subtasks: bigSubs(subtasksListLimit, database.SubtaskStatusCreated)}
			},
			args: flowManagerStatusArgs("subtasks", false, nil),
		},
		{
			name: "running chain",
			db: func() *flowManagerDB {
				return &flowManagerDB{subtasks: bigSubs(runningInfoLimit, database.SubtaskStatusRunning)}
			},
			args: flowManagerStatusArgs("running", false, nil),
		},
		{
			name: "planned list",
			db: func() *flowManagerDB {
				return &flowManagerDB{subtasks: bigSubs(plannedListLimit, database.SubtaskStatusCreated)}
			},
			args: flowManagerStatusArgs("planned", false, nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" is replaced by its summary", func(t *testing.T) {
			t.Parallel()
			called := false
			summarizer := func(context.Context, string) (string, error) {
				called = true
				return "summarized", nil
			}
			result, err := NewFlowStatusTool(1, tt.db(), summarizer).Handle(ctx, GetFlowStatusToolName, json.RawMessage(tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Error("expected the summarizer to be called for oversized output")
			}
			if result != "summarized" {
				t.Errorf("expected summarizer output, got: %s", result)
			}
		})

		t.Run(tt.name+" with a failed summary is an error", func(t *testing.T) {
			t.Parallel()
			summarizer := func(context.Context, string) (string, error) { return "", summarizerDown }
			result, err := NewFlowStatusTool(1, tt.db(), summarizer).Handle(ctx, GetFlowStatusToolName, json.RawMessage(tt.args))
			if !errors.Is(err, summarizerDown) {
				t.Fatalf("expected the summarizer failure to be returned, got err=%v", err)
			}
			if result != "" {
				t.Errorf("a failed summary must not return the oversized answer, got %d bytes", len(result))
			}
		})
	}
}

func TestFlowManager_FlowStatusTool_TasksShowInputAndResultOnlyWhenVerbose(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("no tasks returns a specific message", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{taskReads: [][]database.Task{nil}}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("tasks", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No tasks found") {
			t.Errorf("expected 'No tasks found', got: %s", result)
		}
	})

	tasks := []database.Task{{ID: 1, Status: database.TaskStatusFinished, Title: "t1", Input: "the-input", Result: "the-result"}}

	t.Run("verbose includes input and result", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("tasks", true, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "the-input") || !strings.Contains(result, "the-result") {
			t.Errorf("verbose tasks should include input and result, got: %s", result)
		}
	})

	t.Run("non-verbose hides input and result", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("tasks", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result, "the-input") || strings.Contains(result, "the-result") {
			t.Errorf("non-verbose tasks must not include input or result, got: %s", result)
		}
		if !strings.Contains(result, "t1") {
			t.Errorf("expected task title, got: %s", result)
		}
	})
}

func TestFlowManager_FlowStatusTool_SubtasksNarrowToTheTaskAsked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("without task_id lists the whole flow", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{subtasks: flowManagerSubtasks(database.SubtaskStatusFinished)}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("subtasks", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "All subtasks for flow") {
			t.Errorf("expected flow-level header, got: %s", result)
		}
	})

	t.Run("with task_id narrows to that task", func(t *testing.T) {
		t.Parallel()
		taskID := int64(42)
		db := &flowManagerDB{taskSubs: flowManagerSubtasks(database.SubtaskStatusRunning)}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("subtasks", false, &taskID)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db.filterTaskID != taskID {
			t.Errorf("expected GetFlowTaskSubtasks to be asked for task %d, got %d", taskID, db.filterTaskID)
		}
		if !strings.Contains(result, "Subtasks for task 42") {
			t.Errorf("expected task-level header, got: %s", result)
		}
	})

	t.Run("empty result returns a no-subtasks message", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{subtasks: nil}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("subtasks", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No subtasks found") {
			t.Errorf("expected 'No subtasks found', got: %s", result)
		}
	})

	t.Run("verbose adds description and result", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{{ID: 1, Status: database.SubtaskStatusRunning, Title: "s", TaskID: 1, Description: "sub-desc", Result: "sub-result"}}
		db := &flowManagerDB{subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("subtasks", true, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "sub-desc") || !strings.Contains(result, "sub-result") {
			t.Errorf("verbose subtasks should include description and result, got: %s", result)
		}
	})
}

func TestFlowManager_FlowStatusTool_RunningShowsTheActiveChain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("idle flow reports no active subtask", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{subtasks: flowManagerSubtasks(database.SubtaskStatusFinished, database.SubtaskStatusFailed)}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No running or waiting subtask found") {
			t.Errorf("expected idle message, got: %s", result)
		}
	})

	t.Run("running subtask shows its parent task, input and result", func(t *testing.T) {
		t.Parallel()
		tasks := []database.Task{{ID: 1, Status: database.TaskStatusRunning, Title: "main-task", Input: "task-input", Result: "task-result"}}
		subs := []database.Subtask{{ID: 10, Status: database.SubtaskStatusRunning, Title: "active-sub", TaskID: 1, Result: "sub-result"}}
		db := &flowManagerDB{taskReads: [][]database.Task{tasks}, subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, want := range []string{"Active Task", "main-task", "task-input", "task-result", "active-sub", "sub-result"} {
			if !strings.Contains(result, want) {
				t.Errorf("expected %q in running chain, got: %s", want, result)
			}
		}
		if db.logSubtaskID.Int64 != 10 {
			t.Errorf("expected agent messages to be fetched for subtask 10, got %d", db.logSubtaskID.Int64)
		}
	})

	t.Run("waiting subtask shows the ask note and the submit tool", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{{ID: 5, Status: database.SubtaskStatusWaiting, Title: "ask-sub", TaskID: 2}}
		db := &flowManagerDB{subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "waiting for user input") {
			t.Errorf("expected waiting note, got: %s", result)
		}
		if !strings.Contains(result, SubmitFlowInputToolName) {
			t.Errorf("expected %s named, got: %s", SubmitFlowInputToolName, result)
		}
	})

	t.Run("execution context shows only when verbose", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{{ID: 8, Status: database.SubtaskStatusRunning, Title: "s", TaskID: 1, Context: "exec-context-data"}}

		quiet := &flowManagerDB{subtasks: subs}
		result, err := NewFlowStatusTool(1, quiet, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result, "exec-context-data") {
			t.Errorf("non-verbose must hide execution context, got: %s", result)
		}

		loud := &flowManagerDB{subtasks: subs}
		result, err = NewFlowStatusTool(1, loud, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", true, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "exec-context-data") {
			t.Errorf("verbose must show execution context, got: %s", result)
		}
	})

	t.Run("agent messages keep the newest ten, or fifty when verbose", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{{ID: 4, Status: database.SubtaskStatusRunning, Title: "s", TaskID: 1}}
		logs := make([]database.Msglog, 51)
		for i := range logs {
			logs[i] = database.Msglog{ID: int64(i + 1), Message: fmt.Sprintf("log-%02d", i+1)}
		}

		tests := []struct {
			name    string
			verbose bool
			header  string
			dropped string
			kept    []string
		}{
			{name: "quiet", verbose: false, header: "=== Last 10 agent messages ===", dropped: "log-41", kept: []string{"log-42", "log-51"}},
			{name: "verbose", verbose: true, header: "=== Last 50 agent messages ===", dropped: "log-01", kept: []string{"log-02", "log-51"}},
		}
		for _, tt := range tests {
			db := &flowManagerDB{subtasks: subs, logs: logs}
			result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("running", tt.verbose, nil)))
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.name, err)
			}
			if !strings.Contains(result, tt.header) {
				t.Errorf("%s: expected %q, got: %s", tt.name, tt.header, result)
			}
			if strings.Contains(result, tt.dropped) {
				t.Errorf("%s: %q is older than the limit and must be dropped, got: %s", tt.name, tt.dropped, result)
			}
			for _, want := range tt.kept {
				if !strings.Contains(result, want) {
					t.Errorf("%s: expected %q within the limit, got: %s", tt.name, want, result)
				}
			}
		}
	})
}

func TestFlowManager_FlowStatusTool_PlannedListsOnlyUnstartedSubtasks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("with task_id reads the task's planned subtasks", func(t *testing.T) {
		t.Parallel()
		taskID := int64(7)
		db := &flowManagerDB{planReads: [][]database.Subtask{flowManagerSubtasks(database.SubtaskStatusCreated)}}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("planned", false, &taskID)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db.planTaskID != taskID {
			t.Errorf("expected GetTaskPlannedSubtasks for task %d, got %d", taskID, db.planTaskID)
		}
		if !strings.Contains(result, "Planned subtasks for task 7") {
			t.Errorf("expected task-level header, got: %s", result)
		}
	})

	t.Run("without task_id filters created subtasks from the flow", func(t *testing.T) {
		t.Parallel()
		subs := []database.Subtask{
			{ID: 1, Status: database.SubtaskStatusCreated, Title: "planned-sub", TaskID: 1, Description: "d"},
			{ID: 2, Status: database.SubtaskStatusFinished, Title: "done-sub", TaskID: 1},
		}
		db := &flowManagerDB{subtasks: subs}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("planned", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "planned-sub") {
			t.Errorf("expected the created subtask, got: %s", result)
		}
		if strings.Contains(result, "done-sub") {
			t.Errorf("finished subtask must not appear in the planned list, got: %s", result)
		}
	})

	t.Run("verbose expands the description beyond the short preview", func(t *testing.T) {
		t.Parallel()
		longDesc := strings.Repeat("z", 400)
		subs := []database.Subtask{{ID: 1, Status: database.SubtaskStatusCreated, Title: "s", TaskID: 1, Description: longDesc}}

		short := &flowManagerDB{subtasks: subs}
		result, err := NewFlowStatusTool(1, short, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("planned", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "...") || strings.Contains(result, longDesc) {
			t.Errorf("non-verbose planned should truncate the description to a preview, got len %d", len(result))
		}

		full := &flowManagerDB{subtasks: subs}
		result, err = NewFlowStatusTool(1, full, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("planned", true, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, longDesc) {
			t.Errorf("verbose planned should include the full description, got: %s", result)
		}
	})

	t.Run("all subtasks executed returns a specific message", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{subtasks: flowManagerSubtasks(database.SubtaskStatusFinished, database.SubtaskStatusFailed)}
		result, err := NewFlowStatusTool(1, db, nil).Handle(ctx, GetFlowStatusToolName, json.RawMessage(flowManagerStatusArgs("planned", false, nil)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No planned") {
			t.Errorf("expected 'No planned' message, got: %s", result)
		}
	})
}

func TestFlowManager_AppendSubtaskMsgLogs_KeepsTheNewestWithinTheLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("no logs returns a specific message", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{logs: nil}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 5, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "No agent messages found for subtask 5") {
			t.Errorf("expected the empty notice to name subtask 5, got: %s", result)
		}
	})

	t.Run("fewer logs than the limit shows all, including results", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{logs: []database.Msglog{
			{ID: 1, Message: "msg-alpha", Result: "res-alpha"},
			{ID: 2, Message: "msg-beta"},
		}}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 1, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "msg-alpha") || !strings.Contains(result, "msg-beta") {
			t.Errorf("expected all messages, got: %s", result)
		}
		if !strings.Contains(result, "res-alpha") {
			t.Errorf("expected the message result to be rendered, got: %s", result)
		}
	})

	t.Run("more logs than the limit keeps only the tail", func(t *testing.T) {
		t.Parallel()
		logs := make([]database.Msglog, 5)
		for i := range logs {
			logs[i] = database.Msglog{ID: int64(i + 1), Message: fmt.Sprintf("msg-%d", i+1)}
		}
		db := &flowManagerDB{logs: logs}
		result, err := NewFlowStatusTool(1, db, nil).appendSubtaskMsgLogs(ctx, 1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, absent := range []string{"msg-1", "msg-2", "msg-3"} {
			if strings.Contains(result, absent) {
				t.Errorf("message %q should have been dropped by the limit, got: %s", absent, result)
			}
		}
		for _, present := range []string{"msg-4", "msg-5"} {
			if !strings.Contains(result, present) {
				t.Errorf("recent message %q should appear, got: %s", present, result)
			}
		}
	})

	numLogs := msgLogsLimit/1024 + 2
	oversized := make([]database.Msglog, numLogs)
	for i := range oversized {
		oversized[i] = database.Msglog{ID: int64(i + 1), Message: strings.Repeat("x", 1024)}
	}

	t.Run("oversized output is handed to the summarizer", func(t *testing.T) {
		t.Parallel()
		db := &flowManagerDB{logs: oversized}
		called := false
		summarizer := func(context.Context, string) (string, error) {
			called = true
			return "summarized-logs", nil
		}
		result, err := NewFlowStatusTool(1, db, summarizer).appendSubtaskMsgLogs(ctx, 1, numLogs+10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("expected the summarizer to be called for oversized logs output")
		}
		if result != "summarized-logs" {
			t.Errorf("expected summarizer output, got: %s", result)
		}
	})

	t.Run("a failed summary of oversized output is an error", func(t *testing.T) {
		t.Parallel()
		summarizerDown := errors.New("summarizer down")
		db := &flowManagerDB{logs: oversized}
		summarizer := func(context.Context, string) (string, error) { return "", summarizerDown }
		result, err := NewFlowStatusTool(1, db, summarizer).appendSubtaskMsgLogs(ctx, 1, numLogs+10)
		if !errors.Is(err, summarizerDown) {
			t.Fatalf("expected the summarizer failure to be returned, got err=%v", err)
		}
		if result != "" {
			t.Errorf("a failed summary must not return the oversized list, got %d bytes", len(result))
		}
	})
}

// Subtests are keyed by getInputText, getDescriptionText and getResultText; literal limits make a policy change fail here.
func TestFlowManager_FlowStatusTool_TextHelpersBoundWhatTheyQuote(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	summarizerDown := errors.New("summarizer down")

	tests := []struct {
		name  string
		call  func(*flowStatusTool, string) (string, error)
		limit int
	}{
		{
			name:  "task input",
			call:  func(tl *flowStatusTool, s string) (string, error) { return tl.getInputText(ctx, s) },
			limit: 8192,
		},
		{
			name:  "subtask description",
			call:  func(tl *flowStatusTool, s string) (string, error) { return tl.getDescriptionText(ctx, s) },
			limit: 4096,
		},
		{
			name:  "result",
			call:  func(tl *flowStatusTool, s string) (string, error) { return tl.getResultText(ctx, s) },
			limit: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+": empty stays empty", func(t *testing.T) {
			t.Parallel()
			got, err := tt.call(NewFlowStatusTool(1, &flowManagerDB{}, nil), "")
			if err != nil || got != "" {
				t.Errorf("empty input: got (%q, %v), want (\"\", nil)", got, err)
			}
		})

		t.Run(tt.name+": up to twice the limit is cut and marked, not summarized", func(t *testing.T) {
			t.Parallel()
			called := false
			tool := NewFlowStatusTool(1, &flowManagerDB{}, func(context.Context, string) (string, error) {
				called = true
				return "summarized", nil
			})
			got, err := tt.call(tool, strings.Repeat("a", 2*tt.limit))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if called {
				t.Error("text of exactly twice the limit must be cut, not summarized")
			}
			if want := strings.Repeat("a", tt.limit) + "..."; got != want {
				t.Errorf("expected %d chars plus a truncation marker, got len=%d ending %q", tt.limit, len(got), got[max(0, len(got)-5):])
			}
		})

		t.Run(tt.name+": beyond twice the limit is summarized", func(t *testing.T) {
			t.Parallel()
			called := false
			tool := NewFlowStatusTool(1, &flowManagerDB{}, func(context.Context, string) (string, error) {
				called = true
				return "summarized", nil
			})
			got, err := tt.call(tool, strings.Repeat("b", 2*tt.limit+1))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Error("expected the summarizer to be called for oversized text")
			}
			if got != "summarized" {
				t.Errorf("expected summarizer output, got: %s", got)
			}
		})

		t.Run(tt.name+": a failed summary is an error", func(t *testing.T) {
			t.Parallel()
			tool := NewFlowStatusTool(1, &flowManagerDB{}, func(context.Context, string) (string, error) {
				return "", summarizerDown
			})
			got, err := tt.call(tool, strings.Repeat("c", 2*tt.limit+1))
			if !errors.Is(err, summarizerDown) {
				t.Fatalf("expected the summarizer failure to be returned, got err=%v", err)
			}
			if got != "" {
				t.Errorf("a failed summary must not return the oversized text, got %d bytes", len(got))
			}
		})
	}
}

// synctest's clock moves only while the bubble waits, so a wait is measured exactly and one with no deadline deadlocks.
func TestFlowManager_WaitFlowCompletionTool_ReportsHowTheWaitEnded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	neverFinishes := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	tests := []struct {
		name          string
		args          string
		tasks         []database.Task
		tasksErr      error
		handler       func(context.Context) error
		wantSubstr    string
		wantErrSubstr string
		wantWait      time.Duration
	}{
		{
			name:          "malformed arguments are rejected",
			args:          `{bad}`,
			wantErrSubstr: "failed to parse wait_flow_completion args",
		},
		{
			name:          "a failed task read is an error",
			args:          `{"timeout":10,"message":"x"}`,
			tasksErr:      errors.New("db down"),
			wantErrSubstr: "failed to check flow status: db down",
		},
		{
			name:       "no tasks yet points at submit_flow_input",
			args:       `{"timeout":10,"message":"x"}`,
			tasks:      nil,
			wantSubstr: "Use submit_flow_input to submit the first task description",
		},
		{
			name:       "nothing running points at get_flow_status",
			args:       `{"timeout":10,"message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusWaiting),
			wantSubstr: "not currently running",
		},
		{
			name:       "running task completes within the wait",
			args:       `{"timeout":10,"message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    func(context.Context) error { return nil },
			wantSubstr: "has completed",
		},
		{
			name:       "default timeout applied when zero, still running reports the elapsed wait",
			args:       `{"timeout":0,"message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    neverFinishes,
			wantSubstr: "after waiting 1m0s",
			wantWait:   time.Minute,
		},
		{
			name:       "timeout above the cap is clamped to one hour",
			args:       `{"timeout":99999,"message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    neverFinishes,
			wantSubstr: "after waiting 1h0m0s",
			wantWait:   time.Hour,
		},
		{
			name:       "a timeout within the cap is the wait",
			args:       `{"timeout":10,"message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    neverFinishes,
			wantSubstr: "after waiting 10s",
			wantWait:   10 * time.Second,
		},
		{
			name:          "wait cancelled surfaces an interruption error",
			args:          `{"timeout":10,"message":"x"}`,
			tasks:         flowManagerTasks(database.TaskStatusRunning),
			handler:       func(context.Context) error { return context.Canceled },
			wantErrSubstr: "interrupted",
		},
		{
			name:          "other handler error is propagated",
			args:          `{"timeout":10,"message":"x"}`,
			tasks:         flowManagerTasks(database.TaskStatusRunning),
			handler:       func(context.Context) error { return errors.New("boom") },
			wantErrSubstr: "wait for flow completion failed: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				db := &flowManagerDB{taskReads: [][]database.Task{tt.tasks}, taskErrs: []error{tt.tasksErr}}
				tool := NewWaitFlowCompletionTool(1, db, tt.handler)
				start := time.Now()
				result, err := tool.Handle(ctx, WaitFlowCompletionToolName, json.RawMessage(tt.args))
				if waited := time.Since(start); waited != tt.wantWait {
					t.Errorf("the wait ended after %v, want %v", waited, tt.wantWait)
				}
				if tt.wantErrSubstr != "" {
					if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
						t.Fatalf("expected an error containing %q, got: %v", tt.wantErrSubstr, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(result, tt.wantSubstr) {
					t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
				}
			})
		})
	}
}

func TestFlowManager_StopFlowTool_ReportsTheStateTheStopReached(t *testing.T) {
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
		wantSubstr    string
		wantErrSubstr string
		wantReason    string
	}{
		{
			name:          "malformed arguments are rejected",
			args:          `{bad}`,
			wantErrSubstr: "failed to parse stop_flow args",
		},
		{
			name:          "a failed task read is an error",
			args:          `{"reason":"test","message":"x"}`,
			tasksErr:      errors.New("db down"),
			wantErrSubstr: "failed to check flow status: db down",
		},
		{
			name:       "no running task means already waiting",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusFinished, database.TaskStatusWaiting),
			wantSubstr: "already in 'waiting' state",
		},
		{
			name:       "stop succeeds and the flow reaches waiting",
			args:       `{"reason":"cleanup","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    func(context.Context, string) error { return nil },
			newTasks:   flowManagerTasks(database.TaskStatusFinished),
			wantSubstr: "stopped successfully (reason: cleanup)",
			wantReason: "cleanup",
		},
		{
			name:          "handler error is propagated",
			args:          `{"reason":"test","message":"x"}`,
			tasks:         flowManagerTasks(database.TaskStatusRunning),
			handler:       func(context.Context, string) error { return errors.New("stop failed") },
			wantErrSubstr: "failed to stop flow: stop failed",
		},
		{
			name:          "deadline exceeded reports a timeout",
			args:          `{"reason":"test","message":"x"}`,
			tasks:         flowManagerTasks(database.TaskStatusRunning),
			handler:       func(context.Context, string) error { return context.DeadlineExceeded },
			wantErrSubstr: "timed out",
		},
		{
			name:          "context cancelled reports a timeout",
			args:          `{"reason":"test","message":"x"}`,
			tasks:         flowManagerTasks(database.TaskStatusRunning),
			handler:       func(context.Context, string) error { return context.Canceled },
			wantErrSubstr: "timed out",
		},
		{
			name:       "stop leaves a task still running warns",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    func(context.Context, string) error { return nil },
			newTasks:   flowManagerTasks(database.TaskStatusRunning),
			wantSubstr: "has not reached 'waiting'",
		},
		{
			name:       "stop succeeds but the re-query fails returns a partial message",
			args:       `{"reason":"test","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			handler:    func(context.Context, string) error { return nil },
			newTaskErr: errors.New("db fail"),
			wantSubstr: "Could not verify new status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var reasonSeen string
			handler := tt.handler
			if handler != nil {
				inner := tt.handler
				handler = func(ctx context.Context, reason string) error {
					reasonSeen = reason
					return inner(ctx, reason)
				}
			}
			db := &flowManagerDB{
				taskReads: [][]database.Task{tt.tasks, tt.newTasks},
				taskErrs:  []error{tt.tasksErr, tt.newTaskErr},
			}
			tool := NewStopFlowTool(1, db, handler)
			result, err := tool.Handle(ctx, StopFlowToolName, json.RawMessage(tt.args))
			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected an error containing %q, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
			if tt.wantReason != "" && reasonSeen != tt.wantReason {
				t.Errorf("expected the stop reason %q to reach the handler, got %q", tt.wantReason, reasonSeen)
			}
		})
	}
}

func TestFlowManager_SubmitFlowInputTool_TellsWhatBecameOfTheInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name          string
		args          string
		tasks         []database.Task
		tasksErr      error
		subtasks      []database.Subtask
		handler       func(context.Context, string) error
		newTasks      []database.Task
		wantSubstr    string
		wantErrSubstr string
		wantGuard     bool
	}{
		{
			name:          "malformed arguments are rejected",
			args:          `{bad}`,
			wantErrSubstr: "failed to parse submit_flow_input args",
		},
		{
			name:          "empty input is rejected",
			args:          `{"input":"","message":"x"}`,
			wantErrSubstr: "input must not be empty",
		},
		{
			name:          "a failed task read is an error",
			args:          `{"input":"hi","message":"x"}`,
			tasksErr:      errors.New("db down"),
			handler:       func(context.Context, string) error { return nil }, // a tool that ignores the read fails on its answer, not a nil call
			wantErrSubstr: "failed to check flow status: db down",
		},
		{
			name:       "running task blocks submission with soft guidance",
			args:       `{"input":"hello","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusRunning),
			wantSubstr: `task "task-1" (ID: 1) is currently running`,
		},
		{
			name:       "waiting subtask receives the input as an ask answer",
			args:       `{"input":"my answer","message":"x"}`,
			tasks:      flowManagerTasks(database.TaskStatusWaiting),
			subtasks:   flowManagerSubtasks(database.SubtaskStatusWaiting),
			handler:    func(context.Context, string) error { return nil },
			wantSubstr: "answer to the waiting subtask",
		},
		{
			name:          "idle flow, handler error is propagated",
			args:          `{"input":"do something","message":"x"}`,
			handler:       func(context.Context, string) error { return errors.New("submit failed") },
			wantErrSubstr: "failed to submit flow input: submit failed",
		},
		{
			name: "idle flow, a 'not in waiting state' refusal becomes soft guidance",
			args: `{"input":"do something","message":"x"}`,
			handler: func(context.Context, string) error {
				return errors.New("flow is not in 'waiting' state (current: running)")
			},
			wantSubstr: "currently active (not in 'waiting' state)",
		},
		{
			name: "idle flow, a 'cannot submit input' refusal becomes soft guidance",
			args: `{"input":"do something","message":"x"}`,
			handler: func(context.Context, string) error {
				return errors.New("cannot submit input: flow is busy")
			},
			wantSubstr: "currently active (not in 'waiting' state)",
		},
		{
			name: "idle flow, a stale message chain is a guarded, actionable error",
			args: `{"input":"do something","message":"x"}`,
			handler: func(context.Context, string) error {
				return fmt.Errorf("failed to put flow input: failed to get message chain 63247: sql: no rows in result set")
			},
			wantErrSubstr: "no longer available in the database",
			wantGuard:     true,
		},
		{
			name:          "idle flow, deadline exceeded reports a timeout",
			args:          `{"input":"do something","message":"x"}`,
			handler:       func(context.Context, string) error { return context.DeadlineExceeded },
			wantErrSubstr: "timed out",
		},
		{
			name:          "idle flow, context cancelled reports a timeout",
			args:          `{"input":"do something","message":"x"}`,
			handler:       func(context.Context, string) error { return context.Canceled },
			wantErrSubstr: "timed out",
		},
		{
			name:       "idle flow, a new task is now running",
			args:       `{"input":"do something","message":"x"}`,
			handler:    func(context.Context, string) error { return nil },
			newTasks:   []database.Task{{ID: 5, Status: database.TaskStatusRunning, Title: "new-task"}},
			wantSubstr: `Task "new-task" (ID: 5) is now running`,
		},
		{
			name:       "idle flow, no running task appears before the poll times out",
			args:       `{"input":"do something","message":"x"}`,
			handler:    func(context.Context, string) error { return nil },
			newTasks:   flowManagerTasks(database.TaskStatusCreated),
			wantSubstr: "generator may still be working",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := &flowManagerDB{
				taskReads: [][]database.Task{tt.tasks, tt.newTasks},
				taskErrs:  []error{tt.tasksErr},
				subtasks:  tt.subtasks,
			}
			tool := NewSubmitFlowInputTool(1, db, tt.handler)
			tool.pollInterval = 1 * time.Millisecond
			tool.pollTimeout = 20 * time.Millisecond

			result, err := tool.Handle(ctx, SubmitFlowInputToolName, json.RawMessage(tt.args))
			if got := errors.Is(err, ErrFlowStateGuard); got != tt.wantGuard {
				t.Errorf("state guard = %v, want %v (err: %v)", got, tt.wantGuard, err)
			}
			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected an error containing %q, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result, tt.wantSubstr) {
				t.Errorf("expected %q in result, got: %s", tt.wantSubstr, result)
			}
		})
	}
}

func TestFlowManager_SubmitFlowInputTool_StopsWaitingWhenTheContextIsCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := &flowManagerDB{
		// The guard read sees an idle flow and every poll a task that never runs, so only ctx ends the loop.
		taskReads: [][]database.Task{nil, flowManagerTasks(database.TaskStatusCreated)},
	}
	tool := NewSubmitFlowInputTool(1, db, func(context.Context, string) error {
		cancel()
		return nil
	})
	tool.pollInterval = 1 * time.Millisecond
	// Bounds the test if the loop ignores ctx: it would then time out with no error and fail below.
	tool.pollTimeout = 5 * time.Second

	_, err := tool.Handle(ctx, SubmitFlowInputToolName, json.RawMessage(`{"input":"go","message":"x"}`))
	if err == nil || !strings.Contains(err.Error(), "context cancelled while waiting for task to start") {
		t.Fatalf("expected the wait to stop on the cancelled context, got: %v", err)
	}
}

// Only refusals caused by the flow's state carry ErrFlowStateGuard.
func TestFlowManager_PatchFlowSubtasksTool_AppliesOnlyWhatTheFlowStateAllows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	validOp := SubtaskOperation{Op: SubtaskOpAdd, Title: "new sub", Description: "do thing"}
	invalidOp := SubtaskOperation{Op: SubtaskOpRemove}

	tests := []struct {
		name       string
		args       string
		tasks      []database.Task
		tasksErr   error
		planned    []database.Subtask
		plannedErr error
		allSubs    []database.Subtask
		allSubsErr error
		handler    func(context.Context, int64, SubtaskPatch) error
		newPlanned []database.Subtask
		newPlanErr error
		wantErr    bool
		wantGuard  bool
		want       []string // in the error when wantErr, else in the result
		wantAbsent []string
	}{
		{
			name:    "malformed arguments are rejected",
			args:    `{bad}`,
			wantErr: true,
			want:    []string{"failed to parse patch_flow_subtasks args"},
		},
		{
			name:    "task_id zero is rejected",
			args:    `{"task_id":0,"operations":[],"message":"x"}`,
			wantErr: true,
			want:    []string{"task_id must be a positive integer"},
		},
		{
			name:     "a failed task read is an error",
			args:     `{"task_id":1,"operations":[],"message":"x"}`,
			tasksErr: errors.New("db down"),
			wantErr:  true,
			want:     []string{"failed to check flow status: db down"},
		},
		{
			name:      "running task blocks patching",
			args:      `{"task_id":1,"operations":[],"message":"x"}`,
			tasks:     flowManagerTasks(database.TaskStatusRunning),
			wantErr:   true,
			wantGuard: true,
			want:      []string{`task "task-1" (ID: 1) is currently running; patching is not allowed`},
		},
		{
			name:      "task not in flow is a guarded error",
			args:      `{"task_id":99,"operations":[],"message":"x"}`,
			tasks:     flowManagerTasks(database.TaskStatusFinished),
			wantErr:   true,
			wantGuard: true,
			want:      []string{"task ID 99 was not found in this flow"},
		},
		{
			name:    "empty operations leaves the plan unchanged",
			args:    `{"task_id":1,"operations":[],"message":"x"}`,
			tasks:   flowManagerTasks(database.TaskStatusFinished),
			planned: flowManagerSubtasks(database.SubtaskStatusCreated),
			want:    []string{"the subtask plan for task 1 is unchanged"},
		},
		{
			name:       "a failed plan read is an error",
			args:       flowManagerBuildPatchArgs(1, validOp),
			tasks:      flowManagerTasks(database.TaskStatusFinished),
			plannedErr: errors.New("db fail"),
			wantErr:    true,
			want:       []string{"failed to get planned subtasks for task 1: db fail"},
		},
		{
			name:    "invalid patch operation is a validation error",
			args:    flowManagerBuildPatchArgs(1, invalidOp),
			tasks:   flowManagerTasks(database.TaskStatusFinished),
			planned: flowManagerSubtasks(database.SubtaskStatusCreated),
			wantErr: true,
			want:    []string{"invalid subtask patch"},
		},
		{
			name:       "valid patch reports the count and the new IDs",
			args:       flowManagerBuildPatchArgs(1, validOp),
			tasks:      flowManagerTasks(database.TaskStatusFinished),
			planned:    flowManagerSubtasks(database.SubtaskStatusCreated),
			handler:    func(context.Context, int64, SubtaskPatch) error { return nil },
			newPlanned: []database.Subtask{{ID: 100, Status: database.SubtaskStatusCreated, Title: "new sub", TaskID: 1}},
			want:       []string{"1 operation(s) applied", "ID: 100 | new sub"},
		},
		{
			name:       "patch applied but the re-query fails returns the count and a partial message",
			args:       flowManagerBuildPatchArgs(1, validOp),
			tasks:      flowManagerTasks(database.TaskStatusFinished),
			planned:    flowManagerSubtasks(database.SubtaskStatusCreated),
			handler:    func(context.Context, int64, SubtaskPatch) error { return nil },
			newPlanErr: errors.New("db fail"),
			want:       []string{"1 operation(s) applied", "Could not retrieve updated subtask list"},
		},
		{
			name:      "no plan but a waiting subtask is a guarded error naming it",
			args:      flowManagerBuildPatchArgs(1, validOp),
			tasks:     flowManagerTasks(database.TaskStatusFinished),
			allSubs:   []database.Subtask{{ID: 2, Status: database.SubtaskStatusWaiting, Title: "ask-sub", TaskID: 1}},
			wantErr:   true,
			wantGuard: true,
			want:      []string{`(ID: 2, "ask-sub") waiting for user input`},
		},
		{
			name:  "no plan but a running subtask is a guarded error naming it, not another task's subtask",
			args:  flowManagerBuildPatchArgs(1, validOp),
			tasks: flowManagerTasks(database.TaskStatusFinished),
			allSubs: []database.Subtask{
				{ID: 9, Status: database.SubtaskStatusWaiting, Title: "other-ask", TaskID: 2},
				{ID: 3, Status: database.SubtaskStatusRunning, Title: "run-sub", TaskID: 1},
			},
			wantErr:    true,
			wantGuard:  true,
			want:       []string{`(ID: 3, "run-sub") currently running`},
			wantAbsent: []string{"other-ask"},
		},
		{
			name:      "no plan and no active subtasks is a guarded generic error",
			args:      flowManagerBuildPatchArgs(1, validOp),
			tasks:     flowManagerTasks(database.TaskStatusFinished),
			allSubs:   flowManagerSubtasks(database.SubtaskStatusFinished),
			wantErr:   true,
			wantGuard: true,
			want:      []string{"no 'created' subtasks found for task 1"},
		},
		{
			name:       "no plan and a failed subtask read is an unguarded error",
			args:       flowManagerBuildPatchArgs(1, validOp),
			tasks:      flowManagerTasks(database.TaskStatusFinished),
			allSubsErr: errors.New("db fail"),
			wantErr:    true,
			want:       []string{"failed to check subtask state for task 1: db fail"},
		},
		{
			name:    "handler error is propagated",
			args:    flowManagerBuildPatchArgs(1, validOp),
			tasks:   flowManagerTasks(database.TaskStatusFinished),
			planned: flowManagerSubtasks(database.SubtaskStatusCreated),
			handler: func(context.Context, int64, SubtaskPatch) error { return errors.New("patch failed") },
			wantErr: true,
			want:    []string{"failed to patch subtasks for task 1: patch failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := &flowManagerDB{
				taskReads:   [][]database.Task{tt.tasks},
				taskErrs:    []error{tt.tasksErr},
				planReads:   [][]database.Subtask{tt.planned, tt.newPlanned},
				planErrs:    []error{tt.plannedErr, tt.newPlanErr},
				subtasks:    tt.allSubs,
				subtasksErr: tt.allSubsErr,
			}
			tool := NewPatchFlowSubtasksTool(1, db, tt.handler)
			result, err := tool.Handle(ctx, PatchFlowSubtasksToolName, json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := errors.Is(err, ErrFlowStateGuard); got != tt.wantGuard {
				t.Errorf("state guard = %v, want %v (err: %v)", got, tt.wantGuard, err)
			}
			text := result
			if err != nil {
				text = err.Error()
			}
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Errorf("expected %q, got: %s", want, text)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(text, absent) {
					t.Errorf("did not expect %q, got: %s", absent, text)
				}
			}
		})
	}
}
