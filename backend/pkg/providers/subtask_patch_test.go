package providers

import (
	"fmt"
	"io"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subtaskPatchLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logrus.NewEntry(logger)
}

func subtaskPatchID(v int64) *int64 {
	return &v
}

func subtaskPatchPlan(ids ...int64) []database.Subtask {
	planned := make([]database.Subtask, 0, len(ids))
	for _, id := range ids {
		planned = append(planned, database.Subtask{
			ID: id, Title: fmt.Sprintf("Task %d", id), Description: fmt.Sprintf("Description %d", id),
		})
	}

	return planned
}

// subtaskPatchKept is planned subtask N as it comes out untouched.
func subtaskPatchKept(id int64) tools.SubtaskInfoPatch {
	return tools.SubtaskInfoPatch{ID: id, SubtaskInfo: tools.SubtaskInfo{
		Title: fmt.Sprintf("Task %d", id), Description: fmt.Sprintf("Description %d", id),
	}}
}

func subtaskPatchInfo(id int64, title, description string) tools.SubtaskInfoPatch {
	return tools.SubtaskInfoPatch{ID: id, SubtaskInfo: tools.SubtaskInfo{Title: title, Description: description}}
}

// Removals and modifies apply before adds and reorders; a new subtask has no ID yet.
func TestSubtaskPatch_ApplySubtaskOperations_ProducesThePlanThePatchDescribes(t *testing.T) {
	tests := []struct {
		name    string
		planned []database.Subtask
		ops     []tools.SubtaskOperation
		want    []tools.SubtaskInfoPatch
		wantErr string
	}{
		{
			name:    "an empty patch keeps the plan",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{},
			want:    []tools.SubtaskInfoPatch{subtaskPatchKept(1), subtaskPatchKept(2), subtaskPatchKept(3)},
		},
		{
			name:    "a removal drops the subtask",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(2)}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchKept(1), subtaskPatchKept(3)},
		},
		{
			name:    "several removals drop each subtask",
			planned: subtaskPatchPlan(1, 2, 3, 4),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(3)},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchKept(2), subtaskPatchKept(4)},
		},
		{
			name:    "removing every subtask leaves an empty plan",
			planned: subtaskPatchPlan(1, 2),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(2)},
			},
			want: []tools.SubtaskInfoPatch{},
		},
		{
			name:    "a modify with a title keeps the description and the position",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpModify, ID: subtaskPatchID(2), Title: "Modified Task 2"}},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchKept(1), subtaskPatchInfo(2, "Modified Task 2", "Description 2"), subtaskPatchKept(3),
			},
		},
		{
			name:    "a modify with a description keeps the title",
			planned: subtaskPatchPlan(1),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1), Description: "New Description"}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchInfo(1, "Task 1", "New Description")},
		},
		{
			name:    "a modify with both replaces both",
			planned: subtaskPatchPlan(1),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1), Title: "New Title", Description: "New Description"},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchInfo(1, "New Title", "New Description")},
		},
		{
			name:    "an add without an anchor goes first",
			planned: subtaskPatchPlan(1, 2),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpAdd, Title: "New Task", Description: "New Description"}},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchInfo(0, "New Task", "New Description"), subtaskPatchKept(1), subtaskPatchKept(2),
			},
		},
		{
			name:    "an add after a subtask goes right after it",
			planned: subtaskPatchPlan(1, 2, 3),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpAdd, AfterID: subtaskPatchID(1), Title: "New Task", Description: "New Description"},
			},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchKept(1), subtaskPatchInfo(0, "New Task", "New Description"), subtaskPatchKept(2), subtaskPatchKept(3),
			},
		},
		{
			name:    "an add to an empty plan is the whole plan",
			planned: subtaskPatchPlan(),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpAdd, Title: "New Task", Description: "Description"}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchInfo(0, "New Task", "Description")},
		},
		{
			name:    "adds are placed in the order they come",
			planned: subtaskPatchPlan(1),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpAdd, Title: "Task A", Description: "Desc A"},
				{Op: tools.SubtaskOpAdd, AfterID: subtaskPatchID(1), Title: "Task B", Description: "Desc B"},
			},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchInfo(0, "Task A", "Desc A"), subtaskPatchKept(1), subtaskPatchInfo(0, "Task B", "Desc B"),
			},
		},
		{
			name:    "an add anchored to a subtask removed in the same patch goes last",
			planned: subtaskPatchPlan(1, 2, 3),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(2)},
				{Op: tools.SubtaskOpAdd, AfterID: subtaskPatchID(2), Title: "New", Description: "D"},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchKept(1), subtaskPatchKept(3), subtaskPatchInfo(0, "New", "D")},
		},
		{
			name:    "a reorder without an anchor moves the subtask first",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(3)}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchKept(3), subtaskPatchKept(1), subtaskPatchKept(2)},
		},
		{
			name:    "a reorder after a subtask moves it right after that one",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(1), AfterID: subtaskPatchID(2)}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchKept(2), subtaskPatchKept(1), subtaskPatchKept(3)},
		},
		{
			name:    "the last of several reorders of one subtask wins",
			planned: subtaskPatchPlan(1, 2, 3),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(1), AfterID: subtaskPatchID(2)},
				{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(1), AfterID: subtaskPatchID(3)},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchKept(2), subtaskPatchKept(3), subtaskPatchKept(1)},
		},
		{
			name:    "a reorder to where the subtask already is changes nothing",
			planned: subtaskPatchPlan(1, 2, 3),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(2), AfterID: subtaskPatchID(1)}},
			want:    []tools.SubtaskInfoPatch{subtaskPatchKept(1), subtaskPatchKept(2), subtaskPatchKept(3)},
		},
		{
			name:    "a removal wins over an earlier modify of the same subtask",
			planned: subtaskPatchPlan(1, 2),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1), Title: "Modified Title"},
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchKept(2)},
		},
		{
			name:    "a removal wins over a later modify of the same subtask",
			planned: subtaskPatchPlan(1, 2),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
				{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1), Title: "Modified Title"},
			},
			want: []tools.SubtaskInfoPatch{subtaskPatchKept(2)},
		},
		{
			name:    "removals and modifies apply before adds and reorders",
			planned: subtaskPatchPlan(1, 2, 3, 4),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
				{Op: tools.SubtaskOpModify, ID: subtaskPatchID(3), Title: "Modified Task 3"},
				{Op: tools.SubtaskOpAdd, AfterID: subtaskPatchID(3), Title: "New", Description: "D"},
				{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(2), AfterID: subtaskPatchID(3)},
			},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchInfo(3, "Modified Task 3", "Description 3"),
				subtaskPatchKept(2),
				subtaskPatchInfo(0, "New", "D"),
				subtaskPatchKept(4),
			},
		},
		{
			// fixSubtaskPatch's own table pins each of these operations one by one.
			name:    "a patch the model got wrong is repaired before it is applied",
			planned: subtaskPatchPlan(1, 2),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(99)},
				{Op: tools.SubtaskOpRemove},
				{Op: tools.SubtaskOpAdd, Description: "Some description"},
				{Op: tools.SubtaskOpAdd, Title: "New Task"},
				{Op: tools.SubtaskOpReorder},
				{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(99)},
				{Op: tools.SubtaskOpModify, ID: subtaskPatchID(99), Title: "From Unknown", Description: "D1"},
				{Op: tools.SubtaskOpModify, Title: "From Missing", Description: "D2"},
				{Op: tools.SubtaskOpAdd, AfterID: subtaskPatchID(999), Title: "After Unknown", Description: "D3"},
			},
			want: []tools.SubtaskInfoPatch{
				subtaskPatchInfo(0, "After Unknown", "D3"),
				subtaskPatchInfo(0, "From Missing", "D2"),
				subtaskPatchInfo(0, "From Unknown", "D1"),
				subtaskPatchKept(1),
				subtaskPatchKept(2),
			},
		},
		{
			name:    "a modify with neither title nor description is refused",
			planned: subtaskPatchPlan(1),
			ops:     []tools.SubtaskOperation{{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1)}},
			wantErr: "operation 0: modify operation missing both title and description fields",
		},
		{
			name:    "a reorder of a subtask removed in the same patch is refused",
			planned: subtaskPatchPlan(1, 2),
			ops: []tools.SubtaskOperation{
				{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1)},
				{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(1), AfterID: subtaskPatchID(2)},
			},
			wantErr: "operation 1: subtask with id 1 not found for reorder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applySubtaskOperations(tt.planned, tools.SubtaskPatch{Operations: tt.ops}, subtaskPatchLogger())

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// One bad operation must not throw away a whole refinement.
func TestSubtaskPatch_FixSubtaskPatch_RepairsWhatTheModelGotWrong(t *testing.T) {
	tests := []struct {
		name     string
		planned  []database.Subtask
		patch    tools.SubtaskPatch
		expected tools.SubtaskPatch
	}{
		{
			name: "a modify of an unknown subtask becomes an add",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
				{ID: 1847, Title: "Task 2", Description: "Desc 2"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          subtaskPatchID(1855),
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
			name: "a modify of an unknown subtask without a title is dropped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          subtaskPatchID(9999),
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
			name: "a modify of an unknown subtask without a description is dropped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:    tools.SubtaskOpModify,
						ID:    subtaskPatchID(9999),
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
			name: "a removal of an unknown subtask is dropped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(9999)},
				},
				Message: "Remove non-existent",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Remove non-existent",
			},
		},
		{
			name: "a reorder of an unknown subtask is dropped",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(9999)},
				},
				Message: "Reorder non-existent",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{},
				Message:    "Reorder non-existent",
			},
		},
		{
			name: "an add without a title is dropped",
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
			name: "an add without a description is dropped",
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
			name: "a modify of a planned subtask is kept",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpModify,
						ID:          subtaskPatchID(1846),
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
						ID:          subtaskPatchID(1846),
						Title:       "Updated Title",
						Description: "Updated Desc",
					},
				},
				Message: "Valid modify",
			},
		},
		{
			name: "an anchor to an unknown subtask is cleared",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{
						Op:          tools.SubtaskOpAdd,
						AfterID:     subtaskPatchID(9999),
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
						AfterID:     nil,
						Title:       "New Task",
						Description: "New Desc",
					},
				},
				Message: "Add with invalid afterID",
			},
		},
		{
			name: "valid operations survive among invalid ones",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
				{ID: 1847, Title: "Task 2", Description: "Desc 2"},
				{ID: 1848, Title: "Task 3", Description: "Desc 3"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1846)},
					{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(9999)},
					{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1847), Title: "Updated"},
					{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1855), Title: "New Task", Description: "New Desc"},
					{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1856), Title: "No Desc"},
					{Op: tools.SubtaskOpAdd, Title: "Added Task", Description: "Added Desc"},
					{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(9998)},
				},
				Message: "Complex scenario",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(1846)},
					{Op: tools.SubtaskOpModify, ID: subtaskPatchID(1847), Title: "Updated"},
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "New Task", Description: "New Desc"},
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Added Task", Description: "Added Desc"},
				},
				Message: "Complex scenario",
			},
		},
		{
			name: "a modify without an ID becomes an add and a removal or reorder without one is dropped",
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
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Title", Description: "Desc"},
				},
				Message: "Nil IDs",
			},
		},
		{
			name: "a zero ID counts as no ID",
			planned: []database.Subtask{
				{ID: 1846, Title: "Task 1", Description: "Desc 1"},
			},
			patch: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpModify, ID: subtaskPatchID(0), Title: "Title", Description: "Desc"},
					{Op: tools.SubtaskOpRemove, ID: subtaskPatchID(0)},
					{Op: tools.SubtaskOpReorder, ID: subtaskPatchID(0)},
				},
				Message: "Zero IDs",
			},
			expected: tools.SubtaskPatch{
				Operations: []tools.SubtaskOperation{
					{Op: tools.SubtaskOpAdd, ID: nil, Title: "Title", Description: "Desc"},
				},
				Message: "Zero IDs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, fixSubtaskPatch(tt.planned, tt.patch))
		})
	}
}

func TestSubtaskPatch_ConvertSubtaskInfoPatch_DropsOnlyTheIDs(t *testing.T) {
	got := convertSubtaskInfoPatch([]tools.SubtaskInfoPatch{
		subtaskPatchInfo(7, "Scan ports", "Scan target ports"),
		subtaskPatchInfo(0, "Check for SQL injection", "Test web forms for SQL injection"),
	})

	assert.Equal(t, []tools.SubtaskInfo{
		{Title: "Scan ports", Description: "Scan target ports"},
		{Title: "Check for SQL injection", Description: "Test web forms for SQL injection"},
	}, got)
}
