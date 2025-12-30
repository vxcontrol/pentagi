package providers

import (
	"fmt"
	"slices"

	"pentagi/pkg/database"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

// applySubtaskOperations applies delta operations to the current planned subtasks
// and returns the updated list of SubtaskInfoPatch. Operations are applied in order.
// Returns an error if any operation has missing required fields.
func applySubtaskOperations(
	planned []database.Subtask,
	patch tools.SubtaskPatch,
	logger *logrus.Entry,
) ([]tools.SubtaskInfoPatch, error) {
	logger.WithFields(logrus.Fields{
		"planned_count":    len(planned),
		"operations_count": len(patch.Operations),
		"message":          patch.Message,
	}).Debug("applying subtask operations")

	// Convert database.Subtask to tools.SubtaskInfo with IDs
	result := make([]tools.SubtaskInfoPatch, 0, len(planned))
	for _, st := range planned {
		result = append(result, tools.SubtaskInfoPatch{
			ID: st.ID,
			SubtaskInfo: tools.SubtaskInfo{
				Title:       st.Title,
				Description: st.Description,
			},
		})
	}

	// Build ID -> index map for position lookups
	idToIdx := buildIndexMap(result)

	// Track removals separately to avoid modifying the slice during iteration
	removed := make(map[int64]bool)

	// First pass: process removals and modifications in-place
	for i, op := range patch.Operations {
		opLogger := logger.WithFields(logrus.Fields{
			"operation_index": i,
			"operation":       op.Op,
			"id":              op.ID,
			"after_id":        op.AfterID,
		})

		switch op.Op {
		case tools.SubtaskOpRemove:
			if op.ID == nil {
				err := fmt.Errorf("operation %d: remove operation missing required id field", i)
				opLogger.Error(err.Error())
				return nil, err
			}
			if _, ok := idToIdx[*op.ID]; !ok {
				err := fmt.Errorf("operation %d: subtask with id %d not found for removal", i, *op.ID)
				opLogger.Error(err.Error())
				return nil, err
			}
			removed[*op.ID] = true
			opLogger.WithField("subtask_id", *op.ID).Debug("marked subtask for removal")

		case tools.SubtaskOpModify:
			if op.ID == nil {
				err := fmt.Errorf("operation %d: modify operation missing required id field", i)
				opLogger.Error(err.Error())
				return nil, err
			}
			if op.Title == "" && op.Description == "" {
				err := fmt.Errorf("operation %d: modify operation missing both title and description fields", i)
				opLogger.Error(err.Error())
				return nil, err
			}
			idx, ok := idToIdx[*op.ID]
			if !ok {
				err := fmt.Errorf("operation %d: subtask with id %d not found for modification", i, *op.ID)
				opLogger.Error(err.Error())
				return nil, err
			}
			// Only update fields that are provided
			if op.Title != "" {
				result[idx].Title = op.Title
				opLogger.WithField("new_title", op.Title).Debug("updated subtask title")
			}
			if op.Description != "" {
				result[idx].Description = op.Description
				opLogger.WithField("new_description_len", len(op.Description)).Debug("updated subtask description")
			}
		}
	}

	// Build result list (excluding removed subtasks)
	if len(removed) > 0 {
		filtered := make([]tools.SubtaskInfoPatch, 0, len(result)-len(removed))
		for _, st := range result {
			if !removed[st.ID] {
				filtered = append(filtered, st)
			}
		}
		result = filtered
		logger.WithField("removed_count", len(removed)).Debug("filtered out removed subtasks")
	}

	// Rebuild index map for the filtered result
	idToIdx = buildIndexMap(result)

	// Second pass: process adds and reorders with position awareness
	for i, op := range patch.Operations {
		opLogger := logger.WithFields(logrus.Fields{
			"operation_index": i,
			"operation":       op.Op,
			"id":              op.ID,
			"after_id":        op.AfterID,
		})

		switch op.Op {
		case tools.SubtaskOpAdd:
			if op.Title == "" {
				err := fmt.Errorf("operation %d: add operation missing required title field", i)
				opLogger.Error(err.Error())
				return nil, err
			}
			if op.Description == "" {
				err := fmt.Errorf("operation %d: add operation missing required description field", i)
				opLogger.Error(err.Error())
				return nil, err
			}

			newSubtask := tools.SubtaskInfoPatch{
				ID: 0, // New subtasks don't have an ID yet
				SubtaskInfo: tools.SubtaskInfo{
					Title:       op.Title,
					Description: op.Description,
				},
			}

			insertIdx := calculateInsertIndex(op.AfterID, idToIdx, len(result))
			result = slices.Insert(result, insertIdx, newSubtask)

			// Rebuild index map after insertion
			idToIdx = buildIndexMap(result)

			opLogger.WithFields(logrus.Fields{
				"insert_idx": insertIdx,
				"title":      op.Title,
			}).Debug("inserted new subtask")

		case tools.SubtaskOpReorder:
			if op.ID == nil {
				err := fmt.Errorf("operation %d: reorder operation missing required id field", i)
				opLogger.Error(err.Error())
				return nil, err
			}

			currentIdx, ok := idToIdx[*op.ID]
			if !ok {
				err := fmt.Errorf("operation %d: subtask with id %d not found for reorder", i, *op.ID)
				opLogger.Error(err.Error())
				return nil, err
			}

			// Remove from current position
			subtaskToMove := result[currentIdx]
			result = slices.Delete(result, currentIdx, currentIdx+1)

			// Rebuild index map after deletion
			idToIdx = buildIndexMap(result)

			// Calculate new position and insert
			insertIdx := calculateInsertIndex(op.AfterID, idToIdx, len(result))
			result = slices.Insert(result, insertIdx, subtaskToMove)

			// Rebuild index map after insertion
			idToIdx = buildIndexMap(result)

			opLogger.WithFields(logrus.Fields{
				"from_idx": currentIdx,
				"to_idx":   insertIdx,
			}).Debug("reordered subtask")
		}
	}

	logger.WithFields(logrus.Fields{
		"final_count":   len(result),
		"initial_count": len(planned),
	}).Debug("completed applying subtask operations")

	return result, nil
}

// convertSubtaskInfoPatch removes the ID field from the subtasks info patches
func convertSubtaskInfoPatch(subtasks []tools.SubtaskInfoPatch) []tools.SubtaskInfo {
	result := make([]tools.SubtaskInfo, 0, len(subtasks))
	for _, st := range subtasks {
		result = append(result, tools.SubtaskInfo{
			Title:       st.Title,
			Description: st.Description,
		})
	}
	return result
}

// buildIndexMap creates a map from subtask ID to its index in the slice
func buildIndexMap(subtasks []tools.SubtaskInfoPatch) map[int64]int {
	idToIdx := make(map[int64]int, len(subtasks))
	for i, st := range subtasks {
		if st.ID != 0 {
			idToIdx[st.ID] = i
		}
	}
	return idToIdx
}

// calculateInsertIndex determines the insertion index based on afterID
func calculateInsertIndex(afterID *int64, idToIdx map[int64]int, length int) int {
	if afterID == nil || *afterID == 0 {
		return 0 // Insert at beginning
	}

	if idx, ok := idToIdx[*afterID]; ok {
		return idx + 1 // Insert after the referenced subtask
	}

	// AfterID not found, append to end
	return length
}

// ValidateSubtaskPatch validates the operations in a SubtaskPatch
func ValidateSubtaskPatch(patch tools.SubtaskPatch) error {
	for i, op := range patch.Operations {
		switch op.Op {
		case tools.SubtaskOpAdd:
			if op.Title == "" {
				return fmt.Errorf("operation %d: add requires title", i)
			}
			if op.Description == "" {
				return fmt.Errorf("operation %d: add requires description", i)
			}
		case tools.SubtaskOpRemove:
			if op.ID == nil {
				return fmt.Errorf("operation %d: remove requires id", i)
			}
		case tools.SubtaskOpModify:
			if op.ID == nil {
				return fmt.Errorf("operation %d: modify requires id", i)
			}
			if op.Title == "" && op.Description == "" {
				return fmt.Errorf("operation %d: modify requires at least title or description", i)
			}
		case tools.SubtaskOpReorder:
			if op.ID == nil {
				return fmt.Errorf("operation %d: reorder requires id", i)
			}
		default:
			return fmt.Errorf("operation %d: unknown operation type %q", i, op.Op)
		}
	}
	return nil
}
