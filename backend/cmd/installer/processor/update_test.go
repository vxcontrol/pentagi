package processor

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/cloud/models"
)

// updateFetcher holds a malformed description of 1.2.3, which the SDK rejects before connecting.
func updateFetcher(t *testing.T) (*updateOperationsImpl, string) {
	t.Helper()
	dir := t.TempDir()
	// environmentDir() is the directory of the environment file, and "." without state.
	t.Chdir(dir)

	p := &processor{
		mu: &sync.Mutex{},
		checker: &checker.CheckResult{
			StackUpdates: []checker.StackUpdate{{Stack: "installer", LatestVersion: "1.2.3"}},
		},
	}
	u := &updateOperationsImpl{
		processor:    p,
		described:    &models.PackageInfoResponse{Size: 10},
		describedKey: installerPackageKey("1.2.3"),
	}
	p.updateOps = u
	return u, dir
}

func TestUpdate_ProgressWriter_ReportsEachTenthOnceAndNothingWithoutASize(t *testing.T) {
	for _, tc := range []struct {
		name   string
		total  int64
		writes []int
		want   []int
	}{
		{"the advertised size written in tenths reports every tenth once", 1000,
			slices.Repeat([]int{10}, 100), []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}},
		{"a longer stream stops at a hundred", 1000,
			append(slices.Repeat([]int{10}, 100), 500), []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}},
		{"no size means no percentage", 0, []int{4096}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reported []int
			w := &progressWriter{total: tc.total, report: func(percent int, _, _ int64) {
				reported = append(reported, percent)
			}}
			for _, size := range tc.writes {
				n, err := w.Write(make([]byte, size))
				require.NoError(t, err)
				require.Equal(t, size, n, "a writer that reports fewer bytes than it took stalls io.Copy")
			}
			assert.Equal(t, tc.want, reported)
		})
	}
}

func TestUpdate_CachedPackage_MatchesOnlyTheSameBuild(t *testing.T) {
	held := &models.PackageInfoResponse{Version: "1.2.3", Size: 42}
	u := &updateOperationsImpl{described: held, describedKey: installerPackageKey("1.2.3")}

	assert.Same(t, held, u.cachedPackage(installerPackageKey("1.2.3")))
	assert.Nil(t, u.cachedPackage(installerPackageKey("1.2.4")), "a different version")
	assert.Nil(t, u.cachedPackage("1.2.3/plan9/sparc"), "a different platform")

	assert.Contains(t, installerPackageKey("1.2.3"), runtime.GOOS)
	assert.Contains(t, installerPackageKey("1.2.3"), runtime.GOARCH)

	empty := &updateOperationsImpl{}
	assert.Nil(t, empty.cachedPackage(""), "nothing held is not a match for an empty key")
}

// A kept description would fail every retry of a republished build the same way.
func TestUpdate_FetchInstaller_LeavesNoFileAndNoDescriptionAfterAFailedDownload(t *testing.T) {
	u, dir := updateFetcher(t)

	_, err := u.fetchInstaller(t.Context(), newOperationState(nil))
	require.ErrorContains(t, err, "failed to download installer")

	assert.NoFileExists(t, filepath.Join(dir, installerFileName("1.2.3")),
		"an unverified file must not be left under a name that looks like a release")
	assert.Nil(t, u.described)
}

// An installer renamed to installer_<its own version> sits on the download path.
func TestUpdate_FetchInstaller_RefusesToOverwriteTheRunningInstaller(t *testing.T) {
	u, dir := updateFetcher(t)
	running, err := os.Executable()
	require.NoError(t, err)

	// A hard link is the same inode as the test binary without touching it.
	target := filepath.Join(dir, installerFileName("1.2.3"))
	if err := os.Link(running, target); err != nil {
		t.Skipf("cannot hard-link the test binary here: %v", err)
	}

	_, err = u.fetchInstaller(t.Context(), newOperationState(nil))
	require.ErrorContains(t, err, "refusing to overwrite the running installer")

	info, statErr := os.Stat(target)
	require.NoError(t, statErr)
	assert.Positive(t, info.Size(), "the running installer was truncated")
}

func TestUpdate_InstallerFileName_CarriesTheVersion(t *testing.T) {
	want := "installer_2.0.1"
	if runtime.GOOS == "windows" {
		want += ".exe"
	}
	assert.Equal(t, want, installerFileName("2.0.1"))
}

func TestUpdate_IsSameFile_ComparesInodesNotStrings(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	require.NoError(t, os.WriteFile(real, []byte("x"), 0o644))

	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink(real, link))

	assert.True(t, isSameFile(real, link), "a symlink names the file it points at")
	assert.True(t, isSameFile(real, filepath.Join(dir, ".", "real")), "same path, other spelling")
	assert.False(t, isSameFile(real, filepath.Join(dir, "missing")), "a path that does not exist is nothing")
	assert.False(t, isSameFile(filepath.Join(dir, "missing"), real))
}

// The installer does not apply the update, so this command is what the user runs.
func TestUpdate_MoveCommand_QuotesBothPathsForThePlatform(t *testing.T) {
	unix := InstallerPackage{OS: "linux", Path: "/opt/p/installer_2.1.0", CurrentPath: "/usr/local/bin/inst"}
	assert.Equal(t, `mv "/opt/p/installer_2.1.0" "/usr/local/bin/inst"`, unix.MoveCommand())

	windows := InstallerPackage{
		OS: "windows", Path: `C:\p\installer_2.1.0.exe`, CurrentPath: `C:\Program Files\P\inst.exe`,
	}
	assert.Equal(t, `move /Y "C:\p\installer_2.1.0.exe" "C:\Program Files\P\inst.exe"`, windows.MoveCommand())

	// A command with a guessed destination is worse than no command at all.
	assert.Empty(t, InstallerPackage{OS: "linux", Path: "/opt/p/x"}.MoveCommand())
	assert.Empty(t, InstallerPackage{OS: "linux", CurrentPath: "/usr/local/bin/inst"}.MoveCommand())
}

func TestUpdate_FormatSize_UsesBinaryUnitsUpToTiB(t *testing.T) {
	for size, want := range map[int64]string{
		0:        "0 B",
		512:      "512 B",
		1024:     "1.0 KiB",
		1536:     "1.5 KiB",
		25 << 20: "25.0 MiB",
		3 << 30:  "3.0 GiB",
		5 << 40:  "5.0 TiB",
		1 << 50:  "1024.0 TiB",
		-1:       "-1 B",
	} {
		assert.Equal(t, want, FormatSize(size), "%d bytes", size)
	}
}

// `remove all` and `purge all` end on the installer stack, after all the real work.
func TestUpdate_RemoveInstaller_Succeeds(t *testing.T) {
	u := &updateOperationsImpl{processor: &processor{}}
	assert.NoError(t, u.removeInstaller(t.Context(), newOperationState(nil)))
}
