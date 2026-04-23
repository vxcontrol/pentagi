package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/database"
)

// flowStatusTool implements get_flow_status — reads flow/task/subtask state from DB.
type flowStatusTool struct {
	flowID     int64
	db         database.Querier
	summarizer SummarizeHandler
}

func NewFlowStatusTool(flowID int64, db database.Querier, summarizer SummarizeHandler) *flowStatusTool {
	return &flowStatusTool{flowID: flowID, db: db, summarizer: summarizer}
}

const (
	msgLogLimitNormal     = 10
	msgLogLimitVerbose    = 50
	flowOperationTimeout  = 15 * time.Second
	taskReadyPollInterval = 5 * time.Second
	taskReadyPollTimeout  = 2 * time.Minute
	msgLogsLimit          = 16 * 1024  // 16 KB
	summaryLimit          = 32 * 1024  // 32 KB
	taskListLimit         = 32 * 1024  // 32 KB
	subtasksListLimit     = 48 * 1024  // 48 KB
	plannedListLimit      = 32 * 1024  // 32 KB
	runningInfoLimit      = 48 * 1024  // 48 KB
	inputLimit            = 8 * 1024   // 8 KB
	descriptionLimit      = 4 * 1024   // 4 KB
	resultLimit           = 8 * 1024   // 8 KB
	summarizationLimit    = 128 * 1024 // 128 KB hard limit
)

func (t *flowStatusTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action GetFlowStatusAction
	if err := json.Unmarshal(args, &action); err != nil {
		return "", fmt.Errorf("failed to parse get_flow_status args: %w", err)
	}

	verbose := action.Verbose.Bool()

	switch action.Detail {
	case FlowStatusDetailSummary:
		return t.buildSummary(ctx, verbose)
	case FlowStatusDetailTasks:
		return t.buildTasksList(ctx, verbose)
	case FlowStatusDetailSubtasks:
		return t.buildSubtasksList(ctx, action.TaskID.PtrInt64(), verbose)
	case FlowStatusDetailRunning:
		return t.buildRunningInfo(ctx, verbose)
	case FlowStatusDetailPlanned:
		return t.buildPlannedList(ctx, action.TaskID.PtrInt64(), verbose)
	default:
		return "", fmt.Errorf("unknown detail level %q; use one of: summary, tasks, subtasks, running, planned", action.Detail)
	}
}

