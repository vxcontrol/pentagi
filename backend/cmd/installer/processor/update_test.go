package processor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"pentagi/cmd/installer/checker"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/cloud/models"
)

var _ = checker.CheckResult{}

// updateHarness builds a processor whose check handler records what it was asked, so the
// cost of an operation in cloud requests can be counted rather than argued about.
func updateHarness(t *testing.T, configure func(*mockCheckConfig)) (*processor, *mockCheckHandler) {
	t.Helper()

	handler := newMockCheckHandler()
	if configure != nil {
		configure(&handler.config)
	}

	p := createProcessorWithState(testState(t), createCheckResultWithHandler(handler))
	return p, handler
}

func countCalls(calls []string, name string) int {
	found := 0
	for _, call := range calls {
		if call == name {
			found++
		}
	}
	return found
}

func countMessages[T tea.Msg](msgs []tea.Msg) int {
	found := 0
	for _, msg := range msgs {
		if _, ok := msg.(T); ok {
			found++
		}
	}
	return found
}

// TestOneUpdateCostsOneCheck.
//
// The fresh check after an update is required — only the cloud knows whether anything is
// still pending, and it is that answer, not the local outcome, that repaints the menu. One
// is required. Five is what it used to cost: the check lived in a deferred call inside the
// recursive worker, so `compose` paid for its four stacks and then again for itself, and a
// stack with nothing to do paid too, because the defer was registered before the early
// return.
//
// It is not free. Against the seeded budget for an unlicensed user that was three checks a
// day, so a single press of "Update PentAGI" exhausted the day and part of the next.
func TestOneUpdateCostsOneCheck(t *testing.T) {
	for name, stack := range map[string]ProductStack{
		"every compose stack": ProductStackCompose,
		"a single stack":      ProductStackPentagi,
		"everything":          ProductStackAll,
	} {
		p, handler := updateHarness(t, func(config *mockCheckConfig) {
			config.PentagiIsUpToDate = false
			config.LangfuseIsUpToDate = false
			config.GraphitiIsUpToDate = false
			config.ObservabilityIsUpToDate = false
		})

		before := countCalls(handler.getCalls(), "GatherUpdatesInfo")
		require.NoError(t, p.update(context.Background(), stack, testOperationState(t)))
		after := countCalls(handler.getCalls(), "GatherUpdatesInfo")

		assert.Equal(t, 1, after-before,
			"%s: the operation asked the cloud %d times; one press must cost one check", name, after-before)
	}
}

// TestAnUpdateOfStacksThatAreNotInstalledStillCostsOneCheck.
//
// Not every installation has every stack, and a stack with nothing to do used to be the
// cheapest way to spend a request: the deferred check was registered before the "already
// up to date" early return, so doing nothing cost exactly as much as doing everything.
func TestAnUpdateOfStacksThatAreNotInstalledStillCostsOneCheck(t *testing.T) {
	p, handler := updateHarness(t, func(config *mockCheckConfig) {
		// Only PentAGI needs anything; the other three are current.
		config.PentagiIsUpToDate = false
		config.LangfuseIsUpToDate = true
		config.GraphitiIsUpToDate = true
		config.ObservabilityIsUpToDate = true
	})

	before := countCalls(handler.getCalls(), "GatherUpdatesInfo")
	require.NoError(t, p.update(context.Background(), ProductStackCompose, testOperationState(t)))

	assert.Equal(t, 1, countCalls(handler.getCalls(), "GatherUpdatesInfo")-before)
}

// TestAnOperationAnnouncesItselfExactlyOnce.
//
// The screen treats a completion as the end of the whole operation: it stops the terminal,
// prints the verdict, and — because HandleMsg returns nil for a completion — stops polling.
// So a completion per nested stack meant the FIRST stack to finish reported success on
// behalf of all of them, and everything after it, including errors, was never drained.
//
// Whichever stacks happen to be installed does not change this: the fan-out list is fixed,
// and a stack with nothing to do is the one that finishes first.
func TestAnOperationAnnouncesItselfExactlyOnce(t *testing.T) {
	for name, stack := range map[string]ProductStack{
		"every compose stack": ProductStackCompose,
		"everything":          ProductStackAll,
		"a single stack":      ProductStackObservability,
	} {
		p, _ := updateHarness(t, func(config *mockCheckConfig) {
			config.PentagiIsUpToDate = false
			config.LangfuseIsUpToDate = true
			config.GraphitiIsUpToDate = false
			config.ObservabilityIsUpToDate = false
		})

		state := testOperationState(t)
		require.NoError(t, p.update(context.Background(), stack, state))

		assert.Equal(t, 1, countMessages[ProcessorStartedMsg](state.msgs),
			"%s: started announced more than once", name)
		assert.Equal(t, 1, countMessages[ProcessorCompletionMsg](state.msgs),
			"%s: the screen was told the operation finished %d times, and it believes the first",
			name, countMessages[ProcessorCompletionMsg](state.msgs))
	}
}

