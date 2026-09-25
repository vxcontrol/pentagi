package timezone

import (
	"database/sql"
	"sync"
)

// Catalog answers whether Postgres will accept a zone name in AT TIME ZONE.
//
// Go carries its own copy of tzdata and Postgres carries another, and the two
// drift: a zone IANA split off recently resolves in Go and fails in the server,
// which turns a browser's own zone into a failed query. Asking the server is the
// only way to know, and pg_timezone_names costs about 20ms to scan, so each
// answer is kept.
type Catalog struct {
	db    *sql.DB
	known sync.Map
}

func NewCatalog(db *sql.DB) *Catalog {
	return &Catalog{db: db}
}

// Accepts reports whether the server knows the name. An unreachable server
// answers true, leaving the query itself to fail rather than silently moving
// the day boundaries.
func (c *Catalog) Accepts(name string) bool {
	if c == nil || c.db == nil {
		return true
	}
	if cached, ok := c.known.Load(name); ok {
		return cached.(bool)
	}

	var exists bool
	err := c.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_timezone_names WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		return true
	}

	c.known.Store(name, exists)
	return exists
}

// ResolveFor is Resolve plus the server's opinion: a name Go accepts and the
// server does not falls back to UTC, which is the bucketing those callers had
// before the zone argument existed. Refusing instead would leave them with no
// chart at all, and no zone the server knows can be substituted for theirs.
func (c *Catalog) ResolveFor(name *string) (string, error) {
	resolved, err := Resolve(name)
	if err != nil {
		return "", err
	}
	if !c.Accepts(resolved) {
		return "UTC", nil
	}
	return resolved, nil
}
