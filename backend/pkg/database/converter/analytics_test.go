package converter

import (
	"database/sql"
	"math"
	"pentagi/pkg/database"
	"testing"
	"time"
)

// Helper functions for test data creation

func makeSubtask(id int64, status database.SubtaskStatus, createdAt, updatedAt time.Time) database.Subtask {
	return database.Subtask{
		ID:        id,
		Status:    status,
		Title:     "Test Subtask",
		CreatedAt: sql.NullTime{Time: createdAt, Valid: true},
		UpdatedAt: sql.NullTime{Time: updatedAt, Valid: true},
		TaskID:    1,
	}
}

func makeMsgchain(id int64, msgType database.MsgchainType, taskID int64, subtaskID *int64, createdAt, updatedAt time.Time) database.Msgchain {
	mc := database.Msgchain{
		ID:              id,
		Type:            msgType,
		TaskID:          sql.NullInt64{Int64: taskID, Valid: true},
		CreatedAt:       sql.NullTime{Time: createdAt, Valid: true},
		UpdatedAt:       sql.NullTime{Time: updatedAt, Valid: true},
		DurationSeconds: updatedAt.Sub(createdAt).Seconds(),
	}
	if subtaskID != nil {
		mc.SubtaskID = sql.NullInt64{Int64: *subtaskID, Valid: true}
	}
	return mc
}

func makeToolcall(id int64, status database.ToolcallStatus, taskID, subtaskID *int64, createdAt, updatedAt time.Time) database.Toolcall {
	tc := database.Toolcall{
		ID:        id,
		Status:    status,
		CreatedAt: sql.NullTime{Time: createdAt, Valid: true},
		UpdatedAt: sql.NullTime{Time: updatedAt, Valid: true},
	}
	// Duration is only set for finished/failed toolcalls
	if status == database.ToolcallStatusFinished || status == database.ToolcallStatusFailed {
		tc.DurationSeconds = updatedAt.Sub(createdAt).Seconds()
	} else {
		tc.DurationSeconds = 0
	}
	if taskID != nil {
		tc.TaskID = sql.NullInt64{Int64: *taskID, Valid: true}
	}
	if subtaskID != nil {
		tc.SubtaskID = sql.NullInt64{Int64: *subtaskID, Valid: true}
	}
	return tc
}

// ========== Subtask Duration Tests ==========

func TestCalculateSubtaskDuration_CreatedStatus(t *testing.T) {
	now := time.Now()
	subtask := makeSubtask(1, database.SubtaskStatusCreated, now, now.Add(10*time.Second))

	duration := CalculateSubtaskDuration(subtask, nil)

	if duration != 0 {
		t.Errorf("Expected 0 for created subtask, got %f", duration)
	}
}

func TestCalculateSubtaskDuration_WaitingStatus(t *testing.T) {
	now := time.Now()
	subtask := makeSubtask(1, database.SubtaskStatusWaiting, now, now.Add(10*time.Second))

	duration := CalculateSubtaskDuration(subtask, nil)

	if duration != 0 {
		t.Errorf("Expected 0 for waiting subtask, got %f", duration)
	}
}

func TestCalculateSubtaskDuration_FinishedStatus(t *testing.T) {
	now := time.Now()
	subtask := makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(100*time.Second))

	duration := CalculateSubtaskDuration(subtask, nil)

	if duration < 99 || duration > 101 {
		t.Errorf("Expected ~100 seconds, got %f", duration)
	}
}

func TestCalculateSubtaskDuration_WithMsgchainValidation(t *testing.T) {
	now := time.Now()
	subtaskID := int64(1)

	// Subtask shows 100 seconds
	subtask := makeSubtask(subtaskID, database.SubtaskStatusFinished, now, now.Add(100*time.Second))

	// But msgchain shows only 50 seconds (more accurate)
	msgchains := []database.Msgchain{
		makeMsgchain(1, database.MsgchainTypePrimaryAgent, 1, &subtaskID, now, now.Add(50*time.Second)),
	}

	duration := CalculateSubtaskDuration(subtask, msgchains)

	// Should use minimum (msgchain duration)
	if duration < 49 || duration > 51 {
		t.Errorf("Expected ~50 seconds (msgchain), got %f", duration)
	}
}

// ========== Overlap Compensation Tests ==========

