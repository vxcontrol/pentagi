package terminal

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

var terminalModes = []struct {
	name string
	opts []TerminalOption
}{
	{"through a pty", nil},
	{"through pipes", []TerminalOption{WithNoPty()}},
}

func terminalView(term Terminal) string {
	return ansi.Strip(term.View())
}

// terminalFinish waits for the command as the processor does, then for the terminal to read all output.
func terminalFinish(t *testing.T, term Terminal, cmd *exec.Cmd) error {
	t.Helper()
	errc := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		term.Wait()
		errc <- err
	}()
	select {
	case err := <-errc:
		return err
	case <-time.After(10 * time.Second):
		t.Fatalf("%v never finished", cmd.Args)
		return nil
	}
}

func terminalKillOnCleanup(t *testing.T, term Terminal, cmd *exec.Cmd) {
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = terminalFinish(t, term, cmd)
	})
}

func terminalEventually(t *testing.T, cond func() bool, why string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal(why)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// terminalStdin records what a piped command would read.
type terminalStdin struct{ bytes.Buffer }

func (*terminalStdin) Close() error { return nil }

func TestTerminal_SetSize_ChangesTheReportedSize(t *testing.T) {
	term := NewTerminal(80, 24)
	if w, h := term.GetSize(); w != 80 || h != 24 {
		t.Fatalf("a new terminal reports %dx%d, want 80x24", w, h)
	}

	term.SetSize(100, 30)
	if w, h := term.GetSize(); w != 100 || h != 30 {
		t.Errorf("after SetSize(100, 30) the terminal reports %dx%d", w, h)
	}
}

func TestTerminal_Append_ShowsTheLineInTheView(t *testing.T) {
	term := NewTerminal(80, 24)
	term.Append("test message")

	if !strings.Contains(terminalView(term), "test message") {
		t.Error("appended message not found in view")
	}
}

func TestTerminal_Clear_RemovesAppendedLines(t *testing.T) {
	term := NewTerminal(80, 24)
	term.Append("test message")
	term.Clear()

	if strings.Contains(terminalView(term), "test message") {
		t.Error("message found after clear")
	}
}

func TestTerminal_RestoreModel_AcceptsOnlyATerminal(t *testing.T) {
	term := NewTerminal(80, 24)
	if RestoreModel(term) != term {
		t.Error("RestoreModel did not return the terminal it was given")
	}
	if RestoreModel(&struct{ tea.Model }{}) != nil {
		t.Error("RestoreModel accepted a model that is not a terminal")
	}
}

func TestTerminal_Execute_RefusesASecondCommandWhileOneRuns(t *testing.T) {
	term := NewTerminal(80, 24)
	first := exec.Command("cat")
	if err := term.Execute(first); err != nil {
		t.Fatalf("first Execute: %v", err)
	}
	terminalKillOnCleanup(t, term, first)

	err := term.Execute(exec.Command("echo", "second"))
	if err == nil || !strings.Contains(err.Error(), "already executing") {
		t.Errorf("second Execute = %v, want the terminal to say it is already executing", err)
	}
}

func TestTerminal_Execute_ShowsWhyACommandCouldNotStart(t *testing.T) {
	for _, mode := range terminalModes {
		t.Run(mode.name, func(t *testing.T) {
			term := NewTerminal(80, 24, mode.opts...)
			if err := term.Execute(exec.Command("/nonexistent/pentagi-installer-test")); err == nil {
				t.Fatal("Execute started a command that does not exist")
			}

			if view := terminalView(term); !strings.Contains(view, "failed to execute command") {
				t.Errorf("the view does not say why nothing ran: %q", view)
			}
			if term.IsRunning() {
				t.Error("the terminal still counts the failed command as running")
			}
		})
	}
}

func TestTerminal_Wait_ReadiesTheTerminalForTheNextCommand(t *testing.T) {
	for _, mode := range terminalModes {
		t.Run(mode.name, func(t *testing.T) {
			term := NewTerminal(80, 24, mode.opts...)
			// the second command writes to stderr, so both streams reach the view in both modes
			for _, script := range []string{"echo stdout-one", "echo stderr-two >&2"} {
				cmd := exec.Command("sh", "-c", script)
				if err := term.Execute(cmd); err != nil {
					t.Fatalf("Execute(%q) after the previous command was waited for: %v", script, err)
				}
				if err := terminalFinish(t, term, cmd); err != nil {
					t.Fatalf("%q: %v", script, err)
				}
			}

			if view := terminalView(term); !strings.Contains(view, "stdout-one") || !strings.Contains(view, "stderr-two") {
				t.Errorf("expected outputs not found in view: %q", view)
			}
		})
	}
}

func TestTerminal_Update_SendsKeysToTheRunningCommand(t *testing.T) {
	for _, mode := range terminalModes {
		t.Run(mode.name, func(t *testing.T) {
			term := NewTerminal(80, 24, mode.opts...)
			cmd := exec.Command("cat")
			if err := term.Execute(cmd); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			terminalKillOnCleanup(t, term, cmd)
			if !term.IsRunning() {
				t.Fatal("the terminal does not count cat as running")
			}

			// in pipe mode the output goroutine attaches stdin, and a key sent before that is dropped
			impl := term.(*terminal)
			terminalEventually(t, func() bool {
				impl.mx.Lock()
				defer impl.mx.Unlock()
				return impl.vt != nil || impl.stdinPipe != nil
			}, "the command's input was never attached")

			for _, r := range "hello world" {
				if r == ' ' {
					term.Update(tea.KeyMsg{Type: tea.KeySpace})
				} else {
					term.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
				}
			}
			term.Update(tea.KeyMsg{Type: tea.KeyEnter})
			terminalEventually(t, func() bool { return strings.Contains(terminalView(term), "hello world") },
				"the typed line never reached the view")
		})
	}
}

func TestTerminal_HandleTerminalInput_WritesMappedKeysToThePipe(t *testing.T) {
	for _, tc := range []struct {
		name    string
		key     tea.KeyMsg
		written string
	}{
		{"typed runes are written as text", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hi")}, "hi"},
		{"space is written as a space", tea.KeyMsg{Type: tea.KeySpace}, " "},
		{"enter ends the line", tea.KeyMsg{Type: tea.KeyEnter}, "\n"},
		{"tab is written as a tab", tea.KeyMsg{Type: tea.KeyTab}, "\t"},
		{"backspace is written as a backspace", tea.KeyMsg{Type: tea.KeyBackspace}, "\b"},
		{"ctrl+c is written as an interrupt", tea.KeyMsg{Type: tea.KeyCtrlC}, "\x03"},
		{"ctrl+d is written as end of input", tea.KeyMsg{Type: tea.KeyCtrlD}, "\x04"},
		{"a key without a mapping is left to the viewport", tea.KeyMsg{Type: tea.KeyF1}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdin := &terminalStdin{}
			term := &terminal{cmd: &exec.Cmd{}, stdinPipe: stdin}

			if handled := term.handleTerminalInput(tc.key); handled != (tc.written != "") {
				t.Errorf("handled = %v, want %v", handled, tc.written != "")
			}
			if got := stdin.String(); got != tc.written {
				t.Errorf("the command read %q, want %q", got, tc.written)
			}
		})
	}
}

func TestTerminal_Init_WaitsForThisTerminalsNextUpdate(t *testing.T) {
	term := NewTerminal(80, 24)
	held, _ := term.(*terminal).notifier.acquire()
	if msg := terminalReceive(t, terminalAsync(term.Init()), nil); msg != nil {
		t.Errorf("a second Init while one waits got %#v, want nil", msg)
	}

	term.Append("wakes the waiting Init")
	select {
	case <-held:
	default:
		t.Fatal("Append did not wake the waiting Init")
	}

	next := terminalAsync(term.Init())
	terminalQuiet(t, next)
	if msg := terminalReceive(t, next, func() { term.Append("line") }); msg != (TerminalUpdateMsg{ID: term.ID()}) {
		t.Errorf("Init answered %#v, want this terminal's update", msg)
	}
}

func TestTerminal_Update_ListensForTheNextUpdateOnlyUnderAutoPoll(t *testing.T) {
	for _, tc := range []struct {
		name    string
		opts    []TerminalOption
		otherID bool
		listens bool
	}{
		{"its own update under auto-poll", []TerminalOption{WithAutoPoll()}, false, true},
		{"another terminal's update under auto-poll", []TerminalOption{WithAutoPoll()}, true, false},
		{"its own update without auto-poll", nil, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			term := NewTerminal(80, 24, tc.opts...)
			id := term.ID()
			if tc.otherID {
				id = "other"
			}

			model, next := term.Update(TerminalUpdateMsg{ID: id})
			if model != term {
				t.Error("Update returned another model")
			}
			if (next != nil) != tc.listens {
				t.Errorf("Update returned a next command: %v, want %v", next != nil, tc.listens)
			}
		})
	}
}

func TestTerminal_TerminalFinalizer_ReleasesTheWaiterAndDropsTheNotifier(t *testing.T) {
	impl := NewTerminal(80, 24).(*terminal)
	held, _ := impl.notifier.acquire()
	impl.cmd, impl.stdinPipe = &exec.Cmd{}, &terminalStdin{}

	terminalFinalizer(impl)

	select {
	case <-held:
	default:
		t.Error("the finalizer left the waiting Init blocked")
	}
	if impl.notifier != nil {
		t.Error("the finalizer kept the notifier")
	}
	if impl.cmd != nil || impl.stdinPipe != nil {
		t.Errorf("the finalizer kept the command (%t) or its input (%t)", impl.cmd != nil, impl.stdinPipe != nil)
	}
}

func TestTerminal_NewTerminal_ReleasesTheWaiterOnceTheTerminalIsCollected(t *testing.T) {
	// only the notifier and the ID escape into the command, so the terminal itself is garbage
	waiter := terminalAsync(NewTerminal(80, 24).Init())
	terminalReceive(t, waiter, runtime.GC)
}

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
