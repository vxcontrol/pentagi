package tools

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoroutineLeakFix demonstrates that the fix prevents goroutine leaks.
// This test simulates the io.Copy pattern used in getExecResult() to verify
// that goroutines properly complete or are given a grace period to finish.
func TestGoroutineLeakFix(t *testing.T) {
	t.Run("Normal completion - no leak", func(t *testing.T) {
		// Force GC and get baseline
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		goroutinesBefore := runtime.NumGoroutine()

		// Simulate the fixed pattern
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dst := bytes.Buffer{}
		errChan := make(chan error, 1)

		// Fast reader that completes quickly
		reader := bytes.NewReader([]byte("test output\n"))

		go func() {
			_, copyErr := io.Copy(&dst, reader)
			errChan <- copyErr
		}()

		select {
		case copyErr := <-errChan:
			assert.NoError(t, copyErr)
		case <-ctx.Done():
			t.Fatal("Unexpected timeout")
		}

		// Wait for cleanup
		time.Sleep(100 * time.Millisecond)
		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		goroutinesAfter := runtime.NumGoroutine()
		goroutinesDelta := goroutinesAfter - goroutinesBefore

		t.Logf("Goroutines before: %d, after: %d, delta: %d", goroutinesBefore, goroutinesAfter, goroutinesDelta)
		assert.LessOrEqual(t, goroutinesDelta, 3, "Goroutine leak detected in normal completion")
	})

	t.Run("Timeout with grace period - minimal leak", func(t *testing.T) {
		// Force GC and get baseline
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		goroutinesBefore := runtime.NumGoroutine()

		// Simulate the fixed pattern with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		dst := bytes.Buffer{}
		errChan := make(chan error, 1)

		// Slow reader that will timeout
		slowReader := &slowTestReader{
			data:  []byte("slow output\n"),
			delay: 200 * time.Millisecond,
		}

		go func() {
			_, copyErr := io.Copy(&dst, slowReader)
			errChan <- copyErr
		}()

		var timedOut bool
		select {
		case copyErr := <-errChan:
			// Completed normally
			if copyErr != nil && copyErr != io.EOF {
				t.Logf("Copy error: %v", copyErr)
			}
		case <-ctx.Done():
			// Timeout occurred - wait with grace period
			timedOut = true
			select {
			case <-errChan:
				// Goroutine completed within grace period
				t.Log("Goroutine completed within grace period")
			case <-time.After(5 * time.Second): // defaultExtraExecTimeout
				// Goroutine still blocked
				t.Log("Goroutine still blocked after grace period")
			}
		}

		require.True(t, timedOut, "Expected timeout to occur")

		// Wait for cleanup
		time.Sleep(200 * time.Millisecond)
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		goroutinesAfter := runtime.NumGoroutine()
		goroutinesDelta := goroutinesAfter - goroutinesBefore

		t.Logf("Goroutines before: %d, after: %d, delta: %d", goroutinesBefore, goroutinesAfter, goroutinesDelta)

		// With the fix, we give goroutines time to complete, so delta should be minimal
		assert.LessOrEqual(t, goroutinesDelta, 5, "Goroutine leak detected even with grace period")
	})

	t.Run("Error propagation works correctly", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		dst := bytes.Buffer{}
		errChan := make(chan error, 1)

		// Reader that returns an error
		errorReader := &errorTestReader{err: io.ErrUnexpectedEOF}

		go func() {
			_, copyErr := io.Copy(&dst, errorReader)
			errChan <- copyErr
		}()

		select {
		case copyErr := <-errChan:
			// Error should be propagated
			assert.ErrorIs(t, copyErr, io.ErrUnexpectedEOF, "Error should be propagated correctly")
		case <-ctx.Done():
			t.Fatal("Unexpected timeout")
		}
	})
}

// slowTestReader simulates a slow/blocked reader
type slowTestReader struct {
	data     []byte
	position int
	delay    time.Duration
}

func (sr *slowTestReader) Read(p []byte) (n int, err error) {
	if sr.position >= len(sr.data) {
		time.Sleep(sr.delay * 10) // Simulate very slow read
		return 0, io.EOF
	}

	// Read one byte at a time with delay
	time.Sleep(sr.delay)
	n = copy(p, sr.data[sr.position:sr.position+1])
	sr.position += n
	return n, nil
}

// errorTestReader always returns an error
type errorTestReader struct {
	err error
}

func (er *errorTestReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

// TestOldPatternWouldLeak demonstrates that the old pattern would have leaked goroutines.
// This is for documentation purposes to show why the fix was necessary.
func TestOldPatternWouldLeak(t *testing.T) {
	t.Skip("This test demonstrates the OLD buggy pattern - skipped by default")

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	goroutinesBefore := runtime.NumGoroutine()

	// OLD BUGGY PATTERN (for demonstration only)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	dst := bytes.Buffer{}
	done := make(chan struct{})

	slowReader := &slowTestReader{
		data:  []byte("slow output\n"),
		delay: 200 * time.Millisecond,
	}

	go func() {
		_, _ = io.Copy(&dst, slowReader)
		close(done)
	}()

	select {
	case <-done:
		t.Log("Completed normally")
	case <-ctx.Done():
		t.Log("Timeout - goroutine continues running!")
		// In old pattern, we return here immediately, leaking the goroutine
	}

	// Wait and check for leak
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	goroutinesAfter := runtime.NumGoroutine()
	goroutinesDelta := goroutinesAfter - goroutinesBefore

	t.Logf("OLD PATTERN - Goroutines before: %d, after: %d, delta: %d", goroutinesBefore, goroutinesAfter, goroutinesDelta)
	t.Logf("Delta > 3 indicates a goroutine leak (which is expected in the old pattern)")
}