func TestCalculateSubtasksWithOverlapCompensation_NoOverlap(t *testing.T) {
	now := time.Now()

	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now.Add(10*time.Second), now.Add(20*time.Second)),
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, nil)

	// Check individual subtask durations
	if durations[1] < 9 || durations[1] > 11 {
		t.Errorf("Expected subtask 1 duration ~10s, got %f", durations[1])
	}

	if durations[2] < 9 || durations[2] > 11 {
		t.Errorf("Expected subtask 2 duration ~10s, got %f", durations[2])
	}

	// Check total
	total := durations[1] + durations[2]
	expected := 20.0
	if total < expected-1 || total > expected+1 {
		t.Errorf("Expected total ~%f seconds, got %f", expected, total)
	}
}

func TestCalculateSubtasksWithOverlapCompensation_WithOverlap(t *testing.T) {
	now := time.Now()

	// Both subtasks created at the same time (batch creation)
	// but executed sequentially
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now, now.Add(20*time.Second)), // Same start time!
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, nil)

	// Subtask 1: should be 10s (no compensation needed)
	if durations[1] < 9 || durations[1] > 11 {
		t.Errorf("Expected subtask 1 duration ~10s, got %f", durations[1])
	}

	// Subtask 2: should be 10s (compensated from 20s)
	// Original: 10:00:20 - 10:00:00 = 20s
	// Compensated: 10:00:20 - 10:00:10 = 10s (starts when subtask 1 finished)
	if durations[2] < 9 || durations[2] > 11 {
		t.Errorf("Expected subtask 2 duration ~10s (compensated), got %f", durations[2])
	}

	// Total should be 20s (real wall-clock time)
	total := durations[1] + durations[2]
	expected := 20.0
	if total < expected-1 || total > expected+1 {
		t.Errorf("Expected total ~%f seconds (compensated), got %f", expected, total)
	}
}

func TestCalculateSubtasksWithOverlapCompensation_IgnoresCreated(t *testing.T) {
	now := time.Now()

	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusCreated, now, now.Add(100*time.Second)), // Should be ignored
		makeSubtask(3, database.SubtaskStatusFinished, now.Add(10*time.Second), now.Add(20*time.Second)),
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, nil)

	// Check created subtask is 0
	if durations[2] != 0 {
		t.Errorf("Expected created subtask duration 0, got %f", durations[2])
	}

	// Check finished subtasks
	if durations[1] < 9 || durations[1] > 11 {
		t.Errorf("Expected subtask 1 duration ~10s, got %f", durations[1])
	}

	if durations[3] < 9 || durations[3] > 11 {
		t.Errorf("Expected subtask 3 duration ~10s, got %f", durations[3])
	}

	total := durations[1] + durations[2] + durations[3]
	expected := 20.0 // Only subtasks 1 and 3
	if total < expected-1 || total > expected+1 {
		t.Errorf("Expected total ~%f seconds, got %f", expected, total)
	}
}

// ========== Task Duration Tests ==========

func TestCalculateTaskDuration_OnlySubtasks(t *testing.T) {
	now := time.Now()

	task := database.Task{ID: 1}
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
	}

	duration := CalculateTaskDuration(task, subtasks, nil)

	expected := 10.0
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

func TestCalculateTaskDuration_WithGenerator(t *testing.T) {
	now := time.Now()

	task := database.Task{ID: 1}
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now.Add(5*time.Second), now.Add(15*time.Second)),
	}
	msgchains := []database.Msgchain{
		makeMsgchain(1, database.MsgchainTypeGenerator, 1, nil, now, now.Add(5*time.Second)),
	}

	duration := CalculateTaskDuration(task, subtasks, msgchains)

	// 5s generator + 10s subtask = 15s
	expected := 15.0
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

func TestCalculateTaskDuration_WithRefiner(t *testing.T) {
	now := time.Now()

	task := database.Task{ID: 1}
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now.Add(13*time.Second), now.Add(23*time.Second)),
	}
	msgchains := []database.Msgchain{
		// Refiner runs between subtasks
		makeMsgchain(1, database.MsgchainTypeRefiner, 1, nil, now.Add(10*time.Second), now.Add(13*time.Second)),
	}

	duration := CalculateTaskDuration(task, subtasks, msgchains)

	// 10s subtask1 + 3s refiner + 10s subtask2 = 23s
	expected := 23.0
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

// ========== Toolcalls Count Tests ==========

