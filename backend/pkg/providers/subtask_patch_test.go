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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
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
	assert.Equal(t, "Task 1", result[0].Title) // Title unchanged
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

	// fixSubtaskPatch cleans non-existent afterID to nil, so insertion happens at beginning
	assert.Len(t, result, 2)
	assert.Equal(t, "New Task", result[0].Title) // New task at beginning
	assert.Equal(t, int64(1), result[1].ID)      // Original task moved to second position
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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
}

func TestApplySubtaskOperations_ModifyMissingID(t *testing.T) {
	planned := []database.Subtask{
		{ID: 1, Title: "Task 1", Description: "Description 1"},
	}

	patch := tools.SubtaskPatch{
		Operations: []tools.SubtaskOperation{
			{Op: tools.SubtaskOpModify, Title: "New Title", Description: "New Description"}, // Missing ID
		},
		Message: "Modify with missing ID",
	}

	// fixSubtaskPatch now converts modify with missing ID to add if title and description are present
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 2)                      // Original task + new task added
	assert.Equal(t, "New Title", result[0].Title) // New task at beginning
	assert.Equal(t, "New Description", result[0].Description)
	assert.Equal(t, int64(0), result[0].ID) // New task
	assert.Equal(t, int64(1), result[1].ID) // Original task moved to second position
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

	// Valid ID but missing both fields - this still gets validated and should error
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
			{Op: tools.SubtaskOpModify, ID: &id99, Title: "New Title", Description: "New Description"},
		},
		Message: "Modify non-existent task",
	}

	// fixSubtaskPatch now converts modify with non-existent ID to add (inserted at beginning)
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 2)                // Original task + new task added
	assert.Equal(t, int64(0), result[0].ID) // New task at beginning
	assert.Equal(t, "New Title", result[0].Title)
	assert.Equal(t, "New Description", result[0].Description)
	assert.Equal(t, int64(1), result[1].ID) // Original task moved to second position
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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
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

	// fixSubtaskPatch now filters out invalid operations, so this should succeed with no changes
	result, err := applySubtaskOperations(planned, patch, newTestLogger())
	require.NoError(t, err)
	assert.Len(t, result, 1) // No changes, operation was filtered out
	assert.Equal(t, int64(1), result[0].ID)
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

