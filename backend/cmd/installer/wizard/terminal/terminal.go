package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"pentagi/cmd/installer/wizard/logger"
	"pentagi/cmd/installer/wizard/terminal/vt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/creack/pty"
	"github.com/google/uuid"
)

const (
	terminalBorderColor = "62"
	terminalPadding     = 1
	terminalModel       = "xterm-256color"
	terminalMinWidth    = 20
	terminalMinHeight   = 10
)

type dummyLogger struct{}

func (l *dummyLogger) Printf(format string, v ...any) {
	logger.Log(format, v...)
}

// TerminalUpdateMsg represents a terminal content update event
type TerminalUpdateMsg struct {
	ID string
}

type TerminalOption func(*terminal)

// WithAutoScroll sets scroll to the bottom when new content is appended
func WithAutoScroll() TerminalOption {
	return func(t *terminal) {
		t.autoScroll = true
	}
}

// WithAutoPoll enables automatic polling for new terminal content.
// when set, the terminal will actively check for updates without relying on the BubbleTea update loop ticker.
// this is particularly useful for scenarios with frequent updates to ensure real-time content display.
func WithAutoPoll() TerminalOption {
	return func(t *terminal) {
		t.autoPoll = true
	}
}

// WithStyle sets the style for the terminal viewport
func WithStyle(style lipgloss.Style) TerminalOption {
	return func(t *terminal) {
		t.viewport.Style = style
	}
}

// WithCurrentEnv sets the environment variables for the terminal to the current process's environment
// it is working for cmds (exec.Cmd) without env set, use non-nil Env property to prevent overriding
func WithCurrentEnv() TerminalOption {
	return func(t *terminal) {
		t.env = os.Environ()
	}
}

// WithNoStyled disables styled output (used for pty mode only)
func WithNoStyled() TerminalOption {
	return func(t *terminal) {
		t.noStyled = true
	}
}

// WithNoPty disables pty mode (used for cmds lines output)
func WithNoPty() TerminalOption {
	return func(t *terminal) {
		t.noPty = true
	}
}

type Terminal interface {
	Execute(cmd *exec.Cmd) error
	Append(content string)
	Clear()
	IsRunning() bool
	Wait()

	SetSize(width, height int)
	GetSize() (width, height int)
	ID() string

	tea.Model
}

type terminal struct {
	viewport viewport.Model
	contents []string

	// terminal state
	pty *os.File
	tty *os.File
	cmd *exec.Cmd

	// for non-pty commands
	stdinPipe io.WriteCloser
	cmdLines  []string

	// output buffer
	vt *vt.Terminal
	mx *sync.Mutex
	wg *sync.WaitGroup
	id string

	// notifier for single-subscriber update notifications
	notifier *updateNotifier

	// terminal settings
	autoScroll bool
	autoPoll   bool
	noStyled   bool
	noPty      bool
	env        []string
}

// terminalFinalizer properly cleans up terminal resources
func terminalFinalizer(t *terminal) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.cleanup()

	if t.notifier != nil {
		t.notifier.close()
		t.notifier = nil
	}
}

func NewTerminal(width, height int, opts ...TerminalOption) Terminal {
	id := uuid.New().String()
	t := &terminal{
		viewport: viewport.New(width, height),
		contents: []string{},
		mx:       &sync.Mutex{},
		wg:       &sync.WaitGroup{},
		id:       id,
		notifier: newUpdateNotifier(),
	}

	// set default style
	t.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(terminalBorderColor)).
		Padding(terminalPadding)

	for _, opt := range opts {
		opt(t)
	}

	// set finalizer for automatic resource cleanup
	runtime.SetFinalizer(t, terminalFinalizer)

	return t
}

func (t *terminal) Execute(cmd *exec.Cmd) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	if t.cmd != nil || t.pty != nil || t.tty != nil || t.vt != nil {
		return fmt.Errorf("terminal is already executing a command")
	}

	wrapError := func(err error) error {
		if err != nil {
			t.cleanup()

			msg := fmt.Sprintf("failed to execute command: %v", err)
			t.contents = append(t.contents, msg)
			t.updateViewpoint()
		}

		return err
	}

	if runtime.GOOS == "windows" || t.noPty {
		return wrapError(t.startCmd(cmd))
	} else {
		return wrapError(t.startPty(cmd))
	}
}

