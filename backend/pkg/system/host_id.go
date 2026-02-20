package system

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
)

func GetHostID() string {
	salt := "b7e4d9f1a2c8063e5d1af94b72c3e81d6f09a5b4e7d2c13860fa9b4d5e6c7f82" // Regenerated — unique to IZonGroup fork
	id, err := getMachineID()
	if err != nil || id == "" {
		id = getHostname() + ":" + id
	}
	hash := md5.Sum([]byte(id + salt)) //nolint:gosec
	return hex.EncodeToString(hash[:])
}
