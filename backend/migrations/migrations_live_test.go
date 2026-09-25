//go:build postgres

package migrations

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	metadataIndex = "langchain_pg_embedding_meta_doc_type_flow_id"
	partialIndex  = "langchain_pg_embedding_meta_doc_type_partial"
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

	schema := fmt.Sprintf("migrations_%d", time.Now().UnixNano())
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

	goose.SetBaseFS(EmbedMigrations)
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

func versionBelow(t *testing.T, name string) int64 {
	t.Helper()

	all, err := goose.CollectMigrations("sql", 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	for i, migration := range all {
		if strings.HasSuffix(migration.Source, "_"+name+".sql") {
			if i == 0 {
				t.Fatalf("%s is the first migration, nothing lies below it", name)
			}
			return all[i-1].Version
		}
	}
	t.Fatalf("no migration is named %s", name)
	return 0
}

func thoughtSignedCallID(size int) string {
	rng := rand.NewChaCha8([32]byte{3, 4, 4})
	head := make([]byte, 12)
	_, _ = rng.Read(head)
	signature := make([]byte, size)
	_, _ = rng.Read(signature)

	id := "call_" + hex.EncodeToString(head) + "__thought__" + base64.StdEncoding.EncodeToString(signature)
	return id[:size]
}

func validIndex(t *testing.T, db *sql.DB, name string) (exists, valid bool) {
	t.Helper()

	err := db.QueryRow(`
		SELECT i.indisvalid FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = $1 AND n.nspname = current_schema()`, name).Scan(&valid)
	if err == sql.ErrNoRows {
		return false, false
	}
	if err != nil {
		t.Fatalf("look up index %s: %v", name, err)
	}
	return true, valid
}

func TestMigrations_EmbedMigrations_BuildTheEmbeddingMetadataIndex(t *testing.T) {
	db := openSchema(t)

	if exists, valid := validIndex(t, db, metadataIndex); !exists || !valid {
		t.Errorf("%s: exists=%v valid=%v, want a valid index", metadataIndex, exists, valid)
	}
	if exists, _ := validIndex(t, db, partialIndex); exists {
		t.Errorf("%s is built although no query can use it", partialIndex)
	}
}

func TestMigrations_EmbedMigrations_DropThePartialIndexAStoreBuiltBefore(t *testing.T) {
	db := openSchema(t)

	if err := goose.DownTo(db, "sql", versionBelow(t, "embedding_metadata_indexes")); err != nil {
		t.Fatalf("migrate down below the embedding indexes: %v", err)
	}
	if _, err := db.Exec(`CREATE INDEX ` + partialIndex + ` ON langchain_pg_embedding ((cmetadata ->> 'doc_type'))
		WHERE (cmetadata ->> 'doc_type') IS DISTINCT FROM 'memory'`); err != nil {
		t.Fatalf("build the partial index the way the store did: %v", err)
	}
	if err := goose.Up(db, "sql"); err != nil {
		t.Fatalf("migrate up again: %v", err)
	}

	if exists, _ := validIndex(t, db, partialIndex); exists {
		t.Errorf("%s survived the migrations", partialIndex)
	}
	if exists, valid := validIndex(t, db, metadataIndex); !exists || !valid {
		t.Errorf("%s: exists=%v valid=%v, want a valid index", metadataIndex, exists, valid)
	}
}

func TestMigrations_EmbedMigrations_KeepALongCallIDBelowTheHashIndex(t *testing.T) {
	db := openSchema(t)
	ctx := context.Background()
	flowID := seedFlow(t, db, seedUser(t, db), time.Now())

	callID := thoughtSignedCallID(5000)
	if _, err := database.New(db).CreateToolcall(ctx, database.CreateToolcallParams{
		CallID: callID,
		Status: database.ToolcallStatusRunning,
		Name:   "terminal",
		Args:   json.RawMessage(`{}`),
		FlowID: flowID,
	}); err != nil {
		t.Fatalf("create a toolcall with a %d-byte call id: %v", len(callID), err)
	}

	below := versionBelow(t, "toolcalls_call_id_hash_idx")
	if err := goose.DownTo(db, "sql", below); err != nil {
		t.Fatalf("migrate down below the call id hash index with a %d-byte call id stored: %v", len(callID), err)
	}
	if version, err := goose.GetDBVersion(db); err != nil || version != below {
		t.Fatalf("schema version after migrating down is %d (%v), want %d", version, err, below)
	}

	var stored int
	if err := db.QueryRow(`SELECT count(*) FROM toolcalls WHERE call_id = $1`, callID).Scan(&stored); err != nil {
		t.Fatalf("look up the call id after migrating down: %v", err)
	}
	if stored != 1 {
		t.Errorf("found %d toolcalls with the long call id after migrating down, want 1", stored)
	}
}
