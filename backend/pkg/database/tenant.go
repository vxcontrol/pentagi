package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/crc32"
	"net/url"

	"pentagi/pkg/config"

	"github.com/lib/pq"
)

// requiredExtensions must live in a schema shared by every tenant. A PostgreSQL
// extension is a database-level object installed into exactly one schema: if
// tenant A created "vector" inside its own schema, tenant B's
// CREATE EXTENSION IF NOT EXISTS would be a silent no-op and its migrations
// would then fail with an opaque "type vector does not exist".
var requiredExtensions = []string{"vector", "pg_trgm"}

// EnsureTenantSchema prepares this instance's PostgreSQL namespace and rewrites
// cfg.DatabaseURL so that every downstream consumer — sqlc, gorm, goose and the
// pgxpool backing the langchaingo vector store — resolves unqualified
// identifiers inside it.
//
// It is a strict no-op when TENANT_ID is empty: the DSN is left untouched and
// everything keeps resolving through the default "public" search path exactly
// as before.
func EnsureTenantSchema(ctx context.Context, cfg *config.Config) error {
	if !cfg.HasTenant() {
		return nil
	}

	schema := cfg.SchemaName()
	extSchema := cfg.ExtensionSchema()

	// Short-lived bootstrap connection on the ORIGINAL DSN. Opening it before the
	// search_path rewrite means CREATE EXTENSION resolves against the default
	// path and lands in the shared schema rather than the tenant's.
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open bootstrap database connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to reach database for tenant bootstrap: %w", err)
	}

	// Serialize concurrent first boots so two instances cannot race on schema and
	// extension creation in the shared catalog.
	if err := WithAdvisoryLock(ctx, db, "pentagi-tenant-bootstrap", func(conn *sql.Conn) error {
		// QuoteIdentifier is belt-and-braces: ValidateTenantID already restricts
		// the character set, but this keeps the statement safe if that ever relaxes.
		if _, err := conn.ExecContext(ctx,
			"CREATE SCHEMA IF NOT EXISTS "+pq.QuoteIdentifier(schema),
		); err != nil {
			return fmt.Errorf("failed to create schema %q: %w", schema, err)
		}

		for _, ext := range requiredExtensions {
			if err := ensureSharedExtension(ctx, conn, ext, extSchema); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// Rewrite the DSN once; every consumer reads cfg.DatabaseURL afterwards.
	return RewriteDatabaseURLForTenant(cfg)
}

// RewriteDatabaseURLForTenant appends the tenant search_path to cfg.DatabaseURL.
// Unlike EnsureTenantSchema it does not touch the catalog — use it for tools that
// only need to read or write tenant data (e.g. the installer's password reset).
// It is a no-op when TENANT_ID is empty.
func RewriteDatabaseURLForTenant(cfg *config.Config) error {
	if !cfg.HasTenant() {
		return nil
	}

	rewritten, err := withSearchPath(
		cfg.DatabaseURL,
		cfg.SchemaName()+","+cfg.ExtensionSchema(),
		cfg.DatabaseSearchPathViaOptions,
	)
	if err != nil {
		return err
	}
	cfg.DatabaseURL = rewritten
	return nil
}

// ensureSharedExtension guarantees that ext exists in sharedSchema and is
// therefore reachable from every tenant's search_path. It checks before
// creating so that a database whose extensions were pre-installed by an
// administrator (or by convention — see DATABASE_EXTENSIONS_SCHEMA in
// backend/docs/config.md) works without the application needing CREATE
// privileges.
func ensureSharedExtension(ctx context.Context, conn *sql.Conn, ext, sharedSchema string) error {
	schema, err := extensionSchema(ctx, conn, ext)
	switch {
	case err != nil:
		return err

	case schema == "":
		// Not installed yet — create it explicitly in the shared schema.
		if _, err := conn.ExecContext(ctx, fmt.Sprintf(
			"CREATE EXTENSION IF NOT EXISTS %s SCHEMA %s",
			pq.QuoteIdentifier(ext), pq.QuoteIdentifier(sharedSchema),
		)); err != nil {
			return fmt.Errorf(
				"failed to create extension %q in schema %q (a privileged user must run "+
					"CREATE EXTENSION %s SCHEMA %s once): %w",
				ext, sharedSchema, ext, sharedSchema, err,
			)
		}
		return nil

	case schema != sharedSchema:
		// Installed, but somewhere this tenant's search_path will not reach. Fail
		// with an actionable message rather than letting migrations die on a
		// confusing "type does not exist".
		return fmt.Errorf(
			"extension %q is installed in schema %q, but multi-tenant mode requires it in %q "+
				"so every tenant can reach it; either run ALTER EXTENSION %s SET SCHEMA %s, "+
				"or set DATABASE_EXTENSIONS_SCHEMA=%s to match where it already lives",
			ext, schema, sharedSchema, ext, sharedSchema, schema,
		)

	default:
		return nil
	}
}

// extensionSchema returns the schema an extension is installed into, or "" when
// it is not installed.
func extensionSchema(ctx context.Context, conn *sql.Conn, ext string) (string, error) {
	var schema string
	err := conn.QueryRowContext(ctx, `
		SELECT n.nspname
		  FROM pg_extension e
		  JOIN pg_namespace n ON n.oid = e.extnamespace
		 WHERE e.extname = $1`, ext).Scan(&schema)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("failed to resolve schema of extension %q: %w", ext, err)
	default:
		return schema, nil
	}
}

// VerifySearchPath asserts that connections really do resolve into the expected
// schema. A typo in the DSN would otherwise route a tenant silently onto public,
// where every tenant would share one dataset — a quiet, catastrophic failure.
func VerifySearchPath(ctx context.Context, db *sql.DB, cfg *config.Config) error {
	if !cfg.HasTenant() {
		return nil
	}

	var current sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT current_schema()").Scan(&current); err != nil {
		return fmt.Errorf("failed to resolve current schema: %w", err)
	}
	if current.String != cfg.SchemaName() {
		return fmt.Errorf(
			"search_path resolved to schema %q, expected %q — refusing to start so tenants "+
				"do not silently share one dataset",
			current.String, cfg.SchemaName(),
		)
	}

	return nil
}