func TestCountFinishedToolcalls(t *testing.T) {
	now := time.Now()

	toolcalls := []database.Toolcall{
		makeToolcall(1, database.ToolcallStatusFinished, nil, nil, now, now.Add(time.Second)),
		makeToolcall(2, database.ToolcallStatusFailed, nil, nil, now, now.Add(time.Second)),
		makeToolcall(3, database.ToolcallStatusReceived, nil, nil, now, now.Add(time.Second)),
		makeToolcall(4, database.ToolcallStatusRunning, nil, nil, now, now.Add(time.Second)),
	}

	count := CountFinishedToolcalls(toolcalls)

	if count != 2 {
		t.Errorf("Expected 2 finished toolcalls, got %d", count)
	}
}

func TestCountFinishedToolcallsForSubtask(t *testing.T) {
	now := time.Now()
	subtaskID := int64(1)
	otherSubtaskID := int64(2)

	toolcalls := []database.Toolcall{
		makeToolcall(1, database.ToolcallStatusFinished, nil, &subtaskID, now, now.Add(time.Second)),
		makeToolcall(2, database.ToolcallStatusFinished, nil, &otherSubtaskID, now, now.Add(time.Second)),
		makeToolcall(3, database.ToolcallStatusReceived, nil, &subtaskID, now, now.Add(time.Second)),
	}

	count := CountFinishedToolcallsForSubtask(toolcalls, subtaskID)

	if count != 1 {
		t.Errorf("Expected 1 finished toolcall for subtask, got %d", count)
	}
}

func TestCountFinishedToolcallsForTask(t *testing.T) {
	now := time.Now()
	taskID := int64(1)
	subtaskID := int64(1)

	toolcalls := []database.Toolcall{
		makeToolcall(1, database.ToolcallStatusFinished, &taskID, nil, now, now.Add(time.Second)),    // Task-level
		makeToolcall(2, database.ToolcallStatusFinished, nil, &subtaskID, now, now.Add(time.Second)), // Subtask-level
		makeToolcall(3, database.ToolcallStatusReceived, &taskID, nil, now, now.Add(time.Second)),    // Not finished
	}

	count := CountFinishedToolcallsForTask(toolcalls, taskID, []int64{subtaskID})

	if count != 2 {
		t.Errorf("Expected 2 finished toolcalls for task, got %d", count)
	}
}

// ========== Flow Duration Tests ==========

