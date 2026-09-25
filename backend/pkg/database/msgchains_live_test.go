//go:build postgres

package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"pentagi/pkg/database"
)

func seedMailedUser(t *testing.T, db *sql.DB, mail string) int64 {
	t.Helper()

	var id int64
	if err := db.QueryRow(`
		INSERT INTO users (hash, mail, name, status, role_id, type, provider)
		VALUES ($1, $1, $1, 'active', 1, 'local', 'local')
		RETURNING id`, mail).Scan(&id); err != nil {
		t.Fatalf("seed user %s: %v", mail, err)
	}
	return id
}

// GetUserTotalToolcallsStats is held to the same attribution from the same seed.
func TestMsgchains_GetUserTotalUsageStats_CountsARowOnlyForTheOwnerOfItsFlow(t *testing.T) {
	db := openSchema(t)
	ctx := context.Background()
	q := database.New(db)

	owner := seedMailedUser(t, db, "owner@example.test")
	other := seedMailedUser(t, db, "other@example.test")
	ownFlow := seedFlow(t, db, owner, time.Now())
	otherFlow := seedFlow(t, db, other, time.Now())

	var otherTask int64
	if err := db.QueryRow(`INSERT INTO tasks (input, flow_id) VALUES ('boundary', $1) RETURNING id`,
		otherFlow).Scan(&otherTask); err != nil {
		t.Fatalf("seed task: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO msgchains (type, model, model_provider, usage_in, chain, flow_id, task_id)
		VALUES ('primary_agent', 'm', 'p', 1000, '[]', $1, $2)`, ownFlow, otherTask); err != nil {
		t.Fatalf("seed msgchain: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO toolcalls (call_id, status, name, args, flow_id, task_id, duration_seconds)
		VALUES ('call-1', 'finished', 'terminal', '{}', $1, $2, 5)`, ownFlow, otherTask); err != nil {
		t.Fatalf("seed toolcall: %v", err)
	}

	for _, tc := range []struct {
		user      int64
		usage     int64
		toolcalls int64
		whose     string
	}{
		{owner, 1000, 1, "the owner"},
		{other, 0, 0, "the user whose task the row names"},
	} {
		usage, err := q.GetUserTotalUsageStats(ctx, tc.user)
		if err != nil {
			t.Fatalf("usage of %s: %v", tc.whose, err)
		}
		if usage.TotalUsageIn != tc.usage {
			t.Errorf("usage of %s is %d, want %d", tc.whose, usage.TotalUsageIn, tc.usage)
		}

		calls, err := q.GetUserTotalToolcallsStats(ctx, tc.user)
		if err != nil {
			t.Fatalf("toolcalls of %s: %v", tc.whose, err)
		}
		if calls.TotalCount != tc.toolcalls {
			t.Errorf("toolcalls of %s count %d, want %d", tc.whose, calls.TotalCount, tc.toolcalls)
		}
	}
}
