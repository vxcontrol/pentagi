//go:build windows

package terminal

import (
	"fmt"
	"os/exec"
)

// startPty is not supported on Windows
func (t *terminal) startPty(cmd *exec.Cmd) error {
	return fmt.Errorf("pty mode is not supported on Windows")
}

// managePty is not supported on Windows
func (t *terminal) managePty() {
	// no-op on Windows
}
