package checker

import (
	"errors"
	"strings"
	"sync"
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

// TestUpdateFailureResetsEveryStackFlag is the regression for the branch that used to miss
// one flag.
//
// The flags are read as "up to date, nothing to offer". A flag left at true from an earlier
// successful check keeps saying that after a check that never happened, and the interface
// presents a stale verdict as a current one — with nothing on screen to suggest it is old.
func TestUpdateFailureResetsEveryStackFlag(t *testing.T) {
	result := &CheckResult{
		UpdateServerAccessible:  true,
		PentagiIsUpToDate:       true,
		GraphitiIsUpToDate:      true,
		LangfuseIsUpToDate:      true,
		ObservabilityIsUpToDate: true,
		InstallerIsUpToDate:     true,
		WorkerIsUpToDate:        true,
		StackUpdates:            []StackUpdate{{Stack: "pentagi"}},
	}

	result.recordUpdateFailure(errors.New("dial tcp: no route to host"))

	flags := map[string]bool{
		"pentagi":       result.PentagiIsUpToDate,
		"graphiti":      result.GraphitiIsUpToDate,
		"langfuse":      result.LangfuseIsUpToDate,
		"observability": result.ObservabilityIsUpToDate,
		"installer":     result.InstallerIsUpToDate,
		"worker":        result.WorkerIsUpToDate,
	}
	for stack, upToDate := range flags {
		if upToDate {
			t.Errorf("stack %s kept its verdict from an earlier check after the check failed", stack)
		}
	}

	if result.UpdateServerAccessible {
		t.Error("the server is recorded as reachable after a failed check")
	}
	if result.StackUpdates != nil {
		t.Error("the previous answer is still on the result")
	}
	if result.UpdateFailure == nil {
		t.Fatal("no failure was recorded")
	}
}

// TestUpdateFailureCarriesWhatTheUserCanActOn: a spent quota and a broken proxy call for
// opposite reactions, so the reason and the cooldown have to survive into the result.
func TestUpdateFailureCarriesWhatTheUserCanActOn(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantReason    cloud.FailureReason
		wantRetry     time.Duration
		wantRetryable bool
	}{
		{
			name:          "daily quota says when it resets",
			err:           cloud.Classify(&sdk.QuotaError{Err: sdk.ErrQuotaExceededDaily, Scope: sdk.QuotaScopeDaily, RetryAfter: 2 * time.Hour}),
			wantReason:    cloud.FailureQuotaExceeded,
			wantRetry:     2 * time.Hour,
			wantRetryable: true,
		},
		{
			name:          "a tier without access is not a wait",
			err:           cloud.Classify(sdk.ErrForbidden),
			wantReason:    cloud.FailureForbidden,
			wantRetryable: false,
		},
		{
			name:          "an unclassified error is still a failure",
			err:           errors.New("something else entirely"),
			wantReason:    cloud.FailureUnreachable,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &CheckResult{}
			result.recordUpdateFailure(tt.err)

			failure := result.UpdateFailure
			if failure == nil {
				t.Fatal("no failure recorded")
			}
			if failure.Reason != string(tt.wantReason) {
				t.Errorf("reason = %q, want %q", failure.Reason, tt.wantReason)
			}
			if failure.RetryAfter != tt.wantRetry {
				t.Errorf("retry after = %v, want %v", failure.RetryAfter, tt.wantRetry)
			}
			if failure.Retryable != tt.wantRetryable {
				t.Errorf("retryable = %t, want %t", failure.Retryable, tt.wantRetryable)
			}
			if failure.Message == "" {
				t.Error("the failure carries no message for the user")
			}
		})
	}
}

