package models

import (
	"testing"

	"pentagi/cmd/installer/checker"
)

func TestUpdateOverview_OverviewDocument_TellsWhatTheUpdateChanges(t *testing.T) {
	for _, tc := range []struct {
		name           string
		updates        []checker.StackUpdate
		checkFailed    bool
		want, unwanted []string
	}{
		{
			name: "every release the installation crosses is listed with its own text",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.2.0", Changelog: "target changelog",
				Releases: []checker.ReleaseSummary{
					{Version: "2.1.0", IsStable: true, ReleasedAt: "2026-01-02T03:04:05Z", Changelog: "first changelog"},
					{Version: "2.2.0", IsStable: true, Changelog: "second changelog", ReleaseNotes: "second notes"},
				},
			}},
			want: []string{"2.0.0 → 2.2.0", "### 2.1.0", "first changelog", "second notes", "2026-01-02"},
		},
		{
			name: "an unpublished release shows no date",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
				Releases: []checker.ReleaseSummary{{Version: "2.1.0", IsStable: true, Changelog: "notes"}},
			}},
			unwanted: []string{"Released"},
		},
		{
			name: "a preview release is marked",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
				Releases: []checker.ReleaseSummary{{Version: "2.1.0", IsStable: false, Changelog: "unreleased work"}},
			}},
			want: []string{"### 2.1.0 (preview)"},
		},
		{
			name: "a server without the release list still shows its changelog and notes",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
				Changelog: "only changelog", ReleaseNotes: "only notes",
			}},
			want: []string{"only changelog", "only notes"},
		},
		{
			name: "an unverifiable component is neither a change nor unchanged",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
				Components: []checker.ComponentUpdate{
					{Component: "pentagi", OS: "linux", Arch: "amd64", Outdated: true, Verifiable: true},
					{Component: "pgvector", OS: "linux", Arch: "amd64", Outdated: false, Verifiable: true},
					{Component: "scraper", OS: "linux", Arch: "amd64", Outdated: true, Verifiable: false},
				},
			}},
			want: []string{
				"**pentagi** (linux/amd64) — will be updated",
				"**pgvector** (linux/amd64) — unchanged",
				"**scraper** (linux/amd64) — cannot verify",
			},
		},
		{
			name:     "an installation that cannot be placed is not given an origin",
			updates:  []checker.StackUpdate{{Stack: "pentagi", HasUpdate: true, LatestVersion: "2.1.0"}},
			want:     []string{"→ 2.1.0", "Version unknown"},
			unwanted: []string{" → 2.1.0**"},
		},
		{
			name: "a mixed stack is marked as one",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", CurrentVersionMixed: true, LatestVersion: "2.1.0",
			}},
			want: []string{"different releases"},
		},
		{
			name:     "a uniform stack is not marked as mixed",
			updates:  []checker.StackUpdate{{Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0"}},
			unwanted: []string{"different releases"},
		},
		{
			name: "a truncated release list says it was cut",
			updates: []checker.StackUpdate{{
				Stack: "pentagi", HasUpdate: true, CurrentVersion: "1.0.0", LatestVersion: "2.1.0", ReleasesTruncated: true,
				Releases: []checker.ReleaseSummary{{Version: "2.1.0", IsStable: true, Changelog: "notes"}},
			}},
			want: []string{"Older releases omitted"},
		},
		{
			name:        "a failed check is not reported as up to date",
			checkFailed: true,
			want:        []string{"did not complete"},
			unwanted:    []string{"up to date"},
		},
		{
			name: "a stack with nothing to do is named in its own section",
			updates: []checker.StackUpdate{
				{Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0"},
				{Stack: "worker", HasUpdate: false, CurrentVersion: "2.1.0"},
			},
			want:     []string{"Already up to date", "**worker**"},
			unwanted: []string{"## worker"},
		},
		{
			name:    "no updates at all says so",
			updates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false, CurrentVersion: "2.1.0"}},
			want:    []string{"nothing to apply"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document := overviewDocument(tc.updates, !tc.checkFailed)
			for _, want := range tc.want {
				contains(t, document, want, tc.name)
			}
			for _, unwanted := range tc.unwanted {
				absent(t, document, unwanted, tc.name)
			}
		})
	}
}
