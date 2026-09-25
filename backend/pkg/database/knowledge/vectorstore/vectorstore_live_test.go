//go:build postgres

package vectorstore_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"pentagi/migrations"
	"pentagi/pkg/database/knowledge/vectorstore"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"

	_ "github.com/lib/pq"
)

type zeroEmbedder struct{}

func (zeroEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	return make([][]float32, len(texts)), nil
}

func (zeroEmbedder) EmbedQuery(context.Context, string) ([]float32, error) { return nil, nil }

func TestVectorstore_Options_LeaveTheEmbeddingIndexesToTheMigrations(t *testing.T) {
	dsn := os.Getenv("PENTAGI_TEST_DSN")
	if dsn == "" {
		t.Fatal("PENTAGI_TEST_DSN is unset; this tier needs a postgres to prove anything")
	}
	ctx := context.Background()

	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := fmt.Sprintf("vectorstore_%d", time.Now().UnixNano())
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
	if _, err := db.Exec(`DROP INDEX langchain_pg_embedding_meta_doc_type_flow_id`); err != nil {
		t.Fatalf("drop the metadata index: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("pool config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := pgvector.New(ctx, vectorstore.Options(zeroEmbedder{}, pool, "")...); err != nil {
		t.Fatalf("open the store: %v", err)
	}

	var rebuilt int
	if err := db.QueryRow(`SELECT count(*) FROM pg_indexes WHERE schemaname = $1 AND indexname LIKE '%\_meta\_%'`,
		schema).Scan(&rebuilt); err != nil {
		t.Fatalf("count metadata indexes: %v", err)
	}
	if rebuilt != 0 {
		t.Errorf("opening a store built %d metadata index(es) the migrations own", rebuilt)
	}
}
