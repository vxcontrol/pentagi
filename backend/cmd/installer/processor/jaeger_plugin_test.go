package processor

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A curated release names current artefacts too, and the file name says "linux" whatever the host.
func TestJaegerPlugin_JaegerPluginUpdates_PicksOnlyOutdatedLinuxPlugins(t *testing.T) {
	updates := []checker.StackUpdate{
		{
			Stack: "observability",
			Components: []checker.ComponentUpdate{
				{Component: "jaeger-clickhouse", OS: "linux", Arch: "amd64", Outdated: true, TargetVersion: "0.13.0"},
				{Component: "jaeger-clickhouse", OS: "linux", Arch: "arm64", Outdated: true, TargetVersion: "0.13.0"},
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

// os.Rename keeps the staging file's mode, and a running Jaeger loads whatever sits at the target.
func TestJaegerPlugin_InstallPluginFile_PutsOnlyACompleteExecutableFileInPlace(t *testing.T) {
	for _, tc := range []struct {
		name     string
		target   string
		existing string
		write    func(io.Writer) error
		wantErr  string
		want     string
	}{
		{name: "a failed download leaves the running plugin alone",
			target: "jaeger-clickhouse-linux-amd64", existing: "the plugin currently in use",
			write: func(w io.Writer) error {
				if _, err := w.Write([]byte("partial")); err != nil {
					return err
				}
				return errors.New("connection reset")
			},
			wantErr: "connection reset", want: "the plugin currently in use"},
		{name: "a complete download lands executable in a directory never created before",
			target: filepath.Join("sub", "jaeger-clickhouse-linux-arm64"),
			write: func(w io.Writer) error {
				_, err := w.Write([]byte("plugin bytes"))
				return err
			},
			want: "plugin bytes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), tc.target)
			if tc.existing != "" {
				require.NoError(t, os.WriteFile(target, []byte(tc.existing), 0o755))
			}

			err := installPluginFile(target, tc.write)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantErr)
			}

			content, readErr := os.ReadFile(target)
			require.NoError(t, readErr)
			assert.Equal(t, tc.want, string(content))
			info, statErr := os.Stat(target)
			require.NoError(t, statErr)
			assert.NotZero(t, info.Mode()&0o111, "mode is %v; Jaeger cannot exec a plugin without the bit", info.Mode())
			assert.NoFileExists(t, target+stagingSuffix, "a ~28 MB leftover that nothing else ever removes")
		})
	}
}

// The directory is mounted into a container and may hold an operator's own *.new files.
func TestJaegerPlugin_RemoveStagedPlugins_RemovesOnlyItsOwnLeftovers(t *testing.T) {
	base := filepath.Join(t.TempDir(), checker.JaegerPluginDir)
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
	assert.FileExists(t, somebodyElses)
}
