package terminal

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// terminalAsync runs a command off the test goroutine, as the bubbletea runtime does.
func terminalAsync(cmd tea.Cmd) <-chan tea.Msg {
	out := make(chan tea.Msg, 1)
	go func() { out <- cmd() }()
	return out
}

// terminalQuiet fails when the command answers before any update; a lower bound, so a slow runner passes.
func terminalQuiet(t *testing.T, out <-chan tea.Msg) {
	t.Helper()
	select {
	case msg := <-out:
		t.Fatalf("the waiter answered %#v before any update", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

// terminalReceive pokes until the command answers: a waiter takes the notifier when the test cannot see.
func terminalReceive(t *testing.T, out <-chan tea.Msg, poke func()) tea.Msg {
	t.Helper()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case msg := <-out:
			return msg
		case <-tick.C:
			if poke != nil {
				poke()
			}
		case <-timeout:
			t.Fatal("the command never returned")
			return nil
		}
	}
}
