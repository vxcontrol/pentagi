package cloud

import (
	"strings"

	"github.com/vxcontrol/cloud/models"
)

// UnknownVersion is what a build reports when its own version cannot be expressed as a
// version number — the conventional stand-in for "some build, older than anything we
// publish". Development builds are named after their branch, not a release, so this is the
// normal case outside CI.
const UnknownVersion = "0.0.0"

// latestVersion is accepted by the update server as "the newest published version" when
// asking about a package. It is meaningless as a description of the installer you are
// running, so it is never sent as one.
const latestVersion = "latest"

// NormalizeVersion converts a build version into one the update server accepts.
//
// A leading "v" is dropped, since it is a display convention rather than part of the
// number. Anything that still does not parse as a version — a branch name such as
// "develop", a bare commit hash, an empty string — becomes UnknownVersion. The original
// value is not lost: it travels in the User-Agent, where it identifies the exact build
// without having to satisfy the version grammar.
//
// The check is delegated to the shared validator rather than re-implemented here, so the
// client and the server can never disagree about what a valid version looks like.
func NormalizeVersion(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "v")

	if value == "" || value == latestVersion {
		return UnknownVersion
	}

	if err := validateSemver(value); err != nil {
		return UnknownVersion
	}

	return value
}

// validateSemver reports whether the update server would accept value as a version.
func validateSemver(value string) error {
	return models.GetValidator().Var(value, "semver")
}
