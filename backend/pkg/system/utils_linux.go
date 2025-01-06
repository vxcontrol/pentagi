//go:build linux
// +build linux

package system

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/digitalocean/go-smbios/smbios"
)

type Feature int

const (
	_ Feature = iota

	// System manufacturer. Requires access to SMBIOS data via DMI (i.e. root privileges)
	SystemManufacturer
	// System product name. Requires access to SMBIOS data via DMI (i.e. root privileges)
	SystemProductName
	// System UUID. Makes sense for virtual machines. Requires access to SMBIOS data via DMI (i.e. root privileges)
	SystemUUID
)

var (
	readSMBIOSOnce   sync.Once
	smbiosReadingErr error
	smbiosAttrValues = make(map[Feature]string)

	// See SMBIOS specification https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.3.0.pdf
	smbiosAttrTable = [...]smbiosAttribute{
		// System 0x01
		{0x01, 0x08, 0x04, SystemManufacturer, nil},
		{0x01, 0x08, 0x05, SystemProductName, nil},
		{0x01, 0x14, 0x08, SystemUUID, formatUUID},
	}
)

func formatUUID(b []byte) (string, error) {
	return fmt.Sprintf("%0x-%0x-%0x-%0x-%0x", b[:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func getSMBIOSAttr(feat Feature) (string, error) {
	readSMBIOSOnce.Do(func() {
		smbiosReadingErr = readSMBIOSAttributes()
	})
	if smbiosReadingErr != nil {
		return "", smbiosReadingErr
	}
	return smbiosAttrValues[feat], nil
}

func readSMBIOSAttributes() error {
	smbiosAttrIndex := buildSMBIOSAttrIndex()
	rc, _, err := smbios.Stream()
	if err != nil {
		return fmt.Errorf("unable to open SMBIOS info: %v", err)
	}
	defer rc.Close()
	structures, err := smbios.NewDecoder(rc).Decode()
	if err != nil {
		return fmt.Errorf("unable to decode SMBIOS info: %v", err)
	}
	for _, s := range structures {
		attrList := smbiosAttrIndex[int(s.Header.Type)]
		for _, attr := range attrList {
			val, err := attr.readValueString(s)
			if err != nil {
				return fmt.Errorf("unable to read SMBIOS attribute '%v' of structure type 0x%0x: %v",
					attr.feature, s.Header.Type, err)
			}
			smbiosAttrValues[attr.feature] = val
		}
	}
	return nil
}

func buildSMBIOSAttrIndex() map[int][]*smbiosAttribute {
	res := make(map[int][]*smbiosAttribute)
	for i := range smbiosAttrTable {
		attr := &smbiosAttrTable[i]
		res[attr.structType] = append(res[attr.structType], attr)
	}
	return res
}

type smbiosAttribute struct {
	structType      int
	structMinLength int
	offset          int

	feature Feature

	format func(data []byte) (string, error)
}

func (attr *smbiosAttribute) readValueString(s *smbios.Structure) (string, error) {
	if len(s.Formatted) < attr.structMinLength {
		return "", nil
	}
	if attr.format != nil {
		const headerSize = 4
		return attr.format(s.Formatted[attr.offset-headerSize:])
	}
	return attr.getString(s)
}

func (attr *smbiosAttribute) getString(s *smbios.Structure) (string, error) {
	const headerSize = 4
	strNo := int(s.Formatted[attr.offset-headerSize])
	strNo -= 1
	if strNo < 0 || strNo >= len(s.Strings) {
		return "", fmt.Errorf("invalid string no")
	}
	return s.Strings[strNo], nil
}

func getMachineID() (string, error) {
	const (
		// dbusPath is the default path for dbus machine id.
		dbusPath = "/var/lib/dbus/machine-id"
		// dbusPathEtc is the default path for dbus machine id located in /etc.
		// Some systems (like Fedora 20) only know this path.
		// Sometimes it's the other way round.
		dbusPathEtc = "/etc/machine-id"
	)

	id, err := os.ReadFile(dbusPath)
	if err != nil {
		id, err = os.ReadFile(dbusPathEtc)
	}
	if err != nil {
		return "", err
	}
	machineID := strings.TrimSpace(strings.Trim(string(id), "\n"))

	// root privileges are required to access attributes, the process will be skipped in case of insufficient privileges
	smbiosAttrs := [...]Feature{SystemUUID, SystemManufacturer, SystemProductName}
	for _, attr := range smbiosAttrs {
		attrVal, err := getSMBIOSAttr(attr)
		if err != nil || strings.TrimSpace(attrVal) == "" {
			continue
		}
		machineID = fmt.Sprintf("%s:%s", machineID, strings.ToLower(attrVal))
	}

	return machineID, nil
}
