package version

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The table is shared with scripts/check-version-rule.sh, so the shell and the binary agree on every case.
func TestVersion_GetBinaryVersion_FollowsTheSharedTable(t *testing.T) {
	data, err := os.ReadFile("testdata/version_strings.tsv")
	require.NoError(t, err)
	t.Cleanup(func() { PackageVer, PackageRev = "", "" })

	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		fields := strings.Split(line, "\t")
		require.Len(t, fields, 3, "malformed row %q", line)

		PackageVer, PackageRev = fields[0], fields[1]
		assert.Equal(t, fields[2], GetBinaryVersion(), "PackageVer=%q PackageRev=%q", fields[0], fields[1])
	}
}

func TestVersion_IsDevelopMode_TrueOnlyWithoutARelease(t *testing.T) {
	t.Cleanup(func() { PackageVer = "" })

	// An edition marker and the legacy branch name mean "no tag behind this
	// build", so they are develop mode exactly as an empty value is.
	for ver, want := range map[string]bool{
		"": true, "develop": true, "ce": true, "ee": true,
		"1.0.0": false, "2.1.0": false,
	} {
		PackageVer = ver
		assert.Equal(t, want, IsDevelopMode(), "PackageVer=%q", ver)
	}
}

func TestVersion_GetBinaryVersion_CarriesTheEditionItWasBuiltFor(t *testing.T) {
	t.Cleanup(func() { PackageVer, PackageRev, Edition = "", "", EditionCommunity })

	for _, tc := range []struct {
		edition, ver, rev, want string
	}{
		{edition: EditionEnterprise, want: "ee"},
		{edition: EditionEnterprise, rev: "abc1234", want: "ee.habc1234"},
		{edition: EditionEnterprise, ver: "2.1.0", want: "2.1.0-ee"},
		{edition: EditionEnterprise, ver: "2.1.0", rev: "abc1234", want: "2.1.0-ee.habc1234"},
		// An edition marker in PackageVer is never a release, whichever build it is.
		{edition: EditionEnterprise, ver: "ce", rev: "abc1234", want: "ee.habc1234"},
		// A build that blanked or mistyped the variable still names an edition
		// rather than emitting an empty prerelease.
		{edition: "", ver: "2.1.0", want: "2.1.0-ce"},
		{edition: "nightly", ver: "2.1.0", want: "2.1.0-ce"},
	} {
		Edition, PackageVer, PackageRev = tc.edition, tc.ver, tc.rev
		assert.Equal(t, tc.want, GetBinaryVersion(),
			"Edition=%q PackageVer=%q PackageRev=%q", tc.edition, tc.ver, tc.rev)
	}
}

// The dot before the revision is what makes a stamped release parse: the update
// server and the downstream services run the value through this same validator.
func TestVersion_GetBinaryVersion_ReleaseBuildsAreValidSemver(t *testing.T) {
	t.Cleanup(func() { PackageVer, PackageRev, Edition = "", "", EditionCommunity })

	validate := validator.New()
	for _, tc := range []struct {
		edition, ver, rev string
		wantSemver        bool
	}{
		{edition: EditionCommunity, ver: "2.1.0", wantSemver: true},
		{edition: EditionCommunity, ver: "2.1.0", rev: "abc1234", wantSemver: true},
		{edition: EditionEnterprise, ver: "2.1.0", rev: "abc1234", wantSemver: true},
		{edition: EditionCommunity, ver: "0.9.3", rev: "87ac00f", wantSemver: true},
		// The reason the revision carries a letter: a bare 0123456 would be an
		// all-digit identifier with a leading zero, which semver rejects outright.
		{edition: EditionCommunity, ver: "2.1.0", rev: "0123456", wantSemver: true},
		{edition: EditionEnterprise, ver: "2.1.0", rev: "0000000", wantSemver: true},
		// No release behind the build: there is no number to be valid, and the
		// caller normalises it away rather than be handed an invented one.
		{edition: EditionCommunity},
		{edition: EditionCommunity, rev: "abc1234"},
	} {
		Edition, PackageVer, PackageRev = tc.edition, tc.ver, tc.rev
		got := GetBinaryVersion()
		err := validate.Var(got, "semver")
		if tc.wantSemver {
			assert.NoError(t, err, "%q must parse as semver", got)

			continue
		}
		assert.Error(t, err, "%q is not a release and must not pass as one", got)
	}
}

// A stamped release is a PRERELEASE of the number it names, and semver orders a
// prerelease below its release. Anything asking "is this at least 2.1.0" must
// compare the release part, not the whole string.
func TestVersion_GetBinaryVersion_StampsTheEditionAsAPrerelease(t *testing.T) {
	t.Cleanup(func() { PackageVer, PackageRev, Edition = "", "", EditionCommunity })

	for _, tc := range []struct{ edition, ver, rev, release, prerelease string }{
		{edition: EditionCommunity, ver: "2.1.0", release: "2.1.0", prerelease: "ce"},
		{edition: EditionCommunity, ver: "2.1.0", rev: "abc1234", release: "2.1.0", prerelease: "ce.habc1234"},
		{edition: EditionEnterprise, ver: "0.9.3", rev: "87ac00f", release: "0.9.3", prerelease: "ee.h87ac00f"},
	} {
		Edition, PackageVer, PackageRev = tc.edition, tc.ver, tc.rev
		release, prerelease, found := strings.Cut(GetBinaryVersion(), "-")
		require.True(t, found, "a stamped release must carry a prerelease marker")
		assert.Equal(t, tc.release, release, "the release part must stay the bare number")
		assert.Equal(t, tc.prerelease, prerelease)
	}
}
func TestVersion_GetBinaryName_FallsBackToPentagi(t *testing.T) {
	t.Cleanup(func() { PackageName = "" })

	for name, want := range map[string]string{"": "pentagi", "myservice": "myservice"} {
		PackageName = name
		assert.Equal(t, want, GetBinaryName(), "PackageName=%q", name)
	}
}

// The hash is of the running test binary and must not change between calls.
func TestVersion_BinaryHash_IdentifiesThisExecutable(t *testing.T) {
	hash := BinaryHash()
	require.Len(t, hash, 64, "a sha256 in hex, or empty when unreadable — nothing else")
	assert.Regexp(t, "^[0-9a-f]{64}$", hash)

	path, err := os.Executable()
	require.NoError(t, err)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	sum := sha256.Sum256(content)
	assert.Equal(t, hex.EncodeToString(sum[:]), hash,
		"the hash must be of the running executable, not of anything else on the path")

	assert.Equal(t, hash, BinaryHash(), "the answer must not change under a running process")
}

func TestVersion_BuildRevision_IsNeverEmpty(t *testing.T) {
	assert.Equal(t, "abc1234", buildRevision("abc1234", "0123456789abcdef"))
	assert.Equal(t, "0123456789ab", buildRevision("", "0123456789abcdef"))
	assert.Equal(t, "unknown", buildRevision("", ""))
}