// TestStackAbsentFromTheAnswerStaysUpToDate: the server reports on what it was asked
// about, so silence about a stack is not evidence of an update. Offering one anyway sends
// the user through an update that changes nothing.
func TestStackAbsentFromTheAnswerStaysUpToDate(t *testing.T) {
	result := &CheckResult{}
	answer := &models.CheckUpdatesResponse{
		Updates: []models.UpdateInfo{
			{Stack: models.ProductStackPentagi, HasUpdate: true},
		},
	}

	result.applyUpdateAnswer(answer, reportedComponents{})

	if result.PentagiIsUpToDate {
		t.Error("pentagi has an update but is reported as up to date")
	}
	for stack, upToDate := range map[string]bool{
		"graphiti":      result.GraphitiIsUpToDate,
		"langfuse":      result.LangfuseIsUpToDate,
		"observability": result.ObservabilityIsUpToDate,
		"installer":     result.InstallerIsUpToDate,
		"worker":        result.WorkerIsUpToDate,
	} {
		if !upToDate {
			t.Errorf("stack %s is offered an update the server never mentioned", stack)
		}
	}
	if !result.UpdateServerAccessible || result.UpdateFailure != nil {
		t.Error("a successful check is recorded as a failure")
	}
}

// TestComponentIsOutdatedOnlyWhenTheDigestsActuallyDiffer covers the distinction the
// update screen is built on. Under the stable strategy the answer lists every artefact of
// the release that matches something reported, whether or not it differs — so membership
// is not a work list, and a digest comparison is the only thing that says otherwise.
func TestComponentIsOutdatedOnlyWhenTheDigestsActuallyDiffer(t *testing.T) {
	installedDigest := digest("a")

	tests := []struct {
		name           string
		offered        models.ImageUpdate
		installed      installedArtefact
		wantOutdated   bool
		wantVerifiable bool
	}{
		{
			name: "config digest matches what is installed",
			offered: models.ImageUpdate{
				Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				Repository: "vxcontrol/pentagi", Tag: "latest",
				ImageHash: digest("z"), ConfigHash: ptr(installedDigest),
			},
			installed:      installedArtefact{digest: ptr(installedDigest)},
			wantOutdated:   false,
			wantVerifiable: true,
		},
		{
			// The server names an artefact by any of three digest identities, and a
			// client compares whichever one its daemon exposes. Matching any single one
			// is enough to conclude "I already have this".
			name: "index digest matches even though the config digest does not",
			offered: models.ImageUpdate{
				Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				Repository: "vxcontrol/pentagi", Tag: "latest",
				ImageHash: digest("z"), ConfigHash: ptr(digest("b")), IndexHash: ptr(installedDigest),
			},
			installed:      installedArtefact{digest: ptr(installedDigest)},
			wantOutdated:   false,
			wantVerifiable: true,
		},
		{
			name: "a genuinely different image",
			offered: models.ImageUpdate{
				Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				Repository: "vxcontrol/pentagi", Tag: "latest",
				ImageHash: digest("z"), ConfigHash: ptr(digest("c")),
			},
			installed:      installedArtefact{digest: ptr(installedDigest)},
			wantOutdated:   true,
			wantVerifiable: true,
		},
		{
			name: "nothing installed to compare against",
			offered: models.ImageUpdate{
				Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				Repository: "vxcontrol/pentagi", Tag: "latest",
				ImageHash: digest("z"), ConfigHash: ptr(digest("c")),
			},
			installed:      installedArtefact{},
			wantOutdated:   false,
			wantVerifiable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageComponentUpdate(tt.offered, tt.installed)

			if got.Outdated != tt.wantOutdated {
				t.Errorf("outdated = %t, want %t", got.Outdated, tt.wantOutdated)
			}
			if got.Verifiable != tt.wantVerifiable {
				t.Errorf("verifiable = %t, want %t", got.Verifiable, tt.wantVerifiable)
			}
		})
	}
}

