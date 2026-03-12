package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