func (t *flowStatusTool) buildSummary(ctx context.Context, verbose bool) (string, error) {
	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get flow tasks: %w", err)
	}

	subtasks, err := t.db.GetFlowSubtasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get flow subtasks: %w", err)
	}

	taskCounts := map[string]int{"created": 0, "running": 0, "waiting": 0, "finished": 0, "failed": 0}
	var activeTask *database.Task
	for i, task := range tasks {
		taskCounts[string(task.Status)]++
		if task.Status == database.TaskStatusRunning || task.Status == database.TaskStatusWaiting {
			activeTask = &tasks[i]
		}
	}

	stCounts := map[string]int{"created": 0, "running": 0, "waiting": 0, "finished": 0, "failed": 0}
	var activeST *database.Subtask
	for i, st := range subtasks {
		stCounts[string(st.Status)]++
		if st.Status == database.SubtaskStatusRunning || st.Status == database.SubtaskStatusWaiting {
			activeST = &subtasks[i]
		}
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Flow ID: %d\n", t.flowID)
	fmt.Fprintf(sb, "Flow status: %s\n", inferFlowStatus(tasks))
	fmt.Fprintf(sb, "Tasks    — total: %d, running: %d, waiting: %d, finished: %d, failed: %d, planned: %d\n",
		len(tasks), taskCounts["running"], taskCounts["waiting"], taskCounts["finished"], taskCounts["failed"], taskCounts["created"])
	fmt.Fprintf(sb, "Subtasks — total: %d, running: %d, waiting: %d, finished: %d, failed: %d, planned: %d\n",
		len(subtasks), stCounts["running"], stCounts["waiting"], stCounts["finished"], stCounts["failed"], stCounts["created"])

	if activeTask != nil {
		fmt.Fprintf(sb, "\nActive task:    ID=%-6d | %-8s | %s\n", activeTask.ID, activeTask.Status, activeTask.Title)
		if verbose && activeTask.Input != "" {
			input, err := t.getInputText(ctx, activeTask.Input)
			if err != nil {
				return "", fmt.Errorf("failed to get active task input: %w", err)
			}
			fmt.Fprintf(sb, "  Input: %s\n", input)
		}
	}
	if activeST != nil {
		fmt.Fprintf(sb, "Active subtask: ID=%-6d | %-8s | %s\n", activeST.ID, activeST.Status, activeST.Title)
		if verbose && activeST.Description != "" {
			description, err := t.getDescriptionText(ctx, activeST.Description)
			if err != nil {
				return "", fmt.Errorf("failed to get active subtask description: %w", err)
			}
			fmt.Fprintf(sb, "  Description: %s\n", description)
		}
	}

	if len(tasks) == 0 {
		fmt.Fprintf(sb, "\nNo tasks yet. Flow is waiting for first input.\n")
	}

	summary := sb.String()
	if t.summarizer != nil && len(summary) > summaryLimit {
		summary, err = t.summarizer(ctx, truncateText(summary, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize summary: %w", err)
		}
	}

	return summary, nil
}

func (t *flowStatusTool) buildTasksList(ctx context.Context, verbose bool) (string, error) {
	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get flow tasks: %w", err)
	}

	if len(tasks) == 0 {
		return "No tasks found for this flow.", nil
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Tasks for flow %d:\n\n", t.flowID)
	for _, task := range tasks {
		fmt.Fprintf(sb, "Task ID: %d | Status: %s | Title: %s\n", task.ID, task.Status, task.Title)
		if verbose {
			if task.Input != "" {
				input, err := t.getInputText(ctx, task.Input)
				if err != nil {
					return "", fmt.Errorf("failed to get task input: %w", err)
				}
				fmt.Fprintf(sb, "  Input:  %s\n", input)
			}
			if task.Result != "" {
				result, err := t.getResultText(ctx, task.Result)
				if err != nil {
					return "", fmt.Errorf("failed to get task result: %w", err)
				}
				fmt.Fprintf(sb, "  Result: %s\n", result)
			}
		}
	}

	taskList := sb.String()
	if t.summarizer != nil && len(taskList) > taskListLimit {
		taskList, err = t.summarizer(ctx, truncateText(taskList, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize task list: %w", err)
		}
	}

	return taskList, nil
}

func (t *flowStatusTool) buildSubtasksList(ctx context.Context, taskID *int64, verbose bool) (string, error) {
	var subtasks []database.Subtask
	var err error

	if taskID != nil && *taskID > 0 {
		subtasks, err = t.db.GetFlowTaskSubtasks(ctx, database.GetFlowTaskSubtasksParams{
			FlowID: t.flowID,
			TaskID: *taskID,
		})
	} else {
		subtasks, err = t.db.GetFlowSubtasks(ctx, t.flowID)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		return "No subtasks found.", nil
	}

	sb := &strings.Builder{}
	if taskID != nil && *taskID > 0 {
		fmt.Fprintf(sb, "Subtasks for task %d:\n\n", *taskID)
	} else {
		fmt.Fprintf(sb, "All subtasks for flow %d:\n\n", t.flowID)
	}
	for _, st := range subtasks {
		fmt.Fprintf(sb, "Subtask ID: %d | Task ID: %d | Status: %s | Title: %s\n",
			st.ID, st.TaskID, st.Status, st.Title)
		if verbose {
			if st.Description != "" {
				description, err := t.getDescriptionText(ctx, st.Description)
				if err != nil {
					return "", fmt.Errorf("failed to get subtask description: %w", err)
				}
				fmt.Fprintf(sb, "  Description: %s\n", description)
			}
			if st.Result != "" {
				result, err := t.getResultText(ctx, st.Result)
				if err != nil {
					return "", fmt.Errorf("failed to get subtask result: %w", err)
				}
				fmt.Fprintf(sb, "  Result: %s\n", result)
			}
		}
	}

	subtasksList := sb.String()
	if t.summarizer != nil && len(subtasksList) > subtasksListLimit {
		subtasksList, err = t.summarizer(ctx, truncateText(subtasksList, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize subtasks list: %w", err)
		}
	}

	return subtasksList, nil
}

func (t *flowStatusTool) buildRunningInfo(ctx context.Context, verbose bool) (string, error) {
	subtasks, err := t.db.GetFlowSubtasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get subtasks: %w", err)
	}

	for _, st := range subtasks {
		if st.Status != database.SubtaskStatusRunning && st.Status != database.SubtaskStatusWaiting {
			continue
		}

		// Load the parent task to give the full Task→Subtask chain.
		tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
		if err != nil {
			return "", fmt.Errorf("failed to get flow tasks: %w", err)
		}
		var activeTask *database.Task
		for i, task := range tasks {
			if task.ID == st.TaskID {
				activeTask = &tasks[i]
				break
			}
		}

		sb := &strings.Builder{}

		if activeTask != nil {
			fmt.Fprintf(sb, "=== Active Task ===\n")
			fmt.Fprintf(sb, "Task ID: %d | Status: %s | Title: %s\n", activeTask.ID, activeTask.Status, activeTask.Title)
			if activeTask.Input != "" {
				input, err := t.getInputText(ctx, activeTask.Input)
				if err != nil {
					return "", fmt.Errorf("failed to get active task input: %w", err)
				}
				fmt.Fprintf(sb, "Input:\n%s\n", input)
			}
			if activeTask.Result != "" {
				result, err := t.getResultText(ctx, activeTask.Result)
				if err != nil {
					return "", fmt.Errorf("failed to get active task result: %w", err)
				}
				fmt.Fprintf(sb, "Result so far:\n%s\n", result)
			}
		}

		fmt.Fprintf(sb, "\n=== Active Subtask ===\n")
		fmt.Fprintf(sb, "Subtask ID: %d | Status: %s | Title: %s\n", st.ID, st.Status, st.Title)
		description, err := t.getDescriptionText(ctx, st.Description)
		if err != nil {
			return "", fmt.Errorf("failed to get subtask description: %w", err)
		}
		fmt.Fprintf(sb, "Description:\n%s\n", description)
		if st.Result != "" {
			result, err := t.getResultText(ctx, st.Result)
			if err != nil {
				return "", fmt.Errorf("failed to get subtask result: %w", err)
			}
			fmt.Fprintf(sb, "Result so far:\n%s\n", result)
		}
		if verbose && st.Context != "" {
			execContext, err := t.getDescriptionText(ctx, st.Context)
			if err != nil {
				return "", fmt.Errorf("failed to get subtask context: %w", err)
			}
			fmt.Fprintf(sb, "\nExecution context:\n%s\n", execContext)
		}
		if st.Status == database.SubtaskStatusWaiting {
			fmt.Fprintf(sb, "\nNote: subtask is waiting for user input (ask state).\n")
			fmt.Fprintf(sb, "Use %s to provide the answer and resume execution.\n", SubmitFlowInputToolName)
		}

		msgLimit := msgLogLimitNormal
		if verbose {
			msgLimit = msgLogLimitVerbose
		}
		msgLogs, err := t.appendSubtaskMsgLogs(ctx, st.ID, msgLimit)
		if err != nil {
			return "", fmt.Errorf("failed to get subtask msg logs: %w", err)
		}
		sb.WriteString(msgLogs)

		runningInfo := sb.String()
		if t.summarizer != nil && len(runningInfo) > runningInfoLimit {
			runningInfo, err = t.summarizer(ctx, truncateText(runningInfo, summarizationLimit))
			if err != nil {
				return "", fmt.Errorf("failed to summarize running info: %w", err)
			}
		}

		return runningInfo, nil
	}

	return "No running or waiting subtask found. Flow is idle (waiting for next input).", nil
}

func (t *flowStatusTool) buildPlannedList(ctx context.Context, taskID *int64, verbose bool) (string, error) {
	var subtasks []database.Subtask
	var err error

	if taskID != nil && *taskID > 0 {
		subtasks, err = t.db.GetTaskPlannedSubtasks(ctx, *taskID)
		if err != nil {
			return "", fmt.Errorf("failed to get planned subtasks: %w", err)
		}
	} else {
		all, err2 := t.db.GetFlowSubtasks(ctx, t.flowID)
		if err2 != nil {
			return "", fmt.Errorf("failed to get flow subtasks: %w", err2)
		}
		for _, st := range all {
			if st.Status == database.SubtaskStatusCreated {
				subtasks = append(subtasks, st)
			}
		}
	}

	if len(subtasks) == 0 {
		return "No planned (created) subtasks found. All subtasks have been started or completed.", nil
	}

	sb := &strings.Builder{}
	if taskID != nil && *taskID > 0 {
		fmt.Fprintf(sb, "Planned subtasks for task %d:\n\n", *taskID)
	} else {
		fmt.Fprintf(sb, "All planned subtasks for flow %d:\n\n", t.flowID)
	}
	for _, st := range subtasks {
		fmt.Fprintf(sb, "Subtask ID: %d | Task ID: %d | Title: %s\n", st.ID, st.TaskID, st.Title)
		if st.Description != "" {
			if verbose {
				description, err := t.getDescriptionText(ctx, st.Description)
				if err != nil {
					return "", fmt.Errorf("failed to get subtask description: %w", err)
				}
				fmt.Fprintf(sb, "  Description: %s\n", description)
			} else {
				fmt.Fprintf(sb, "  Description: %s\n", truncateText(st.Description, 200))
			}
		}
	}

	plannedList := sb.String()
	if t.summarizer != nil && len(plannedList) > plannedListLimit {
		plannedList, err = t.summarizer(ctx, truncateText(plannedList, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize planned list: %w", err)
		}
	}

	return plannedList, nil
}

// appendSubtaskMsgLogs appends the last `limit` agent messages for a subtask to the builder.
func (t *flowStatusTool) appendSubtaskMsgLogs(ctx context.Context, subtaskID int64, limit int) (string, error) {
	nullID := database.Int64ToNullInt64(&subtaskID)
	logs, err := t.db.GetSubtaskMsgLogs(ctx, nullID)
	if err != nil {
		return "", fmt.Errorf("failed to get subtask msg logs: %w", err)
	}

	if len(logs) == 0 {
		return fmt.Sprintf("No agent messages found for subtask %d.", subtaskID), nil
	}

	start := 0
	if len(logs) > limit {
		start = len(logs) - limit
	}
	logs = logs[start:]

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "\n=== Last %d agent messages ===\n", len(logs))
	for _, log := range logs {
		fmt.Fprintf(sb, "[%s] %s\n", log.Type, truncateText(log.Message, 1024))
		if log.Result != "" {
			fmt.Fprintf(sb, "  → %s\n", truncateText(log.Result, 300))
		}
	}

	logsList := sb.String()
	if t.summarizer != nil && len(logsList) > msgLogsLimit {
		logsList, err = t.summarizer(ctx, truncateText(logsList, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize msg logs: %w", err)
		}
	}

	return logsList, nil
}

func (t *flowStatusTool) getInputText(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", nil
	}

	if t.summarizer != nil && len(input) > 2*inputLimit {
		summarized, err := t.summarizer(ctx, truncateText(input, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize input: %w", err)
		}
		return summarized, nil
	}

	return truncateText(input, inputLimit), nil
}

func (t *flowStatusTool) getDescriptionText(ctx context.Context, description string) (string, error) {
	if description == "" {
		return "", nil
	}

	if t.summarizer != nil && len(description) > 2*descriptionLimit {
		summarized, err := t.summarizer(ctx, truncateText(description, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize description: %w", err)
		}
		return summarized, nil
	}

	return truncateText(description, descriptionLimit), nil
}

func (t *flowStatusTool) getResultText(ctx context.Context, result string) (string, error) {
	if result == "" {
		return "", nil
	}

	if t.summarizer != nil && len(result) > 2*resultLimit {
		summarized, err := t.summarizer(ctx, truncateText(result, summarizationLimit))
		if err != nil {
			return "", fmt.Errorf("failed to summarize result: %w", err)
		}
		return summarized, nil
	}

	return truncateText(result, resultLimit), nil
}

// inferFlowStatus derives a human-readable flow status from the task list.
func inferFlowStatus(tasks []database.Task) string {
	for _, task := range tasks {
		if task.Status == database.TaskStatusRunning {
			return "running"
		}
	}
	for _, task := range tasks {
		if task.Status == database.TaskStatusWaiting {
			return "waiting (subtask asking for user input)"
		}
	}
	return "waiting (ready for next input)"
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// stopFlowTool implements stop_flow.
type stopFlowTool struct {
	flowID  int64
	db      database.Querier
	handler func(ctx context.Context, reason string) error
}

func NewStopFlowTool(flowID int64, db database.Querier, handler func(ctx context.Context, reason string) error) *stopFlowTool {
	return &stopFlowTool{flowID: flowID, db: db, handler: handler}
}

func (t *stopFlowTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action StopFlowAction
	if err := json.Unmarshal(args, &action); err != nil {
		return "", fmt.Errorf("failed to parse stop_flow args: %w", err)
	}

	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to check flow status: %w", err)
	}

	isRunning := false
	for _, task := range tasks {
		if task.Status == database.TaskStatusRunning {
			isRunning = true
			break
		}
	}

	if !isRunning {
		return "No running task found — the flow is already in 'waiting' state and ready to accept input.", nil
	}

	stopCtx, stopCancel := context.WithTimeout(ctx, flowOperationTimeout)
	defer stopCancel()

	if err := t.handler(stopCtx, action.Reason); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", fmt.Errorf(
				"stop timed out after %s — the task may still be winding down. "+
					"Call %s to check the current state before proceeding",
				flowOperationTimeout, GetFlowStatusToolName)
		}
		return "", fmt.Errorf("failed to stop flow: %w", err)
	}

	// Re-query task statuses to report the actual state the flow reached.
	newTasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		// Stop succeeded but we cannot verify the new state — return a safe partial message.
		return fmt.Sprintf(
			"Flow stop initiated (reason: %s). Could not verify new status: %s. "+
				"Call %s to confirm before proceeding.",
			action.Reason, err, GetFlowStatusToolName), nil
	}

	for _, task := range newTasks {
		if task.Status == database.TaskStatusRunning {
			return fmt.Sprintf(
				"Stop requested (reason: %s), but a task is still running — the flow has not reached 'waiting' yet. "+
					"Call %s to confirm before making further changes.",
				action.Reason, GetFlowStatusToolName), nil
		}
	}

	return fmt.Sprintf(
		"Flow stopped successfully (reason: %s). The flow is now in 'waiting' state and ready to accept new input or subtask patches.",
		action.Reason), nil
}

