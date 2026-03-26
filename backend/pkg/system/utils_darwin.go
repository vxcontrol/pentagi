//go:build darwin
// +build darwin

package system

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var execLock sync.Mutex

func getMachineID() (string, error) {
	out, err := execCmd("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		return "", err
	}
	id, err := extractID(out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Trim(id, "\n")), nil
}

func extractID(lines string) (string, error) {
	const uuidParamName = "IOPlatformUUID"
	for _, line := range strings.Split(lines, "\n") {
		if strings.Contains(line, uuidParamName) {
			parts := strings.SplitAfter(line, `" = "`)
			if len(parts) == 2 {
				return strings.TrimRight(parts[1], `"`), nil
			}
		}
	}
	return "", fmt.Errorf("failed to extract the '%s' value from the `ioreg` output", uuidParamName)
}

func execCmd(scmd string, args ...string) (string, error) {
	execLock.Lock()
	defer execLock.Unlock()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(scmd, args...)
	cmd.Stdin = strings.NewReader("")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}