// TestAnOutdatedImageCarriesTheTargetVerificationNeeds.
//
// The post-update verification compares what `docker inspect` reports after the
// pull (the config digest) against ComponentUpdate.TargetVersion. Without this
// field every outdated image had a target of "", was routed to "not verifiable",
// and the verification printed its pass line for an update that may not have
// taken — while the same CheckResult still said HasUpdate for the stack.
//
// The only test of that comparison used to hand-write a ComponentUpdate with
// TargetVersion filled — a shape the real producer could not emit.
func TestAnOutdatedImageCarriesTheTargetVerificationNeeds(t *testing.T) {
	installedDigest, offeredConfig := digest("a"), digest("b")

	outdated := buildImageComponentUpdate(models.ImageUpdate{
		Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
		Repository: "vxcontrol/pentagi", Tag: "latest",
		ImageHash: digest("m"), ConfigHash: ptr(offeredConfig),
	}, installedArtefact{digest: ptr(installedDigest)})

	if !outdated.Outdated {
		t.Fatalf("fixture broken: the digests differ, the component must be outdated")
	}
	if outdated.TargetDigest != offeredConfig {
		t.Errorf("TargetDigest = %q, want the offered config digest %q — the installed side of"+
			" verification is docker inspect's Id, and only the config digest can ever equal it",
			outdated.TargetDigest, offeredConfig)
	}
	if outdated.TargetVersion != "" {
		t.Errorf("TargetVersion = %q for an image; the field means a VERSION everywhere else,"+
			" and an image has none of its own in the answer", outdated.TargetVersion)
	}

	// No config digest known server-side: the target stays empty so verification
	// degrades to "cannot verify" rather than comparing against nothing.
	unknowable := buildImageComponentUpdate(models.ImageUpdate{
		Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
		Repository: "vxcontrol/pentagi", Tag: "latest",
		ImageHash: digest("m"),
	}, installedArtefact{digest: ptr(installedDigest)})
	if unknowable.TargetDigest != "" {
		t.Errorf("TargetDigest = %q for an image whose config digest the server does not know;"+
			" inventing one turns \"cannot verify\" into a false mismatch", unknowable.TargetDigest)
	}
}

// TestAFileCarriesBothItsVersionAndItsHash.
//
// The two are not interchangeable and used to be one field, which is how the Jaeger plugin
// — a file with a hash and NO version at all — ended up verified as "expected 0.13.0, got
// a423c6…" after every successful update.
//
// Both are needed and for different things: the download asks for the version
// (`?version=`), and the result is checked against the hash, because the hash is the only
// identity the installed side can report for a file on disk.
func TestAFileCarriesBothItsVersionAndItsHash(t *testing.T) {
	offeredHash := digest("plugin-0-13-0")

	plugin := buildFileComponentUpdate(models.FileUpdate{
		Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
		Arch: models.ArchTypeAMD64, Version: "0.13.0", FileHash: offeredHash,
		Action: models.ComponentActionUpgrade,
	}, installedArtefact{digest: ptr(digest("plugin-0-12-0"))})

	if plugin.TargetVersion != "0.13.0" {
		t.Errorf("TargetVersion = %q, want the version the download asks for", plugin.TargetVersion)
	}
	if plugin.TargetDigest != offeredHash {
		t.Errorf("TargetDigest = %q, want the offered file hash %q — verification compares"+
			" against what is on disk, and a file on disk has a hash, not a version",
			plugin.TargetDigest, offeredHash)
	}
	if plugin.TargetDigest == plugin.TargetVersion {
		t.Error("the version and the digest must not be the same value; that is the defect")
	}
}

// TestACurrentVerdictIsNotOverriddenByALocalDigestMiss.
//
// Docker exposes three sha256 identities per image and this client reports the CONFIG
// digest. An answer that names only the manifest digest therefore matches nothing locally
// — CarriesDigest returns false for an installation that is exactly up to date — and the
// digest comparison alone would mark it outdated. The server knows better: it compared
// against the digest this client sent, so its verdict is the authority.
//
// Without this the overview tells the user a component will be updated in the same breath
// as the server saying it will not, and the whole point of the redesign is that the
// interface stops making claims the answer does not support.
func TestACurrentVerdictIsNotOverriddenByALocalDigestMiss(t *testing.T) {
	installed := digest("config-of-the-newest-build")

	current := buildImageComponentUpdate(models.ImageUpdate{
		Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
		Repository: "vxcontrol/pentagi", Tag: "latest",
		// Only the manifest digest — the identity this client never reports.
		ImageHash: digest("manifest"),
		Action:    models.ComponentActionCurrent,
	}, installedArtefact{digest: ptr(installed)})

	if current.Outdated {
		t.Error("the server answered `current` and the screen would still say the component" +
			" will be updated: a digest the client cannot report is not evidence of a difference")
	}

	// The mirror case must keep working: an `upgrade` verdict is obeyed even when the
	// digests happen to match, because the server is the authority in both directions.
	upgrade := buildImageComponentUpdate(models.ImageUpdate{
		Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
		Repository: "vxcontrol/pentagi", Tag: "latest",
		ImageHash: digest("manifest"), ConfigHash: ptr(installed),
		Action: models.ComponentActionUpgrade,
	}, installedArtefact{digest: ptr(installed)})
	if !upgrade.Outdated {
		t.Error("an `upgrade` verdict must survive a digest comparison that found no difference")
	}

	// Files answer to the same rule: a `current` verdict has to survive a version string
	// that merely spells the same build differently.
	file := buildFileComponentUpdate(models.FileUpdate{
		Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
		Arch: models.ArchTypeAMD64, Version: "v0.13.0", Action: models.ComponentActionCurrent,
	}, installedArtefact{version: ptr("0.13.0")})
	if file.Outdated {
		t.Error("the server answered `current` for a file and the version strings differed only" +
			" in spelling; the screen would offer an update the server did not")
	}
}