// submitFlowInputTool implements submit_flow_input.
type submitFlowInputTool struct {
	flowID       int64
	db           database.Querier
	handler      func(ctx context.Context, input string) error
	pollInterval time.Duration
	pollTimeout  time.Duration
}

func NewSubmitFlowInputTool(flowID int64, db database.Querier, handler func(ctx context.Context, input string) error) *submitFlowInputTool {
	return &submitFlowInputTool{
		flowID:       flowID,
		db:           db,
		handler:      handler,
		pollInterval: taskReadyPollInterval,
		pollTimeout:  taskReadyPollTimeout,
	}
}

func (t *submitFlowInputTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SubmitFlowInputAction
	if err := json.Unmarshal(args, &action); err != nil {
		return "", fmt.Errorf("failed to parse %s args: %w", SubmitFlowInputToolName, err)
	}

	if action.Input == "" {
		return "", fmt.Errorf("input must not be empty")
	}

	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to check flow status: %w", err)
	}

	for _, task := range tasks {
		if task.Status == database.TaskStatusRunning {
			return "", fmt.Errorf(
				"task %q (ID: %d) is currently running; "+
					"cannot submit input while a task is executing. "+
					"Call %s first to cancel it, then retry %s",
				task.Title, task.ID, StopFlowToolName, SubmitFlowInputToolName)
		}
	}

	// Determine mode for better response message
	waitingForAsk := false
	subtasks, err := t.db.GetFlowSubtasks(ctx, t.flowID)
	if err == nil {
		for _, st := range subtasks {
			if st.Status == database.SubtaskStatusWaiting {
				waitingForAsk = true
				break
			}
		}
	}

	inputCtx, inputCancel := context.WithTimeout(ctx, flowOperationTimeout)
	defer inputCancel()

	if err := t.handler(inputCtx, action.Input); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", fmt.Errorf(
				"submit timed out after %s — the flow may not have received the input. "+
					"Call %s to check the current state before retrying",
				flowOperationTimeout, GetFlowStatusToolName)
		}
		// A missing message chain means the waiting subtask's execution context was lost
		// (most likely after a system restart or database cleanup while the subtask was at an ask checkpoint).
		// The subtask can no longer be resumed; the flow must be stopped and restarted.
		if strings.Contains(err.Error(), "no rows in result set") {
			return "", fmt.Errorf(
				"the waiting subtask's execution context is no longer available in the database — "+
					"this typically happens after a system restart when the subtask was paused at an ask checkpoint. "+
					"Recovery: (1) call %s to cancel the stale task; "+
					"(2) call %s with a fresh description of what to do next; "+
					"if the flow already has planned subtasks that should still run, "+
					"use %s to inspect the remaining plan and %s to patch it before resuming",
				StopFlowToolName, SubmitFlowInputToolName,
				GetFlowStatusToolName, PatchFlowSubtasksToolName)
		}
		return "", fmt.Errorf("failed to submit flow input: %w", err)
	}

	if waitingForAsk {
		return "Input delivered as the answer to the waiting subtask's question. The subtask will resume execution.", nil
	}

	// Input triggered task creation — wait for the generator to produce a running task.
	return t.waitForTaskReady(ctx)
}

