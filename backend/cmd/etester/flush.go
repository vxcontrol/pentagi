package main

import (
	"fmt"
	"os"

	"pentagi/pkg/terminal"
)

// flush deletes all documents from the embedding store
func (t *Tester) flush() error {
	terminal.Warning("This will delete ALL documents from the embedding store.")
	response, err := terminal.GetYesNoInputContext(t.ctx, "Are you sure you want to continue?", os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to get yes/no input: %w", err)
	}

	if !response {
		terminal.Info("Operation cancelled.")
		return nil
	}

	tx, err := t.conn.Begin(t.ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(t.ctx)

	result, err := tx.Exec(t.ctx, fmt.Sprintf("DELETE FROM %s", t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	if err := tx.Commit(t.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected := result.RowsAffected()
	terminal.Success("\nSuccessfully deleted %d documents from the embedding store.", rowsAffected)

	return nil
}
