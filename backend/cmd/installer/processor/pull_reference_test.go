package processor

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTheComposeVariableIsDerivedFromTheComponentName.
//
// The compose files spell it `${<COMPONENT>_IMAGE:-<default>}` with the
// component name uppercased and dashes turned into underscores. Deriving it here
// rather than keeping a table means a new component needs no second
// registration — and the pairing is checked against the real files below, so a
// drift is a failing test rather than a pull that silently used the default.
func TestTheComposeVariableIsDerivedFromTheComponentName(t *testing.T) {
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

// TestOnlyResolvedComponentsGetAReference.
//
// A component the server could not resolve carries no reference. Writing an
// empty value for it would erase whatever a previous answer had pinned —
// `${VAR:-default}` treats empty as absent, so the pull would quietly fall back
// to the shipped default and the installation would drift off the curated set
// without anything saying so.
func TestOnlyResolvedComponentsGetAReference(t *testing.T) {
	updates := []checker.StackUpdate{
		{
			Stack: "pentagi",
			Components: []checker.ComponentUpdate{
				{Component: "pentagi", Action: "upgrade", PullReference: "vxcontrol/pentagi:2.3.4"},
				{Component: "pgvector", Action: "current", PullReference: "vxcontrol/pgvector:latest"},
				// Nothing published for it: no reference, and none must be written.
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

	vars := pullReferencesForStack(updates, "pentagi")
	assert.Equal(t, map[string]string{
		"PENTAGI_IMAGE":  "vxcontrol/pentagi:2.3.4",
		"PGVECTOR_IMAGE": "vxcontrol/pgvector:latest",
	}, vars, "an unresolved component must not be written, and another stack must not leak in")

	// `current` components are written too, and deliberately: the variable has to
	// keep naming the artefact the installation is on, or the next `compose up`
	// resolves it back to the file's default.
	assert.Contains(t, vars, "PGVECTOR_IMAGE")

	assert.Equal(t, map[string]string{"LANGFUSE_WEB_IMAGE": "langfuse/langfuse:3"},
		pullReferencesForStack(updates, "langfuse"))
	assert.Empty(t, pullReferencesForStack(updates, "observability"))
}

// TestTheReferencesReachTheFileComposeActuallyReads.
//
// This is the defect that made the whole mechanism inert, and it was invisible: the code
// called SetVars, which STAGES a change into the installer's own state file, to be written
// into `.env` only when the user applies their changes. But the pull that follows is
// `docker compose --env-file <.env>`, so a staged variable is one compose has never heard
// of — `${PENTAGI_IMAGE:-vxcontrol/pentagi:latest}` fell back to the default, the pull
// fetched something the update server never offered, and the update reported success.
//
// Two things are asserted and both matter. The references have to reach the file, and they
// must NOT make the installation dirty: a value the installer wrote for itself would
// otherwise prompt the user to apply changes they never made.
func TestTheReferencesReachTheFileComposeActuallyReads(t *testing.T) {
	state := newMockState()
	p := &processor{
		state: state,
		checker: &checker.CheckResult{
			StackUpdates: []checker.StackUpdate{{
				Stack: "pentagi",
				Components: []checker.ComponentUpdate{
					{Component: "pentagi", Action: "upgrade", PullReference: "vxcontrol/pentagi:2.3.4"},
				},
			}},
		},
	}

	require.NoError(t, p.applyPullReferences(ProductStackPentagi))

	assert.Equal(t, map[string]string{"PENTAGI_IMAGE": "vxcontrol/pentagi:2.3.4"}, state.written,
		"the reference was staged instead of written; docker compose reads the file, not the staging area")
	assert.False(t, state.IsDirty(),
		"a reference the installer wrote for itself must not ask the user to apply changes they never made")
}

// TestEveryFetchingOperationWritesTheReferencesFirst.
//
// The references used to be written by ONE call site — the update path — while four others
// reached a pull without them: download, install, and the three apply-changes phases. All
// five end in updateStack or downloadStack, so that is where the write belongs; a call site
// that forgets is not a compile error, it is a pull of the compose default in silence.
//
// The tear-down operations are the other half of the claim: `down` needs no reference, and
// writing one while removing a stack would be a side effect nobody asked for.
func TestEveryFetchingOperationWritesTheReferencesFirst(t *testing.T) {
	newProcessor := func() (*processor, *mockState) {
		state := newMockState()
		p := &processor{
			state: state,
			checker: &checker.CheckResult{
				StackUpdates: []checker.StackUpdate{{
					Stack: "pentagi",
					Components: []checker.ComponentUpdate{
						{Component: "pentagi", Action: "upgrade", PullReference: "vxcontrol/pentagi:2.3.4"},
					},
				}},
			},
		}
		p.composeOps = newComposeOperations(p)
		return p, state
	}

	// The compose command itself cannot run here — there is no docker and no compose file —
	// so what is asserted is the state of `.env` by the time the operation gets that far.
	for name, run := range map[string]func(*processor) error{
		"update": func(p *processor) error {
			return p.composeOps.updateStack(context.Background(), ProductStackPentagi, newOperationState(nil))
		},
		"download": func(p *processor) error {
			return p.composeOps.downloadStack(context.Background(), ProductStackPentagi, newOperationState(nil))
		},
	} {
		p, state := newProcessor()
		_ = run(p)
		assert.Equal(t, "vxcontrol/pentagi:2.3.4", state.written["PENTAGI_IMAGE"],
			"%s reached the compose command without writing the reference the server chose", name)
	}

	for name, run := range map[string]func(*processor) error{
		"remove": func(p *processor) error {
			return p.composeOps.removeStack(context.Background(), ProductStackPentagi, newOperationState(nil))
		},
		"purge": func(p *processor) error {
			return p.composeOps.purgeStack(context.Background(), ProductStackPentagi, newOperationState(nil))
		},
	} {
		p, state := newProcessor()
		_ = run(p)
		assert.Empty(t, state.written, "%s must not write image references while tearing a stack down", name)
	}
}

// TestEveryComposeImageIsParameterised is the other end of the mechanism.
//
// A hardcoded `image:` cannot be steered at all: the server's answer is written
// into a variable nothing reads, the pull takes whatever the file says, and the
// update reports success having changed nothing. This walks the real files.
func TestEveryComposeImageIsParameterised(t *testing.T) {
	root := repositoryRoot(t)
	files := []string{
		"docker-compose.yml",
		"docker-compose-langfuse.yml",
		"docker-compose-observability.yml",
		"docker-compose-graphiti.yml",
	}

	imageLine := regexp.MustCompile(`(?m)^\s*image:\s*(\S+)\s*$`)
	parameterised := regexp.MustCompile(`^\$\{([A-Z0-9_]+):-(\S+)\}$`)

	seen := 0
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err, "the compose files are part of the contract with the server")

		for _, match := range imageLine.FindAllStringSubmatch(string(body), -1) {
			seen++
			reference := match[1]
			parts := parameterised.FindStringSubmatch(reference)
			require.NotNil(t, parts,
				"%s: %q is hardcoded, so the server cannot steer what this service pulls", name, reference)

			// `:-` and not `-`: the first substitutes the default for an absent
			// OR empty variable, the second only for an absent one. An
			// installation that wrote an empty value would otherwise pull "".
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

// repositoryRoot walks up to the directory holding the compose files.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, parent, dir, "no docker-compose.yml above the test directory")
		dir = parent
	}
	t.Fatal("could not find the repository root")
	return ""
}

// TestASharperTagIsNotDrift is the distinction the whole check turns on.
//
// The update path writes the server's chosen reference into these variables, and
// under `stable` that is deliberately a more specific tag than the file ships:
// `vxcontrol/pentagi:2.3.4` against `vxcontrol/pentagi:latest`. Reporting that as
// drift would tell the user their installation is misconfigured every time an
// update worked.
//
// A different REPOSITORY is the real case — an installation pinned to
// `postgres:15` while this build ships `postgres:16`, or an image rebuilt under
// somebody's own name — and it is the one the server answers `unknown` for.
func TestASharperTagIsNotDrift(t *testing.T) {
	defaults := map[string]string{
		"PENTAGI_IMAGE":  "vxcontrol/pentagi:latest",
		"POSTGRES_IMAGE": "postgres:16",
		"NEO4J_IMAGE":    "neo4j:5.26.2",
		"REDIS_IMAGE":    "redis:7",
	}

	drift := referenceDrift(defaults, map[string]string{
		// The update path pinned it. Same repository, sharper tag: working.
		"PENTAGI_IMAGE": "vxcontrol/pentagi:2.3.4",
		// A different major of the same image is still the same repository, so
		// this is NOT reported: moving between them is a data migration and not
		// something to nag about on an update screen.
		"POSTGRES_IMAGE": "postgres:15",
		// A rebuild under another name is a configuration this build cannot
		// support, and the server answers `repository_not_tracked` for it.
		"NEO4J_IMAGE": "acme/neo4j-fork:5.26.2",
		// Unset: the compose default applies, which is exactly right.
		"REDIS_IMAGE": "",
	})

	require.Len(t, drift, 1, "only a different repository is drift: %+v", drift)
	assert.Equal(t, "NEO4J_IMAGE", drift[0].Variable)
	assert.Equal(t, "acme/neo4j-fork:5.26.2", drift[0].Configured)
	assert.Equal(t, "neo4j:5.26.2", drift[0].Default)
}

// TestTheDefaultsAreReadFromTheFilesThemselves. A table in Go would be a second
// source of truth able to disagree with the files docker compose actually reads.
func TestTheDefaultsAreReadFromTheFilesThemselves(t *testing.T) {
	root := repositoryRoot(t)
	defaults, err := composeReferenceDefaults(func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, name))
	})
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(defaults), 20, "every image is parameterised, so every one has a default")
	assert.Equal(t, "vxcontrol/pentagi:latest", defaults["PENTAGI_IMAGE"])
	assert.Equal(t, "postgres:16", defaults["POSTGRES_IMAGE"])
	// The two ClickHouse services ship the same image under different variables,
	// which is why the variable and not the image is the key.
	assert.Equal(t, "clickhouse/clickhouse-server:24", defaults["CLICKHOUSE_IMAGE"])
	assert.Equal(t, "clickhouse/clickhouse-server:24", defaults["CLICKSTORE_IMAGE"])

	// A build that shipped no parameterised image at all would make every
	// installation look perfectly aligned.
	_, err = composeReferenceDefaults(func(string) ([]byte, error) { return []byte("services:\n"), nil })
	assert.Error(t, err, "finding nothing must be an error, not a clean bill of health")
}
