package version

import (
	"fmt"
)

// PackageName is service name or binary name
var PackageName string

// PackageVer is semantic version of the binary
var PackageVer string

// PackageRev is revision of the binary build
var PackageRev string

func GetBinaryVersion() string {
	version := "develop"
	if PackageVer != "" {
		version = PackageVer
	}
	if PackageRev != "" {
		version = fmt.Sprintf("%s-%s", version, PackageRev)
	}
	return version
}

func IsDevelopMode() bool {
	return PackageVer == ""
}

func GetBinaryName() string {
	if PackageName != "" {
		return PackageName
	}
	return "pentagi"
}
