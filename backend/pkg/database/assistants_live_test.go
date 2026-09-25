//go:build postgres

package database_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"pentagi/pkg/database"
)

func TestAssistants_UpdateAssistantStatus_LeavesADeletedAssistantAlone(t *testing.T) {
	db := openSchema(t)
	q := database.New(db)
	ctx := t.Context()

	flowID := seedFlow(t, db, seedUser(t, db), time.Now())
	assistant, err := q.CreateAssistant(ctx, database.CreateAssistantParams{
		Title:             "deleted while building",
		Status:            database.AssistantStatusCreated,
		Model:             "m",
		ModelProviderName: "p",
		ModelProviderType: database.ProviderTypeOpenai,
		Language:          "en",
		Functions:         json.RawMessage(`{}`),
		FlowID:            flowID,
	})
	if err != nil {
		t.Fatalf("create assistant: %v", err)
	}
	if _, err := q.DeleteAssistant(ctx, assistant.ID); err != nil {
		t.Fatalf("delete assistant: %v", err)
	}

	_, err = q.UpdateAssistantStatus(ctx, database.UpdateAssistantStatusParams{
		ID:     assistant.ID,
		Status: database.AssistantStatusFailed,
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("updating a deleted assistant must match no row, got %v", err)
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM assistants WHERE id = $1`, assistant.ID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != string(database.AssistantStatusCreated) {
		t.Fatalf("a deleted assistant's status changed to %q", status)
	}
}
