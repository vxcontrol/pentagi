//go:build postgres

package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"pentagi/migrations"
	"pentagi/pkg/config"
	"pentagi/pkg/database"

	"github.com/pressly/goose/v3"
)

func openTenant(t *testing.T) (*sql.DB, string) {
	t.Helper()

	dsn := os.Getenv("PENTAGI_TEST_DSN")
	if dsn == "" {
		t.Fatal("PENTAGI_TEST_DSN is unset; this tier needs a postgres to prove anything")
	}

	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := fmt.Sprintf("renumbered_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	// A schema left behind keeps its own pg_trgm, and every later run then fails on gin_trgm_ops.
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop schema %s: %v", schema, err)
		}
	})

	tenantURL, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	query := tenantURL.Query()
	query.Set("options", "-c search_path="+schema+",public")
	tenantURL.RawQuery = query.Encode()

	db, err := sql.Open("postgres", tenantURL.String())
	if err != nil {
		t.Fatalf("open tenant: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	goose.SetTableName(schema + ".goose_db_version")
	if err := goose.Up(db, "sql"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db, schema
}

func recordEarlierNumbering(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`DELETE FROM goose_db_version WHERE version_id > 20260720`); err != nil {
		t.Fatalf("drop the migrations the earlier numbering did not have: %v", err)
	}
	if _, err := db.Exec(`
		UPDATE goose_db_version SET version_id = CASE version_id
			WHEN 20260717 THEN 20260823
			WHEN 20260718 THEN 20260910
			WHEN 20260719 THEN 20260911
			WHEN 20260720 THEN 20260915
		END
		WHERE version_id BETWEEN 20260717 AND 20260720`); err != nil {
		t.Fatalf("record the earlier numbering: %v", err)
	}
}

func countVersions(t *testing.T, db *sql.DB, versions string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(`SELECT count(*) FROM goose_db_version WHERE version_id IN (` + versions + `)`).Scan(&n); err != nil {
		t.Fatalf("count versions %s: %v", versions, err)
	}
	return n
}

const earlierVersions = "20260823, 20260910, 20260911, 20260915"

func migrateUp(db *sql.DB) error { return goose.Up(db, "sql") }

func TestRenumberedMigrations_RenumberMigrations_MovesADatabaseRecordedUnderTheEarlierNumbering(t *testing.T) {
	db, schema := openTenant(t)
	recordEarlierNumbering(t, db)

	if err := database.RunMigrations(context.Background(), db, &config.Config{TenantID: schema}, migrateUp); err != nil {
		t.Fatalf("migrations over the earlier numbering: %v", err)
	}

	all, err := goose.CollectMigrations("sql", 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	latest := all[len(all)-1].Version
	if version, err := goose.GetDBVersion(db); err != nil || version != latest {
		t.Errorf("schema version is %d (%v), want %d", version, err, latest)
	}
	if n := countVersions(t, db, earlierVersions); n != 0 {
		t.Errorf("%d versions of the earlier numbering are still recorded", n)
	}
}

func TestRenumberedMigrations_RenumberMigrations_LeavesADatabaseOfTheLineThatReusesTheNumbers(t *testing.T) {
	db, schema := openTenant(t)
	recordEarlierNumbering(t, db)
	if _, err := db.Exec(`INSERT INTO goose_db_version (version_id, is_applied) VALUES (20260801, true)`); err != nil {
		t.Fatalf("record a version of the other line: %v", err)
	}

	if err := database.RunMigrations(context.Background(), db, &config.Config{TenantID: schema}, migrateUp); err == nil {
		t.Error("migrations started over a database of the other line")
	}
	if n := countVersions(t, db, earlierVersions); n != 4 {
		t.Errorf("%d of the 4 recorded versions survived, want all of them untouched", n)
	}
}
