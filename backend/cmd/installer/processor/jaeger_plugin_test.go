package processor

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/cloud/models"
)

// TestOnlyOutdatedLinuxPluginsAreFetched.
//
// Membership in the answer is not the test: under a curated release the server names every
// artefact it knows about, current ones included, so fetching everything listed would
// re-download 28 MB per architecture on every observability update forever.
//
// The OS check is the other half. The file name has "linux" inside it because the plugin
// runs in the Jaeger container whatever the host is, so an answer for another platform
// would write non-linux content under a linux name — and Jaeger would fail to start.
func TestOnlyOutdatedLinuxPluginsAreFetched(t *testing.T) {
	updates := []checker.StackUpdate{
		{
			Stack: "observability",
			Components: []checker.ComponentUpdate{
				{Component: "jaeger-clickhouse", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: "0.13.0"},
				{Component: "jaeger-clickhouse", OS: "linux", Arch: "arm64", Outdated: true, TargetVersion: "0.13.0"},
				// Current: listed by the release, nothing to do.
				{Component: "jaeger", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: "1.56.0"},
			},
		},
		{
			Stack: "pentagi",
			Components: []checker.ComponentUpdate{
				{Component: "pentagi", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: "2.3.4"},
			},
		},
	}

	wanted := jaegerPluginUpdates(updates)
	require.Len(t, wanted, 2, "both architectures are kept in step: %+v", wanted)
	assert.Equal(t, "amd64", wanted[0].Arch)
	assert.Equal(t, "arm64", wanted[1].Arch)

	// Every reason a component must be skipped, one at a time.
	for name, component := range map[string]checker.ComponentUpdate{
		"already current":   {Component: "jaeger-clickhouse", OS: "linux", Arch: "amd64", Outdated: false, TargetVersion: "0.13.0"},
		"no version to ask": {Component: "jaeger-clickhouse", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: ""},
		"not linux":         {Component: "jaeger-clickhouse", OS: "darwin", Arch: "amd64", Outdated: true, TargetVersion: "0.13.0"},
		"another component": {Component: "installer", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: "0.13.0"},
	} {
		got := jaegerPluginUpdates([]checker.StackUpdate{
			{Stack: "observability", Components: []checker.ComponentUpdate{component}},
		})
		assert.Empty(t, got, "%s must not be fetched", name)
	}
}

// TestTheComponentNameMatchesTheContract. jaegerPluginUpdates compares against a literal
// string; if the vocabulary ever renames the component, the comparison silently stops
// matching and the plugin is never updated again, with no error anywhere.
func TestTheComponentNameMatchesTheContract(t *testing.T) {
	assert.Equal(t, "jaeger-clickhouse", string(models.ComponentTypeJaegerClickhouse))
}

// TestAFailedDownloadLeavesTheRunningPluginAlone.
//
// This is what staging is for. The file being replaced is the one a running Jaeger has
// loaded, so a download that dies halfway must not be visible at the target path at all —
// a truncated plugin brings the container back up unable to start, with a gRPC error that
// says nothing about a bad download.
func TestAFailedDownloadLeavesTheRunningPluginAlone(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "jaeger-clickhouse-linux-amd64")
	require.NoError(t, os.WriteFile(target, []byte("the plugin currently in use"), 0o755))

	err := installPluginFile(target, func(w io.Writer) error {
		// Half of it arrives, then the transfer fails — the ordinary shape of a broken
		// download, and the one that would truncate an in-place write.
		if _, err := w.Write([]byte("partial")); err != nil {
			return err
		}
		return errors.New("connection reset")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection reset")

	content, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	assert.Equal(t, "the plugin currently in use", string(content),
		"the plugin Jaeger is running was replaced by a failed download")

	assert.NoFileExists(t, target+stagingSuffix,
		"a ~28 MB leftover nothing else ever removes: integrity verification only knows the files the installer ships")
}

// TestTheInstalledPluginIsExecutable.
//
// os.Rename does NOT carry over the destination's mode — the result keeps the staging
// file's. A plugin created 0644 and renamed over an executable one ends up not executable,
// and Jaeger fails to start with an opaque gRPC plugin error rather than a permission one.
func TestTheInstalledPluginIsExecutable(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "jaeger-clickhouse-linux-arm64")

	require.NoError(t, installPluginFile(target, func(w io.Writer) error {
		_, err := w.Write([]byte("plugin bytes"))
		return err
	}))

	info, err := os.Stat(target)
	require.NoError(t, err, "the directory the mount expects is created when the stack was never installed")
	assert.Equal(t, "plugin bytes", readFile(t, target))
	assert.NotZero(t, info.Mode()&0o111, "mode is %v; Jaeger cannot exec a plugin without the bit", info.Mode())
	assert.NoFileExists(t, target+stagingSuffix, "the staging file must not survive a success")
}