// waitForTaskReady polls GetFlowTasks every pollInterval until a running or waiting
// task appears or pollTimeout is reached. It is called after submit_flow_input
// triggers new task creation so the LLM gets confirmation that the generator finished.
func (t *submitFlowInputTool) waitForTaskReady(ctx context.Context) (string, error) {
	deadline := time.Now().Add(t.pollTimeout)
	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	for {
		tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
		if err == nil {
			for _, task := range tasks {
				if task.Status == database.TaskStatusRunning || task.Status == database.TaskStatusWaiting {
					return fmt.Sprintf(
						"Input accepted. Task %q (ID: %d) is now running — the generator has produced its subtask plan. "+
							"Call %s with detail='running' to see what the agent is doing.",
						task.Title, task.ID, GetFlowStatusToolName), nil
				}
			}
		}

		if time.Now().After(deadline) {
			break
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf(
				"context cancelled while waiting for task to start. "+
					"Call %s to check whether a task was created",
				GetFlowStatusToolName)
		case <-ticker.C:
		}
	}

	return fmt.Sprintf(
		"Input accepted but no running task appeared within %s. "+
			"The generator may still be working. Call %s to check the current state.",
		t.pollTimeout, GetFlowStatusToolName), nil
}

// patchFlowSubtasksTool implements patch_flow_subtasks.
type patchFlowSubtasksTool struct {
	flowID  int64
	db      database.Querier
	handler func(ctx context.Context, taskID int64, patch SubtaskPatch) error
}

