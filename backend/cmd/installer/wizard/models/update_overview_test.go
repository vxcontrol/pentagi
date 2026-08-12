package models

import (
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
)

// The wizard has no harness for driving a bubbletea screen, so what is tested
// here is the part that carries the meaning: the document the screen renders.
// Everything above it — the viewport, the scrolling, the re-render on resize —
// is mechanics shared with the EULA screen, which has been in the field for
// releases.

func contains(t *testing.T, document, want, why string) {
	t.Helper()
	if !strings.Contains(document, want) {
		t.Errorf("%s\nexpected to find %q in:\n%s", why, want, document)
	}
}

func absent(t *testing.T, document, unwanted, why string) {
	t.Helper()
	if strings.Contains(document, unwanted) {
		t.Errorf("%s\nexpected NOT to find %q in:\n%s", why, unwanted, document)
	}
}

func TestTheOverviewNamesEveryReleaseTheInstallationCrosses(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack:          "pentagi",
			HasUpdate:      true,
			CurrentVersion: "2.0.0",
			LatestVersion:  "2.2.0",
			Changelog:      "target changelog",
			Releases: []checker.ReleaseSummary{
				{Version: "2.1.0", IsStable: true, ReleasedAt: "2026-01-02T03:04:05Z", Changelog: "first changelog"},
				{Version: "2.2.0", IsStable: true, Changelog: "second changelog", ReleaseNotes: "second notes"},
			},
		},
	}, true)

	contains(t, document, "2.0.0 → 2.2.0", "the move must be stated in both directions")
	contains(t, document, "2.1.0", "an intermediate release must be listed")
	contains(t, document, "first changelog", "each release carries its OWN text — that is the point of the list")
	contains(t, document, "second notes", "the target's release notes must be there")
	contains(t, document, "2026-01-02", "a published release shows when it was published")
}

// TestAnUnpublishedReleaseShowsNoDate. ReleasedAt is empty for a release nobody
// pressed Publish on, and printing "Released " with nothing after it looks like
// a bug in the installer rather than an absent value on the server.
func TestAnUnpublishedReleaseShowsNoDate(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
			Releases: []checker.ReleaseSummary{{Version: "2.1.0", IsStable: true, Changelog: "notes"}},
		},
	}, true)

	absent(t, document, "Released", "a release with no publication date must print no date line")
}

// TestAPreviewReleaseIsMarked. A preview release in the list is legitimate — a
// preview client crosses them — but reading it as a stable release is how an
// operator concludes the product shipped something it has not.
func TestAPreviewReleaseIsMarked(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
			Releases: []checker.ReleaseSummary{
				{Version: "2.1.0", IsStable: false, Changelog: "unreleased work"},
			},
		},
	}, true)

	contains(t, document, "preview", "a preview release must not read as a stable one")
}

// TestAServerWithoutTheReleaseListStillShowsItsChangelog. Every deployment older
// than the `releases` field sends only the target's text, and the overview must
// not go blank against one — that is most of the estate on the day this ships.
func TestAServerWithoutTheReleaseListStillShowsItsChangelog(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
			Changelog: "only changelog", ReleaseNotes: "only notes",
		},
	}, true)

	contains(t, document, "only changelog", "the deprecated single changelog is the fallback")
	contains(t, document, "only notes", "so are the deprecated release notes")
}

// TestTheThreeComponentVerdictsAreDistinct.
//
// "cannot verify" is not a synonym for "unchanged". The server omits a digest it
// does not know, and an installation told "unchanged" on that basis has been
// told it is current on no evidence at all — which is the failure this whole
// redesign exists to remove.
func TestTheThreeComponentVerdictsAreDistinct(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
			Components: []checker.ComponentUpdate{
				{Component: "pentagi", OS: "linux", Arch: "amd64", Outdated: true, Verifiable: true},
				{Component: "pgvector", OS: "linux", Arch: "amd64", Outdated: false, Verifiable: true},
				{Component: "scraper", OS: "linux", Arch: "amd64", Outdated: true, Verifiable: false},
			},
		},
	}, true)

	contains(t, document, "**pentagi** (linux/amd64) — will be updated", "an outdated component changes")
	contains(t, document, "**pgvector** (linux/amd64) — unchanged", "a current component does not")
	contains(t, document, "**scraper** (linux/amd64) — cannot verify",
		"an unverifiable component must not be reported as a change, nor as unchanged")
}

// TestAnInstallationThatCannotBePlacedIsNotGivenAnOrigin. С1 and С6: nothing was
// pulled yet, or the installation was touched outside the installer. An arrow
// from an empty version would claim we know where it starts.
func TestAnInstallationThatCannotBePlacedIsNotGivenAnOrigin(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{Stack: "pentagi", HasUpdate: true, LatestVersion: "2.1.0"},
	}, true)

	contains(t, document, "→ 2.1.0", "the target is still named")
	absent(t, document, " → 2.1.0**", "there is no version to move FROM")
	contains(t, document, "Version unknown", "the reader must be told why there is no origin")
}

// TestAMixedStackSaysSo. current_version names the OLDEST component of a mixed
// stack, and printing it bare turns "the oldest thing you have is from 2.0.0"
// into "you are on 2.0.0".
func TestAMixedStackSaysSo(t *testing.T) {
	mixed := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0",
			CurrentVersionMixed: true, LatestVersion: "2.1.0",
		},
	}, true)
	contains(t, mixed, "different releases", "a mixed stack must be marked as one")

	uniform := overviewDocument([]checker.StackUpdate{
		{Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0"},
	}, true)
	absent(t, uniform, "different releases", "a uniform stack must not be marked — or the mark says nothing")
}

func TestATruncatedReleaseListSaysItWasCut(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{
			Stack: "pentagi", HasUpdate: true, CurrentVersion: "1.0.0", LatestVersion: "2.1.0",
			ReleasesTruncated: true,
			Releases:          []checker.ReleaseSummary{{Version: "2.1.0", IsStable: true, Changelog: "notes"}},
		},
	}, true)

	contains(t, document, "Older releases omitted", "a cut list that does not say so is a lie by omission")
}

// TestAFailedCheckIsNotReportedAsUpToDate is the one that matters most.
//
// The screen is reachable from a menu entry that was drawn from an EARLIER check
// result, and a later check can fail. Falling back to "everything is up to date"
// there tells the user the opposite of the truth about their installation.
func TestAFailedCheckIsNotReportedAsUpToDate(t *testing.T) {
	document := overviewDocument(nil, false)

	absent(t, document, "up to date", "a check that never completed proves nothing about the installation")
	contains(t, document, "did not complete", "the reader must be told the check failed")
}

func TestAStackWithNothingToDoIsListedSeparately(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0"},
		{Stack: "worker", HasUpdate: false, CurrentVersion: "2.1.0"},
	}, true)

	contains(t, document, "Already up to date", "a stack with nothing to do belongs in its own section")
	contains(t, document, "**worker**", "and it must still be named — silence reads as 'not installed'")

	// A stack with no update must not get a section of its own with a version
	// move, which is what a loop over the whole list would produce.
	absent(t, document, "## worker", "an unchanged stack is not a section")
}

func TestNoUpdatesAtAllSaysSo(t *testing.T) {
	document := overviewDocument([]checker.StackUpdate{
		{Stack: "pentagi", HasUpdate: false, CurrentVersion: "2.1.0"},
	}, true)

	contains(t, document, "nothing to apply", "an empty plan must be stated, not left blank")
}
