package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestTasks_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	subtasks := []Subtask{modelsSubtask()}

	modelsCheckValid(t, []modelsValidCase{
		{"a created task status", TaskStatusCreated, ""},
		{"a running task status", TaskStatusRunning, ""},
		{"a waiting task status", TaskStatusWaiting, ""},
		{"a finished task status", TaskStatusFinished, ""},
		{"a failed task status", TaskStatusFailed, ""},
		{"an empty task status", TaskStatus(""), "invalid TaskStatus: "},
		{"an unknown task status", TaskStatus("cancelled"), "invalid TaskStatus: cancelled"},

		{"a complete task", modelsTask(), ""},
		{"a task without a title", modelsWith(modelsTask(), func(tk *Task) { tk.Title = "" }), "Task.Title:required"},
		{"a task without input", modelsWith(modelsTask(), func(tk *Task) { tk.Input = "" }), "Task.Input:required"},
		{"a task in an unknown status", modelsWith(modelsTask(), func(tk *Task) { tk.Status = "invalid" }), "Task.Status:valid"},

		{"a task with its subtasks", TaskSubtasks{Subtasks: subtasks, Task: modelsTask()}, ""},
		{"an invalid task with its subtasks", TaskSubtasks{Subtasks: subtasks,
			Task: modelsWith(modelsTask(), func(tk *Task) { tk.Title = "" })}, "Task.Title:required"},
		{"a task with an invalid subtask", TaskSubtasks{
			Subtasks: []Subtask{modelsWith(modelsSubtask(), func(s *Subtask) { s.Title = "" })},
			Task:     modelsTask(),
		}, "Subtask.Title:required"},
	})
}

func TestTasks_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "tasks", (&Task{}).TableName())
	assert.Equal(t, "tasks", (&TaskSubtasks{}).TableName())
}