// Files carry no digest triple: the server identifies their builds by version, and by
// hash for the one that ships without a version at all.
func TestFileIsOutdatedByVersionAndFallsBackToTheHash(t *testing.T) {
	installedDigest := digest("a")

	tests := []struct {
		name           string
		offered        models.FileUpdate
		installed      installedArtefact
		wantOutdated   bool
		wantVerifiable bool
	}{
		{
			name: "a newer installer build",
			offered: models.FileUpdate{
				Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin, Arch: models.ArchTypeARM64,
				PackageName: "installer", Version: "2.1.0", FileHash: digest("z"),
			},
			installed:      installedArtefact{version: ptr("2.0.0")},
			wantOutdated:   true,
			wantVerifiable: true,
		},
		{
			name: "the same installer build",
			offered: models.FileUpdate{
				Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin, Arch: models.ArchTypeARM64,
				PackageName: "installer", Version: "2.0.0", FileHash: digest("z"),
			},
			installed:      installedArtefact{version: ptr("2.0.0")},
			wantOutdated:   false,
			wantVerifiable: true,
		},
		{
			// The Jaeger plugin ships without a version, so the hash is all there is.
			name: "no version reported falls back to the hash",
			offered: models.FileUpdate{
				Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				PackageName: "jaeger-clickhouse", Version: "0.13.0", FileHash: installedDigest,
			},
			installed:      installedArtefact{digest: ptr(installedDigest)},
			wantOutdated:   false,
			wantVerifiable: true,
		},
		{
			name: "nothing installed to compare against",
			offered: models.FileUpdate{
				Component: models.ComponentTypeInstaller, OS: models.OSTypeDarwin, Arch: models.ArchTypeARM64,
				PackageName: "installer", Version: "2.1.0", FileHash: digest("z"),
			},
			installed:      installedArtefact{},
			wantOutdated:   false,
			wantVerifiable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFileComponentUpdate(tt.offered, tt.installed)

			if got.Outdated != tt.wantOutdated {
				t.Errorf("outdated = %t, want %t", got.Outdated, tt.wantOutdated)
			}
			if got.Verifiable != tt.wantVerifiable {
				t.Errorf("verifiable = %t, want %t", got.Verifiable, tt.wantVerifiable)
			}
		})
	}
}

