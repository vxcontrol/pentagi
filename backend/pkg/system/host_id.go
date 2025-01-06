package system

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
)

func GetHostID() string {
	salt := "acfee3b28d2d95904730177369171ac430c08bab050350f173d92b14563eccee"
	id, err := getMachineID()
	if err != nil || id == "" {
		id = getHostname() + ":" + id
	}
	hash := md5.Sum([]byte(id + salt)) //nolint:gosec
	return hex.EncodeToString(hash[:])
}
