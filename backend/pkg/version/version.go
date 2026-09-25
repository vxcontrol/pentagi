package version

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

// PackageName is service name or binary name
var PackageName string

// PackageVer is semantic version of the binary
var PackageVer string

// PackageRev is revision of the binary build
var PackageRev string

// Edition names the distribution this binary belongs to and is the prerelease
// identifier every version string carries. Change it in one place: the
// Enterprise pipeline stamps EditionEnterprise over it with
// -X pentagi/pkg/version.Edition=ee, which is why it is a var.
var Edition = EditionCommunity

const (
	// EditionCommunity is the open-source distribution built from GitHub.
	EditionCommunity = "ce"
	// EditionEnterprise is the distribution built from the GitLab repository.
	EditionEnterprise = "ee"

	// revisionPrefix keeps the revision an ALPHANUMERIC semver identifier. A bare
	// hex revision that happens to be all digits with a leading zero -- 0123456,
	// about one build in 268 -- is not a legal numeric identifier, and the whole
	// version would stop parsing. The letter says "hex" and removes the class.
	revisionPrefix = "h"

	// legacyDevelopVer is what unreleased builds used to pass in PACKAGE_VER.
	// Accepted as "no release" so an old pipeline does not stamp a release
	// literally named "develop".
	legacyDevelopVer = "develop"
)

// GetBinaryVersion names the running build in a form the downstream services
// can parse. The edition is always present and always a prerelease identifier,
// and the revision is a second identifier after a dot rather than another
// hyphenated word, so a stamped build is valid semver:
//
//	(no release, no revision)  ce
//	(no release)               ce.h87ac00f
//	release tag                2.1.0-ce
//	release tag and revision   2.1.0-ce.h87ac00f
//
// The first two are deliberately not semver: there is no release behind them,
// and callers that need a number normalise them away rather than be handed one
// that was invented here.
func GetBinaryVersion() string {
	edition := editionName()
	if releaseVer() == "" {
		if PackageRev == "" {
			return edition
		}

		return fmt.Sprintf("%s.%s%s", edition, revisionPrefix, PackageRev)
	}

	if PackageRev == "" {
		return fmt.Sprintf("%s-%s", releaseVer(), edition)
	}

	return fmt.Sprintf("%s-%s.%s%s", releaseVer(), edition, revisionPrefix, PackageRev)
}

// editionName falls back to the community marker so a build that blanked the
// variable still reports an edition rather than an empty prerelease.
func editionName() string {
	switch Edition {
	case EditionCommunity, EditionEnterprise:
		return Edition
	default:
		return EditionCommunity
	}
}

// releaseVer is PackageVer when it names a release, and "" otherwise. An
// edition marker or the legacy branch name is not a release: a pipeline that
// passes one is saying there is no tag behind this build.
func releaseVer() string {
	switch PackageVer {
	case "", EditionCommunity, EditionEnterprise, legacyDevelopVer:
		return ""
	default:
		return PackageVer
	}
}

// GetBuildRevision names the running build: the revision the pipeline stamped, else the first
// 12 hex digits of the binary's sha256, else "unknown" -- never empty.
func GetBuildRevision() string {
	return buildRevision(PackageRev, BinaryHash())
}

func buildRevision(rev, binaryHash string) string {
	switch {
	case rev != "":
		return rev
	case binaryHash == "":
		return "unknown"
	case len(binaryHash) < 12:
		return binaryHash
	default:
		return binaryHash[:12]
	}
}

// IsDevelopMode reports that no release tag stands behind this build.
func IsDevelopMode() bool {
	return releaseVer() == ""
}

func GetBinaryName() string {
	if PackageName != "" {
		return PackageName
	}
	return "pentagi"
}

// BinaryHash is the sha256 of the executable this process was started from.
//
// It answers what the version string cannot: two builds cut from the same branch
// carry the same "ce" or the same "0.9.3-ce.h87ac00f", and outside a release that
// is the normal case rather than the exception. Anything describing this process to
// somebody else needs it, and the value belongs here beside GetBinaryVersion for the
// same reason — both answer "which build is this".
//
// Computed once. Reading a hundred megabytes is not free, and the file this process
// was started from cannot become a different file while it runs: replacing the
// binary on disk produces a new file that this process is not executing, so hashing
// again would report something that is not running.
//
// Empty when it cannot be read — a hash that could not be computed is not a fact
// worth inventing, and every reader of this treats an empty value as "unknown".
func BinaryHash() string {
	binaryHashOnce.Do(func() {
		path, err := os.Executable()
		if err != nil {
			return
		}
		file, err := os.Open(path)
		if err != nil {
			return
		}
		defer file.Close()

		digest := sha256.New()
		if _, err := io.Copy(digest, file); err != nil {
			return
		}
		binaryHash = hex.EncodeToString(digest.Sum(nil))
	})
	return binaryHash
}

var (
	binaryHashOnce sync.Once
	binaryHash     string
)
