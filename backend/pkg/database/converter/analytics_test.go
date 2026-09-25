package converter

import (
	"database/sql"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"

	"github.com/stretchr/testify/assert"
)

// analyticsStart is a fixed clock, so every duration below is exact.
var analyticsStart = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func analyticsAt(second int) sql.NullTime {
	return sql.NullTime{Time: analyticsStart.Add(time.Duration(second) * time.Second), Valid: true}
}

// analyticsID turns an id into a nullable column; zero is NULL.
func analyticsID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id != 0}
}

func analyticsSubtask(id int64, status database.SubtaskStatus, from, to int) database.Subtask {
	return database.Subtask{ID: id, TaskID: 1, Title: "Test Subtask", Status: status, CreatedAt: analyticsAt(from), UpdatedAt: analyticsAt(to)}
}

func analyticsMsgchain(msgType database.MsgchainType, taskID, subtaskID int64, from, to int) database.Msgchain {
	return database.Msgchain{
		Type:            msgType,
		TaskID:          analyticsID(taskID),
		SubtaskID:       analyticsID(subtaskID),
		CreatedAt:       analyticsAt(from),
		UpdatedAt:       analyticsAt(to),
		DurationSeconds: float64(to - from),
	}
}

func analyticsToolcall(id int64, status database.ToolcallStatus, taskID, subtaskID int64) database.Toolcall {
	return database.Toolcall{ID: id, Status: status, TaskID: analyticsID(taskID), SubtaskID: analyticsID(subtaskID)}
}

func TestAnalytics_CalculateSubtaskDuration_CountsOnlyAStartedSubtask(t *testing.T) {
	for _, tt := range []struct {
		name      string
		subtask   database.Subtask
		msgchains []database.Msgchain
		want      float64
	}{
		{"a created subtask has not started", analyticsSubtask(1, database.SubtaskStatusCreated, 0, 10), nil, 0},
		{"a waiting subtask has not started", analyticsSubtask(1, database.SubtaskStatusWaiting, 0, 10), nil, 0},
		{"a finished subtask counts from creation to its last update",
			analyticsSubtask(1, database.SubtaskStatusFinished, 0, 100), nil, 100},
		{"a shorter primary agent chain is the more conservative duration",
			analyticsSubtask(1, database.SubtaskStatusFinished, 0, 100),
			[]database.Msgchain{analyticsMsgchain(database.MsgchainTypePrimaryAgent, 1, 1, 0, 50)}, 50},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, CalculateSubtaskDuration(tt.subtask, tt.msgchains), 1e-9)
		})
	}
}

func TestAnalytics_CalculateSubtasksWithOverlapCompensation_StartsEachSubtaskWhenThePreviousEnded(t *testing.T) {
	finished, created, waiting := database.SubtaskStatusFinished, database.SubtaskStatusCreated, database.SubtaskStatusWaiting

	for _, tt := range []struct {
		name      string
		subtasks  []database.Subtask
		msgchains []database.Msgchain
		want      map[int64]float64
	}{
		{
			name:     "subtasks that ran back to back",
			subtasks: []database.Subtask{analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 10, 20)},
			want:     map[int64]float64{1: 10, 2: 10},
		},
		{
			name:     "a subtask created with its predecessor starts when that one ends",
			subtasks: []database.Subtask{analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 0, 20)},
			want:     map[int64]float64{1: 10, 2: 10},
		},
		{
			name: "five subtasks created in one batch",
			subtasks: []database.Subtask{
				analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 0, 20), analyticsSubtask(3, finished, 0, 30),
				analyticsSubtask(4, finished, 0, 40), analyticsSubtask(5, finished, 0, 50),
			},
			want: map[int64]float64{1: 10, 2: 10, 3: 10, 4: 10, 5: 10},
		},
		{
			name: "a created subtask counts nothing and moves no start",
			subtasks: []database.Subtask{
				analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, created, 0, 100), analyticsSubtask(3, finished, 10, 20),
			},
			want: map[int64]float64{1: 10, 2: 0, 3: 10},
		},
		{
			name: "created and waiting subtasks count nothing",
			subtasks: []database.Subtask{
				analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, created, 0, 100),
				analyticsSubtask(3, waiting, 0, 100), analyticsSubtask(4, finished, 0, 25),
			},
			want: map[int64]float64{1: 10, 2: 0, 3: 0, 4: 15},
		},
		{
			name:     "a shorter primary agent chain caps the compensated duration",
			subtasks: []database.Subtask{analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 0, 25)},
			msgchains: []database.Msgchain{
				analyticsMsgchain(database.MsgchainTypePrimaryAgent, 1, 1, 0, 10),
				analyticsMsgchain(database.MsgchainTypePrimaryAgent, 1, 2, 10, 15),
			},
			want: map[int64]float64{1: 10, 2: 5},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDeltaMapValues(t, tt.want, CalculateSubtasksWithOverlapCompensation(tt.subtasks, tt.msgchains), 1e-9)
		})
	}
}

