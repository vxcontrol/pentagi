package rdb

import (
	"math"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestTable_PageOffset_StaysPastTheRowsAlreadySkipped(t *testing.T) {
	tests := []struct {
		name string
		page int
		size int
		want int
	}{
		{name: "first page", page: 1, size: 5, want: 0},
		{name: "second page", page: 2, size: 5, want: 5},
		{name: "no page size to divide by", page: 7, size: 0, want: 0},
		{name: "widest product that fits", page: 1844674407370955162, size: 5, want: 9223372036854775805},
		{name: "product wider than the counter", page: 1844674407370955163, size: 5, want: math.MaxInt},
		{name: "widest page of all", page: math.MaxInt, size: 5, want: math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pageOffset(tt.page, tt.size)

			if got < 0 {
				t.Fatalf("page %d of %d starts %d rows in, and a driver drops a negative offset instead of refusing it — the caller gets the first page", tt.page, tt.size, got)
			}

			if got != tt.want {
				t.Errorf("page %d of %d starts after %d rows, want %d", tt.page, tt.size, got, tt.want)
			}
		})
	}
}

func TestTable_Init_AppliesTheDeclaredPageSize(t *testing.T) {
	tests := []struct {
		name string
		size int
		want int
	}{
		{name: "absent", size: 0, want: 5},
		{name: "unlimited", size: -1, want: -1},
		{name: "asked for", size: 20, want: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := TableQuery{Page: 1, Size: tt.size, Type: "init"}

			if err := query.Init("flows", nil); err != nil {
				t.Fatalf("Init: %v", err)
			}

			if query.Size != tt.want {
				t.Errorf("page size %d became %d, want %d", tt.size, query.Size, tt.want)
			}
		})
	}

	field, ok := reflect.TypeFor[TableQuery]().FieldByName("Size")
	if !ok {
		t.Fatal("TableQuery has no Size field — the tag this check reads has moved")
	}

	if declared := field.Tag.Get("default"); declared != "5" {
		t.Errorf("the tag promises callers a page of %s and Swagger repeats it, while a request without pageSize gets 5", declared)
	}
}

type tableOwnRow struct {
	ID uint64
}

func (tableOwnRow) TableName() string { return "table_own_rows" }

func TestTable_Query_ReadsTheTableItWasGivenNotTheModelsOwn(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// every pooled connection to ":memory:" opens its own empty database
	db.DB().SetMaxOpenConns(1)

	for _, stmt := range []string{
		"CREATE TABLE table_own_rows (id INTEGER PRIMARY KEY)",
		"CREATE TABLE table_given_rows (id INTEGER PRIMARY KEY)",
		"INSERT INTO table_own_rows (id) VALUES (3), (4)",
		"INSERT INTO table_given_rows (id) VALUES (7)",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}

	query := TableQuery{Page: 1, Size: 5, Type: "init"}
	if err := query.Init("table_given_rows", nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var rows []tableOwnRow
	total, err := query.Query(db, &rows)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	if total != 1 || !reflect.DeepEqual(rows, []tableOwnRow{{ID: 7}}) {
		t.Errorf("Query counted %d and read %v, want the one row 7 of table_given_rows", total, rows)
	}
}
