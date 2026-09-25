package processor

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The cloud does not steer an installation back to a supported reference, and
// deliberately: it would need a copy of these compose defaults kept in step with
// them, and `postgres:15 → 16` is a major upgrade with a data migration behind
// it — not something a routine update check may cause as a side effect. It
// answers `unknown` with a reason instead.
//
// The installer is the side that CAN say it, because it ships the reference
// defaults itself. What it does with that is report the difference and let the
// user decide, separately from updating.

// composeImagePattern matches `image: ${VAR:-default}` in a compose file, which
// is the only form these files use — TestPullReference_EveryComposeImageIsParameterised
// is what keeps that true.
var composeImagePattern = regexp.MustCompile(`(?m)^\s*image:\s*\$\{([A-Z0-9_]+):-(\S+)\}\s*$`)

// composeFiles are the four the installer ships and the variables in them are
// the reference defaults of this build.
var composeFiles = []string{
	"docker-compose.yml",
	"docker-compose-langfuse.yml",
	"docker-compose-observability.yml",
	"docker-compose-graphiti.yml",
}

// ReferenceDrift is one variable whose value is not what this build of the
// installer would have used.
type ReferenceDrift struct {
	// Variable is the compose variable, Configured what `.env` says, Default what
	// this installer ships.
	Variable   string
	Configured string
	Default    string
}

func (d ReferenceDrift) String() string {
	return fmt.Sprintf("%s=%s (this build ships %s)", d.Variable, d.Configured, d.Default)
}

// composeReferenceDefaults reads the shipped defaults out of the compose files.
//
// From the files rather than from a table in Go: the files are what `docker
// compose` actually reads, so a table would be a second source of truth able to
// disagree with the thing it describes.
func composeReferenceDefaults(read func(string) ([]byte, error)) (map[string]string, error) {
	defaults := map[string]string{}
	for _, name := range composeFiles {
		body, err := read(name)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		for _, match := range composeImagePattern.FindAllStringSubmatch(string(body), -1) {
			defaults[match[1]] = match[2]
		}
	}
	if len(defaults) == 0 {
		return nil, fmt.Errorf("no parameterised images found in the compose files")
	}
	return defaults, nil
}

// referenceDrift lists the variables `.env` sets to something other than the
// shipped default.
//
// A variable the update path wrote is NOT drift: it is the server's answer, and
// under `stable` it is deliberately a more specific tag than the file's default
// (`vxcontrol/pentagi:2.3.4` against `vxcontrol/pentagi:latest`). Only a
// different REPOSITORY is a configuration this installer cannot support —
// a different tag of the same repository is the mechanism working.
func referenceDrift(defaults, configured map[string]string) []ReferenceDrift {
	var drift []ReferenceDrift
	for variable, shipped := range defaults {
		value := strings.TrimSpace(configured[variable])
		if value == "" || value == shipped {
			continue
		}
		if repositoryOf(value) == repositoryOf(shipped) {
			continue
		}
		drift = append(drift, ReferenceDrift{
			Variable: variable, Configured: value, Default: shipped,
		})
	}
	sort.Slice(drift, func(i, j int) bool { return drift[i].Variable < drift[j].Variable })
	return drift
}

// repositoryOf drops the tag. A colon separates a tag only after the last slash,
// so a registry port survives; a digest reference contributes no tag.
func repositoryOf(reference string) string {
	if at := strings.LastIndex(reference, "@"); at >= 0 {
		reference = reference[:at]
	}
	colon := strings.LastIndex(reference, ":")
	if colon < 0 || colon < strings.LastIndex(reference, "/") {
		return reference
	}
	return reference[:colon]
}

// ReferenceDrift reports the compose variables this installation sets to a
// repository this build of the installer does not ship.
//
// Exposed as a query rather than folded into the update: realignment is a
// deliberate act, not a side effect of a routine update check. What
// the interface does with the answer — show it, offer to reset those variables —
// is its decision; what this guarantees is that the difference is visible at all,
// which it was not before: the server answers `unknown` with a reason and the
// installation simply never updates that component.
func (p *processor) ReferenceDrift() ([]ReferenceDrift, error) {
	if p.files == nil || p.state == nil {
		return nil, nil
	}
	defaults, err := composeReferenceDefaults(p.files.GetContent)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(defaults))
	for variable := range defaults {
		names = append(names, variable)
	}
	vars, _ := p.state.GetVars(names)

	configured := make(map[string]string, len(vars))
	for variable, envVar := range vars {
		configured[variable] = envVar.Value
	}
	return referenceDrift(defaults, configured), nil
}
