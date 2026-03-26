package terminal

import (
	"sync"
	"testing"
	"time"
)

// Tests for update notifier functionality

func TestUpdateNotifierSingleSubscriber(t *testing.T) {
	b := newUpdateNotifier()

	// first subscribe should return context
	ch1, ok := b.acquire()
	if !ok || ch1 == nil {
		t.Fatal("acquire should return a channel and true")
	}

	// release should signal the context
	go func() {
		time.Sleep(50 * time.Millisecond)
		b.release()
	}()

	select {
	case <-ch1:
		// success - context was signalled
	case <-time.After(100 * time.Millisecond):
		t.Error("subscriber did not release")
	}
}

func TestUpdateNotifierOnlyOneAcquirer(t *testing.T) {
	b := newUpdateNotifier()

	// multiple subscribers should get the same context
	ch1, ok1 := b.acquire()
	_, ok2 := b.acquire()
	_, ok3 := b.acquire()

	// verify they are the same context
	if !(ok1 && ch1 != nil) || ok2 || ok3 {
		t.Error("only first acquire should succeed; others should fail")
	}

	// release should signal the shared context
	go func() {
		time.Sleep(50 * time.Millisecond)
		b.release()
	}()

	// all should receive the release
	timeout := time.After(100 * time.Millisecond)
	received := 0

	ch := ch1
	for received < 1 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Errorf("only %d out of 1 subscribers received release", received)
			return
		}
	}
}

func TestUpdateNotifierNewAcquireAfterRelease(t *testing.T) {
	b := newUpdateNotifier()

	// first subscribe
	ch1, ok := b.acquire()
	if !ok || ch1 == nil {
		t.Fatal("first acquire should succeed")
	}

	// release cancels the context
	b.release()

	// verify channel is closed by reading from it
	select {
	case <-ch1:
		// success - channel closed
	case <-time.After(10 * time.Millisecond):
		t.Error("channel should be closed after release")
	}

	// new subscribe should create new context
	// next acquire should succeed again
	ch2, ok := b.acquire()
	if !ok || ch2 == nil {
		t.Error("acquire should succeed after release")
	}
	// and should be signalled on next release
	go func() {
		time.Sleep(20 * time.Millisecond)
		b.release()
	}()
	select {
	case <-ch2:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("acquired channel should be closed after subsequent release")
	}
}

func TestUpdateNotifierAfterClose(t *testing.T) {
	b := newUpdateNotifier()

	// subscribe before close
	ch1, ok := b.acquire()
	if !ok || ch1 == nil {
		t.Fatal("first acquire should succeed")
	}

	// close notifier
	b.close()

	// verify existing wait channel is closed
	select {
	case <-ch1:
		// success
	case <-time.After(10 * time.Millisecond):
		t.Error("existing wait should be closed after notifier close")
	}

	// new subscribe after close should return same cancelled context
	if _, ok := b.acquire(); ok {
		t.Error("acquire after close should fail")
	}

	// releasing after close should not panic
	b.release() // should not panic

	// closing again should not panic
	b.close() // should not panic
}

func TestUpdateNotifierConcurrentAccess(t *testing.T) {
	b := newUpdateNotifier()

	// test concurrent subscribe and release
	var wg sync.WaitGroup
	received := make([]bool, 10)

	// start multiple subscribers concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ch, ok := b.acquire()
			if !ok || ch == nil {
				return
			}
			select {
			case <-ch:
				received[idx] = true
			case <-time.After(200 * time.Millisecond):
				// timeout
			}
		}(i)
	}

	// wait a bit for subscribers to register
	time.Sleep(50 * time.Millisecond)

	// release
	b.release()

	wg.Wait()

	// check all received
	// at least one should have received
	any := false
	for _, recv := range received {
		if recv {
			any = true
		}
	}
	if !any {
		t.Errorf("no subscriber received release")
	}
}

func TestUpdateNotifierReleaseWithoutAcquirer(t *testing.T) {
	b := newUpdateNotifier()

	// release without any active subscribers should not panic
	b.release() // should not panic

	// first subscribe after empty release should still work
	ch, ok := b.acquire()
	if !ok || ch == nil {
		t.Fatal("acquire should succeed")
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		b.release()
	}()

	select {
	case <-ch:
		// success
	case <-time.After(100 * time.Millisecond):
		t.Error("subscribe after empty release should still work")
	}
}