func TestCalculateFlowDuration_WithTasks(t *testing.T) {
	now := time.Now()

	tasks := []database.Task{
		{ID: 1},
	}

	subtasksMap := map[int64][]database.Subtask{
		1: {
			makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		},
	}

	msgchainsMap := map[int64][]database.Msgchain{
		1: {},
	}

	assistantMsgchains := []database.Msgchain{}

	duration := CalculateFlowDuration(tasks, subtasksMap, msgchainsMap, assistantMsgchains)

	expected := 10.0
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

func TestCalculateFlowDuration_WithAssistantMsgchains(t *testing.T) {
	now := time.Now()

	tasks := []database.Task{
		{ID: 1},
	}

	subtasksMap := map[int64][]database.Subtask{
		1: {
			makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		},
	}

	msgchainsMap := map[int64][]database.Msgchain{
		1: {},
	}

	// Assistant msgchains without task/subtask binding (flow-level)
	assistantMsgchain := database.Msgchain{
		ID:              1,
		Type:            database.MsgchainTypeAssistant,
		FlowID:          1,
		TaskID:          sql.NullInt64{Valid: false},
		SubtaskID:       sql.NullInt64{Valid: false},
		DurationSeconds: 5.0,
		CreatedAt:       sql.NullTime{Time: now.Add(20 * time.Second), Valid: true},
		UpdatedAt:       sql.NullTime{Time: now.Add(25 * time.Second), Valid: true},
	}
	assistantMsgchains := []database.Msgchain{assistantMsgchain}

	duration := CalculateFlowDuration(tasks, subtasksMap, msgchainsMap, assistantMsgchains)

	expected := 15.0 // 10s task + 5s assistant msgchain
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

func TestCalculateFlowDuration_IgnoresMsgchainsWithTaskOrSubtask(t *testing.T) {
	now := time.Now()
	taskID := int64(1)
	subtaskID := int64(1)

	tasks := []database.Task{}
	subtasksMap := map[int64][]database.Subtask{}
	msgchainsMap := map[int64][]database.Msgchain{}

	// These should be ignored in flow-level calculation (have task/subtask binding)
	assistantMsgchainWithTask := database.Msgchain{
		ID:              1,
		Type:            database.MsgchainTypeAssistant,
		FlowID:          1,
		TaskID:          sql.NullInt64{Int64: taskID, Valid: true},
		SubtaskID:       sql.NullInt64{Valid: false},
		DurationSeconds: 10.0,
		CreatedAt:       sql.NullTime{Time: now, Valid: true},
		UpdatedAt:       sql.NullTime{Time: now.Add(10 * time.Second), Valid: true},
	}

	assistantMsgchainWithSubtask := database.Msgchain{
		ID:              2,
		Type:            database.MsgchainTypeAssistant,
		FlowID:          1,
		TaskID:          sql.NullInt64{Valid: false},
		SubtaskID:       sql.NullInt64{Int64: subtaskID, Valid: true},
		DurationSeconds: 10.0,
		CreatedAt:       sql.NullTime{Time: now, Valid: true},
		UpdatedAt:       sql.NullTime{Time: now.Add(10 * time.Second), Valid: true},
	}

	// Only this one should count (no task/subtask binding)
	assistantMsgchainFlowLevel := database.Msgchain{
		ID:              3,
		Type:            database.MsgchainTypeAssistant,
		FlowID:          1,
		TaskID:          sql.NullInt64{Valid: false},
		SubtaskID:       sql.NullInt64{Valid: false},
		DurationSeconds: 5.0,
		CreatedAt:       sql.NullTime{Time: now, Valid: true},
		UpdatedAt:       sql.NullTime{Time: now.Add(5 * time.Second), Valid: true},
	}

	assistantMsgchains := []database.Msgchain{
		assistantMsgchainWithTask,
		assistantMsgchainWithSubtask,
		assistantMsgchainFlowLevel,
	}

	duration := CalculateFlowDuration(tasks, subtasksMap, msgchainsMap, assistantMsgchains)

	expected := 5.0
	if duration < expected-1 || duration > expected+1 {
		t.Errorf("Expected ~%f seconds, got %f", expected, duration)
	}
}

// ========== Mathematical Correctness Tests ==========

func TestSubtasksSumEqualsTaskSubtasksPart(t *testing.T) {
	now := time.Now()

	task := database.Task{ID: 1}

	// Create subtasks with batch creation (same created_at)
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now, now.Add(20*time.Second)),
		makeSubtask(3, database.SubtaskStatusFinished, now, now.Add(30*time.Second)),
	}

	// No generator/refiner for this test
	msgchains := []database.Msgchain{}

	// Calculate individual subtask durations with compensation
	compensatedDurations := CalculateSubtasksWithOverlapCompensation(subtasks, msgchains)

	// Sum individual subtask durations
	var subtasksSum float64
	for _, duration := range compensatedDurations {
		subtasksSum += duration
	}

	// Calculate task duration (should equal subtasks sum since no generator/refiner)
	taskDuration := CalculateTaskDuration(task, subtasks, msgchains)

	// They should be equal
	if math.Abs(subtasksSum-taskDuration) > 0.1 {
		t.Errorf("Subtasks sum (%f) should equal task duration (%f)", subtasksSum, taskDuration)
	}
}

func TestCompensation_ExtremeBatchCreation(t *testing.T) {
	now := time.Now()

	// All 5 subtasks created at exactly the same time
	// Execute sequentially, 10 seconds each
	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now, now.Add(20*time.Second)),
		makeSubtask(3, database.SubtaskStatusFinished, now, now.Add(30*time.Second)),
		makeSubtask(4, database.SubtaskStatusFinished, now, now.Add(40*time.Second)),
		makeSubtask(5, database.SubtaskStatusFinished, now, now.Add(50*time.Second)),
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, nil)

	// Each should be ~10 seconds (compensated)
	for i := int64(1); i <= 5; i++ {
		if durations[i] < 9 || durations[i] > 11 {
			t.Errorf("Expected subtask %d duration ~10s, got %f", i, durations[i])
		}
	}

	// Total should be 50s (real wall-clock)
	var total float64
	for _, d := range durations {
		total += d
	}

	expected := 50.0
	if total < expected-1 || total > expected+1 {
		t.Errorf("Expected total ~%f seconds, got %f", expected, total)
	}
}