func (t *terminal) startCmd(cmd *exec.Cmd) error {
	// set up environment
	if cmd.Env == nil {
		cmd.Env = t.env
	}

	// initialize command lines buffer
	t.cmdLines = []string{}

	// set up pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// set up stdin pipe for interactive commands
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		stdoutPipe.Close()
		stderrPipe.Close()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// store pipes for cleanup and input handling
	t.cmd = cmd

	// start the command
	if err := cmd.Start(); err != nil {
		stdoutPipe.Close()
		stderrPipe.Close()
		stdinPipe.Close()
		return fmt.Errorf("failed to start command: %w", err)
	}

	// start managing the command output
	go t.manageCmd(stdoutPipe, stderrPipe, stdinPipe)

	return nil
}

// manageCmd manages command pipes and their output
func (t *terminal) manageCmd(stdoutPipe, stderrPipe io.ReadCloser, stdinPipe io.WriteCloser) {
	t.wg.Add(1)
	defer t.wg.Done()

	defer func() {
		t.mx.Lock()
		defer t.mx.Unlock()

		// close pipes
		stdoutPipe.Close()
		stderrPipe.Close()
		stdinPipe.Close()

		t.cleanup()
		t.updateViewpoint()
	}()

	// store stdin pipe for input handling
	t.mx.Lock()
	t.stdinPipe = stdinPipe
	t.mx.Unlock()

	handleError := func(msg string, err error) {
		t.contents = append(t.contents, fmt.Sprintf("%s: %v", msg, err))
	}

	// create channels for coordinating output from both streams
	lineChan := make(chan string, 10)
	errorChan := make(chan error, 2)
	doneChan := make(chan struct{}, 2)

	// read from stdout line by line
	go func() {
		defer func() { doneChan <- struct{}{} }()
		scanner := bufio.NewScanner(stdoutPipe)

		for scanner.Scan() {
			lineChan <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("stdout scan error: %w", err)
		}
	}()

	// read from stderr line by line
	go func() {
		defer func() { doneChan <- struct{}{} }()
		scanner := bufio.NewScanner(stderrPipe)

		for scanner.Scan() {
			lineChan <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("stderr scan error: %w", err)
		}
	}()

	// main loop to process output and errors
	for readersDone := 0; readersDone < 2; {
		select {
		case line := <-lineChan:
			// add line to command lines buffer
			t.mx.Lock()
			t.cmdLines = append(t.cmdLines, line)
			t.updateViewpoint()
			t.mx.Unlock()

		case err := <-errorChan:
			// handle read errors
			t.mx.Lock()
			handleError("error reading output", err)

			// try to kill process if it's still running
			if t.cmd != nil && t.cmd.Process != nil {
				if killErr := t.cmd.Process.Kill(); killErr != nil {
					handleError("failed to terminate process", killErr)
				} else {
					t.contents = append(t.contents, "process terminated")
				}
			}
			t.mx.Unlock()

		case <-doneChan:
			readersDone++
		}
	}

	// drain any remaining output
	for len(lineChan) > 0 {
		line := <-lineChan
		t.mx.Lock()
		t.cmdLines = append(t.cmdLines, line)
		t.updateViewpoint()
		t.mx.Unlock()
	}
}

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

// cleanup properly releases all terminal resources (must be called with lock held)
func (t *terminal) cleanup() {
	if t.tty != nil {
		_ = t.tty.Close()
		t.tty = nil
	}
	if t.pty != nil {
		_ = t.pty.Close()
		t.pty = nil
	}
	if t.stdinPipe != nil {
		_ = t.stdinPipe.Close()
		t.stdinPipe = nil
	}
	if t.vt != nil {
		t.contents = append(t.contents, t.vt.Dump(!t.noStyled)...)
		t.contents = append(t.contents, "")
		t.vt = nil
	}
	if t.cmdLines != nil {
		t.contents = append(t.contents, t.cmdLines...)
		t.contents = append(t.contents, "")
		t.cmdLines = nil
	}
	t.cmd = nil
}

func (t *terminal) updateViewpoint() {
	// read from term
	var lines []string
	if t.vt != nil {
		lines = t.vt.Dump(!t.noStyled)
	} else if t.cmdLines != nil {
		lines = t.cmdLines
	}

	ws := t.getWinSize()
	style := lipgloss.NewStyle().Width(int(ws.Cols))
	t.viewport.SetContent(style.Render(strings.Join(append(t.contents, lines...), "\n")))
	if t.autoScroll {
		t.viewport.GotoBottom()
	}

	t.notifyUpdate()
}

// notifyUpdate sends an update notification to the UI (non-blocking)
func (t *terminal) notifyUpdate() {
	if t.notifier != nil {
		t.notifier.release()
	}
}

