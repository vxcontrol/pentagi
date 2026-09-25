package terminal

import (
	"sync"
	"testing"
	"time"
)

func TestTeacmd_UpdateNotifier_GivesOneWaiterAChannelThatReleaseCloses(t *testing.T) {
	for _, tc := range []struct {
		name         string
		releaseFirst bool
	}{
		{"a new notifier", false},
		{"a notifier released while nobody waited", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := newUpdateNotifier()
			if tc.releaseFirst {
				n.release()
			}

			for round := 1; round <= 2; round++ {
				ch, ok := n.acquire()
				if !ok || ch == nil {
					t.Fatalf("round %d: acquire = (%v, %v), want a channel", round, ch, ok)
				}
				if _, ok := n.acquire(); ok {
					t.Fatalf("round %d: a second waiter acquired before the release", round)
				}
				select {
				case <-ch:
					t.Fatalf("round %d: the channel was closed before the release", round)
				default:
				}

				n.release()
				select {
				case <-ch:
				default:
					t.Fatalf("round %d: the release left the waiter's channel open", round)
				}
			}
		})
	}
}

func TestTeacmd_UpdateNotifier_CloseEndsTheWaitAndRefusesLaterWaiters(t *testing.T) {
	n := newUpdateNotifier()
	ch, ok := n.acquire()
	if !ok {
		t.Fatal("the first acquire failed")
	}

	n.close()
	select {
	case <-ch:
	default:
		t.Fatal("close left the waiter's channel open")
	}
	if _, ok := n.acquire(); ok {
		t.Error("a waiter acquired a closed notifier")
	}

	// release and a second close must be harmless on a closed notifier
	n.release()
	n.close()
}

func TestTeacmd_UpdateNotifier_LetsOneOfManyConcurrentWaitersIn(t *testing.T) {
	n := newUpdateNotifier()
	won := make(chan (<-chan struct{}), 10)
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if ch, ok := n.acquire(); ok {
				won <- ch
			}
		})
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the concurrent acquires never returned")
	}

	if len(won) != 1 {
		t.Fatalf("%d of 10 concurrent waiters acquired, want 1", len(won))
	}
	ch := <-won
	n.release()
	select {
	case <-ch:
	default:
		t.Fatal("the release did not reach the waiter that won")
	}
}

func TestTeacmd_WaitForTerminalUpdate_WakesOnlyTheWaiterHoldingTheNotifier(t *testing.T) {
	n := newUpdateNotifier()
	if _, ok := n.acquire(); !ok {
		t.Fatal("the first acquire failed")
	}
	if msg := terminalReceive(t, terminalAsync(waitForTerminalUpdate(n, "term-1")), nil); msg != nil {
		t.Errorf("a second waiter got %#v, want nil", msg)
	}

	n.release()
	winner := terminalAsync(waitForTerminalUpdate(n, "term-1"))
	terminalQuiet(t, winner)
	if msg := terminalReceive(t, winner, n.release); msg != (TerminalUpdateMsg{ID: "term-1"}) {
		t.Errorf("the waiter got %#v, want the update of term-1", msg)
	}
}
