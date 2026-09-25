package checker

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"pentagi/cmd/installer/cloud"
	"pentagi/cmd/installer/loader"

	"github.com/vxcontrol/cloud/models"
	"github.com/vxcontrol/cloud/sdk"
)

func ptr[T any](v T) *T { return &v }

func digest(fill string) string {
	return strings.Repeat(fill, 64/len(fill))
}

func TestUpdates_RecordUpdateFailure_ResetsEveryVerdictAndSaysWhy(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantReason    cloud.FailureReason
		wantRetry     time.Duration
		wantRetryable bool
	}{
		{"a daily quota says when it resets", cloud.Classify(&sdk.QuotaError{
			Err: sdk.ErrQuotaExceededDaily, Scope: sdk.QuotaScopeDaily, RetryAfter: 2 * time.Hour,
		}), cloud.FailureQuotaExceeded, 2 * time.Hour, true},
		{"a tier without access is not a wait", cloud.Classify(sdk.ErrForbidden), cloud.FailureForbidden, 0, false},
		{"an unclassified error is still a failure", errors.New("dial tcp: no route to host"), cloud.FailureUnreachable, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CheckResult{
				UpdateServerAccessible: true,
				PentagiIsUpToDate:      true, GraphitiIsUpToDate: true, LangfuseIsUpToDate: true,
				ObservabilityIsUpToDate: true, InstallerIsUpToDate: true, WorkerIsUpToDate: true,
				StackUpdates: []StackUpdate{{Stack: "pentagi"}},
			}

			result.recordUpdateFailure(tt.err)

			for stack, upToDate := range map[string]bool{
				"pentagi": result.PentagiIsUpToDate, "graphiti": result.GraphitiIsUpToDate,
				"langfuse": result.LangfuseIsUpToDate, "observability": result.ObservabilityIsUpToDate,
				"installer": result.InstallerIsUpToDate, "worker": result.WorkerIsUpToDate,
			} {
				if upToDate {
					t.Errorf("stack %s kept its verdict from an earlier check after the check failed", stack)
				}
			}
			if result.UpdateServerAccessible || result.StackUpdates != nil {
				t.Errorf("reachable = %t, answer = %v after a failed check", result.UpdateServerAccessible, result.StackUpdates)
			}
			want := UpdateCheckFailure{
				Reason: string(tt.wantReason), Message: tt.err.Error(), RetryAfter: tt.wantRetry, Retryable: tt.wantRetryable,
			}
			if result.UpdateFailure == nil || *result.UpdateFailure != want {
				t.Errorf("failure = %+v, want %+v", result.UpdateFailure, want)
			}
		})
	}
}

func TestUpdates_ApplyUpdateAnswer_KeepsAStackTheAnswerOmitsUpToDate(t *testing.T) {
	result := &CheckResult{UpdateFailure: &UpdateCheckFailure{Reason: "stale"}}
	answer := &models.CheckUpdatesResponse{Updates: []models.UpdateInfo{
		{Stack: models.ProductStackPentagi, HasUpdate: true},
	}}

	result.applyUpdateAnswer(answer, reportedComponents{})

	if result.PentagiIsUpToDate {
		t.Error("pentagi has an update but is reported as up to date")
	}
	for stack, upToDate := range map[string]bool{
		"graphiti": result.GraphitiIsUpToDate, "langfuse": result.LangfuseIsUpToDate,
		"observability": result.ObservabilityIsUpToDate, "installer": result.InstallerIsUpToDate,
		"worker": result.WorkerIsUpToDate,
	} {
		if !upToDate {
			t.Errorf("stack %s is offered an update the server never mentioned", stack)
		}
	}
	if !result.UpdateServerAccessible || result.UpdateFailure != nil {
		t.Error("a successful check is recorded as a failure")
	}
}

// The Jaeger plugin is reported once per architecture, so the answer must be matched on the platform too.
func TestUpdates_ApplyUpdateAnswer_MatchesEachComponentOnItsPlatform(t *testing.T) {
	amd64Digest, arm64Digest := digest("a"), digest("b")
	reported := reportedComponents{files: []models.FileComponentInfo{
		{Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64, FileHash: ptr(amd64Digest)},
		{Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeARM64, FileHash: ptr(arm64Digest)},
	}}
	answer := &models.CheckUpdatesResponse{Updates: []models.UpdateInfo{{
		Stack: models.ProductStackObservability,
		Files: []models.FileUpdate{
			{Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				PackageName: "jaeger-clickhouse", Version: "0.13.0", FileHash: amd64Digest},
			{Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeARM64,
				PackageName: "jaeger-clickhouse", Version: "0.13.0", FileHash: arm64Digest},
		},
	}}}

	result := &CheckResult{}
	result.applyUpdateAnswer(answer, reported)

	if len(result.StackUpdates) != 1 || len(result.StackUpdates[0].Components) != 2 {
		t.Fatalf("got %+v, want one stack with two components", result.StackUpdates)
	}
	for _, component := range result.StackUpdates[0].Components {
		if component.Outdated || !component.Verifiable {
			t.Errorf("%s/%s matches what is installed: outdated=%t verifiable=%t",
				component.Component, component.Arch, component.Outdated, component.Verifiable)
		}
	}
}

