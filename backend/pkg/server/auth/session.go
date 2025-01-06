package auth

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"sync"

	"pentagi/pkg/system"
	"pentagi/pkg/version"
)

var (
	cookieStoreKeys [][]byte
	cookieStoreOnce sync.Once
)

// MakeCookieStoreKey is function to generate auth and encryption keys for cookie store
func MakeCookieStoreKey(globalSalt string) [][]byte {
	cookieStoreOnce.Do(func() {
		baseHash := func(values ...string) string {
			hash := sha256.Sum256([]byte(strings.Join(values, "|")))
			return hex.EncodeToString(hash[:])
		}
		authKey := strings.Join([]string{
			baseHash(version.GetBinaryVersion(), "a8d0abae36f749588f4393e6fc292690", globalSalt),
			system.GetHostID(),
			globalSalt,
		}, "|")
		encKey := strings.Join([]string{
			baseHash(version.GetBinaryVersion(), "7c9be62adec5076970fa946e78f256e2", globalSalt),
			system.GetHostID(),
			globalSalt,
		}, "|")
		authKeyBytes := sha512.Sum512([]byte(authKey))
		encKeyBytes := sha256.Sum256([]byte(encKey))
		cookieStoreKeys = [][]byte{authKeyBytes[:], encKeyBytes[:]}
	})
	return cookieStoreKeys
}
