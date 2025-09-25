package terminal

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/cmd/installer/wizard/terminal/vt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestNewTerminal(t *testing.T) {
	term := NewTerminal(80, 24)
	if term == nil {
		t.Fatal("NewTerminal returned nil")
	}

	width, height := term.GetSize()
	if width != 80 || height != 24 {
		t.Errorf("expected size 80x24, got %dx%d", width, height)
	}
}

func TestTerminalSetSize(t *testing.T) {
	term := NewTerminal(80, 24)
	term.SetSize(100, 30)

	width, height := term.GetSize()
	if width != 100 || height != 30 {
		t.Errorf("expected size 100x30, got %dx%d", width, height)
	}
}

func TestTerminalAppend(t *testing.T) {
	term := NewTerminal(80, 24)
	term.Append("test message")

	view := term.View()
	cleanView := ansi.Strip(view)
	if !strings.Contains(cleanView, "test message") {
		t.Error("appended message not found in view")
	}
}

func TestTerminalClear(t *testing.T) {
	term := NewTerminal(80, 24)
	term.Append("test message")
	term.Clear()

	view := term.View()
	cleanView := ansi.Strip(view)
	if strings.Contains(cleanView, "test message") {
		t.Error("message found after clear")
	}
}

func TestExecuteEcho(t *testing.T) {
	term := NewTerminal(80, 24)
	cmd := exec.Command("echo", "hello world")

	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// wait for command to complete and output to be processed
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for command completion")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "hello world") {
				return // success
			}
		}
	}
}

func TestExecuteCat(t *testing.T) {
	term := NewTerminal(80, 24)

	// create temp file with content
	tmpFile := t.TempDir() + "/test.txt"
	content := "line1\nline2\nline3\n"

	if err := writeFile(tmpFile, content); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := exec.Command("cat", tmpFile)
	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// wait for output
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for cat command")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "line1") && strings.Contains(cleanView, "line2") {
				return
			}
		}
	}
}

func TestExecuteGrep(t *testing.T) {
	term := NewTerminal(80, 24)

	tmpFile := t.TempDir() + "/test.txt"
	content := "apple\nbanana\ncherry\napricot\n"

	if err := writeFile(tmpFile, content); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cmd := exec.Command("grep", "ap", tmpFile)
	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for grep command")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "apple") && strings.Contains(cleanView, "apricot") {
				return
			}
		}
	}
}

func TestExecuteInteractiveInput(t *testing.T) {
	term := NewTerminal(80, 24)

	// use 'cat' without arguments to read from stdin
	cmd := exec.Command("cat")
	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// simulate user input via Update method in goroutine
	go func() {
		time.Sleep(100 * time.Millisecond)

		// send "hello" and enter
		for _, r := range "hello" {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
			term.Update(msg)
			time.Sleep(10 * time.Millisecond)
		}

		// send enter
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		term.Update(enterMsg)
		time.Sleep(50 * time.Millisecond)

		// send ctrl+d to close input
		ctrlDMsg := tea.KeyMsg{Type: tea.KeyCtrlD}
		term.Update(ctrlDMsg)
	}()

	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for interactive input")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "hello") {
				return
			}
		}
	}
}

func TestExecuteMultipleCommands(t *testing.T) {
	term := NewTerminal(80, 24)

	// test sequential execution
	commands := []struct {
		cmd    []string
		expect string
	}{
		{[]string{"echo", "first"}, "first"},
		{[]string{"echo", "second"}, "second"},
		{[]string{"echo", "third"}, "third"},
	}

	for i, cmdTest := range commands {
		if i > 0 {
			// wait for previous command to finish
			time.Sleep(200 * time.Millisecond)
		}

		cmd := exec.Command(cmdTest.cmd[0], cmdTest.cmd[1:]...)
		err := term.Execute(cmd)
		if err != nil {
			t.Fatalf("Execute command %d failed: %v", i, err)
		}

		// wait for output
		timeout := time.NewTimer(2 * time.Second)
		ticker := time.NewTicker(50 * time.Millisecond)

		found := false
		for !found {
			select {
			case <-timeout.C:
				t.Fatalf("timeout waiting for command %d output", i)
			case <-ticker.C:
				view := term.View()
				cleanView := ansi.Strip(view)
				if strings.Contains(cleanView, cmdTest.expect) {
					found = true
				}
			}
		}
		timeout.Stop()
		ticker.Stop()

		if !found {
			t.Errorf("command %d output not found in view", i)
		}
	}
}

