package converter

import (
	"math"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"sort"
	"time"
)

// ========== Subtask Duration Calculation ==========

// CalculateSubtaskDuration calculates the actual execution duration of a subtask
// Uses linear time (created_at -> updated_at) with compensation for:
// - Subtasks in 'created' or 'waiting' status (returns 0)
// - Running subtasks (returns time from created_at to now)
// - Finished/Failed subtasks (returns time from created_at to updated_at)
// Optionally validates against primary_agent msgchain duration if available
func CalculateSubtaskDuration(subtask database.Subtask, msgchains []database.Msgchain) float64 {
	// Ignore subtasks that haven't started or are waiting
	if subtask.Status == database.SubtaskStatusCreated || subtask.Status == database.SubtaskStatusWaiting {
		return 0
	}

	// Calculate linear duration
	var linearDuration float64
	if subtask.Status == database.SubtaskStatusRunning {
		// For running subtasks: from created_at to now
		linearDuration = time.Since(subtask.CreatedAt.Time).Seconds()
	} else {
		// For finished/failed: from created_at to updated_at
		linearDuration = subtask.UpdatedAt.Time.Sub(subtask.CreatedAt.Time).Seconds()
	}

	// Try to find primary_agent msgchain for validation
	var msgchainDuration float64
	for _, mc := range msgchains {
		if mc.Type == database.MsgchainTypePrimaryAgent &&
			mc.SubtaskID.Valid && mc.SubtaskID.Int64 == subtask.ID {
			msgchainDuration += mc.DurationSeconds
		}
	}

	// If msgchain exists, use the minimum (more conservative estimate)
	if msgchainDuration > 0 {
		return math.Min(linearDuration, msgchainDuration)
	}

	return linearDuration
}

// SubtaskDurationInfo holds calculated duration info for a subtask
type SubtaskDurationInfo struct {
	SubtaskID int64
	Duration  float64
}