func TestCompensation_MixedStatus(t *testing.T) {
	now := time.Now()

	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusCreated, now, now.Add(100*time.Second)), // Not started
		makeSubtask(3, database.SubtaskStatusWaiting, now, now.Add(100*time.Second)), // Waiting
		makeSubtask(4, database.SubtaskStatusFinished, now, now.Add(25*time.Second)), // After subtask 1
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, nil)

	// Subtask 1: 10s
	if durations[1] < 9 || durations[1] > 11 {
		t.Errorf("Expected subtask 1 duration ~10s, got %f", durations[1])
	}

	// Subtask 2: 0 (created)
	if durations[2] != 0 {
		t.Errorf("Expected subtask 2 duration 0, got %f", durations[2])
	}

	// Subtask 3: 0 (waiting)
	if durations[3] != 0 {
		t.Errorf("Expected subtask 3 duration 0, got %f", durations[3])
	}

	// Subtask 4: should be compensated to start after subtask 1
	// Original: 10:00:25 - 10:00:00 = 25s
	// Compensated: 10:00:25 - 10:00:10 = 15s
	if durations[4] < 14 || durations[4] > 16 {
		t.Errorf("Expected subtask 4 duration ~15s (compensated), got %f", durations[4])
	}
}

func TestCompensation_WithMsgchainValidation(t *testing.T) {
	now := time.Now()
	subtaskID1 := int64(1)
	subtaskID2 := int64(2)

	subtasks := []database.Subtask{
		makeSubtask(1, database.SubtaskStatusFinished, now, now.Add(10*time.Second)),
		makeSubtask(2, database.SubtaskStatusFinished, now, now.Add(25*time.Second)), // Overlap
	}

	// Msgchain for subtask 2 shows it only took 5s (more accurate than compensated 15s)
	msgchains := []database.Msgchain{
		makeMsgchain(1, database.MsgchainTypePrimaryAgent, 1, &subtaskID1, now, now.Add(10*time.Second)),
		makeMsgchain(2, database.MsgchainTypePrimaryAgent, 1, &subtaskID2, now.Add(10*time.Second), now.Add(15*time.Second)),
	}

	durations := CalculateSubtasksWithOverlapCompensation(subtasks, msgchains)

	// Subtask 1: 10s
	if durations[1] < 9 || durations[1] > 11 {
		t.Errorf("Expected subtask 1 duration ~10s, got %f", durations[1])
	}

	// Subtask 2: compensated would be 15s, but msgchain shows 5s → use 5s
	if durations[2] < 4 || durations[2] > 6 {
		t.Errorf("Expected subtask 2 duration ~5s (msgchain), got %f", durations[2])
	}

	// Total should be 15s (with msgchain validation)
	total := durations[1] + durations[2]
	expected := 15.0
	if total < expected-1 || total > expected+1 {
		t.Errorf("Expected total ~%f seconds, got %f", expected, total)
	}
}

// ========== Integration Test ==========

func TestBuildFlowExecutionStats_CompleteFlow(t *testing.T) {
	now := time.Now()

	flowID := int64(1)
	flowTitle := "Test Flow"

	tasks := []database.GetTasksForFlowRow{
		{ID: 1,
			Title:     "Test Task",
			CreatedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(10 * time.Second), Valid: true},
		},
	}

	subtaskID := int64(1)
	subtasks := []database.GetSubtasksForTasksRow{
		{ID: 1,
			TaskID:    1,
			Title:     "Test Subtask",
			Status:    database.SubtaskStatusFinished,
			CreatedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(10 * time.Second), Valid: true},
		},
	}

	msgchains := []database.GetMsgchainsForFlowRow{
		{
			ID:              1,
			Type:            database.MsgchainTypePrimaryAgent,
			FlowID:          flowID,
			TaskID:          sql.NullInt64{Int64: 1, Valid: true},
			SubtaskID:       sql.NullInt64{Int64: subtaskID, Valid: true},
			DurationSeconds: 10.0,
			CreatedAt:       sql.NullTime{Time: now, Valid: true},
			UpdatedAt:       sql.NullTime{Time: now.Add(10 * time.Second), Valid: true},
		},
	}

	toolcalls := []database.GetToolcallsForFlowRow{
		{
			ID:              1,
			Status:          database.ToolcallStatusFinished,
			FlowID:          flowID,
			TaskID:          sql.NullInt64{Valid: false},
			SubtaskID:       sql.NullInt64{Int64: subtaskID, Valid: true},
			DurationSeconds: 5.0,
			CreatedAt:       sql.NullTime{Time: now, Valid: true},
			UpdatedAt:       sql.NullTime{Time: now.Add(5 * time.Second), Valid: true},
		},
	}

	assistantsCount := 2
	stats := BuildFlowExecutionStats(flowID, flowTitle, tasks, subtasks, msgchains, toolcalls, assistantsCount)

	if stats.FlowID != 1 {
		t.Errorf("Expected flow ID 1, got %d", stats.FlowID)
	}

	if stats.TotalDurationSeconds < 9 || stats.TotalDurationSeconds > 11 {
		t.Errorf("Expected ~10 seconds, got %f", stats.TotalDurationSeconds)
	}

	if stats.TotalToolcallsCount != 1 {
		t.Errorf("Expected 1 toolcall, got %d", stats.TotalToolcallsCount)
	}

	if stats.TotalAssistantsCount != 2 {
		t.Errorf("Expected 2 assistants, got %d", stats.TotalAssistantsCount)
	}

	if len(stats.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(stats.Tasks))
	}

	if len(stats.Tasks[0].Subtasks) != 1 {
		t.Fatalf("Expected 1 subtask, got %d", len(stats.Tasks[0].Subtasks))
	}
}

