package models

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"pentagi/cmd/installer/processor"
	"pentagi/cmd/installer/wizard/locale"
)

// As with the update overview, what is tested here is the part that carries the meaning:
// the text the screen puts in front of the user before anything is downloaded. The
// viewport and the terminal panel are mechanics shared with the operation form.

func offerLines(t *testing.T, running string, offered *processor.InstallerPackage, err error) string {
	t.Helper()
	return strings.Join(installerUpdateLines(running, offered, err), "\n")
}

// TestTheScreenNamesTheBuildBeforeOfferingToFetchIt.
//
// This is the whole reason the screen exists. The entry used to lead into a generic
// "are you sure?" whose help text described an operation nobody had written: replacing the
// running binary and exiting. Consent to something undescribed is not consent, so the
// version pair, the size and the destination path all have to be on screen first.
func TestTheScreenNamesTheBuildBeforeOfferingToFetchIt(t *testing.T) {
	document := offerLines(t, "2.0.0", &processor.InstallerPackage{
		Version: "2.1.0",
		OS:      "linux",
		Arch:    "amd64",
		Size:    25 << 20,
		Path:    "/opt/pentagi/installer_2.1.0",
	}, nil)

	contains(t, document, "2.0.0", "the version being run must be named")
	contains(t, document, "2.1.0", "the version on offer must be named")
	contains(t, document, "linux/amd64", "the package is per-platform and the screen says which")
	contains(t, document, "25.0 MiB", "the size comes from packages/info; a raw byte count is not an answer")
	contains(t, document, "/opt/pentagi/installer_2.1.0", "where the file lands is part of what is agreed to")
	contains(t, document, locale.InstallerUpdatePressEnter, "the one action must be spelled out")
}

// TestTheScreenSaysTheRunningInstallerSurvives.
//
// The claim, not the phrasing: applying the update is out of scope by design — a process
// cannot reliably replace the binary it is executing, and doing it badly leaves the user
// with no installer at all. The text this replaced promised exactly that, so the promise
// is pinned here.
func TestTheScreenSaysTheRunningInstallerSurvives(t *testing.T) {
	document := offerLines(t, "2.0.0", &processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_2.1.0",
	}, nil)

	contains(t, document, locale.InstallerUpdateManualNote,
		"the screen must say what happens to the installer being run")

	note := strings.ToLower(locale.InstallerUpdateManualNote)
	if !strings.Contains(note, "not replaced") {
		t.Errorf("the note must say the running installer is NOT replaced, got: %q",
			locale.InstallerUpdateManualNote)
	}

	absent(t, strings.ToLower(document), "will exit",
		"nothing here exits the app; the old help text said it did and it never happened")
}

// TestAnOfferOfTheVersionAlreadyRunningSaysSo.
//
// The entry only appears when the server reports a newer build, but that answer can be
// older than the screen — a re-check while the app is open, or a strategy changed under it.
// Two identical versions printed one above the other read as an update either way.
func TestAnOfferOfTheVersionAlreadyRunningSaysSo(t *testing.T) {
	same := offerLines(t, "2.1.0", &processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_2.1.0",
	}, nil)
	contains(t, same, locale.InstallerUpdateSameVersion,
		"the same version twice must be called out, not left to be spotted")

	moved := offerLines(t, "2.0.0", &processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024, Path: "installer_2.1.0",
	}, nil)
	absent(t, moved, locale.InstallerUpdateSameVersion,
		"a real update must not be told it is already installed")
}

// TestAFailedLookupShowsTheFailureAndNothingElse.
//
// Without the description there is no size, no sha256 and no signature — nothing to check
// a download against. Showing a half-filled offer would invite Enter on a screen where
// Enter is refused, so the failure replaces the offer rather than joining it.
func TestAFailedLookupShowsTheFailureAndNothingElse(t *testing.T) {
	document := offerLines(t, "2.0.0", nil, errors.New("update server is unreachable"))

	contains(t, document, "update server is unreachable", "the reason must reach the user")
	absent(t, document, locale.InstallerUpdatePressEnter,
		"offering the action while the description is missing invites a keypress that is refused")
}

// TestTheOfferIsAwaitedRatherThanGuessed. Both nil means the answer has not arrived; the
// panel must say it is waiting instead of rendering an offer of empty strings.
func TestTheOfferIsAwaitedRatherThanGuessed(t *testing.T) {
	document := offerLines(t, "2.0.0", nil, nil)

	contains(t, document, locale.InstallerUpdateAsking, "an unanswered screen says it is asking")
	absent(t, document, locale.InstallerUpdatePressEnter, "there is nothing to press Enter for yet")
}

