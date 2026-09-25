package processor

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullReference_ComposeImageVar_UppercasesTheComponentAndTurnsDashesIntoUnderscores(t *testing.T) {
	for component, want := range map[string]string{
		"pentagi":         "PENTAGI_IMAGE",
		"langfuse-worker": "LANGFUSE_WORKER_IMAGE",
		"node-exporter":   "NODE_EXPORTER_IMAGE",
		"clickstore":      "CLICKSTORE_IMAGE",
		"otel":            "OTEL_IMAGE",
	} {
		assert.Equal(t, want, composeImageVar(component))
	}
}

// An empty value would erase a reference a previous answer pinned.
func TestPullReference_PullReferencesForStack_KeepsOnlyResolvedComponentsOfThatStack(t *testing.T) {
	updates := []checker.StackUpdate{
		{
			Stack: "pentagi",
			Components: []checker.ComponentUpdate{
				{Component: "pentagi", Action: "upgrade", PullReference: "vxcontrol/pentagi:2.3.4"},
				// Written too, or the next `compose up` falls back to the file's default.
				{Component: "pgvector", Action: "current", PullReference: "vxcontrol/pgvector:latest"},
				{Component: "scraper", Action: "unknown", Reason: "tag_not_published"},
			},
		},
		{
			Stack: "langfuse",
			Components: []checker.ComponentUpdate{
				{Component: "langfuse-web", Action: "current", PullReference: "langfuse/langfuse:3"},
			},
		},
	}

	assert.Equal(t, map[string]string{
		"PENTAGI_IMAGE":  "vxcontrol/pentagi:2.3.4",
		"PGVECTOR_IMAGE": "vxcontrol/pgvector:latest",
	}, pullReferencesForStack(updates, "pentagi"),
		"an unresolved component must not be written, and another stack must not leak in")
	assert.Equal(t, map[string]string{"LANGFUSE_WEB_IMAGE": "langfuse/langfuse:3"},
		pullReferencesForStack(updates, "langfuse"))
	assert.Empty(t, pullReferencesForStack(updates, "observability"))
}

// `docker compose --env-file` reads the file, not the installer's staging area.
func TestPullReference_ApplyPullReferences_WritesTheEnvFileWithoutMakingTheStateDirty(t *testing.T) {
	p, state := processorWithPullReference(t)

	require.NoError(t, p.applyPullReferences(ProductStackPentagi))

	assert.Equal(t, map[string]string{"PENTAGI_IMAGE": "vxcontrol/pentagi:2.3.4"}, state.written,
		"the reference was staged instead of written")
	assert.False(t, state.IsDirty())
}

// A hardcoded `image:` cannot be steered by the server's answer; this walks the real files.
func TestPullReference_EveryComposeImageIsParameterised(t *testing.T) {
	root := repositoryRoot(t)
	imageLine := regexp.MustCompile(`(?m)^\s*image:\s*(\S+)\s*$`)
	parameterised := regexp.MustCompile(`^\$\{([A-Z0-9_]+):-(\S+)\}$`)

	seen := 0
	for _, name := range []string{
		"docker-compose.yml",
		"docker-compose-langfuse.yml",
		"docker-compose-observability.yml",
		"docker-compose-graphiti.yml",
	} {
		body, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err, "the compose files are part of the contract with the server")

		for _, match := range imageLine.FindAllStringSubmatch(string(body), -1) {
			seen++
			reference := match[1]
			parts := parameterised.FindStringSubmatch(reference)
			require.NotNil(t, parts,
				"%s: %q is hardcoded, so the server cannot steer what this service pulls", name, reference)

			// `:-` and not `-`: only the first substitutes the default for an EMPTY variable.
			assert.True(t, len(parts) == 3 && parts[2] != "",
				"%s: %s must keep a default so an installation with no answer behaves as before",
				name, parts[1])
			assert.True(t, len(parts[1]) > len("_IMAGE") && parts[1][len(parts[1])-len("_IMAGE"):] == "_IMAGE",
				"%s: %q does not follow <COMPONENT>_IMAGE, so composeImageVar will never write it",
				name, parts[1])
		}
	}
	assert.GreaterOrEqual(t, seen, 20, "the walk found only %d images; it is not reading the files", seen)
}