func TestBuildFlowExecutionStats_MathematicalConsistency(t *testing.T) {
	now := time.Now()

	flowID := int64(1)
	flowTitle := "Test Flow"

	tasks := []database.GetTasksForFlowRow{
		{ID: 1, Title: "Test Task", CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now.Add(100 * time.Second), Valid: true}},
	}

	// Batch-created subtasks (all created at same time)
	subtasks := []database.GetSubtasksForTasksRow{
		{ID: 1, TaskID: 1, Title: "Subtask 1", Status: database.SubtaskStatusFinished, CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now.Add(10 * time.Second), Valid: true}},
		{ID: 2, TaskID: 1, Title: "Subtask 2", Status: database.SubtaskStatusFinished, CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now.Add(20 * time.Second), Valid: true}},
		{ID: 3, TaskID: 1, Title: "Subtask 3", Status: database.SubtaskStatusFinished, CreatedAt: sql.NullTime{Time: now, Valid: true}, UpdatedAt: sql.NullTime{Time: now.Add(30 * time.Second), Valid: true}},
	}

	// Generator runs for 5s before subtasks
	msgchains := []database.GetMsgchainsForFlowRow{
		{
			ID:              1,
			Type:            database.MsgchainTypeGenerator,
			FlowID:          flowID,
			TaskID:          sql.NullInt64{Int64: 1, Valid: true},
			SubtaskID:       sql.NullInt64{Valid: false},
			DurationSeconds: 5.0,
			CreatedAt:       sql.NullTime{Time: now.Add(-5 * time.Second), Valid: true},
			UpdatedAt:       sql.NullTime{Time: now, Valid: true},
		},
	}

	toolcalls := []database.GetToolcallsForFlowRow{}

	assistantsCount := 0
	stats := BuildFlowExecutionStats(flowID, flowTitle, tasks, subtasks, msgchains, toolcalls, assistantsCount)

	// Critical mathematical consistency check:
	// Sum of subtask durations should equal task subtasks part
	var subtasksSum float64
	for _, subtask := range stats.Tasks[0].Subtasks {
		subtasksSum += subtask.TotalDurationSeconds
	}

	// Task duration should be: subtasks (30s compensated) + generator (5s) = 35s
	expectedTaskDuration := 35.0
	if stats.Tasks[0].TotalDurationSeconds < expectedTaskDuration-1 || stats.Tasks[0].TotalDurationSeconds > expectedTaskDuration+1 {
		t.Errorf("Expected task duration ~%fs, got %f", expectedTaskDuration, stats.Tasks[0].TotalDurationSeconds)
	}

	// Subtasks sum should be 30s (compensated: 10 + 10 + 10)
	expectedSubtasksSum := 30.0
	if subtasksSum < expectedSubtasksSum-1 || subtasksSum > expectedSubtasksSum+1 {
		t.Errorf("Expected subtasks sum ~%fs, got %f", expectedSubtasksSum, subtasksSum)
	}

	// Task duration should be >= subtasks sum (includes generator/refiner)
	if stats.Tasks[0].TotalDurationSeconds < subtasksSum {
		t.Errorf("Task duration (%f) cannot be less than subtasks sum (%f)", stats.Tasks[0].TotalDurationSeconds, subtasksSum)
	}

	// Flow duration should equal task duration (no flow-level toolcalls in this test)
	if math.Abs(stats.TotalDurationSeconds-stats.Tasks[0].TotalDurationSeconds) > 0.1 {
		t.Errorf("Flow duration (%f) should equal task duration (%f) with no flow-level toolcalls", stats.TotalDurationSeconds, stats.Tasks[0].TotalDurationSeconds)
	}
}