// TestTheScreenHandsOverTheCommandThatInstallsTheBuild.
//
// The download is only half of a self-update, and the installer deliberately does not do
// the other half. A screen that stops at "downloaded, verified" leaves a file with a name
// nothing will ever run — so the exact command has to be on screen, with both paths in it.
func TestTheScreenHandsOverTheCommandThatInstallsTheBuild(t *testing.T) {
	document := strings.Join(installerMoveInstructions(&processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024,
		Path:        "/opt/pentagi/installer_2.1.0",
		CurrentPath: "/usr/local/bin/pentagi-installer",
	}), "\n")

	contains(t, document, "/opt/pentagi/installer_2.1.0", "where the build is must be said")
	contains(t, document, `mv "/opt/pentagi/installer_2.1.0" "/usr/local/bin/pentagi-installer"`,
		"the command must name both paths and quote them — installations live under paths with spaces")
	absent(t, document, locale.InstallerUpdateWindowsNote,
		"renaming over a running binary is ordinary on linux; the warning belongs to Windows only")
}

// TestOnWindowsTheUserIsToldToCloseTheInstallerFirst. Windows refuses to replace a running
// program, so the command alone would fail with a permission error the user cannot read.
func TestOnWindowsTheUserIsToldToCloseTheInstallerFirst(t *testing.T) {
	document := strings.Join(installerMoveInstructions(&processor.InstallerPackage{
		Version: "2.1.0", OS: "windows", Arch: "amd64",
		Path:        `C:\pentagi\installer_2.1.0.exe`,
		CurrentPath: `C:\Program Files\PentAGI\installer.exe`,
	}), "\n")

	contains(t, document, `move /Y "C:\pentagi\installer_2.1.0.exe" "C:\Program Files\PentAGI\installer.exe"`,
		"mv is not a Windows command, and the destination has a space in it")
	contains(t, document, locale.InstallerUpdateWindowsNote, "the command fails while the installer runs")
}

// TestWithoutTheRunningPathNoCommandIsInvented.
//
// os.Executable can fail. A command built around a guessed destination is worse than no
// command: run it and the build lands somewhere nobody looks, or over something else.
func TestWithoutTheRunningPathNoCommandIsInvented(t *testing.T) {
	document := strings.Join(installerMoveInstructions(&processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Path: "/opt/pentagi/installer_2.1.0", CurrentPath: "",
	}), "\n")

	contains(t, document, "/opt/pentagi/installer_2.1.0", "where the build is is still known")
	contains(t, document, locale.InstallerUpdateMoveUnknown, "what has to happen must still be said")
	absent(t, document, "mv \"", "no command may be printed with a guessed destination in it")
}

// TestAFileAlreadyOnDiskIsAnnouncedBeforeEnterIsPressed.
//
// The confirmation exists for exactly one case — something already occupies that name —
// and a confirmation is not the place to learn that for the first time.
func TestAFileAlreadyOnDiskIsAnnouncedBeforeEnterIsPressed(t *testing.T) {
	present := offerLines(t, "2.0.0", &processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024,
		Path: "installer_2.1.0", Downloaded: true,
	}, nil)
	contains(t, present, locale.InstallerUpdateAlreadyHere, "the file being there must be said up front")

	absentFile := offerLines(t, "2.0.0", &processor.InstallerPackage{
		Version: "2.1.0", OS: "linux", Arch: "amd64", Size: 1024,
		Path: "installer_2.1.0", Downloaded: false,
	}, nil)
	absent(t, absentFile, locale.InstallerUpdateAlreadyHere,
		"downloading over nothing is not a question and must not read like one")
}

// TestEveryScreenModelIsRestorable.
//
// RestoreModel is how the app turns whatever a screen's Update returned back into a
// BaseScreenModel. What a missing branch costs is worth stating precisely, because the
// obvious answer is wrong: for a screen whose Update has a pointer receiver and returns
// itself — which is every screen here — the mutation has already happened in place and
// app.forwardMsgToCurrentModel keeps the pointer it already holds, so nothing freezes.
//
// The branch is load-bearing for the case the type allows but nobody has written yet: a
// model that returns a DIFFERENT model. That one is dropped silently, with no error
// anywhere, and it is the reason the contract is "every screen is listed" rather than
// "list the ones that need it" — the day a screen starts returning something else is not
// the day anybody remembers this rule.
func TestEveryScreenModelIsRestorable(t *testing.T) {
	for name, model := range map[string]tea.Model{
		"InstallerUpdateModel": (*InstallerUpdateModel)(nil),
		"UpdateOverviewModel":  (*UpdateOverviewModel)(nil),
	} {
		if RestoreModel(model) == nil {
			t.Errorf("%s is missing its branch in RestoreModel", name)
		}
	}
}
