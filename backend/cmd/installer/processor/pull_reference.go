package processor

import (
	"fmt"
	"strings"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/wizard/logger"
)

// Every `image:` in the compose files is `${<COMPONENT>_IMAGE:-<default>}`, and
// this is the other half of that: before a pull, the reference the server chose
// is written into `.env` so compose resolves it.
//
// The installer does NOT pick a tag. Which tag to pull is the cloud's decision
// and varies with the update strategy, so a client that rebuilt `repository:tag`
// out of the repository and tag fields would silently override the answer with a
// guess — and land on an artefact it was never offered.
//
// The `:-` default in compose is what makes this safe to skip: an absent or
// EMPTY variable falls back to the reference the file shipped with, so an
// installation that never got an answer behaves exactly as before.

// composeImageVar is the environment variable a component's image reference goes
// into.
//
// Derived from the component name rather than kept in a table: the compose files
// use exactly this spelling, so a new component needs no second registration
// that could be forgotten — and a mismatch shows up as "the default was used",
// which is visible in `docker compose config`.
func composeImageVar(component string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(component), "-", "_")) + "_IMAGE"
}

// pullReferencesForStack collects the variables to write for one stack.
//
// Only components the server actually resolved: an `unknown` one carries no
// reference, and writing an empty value would be worse than writing nothing —
// `${VAR:-default}` treats empty as absent, so it would work, but it would also
// erase a reference a previous answer had pinned.
func pullReferencesForStack(updates []checker.StackUpdate, stack string) map[string]string {
	vars := map[string]string{}
	for _, update := range updates {
		if update.Stack != stack {
			continue
		}
		for _, component := range update.Components {
			if component.PullReference == "" || component.Component == "" {
				continue
			}
			vars[composeImageVar(component.Component)] = component.PullReference
		}
	}
	return vars
}

// applyPullReferences writes them, and reports what it wrote so an operator can
// see in the log which reference a pull was about to use.
//
// A failure here is returned rather than logged and swallowed: the whole point
// is that the pull uses the reference the server named, and a pull that quietly
// used the compose default instead is an update that reports success while
// changing nothing.
func (p *processor) applyPullReferences(stack ProductStack) error {
	if p.checker == nil || p.state == nil {
		return nil
	}
	vars := pullReferencesForStack(p.checker.StackUpdates, string(stack))
	if len(vars) == 0 {
		return nil
	}

	names := make([]string, 0, len(vars))
	for name, reference := range vars {
		names = append(names, fmt.Sprintf("%s=%s", name, reference))
	}
	logger.Log("[processor] PULL REFERENCES for %s: %s", stack, strings.Join(names, " "))

	// WriteVars and not SetVars. SetVars STAGES a change: it lands in the installer's own
	// state file and only reaches `.env` when the user applies their changes. But the pull
	// that follows is `docker compose --env-file <.env>`, so a staged variable is one
	// compose has never heard of — it falls back to the default the compose file ships
	// with, pulls that, and the update reports success having fetched something the server
	// never offered. The whole mechanism was inert until this line said WriteVars.
	if err := p.state.WriteVars(vars); err != nil {
		return fmt.Errorf("failed to write image references for %s: %w", stack, err)
	}
	return nil
}