func (t *terminal) Append(content string) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.contents = append(t.contents, content)
	t.updateViewpoint()
}

func (t *terminal) Clear() {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.contents = []string{}
	t.updateViewpoint()
}

func (t *terminal) SetSize(width, height int) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.setSize(width, height)
	t.updateViewpoint()
}

func (t *terminal) setSize(width, height int) {
	t.viewport.Width = max(width, terminalMinWidth)
	t.viewport.Height = max(height, terminalMinHeight)

	ws := t.getWinSize()

	if t.pty != nil {
		_ = pty.Setsize(t.pty, ws) // best effort
	}

	if t.vt != nil {
		t.vt.Resize(int(ws.Cols), int(ws.Rows))
	}
}

func (t *terminal) getWinSize() *pty.Winsize {
	dx, dy := t.viewport.Style.GetFrameSize()
	width, height := t.viewport.Width-dx, t.viewport.Height-dy

	return &pty.Winsize{
		Rows: uint16(height),
		Cols: uint16(width),
		X:    uint16(width * 8),
		Y:    uint16(height * 16),
	}
}

func (t *terminal) GetSize() (width, height int) {
	t.mx.Lock()
	defer t.mx.Unlock()

	return t.viewport.Width, t.viewport.Height
}

func (t *terminal) ID() string {
	return t.id
}

func (t *terminal) IsRunning() bool {
	t.mx.Lock()
	defer t.mx.Unlock()

	return t.cmd != nil && (t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited())
}

func (t *terminal) Wait() {
	t.wg.Wait()
}

func (t *terminal) Init() tea.Cmd {
	// acquire single active subscription; return nil if one is already active
	if t.notifier == nil {
		t.notifier = newUpdateNotifier()
	}

	return waitForTerminalUpdate(t.notifier, t.id)
}

func (t *terminal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	t.mx.Lock()
	defer t.mx.Unlock()

	switch msg := msg.(type) {
	case TerminalUpdateMsg:
		if msg.ID != t.id {
			return t, nil // ignore messages from other terminals
		}

		// content was updated, start listening for next update
		// when autoPoll is enabled, we always resubscribe to updates
		// to capture future appends or command outputs
		if t.autoPoll {
			return t, waitForTerminalUpdate(t.notifier, t.id)
		}

		return t, nil

	case tea.WindowSizeMsg:
		t.setSize(msg.Width, msg.Height)
		return t, nil

	case tea.KeyMsg:
		if t.handleTerminalInput(msg) {
			return t, nil
		}

	case tea.MouseMsg:
		// TODO: handle mouse events in terminal while running command
	}

	var cmd tea.Cmd
	// update viewport for scrolling
	t.viewport, cmd = t.viewport.Update(msg)

	return t, cmd
}

func (t *terminal) handleTerminalInput(msg tea.KeyMsg) bool {
	if t.cmd == nil {
		return false
	}

	switch msg.Type {
	// use these keys to scroll the viewport
	case tea.KeyPgUp, tea.KeyPgDown, tea.KeyHome, tea.KeyEnd:
		return false
	}

	// for pty mode, use virtual terminal key handling
	if t.vt != nil {
		keyEvent := teaKeyToUVKey(msg)
		if keyEvent == nil {
			return false
		}
		t.vt.SendKey(keyEvent)
		return true
	}

	// for non-pty mode (cmd), write directly to stdin pipe
	if t.stdinPipe != nil {
		var data []byte
		switch msg.Type {
		case tea.KeyRunes:
			data = []byte(string(msg.Runes))
		case tea.KeyEnter:
			data = []byte("\n")
		case tea.KeySpace:
			data = []byte(" ")
		case tea.KeyTab:
			data = []byte("\t")
		case tea.KeyBackspace:
			data = []byte("\b")
		case tea.KeyCtrlC:
			data = []byte("\x03")
		case tea.KeyCtrlD:
			data = []byte("\x04")
		}

		if len(data) > 0 {
			if _, err := t.stdinPipe.Write(data); err != nil {
				// handle write error silently for now
				return false
			}
			return true
		}
	}

	return false
}

func (t *terminal) View() string {
	t.mx.Lock()
	defer t.mx.Unlock()

	return t.viewport.View()
}

// RestoreModel may return nil if the model is not a terminal model
func RestoreModel(model tea.Model) Terminal {
	if t, ok := model.(*terminal); ok {
		return t
	}
	return nil
}