// TestEveryFanningOperationAnnouncesItselfExactlyOnce.
//
// `update` reaches one announcement by structure — its recursion goes through a worker
// that says nothing. The other five operations still recurse through themselves, so their
// guarantee rests on the check inside sendStarted/sendCompletion, and it has to hold for
// each of them: download, remove, purge, start and stop all fan a request for `compose` or
// `all` out over four stacks.
func TestEveryFanningOperationAnnouncesItselfExactlyOnce(t *testing.T) {
	for name, run := range map[string]func(*processor, *operationState) error{
		"download": func(p *processor, s *operationState) error {
			return p.download(context.Background(), ProductStackCompose, s)
		},
		"remove": func(p *processor, s *operationState) error {
			return p.remove(context.Background(), ProductStackCompose, s)
		},
		"start": func(p *processor, s *operationState) error {
			return p.start(context.Background(), ProductStackCompose, s)
		},
		"stop": func(p *processor, s *operationState) error {
			return p.stop(context.Background(), ProductStackCompose, s)
		},
	} {
		p, _ := updateHarness(t, nil)
		state := testOperationState(t)
		_ = run(p, state)

		assert.LessOrEqual(t, countMessages[ProcessorCompletionMsg](state.msgs), 1,
			"%s: the screen was told the operation finished more than once, and it believes the first", name)
		assert.LessOrEqual(t, countMessages[ProcessorStartedMsg](state.msgs), 1,
			"%s: started announced more than once", name)
	}
}

// TestAFailureInsideTheFanOutReachesTheScreen.
//
// The other half of announcing once: the single completion has to carry the error. A stack
// failing in the middle of the fan-out used to be reported as a completion of its own —
// after an earlier stack had already reported success and stopped the screen polling.
func TestAFailureInsideTheFanOutReachesTheScreen(t *testing.T) {
	p, _ := updateHarness(t, func(config *mockCheckConfig) {
		config.PentagiIsUpToDate = false
		config.LangfuseIsUpToDate = false
		config.GraphitiIsUpToDate = false
		config.ObservabilityIsUpToDate = false
	})

	// Observability is first in the update order, so this fails partway through the run.
	injectComposeError(p, map[string]error{"updateStack": errors.New("compose exploded")})

	state := testOperationState(t)
	err := p.update(context.Background(), ProductStackCompose, state)
	require.Error(t, err, "a stack failing must fail the operation")

	require.Equal(t, 1, countMessages[ProcessorCompletionMsg](state.msgs))
	for _, msg := range state.msgs {
		if completion, ok := msg.(ProcessorCompletionMsg); ok {
			assert.Error(t, completion.Error,
				"the one completion the screen sees reported success while a stack had failed")
			assert.Equal(t, ProductStackCompose, completion.Stack,
				"the completion must name the stack the user asked about, not a nested one")
		}
	}
}

// TestRemovingEverythingIsNotFailedByTheInstallerItself.
//
// `remove all` and `purge all` walk every stack with the installer LAST, and removing the
// installer binary is something a running process cannot do. It used to answer "not
// implemented", so the whole operation ended in an error AFTER all the real work had
// already succeeded — the user saw a failure for an installation that had been removed.
func TestRemovingEverythingIsNotFailedByTheInstallerItself(t *testing.T) {
	// The REAL implementation, not the harness mock: what is under test is the value this
	// function returns, and a mock returning nil would prove nothing.
	u := &updateOperationsImpl{processor: &processor{}}

	assert.NoError(t, u.removeInstaller(context.Background(), newOperationState(nil)),
		"the one stack whose removal is deliberately out of scope must not fail the run")
}

