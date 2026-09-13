package tools

import (
	"testing"

	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

func TestGetGlobalFilters_NoTaskOrSubtask(t *testing.T) {
	t.Parallel()

	filters := map[string]any{
		"flow_id":  "1",
		"doc_type": "memory",
	}

	isSpecific, global := getGlobalFilters(filters)

	if isSpecific {
		t.Error("expected isSpecific=false when no task_id/subtask_id present")
	}
	if len(global) != len(filters) {
		t.Errorf("expected global filters unchanged in size, got %d want %d", len(global), len(filters))
	}
}

func TestGetGlobalFilters_WithTaskID(t *testing.T) {
	t.Parallel()

	filters := map[string]any{
		"flow_id":  "1",
		"doc_type": "memory",
		"task_id":  "42",
	}

	isSpecific, global := getGlobalFilters(filters)

	if !isSpecific {
		t.Error("expected isSpecific=true when task_id present")
	}
	if _, ok := global["task_id"]; ok {
		t.Error("expected task_id to be removed from global filters")
	}
	if len(global) != 2 {
		t.Errorf("expected 2 remaining global filters, got %d", len(global))
	}
}

func TestGetGlobalFilters_WithSubtaskID(t *testing.T) {
	t.Parallel()

	filters := map[string]any{
		"flow_id":    "1",
		"subtask_id": "7",
	}

	isSpecific, global := getGlobalFilters(filters)

	if !isSpecific {
		t.Error("expected isSpecific=true when subtask_id present")
	}
	if _, ok := global["subtask_id"]; ok {
		t.Error("expected subtask_id to be removed from global filters")
	}
}

func TestGetGlobalFilters_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	filters := map[string]any{
		"flow_id": "1",
		"task_id": "42",
	}

	_, global := getGlobalFilters(filters)
	global["extra"] = "value"

	if _, ok := filters["extra"]; ok {
		t.Error("modifying the returned global filters map should not mutate the input map")
	}
	if _, ok := filters["task_id"]; !ok {
		t.Error("original filters map should still contain task_id")
	}
}

func TestMemory_IsAvailable_NilStore(t *testing.T) {
	t.Parallel()

	m := &memory{store: nil}

	if m.IsAvailable() {
		t.Error("expected IsAvailable() to be false when store is nil")
	}
}

func TestMemory_IsAvailable_NonNilStore(t *testing.T) {
	t.Parallel()

	m := &memory{store: &pgvector.Store{}}

	if !m.IsAvailable() {
		t.Error("expected IsAvailable() to be true when store is non-nil")
	}
}
