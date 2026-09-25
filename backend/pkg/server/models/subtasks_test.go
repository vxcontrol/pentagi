package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestSubtasks_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	modelsCheckValid(t, []modelsValidCase{
		{"a created subtask status", SubtaskStatusCreated, ""},
		{"a running subtask status", SubtaskStatusRunning, ""},
		{"a waiting subtask status", SubtaskStatusWaiting, ""},
		{"a finished subtask status", SubtaskStatusFinished, ""},
		{"a failed subtask status", SubtaskStatusFailed, ""},
		{"an empty subtask status", SubtaskStatus(""), "invalid SubtaskStatus: "},
		{"an unknown subtask status", SubtaskStatus("pending"), "invalid SubtaskStatus: pending"},

		{"a complete subtask", modelsSubtask(), ""},
		{"a subtask without a title", modelsWith(modelsSubtask(), func(s *Subtask) { s.Title = "" }), "Subtask.Title:required"},
		{"a subtask without a description", modelsWith(modelsSubtask(), func(s *Subtask) { s.Description = "" }),
			"Subtask.Description:required"},
		{"a subtask in an unknown status", modelsWith(modelsSubtask(), func(s *Subtask) { s.Status = "invalid" }),
			"Subtask.Status:valid"},
	})
}

func TestSubtasks_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "subtasks", (&Subtask{}).TableName())
}