// TestStaleStagingFilesAreRemovedByExactName.
//
// An interrupted download leaves ~28 MB that nothing else will ever clean up. But this
// directory is mounted into a container and an operator may keep their own files in it, so
// the cleanup names the two files it created rather than globbing "*.new".
func TestStaleStagingFilesAreRemovedByExactName(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, checker.JaegerPluginDir)
	require.NoError(t, os.MkdirAll(base, 0o755))

	for _, name := range checker.JaegerPluginBinaries {
		require.NoError(t, os.WriteFile(filepath.Join(base, name+stagingSuffix), []byte("half"), 0o755))
	}
	somebodyElses := filepath.Join(base, "operators-own-notes.new")
	require.NoError(t, os.WriteFile(somebodyElses, []byte("keep me"), 0o644))

	u := &updateOperationsImpl{processor: &processor{}}
	u.removeStagedPlugins(base, newOperationState(nil))

	for _, name := range checker.JaegerPluginBinaries {
		assert.NoFileExists(t, filepath.Join(base, name+stagingSuffix), "%s leftover survived", name)
	}
	assert.FileExists(t, somebodyElses, "a delete that guesses is how a tool destroys what it was not asked to touch")
}

// TestThePluginIsExemptFromIntegrityVerification.
//
// Without this the update undoes itself: a downloaded plugin differs from the copy the
// installer embeds, so verifyDirectoryContentIntegrity calls it modified and — under
// force, which is the path a stack update takes — copies the embedded version back over
// it. Tied to the plugin table rather than spelled out again, so a third architecture
// cannot be added on one side only.
func TestThePluginIsExemptFromIntegrityVerification(t *testing.T) {
	fs := &fileSystemOperationsImpl{}
	for arch, name := range checker.JaegerPluginBinaries {
		path := checker.JaegerPluginDir + "/" + name
		assert.True(t, fs.isExcludedFromVerification(path),
			"%s (%s) is not excluded, so a stack update would copy the embedded copy back over the download", path, arch)
	}
}

// TestJaegerIsRestartedOnlyWhenThePluginWasReplaced.
//
// `docker compose up -d` recreates a container only when its SPEC changes, and a different
// file inside a bind mount is not a spec change. So when the plugin was replaced and the
// jaeger image was not, the `up -d` in the update path leaves the container running with
// the plugin it exec'd at start — the update reports success and Jaeger keeps writing
// traces through the old plugin. The restart is what closes that, and it must not fire
// when nothing was replaced: it is downtime nobody asked for.
func TestJaegerIsRestartedOnlyWhenThePluginWasReplaced(t *testing.T) {
	for name, replaced := range map[string]bool{"plugin replaced": true, "nothing to do": false} {
		p, composeOps, _, _ := newProcessorForLogicTestsWithConfig(t, func(config *mockCheckConfig) {
			config.ObservabilityIsUpToDate = false
		})
		p.updateOps.(*baseMockUpdateOperations).jaegerPluginReplaced = replaced

		require.NoError(t, p.update(t.Context(), ProductStackObservability, testOperationState(t)))

		restarted := false
		for _, c := range composeOps.getCalls() {
			args, _ := c.Args.([]string)
			if c.Method == "performStackCommand" && slices.Equal(args, []string{"restart", jaegerServiceName}) {
				restarted = true
			}
		}
		assert.Equal(t, replaced, restarted, "%s: jaeger restarted=%v", name, restarted)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, fmt.Sprintf("reading %s", path))
	return string(content)
}
