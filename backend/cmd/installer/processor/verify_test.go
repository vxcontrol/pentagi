package processor

import (
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func componentDigest(fill string) string {
	return strings.Repeat(fill, 64/len(fill))
}

func TestVerify_CaptureStackTargets_RecordsWhatEachComponentShouldBecome(t *testing.T) {
	for _, tc := range []struct {
		name      string
		component checker.ComponentUpdate
		want      string
	}{
		{"an outdated component should become the offered build", checker.ComponentUpdate{
			Verifiable: true, Outdated: true, CurrentVersion: componentDigest("a"), TargetDigest: componentDigest("b"),
		}, componentDigest("b")},
		// Under the stable strategy the answer lists every artefact of the release.
		{"a current component should stay what it runs", checker.ComponentUpdate{
			Verifiable: true, Outdated: false, CurrentVersion: componentDigest("c"),
		}, componentDigest("c")},
		// Promoting the installed value into a target would turn "cannot tell" into a claim.
		{"an unverifiable component carries no target", checker.ComponentUpdate{
			Verifiable: false, CurrentVersion: componentDigest("d"),
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.component.Component, tc.component.OS, tc.component.Arch = "pentagi", "linux", "amd64"
			before := &checker.CheckResult{StackUpdates: []checker.StackUpdate{
				{Stack: "langfuse", Components: []checker.ComponentUpdate{{Component: "langfuse-web"}}},
				{Stack: "pentagi", HasUpdate: true, Components: []checker.ComponentUpdate{tc.component}},
			}}

			assert.Equal(t, []componentTarget{{name: "pentagi/linux/amd64", digest: tc.want}},
				captureStackTargets(before, "pentagi"))
		})
	}
}

func TestVerify_StackHasNoUpdate_TrustsOnlyASuccessfulFreshAnswer(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result *checker.CheckResult
		want   bool
	}{
		{"the answer says nothing is pending",
			&checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false}}}, true},
		{"the answer still offers an update",
			&checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: true}}}, false},
		// Silence about a stack is not evidence that something is pending for it.
		{"the stack is not mentioned",
			&checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "langfuse"}}}, true},
		// Without a fresh answer a difference cannot be excused as the registry moving ahead.
		{"the refreshed check failed", &checker.CheckResult{
			UpdateFailure: &checker.UpdateCheckFailure{Reason: "unreachable"},
			StackUpdates:  []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false}},
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, stackHasNoUpdate(tc.result, "pentagi"))
		})
	}
}

// A moving tag can point somewhere new between the answer and the pull.
func TestVerify_VerifyStackUpdate_TellsAMatchANewerBuildAMismatchAndTheUnverifiableApart(t *testing.T) {
	offered, arrived := componentDigest("a"), componentDigest("b")
	for _, tc := range []struct {
		name         string
		stillOffered bool
		installed    string
		targets      []componentTarget
		want         []string
		wantNot      string
	}{
		{name: "the registry moved ahead", installed: arrived,
			targets: []componentTarget{{name: "pentagi/linux/amd64", digest: offered}},
			want:    []string{"pentagi/linux/amd64: newer than the offered build"}},
		{name: "the update did not take", stillOffered: true, installed: arrived,
			targets: []componentTarget{{name: "pentagi/linux/amd64", digest: offered}},
			want:    []string{"pentagi/linux/amd64: expected " + offered + ", got " + arrived}},
		{name: "a match and an unverifiable component", installed: offered,
			targets: []componentTarget{
				{name: "pentagi/linux/amd64", digest: offered},
				{name: "scraper/linux/amd64", digest: ""},
			},
			want: []string{
				"pentagi/linux/amd64: matches the published build",
				"scraper/linux/amd64: no published digest",
			},
			wantNot: "do not match"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &processor{checker: &checker.CheckResult{
				StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: tc.stillOffered}},
				InstalledComponents: []checker.InstalledComponent{
					{Component: "pentagi", OS: "linux", Arch: "amd64", Digest: tc.installed},
					{Component: "scraper", OS: "linux", Arch: "amd64", Digest: componentDigest("c")},
				},
			}}
			state := testOperationState(t)

			p.verifyStackUpdate(ProductStackPentagi, tc.targets, state)

			reported := state.output.String()
			for _, want := range tc.want {
				assert.Contains(t, reported, want)
			}
			if tc.wantNot != "" {
				assert.NotContains(t, reported, tc.wantNot, "an unverifiable component was counted as a mismatch")
			}
		})
	}
}

// The installer operation downloads a build and leaves the running binary alone.
func TestVerify_VerifiableUpdateStacks_LeaveOutOnlyTheInstaller(t *testing.T) {
	require.False(t, verifiableUpdateStacks[ProductStackInstaller])
	for _, stack := range []ProductStack{
		ProductStackPentagi, ProductStackGraphiti, ProductStackLangfuse, ProductStackObservability, ProductStackWorker,
	} {
		assert.True(t, verifiableUpdateStacks[stack], "stack %s replaces what is installed but is never verified", stack)
	}
}
