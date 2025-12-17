package providers

import (
	"io"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logrus.NewEntry(logger)
}

func TestApplySubtaskOperations_EmptyPatch(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{},
		Message:    "No changes needed",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "Task 1", result[0].Title)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, int64(3), result[2].ID)
}

func TestApplySubtaskOperations_RemoveOperation(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
	}

	id2 := int64(2)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove, ID: &id2},
		},
		Message: "Removed task 2",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(3), result[1].ID)
}

func TestApplySubtaskOperations_RemoveMultiple(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
		{ID: 4, Title: "Task 4", Description: "Description 4"},
	}

	id1, id3 := int64(1), int64(3)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove, ID: &id1},
			{Op: tools.SubtaskOpRemove, ID: &id3},
		},
		Message: "Removed tasks 1 and 3",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), result[0].ID)
	assert.Equal(t, int64(4), result[1].ID)
}

func TestApplySubtaskOperations_RemoveNonExistent(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id99 := int64(99)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove, ID: &id99},
		},
		Message: "Try to remove non-existent task",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found for removal")
}

func TestApplySubtaskOperations_ModifyTitle(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
	}

	id1 := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, ID: &id1, Title: "Updated Task 1"},
		},
		Message: "Updated title",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "Updated Task 1", result[0].Title)
	assert.Equal(t, "Description 1", result[0].Description) // Description unchanged
	assert.Equal(t, "Task 2", result[1].Title)              // Other task unchanged
}

func TestApplySubtaskOperations_ModifyDescription(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id1 := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, ID: &id1, Description: "New Description"},
		},
		Message: "Updated description",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "Task 1", result[0].Title)              // Title unchanged
	assert.Equal(t, "New Description", result[0].Description)
}

func TestApplySubtaskOperations_ModifyBoth(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id1 := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, ID: &id1, Title: "New Title", Description: "New Description"},
		},
		Message: "Updated both",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "New Title", result[0].Title)
	assert.Equal(t, "New Description", result[0].Description)
}

func TestApplySubtaskOperations_AddAtBeginning(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, Title: "New Task", Description: "New Description"},
		},
		Message: "Added at beginning",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, int64(0), result[0].ID) // New task has ID 0
	assert.Equal(t, "New Task", result[0].Title)
	assert.Equal(t, int64(1), result[1].ID)
	assert.Equal(t, int64(2), result[2].ID)
}

func TestApplySubtaskOperations_AddAfterSpecific(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
	}

	afterID := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, AfterID: &afterID, Title: "New Task", Description: "New Description"},
		},
		Message: "Added after task 1",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 4)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "New Task", result[1].Title)
	assert.Equal(t, int64(2), result[2].ID)
	assert.Equal(t, int64(3), result[3].ID)
}

func TestApplySubtaskOperations_AddAfterNonExistent(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	afterID := int64(99)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, AfterID: &afterID, Title: "New Task", Description: "New Description"},
		},
		Message: "Added after non-existent",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	// Should append to end when AfterID not found
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "New Task", result[1].Title)
}

func TestApplySubtaskOperations_ReorderToBeginning(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
	}

	id3 := int64(3)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpReorder, ID: &id3}, // AfterID nil = move to beginning
		},
		Message: "Moved task 3 to beginning",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, int64(3), result[0].ID)
	assert.Equal(t, int64(1), result[1].ID)
	assert.Equal(t, int64(2), result[2].ID)
}

func TestApplySubtaskOperations_ReorderAfterSpecific(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
		{ID: 3, Title: "Task 3", Description: "Description 3"},
	}

	id1, afterID := int64(1), int64(2)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpReorder, ID: &id1, AfterID: &afterID},
		},
		Message: "Moved task 1 after task 2",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, int64(2), result[0].ID)
	assert.Equal(t, int64(1), result[1].ID)
	assert.Equal(t, int64(3), result[2].ID)
}

func TestApplySubtaskOperations_ComplexScenario(t *testing.T) {
	// Simulates a real refiner scenario:
	// - Remove completed subtask
	// - Modify an existing subtask based on findings
	// - Add a new subtask to address a newly discovered issue
	planned := []database.Subtask{
		{ID: 10, Title: "Scan ports", Description: "Scan target ports"},
		{ID: 11, Title: "Enumerate services", Description: "Enumerate running services"},
		{ID: 12, Title: "Test vulnerabilities", Description: "Test for known vulnerabilities"},
	}

	id10, id11, afterID := int64(10), int64(11), int64(11)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove, ID: &id10},
			{Op: tools.SubtaskOpModify, ID: &id11, Description: "Enumerate services, focusing on web services found on port 80 and 443"},
			{Op: tools.SubtaskOpAdd, AfterID: &afterID, Title: "Check for SQL injection", Description: "Test web forms for SQL injection vulnerabilities"},
		},
		Message: "Refined plan based on port scan results",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)

	// First should be modified enumerate services
	assert.Equal(t, int64(11), result[0].ID)
	assert.Equal(t, "Enumerate services", result[0].Title)
	assert.Contains(t, result[0].Description, "port 80 and 443")

	// Second should be the new SQL injection task
	assert.Equal(t, int64(0), result[1].ID) // New task
	assert.Equal(t, "Check for SQL injection", result[1].Title)

	// Third should be the original vulnerability test
	assert.Equal(t, int64(12), result[2].ID)
}

