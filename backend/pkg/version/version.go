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

// BinaryHash is the sha256 of the executable this process was started from.
//
// It answers what the version string cannot: two builds cut from the same branch
// carry the same "develop" or the same "0.9.3-87ac00f", and outside a release that
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