func TestUpdates_ApplyUpdateAnswer_CarriesTheReleaseTextsToTheUpdateScreen(t *testing.T) {
	releasedAt := time.Date(2025, 7, 1, 12, 0, 0, 0, time.FixedZone("CEST", 2*60*60))
	answer := &models.CheckUpdatesResponse{Updates: []models.UpdateInfo{{
		Stack:          models.ProductStackPentagi,
		HasUpdate:      true,
		CurrentVersion: ptr("2.0.0"),
		LatestVersion:  ptr("2.1.0"),
		Changelog:      ptr("## Changed\n- everything"),
		ReleaseNotes:   ptr("## Notes\n- read this"),
		Releases: []models.ReleaseNote{{
			Version: "2.1.0", IsStable: true, ReleasedAt: &releasedAt, Changelog: "- changed", ReleaseNotes: "- noted",
		}},
	}}}

	result := &CheckResult{}
	result.applyUpdateAnswer(answer, reportedComponents{})

	if len(result.StackUpdates) != 1 {
		t.Fatalf("got %d stacks, want 1", len(result.StackUpdates))
	}
	got := result.StackUpdates[0]
	want := StackUpdate{
		Stack: "pentagi", HasUpdate: true, CurrentVersion: "2.0.0", LatestVersion: "2.1.0",
		Changelog: "## Changed\n- everything", ReleaseNotes: "## Notes\n- read this",
		Releases: []ReleaseSummary{{
			Version: "2.1.0", IsStable: true, ReleasedAt: "2025-07-01T10:00:00Z", Changelog: "- changed", ReleaseNotes: "- noted",
		}},
		Components: []ComponentUpdate{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stack = %+v\nwant   %+v", got, want)
	}
}

func TestUpdates_BuildImageComponentUpdate_TrustsTheServerActionOverADigestComparison(t *testing.T) {
	installed := digest("a")
	offer := func(edit func(*models.ImageUpdate)) models.ImageUpdate {
		update := models.ImageUpdate{
			Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Repository: "vxcontrol/pentagi", Tag: "latest",
		}
		edit(&update)
		return update
	}
	reason := models.ReasonTagNotPublished
	tests := []struct {
		name                 string
		offered              models.ImageUpdate
		installed            *string
		outdated, verifiable bool
		targetDigest, reason string
	}{
		{"the config digest matches what is installed", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.ConfigHash = digest("z"), ptr(installed)
		}), ptr(installed), false, true, installed, ""},
		{"the index digest matches though the config digest does not", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.ConfigHash, u.IndexHash = digest("z"), ptr(digest("b")), ptr(installed)
		}), ptr(installed), false, true, digest("b"), ""},
		{"a genuinely different image targets the offered config digest", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.ConfigHash = digest("z"), ptr(digest("c"))
		}), ptr(installed), true, true, digest("c"), ""},
		{"nothing installed to compare against", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.ConfigHash = digest("z"), ptr(digest("c"))
		}), nil, false, false, digest("c"), ""},
		{"no config digest known leaves the target empty", offer(func(u *models.ImageUpdate) {
			u.ImageHash = digest("m")
		}), ptr(installed), true, true, "", ""},
		{"current survives a digest the client never reports", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.Action = digest("m"), models.ComponentActionCurrent
		}), ptr(installed), false, true, "", ""},
		{"upgrade survives matching digests", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.ConfigHash, u.Action = digest("m"), ptr(installed), models.ComponentActionUpgrade
		}), ptr(installed), true, true, installed, ""},
		{"an image never pulled is an update", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.Action, u.PullReference = digest("b"), models.ComponentActionInstall, "vxcontrol/pentagi:latest"
		}), nil, true, false, "", ""},
		{"unknown carries a reason and nothing to act on", offer(func(u *models.ImageUpdate) {
			u.Action, u.Reason = models.ComponentActionUnknown, &reason
		}), ptr(installed), false, false, "", "tag_not_published"},
		{"current with the installed digest stays current", offer(func(u *models.ImageUpdate) {
			u.ImageHash, u.Action, u.PullReference = installed, models.ComponentActionCurrent, "vxcontrol/pentagi:latest"
		}), ptr(installed), false, true, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageComponentUpdate(tt.offered, installedArtefact{digest: tt.installed})

			if got.Outdated != tt.outdated || got.Verifiable != tt.verifiable {
				t.Errorf("outdated=%t verifiable=%t, want %t and %t", got.Outdated, got.Verifiable, tt.outdated, tt.verifiable)
			}
			if got.TargetDigest != tt.targetDigest || got.TargetVersion != "" {
				t.Errorf("target digest %q version %q, want digest %q and no version", got.TargetDigest, got.TargetVersion, tt.targetDigest)
			}
			if got.Action != string(tt.offered.Action) || got.PullReference != tt.offered.PullReference || got.Reason != tt.reason {
				t.Errorf("action %q pull %q reason %q, want %q, %q, %q", got.Action, got.PullReference, got.Reason,
					tt.offered.Action, tt.offered.PullReference, tt.reason)
			}
		})
	}
}

