package processor

import (
	"fmt"

	"pentagi/cmd/installer/checker"
)

// Verifying an update means answering one question: did the artefacts that are now
// installed become the ones the server offered? It is asked in two steps.
//
// The first is local. Before the update, the server named a target digest for each
// artefact; afterwards, the components are observed again and compared against those
// targets. The second step is the fresh update check that follows, and it is what actually
// decides whether the menus still offer an update — because only the server knows whether
// something newer appeared in the meantime.
//
// The two disagree in one legitimate case. A moving tag such as "latest" points at whatever
// the registry holds now, and the registry can move between the answer and the pull. Then
// what is installed matches no target and yet is not wrong — it is newer. The fresh check
// settles it: if the server no longer offers an update for the stack, whatever is installed
// is current, and a digest that differs from the old target means the registry moved ahead
// rather than that the update failed.

// verifiableUpdateStacks are the stacks whose update replaces what is installed, and so
// the only ones a comparison afterwards can say anything about. The installer is absent
// deliberately: its operation downloads a build and stops there, leaving the running binary
// as it was, so a comparison would report the intended outcome as a mismatch.
var verifiableUpdateStacks = map[ProductStack]bool{
	ProductStackPentagi:       true,
	ProductStackGraphiti:      true,
	ProductStackLangfuse:      true,
	ProductStackObservability: true,
	ProductStackWorker:        true,
}

// componentTarget is the digest the server named for one artefact, captured before the
// update so it survives the answer being replaced by a newer one.
type componentTarget struct {
	name string
	// digest is empty when the answer named no digest for this artefact. That is "nothing
	// to compare", not a mismatch — the server omits what it does not know.
	digest string
}

// captureStackTargets records what the current answer says a stack's artefacts should
// become. It must run before the update replaces the answer.
func captureStackTargets(result *checker.CheckResult, stack string) []componentTarget {
	if result == nil {
		return nil
	}

	for _, update := range result.StackUpdates {
		if update.Stack != stack {
			continue
		}

		targets := make([]componentTarget, 0, len(update.Components))
		for _, component := range update.Components {
			targets = append(targets, componentTarget{
				name: componentIdentity(component.Component, component.OS, component.Arch),
				// TargetDigest is the offered identity in the same form the installed
				// side reports it — a config digest for images, a file hash for files.
				digest: componentTargetDigest(component),
			})
		}

		return targets
	}

	return nil
}

// componentTargetDigest is the value an updated artefact should end up carrying, in the
// same identity the installed side reports: a config digest for an image, a file hash for
// a file. TargetDigest carries exactly that for both kinds.
//
// It used to read TargetVersion, which held a digest for images and a VERSION for files —
// so the Jaeger plugin, which has a hash and no version, was checked as "expected 0.13.0,
// got a423c6…" and reported a mismatch after every successful update.
func componentTargetDigest(component checker.ComponentUpdate) string {
	if !component.Verifiable {
		return ""
	}
	if component.Outdated {
		// The check concluded this differs from what is installed, so the target is what
		// the answer offered rather than what was there.
		return component.TargetDigest
	}
	// Already current: the installed value is the target.
	return component.CurrentVersion
}

func componentIdentity(component, os, arch string) string {
	return component + "/" + os + "/" + arch
}

// verifyStackUpdate compares what is installed now against what was offered before, and
// writes the outcome to the terminal.
//
// It never fails the operation. The update itself either worked or reported an error; this
// is a report about the result, and turning a registry that moved ahead into a failed
// update would be worse than saying nothing.
func (p *processor) verifyStackUpdate(
	stack ProductStack, targets []componentTarget, state *operationState,
) {
	if len(targets) == 0 {
		return
	}

	p.appendLog(fmt.Sprintf(MsgVerifyingUpdatedComponents, stack), stack, state)

	installed := make(map[string]string, len(p.checker.InstalledComponents))
	for _, component := range p.checker.InstalledComponents {
		identity := componentIdentity(component.Component, component.OS, component.Arch)
		installed[identity] = firstNonEmptyString(component.Digest, component.Version)
	}

	// The fresh check that just ran is the authority on whether anything is still pending.
	stackCurrent := stackHasNoUpdate(p.checker, string(stack))

	mismatches := 0
	for _, target := range targets {
		actual := installed[target.name]

		switch {
		case target.digest == "" || actual == "":
			p.appendLog(fmt.Sprintf(MsgComponentNotVerifiable, target.name), stack, state)

		case actual == target.digest:
			p.appendLog(fmt.Sprintf(MsgComponentUpToDate, target.name), stack, state)

		case stackCurrent:
			// Different from what was offered, yet the server says there is nothing left
			// to apply: the registry moved on between the answer and the pull.
			p.appendLog(fmt.Sprintf(MsgComponentNewerThanOffered, target.name), stack, state)

		default:
			mismatches++
			p.appendLog(fmt.Sprintf(MsgComponentMismatch, target.name, target.digest, actual), stack, state)
		}
	}

	if mismatches == 0 {
		p.appendLog(fmt.Sprintf(MsgVerificationPassed, stack), stack, state)
		return
	}
	p.appendLog(fmt.Sprintf(MsgVerificationMismatch, mismatches, stack), stack, state)
}

// stackHasNoUpdate reports whether the freshest answer leaves nothing to apply for a stack.
// A stack the answer does not mention counts as current: the server reports on what it was
// asked about, and silence is not evidence of an update.
func stackHasNoUpdate(result *checker.CheckResult, stack string) bool {
	if result == nil || result.UpdateFailure != nil {
		return false
	}

	for _, update := range result.StackUpdates {
		if update.Stack == stack {
			return !update.HasUpdate
		}
	}

	return true
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