func TestRestoreModel(t *testing.T) {
	term := NewTerminal(80, 24)

	// test with valid terminal
	restored := RestoreModel(term)
	if restored == nil {
		t.Error("RestoreModel returned nil for valid terminal")
	}

	// test with invalid model
	invalidModel := &struct{ tea.Model }{}
	restored = RestoreModel(invalidModel)
	if restored != nil {
		t.Error("RestoreModel should return nil for invalid model")
	}
}

func TestTeaKeyToUVKey(t *testing.T) {
	tests := []struct {
		name     string
		key      tea.KeyMsg
		expected vt.KeyPressEvent
	}{
		{
			name:     "Arrow Up",
			key:      tea.KeyMsg{Type: tea.KeyUp},
			expected: vt.KeyPressEvent{Code: vt.KeyUp, Mod: 0},
		},
		{
			name:     "Arrow Down",
			key:      tea.KeyMsg{Type: tea.KeyDown},
			expected: vt.KeyPressEvent{Code: vt.KeyDown, Mod: 0},
		},
		{
			name:     "Arrow Left",
			key:      tea.KeyMsg{Type: tea.KeyLeft},
			expected: vt.KeyPressEvent{Code: vt.KeyLeft, Mod: 0},
		},
		{
			name:     "Arrow Right",
			key:      tea.KeyMsg{Type: tea.KeyRight},
			expected: vt.KeyPressEvent{Code: vt.KeyRight, Mod: 0},
		},
		{
			name:     "Enter",
			key:      tea.KeyMsg{Type: tea.KeyEnter},
			expected: vt.KeyPressEvent{Code: vt.KeyEnter, Mod: 0},
		},
		{
			name:     "Tab",
			key:      tea.KeyMsg{Type: tea.KeyTab},
			expected: vt.KeyPressEvent{Code: vt.KeyTab, Mod: 0},
		},
		{
			name:     "Space",
			key:      tea.KeyMsg{Type: tea.KeySpace},
			expected: vt.KeyPressEvent{Code: vt.KeySpace, Mod: 0},
		},
		{
			name:     "Regular character 'a'",
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			expected: vt.KeyPressEvent{Code: 'a', Mod: 0},
		},
		{
			name:     "Ctrl+C",
			key:      tea.KeyMsg{Type: tea.KeyCtrlC},
			expected: vt.KeyPressEvent{Code: 'c', Mod: vt.ModCtrl},
		},
		{
			name:     "Alt+Up",
			key:      tea.KeyMsg{Type: tea.KeyUp, Alt: true},
			expected: vt.KeyPressEvent{Code: vt.KeyUp, Mod: vt.ModAlt},
		},
		{
			name:     "Shift+Tab",
			key:      tea.KeyMsg{Type: tea.KeyShiftTab},
			expected: vt.KeyPressEvent{Code: vt.KeyTab, Mod: vt.ModShift},
		},
		{
			name:     "F1",
			key:      tea.KeyMsg{Type: tea.KeyF1},
			expected: vt.KeyPressEvent{Code: vt.KeyF1, Mod: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := teaKeyToUVKey(test.key)
			if result == nil {
				t.Errorf("teaKeyToUVKey(%+v) returned nil", test.key)
				return
			}

			keyPress, ok := result.(vt.KeyPressEvent)
			if !ok {
				t.Errorf("teaKeyToUVKey(%+v) returned non-KeyPressEvent: %T", test.key, result)
				return
			}

			if keyPress.Code != test.expected.Code || keyPress.Mod != test.expected.Mod {
				t.Errorf("teaKeyToUVKey(%+v) = {Code: %v, Mod: %v}, expected {Code: %v, Mod: %v}",
					test.key, keyPress.Code, keyPress.Mod, test.expected.Code, test.expected.Mod)
			}
		})
	}
}