// TestFixSubtaskPatch tests the fixSubtaskPatch function with various LLM-generated error cases
func TestFixSubtaskPatch(t *testing.T) {
	tests := []struct {
		name     string
		planned  []database.Subtask
		patch    tools.SubtaskPatch
		expected tools.SubtaskPatch
	}{
		{
			name: "modify with non-existent ID converts to add",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
				{ID: 1847, Title: "Task 2", Description: "Desc 2"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          int64Ptr(1855), // Non-existent ID
						Title:       "New Task",
						Description: "New Description",
					},
				},
				Message: "Trying to modify non-existent task",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpAdd,
						ID:          nil,
						Title:       "New Task",
						Description: "New Description",
					},
				},
				Message: "Trying to modify non-existent task",
			},
		},
		{
			name: "modify with non-existent ID and missing title skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          int64Ptr(9999),
						Description: "Only description",
					},
				},
				Message: "Invalid modify",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Invalid modify",
			},
		},
		{
			name: "modify with non-existent ID and missing description skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:    tools.SubtaskOpModify,
						ID:    int64Ptr(9999),
						Title: "Only title",
					},
				},
				Message: "Invalid modify",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Invalid modify",
			},
		},
		{
			name: "remove with non-existent ID skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: int64Ptr(9999)},
				},
				Message: "Remove non-existent",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Remove non-existent",
			},
		},
		{
			name: "reorder with non-existent ID skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpReorder, ID: int64Ptr(9999)},
				},
				Message: "Reorder non-existent",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Reorder non-existent",
			},
		},
		{
			name: "add with empty title skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, Description: "Desc only"},
				},
				Message: "Add without title",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Add without title",
			},
		},
		{
			name: "add with empty description skipped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, Title: "Title only"},
				},
				Message: "Add without description",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Add without description",
			},
		},
		{
			name: "valid modify with existing ID preserved",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          int64Ptr(1846),
						Title:       "Updated Title",
						Description: "Updated Desc",
					},
				},
				Message: "Valid modify",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          int64Ptr(1846),
						Title:       "Updated Title",
						Description: "Updated Desc",
					},
				},
				Message: "Valid modify",
			},
		},
		{
			name: "afterID with non-existent value cleaned",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpAdd,
						AfterID:     int64Ptr(9999), // Non-existent
						Title:       "New Task",
						Description: "New Desc",
					},
				},
				Message: "Add with invalid afterID",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpAdd,
						AfterID:     nil, // Cleaned
						Title:       "New Task",
						Description: "New Desc",
					},
				},
				Message: "Add with invalid afterID",
			},
		},
		{
			name: "complex scenario with multiple errors",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
				{ID: 1847, Title: "Task 2", Description: "Desc 2"},
				{ID: 1848, Title: "Task 3", Description: "Desc 3"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					// Valid remove
					{Op: tools.SubtaskOpRemove, ID: int64Ptr(1846)},
					// Invalid remove (non-existent ID) - should be skipped
					{Op: tools.SubtaskOpRemove, ID: int64Ptr(9999)},
					// Valid modify
					{Op: tools.SubtaskOpModify, ID: int64Ptr(1847), Title: "Updated"},
					// Invalid modify (non-existent ID) - should convert to add
					{Op: tools.SubtaskOpModify, ID: int64Ptr(1855), Title: "New Task", Description: "New Desc"},
					// Invalid modify (non-existent ID, missing fields) - should be skipped
					{Op: tools.SubtaskOpModify, ID: int64Ptr(1856), Title: "No Desc"},
					// Valid add
					{Op: tools.SubtaskOpAdd, Title: "Added Task", Description: "Added Desc"},
					// Invalid reorder (non-existent ID) - should be skipped
					{Op: tools.SubtaskOpReorder, ID: int64Ptr(9998)},
				},
				Message: "Complex scenario",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: int64Ptr(1846)},
					{Op: tools.SubtaskOpModify, ID: int64Ptr(1847), Title: "Updated"},
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "New Task", Description: "New Desc"},
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Added Task", Description: "Added Desc"},
				},
				Message: "Complex scenario",
			},
		},
		{
			name: "empty ID (nil) for modify with valid fields converts to add",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: nil, Title: "Title", Description: "Desc"},
					{Op: tools.SubtaskOpRemove, ID: nil},
					{Op: tools.SubtaskOpReorder, ID: nil},
				},
				Message: "Nil IDs",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					// Modify with nil ID and valid fields converts to ADD
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Title", Description: "Desc"},
					// Remove and reorder with nil IDs are skipped
				},
				Message: "Nil IDs",
			},
		},
		{
			name: "zero ID for modify with valid fields converts to add",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: int64Ptr(0), Title: "Title", Description: "Desc"},
					{Op: tools.SubtaskOpRemove, ID: int64Ptr(0)},
					{Op: tools.SubtaskOpReorder, ID: int64Ptr(0)},
				},
				Message: "Zero IDs",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					// Modify with zero ID and valid fields converts to ADD
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Title", Description: "Desc"},
					// Remove and reorder with zero IDs are skipped
				},
				Message: "Zero IDs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixSubtaskPatch(tt.planned, tt.patch)

			assert.Equal(t, tt.expected.Message, result.Message, "Message mismatch")
			assert.Equal(t, len(tt.expected.Operations), len(result.Operations), "Operations count mismatch")

			for i, expectedOp := range tt.expected.Operations {
				if i >= len(result.Operations) {
					t.Errorf("Missing operation %d", i)
					continue
				}

				resultOp := result.Operations[i]

				assert.Equal(t, expectedOp.Op, resultOp.Op, "Operation %d: Op mismatch", i)
				assert.Equal(t, expectedOp.Title, resultOp.Title, "Operation %d: Title mismatch", i)
				assert.Equal(t, expectedOp.Description, resultOp.Description, "Operation %d: Description mismatch", i)

				// Check ID
				if expectedOp.ID == nil {
					assert.Nil(t, resultOp.ID, "Operation %d: ID should be nil", i)
				} else {
					require.NotNil(t, resultOp.ID, "Operation %d: ID should not be nil", i)
					assert.Equal(t, *expectedOp.ID, *resultOp.ID, "Operation %d: ID value mismatch", i)
				}

				// Check AfterID
				if expectedOp.AfterID == nil {
					assert.Nil(t, resultOp.AfterID, "Operation %d: AfterID should be nil", i)
				} else {
					require.NotNil(t, resultOp.AfterID, "Operation %d: AfterID should not be nil", i)
					assert.Equal(t, *expectedOp.AfterID, *resultOp.AfterID, "Operation %d: AfterID value mismatch", i)
				}
			}
		})
	}
}

