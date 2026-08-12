package processor

import (
	"runtime"
	"strings"
	"sync"
	"testing"

	"pentagi/cmd/installer/checker"
)

func componentDigest(fill string) string {
	return strings.Repeat(fill, 64/len(fill))
}

// TestTargetsAreCapturedBeforeTheAnswerIsReplaced covers the ordering the whole
// verification depends on. The update refreshes the check, so by the time the comparison
// runs the answer describes the NEW state — comparing against it would compare the result
// with itself and always agree.
func TestTargetsAreCapturedBeforeTheAnswerIsReplaced(t *testing.T) {
	before := &checker.CheckResult{StackUpdates: []checker.StackUpdate{{
		Stack:     "pentagi",
		HasUpdate: true,
		Components: []checker.ComponentUpdate{{
			Component: "pentagi", OS: "linux", Arch: "amd64",
			Verifiable: true, Outdated: true,
			CurrentVersion: componentDigest("a"),
			TargetDigest:   componentDigest("b"),
		}},
	}}}

	targets := captureStackTargets(before, "pentagi")

	if len(targets) != 1 {
		t.Fatalf("captured %d targets, want 1", len(targets))
	}
	if targets[0].digest != componentDigest("b") {
		t.Errorf("target digest = %q, want the offered build", targets[0].digest)
	}
	if targets[0].name != "pentagi/linux/amd64" {
		t.Errorf("target name = %q, want pentagi/linux/amd64", targets[0].name)
	}
}

// TestTargetOfAnAlreadyCurrentComponentIsWhatItRuns: under the stable strategy the answer
// lists every artefact of the release, including ones that already match. For those the
// target is what is installed, so the verification confirms nothing changed rather than
// expecting a change that was never due.
func TestTargetOfAnAlreadyCurrentComponentIsWhatItRuns(t *testing.T) {
	current := componentDigest("c")
	before := &checker.CheckResult{StackUpdates: []checker.StackUpdate{{
		Stack: "pentagi",
		Components: []checker.ComponentUpdate{{
			Component: "scraper", OS: "linux", Arch: "amd64",
			Verifiable: true, Outdated: false,
			CurrentVersion: current,
			TargetDigest:   "",
		}},
	}}}

	targets := captureStackTargets(before, "pentagi")

	if len(targets) != 1 || targets[0].digest != current {
		t.Fatalf("targets = %+v, want the installed digest as the target", targets)
	}
}

// TestUnverifiableComponentCarriesNoTarget: the server omits a digest it does not know, and
// inventing one would turn "cannot tell" into a mismatch report.
func TestUnverifiableComponentCarriesNoTarget(t *testing.T) {
	before := &checker.CheckResult{StackUpdates: []checker.StackUpdate{{
		Stack: "pentagi",
		Components: []checker.ComponentUpdate{{
			Component: "pgvector", OS: "linux", Arch: "amd64",
			// Something IS installed — the check knows what it is. What is missing is a
			// published digest to compare it against, and without the guard the installed
			// value would be promoted into a target, turning "cannot tell" into a claim.
			Verifiable:     false,
			CurrentVersion: componentDigest("d"),
		}},
	}}}

	targets := captureStackTargets(before, "pentagi")

	if len(targets) != 1 {
		t.Fatalf("captured %d targets, want 1", len(targets))
	}
	if targets[0].digest != "" {
		t.Errorf("target digest = %q, want empty for an artefact the server could not name", targets[0].digest)
	}
}

func TestStackHasNoUpdate(t *testing.T) {
	tests := []struct {
		name   string
		result *checker.CheckResult
		stack  string
		want   bool
	}{
		{
			name:   "the answer says nothing is pending",
			result: &checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false}}},
			stack:  "pentagi",
			want:   true,
		},
		{
			name:   "the answer still offers an update",
			result: &checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: true}}},
			stack:  "pentagi",
			want:   false,
		},
		{
			// The server reports on what it was asked about; silence about a stack is not
			// evidence that something is pending for it.
			name:   "the stack is not mentioned",
			result: &checker.CheckResult{StackUpdates: []checker.StackUpdate{{Stack: "langfuse"}}},
			stack:  "pentagi",
			want:   true,
		},
		{
			// Without a fresh answer there is no authority to appeal to, so a difference
			// cannot be excused as the registry having moved ahead.
			name: "the refreshed check failed",
			result: &checker.CheckResult{
				UpdateFailure: &checker.UpdateCheckFailure{Reason: "unreachable"},
				StackUpdates:  []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false}},
			},
			stack: "pentagi",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stackHasNoUpdate(tt.result, tt.stack); got != tt.want {
				t.Errorf("stackHasNoUpdate = %t, want %t", got, tt.want)
			}
		})
	}
}