func TestExecuteConcurrency(t *testing.T) {
	term := NewTerminal(80, 24)

	// try to execute two commands simultaneously
	cmd1 := exec.Command("echo", "first")
	err1 := term.Execute(cmd1)
	if err1 != nil {
		t.Fatalf("first Execute failed: %v", err1)
	}

	// second command should fail because terminal is busy
	cmd2 := exec.Command("echo", "second")
	err2 := term.Execute(cmd2)
	if err2 == nil {
		t.Error("second Execute should have failed while first is running")
	}
	if !strings.Contains(err2.Error(), "already executing") {
		t.Errorf("unexpected error message: %v", err2)
	}
}

// verifies that waiting on the external cmd and then Terminal.Wait() makes
// subsequent Execute calls safe (no race) in non-PTY mode
func TestWaitBeforeNextExecute_NoPty(t *testing.T) {
	term := NewTerminal(80, 24, WithNoPty())

	var cmd1 *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd1 = exec.Command("cmd", "/c", "echo one")
	} else {
		cmd1 = exec.Command("sh", "-c", "echo one")
	}

	if err := term.Execute(cmd1); err != nil {
		t.Fatalf("first Execute failed: %v", err)
	}

	// client waits for process completion first
	if err := cmd1.Wait(); err != nil {
		t.Fatalf("cmd1.Wait failed: %v", err)
	}

	// ensure terminal finished internal cleanup
	term.Wait()

	var cmd2 *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd2 = exec.Command("cmd", "/c", "echo two")
	} else {
		cmd2 = exec.Command("sh", "-c", "echo two")
	}

	if err := term.Execute(cmd2); err != nil {
		t.Fatalf("second Execute failed after Wait(): %v", err)
	}

	if err := cmd2.Wait(); err != nil {
		t.Fatalf("cmd2.Wait failed: %v", err)
	}
	term.Wait()

	// verify content contains outputs from both commands
	cleanView := ansi.Strip(term.View())
	if !(strings.Contains(cleanView, "one") && strings.Contains(cleanView, "two")) {
		t.Fatalf("expected outputs not found in view: %q", cleanView)
	}
}

// verifies that waiting on cmd and then Terminal.Wait() is safe in PTY mode
func TestWaitBeforeNextExecute_Pty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping PTY test on Windows")
	}

	term := NewTerminal(80, 24)

	cmd1 := exec.Command("sh", "-c", "echo one")
	if err := term.Execute(cmd1); err != nil {
		t.Fatalf("first Execute failed: %v", err)
	}

	if err := cmd1.Wait(); err != nil {
		t.Fatalf("cmd1.Wait failed: %v", err)
	}
	term.Wait()

	cmd2 := exec.Command("sh", "-c", "echo two")
	if err := term.Execute(cmd2); err != nil {
		t.Fatalf("second Execute failed after Wait(): %v", err)
	}

	if err := cmd2.Wait(); err != nil {
		t.Fatalf("cmd2.Wait failed: %v", err)
	}
	term.Wait()

	cleanView := ansi.Strip(term.View())
	if !(strings.Contains(cleanView, "one") && strings.Contains(cleanView, "two")) {
		t.Fatalf("expected outputs not found in view: %q", cleanView)
	}
}

// helper function to write file content
func writeFile(filename, content string) error {
	cmd := exec.Command("sh", "-c", "cat > "+filename)
	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

// benchmark basic terminal operations
func BenchmarkTerminalAppend(b *testing.B) {
	term := NewTerminal(80, 24)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		term.Append("benchmark message")
	}
}

func BenchmarkTerminalView(b *testing.B) {
	term := NewTerminal(80, 24)
	term.Append("some content to render")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = term.View()
	}
}

func BenchmarkKeySequenceConversion(b *testing.B) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = teaKeyToUVKey(msg)
	}
}

func TestTerminalEvents(t *testing.T) {
	term := NewTerminal(80, 24, WithAutoPoll())

	// first init acquires subscription
	cmd1 := term.Init()
	if cmd1 == nil {
		t.Fatal("Init() should return a command for first subscription")
	}

	go cmd1()
	time.Sleep(100 * time.Millisecond) // wait for subscription to be acquired

	// second init should return nil (already subscribed)
	if cmd2 := term.Init(); cmd2 == nil || cmd2() != nil {
		t.Fatal("Init() should return cmd with nil message when there is already an active subscriber")
	}

	// append should trigger event
	term.Append("test message")

	// after update message, Update must return a new wait command
	model, nextCmd := term.Update(TerminalUpdateMsg{ID: term.ID()})
	if model == nil {
		t.Error("Update should return model")
	}
	if nextCmd == nil {
		t.Error("Update should return next command for continued listening")
	}

	// simulate receiving the update message from another terminal
	model, nextCmd = term.Update(TerminalUpdateMsg{ID: "other"})
	if model == nil {
		t.Error("Update should return model")
	}
	if nextCmd != nil {
		t.Error("Update should not return next command for other terminal")
	}
}