// TestTheAnswerIsMatchedPerPlatform: the Jaeger plugin is reported twice, once per
// architecture. Matching the answer on the component name alone would collapse the two and
// compare one architecture's file against the other's.
func TestTheAnswerIsMatchedPerPlatform(t *testing.T) {
	amd64Digest, arm64Digest := digest("a"), digest("b")

	reported := reportedComponents{files: []models.FileComponentInfo{
		{
			Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
			Arch: models.ArchTypeAMD64, FileHash: ptr(amd64Digest),
		},
		{
			Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
			Arch: models.ArchTypeARM64, FileHash: ptr(arm64Digest),
		},
	}}
	answer := &models.CheckUpdatesResponse{Updates: []models.UpdateInfo{{
		Stack: models.ProductStackObservability, HasUpdate: false,
		Files: []models.FileUpdate{
			{
				Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
				Arch: models.ArchTypeAMD64, PackageName: "jaeger-clickhouse",
				Version: "0.13.0", FileHash: amd64Digest,
			},
			{
				Component: models.ComponentTypeJaegerClickhouse, OS: models.OSTypeLinux,
				Arch: models.ArchTypeARM64, PackageName: "jaeger-clickhouse",
				Version: "0.13.0", FileHash: arm64Digest,
			},
		},
	}}}

	result := &CheckResult{}
	result.applyUpdateAnswer(answer, reported)

	if len(result.StackUpdates) != 1 {
		t.Fatalf("got %d stacks, want 1", len(result.StackUpdates))
	}
	for _, component := range result.StackUpdates[0].Components {
		if component.Outdated {
			t.Errorf("%s/%s matches what is installed but is reported as outdated",
				component.Component, component.Arch)
		}
		if !component.Verifiable {
			t.Errorf("%s/%s should have been comparable", component.Component, component.Arch)
		}
	}
}

// TestReleaseTextsReachTheResult: the update screen shows the changelog and release notes
// before anything is applied, so they have to survive the answer, not just the verdict.
func TestReleaseTextsReachTheResult(t *testing.T) {
	answer := &models.CheckUpdatesResponse{Updates: []models.UpdateInfo{{
		Stack:          models.ProductStackPentagi,
		HasUpdate:      true,
		CurrentVersion: ptr("2.0.0"),
		LatestVersion:  ptr("2.1.0"),
		Changelog:      ptr("## Changed\n- everything"),
		ReleaseNotes:   ptr("## Notes\n- read this"),
	}}}

	result := &CheckResult{}
	result.applyUpdateAnswer(answer, reportedComponents{})

	if len(result.StackUpdates) != 1 {
		t.Fatalf("got %d stacks, want 1", len(result.StackUpdates))
	}
	stack := result.StackUpdates[0]
	if stack.CurrentVersion != "2.0.0" || stack.LatestVersion != "2.1.0" {
		t.Errorf("versions = %q → %q, want 2.0.0 → 2.1.0", stack.CurrentVersion, stack.LatestVersion)
	}
	if stack.Changelog == "" || stack.ReleaseNotes == "" {
		t.Error("the release texts were dropped, leaving the update screen with nothing to show")
	}
}

func TestUpdateStrategy(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		want       models.UpdateStrategy
	}{
		// Unset is the normal state of every installation: the variable has never been
		// written by anything. What it resolves to is therefore the strategy the whole
		// fleet is on, which is why it is `preview` — under `stable` a stack with no
		// published release is offered nothing, and "nothing is published yet" reaches
		// the user as "there are no updates".
		{name: "unset follows the channels already in use", configured: "", want: models.UpdateStrategyPreview},
		{name: "explicit stable", configured: "stable", want: models.UpdateStrategyStable},
		{name: "case is normalised", configured: "NIGHTLY", want: models.UpdateStrategyNightly},
		{name: "surrounding whitespace", configured: "  preview  ", want: models.UpdateStrategyPreview},
		// A typo must not leave an installation unable to hear about security updates.
		{name: "a typo falls back rather than failing", configured: "bleeding-edge", want: models.UpdateStrategyPreview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appState := &mockState{vars: map[string]loader.EnvVar{
				"UPDATE_STRATEGY": {Value: tt.configured},
			}}
			if got := updateStrategy(appState); got != tt.want {
				t.Errorf("updateStrategy(%q) = %q, want %q", tt.configured, got, tt.want)
			}
		})
	}
}