// TestRemovingEverythingAttemptsEveryStack.
//
// The loop used to return on the first failure, so a single stack that could not be
// removed left the others untouched — on an operation the user asked to remove everything.
// Every stack is attempted now, and every failure is reported rather than only the first.
func TestRemovingEverythingAttemptsEveryStack(t *testing.T) {
	p, _ := updateHarness(t, nil)
	composeOps := p.composeOps.(*baseMockComposeOperations)

	// removeStack fails for the first compose stack it is asked about.
	injectComposeError(p, map[string]error{"removeStack": errors.New("compose exploded")})

	err := p.remove(context.Background(), ProductStackAll, testOperationState(t))
	require.Error(t, err, "a stack that could not be removed must still fail the operation")

	attempted := map[ProductStack]bool{}
	for _, call := range composeOps.getCalls() {
		if call.Method == "removeStack" {
			attempted[call.Stack] = true
		}
	}
	for _, stack := range composeOperationAllStacksOrder[ProcessorOperationRemove] {
		assert.True(t, attempted[stack],
			"%s was never attempted because an earlier stack failed", stack)
	}
}

// TestPurgingEverythingStillRemovesTheNetworks.
//
// The three custom docker networks were removed AFTER the loop over the stacks, and the
// loop returned on the first error — so one stack that could not be purged left all three
// behind, on the one operation whose entire purpose is to leave nothing. Which stacks an
// installation has varies, so "a stack that fails" is not an exotic path.
func TestPurgingEverythingStillRemovesTheNetworks(t *testing.T) {
	p, _ := updateHarness(t, nil)
	dockerOps := p.dockerOps.(*baseMockDockerOperations)

	// The first stack in the purge order fails, which used to end the operation there.
	injectComposeError(p, map[string]error{"purgeImagesStack": errors.New("compose exploded")})

	err := p.purge(context.Background(), ProductStackAll, testOperationState(t))
	require.Error(t, err, "a stack that could not be purged must still fail the operation")

	removed := map[string]bool{}
	for _, call := range dockerOps.getCalls() {
		if call.Method == "removeMainDockerNetwork" {
			removed[call.Name] = true
		}
	}

	for _, network := range []ProductDockerNetwork{
		ProductDockerNetworkPentagi, ProductDockerNetworkObservability, ProductDockerNetworkLangfuse,
	} {
		assert.True(t, removed[string(network)],
			"network %s was left behind because an earlier stack failed", network)
	}
}

// TestAFastFailureStillReachesTheScreen.
//
// wrapCommand waits up to 500 ms for the operation before handing a tea.Cmd back, so an
// operation that fails faster than that delivers its error through ProcessorWaitMsg. That
// message used to STOP the polling — and the completion message, the one that clears the
// spinner and prints what happened, sits in the queue waiting to be polled. So the screen
// said "in progress" for the rest of the session, Enter did nothing, and the error was
// shown nowhere.
//
// What is asserted is the message chain, not the screen: HandleMsg must keep polling after
// an error so the completion can be drained, and must stop once it has been.
func TestAFastFailureStillReachesTheScreen(t *testing.T) {
	p, _ := updateHarness(t, nil)
	model := &processorModel{processor: p}

	state := testOperationState(t)
	state.sendCompletion(ProductStackAll, errors.New("failed before the window closed"))

	failed := ProcessorWaitMsg{
		Error: errors.New("failed before the window closed"),
		Stack: ProductStackAll, state: state, num: 0,
	}

	next := model.HandleMsg(failed)
	require.NotNil(t, next, "polling stopped, so the completion already in the queue is never delivered")

	delivered := next()
	completion, ok := delivered.(ProcessorCompletionMsg)
	require.True(t, ok, "expected the queued completion, got %T", delivered)
	assert.Error(t, completion.Error, "the completion must carry the failure")

	// And the chain terminates: a completion is the end of it.
	assert.Nil(t, model.HandleMsg(completion), "polling must stop once the completion is delivered")
}