func TestAnalytics_CalculateTaskDuration_AddsThePlanningAgentsToTheSubtasks(t *testing.T) {
	finished := database.SubtaskStatusFinished

	for _, tt := range []struct {
		name      string
		subtasks  []database.Subtask
		msgchains []database.Msgchain
		want      float64
	}{
		{"only subtasks", []database.Subtask{analyticsSubtask(1, finished, 0, 10)}, nil, 10},
		{"the generator runs before the subtasks", []database.Subtask{analyticsSubtask(1, finished, 5, 15)},
			[]database.Msgchain{analyticsMsgchain(database.MsgchainTypeGenerator, 1, 0, 0, 5)}, 15},
		{"the refiner runs between subtasks",
			[]database.Subtask{analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 13, 23)},
			[]database.Msgchain{analyticsMsgchain(database.MsgchainTypeRefiner, 1, 0, 10, 13)}, 23},
		{"the reporter runs after the subtasks", []database.Subtask{analyticsSubtask(1, finished, 0, 10)},
			[]database.Msgchain{analyticsMsgchain(database.MsgchainTypeReporter, 1, 0, 10, 14)}, 14},
		{"batch-created subtasks count once each",
			[]database.Subtask{analyticsSubtask(1, finished, 0, 10), analyticsSubtask(2, finished, 0, 20), analyticsSubtask(3, finished, 0, 30)},
			nil, 30},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, CalculateTaskDuration(database.Task{ID: 1}, tt.subtasks, tt.msgchains), 1e-9)
		})
	}
}

func TestAnalytics_CalculateFlowDuration_AddsOnlyFlowLevelAssistantChains(t *testing.T) {
	oneTask := []database.Task{{ID: 1}}
	itsSubtask := map[int64][]database.Subtask{1: {analyticsSubtask(1, database.SubtaskStatusFinished, 0, 10)}}
	noChains := map[int64][]database.Msgchain{1: {}}

	for _, tt := range []struct {
		name       string
		tasks      []database.Task
		subtasks   map[int64][]database.Subtask
		msgchains  map[int64][]database.Msgchain
		assistants []database.Msgchain
		want       float64
	}{
		{"tasks only", oneTask, itsSubtask, noChains, nil, 10},
		{"a flow-level assistant chain adds its time", oneTask, itsSubtask, noChains,
			[]database.Msgchain{analyticsMsgchain(database.MsgchainTypeAssistant, 0, 0, 20, 25)}, 15},
		{"an assistant chain bound to a task or a subtask is not flow-level", nil, nil, nil,
			[]database.Msgchain{
				analyticsMsgchain(database.MsgchainTypeAssistant, 1, 0, 0, 10),
				analyticsMsgchain(database.MsgchainTypeAssistant, 0, 1, 0, 10),
				analyticsMsgchain(database.MsgchainTypeAssistant, 0, 0, 0, 5),
			}, 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, CalculateFlowDuration(tt.tasks, tt.subtasks, tt.msgchains, tt.assistants), 1e-9)
		})
	}
}

// Subtests are keyed by unit and count the same calls.
func TestAnalytics_CountFinishedToolcalls_CountsOnlyFinishedAndFailedCalls(t *testing.T) {
	calls := []database.Toolcall{
		analyticsToolcall(1, database.ToolcallStatusFinished, 1, 0),
		analyticsToolcall(2, database.ToolcallStatusFailed, 1, 0),
		analyticsToolcall(3, database.ToolcallStatusReceived, 1, 0),
		analyticsToolcall(4, database.ToolcallStatusFinished, 0, 1),
		analyticsToolcall(5, database.ToolcallStatusFailed, 0, 1),
		analyticsToolcall(6, database.ToolcallStatusRunning, 0, 1),
		analyticsToolcall(7, database.ToolcallStatusFinished, 0, 2),
		analyticsToolcall(8, database.ToolcallStatusFinished, 2, 0),
	}

	t.Run("across the flow", func(t *testing.T) {
		assert.Equal(t, 6, CountFinishedToolcalls(calls))
	})
	t.Run("for one subtask", func(t *testing.T) {
		assert.Equal(t, 2, CountFinishedToolcallsForSubtask(calls, 1))
	})
	t.Run("for a task and its subtasks", func(t *testing.T) {
		assert.Equal(t, 4, CountFinishedToolcallsForTask(calls, 1, []int64{1}))
	})
}