func TestBuildFlowExecutionStats_MultipleTasksWithRefiner(t *testing.T) {
	now := time.Now()

	flowID := int64(1)
	flowTitle := "Complex Flow"

	tasks := []database.GetTasksForFlowRow{
		{
			ID:        1,
			Title:     "Task 1",
			CreatedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(50 * time.Second), Valid: true},
		},
		{
			ID:        2,
			Title:     "Task 2",
			CreatedAt: sql.NullTime{Time: now.Add(50 * time.Second), Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(100 * time.Second), Valid: true},
		},
	}

	subtasks := []database.GetSubtasksForTasksRow{
		{
			ID:        1,
			TaskID:    1,
			Title:     "T1 S1",
			Status:    database.SubtaskStatusFinished,
			CreatedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(20 * time.Second), Valid: true},
		},
		{
			ID:        2,
			TaskID:    1,
			Title:     "T1 S2",
			Status:    database.SubtaskStatusFinished,
			CreatedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(40 * time.Second), Valid: true},
		},
		{
			ID:        3,
			TaskID:    2,
			Title:     "T2 S1",
			Status:    database.SubtaskStatusFinished,
			CreatedAt: sql.NullTime{Time: now.Add(50 * time.Second), Valid: true},
			UpdatedAt: sql.NullTime{Time: now.Add(100 * time.Second), Valid: true},
		},
	}

	msgchains := []database.GetMsgchainsForFlowRow{
		// Task 1: generator (5s) + refiner between subtasks (5s)
		{
			ID:              1,
			Type:            database.MsgchainTypeGenerator,
			FlowID:          flowID,
			TaskID:          sql.NullInt64{Int64: 1, Valid: true},
			SubtaskID:       sql.NullInt64{Valid: false},
			DurationSeconds: 5.0,
			CreatedAt:       sql.NullTime{Time: now.Add(-5 * time.Second), Valid: true},
			UpdatedAt:       sql.NullTime{Time: now, Valid: true},
		},
		{
			ID:              2,
			Type:            database.MsgchainTypeRefiner,
			FlowID:          flowID,
			TaskID:          sql.NullInt64{Int64: 1, Valid: true},
			SubtaskID:       sql.NullInt64{Valid: false},
			DurationSeconds: 5.0,
			CreatedAt:       sql.NullTime{Time: now.Add(20 * time.Second), Valid: true},
			UpdatedAt:       sql.NullTime{Time: now.Add(25 * time.Second), Valid: true},
		},
	}

	toolcalls := []database.GetToolcallsForFlowRow{}

	assistantsCount := 1
	stats := BuildFlowExecutionStats(flowID, flowTitle, tasks, subtasks, msgchains, toolcalls, assistantsCount)

	// Verify we have 2 tasks
	if len(stats.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(stats.Tasks))
	}

	// Task 1: subtasks (20 + 20 compensated = 40s) + generator (5s) + refiner (5s) = 50s
	task1Duration := stats.Tasks[0].TotalDurationSeconds
	expectedTask1 := 50.0
	if task1Duration < expectedTask1-1 || task1Duration > expectedTask1+1 {
		t.Errorf("Expected task 1 duration ~%fs, got %f", expectedTask1, task1Duration)
	}

	// Task 1 subtasks sum should be 40s (compensated)
	var task1SubtasksSum float64
	for _, subtask := range stats.Tasks[0].Subtasks {
		task1SubtasksSum += subtask.TotalDurationSeconds
	}

	expectedSubtasks1 := 40.0
	if task1SubtasksSum < expectedSubtasks1-1 || task1SubtasksSum > expectedSubtasks1+1 {
		t.Errorf("Expected task 1 subtasks sum ~%fs, got %f", expectedSubtasks1, task1SubtasksSum)
	}

	// Task 1 duration should be: subtasks + generator + refiner
	if task1Duration < task1SubtasksSum {
		t.Errorf("Task 1 duration (%f) should be >= subtasks sum (%f)", task1Duration, task1SubtasksSum)
	}
}

// ========== Integration Test ==========
