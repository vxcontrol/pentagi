//go:build postgres

package rdb

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type tableTitledRow struct {
	ID    uint64
	Title string
}

// tableOpenPostgres fails rather than skips without a server: Init maps text fields to LOWER(column::text), which sqlite cannot parse.
func tableOpenPostgres(t *testing.T) (*gorm.DB, string) {
	t.Helper()

	dsn := os.Getenv("PENTAGI_TEST_DSN")
	if dsn == "" {
		t.Fatal("PENTAGI_TEST_DSN is unset; this tier needs a postgres to prove anything")
	}

	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	table := fmt.Sprintf("table_titled_rows_%d", time.Now().UnixNano())
	for _, stmt := range []string{
		"CREATE TABLE " + table + " (id BIGINT PRIMARY KEY, title TEXT NOT NULL)",
		"INSERT INTO " + table + " (id, title) VALUES (1, 'Recon Example'), (2, 'Other')",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	t.Cleanup(func() {
		if err := db.Exec("DROP TABLE " + table).Error; err != nil {
			t.Errorf("drop %s: %v", table, err)
		}
	})

	return db, table
}

func TestTable_Query_ComparesTextFieldsWhateverTheCaseOfTheValue(t *testing.T) {
	db, table := tableOpenPostgres(t)

	tests := []struct {
		name     string
		operator string
		value    any
		want     []tableTitledRow
	}{
		{name: "equal to the stored title", operator: "=", value: "Recon Example", want: []tableTitledRow{{ID: 1, Title: "Recon Example"}}},
		{name: "not equal to the stored title", operator: "!=", value: "Recon Example", want: []tableTitledRow{{ID: 2, Title: "Other"}}},
		{name: "like part of the stored title", operator: "like", value: "Recon", want: []tableTitledRow{{ID: 1, Title: "Recon Example"}}},
		{name: "in a list with the stored title", operator: "in", value: []any{"Recon Example"}, want: []tableTitledRow{{ID: 1, Title: "Recon Example"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := TableQuery{
				Page:    1,
				Size:    5,
				Type:    "filter",
				Filters: []TableFilter{{Field: "title", Operator: tt.operator, Value: tt.value}},
			}
			if err := query.Init(table, map[string]any{"title": "{{table}}.title"}); err != nil {
				t.Fatalf("Init: %v", err)
			}

			var rows []tableTitledRow
			total, err := query.Query(db, &rows)
			if err != nil {
				t.Fatalf("Query: %v", err)
			}

			if total != uint64(len(tt.want)) || !reflect.DeepEqual(rows, tt.want) {
				t.Errorf("title %s %v counted %d and read %v, want %v", tt.operator, tt.value, total, rows, tt.want)
			}
		})
	}
}