// Helper function to create int64 pointer
func int64Ptr(v int64) *int64 {
	return &v
}

// TestApplySubtaskOperations_EdgeCases tests edge cases found during audit
func TestApplySubtaskOperations_EdgeCases(t *testing.T) {
	t.Run("modify then remove same task", func(t *testing.T) {
		// Test that modify is applied even if task is later removed
		planned := []database.Subtask{
			{ID: 10, Title: "Task 1", Description: "Desc 1"},
			{ID: 11, Title: "Task 2", Description: "Desc 2"},
		}

		id10 := int64(10)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpModify, ID: &id10, Title: "Modified Title"},
				{Op: tools.SubtaskOpRemove, ID: &id10},
			},
			Message: "Modify then remove",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		// Task 10 should be removed (modify was applied but then removed)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(11), result[0].ID)
		assert.Equal(t, "Task 2", result[0].Title)
	})

	t.Run("multiple reorder of same task", func(t *testing.T) {
		// Test that multiple reorders result in final position
		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
			{ID: 3, Title: "Task 3", Description: "Desc 3"},
		}

		id1, afterID2, afterID3 := int64(1), int64(2), int64(3)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpReorder, ID: &id1, AfterID: &afterID2}, // Task1 after Task2
				{Op: tools.SubtaskOpReorder, ID: &id1, AfterID: &afterID3}, // Task1 after Task3 (final)
			},
			Message: "Multiple reorders",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		assert.Len(t, result, 3)
		assert.Equal(t, int64(2), result[0].ID) // Task 2
		assert.Equal(t, int64(3), result[1].ID) // Task 3
		assert.Equal(t, int64(1), result[2].ID) // Task 1 (moved to end)
	})

	t.Run("reorder to current position (no-op)", func(t *testing.T) {
		// Test reordering to the same position
		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
			{ID: 3, Title: "Task 3", Description: "Desc 3"},
		}

		id2, afterID1 := int64(2), int64(1)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpReorder, ID: &id2, AfterID: &afterID1}, // Task2 already after Task1
			},
			Message: "No-op reorder",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		// Order should remain the same
		assert.Len(t, result, 3)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, int64(3), result[2].ID)
	})

	t.Run("remove then modify same task filtered by fixSubtaskPatch", func(t *testing.T) {
		// fixSubtaskPatch should filter out modify with non-existent ID (after remove in first pass)
		// But since operations are processed in order, remove happens in first pass,
		// modify happens in first pass too (before removal is applied to result)
		// So modify will be applied, then task will be removed
		planned := []database.Subtask{
			{ID: 10, Title: "Task 1", Description: "Desc 1"},
			{ID: 11, Title: "Task 2", Description: "Desc 2"},
		}

		id10 := int64(10)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: &id10},
				{Op: tools.SubtaskOpModify, ID: &id10, Title: "Modified Title"},
			},
			Message: "Remove then modify",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		// Task 10 should be removed (modify was applied in first pass but then removed)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(11), result[0].ID)
	})

	t.Run("add with non-existent afterID inserts at beginning", func(t *testing.T) {
		// fixSubtaskPatch cleans non-existent afterID to nil → insert at beginning
		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
		}

		afterID := int64(999)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpAdd, AfterID: &afterID, Title: "New Task", Description: "New Desc"},
			},
			Message: "Add with invalid afterID",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		assert.Len(t, result, 3)
		assert.Equal(t, "New Task", result[0].Title) // Inserted at beginning (afterID cleaned to nil)
		assert.Equal(t, int64(1), result[1].ID)
		assert.Equal(t, int64(2), result[2].ID)
	})

	t.Run("modify existing task preserves position", func(t *testing.T) {
		// Verify that modify doesn't change task position
		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
			{ID: 3, Title: "Task 3", Description: "Desc 3"},
		}

		id2 := int64(2)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpModify, ID: &id2, Title: "Modified Task 2"},
			},
			Message: "Modify preserves position",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		assert.Len(t, result, 3)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, int64(2), result[1].ID)
		assert.Equal(t, "Modified Task 2", result[1].Title)
		assert.Equal(t, int64(3), result[2].ID)
	})

	t.Run("complex interleaved operations", func(t *testing.T) {
		// Test complex scenario with add, remove, modify, reorder
		// Initial: [Task1, Task2, Task3, Task4]
		// Operations:
		// 1. Remove Task1 (first pass)
		// 2. Modify Task3 (first pass)
		// After first pass: [Task2, Task3(modified), Task4]
		// 3. Add "New" after Task3 (second pass) → [Task2, Task3(modified), New, Task4]
		// 4. Reorder Task2 after Task3 (second pass) → [Task3(modified), Task2, New, Task4]

		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
			{ID: 3, Title: "Task 3", Description: "Desc 3"},
			{ID: 4, Title: "Task 4", Description: "Desc 4"},
		}

		id1, id2, id3, afterID3 := int64(1), int64(2), int64(3), int64(3)
		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: &id1},                                        // Remove Task 1
				{Op: tools.SubtaskOpModify, ID: &id3, Title: "Modified Task 3"},              // Modify Task 3
				{Op: tools.SubtaskOpAdd, AfterID: &afterID3, Title: "New", Description: "D"}, // Add after Task 3
				{Op: tools.SubtaskOpReorder, ID: &id2, AfterID: &id3},                        // Move Task 2 after Task 3
			},
			Message: "Complex operations",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		// Expected order after all operations: [Task3(modified), Task2, New, Task4]
		assert.Len(t, result, 4)
		assert.Equal(t, int64(3), result[0].ID) // Task 3 (modified, stays in place)
		assert.Equal(t, "Modified Task 3", result[0].Title)
		assert.Equal(t, int64(2), result[1].ID) // Task 2 (moved after Task 3)
		assert.Equal(t, int64(0), result[2].ID) // New task (added after Task 3)
		assert.Equal(t, "New", result[2].Title)
		assert.Equal(t, int64(4), result[3].ID) // Task 4 (unchanged position)
	})

	t.Run("empty operations does not change order", func(t *testing.T) {
		planned := []database.Subtask{
			{ID: 1, Title: "Task 1", Description: "Desc 1"},
			{ID: 2, Title: "Task 2", Description: "Desc 2"},
		}

		patch := tools.SubtaskPatch{
			Operations: []tools.SubtaskOperation{},
			Message:    "No operations",
		}

		result, err := applySubtaskOperations(planned, patch, newTestLogger())
		require.NoError(t, err)

		assert.Len(t, result, 2)
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, int64(2), result[1].ID)
	})
}
