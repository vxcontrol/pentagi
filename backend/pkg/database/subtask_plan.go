package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// SaveSubtaskPlan commits deletion and creation together. Text is normalized
// at this boundary because PostgreSQL text cannot contain NUL bytes.
func SaveSubtaskPlan(ctx context.Context, db Querier, deleteIDs []int64, plan []CreateSubtaskParams) error {
	writer, ok := db.(interface {
		SaveSubtaskPlan(context.Context, []int64, []CreateSubtaskParams) error
	})
	if !ok {
		return errors.New("subtask plan storage does not support transactions")
	}
	return writer.SaveSubtaskPlan(ctx, deleteIDs, plan)
}

func (q *Queries) SaveSubtaskPlan(ctx context.Context, deleteIDs []int64, plan []CreateSubtaskParams) error {
	beginner, ok := q.db.(interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	})
	if !ok {
		return errors.New("subtask plan storage cannot begin a transaction")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin subtask plan transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	queries := q.WithTx(tx)
	if len(deleteIDs) > 0 {
		if err := queries.DeleteSubtasks(ctx, deleteIDs); err != nil {
			return fmt.Errorf("delete planned subtasks: %w", err)
		}
	}
	for _, item := range plan {
		item.Title = SanitizeUTF8(item.Title)
		item.Description = SanitizeUTF8(item.Description)
		if _, err := queries.CreateSubtask(ctx, item); err != nil {
			return fmt.Errorf("create subtask for task %d: %w", item.TaskID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subtask plan: %w", err)
	}
	return nil
}