// CalculateSubtasksWithOverlapCompensation calculates duration for each subtask
// accounting for potential overlap in created_at timestamps when subtasks are created in batch
// Returns map of subtask_id -> compensated_duration
func CalculateSubtasksWithOverlapCompensation(subtasks []database.Subtask, msgchains []database.Msgchain) map[int64]float64 {
	result := make(map[int64]float64)

	if len(subtasks) == 0 {
		return result
	}

	// Sort subtasks by ID (which is monotonic and represents execution order)
	sorted := make([]database.Subtask, len(subtasks))
	copy(sorted, subtasks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	var previousEndTime time.Time

	for _, subtask := range sorted {
		// Skip subtasks that haven't started
		if subtask.Status == database.SubtaskStatusCreated || subtask.Status == database.SubtaskStatusWaiting {
			result[subtask.ID] = 0
			continue
		}

		// Determine actual start time (compensating for overlap)
		startTime := subtask.CreatedAt.Time
		if !previousEndTime.IsZero() && startTime.Before(previousEndTime) {
			// If current subtask was created before previous one finished,
			// use previous end time as start time
			startTime = previousEndTime
		}

		// Determine end time
		var endTime time.Time
		if subtask.Status == database.SubtaskStatusRunning {
			endTime = time.Now()
		} else {
			endTime = subtask.UpdatedAt.Time
		}

		// Calculate duration for this subtask
		duration := 0.0
		if endTime.After(startTime) {
			duration = endTime.Sub(startTime).Seconds()

			// Validate against sum of all primary_agent msgchains for this subtask
			var msgchainDuration float64
			for _, mc := range msgchains {
				if mc.Type == database.MsgchainTypePrimaryAgent &&
					mc.SubtaskID.Valid && mc.SubtaskID.Int64 == subtask.ID {
					msgchainDuration += mc.DurationSeconds
				}
			}

			if msgchainDuration > 0 {
				duration = math.Min(duration, msgchainDuration)
			}
		}

		result[subtask.ID] = duration
		previousEndTime = endTime
	}

	return result
}

// ========== Task Duration Calculation ==========

// CalculateTaskDuration calculates total task execution time including:
// 1. Generator agent execution (before subtasks)
// 2. All subtasks execution (with overlap compensation)
// 3. Refiner agent executions (between subtasks)
// 4. Task reporter agent execution (after subtasks)
func CalculateTaskDuration(task database.Task, subtasks []database.Subtask, msgchains []database.Msgchain) float64 {
	// 1. Calculate subtasks duration with overlap compensation
	subtaskDurations := CalculateSubtasksWithOverlapCompensation(subtasks, msgchains)
	var subtasksDuration float64
	for _, duration := range subtaskDurations {
		subtasksDuration += duration
	}

	// 2. Calculate generator agent duration (runs before subtasks)
	generatorDuration := getMsgchainDuration(msgchains, database.MsgchainTypeGenerator, task.ID, nil)

	// 3. Calculate total refiner agent duration (runs between subtasks)
	refinerDuration := sumMsgchainsDuration(msgchains, database.MsgchainTypeRefiner, task.ID)

	// 4. Calculate task reporter agent duration (runs after subtasks)
	reporterDuration := getMsgchainDuration(msgchains, database.MsgchainTypeReporter, task.ID, nil)

	return subtasksDuration + generatorDuration + refinerDuration + reporterDuration
}

// getMsgchainDuration returns duration of a single msgchain matching criteria
func getMsgchainDuration(msgchains []database.Msgchain, msgType database.MsgchainType, taskID int64, subtaskID *int64) float64 {
	for _, mc := range msgchains {
		if mc.Type == msgType && mc.TaskID.Valid && mc.TaskID.Int64 == taskID {
			// Check subtaskID match if specified
			if subtaskID != nil {
				if !mc.SubtaskID.Valid || mc.SubtaskID.Int64 != *subtaskID {
					continue
				}
			} else {
				// If subtaskID is nil, we want msgchains without subtask_id
				if mc.SubtaskID.Valid {
					continue
				}
			}
			return mc.DurationSeconds
		}
	}
	return 0
}

// sumMsgchainsDuration returns sum of durations for all msgchains matching criteria
func sumMsgchainsDuration(msgchains []database.Msgchain, msgType database.MsgchainType, taskID int64) float64 {
	var total float64
	for _, mc := range msgchains {
		if mc.Type == msgType && mc.TaskID.Valid && mc.TaskID.Int64 == taskID && !mc.SubtaskID.Valid {
			total += mc.DurationSeconds
		}
	}
	return total
}

// ========== Flow Duration Calculation ==========

// CalculateFlowDuration calculates total flow execution time including:
// 1. All tasks duration (which includes generator, subtasks, and refiner)
// 2. Assistant msgchains duration (flow-level, without task binding)
func CalculateFlowDuration(tasks []database.Task, subtasksMap map[int64][]database.Subtask,
	msgchainsMap map[int64][]database.Msgchain, assistantMsgchains []database.Msgchain) float64 {

	// 1. Calculate total tasks duration
	var tasksDuration float64
	for _, task := range tasks {
		subtasks := subtasksMap[task.ID]
		msgchains := msgchainsMap[task.ID]
		tasksDuration += CalculateTaskDuration(task, subtasks, msgchains)
	}

	// 2. Calculate assistant msgchains duration (flow-level operations without task binding)
	var assistantDuration float64
	for _, mc := range assistantMsgchains {
		if mc.Type == database.MsgchainTypeAssistant && !mc.TaskID.Valid && !mc.SubtaskID.Valid {
			assistantDuration += mc.DurationSeconds
		}
	}

	return tasksDuration + assistantDuration
}

// ========== Toolcalls Count Calculation ==========

// CountFinishedToolcalls counts only finished and failed toolcalls (excludes created/running)
func CountFinishedToolcalls(toolcalls []database.Toolcall) int {
	count := 0
	for _, tc := range toolcalls {
		if tc.Status == database.ToolcallStatusFinished || tc.Status == database.ToolcallStatusFailed {
			count++
		}
	}
	return count
}

// CountFinishedToolcallsForSubtask counts finished toolcalls for a specific subtask
func CountFinishedToolcallsForSubtask(toolcalls []database.Toolcall, subtaskID int64) int {
	count := 0
	for _, tc := range toolcalls {
		if tc.SubtaskID.Valid && tc.SubtaskID.Int64 == subtaskID {
			if tc.Status == database.ToolcallStatusFinished || tc.Status == database.ToolcallStatusFailed {
				count++
			}
		}
	}
	return count
}

// CountFinishedToolcallsForTask counts finished toolcalls for a task (including subtasks)
func CountFinishedToolcallsForTask(toolcalls []database.Toolcall, taskID int64, subtaskIDs []int64) int {
	subtaskIDSet := make(map[int64]bool)
	for _, id := range subtaskIDs {
		subtaskIDSet[id] = true
	}

	count := 0
	for _, tc := range toolcalls {
		// Count task-level toolcalls
		if tc.TaskID.Valid && tc.TaskID.Int64 == taskID && !tc.SubtaskID.Valid {
			if tc.Status == database.ToolcallStatusFinished || tc.Status == database.ToolcallStatusFailed {
				count++
			}
		}
		// Count subtask-level toolcalls
		if tc.SubtaskID.Valid && subtaskIDSet[tc.SubtaskID.Int64] {
			if tc.Status == database.ToolcallStatusFinished || tc.Status == database.ToolcallStatusFailed {
				count++
			}
		}
	}
	return count
}

// ========== Hierarchical Stats Building ==========

// BuildFlowExecutionStats builds hierarchical execution statistics for a flow
func BuildFlowExecutionStats(flowID int64, flowTitle string, tasks []database.GetTasksForFlowRow,
	subtasks []database.GetSubtasksForTasksRow, msgchains []database.GetMsgchainsForFlowRow,
	toolcalls []database.GetToolcallsForFlowRow, assistantsCount int) *model.FlowExecutionStats {

	// Convert row types to internal structures
	subtasksMap := make(map[int64][]database.Subtask)
	for _, s := range subtasks {
		subtasksMap[s.TaskID] = append(subtasksMap[s.TaskID], database.Subtask{
			ID:        s.ID,
			TaskID:    s.TaskID,
			Title:     s.Title,
			Status:    s.Status,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	msgchainsMap := make(map[int64][]database.Msgchain)
	assistantMsgchains := make([]database.Msgchain, 0)
	for _, mc := range msgchains {
		msgchain := database.Msgchain{
			ID:              mc.ID,
			Type:            mc.Type,
			FlowID:          mc.FlowID,
			TaskID:          mc.TaskID,
			SubtaskID:       mc.SubtaskID,
			DurationSeconds: mc.DurationSeconds,
			CreatedAt:       mc.CreatedAt,
			UpdatedAt:       mc.UpdatedAt,
		}

		if mc.TaskID.Valid {
			msgchainsMap[mc.TaskID.Int64] = append(msgchainsMap[mc.TaskID.Int64], msgchain)
		} else if mc.Type == database.MsgchainTypeAssistant {
			// Collect flow-level assistant msgchains
			assistantMsgchains = append(assistantMsgchains, msgchain)
		}
	}

	toolcallsMap := make(map[int64][]database.Toolcall)
	flowToolcalls := make([]database.Toolcall, 0, len(toolcalls))
	for _, tc := range toolcalls {
		toolcall := database.Toolcall{
			ID:              tc.ID,
			Status:          tc.Status,
			FlowID:          tc.FlowID,
			TaskID:          tc.TaskID,
			SubtaskID:       tc.SubtaskID,
			DurationSeconds: tc.DurationSeconds,
			CreatedAt:       tc.CreatedAt,
			UpdatedAt:       tc.UpdatedAt,
		}
		if tc.FlowID == flowID {
			flowToolcalls = append(flowToolcalls, toolcall)
		}

		if tc.TaskID.Valid {
			toolcallsMap[tc.TaskID.Int64] = append(toolcallsMap[tc.TaskID.Int64], toolcall)
		} else if tc.SubtaskID.Valid {
			// Find task for this subtask
			for taskID, subs := range subtasksMap {
				for _, sub := range subs {
					if sub.ID == tc.SubtaskID.Int64 {
						toolcallsMap[taskID] = append(toolcallsMap[taskID], toolcall)
						break
					}
				}
			}
		}
	}

	// Build task stats
	taskStats := make([]*model.TaskExecutionStats, 0, len(tasks))

	for _, taskRow := range tasks {
		task := database.Task{
			ID:        taskRow.ID,
			Title:     taskRow.Title,
			CreatedAt: taskRow.CreatedAt,
			UpdatedAt: taskRow.UpdatedAt,
		}

		subs := subtasksMap[task.ID]
		mcs := msgchainsMap[task.ID]
		tcs := toolcallsMap[task.ID]

		// Calculate compensated durations for all subtasks at once
		compensatedDurations := CalculateSubtasksWithOverlapCompensation(subs, mcs)

		// Build subtask stats using compensated durations
		subtaskStats := make([]*model.SubtaskExecutionStats, 0, len(subs))
		subtaskIDs := make([]int64, 0, len(subs))

		for _, subtask := range subs {
			subtaskIDs = append(subtaskIDs, subtask.ID)
			duration := compensatedDurations[subtask.ID]
			count := CountFinishedToolcallsForSubtask(tcs, subtask.ID)

			subtaskStats = append(subtaskStats, &model.SubtaskExecutionStats{
				SubtaskID:            subtask.ID,
				SubtaskTitle:         subtask.Title,
				TotalDurationSeconds: duration,
				TotalToolcallsCount:  count,
			})
		}

		// Build task stats
		taskDuration := CalculateTaskDuration(task, subs, mcs)
		taskCount := CountFinishedToolcallsForTask(tcs, task.ID, subtaskIDs)

		taskStats = append(taskStats, &model.TaskExecutionStats{
			TaskID:               task.ID,
			TaskTitle:            task.Title,
			TotalDurationSeconds: taskDuration,
			TotalToolcallsCount:  taskCount,
			Subtasks:             subtaskStats,
		})
	}

	// Build flow stats
	tasksInternal := make([]database.Task, len(tasks))
	for i, t := range tasks {
		tasksInternal[i] = database.Task{
			ID:        t.ID,
			Title:     t.Title,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		}
	}

	flowDuration := CalculateFlowDuration(tasksInternal, subtasksMap, msgchainsMap, assistantMsgchains)
	flowCount := CountFinishedToolcalls(flowToolcalls)

	return &model.FlowExecutionStats{
		FlowID:               flowID,
		FlowTitle:            flowTitle,
		TotalDurationSeconds: flowDuration,
		TotalToolcallsCount:  flowCount,
		TotalAssistantsCount: assistantsCount,
		Tasks:                taskStats,
	}
}
