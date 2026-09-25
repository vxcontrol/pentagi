//go:build postgres

package database_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
)

const btreeRowLimit = 2704

func thoughtSignedCallID(size int) string {
	rng := rand.NewChaCha8([32]byte{3, 4, 4})
	head := make([]byte, 12)
	_, _ = rng.Read(head)
	signature := make([]byte, size)
	_, _ = rng.Read(signature)

	id := "call_" + hex.EncodeToString(head) + "__thought__" + base64.StdEncoding.EncodeToString(signature)
	return id[:size]
}

func TestToolcalls_GetCallToolcall_FindsACallIDLongerThanABtreeRow(t *testing.T) {
	db := openSchema(t)
	ctx := context.Background()
	flowID := seedFlow(t, db, seedUser(t, db), time.Now())

	callID := thoughtSignedCallID(5000)

	q := database.New(db)
	created, err := q.CreateToolcall(ctx, database.CreateToolcallParams{
		CallID: callID,
		Status: database.ToolcallStatusRunning,
		Name:   "terminal",
		Args:   json.RawMessage(`{}`),
		FlowID: flowID,
	})
	if err != nil {
		t.Fatalf("create a toolcall with a %d-byte call id: %v", len(callID), err)
	}

	var stored int
	if err := db.QueryRow(`SELECT pg_column_size(call_id) FROM toolcalls WHERE id = $1`, created.ID).Scan(&stored); err != nil {
		t.Fatalf("measure the stored call id: %v", err)
	}
	if stored <= btreeRowLimit {
		t.Fatalf("the call id compresses to %d bytes, under the %d-byte btree row limit, so it proves nothing about long ids",
			stored, btreeRowLimit)
	}

	found, err := q.GetCallToolcall(ctx, callID)
	if err != nil {
		t.Fatalf("find the toolcall by its %d-byte call id: %v", len(callID), err)
	}
	if found.ID != created.ID {
		t.Errorf("lookup by call id returned toolcall %d, want %d", found.ID, created.ID)
	}
	if found.CallID != callID {
		t.Errorf("the stored call id is %d bytes and differs from the %d bytes sent", len(found.CallID), len(callID))
	}

	if _, err := db.Exec(`SET enable_seqscan = off`); err != nil {
		t.Fatalf("disable seqscan: %v", err)
	}
	rows, err := db.Query(`EXPLAIN SELECT id FROM toolcalls WHERE call_id = $1`, callID)
	if err != nil {
		t.Fatalf("explain the lookup: %v", err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("read the plan: %v", err)
		}
		plan.WriteString(line + "\n")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read the plan: %v", err)
	}
	if !strings.Contains(plan.String(), "toolcalls_call_id_idx") {
		t.Errorf("the lookup by call id no longer reaches toolcalls_call_id_idx:\n%s", plan.String())
	}
}