func TestUpdates_BuildFileComponentUpdate_ComparesTheVersionThenTheHash(t *testing.T) {
	tests := []struct {
		name                 string
		offered              models.FileUpdate
		installed            installedArtefact
		outdated, verifiable bool
	}{
		{"a newer installer build", models.FileUpdate{Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin,
			Arch: models.ArchTypeARM64, PackageName: "installer", Version: "2.1.0", FileHash: digest("z")},
			installedArtefact{version: ptr("2.0.0")}, true, true},
		{"the same installer build", models.FileUpdate{Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin,
			Arch: models.ArchTypeARM64, PackageName: "installer", Version: "2.0.0", FileHash: digest("z")},
			installedArtefact{version: ptr("2.0.0")}, false, true},
		{"no version reported falls back to the hash", models.FileUpdate{Component: models.ComponentTypeJaegerClickhouse,
			OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64, PackageName: "jaeger-clickhouse", Version: "0.13.0", FileHash: digest("a")},
			installedArtefact{digest: ptr(digest("a"))}, false, true},
		{"nothing installed to compare against", models.FileUpdate{Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin,
			Arch: models.ArchTypeARM64, PackageName: "installer", Version: "2.1.0", FileHash: digest("z")},
			installedArtefact{}, false, false},
		{"an upgrade names the version to download and the hash to verify", models.FileUpdate{
			Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Version: "0.13.0", FileHash: digest("plugin-0-13-0"), Action: models.ComponentActionUpgrade},
			installedArtefact{digest: ptr(digest("plugin-0-12-0"))}, true, true},
		{"current survives a version spelled differently", models.FileUpdate{Component: models.ComponentTypeJaegerClickhouse,
			OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64, Version: "v0.13.0", Action: models.ComponentActionCurrent},
			installedArtefact{version: ptr("0.13.0")}, false, true},
		{"a file never installed is an update", models.FileUpdate{Component: models.ComponentTypeInstaller, OS: models.OSTypeLinux,
			Arch: models.ArchTypeAMD64, Action: models.ComponentActionInstall, PackageName: "installer", Version: "2.1.0",
			FileHash: digest("b")}, installedArtefact{}, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFileComponentUpdate(tt.offered, tt.installed)

			if got.Outdated != tt.outdated || got.Verifiable != tt.verifiable {
				t.Errorf("outdated=%t verifiable=%t, want %t and %t", got.Outdated, got.Verifiable, tt.outdated, tt.verifiable)
			}
			if got.TargetVersion != tt.offered.Version || got.TargetDigest != tt.offered.FileHash || got.Action != string(tt.offered.Action) {
				t.Errorf("target version %q digest %q action %q, want the offered %q, %q, %q", got.TargetVersion, got.TargetDigest,
					got.Action, tt.offered.Version, tt.offered.FileHash, tt.offered.Action)
			}
		})
	}
}

func TestUpdates_UpdateStrategy_FallsBackToPreviewForAnythingUnknown(t *testing.T) {
	tests := []struct {
		name, configured string
		want             models.UpdateStrategy
	}{
		{"unset follows the channels already in use", "", models.UpdateStrategyPreview},
		{"explicit stable", "stable", models.UpdateStrategyStable},
		{"case is normalised", "NIGHTLY", models.UpdateStrategyNightly},
		{"surrounding whitespace", "  preview  ", models.UpdateStrategyPreview},
		{"a typo falls back rather than failing", "bleeding-edge", models.UpdateStrategyPreview},
	}

	for _, tt := range tests {
		appState := &mockState{vars: map[string]loader.EnvVar{"UPDATE_STRATEGY": {Value: tt.configured}}}
		if got := updateStrategy(appState); got != tt.want {
			t.Errorf("%s: updateStrategy(%q) = %q, want %q", tt.name, tt.configured, got, tt.want)
		}
	}
}