// TestReportedRequestPassesTheContract ties the whole gathering to the one thing that
// matters about it: a request that fails validation is answered as a server error, which
// is indistinguishable from an outage and costs the answer for every component at once.
func TestReportedRequestPassesTheContract(t *testing.T) {
	handler := &defaultCheckHandler{
		mx:       &sync.Mutex{},
		appState: &mockState{vars: map[string]loader.EnvVar{}, envPath: t.TempDir() + "/.env"},
	}
	// Every stack connected and local, so the gathering walks every branch it has.
	result := &CheckResult{
		GraphitiConnected: true, LangfuseConnected: true, ObservabilityConnected: true,
	}

	components := handler.gatherComponents(t.Context(), result)

	request := models.CheckUpdatesRequest{
		InstallerVersion: cloud.NormalizeVersion(InstallerVersion),
		InstallerOS:      models.OSTypeLinux,
		InstallerArch:    models.ArchTypeAMD64,
		Strategy:         models.UpdateStrategyStable,
		Images:           components.images,
		Files:            components.files,
	}
	if err := request.Valid(); err != nil {
		t.Fatalf("the gathered request would be rejected: %v", err)
	}
}

// TestTheServersActionDecidesWhereAComparisonCannot.
//
// Membership in the answer is not a work list — under a curated release every
// artefact matching a reported component is listed — so the installer compares
// digests. But a comparison needs two sides, and the two cases with only one are
// exactly the ones that matter most:
//
//   - `install`: nothing is pulled, so there is no installed digest and the
//     comparison leaves "not outdated" — which reads as "nothing to do" for the
//     one component that most needs doing, and hides the download button;
//   - `unknown`: there is no artefact at all, so whatever a comparison produced
//     is about nothing.
func TestTheServersActionDecidesWhereAComparisonCannot(t *testing.T) {
	installed := strings.Repeat("a", 64)
	offeredDigest := strings.Repeat("b", 64)
	reason := models.ReasonTagNotPublished

	t.Run("nothing pulled yet is an update, not a match", func(t *testing.T) {
		component := buildImageComponentUpdate(models.ImageUpdate{
			Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Action: models.ComponentActionInstall, Repository: "vxcontrol/pentagi", Tag: "latest",
			PullReference: "vxcontrol/pentagi:latest", ImageHash: offeredDigest,
		}, installedArtefact{})

		if !component.Outdated {
			t.Error("an image that was never pulled has to be offered, or the download button never appears")
		}
		if component.Action != "install" {
			t.Errorf("the server's action must survive: %q", component.Action)
		}
		if component.PullReference != "vxcontrol/pentagi:latest" {
			t.Errorf("the reference the server chose is what goes into the compose variable: %q",
				component.PullReference)
		}
	})

	t.Run("unknown carries a reason and nothing to act on", func(t *testing.T) {
		component := buildImageComponentUpdate(models.ImageUpdate{
			Component: models.ComponentTypePostgres, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Action: models.ComponentActionUnknown, Reason: &reason,
			Repository: "postgres", Tag: "15",
		}, installedArtefact{digest: &installed})

		if component.Outdated {
			t.Error("there is no artefact, so there is nothing to change to")
		}
		if component.Verifiable {
			t.Error("and nothing to verify against")
		}
		if component.Reason != "tag_not_published" {
			t.Errorf("the reason separates 'nothing to do' from 'we publish nothing for this': %q",
				component.Reason)
		}
		if component.PullReference != "" {
			t.Errorf("an unresolved component names nothing to pull: %q", component.PullReference)
		}
	})

	t.Run("current with a matching digest stays current", func(t *testing.T) {
		component := buildImageComponentUpdate(models.ImageUpdate{
			Component: models.ComponentTypePentagi, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Action: models.ComponentActionCurrent, Repository: "vxcontrol/pentagi", Tag: "latest",
			PullReference: "vxcontrol/pentagi:latest", ImageHash: installed,
		}, installedArtefact{digest: &installed})

		if component.Outdated {
			t.Error("the installed digest is the offered one")
		}
		if !component.Verifiable {
			t.Error("both sides are known, so the comparison means something")
		}
	})

	t.Run("a file that was never installed is an update too", func(t *testing.T) {
		component := buildFileComponentUpdate(models.FileUpdate{
			Component: models.ComponentTypeInstaller, OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
			Action: models.ComponentActionInstall, PackageName: "installer", Version: "2.1.0",
			FileHash: offeredDigest,
		}, installedArtefact{})

		if !component.Outdated {
			t.Error("a file that was never installed has to be offered")
		}
		if component.Action != "install" {
			t.Errorf("the server's action must survive: %q", component.Action)
		}
	})
}