func TestTerminalFinalizer(t *testing.T) {
	term := NewTerminal(80, 24)

	// terminal should start normally
	termImpl := term.(*terminal)
	if termImpl.notifier == nil {
		t.Error("new terminal should have notifier")
	}

	// execute a command to create some resources
	cmd := exec.Command("echo", "test")
	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// wait for command completion
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for command completion")
		case <-ticker.C:
			if !term.IsRunning() {
				// manually call finalizer to test resource cleanup
				terminalFinalizer(termImpl)

				// verify notifier is cleaned up
				if termImpl.notifier != nil {
					t.Error("notifier should be nil after finalizer")
				}

				// verify terminal resources are cleaned
				termImpl.mx.Lock()
				if termImpl.cmd != nil || termImpl.pty != nil {
					termImpl.mx.Unlock()
					t.Error("terminal resources should be cleaned after finalizer")
					return
				}
				termImpl.mx.Unlock()
				return
			}
		}
	}
}

func TestResourceCleanup(t *testing.T) {
	term := NewTerminal(80, 24)

	// execute a command
	cmd := exec.Command("echo", "test")
	err := term.Execute(cmd)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// wait for completion
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for command completion")
		case <-ticker.C:
			if !term.IsRunning() {
				// command completed, resources should be cleaned
				termImpl := term.(*terminal)
				termImpl.mx.Lock()
				if termImpl.cmd != nil || termImpl.pty != nil {
					termImpl.mx.Unlock()
					t.Error("resources not cleaned after command completion")
					return
				}
				termImpl.mx.Unlock()
				return
			}
		}
	}
}

func TestResourceRelease(t *testing.T) {
	var wg sync.WaitGroup
	term := NewTerminal(80, 24)

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond)
		term.Execute(exec.Command("echo", "test"))
	}()

	cmd := term.Init()
	if cmd == nil {
		t.Fatal("Init() should return a command")
	}

	// wait for command output
	cmd()

	view := term.View()
	cleanView := ansi.Strip(view)
	if !strings.Contains(cleanView, "test") {
		t.Fatal("command output not found in view")
	}

	wg.Wait()
	term = nil
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cmd()
		cancel()
	}()

	// wait for resources to be released
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for command completion")
		case <-ctx.Done():
			t.Log("context done")
			return
		case <-ticker.C:
			runtime.GC()
		}
	}
}

// Tests for startCmd functionality specifically
func TestStartCmdBasic(t *testing.T) {
	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Force use of startCmd instead of startPty
	cmd := exec.Command("echo", "Hello from startCmd!")
	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for command to complete
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for startCmd completion")
		case <-ticker.C:
			if !term.IsRunning() {
				view := term.View()
				cleanView := ansi.Strip(view)
				if !strings.Contains(cleanView, "Hello from startCmd!") {
					t.Errorf("Expected output not found. Got: %q", cleanView)
				}
				return
			}
		}
	}
}

func TestStartCmdInteractive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping interactive test on Windows")
	}

	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Use cat for interactive testing
	cmd := exec.Command("cat")
	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for command to start
	time.Sleep(100 * time.Millisecond)

	// Verify command is running
	if !term.IsRunning() {
		t.Fatal("Command should be running")
	}

	// Send input through stdinPipe
	testInput := "Hello from stdin!\n"
	term.mx.Lock()
	if term.stdinPipe != nil {
		_, err := term.stdinPipe.Write([]byte(testInput))
		if err != nil {
			term.mx.Unlock()
			t.Fatalf("Failed to write to stdin: %v", err)
		}
	} else {
		term.mx.Unlock()
		t.Fatal("stdinPipe should not be nil")
	}
	term.mx.Unlock()

	// Wait for output to appear
	outputTimeout := time.NewTimer(2 * time.Second)
	defer outputTimeout.Stop()

	outputTicker := time.NewTicker(50 * time.Millisecond)
	defer outputTicker.Stop()

	outputFound := false
	for !outputFound {
		select {
		case <-outputTimeout.C:
			t.Fatal("timeout waiting for interactive output")
		case <-outputTicker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "Hello from stdin!") {
				outputFound = true
			}
		}
	}

	// Send EOF to terminate
	term.mx.Lock()
	if term.stdinPipe != nil {
		term.stdinPipe.Write([]byte{4}) // Ctrl+D
	}
	term.mx.Unlock()

	// Wait for completion
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			// Command might still be running, that's ok if we got output
			if outputFound {
				t.Log("Command may still be running, but output was received successfully")
				return
			}
			t.Fatal("timeout waiting for command completion")
		case <-ticker.C:
			if !term.IsRunning() {
				return // success
			}
		}
	}
}