func TestUpdateNotifierContextReuse(t *testing.T) {
	b := newUpdateNotifier()

	// multiple subscribers should get same context
	ch1, ok1 := b.acquire()
	_, ok2 := b.acquire()
	_, ok3 := b.acquire()

	// verify they are exactly the same context
	if !(ok1 && ch1 != nil) || ok2 || ok3 {
		t.Error("only first acquire should succeed; others should fail")
	}

	// release should cancel the context (close channel)
	b.release()

	// all context.Done() channels should be closed
	select {
	case <-ch1:
		// success
	case <-time.After(10 * time.Millisecond):
		t.Error("ch1 should be closed after release")
	}

	// no ch2

	// no ch3
}

func TestUpdateNotifierStartsWithActiveChannel(t *testing.T) {
	b := newUpdateNotifier()

	// first subscribe should create new active context (since initial is cancelled)
	ch1, ok := b.acquire()
	if !ok || ch1 == nil {
		t.Fatal("first acquire should succeed")
	}

	// channel must not be closed before release
	select {
	case <-ch1:
		t.Error("first acquire should create active channel")
	case <-time.After(10 * time.Millisecond):
		// success - channel is active
	}

	// second subscribe should return same context
	// second acquire should fail until release
	if _, ok := b.acquire(); ok {
		t.Error("second acquire should fail before release")
	}

	// release should cancel the context
	b.release()

	// context should now be cancelled
	select {
	case <-ch1:
		// success
	case <-time.After(10 * time.Millisecond):
		t.Error("channel should be closed after release")
	}

	// new subscribe after release should create new context
	if _, ok := b.acquire(); !ok {
		t.Error("acquire after release should succeed")
	}
}

func TestWaitForTerminalUpdate(t *testing.T) {
	b := newUpdateNotifier()

	// create command
	cmd := waitForTerminalUpdate(b, "test")
	if cmd == nil {
		t.Fatal("waitForTerminalUpdate should return a command")
	}

	// run command in goroutine
	msgChan := make(chan any, 1)
	go func() {
		msg := cmd()
		msgChan <- msg
	}()

	// wait a bit then release
	time.Sleep(50 * time.Millisecond)
	b.release()

	// check we received the message
	select {
	case msg := <-msgChan:
		if _, ok := msg.(TerminalUpdateMsg); !ok {
			t.Errorf("expected TerminalUpdateMsg, got %T", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("waitForTerminalUpdate command did not return")
	}
}

// ensure only a single waiter receives TerminalUpdateMsg and others get nil
func TestWaitForTerminalUpdateSingleWinner(t *testing.T) {
	b := newUpdateNotifier()

	cmd1 := waitForTerminalUpdate(b, "test")
	cmd2 := waitForTerminalUpdate(b, "test")

	if cmd1 == nil || cmd2 == nil {
		t.Fatal("waitForTerminalUpdate should return non-nil commands")
	}

	msgCh := make(chan any, 2)

	go func() { msgCh <- cmd1() }()
	go func() { msgCh <- cmd2() }()

	// release once – only one waiter must win
	time.Sleep(20 * time.Millisecond)
	b.release()

	// collect both results
	var msgs []any
	timeout := time.After(200 * time.Millisecond)
	for len(msgs) < 2 {
		select {
		case m := <-msgCh:
			msgs = append(msgs, m)
		case <-timeout:
			t.Fatal("timeout waiting for waiter results")
		}
	}

	wins := 0
	nils := 0
	for _, m := range msgs {
		if m == nil {
			nils++
			continue
		}
		if _, ok := m.(TerminalUpdateMsg); ok {
			wins++
		}
	}

	if wins != 1 {
		t.Errorf("expected exactly 1 TerminalUpdateMsg winner, got %d", wins)
	}
	if nils != 1 {
		t.Errorf("expected exactly 1 nil message for loser, got %d", nils)
	}
}
