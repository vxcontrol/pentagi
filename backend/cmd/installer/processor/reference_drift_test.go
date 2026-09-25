package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A sharper tag of the shipped repository is what a working update pins, not drift.
func TestReferenceDrift_ReferenceDrift_ReportsOnlyADifferentRepository(t *testing.T) {
	defaults := map[string]string{
		"PENTAGI_IMAGE":  "vxcontrol/pentagi:latest",
		"POSTGRES_IMAGE": "postgres:16",
		"NEO4J_IMAGE":    "neo4j:5.26.2",
		"REDIS_IMAGE":    "redis:7",
	}

	drift := referenceDrift(defaults, map[string]string{
		"PENTAGI_IMAGE": "vxcontrol/pentagi:2.3.4",
		// Another major of the same repository is a data migration, not drift.
		"POSTGRES_IMAGE": "postgres:15",
		"NEO4J_IMAGE":    "acme/neo4j-fork:5.26.2",
		"REDIS_IMAGE":    "",
	})

	require.Len(t, drift, 1, "only a different repository is drift: %+v", drift)
	assert.Equal(t, "NEO4J_IMAGE", drift[0].Variable)
	assert.Equal(t, "acme/neo4j-fork:5.26.2", drift[0].Configured)
	assert.Equal(t, "neo4j:5.26.2", drift[0].Default)
}

func TestReferenceDrift_ComposeReferenceDefaults_ReadsTheShippedComposeFiles(t *testing.T) {
	root := repositoryRoot(t)
	defaults, err := composeReferenceDefaults(func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, name))
	})
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(defaults), 20, "every image is parameterised, so every one has a default")
	assert.Equal(t, "vxcontrol/pentagi:latest", defaults["PENTAGI_IMAGE"])
	assert.Equal(t, "postgres:16", defaults["POSTGRES_IMAGE"])
	// Two services ship the same image under different variables, so the variable is the key.
	assert.Equal(t, "clickhouse/clickhouse-server:24", defaults["CLICKHOUSE_IMAGE"])
	assert.Equal(t, "clickhouse/clickhouse-server:24", defaults["CLICKSTORE_IMAGE"])

	// Finding nothing would make every installation look perfectly aligned.
	_, err = composeReferenceDefaults(func(string) ([]byte, error) { return []byte("services:\n"), nil })
	assert.EqualError(t, err, "no parameterised images found in the compose files")
}