func TestStartCmdStderrHandling(t *testing.T) {
	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Command that writes to stderr
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo Error output 1>&2")
	} else {
		cmd = exec.Command("sh", "-c", "echo 'Error output' >&2")
	}

	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for completion
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for stderr command completion")
		case <-ticker.C:
			if !term.IsRunning() {
				view := term.View()
				cleanView := ansi.Strip(view)
				if !strings.Contains(cleanView, "Error output") {
					t.Errorf("Stderr output not found. Got: %q", cleanView)
				}
				return
			}
		}
	}
}

func TestStartCmdPlainTextOutput(t *testing.T) {
	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Command that outputs plain text (no ANSI processing in cmd mode)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("echo", "Simple text output")
	} else {
		cmd = exec.Command("echo", "Simple text output")
	}

	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for completion
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for command completion")
		case <-ticker.C:
			if !term.IsRunning() {
				// Get final view content
				view := term.View()
				cleanView := ansi.Strip(view)

				if !strings.Contains(cleanView, "Simple text output") {
					t.Errorf("Expected 'Simple text output' not found. Got: %q", cleanView)
				}

				// Verify that cmdLines buffer was used (not vt)
				term.mx.Lock()
				vtExists := term.vt != nil
				cmdLinesExist := term.cmdLines != nil
				term.mx.Unlock()

				if vtExists {
					t.Error("vt should not be created in startCmd mode")
				}

				if !cmdLinesExist {
					t.Log("cmdLines was already cleaned up by manageCmd, which is expected behavior")
				}

				return
			}
		}
	}
}

func TestStartCmdSimpleKeyHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping key handling test on Windows")
	}

	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Start cat command for key input testing
	cmd := exec.Command("cat")
	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for command to start
	time.Sleep(100 * time.Millisecond)

	// Test simple key input
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("hello")},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune("world")},
		{Type: tea.KeyEnter},
		{Type: tea.KeyCtrlD}, // EOF
	}

	for _, key := range testKeys {
		term.mx.Lock()
		handled := term.handleTerminalInput(key)
		term.mx.Unlock()

		if !handled {
			t.Errorf("handleTerminalInput should handle key: %+v", key)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Wait for output
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for key input processing")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "hello world") {
				return // success
			}
		}
	}
}

func TestStartCmdInputHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping input test on Windows")
	}

	term := NewTerminal(80, 24).(*terminal)
	defer func() {
		term.mx.Lock()
		term.cleanup()
		term.mx.Unlock()
	}()

	// Start cat command
	cmd := exec.Command("cat")
	err := term.startCmd(cmd)
	if err != nil {
		t.Fatalf("startCmd failed: %v", err)
	}

	// Wait for command to start
	time.Sleep(100 * time.Millisecond)

	// Test key input handling through handleTerminalInput
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("test")},
		{Type: tea.KeySpace},
		{Type: tea.KeyRunes, Runes: []rune("input")},
		{Type: tea.KeyEnter},
		{Type: tea.KeyCtrlD}, // EOF
	}

	for _, key := range testKeys {
		term.mx.Lock()
		handled := term.handleTerminalInput(key)
		term.mx.Unlock()

		if !handled {
			t.Errorf("handleTerminalInput should handle key: %+v", key)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Wait for output and completion
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatal("timeout waiting for input handling completion")
		case <-ticker.C:
			view := term.View()
			cleanView := ansi.Strip(view)
			if strings.Contains(cleanView, "test input") {
				return // success
			}
		}
	}
}