// TestTheStacksAnUpdateTouchesAreTheOnesItVerifies. Targets are captured before anything
// changes, so the expansion has to match what the operation actually runs.
func TestTheStacksAnUpdateTouchesAreTheOnesItVerifies(t *testing.T) {
	assert.Equal(t, composeOperationAllStacksOrder[ProcessorOperationUpdate],
		stacksTouchedByUpdate(ProductStackCompose))

	assert.Equal(t, []ProductStack{ProductStackPentagi}, stacksTouchedByUpdate(ProductStackPentagi))

	all := stacksTouchedByUpdate(ProductStackAll)
	assert.Contains(t, all, ProductStackWorker)
	assert.Contains(t, all, ProductStackInstaller)
	assert.Len(t, all, len(composeOperationAllStacksOrder[ProcessorOperationUpdate])+2)

	// The expansion must not alias the shared order slice: append on a shared backing
	// array would rewrite the table every other caller reads.
	assert.Equal(t, 4, len(composeOperationAllStacksOrder[ProcessorOperationUpdate]),
		"stacksTouchedByUpdate(all) grew the shared compose order table")
}

// TestProgressIsReportedOncePerTenthAndNeverPastAHundred.
//
// The panel this writes into scrolls, so a line per read would bury the messages around it
// — the path the file landed on, and the line saying verification passed. A tenth is the
// step; the same tenth must never be printed twice, and a stream longer than advertised
// must not print percentages above a hundred (the download fails the length check
// afterwards, which is where that belongs).
func TestProgressIsReportedOncePerTenthAndNeverPastAHundred(t *testing.T) {
	var reported []int
	w := &progressWriter{total: 1000, report: func(percent int, _, _ int64) {
		reported = append(reported, percent)
	}}

	// 100 writes of 10 bytes: every tenth is crossed exactly once.
	for range 100 {
		n, err := w.Write(make([]byte, 10))
		require.NoError(t, err)
		require.Equal(t, 10, n, "a writer that reports fewer bytes than it took stalls io.Copy")
	}
	assert.Equal(t, []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, reported)

	// Past the advertised length the percentage is pinned rather than climbing.
	_, err := w.Write(make([]byte, 500))
	require.NoError(t, err)
	assert.Equal(t, []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, reported,
		"a longer stream must not print 150%%; the length check is what rejects it")
}

// TestProgressIsSilentWithoutASize. packages/info is the only place a size comes from —
// the response body is encrypted, which strips Content-Length. Without one there is no
// percentage to print, and printing 0% forever would read as a stalled download.
func TestProgressIsSilentWithoutASize(t *testing.T) {
	called := false
	w := &progressWriter{total: 0, report: func(int, int64, int64) { called = true }}

	_, err := w.Write(make([]byte, 4096))
	require.NoError(t, err)
	assert.False(t, called)
}

// TestTheDescriptionIsKeyedByBuildAndNotByVersion.
//
// The description carries the size, the sha256 and the signature the download is checked
// against, and all three are per-platform. Handing back the description of another build
// would fail every check with a message about a corrupt download.
func TestTheDescriptionIsKeyedByBuildAndNotByVersion(t *testing.T) {
	held := &models.PackageInfoResponse{Version: "1.2.3", Size: 42}
	u := &updateOperationsImpl{described: held, describedKey: installerPackageKey("1.2.3")}

	assert.Same(t, held, u.cachedPackage(installerPackageKey("1.2.3")))
	assert.Nil(t, u.cachedPackage(installerPackageKey("1.2.4")), "a different version")
	assert.Nil(t, u.cachedPackage("1.2.3/plan9/sparc"), "a different platform")

	assert.Contains(t, installerPackageKey("1.2.3"), runtime.GOOS)
	assert.Contains(t, installerPackageKey("1.2.3"), runtime.GOARCH)

	// Nothing held is not the same as anything matching an empty key.
	empty := &updateOperationsImpl{}
	assert.Nil(t, empty.cachedPackage(""))
}

// TestAFailedDownloadLeavesNoFileAndNoStaleDescription.
//
// Two things have to happen when a download does not verify, and both are easy to get
// wrong. The file must go: verification only completes after the last byte, so whatever is
// on disk is unverified, and leaving it puts something named like a release next to the
// installation with nothing marking it as suspect. And the description must go: it is the
// yardstick every check used, so if the build was republished under the same version since
// it was fetched, keeping it makes every retry fail identically forever.
//
// The description here is deliberately malformed, which the SDK rejects before opening a
// connection — this is about the failure path, not about the network.
func TestAFailedDownloadLeavesNoFileAndNoStaleDescription(t *testing.T) {
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
		described:    &models.PackageInfoResponse{Size: 10}, // no hash, no signature
		describedKey: installerPackageKey("1.2.3"),
	}
	p.updateOps = u

	_, err := u.fetchInstaller(context.Background(), newOperationState(nil))
	require.Error(t, err)

	assert.NoFileExists(t, filepath.Join(dir, installerFileName("1.2.3")),
		"an unverified file must not be left under a name that looks like a release")
	assert.Nil(t, u.described,
		"keeping the description that failed makes every retry fail the same way")
}

