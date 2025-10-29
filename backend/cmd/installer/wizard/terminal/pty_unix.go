//go:build !windows

package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"pentagi/cmd/installer/wizard/terminal/vt"

	"github.com/creack/pty"
)

// startPty sets term, tty, pty, and cmd properties and starts the command
func (t *terminal) startPty(cmd *exec.Cmd) error {
	var err error
	t.pty, t.tty, err = pty.Open()
	if err != nil {
		return err
	}

	// set up environment
	if cmd.Env == nil {
		cmd.Env = t.env
	}

	// ensure TERM is set correctly
	termSet := false
	termEnv := fmt.Sprintf("TERM=%s", terminalModel)
	for i, env := range cmd.Env {
		if len(env) >= 5 && env[:5] == "TERM=" {
			cmd.Env[i] = termEnv
			termSet = true
			break
		}
	}
	if !termSet {
		cmd.Env = append(cmd.Env, termEnv)
	}

	ws := t.getWinSize()
	t.vt = vt.NewTerminal(int(ws.Cols), int(ws.Rows), t.pty)
	t.vt.SetLogger(&dummyLogger{})

	tearDownPty := func(err error) error {
		return errors.Join(err, t.tty.Close(), t.pty.Close())
	}

	// according to the creack/pty library implementation (just copy to keep tty open)
	t.cmd = cmd
	if t.cmd.Stdout == nil {
		t.cmd.Stdout = t.tty
	}
	if t.cmd.Stderr == nil {
		t.cmd.Stderr = t.tty
	}
	if t.cmd.Stdin == nil {
		t.cmd.Stdin = t.tty
	}
	if t.cmd.SysProcAttr == nil {
		t.cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	t.cmd.SysProcAttr.Setsid = true
	t.cmd.SysProcAttr.Setctty = true

	if err := pty.Setsize(t.pty, ws); err != nil {
		return tearDownPty(err)
	}

	if err := t.cmd.Start(); err != nil {
		return tearDownPty(err)
	}

	// close parent's copy of slave side to ensure EOF is delivered when child exits
	// child keeps its own descriptors; we must not hold t.tty open in parent
	if t.tty != nil {
		_ = t.tty.Close()
		t.tty = nil
	}

	go t.managePty()

	return nil
}

// managePty manages the pseudoterminal and its output
func (t *terminal) managePty() {
	t.wg.Add(1)
	defer t.wg.Done()

	defer func() {
		t.mx.Lock()
		defer t.mx.Unlock()

		t.cleanup()
		t.updateViewpoint()
	}()

	// get reader while holding lock briefly
	t.mx.Lock()
	if t.pty == nil {
		t.mx.Unlock()
		return
	}
	// large buffer for better ANSI sequence capture
	reader := bufio.NewReaderSize(t.pty, 32768)
	buf := make([]byte, 32768)
	t.mx.Unlock()

	handleError := func(msg string, err error) {
		t.contents = append(t.contents, fmt.Sprintf("%s: %v", msg, err))
	}

	for {
		n, err := reader.Read(buf)

		// update output buffer
		if n > 0 {
			t.mx.Lock()
			if _, err := t.vt.Write(buf[:n]); err != nil {
				handleError("error writing to terminal", err)
			}
			t.updateViewpoint()
			t.mx.Unlock()
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				// normal termination
				break
			}
			// on linux, reading from ptmx after slave closes returns EIO; treat as EOF
			if errors.Is(err, syscall.EIO) {
				break
			}

			// handle other errors
			t.mx.Lock()
			handleError("error reading output", err)

			// try to kill process if it's still running
			if t.cmd != nil && t.cmd.Process != nil && (t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited()) {
				if killErr := t.cmd.Process.Kill(); killErr != nil {
					handleError("failed to terminate process", killErr)
				} else {
					t.contents = append(t.contents, "process terminated")
				}
			}

			t.mx.Unlock()
			break
		}
	}
}