func TestApplySubtaskOperations_RemoveAllTasks(t *testing.T) {
	// Simulates task completion - all remaining subtasks are removed
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
		{ID: 2, Title: "Task 2", Description: "Description 2"},
	}

	id1, id2 := int64(1), int64(2)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove, ID: &id1},
			{Op: tools.SubtaskOpRemove, ID: &id2},
		},
		Message: "Task completed, removing all remaining subtasks",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 0)
}

func TestApplySubtaskOperations_EmptyPlanned(t *testing.T) {
	planned := []database.Subtask{}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, Title: "New Task", Description: "Description"},
		},
		Message: "Adding first task",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 1)
	assert.Equal(t, "New Task", result[0].Title)
}

func TestApplySubtaskOperations_RemoveMissingID(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpRemove}, // Missing ID
		},
		Message: "Remove with missing ID",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove operation missing required id field")
}

func TestApplySubtaskOperations_ModifyMissingID(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, Title: "New Title"}, // Missing ID
		},
		Message: "Modify with missing ID",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modify operation missing required id field")
}

func TestApplySubtaskOperations_ModifyMissingTitleAndDescription(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id1 := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, ID: &id1}, // Missing both title and description
		},
		Message: "Modify with missing title and description",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modify operation missing both title and description")
}

func TestApplySubtaskOperations_ModifyNonExistent(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id99 := int64(99)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, ID: &id99, Title: "New Title"},
		},
		Message: "Modify non-existent task",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found for modification")
}

func TestApplySubtaskOperations_AddMissingTitle(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, Description: "Some description"}, // Missing title
		},
		Message: "Add with missing title",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add operation missing required title field")
}

func TestApplySubtaskOperations_AddMissingDescription(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, Title: "New Task"}, // Missing description
		},
		Message: "Add with missing description",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "add operation missing required description field")
}

func TestApplySubtaskOperations_ReorderMissingID(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpReorder}, // Missing ID
		},
		Message: "Reorder with missing ID",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reorder operation missing required id field")
}

func TestApplySubtaskOperations_ReorderNonExistent(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	id99 := int64(99)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpReorder, ID: &id99},
		},
		Message: "Reorder non-existent task",
	}

	_, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found for reorder")
}

func TestApplySubtaskOperations_MultipleAddsWithPositioning(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	afterID1 := int64(1)
	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpAdd, Title: "Task A", Description: "Desc A"},
			{Op: tools.SubtaskOpAdd, AfterID: &afterID1, Title: "Task B", Description: "Desc B"},
		},
		Message: "Multiple adds",
	}

	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)

	assert.Len(t, result, 3)
	// Task A at beginning
	assert.Equal(t, "Task A", result[0].Title)
	// Task 1 in middle
	assert.Equal(t, int64(1), result[1].ID)
	// Task B after Task 1
	assert.Equal(t, "Task B", result[2].Title)
}

func TestValidateSubtaskPatch_ValidOperations(t *testing.T) {
	id := int64(1)

	tests := []struct {
		name  string
		patch tools.SubtaskPatch
	}{
		{
			name: "empty operations",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
			},
		},
		{
			name: "valid add",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, Title: "Title", Description: "Desc"},
				},
			},
		},
		{
			name: "valid remove",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: &id},
				},
			},
		},
		{
			name: "valid modify with title",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: &id, Title: "New Title"},
				},
			},
		},
		{
			name: "valid modify with description",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: &id, Description: "New Desc"},
				},
			},
		},
		{
			name: "valid reorder",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpReorder, ID: &id},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubtaskPatch(tt.patch)
			assert.NoError(t, err)
		})
	}
}

func TestValidateSubtaskPatch_InvalidOperations(t *testing.T) {
	id := int64(1)

	tests := []struct {
		name          string
		patch         tools.SubtaskPatch
		expectedError string
	}{
		{
			name: "add missing title",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, Description: "Desc"},
				},
			},
			expectedError: "add requires title",
		},
		{
			name: "add missing description",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, Title: "Title"},
				},
			},
			expectedError: "add requires description",
		},
		{
			name: "remove missing id",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove},
				},
			},
			expectedError: "remove requires id",
		},
		{
			name: "modify missing id",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, Title: "Title"},
				},
			},
			expectedError: "modify requires id",
		},
		{
			name: "modify missing both title and description",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: &id},
				},
			},
			expectedError: "modify requires at least title or description",
		},
		{
			name: "reorder missing id",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpReorder},
				},
			},
			expectedError: "reorder requires id",
		},
		{
			name: "unknown operation type",
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: "invalid_op"},
				},
			},
			expectedError: "unknown operation type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubtaskPatch(tt.patch)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

