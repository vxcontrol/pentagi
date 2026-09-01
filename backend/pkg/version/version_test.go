package version

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBinaryVersion_Default(t *testing.T) {
	PackageVer = ""
	PackageRev = ""

	result := GetBinaryVersion()
	assert.Equal(t, "develop", result)
}

func TestGetBinaryVersion_WithVersion(t *testing.T) {
	PackageVer = "1.2.0"
	PackageRev = ""
	defer func() { PackageVer = "" }()

	result := GetBinaryVersion()
	assert.Equal(t, "1.2.0", result)
}

func TestGetBinaryVersion_WithVersionAndRevision(t *testing.T) {
	PackageVer = "1.2.0"
	PackageRev = "abc1234"
	defer func() {
		PackageVer = ""
		PackageRev = ""
	}()

	result := GetBinaryVersion()
	assert.Equal(t, "1.2.0-abc1234", result)
}

func TestGetBinaryVersion_WithRevisionOnly(t *testing.T) {
	PackageVer = ""
	PackageRev = "abc1234"
	defer func() { PackageRev = "" }()

	result := GetBinaryVersion()
	assert.Equal(t, "develop-abc1234", result)
}

func TestIsDevelopMode_True(t *testing.T) {
	PackageVer = ""

	assert.True(t, IsDevelopMode())
}

func TestIsDevelopMode_False(t *testing.T) {
	PackageVer = "1.0.0"
	defer func() { PackageVer = "" }()

	assert.False(t, IsDevelopMode())
}

func TestGetBinaryName_Default(t *testing.T) {
	PackageName = ""

	result := GetBinaryName()
	assert.Equal(t, "pentagi", result)
}

func TestGetBinaryName_Custom(t *testing.T) {
	PackageName = "myservice"
	defer func() { PackageName = "" }()

	result := GetBinaryName()
	assert.Equal(t, "myservice", result)
}

// TestBinaryHashIdentifiesThisExecutable.
//
// It has to be the digest of the file this process was started from — the test
// binary here — and it has to be the SAME on every call: the value is cached
// precisely because reading the executable again could only produce a different
// answer if somebody replaced a file this process is no longer executing.
func TestBinaryHashIdentifiesThisExecutable(t *testing.T) {
	hash := BinaryHash()
	require.Len(t, hash, 64, "a sha256 in hex, or empty when unreadable — nothing else")
	assert.Regexp(t, "^[0-9a-f]{64}$", hash)

	path, err := os.Executable()
	require.NoError(t, err)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(sha256Sum(content)), hash,
		"the hash must be of the running executable, not of anything else on the path")

	assert.Equal(t, hash, BinaryHash(), "the answer must not change under a running process")
}

func sha256Sum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}
