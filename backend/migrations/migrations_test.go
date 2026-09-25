package migrations

import (
	"strings"
	"testing"
)

// goose keys a migration on the digits before the first underscore; a clash panics the server at startup.
func TestMigrations_EmbedMigrations_VersionsAreUniqueAcrossFileNames(t *testing.T) {
	entries, err := EmbedMigrations.ReadDir("sql")
	if err != nil {
		t.Fatalf("read sql dir: %v", err)
	}

	seen := make(map[string]string, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		version, _, found := strings.Cut(name, "_")
		if !found {
			t.Errorf("%s has no version prefix", name)
			continue
		}
		if previous, clash := seen[version]; clash {
			t.Errorf("%s and %s share goose version %s", previous, name, version)
			continue
		}
		seen[version] = name
	}

	if len(seen) == 0 {
		t.Fatal("no migrations found, so this proves nothing")
	}
}