func NewPatchFlowSubtasksTool(
	flowID int64,
	db database.Querier,
	handler func(ctx context.Context, taskID int64, patch SubtaskPatch) error,
) *patchFlowSubtasksTool {
	return &patchFlowSubtasksTool{flowID: flowID, db: db, handler: handler}
}

func (t *patchFlowSubtasksTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action PatchFlowSubtasksAction
	if err := json.Unmarshal(args, &action); err != nil {
		return "", fmt.Errorf("failed to parse patch_flow_subtasks args: %w", err)
	}

	if action.TaskID <= 0 {
		return "", fmt.Errorf("task_id must be a positive integer")
	}

	// Validate flow is not running
	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to check flow status: %w", err)
	}

	for _, task := range tasks {
		if task.Status == database.TaskStatusRunning {
			return "", fmt.Errorf(
				"task %q (ID: %d) is currently running; "+
					"patching is not allowed while a task is executing. "+
					"Call %s first, then retry %s",
				task.Title, task.ID, StopFlowToolName, PatchFlowSubtasksToolName)
		}
	}

	// Verify task_id belongs to this flow.
	taskBelongsToFlow := false
	for _, task := range tasks {
		if task.ID == action.TaskID {
			taskBelongsToFlow = true
			break
		}
	}
	if !taskBelongsToFlow {
		return "", fmt.Errorf(
			"task ID %d was not found in this flow; "+
				"obtain a valid task ID from %s with detail='tasks'",
			action.TaskID, GetFlowStatusToolName)
	}

	// Verify there are planned subtasks to patch.
	planned, err := t.db.GetTaskPlannedSubtasks(ctx, action.TaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get planned subtasks for task %d: %w", action.TaskID, err)
	}

	if len(planned) == 0 && len(action.Operations) > 0 {
		subtasks, err := t.db.GetFlowSubtasks(ctx, t.flowID)
		if err != nil {
			return "", fmt.Errorf("failed to check subtask state for task %d: %w", action.TaskID, err)
		}

		for _, st := range subtasks {
			if st.TaskID != action.TaskID {
				continue
			}
			switch st.Status {
			case database.SubtaskStatusWaiting:
				return "", fmt.Errorf(
					"task %d has a subtask (ID: %d, %q) waiting for user input. "+
						"Only 'created' subtasks can be patched, but you can include the waiting subtask's ID in your operations to modify or remove it. "+
						"Alternatively, answer it via %s first",
					action.TaskID, st.ID, st.Title, SubmitFlowInputToolName)
			case database.SubtaskStatusRunning:
				return "", fmt.Errorf(
					"task %d has a subtask (ID: %d, %q) currently running. "+
						"Call %s first, then retry",
					action.TaskID, st.ID, st.Title, StopFlowToolName)
			}
		}

		return "", fmt.Errorf(
			"no 'created' subtasks found for task %d; "+
				"all subtasks have been executed or the task has no plan yet. "+
				"Use %s to create a new task instead",
			action.TaskID, SubmitFlowInputToolName)
	}

	patch := SubtaskPatch{
		Operations: action.Operations,
		Message:    action.Message,
	}

	if err := patch.Validate(); err != nil {
		return "", fmt.Errorf("invalid subtask patch: %w", err)
	}

	if len(action.Operations) == 0 {
		return fmt.Sprintf("No operations provided — the subtask plan for task %d is unchanged.", action.TaskID), nil
	}

	if err := t.handler(ctx, action.TaskID, patch); err != nil {
		return "", fmt.Errorf("failed to patch subtasks for task %d: %w", action.TaskID, err)
	}

	// Query the new subtask list so the LLM can correlate the patched entries with their new IDs.
	newPlanned, err := t.db.GetTaskPlannedSubtasks(ctx, action.TaskID)
	if err != nil {
		// Not fatal — operations were applied; just warn and skip the list.
		return fmt.Sprintf(
			"%d operation(s) applied to the subtask plan for task %d. "+
				"Could not retrieve updated subtask list: %s. "+
				"Call %s with detail='planned' and task_id=%d to verify.",
			len(action.Operations), action.TaskID, err,
			GetFlowStatusToolName, action.TaskID), nil
	}

	sb := &strings.Builder{}
	fmt.Fprintf(sb,
		"%d operation(s) applied to the subtask plan for task %d. "+
			"Updated planned subtasks (new IDs assigned after recreation):\n",
		len(action.Operations), action.TaskID)
	for _, st := range newPlanned {
		fmt.Fprintf(sb, "  ID: %d | %s\n", st.ID, st.Title)
	}
	fmt.Fprintf(sb,
		"Call %s to start execution with the updated plan.",
		SubmitFlowInputToolName)

	return sb.String(), nil
}
