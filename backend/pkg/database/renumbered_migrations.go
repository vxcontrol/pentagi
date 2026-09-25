package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var renumberedMigrations = map[int64]int64{
	20260823: 20260717,
	20260910: 20260718,
	20260911: 20260719,
	20260915: 20260720,
}

const lastSharedMigration = 20260716

func renumberPlan(versions []int64) map[int64]int64 {
	plan := make(map[int64]int64)
	for _, version := range versions {
		if version <= lastSharedMigration {
			continue
		}
		renumbered, ok := renumberedMigrations[version]
		if !ok {
			return nil
		}
		plan[version] = renumbered
	}

	if len(plan) == 0 {
		return nil
	}
	return plan
}

func renumberMigrations(ctx context.Context, db *sql.DB, schema string) error {
	table := pq.QuoteIdentifier(schema) + ".goose_db_version"

	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, table).Scan(&exists); err != nil {
		return fmt.Errorf("failed to look up %s: %w", table, err)
	}
	if !exists {
		return nil
	}

	rows, err := db.QueryContext(ctx, `SELECT DISTINCT version_id FROM `+table)
	if err != nil {
		return fmt.Errorf("failed to read recorded migration versions: %w", err)
	}
	var versions []int64
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return fmt.Errorf("failed to read a recorded migration version: %w", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("failed to read recorded migration versions: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read recorded migration versions: %w", err)
	}

	plan := renumberPlan(versions)
	if plan == nil {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin renumbering migration versions: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for old, renumbered := range plan {
		if _, err := tx.ExecContext(ctx,
			`UPDATE `+table+` SET version_id = $1 WHERE version_id = $2`, renumbered, old); err != nil {
			return fmt.Errorf("failed to renumber migration version %d: %w", old, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit renumbered migration versions: %w", err)
	}

	logrus.WithField("versions", plan).Info("moved migration versions recorded under the earlier numbering")
	return nil
}
