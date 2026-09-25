//go:build postgres

package database_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"pentagi/migrations"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// openSchema fails rather than skips without a server: a skip would read as coverage the shipped SQL does not have.
func openSchema(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("PENTAGI_TEST_DSN")
	if dsn == "" {
		t.Fatal("PENTAGI_TEST_DSN is unset; this tier needs a postgres to prove anything")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// Registered first so it runs last, after the schema drop below.
	t.Cleanup(func() { db.Close() })
	// search_path is session state, and a pool could hand a later query to a connection that never received it.
	db.SetMaxOpenConns(1)

	schema := fmt.Sprintf("stats_by_day_%d", time.Now().UnixNano())
	if _, err := db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	// A schema left behind keeps its own pg_trgm, and every later run then fails on gin_trgm_ops.
	t.Cleanup(func() {
		if _, err := db.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop schema %s: %v", schema, err)
		}
	})

	if _, err := db.Exec("SET search_path TO " + schema + ", public"); err != nil {
		t.Fatalf("search_path: %v", err)
	}

	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName(schema + ".goose_db_version")
	if err := goose.Up(db, "sql"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var landed int
	err = db.QueryRow(`
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema = $1 AND table_name IN ('flows', 'users')`, schema).Scan(&landed)
	if err != nil {
		t.Fatalf("locate migrated tables: %v", err)
	}
	if landed != 2 {
		t.Fatalf("migrations did not land in %s: %d of 2 tables found there", schema, landed)
	}

	return db
}

func seedFlow(t *testing.T, db *sql.DB, userID int64, createdAt time.Time) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(`
		INSERT INTO flows (title, model, model_provider_name, language, user_id, model_provider_type, created_at)
		VALUES ('boundary', 'm', 'p', 'en', $1, 'openai', $2)
		RETURNING id`, userID, createdAt).Scan(&id)
	if err != nil {
		t.Fatalf("seed flow at %s: %v", createdAt, err)
	}
	return id
}

func seedUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(`
		INSERT INTO users (hash, mail, name, status, role_id, type, provider)
		VALUES ('h', 'boundary@example.test', 'boundary', 'active', 1, 'local', 'local')
		RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}