// TestVerificationDistinguishesNewerFromWrong is the case the plan calls out: a moving tag
// can point somewhere new between the answer and the pull, so what arrives matches no
// target and is still correct. The fresh check settles it — if the server no longer offers
// an update, whatever is installed is current.
func TestVerificationDistinguishesNewerFromWrong(t *testing.T) {
	target, arrived := componentDigest("a"), componentDigest("b")

	tests := []struct {
		name          string
		stillOffered  bool
		wantSubstring string
	}{
		{
			name:          "registry moved ahead",
			stillOffered:  false,
			wantSubstring: "newer than the offered build",
		},
		{
			name:          "the update did not take",
			stillOffered:  true,
			wantSubstring: "expected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &checker.CheckResult{
				StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: tt.stillOffered}},
				InstalledComponents: []checker.InstalledComponent{{
					Component: "pentagi", OS: "linux", Arch: "amd64", Digest: arrived,
				}},
			}
			p := &processor{checker: result}
			state := &operationState{mx: &sync.Mutex{}, ctx: t.Context()}

			p.verifyStackUpdate(ProductStackPentagi,
				[]componentTarget{{name: "pentagi/linux/amd64", digest: target}}, state)

			if reported := state.output.String(); !strings.Contains(reported, tt.wantSubstring) {
				t.Errorf("verification said %q, want it to mention %q", reported, tt.wantSubstring)
			}
		})
	}
}

// TestVerificationReportsAMatchAndAnUnverifiableComponentDifferently: "matches" and "cannot
// be checked" are different facts, and reporting the second as the first would claim proof
// nobody has.
func TestVerificationReportsAMatchAndAnUnverifiableComponentDifferently(t *testing.T) {
	matched := componentDigest("a")

	result := &checker.CheckResult{
		StackUpdates: []checker.StackUpdate{{Stack: "pentagi", HasUpdate: false}},
		InstalledComponents: []checker.InstalledComponent{
			{Component: "pentagi", OS: "linux", Arch: "amd64", Digest: matched},
			{Component: "scraper", OS: "linux", Arch: "amd64", Digest: componentDigest("c")},
		},
	}
	p := &processor{checker: result}
	state := &operationState{mx: &sync.Mutex{}, ctx: t.Context()}

	p.verifyStackUpdate(ProductStackPentagi, []componentTarget{
		{name: "pentagi/linux/amd64", digest: matched},
		{name: "scraper/linux/amd64", digest: ""},
	}, state)

	text := state.output.String()
	if !strings.Contains(text, "matches the published build") {
		t.Error("the matching component was not reported as matching")
	}
	if !strings.Contains(text, "no published digest") {
		t.Error("the unverifiable component was not distinguished from a match")
	}
	if strings.Contains(text, "do not match") {
		t.Error("an unverifiable component was counted as a mismatch")
	}
}

// TestInstallerUpdatesAreNotVerifiedAgainstTheRunningBinary: the installer operation
// downloads a build and stops, deliberately leaving the running binary alone. Comparing
// them would report the intended outcome as a failure.
func TestInstallerUpdatesAreNotVerifiedAgainstTheRunningBinary(t *testing.T) {
	if verifiableUpdateStacks[ProductStackInstaller] {
		t.Error("the installer stack is verified against a binary its update never replaces")
	}
	for _, stack := range []ProductStack{
		ProductStackPentagi, ProductStackGraphiti,
		ProductStackLangfuse, ProductStackObservability, ProductStackWorker,
	} {
		if !verifiableUpdateStacks[stack] {
			t.Errorf("stack %s replaces what is installed but is never verified", stack)
		}
	}
}

func TestInstallerFileName(t *testing.T) {
	name := installerFileName("2.1.0")

	// The version is in the name so a download never lands on the installer in use.
	if !strings.Contains(name, "2.1.0") {
		t.Errorf("file name %q does not identify the build", name)
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		t.Errorf("file name %q is not executable on this platform", name)
	}
	if runtime.GOOS != "windows" && strings.HasSuffix(name, ".exe") {
		t.Errorf("file name %q carries a windows suffix", name)
	}
}
