//go:build postgres

package timezone

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	_ "time/tzdata"

	_ "github.com/lib/pq"
)

func catalogOpenDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("PENTAGI_TEST_DSN")
	if dsn == "" {
		t.Fatal("PENTAGI_TEST_DSN is unset; the catalogue is the server's, not ours")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// sql.Open connects to nothing and Accepts answers true on a failed query: unpinged, the tier passes against a dead address.
	if err := db.Ping(); err != nil {
		t.Fatalf("ping %s: %v", dsn, err)
	}

	return db
}

func catalogNarrow(t *testing.T, db *sql.DB, keep ...string) {
	t.Helper()

	db.SetMaxOpenConns(1)

	schema := fmt.Sprintf("tz_probe_%d", os.Getpid())
	t.Cleanup(func() {
		if _, err := db.Exec("SET search_path = public"); err != nil {
			t.Errorf("restore search_path: %v", err)
		}
		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema)); err != nil {
			t.Errorf("drop %s: %v", schema, err)
		}
	})

	if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema)); err != nil {
		t.Fatalf("drop stale %s: %v", schema, err)
	}
	if _, err := db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schema)); err != nil {
		t.Fatalf("create %s: %v", schema, err)
	}

	list := ""
	for i, name := range keep {
		if i > 0 {
			list += ", "
		}
		list += fmt.Sprintf("'%s'", name)
	}
	view := fmt.Sprintf(
		"CREATE VIEW %s.pg_timezone_names AS SELECT * FROM pg_catalog.pg_timezone_names WHERE name IN (%s)",
		schema, list,
	)
	if _, err := db.Exec(view); err != nil {
		t.Fatalf("create view: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SET search_path = %s, pg_catalog, public", schema)); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	var visible int
	if err := db.QueryRow("SELECT count(*) FROM pg_timezone_names").Scan(&visible); err != nil {
		t.Fatalf("count narrowed catalogue: %v", err)
	}
	if visible != len(keep) {
		t.Fatalf("narrowed catalogue holds %d names, want %d — the view is not in front", visible, len(keep))
	}
}

func TestCatalog_ResolveFor_FallsBackForAZoneTheServerDoesNotKnow(t *testing.T) {
	db := catalogOpenDB(t)
	catalogNarrow(t, db, "UTC", "Asia/Riyadh")

	catalog := NewCatalog(db)
	name := func(s string) *string { return &s }

	tests := []struct {
		name string
		in   *string
		want string
	}{
		{name: "a zone this server carries", in: name("Asia/Riyadh"), want: "Asia/Riyadh"},
		{name: "a zone Go carries and this server does not", in: name("Europe/Berlin"), want: "UTC"},
		{name: "UTC stays UTC", in: name("UTC"), want: "UTC"},
		{name: "an absent zone means UTC", in: nil, want: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := catalog.ResolveFor(tt.in)
			if err != nil {
				t.Fatalf("ResolveFor: %v", err)
			}
			if got != tt.want {
				t.Errorf("ResolveFor = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCatalog_ResolveFor_StillRefusesWhatGoRefuses(t *testing.T) {
	catalog := NewCatalog(catalogOpenDB(t))

	for _, bad := range []string{"Local", "Middle-earth/Shire", "UTC'; DROP TABLE flows; --"} {
		if _, err := catalog.ResolveFor(&bad); err == nil {
			t.Errorf("ResolveFor(%q) returned no error", bad)
		}
	}
}