func TestAnalytics_BuildFlowExecutionStats_BuildsTheWholeTree(t *testing.T) {
	finished := database.SubtaskStatusFinished
	subtaskRow := func(id, taskID int64, title string, from, to int) database.GetSubtasksForTasksRow {
		return database.GetSubtasksForTasksRow{ID: id, TaskID: taskID, Title: title, Status: finished,
			CreatedAt: analyticsAt(from), UpdatedAt: analyticsAt(to)}
	}
	chainRow := func(msgType database.MsgchainType, taskID, subtaskID int64, from, to int) database.GetMsgchainsForFlowRow {
		return database.GetMsgchainsForFlowRow{Type: msgType, FlowID: 1, TaskID: analyticsID(taskID), SubtaskID: analyticsID(subtaskID),
			DurationSeconds: float64(to - from), CreatedAt: analyticsAt(from), UpdatedAt: analyticsAt(to)}
	}

	for _, tt := range []struct {
		name       string
		tasks      []database.GetTasksForFlowRow
		subtasks   []database.GetSubtasksForTasksRow
		msgchains  []database.GetMsgchainsForFlowRow
		toolcalls  []database.GetToolcallsForFlowRow
		assistants int
		want       *model.FlowExecutionStats
	}{
		{
			name:      "a task whose subtask made a tool call",
			tasks:     []database.GetTasksForFlowRow{{ID: 1, Title: "Test Task", CreatedAt: analyticsAt(0), UpdatedAt: analyticsAt(10)}},
			subtasks:  []database.GetSubtasksForTasksRow{subtaskRow(1, 1, "Test Subtask", 0, 10)},
			msgchains: []database.GetMsgchainsForFlowRow{chainRow(database.MsgchainTypePrimaryAgent, 1, 1, 0, 10)},
			toolcalls: []database.GetToolcallsForFlowRow{{ID: 1, Status: database.ToolcallStatusFinished, FlowID: 1,
				SubtaskID: analyticsID(1), DurationSeconds: 5, CreatedAt: analyticsAt(0), UpdatedAt: analyticsAt(5)}},
			assistants: 2,
			want: &model.FlowExecutionStats{FlowID: 1, FlowTitle: "Test Flow", TotalDurationSeconds: 10, TotalToolcallsCount: 1,
				TotalAssistantsCount: 2, Tasks: []*model.TaskExecutionStats{{
					TaskID: 1, TaskTitle: "Test Task", TotalDurationSeconds: 10, TotalToolcallsCount: 1,
					Subtasks: []*model.SubtaskExecutionStats{{SubtaskID: 1, SubtaskTitle: "Test Subtask", TotalDurationSeconds: 10, TotalToolcallsCount: 1}},
				}}},
		},
		{
			name:  "batch-created subtasks after a generator",
			tasks: []database.GetTasksForFlowRow{{ID: 1, Title: "Test Task", CreatedAt: analyticsAt(0), UpdatedAt: analyticsAt(100)}},
			subtasks: []database.GetSubtasksForTasksRow{
				subtaskRow(1, 1, "Subtask 1", 0, 10), subtaskRow(2, 1, "Subtask 2", 0, 20), subtaskRow(3, 1, "Subtask 3", 0, 30),
			},
			msgchains: []database.GetMsgchainsForFlowRow{chainRow(database.MsgchainTypeGenerator, 1, 0, -5, 0)},
			want: &model.FlowExecutionStats{FlowID: 1, FlowTitle: "Test Flow", TotalDurationSeconds: 35, Tasks: []*model.TaskExecutionStats{{
				TaskID: 1, TaskTitle: "Test Task", TotalDurationSeconds: 35,
				Subtasks: []*model.SubtaskExecutionStats{
					{SubtaskID: 1, SubtaskTitle: "Subtask 1", TotalDurationSeconds: 10},
					{SubtaskID: 2, SubtaskTitle: "Subtask 2", TotalDurationSeconds: 10},
					{SubtaskID: 3, SubtaskTitle: "Subtask 3", TotalDurationSeconds: 10},
				},
			}}},
		},
		{
			name: "two tasks, one with a generator and a refiner",
			tasks: []database.GetTasksForFlowRow{
				{ID: 1, Title: "Task 1", CreatedAt: analyticsAt(0), UpdatedAt: analyticsAt(50)},
				{ID: 2, Title: "Task 2", CreatedAt: analyticsAt(50), UpdatedAt: analyticsAt(100)},
			},
			subtasks: []database.GetSubtasksForTasksRow{
				subtaskRow(1, 1, "T1 S1", 0, 20), subtaskRow(2, 1, "T1 S2", 0, 40), subtaskRow(3, 2, "T2 S1", 50, 100),
			},
			msgchains: []database.GetMsgchainsForFlowRow{
				chainRow(database.MsgchainTypeGenerator, 1, 0, -5, 0),
				chainRow(database.MsgchainTypeRefiner, 1, 0, 20, 25),
			},
			assistants: 1,
			want: &model.FlowExecutionStats{FlowID: 1, FlowTitle: "Test Flow", TotalDurationSeconds: 100, TotalAssistantsCount: 1,
				Tasks: []*model.TaskExecutionStats{
					{TaskID: 1, TaskTitle: "Task 1", TotalDurationSeconds: 50, Subtasks: []*model.SubtaskExecutionStats{
						{SubtaskID: 1, SubtaskTitle: "T1 S1", TotalDurationSeconds: 20},
						{SubtaskID: 2, SubtaskTitle: "T1 S2", TotalDurationSeconds: 20},
					}},
					{TaskID: 2, TaskTitle: "Task 2", TotalDurationSeconds: 50, Subtasks: []*model.SubtaskExecutionStats{
						{SubtaskID: 3, SubtaskTitle: "T2 S1", TotalDurationSeconds: 50},
					}},
				}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFlowExecutionStats(1, "Test Flow", tt.tasks, tt.subtasks, tt.msgchains, tt.toolcalls, tt.assistants)
			assert.Equal(t, tt.want, got)
		})
	}
}