// RunMigrations applies pending migrations while holding an advisory lock, so
// that two instances booting simultaneously cannot execute the same migration
// set concurrently. Without a tenant the lock is still taken, which also fixes
// the pre-existing race between two single-instance deployments sharing a
// database.
func RunMigrations(ctx context.Context, db *sql.DB, cfg *config.Config, up func(*sql.DB) error) error {
	return WithAdvisoryLock(ctx, db, "pentagi-migrations-"+cfg.SchemaName(), func(*sql.Conn) error {
		return up(db)
	})
}

// WithAdvisoryLock runs fn while holding a PostgreSQL session-level advisory
// lock derived from key. The lock is taken on a dedicated connection because
// advisory locks are session-scoped and *sql.DB is a pool.
func WithAdvisoryLock(ctx context.Context, db *sql.DB, key string, fn func(*sql.Conn) error) error {
	// crc32 into the signed 32-bit space keeps the key stable and collision-free
	// enough for the two distinct locks this application takes.
	lockID := int64(int32(crc32.ChecksumIEEE([]byte(key))))

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire database connection for advisory lock: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockID); err != nil {
		return fmt.Errorf("failed to acquire advisory lock %q: %w", key, err)
	}
	defer func() {
		// Best effort: closing the connection releases the lock regardless.
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lockID)
	}()

	return fn(conn)
}

// withSearchPath returns dsn with the tenant's search_path applied as a
// PostgreSQL startup parameter (or, with viaOptions, wrapped as
// options=--search_path=<value> for poolers that need it — see
// DATABASE_SEARCH_PATH_VIA_OPTIONS in backend/docs/config.md). Supports both
// URL-style DSNs and libpq keyword strings.
func withSearchPath(dsn, searchPath string, viaOptions bool) (string, error) {
	key, value := "search_path", searchPath
	if viaOptions {
		key, value = "options", "--search_path="+searchPath
	}

	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		// Not a URL — fall back to libpq keyword/value syntax.
		return fmt.Sprintf("%s %s=%s", dsn, key, value), nil
	}

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
