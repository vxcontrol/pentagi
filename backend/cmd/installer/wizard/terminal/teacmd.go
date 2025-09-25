package terminal

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// updateNotifier coordinates a single waiter for the next terminal update
type updateNotifier struct {
	mx       sync.Mutex
	ch       chan struct{}
	acquired bool
	closed   bool
}

func newUpdateNotifier() *updateNotifier {
	return &updateNotifier{ch: make(chan struct{})}
}

// acquire returns a channel to wait for the next update. Only the first caller
// succeeds; subsequent calls return (nil, false) until the next release.
func (n *updateNotifier) acquire() (<-chan struct{}, bool) {
	n.mx.Lock()
	defer n.mx.Unlock()

	if n.closed {
		return nil, false
	}
	if n.acquired {
		return nil, false
	}
	if n.ch == nil {
		n.ch = make(chan struct{})
	}
	n.acquired = true
	return n.ch, true
}

// release signals the update to the active waiter and resets state
func (n *updateNotifier) release() {
	n.mx.Lock()
	defer n.mx.Unlock()

	if n.closed {
		return
	}
	if n.ch != nil {
		close(n.ch)
		n.ch = nil
	}
	n.acquired = false
}

// close terminates any pending waiter and resets state
func (n *updateNotifier) close() {
	n.mx.Lock()
	defer n.mx.Unlock()
	if n.closed {
		return
	}
	if n.ch != nil {
		close(n.ch)
		n.ch = nil
	}
	n.acquired = false
	n.closed = true
}

// waitForTerminalUpdate returns a command that waits for terminal content updates
func waitForTerminalUpdate(n *updateNotifier, id string) tea.Cmd {
	return func() tea.Msg {
		ch, ok := n.acquire()
		if !ok || ch == nil {
			return nil
		}
		<-ch
		return TerminalUpdateMsg{ID: id}
	}
}