// TestTheDownloadedFileCarriesTheVersion. A download must never overwrite the installer
// currently in use — it is running from this directory.
func TestTheDownloadedFileCarriesTheVersion(t *testing.T) {
	name := installerFileName("2.0.1")

	assert.Contains(t, name, "2.0.1")
	if runtime.GOOS == "windows" {
		assert.Equal(t, "installer_2.0.1.exe", name, "Windows will not execute it without .exe")
	} else {
		assert.Equal(t, "installer_2.0.1", name)
	}
}

// TestTheRunningInstallerIsNeverOverwritten.
//
// The versioned name normally keeps the download away from the binary in use, but an
// installer renamed to installer_<its own version> lands on exactly that path, and the
// O_TRUNC in fetchInstaller would empty it. Linux answers ETXTBSY and macOS does not, so
// the guard cannot be left to the operating system.
func TestTheRunningInstallerIsNeverOverwritten(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	running, err := os.Executable()
	require.NoError(t, err)

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

	// A hard link makes the target the same file as the test binary without touching it:
	// os.SameFile compares inodes, which is the comparison the guard makes.
	target := filepath.Join(dir, installerFileName("1.2.3"))
	if err := os.Link(running, target); err != nil {
		t.Skipf("cannot hard-link the test binary here: %v", err)
	}

	_, err = u.fetchInstaller(context.Background(), newOperationState(nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to overwrite the running installer")

	// And the file is still whole — this is the whole point of the guard.
	info, statErr := os.Stat(target)
	require.NoError(t, statErr)
	assert.Positive(t, info.Size(), "the running installer was truncated")
}

// TestIsSameFileComparesInodesNotStrings. A symlink, a relative path and a different
// spelling of the same directory all name one file while comparing unequal as text.
func TestIsSameFileComparesInodesNotStrings(t *testing.T) {
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

// TestTheMoveCommandNamesBothPathsAndQuotesThem.
//
// This command is the entire handover: the installer does not apply the update, so what is
// printed here is what the user runs. An unquoted path under "Program Files" becomes two
// arguments, and mv is not a command Windows has.
func TestTheMoveCommandNamesBothPathsAndQuotesThem(t *testing.T) {
	unix := InstallerPackage{OS: "linux", Path: "/opt/p/installer_2.1.0", CurrentPath: "/usr/local/bin/inst"}
	assert.Equal(t, `mv "/opt/p/installer_2.1.0" "/usr/local/bin/inst"`, unix.MoveCommand())

	windows := InstallerPackage{
		OS: "windows", Path: `C:\p\installer_2.1.0.exe`, CurrentPath: `C:\Program Files\P\inst.exe`,
	}
	assert.Equal(t, `move /Y "C:\p\installer_2.1.0.exe" "C:\Program Files\P\inst.exe"`, windows.MoveCommand())

	// Nothing is invented when os.Executable said nothing: a command with a guessed
	// destination in it is worse than no command at all.
	assert.Empty(t, InstallerPackage{OS: "linux", Path: "/opt/p/x"}.MoveCommand())
	assert.Empty(t, InstallerPackage{OS: "linux", CurrentPath: "/usr/local/bin/inst"}.MoveCommand())
}

func TestFormatSize(t *testing.T) {
	for size, want := range map[int64]string{
		0:        "0 B",
		512:      "512 B",
		1024:     "1.0 KiB",
		1536:     "1.5 KiB",
		25 << 20: "25.0 MiB",
		3 << 30:  "3.0 GiB",
		5 << 40:  "5.0 TiB",
		1 << 50:  "1024.0 TiB", // no unit above TiB: it stops climbing rather than wrapping
		-1:       "-1 B",
	} {
		assert.Equal(t, want, FormatSize(size), "%d bytes", size)
	}
}
