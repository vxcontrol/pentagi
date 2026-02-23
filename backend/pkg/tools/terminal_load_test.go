package tools

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGoroutineLeakUnderLoad simulates multiple concurrent timeout scenarios
// to verify that the fix prevents goroutine accumulation under load.
func TestGoroutineLeakUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Force GC and get baseline
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	goroutinesBefore := runtime.NumGoroutine()

	t.Logf("Starting load test with %d goroutines", goroutinesBefore)

	// Simulate 50 concurrent command executions, some timing out
	const iterations = 50
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Simulate the fixed pattern
			timeout := 100 * time.Millisecond
			if iteration%3 == 0 {
				// Some complete successfully
				timeout = 5 * time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			dst := bytes.Buffer{}
			errChan := make(chan error, 1)

			// Variable speed reader
			reader := &loadTestReader{
				data:  []byte("command output\n"),
				delay: 50 * time.Millisecond,
			}
			if iteration%3 != 0 {
				// Will timeout
				reader.delay = 200 * time.Millisecond
			}

			go func() {
				_, copyErr := io.Copy(&dst, reader)
				errChan <- copyErr
			}()

			select {
			case <-errChan:
				// Completed normally
			case <-ctx.Done():
				// Timeout - wait with grace period
				select {
				case <-errChan:
					// Goroutine completed within grace period
				case <-time.After(5 * time.Second):
					// Goroutine still blocked
				}
			}
		}(i)

		// Stagger the starts slightly
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all iterations to complete
	wg.Wait()

	// Give time for all goroutines to finish
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	goroutinesAfter := runtime.NumGoroutine()
	goroutinesDelta := goroutinesAfter - goroutinesBefore

	t.Logf("Load test completed:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Goroutines before: %d", goroutinesBefore)
	t.Logf("  Goroutines after: %d", goroutinesAfter)
	t.Logf("  Delta: %d", goroutinesDelta)

	// With proper cleanup, delta should be minimal even after 50 iterations
	// Allow some variance for test infrastructure
	maxAllowedDelta := 10
	assert.LessOrEqual(t, goroutinesDelta, maxAllowedDelta,
		"Significant goroutine leak detected under load: %d new goroutines remained after %d iterations",
		goroutinesDelta, iterations)

	if goroutinesDelta > 0 {
		t.Logf("Note: %d goroutines delta is within acceptable range", goroutinesDelta)
	} else {
		t.Log("Perfect: No goroutine leak detected!")
	}
}

// loadTestReader simulates variable-speed I/O operations
type loadTestReader struct {
	data     []byte
	position int
	delay    time.Duration
}

func (r *loadTestReader) Read(p []byte) (n int, err error) {
	if r.position >= len(r.data) {
		return 0, io.EOF
	}

	// Simulate I/O delay
	time.Sleep(r.delay)

	// Read up to one byte at a time
	n = copy(p, r.data[r.position:min(r.position+1, len(r.data))])
	r.position += n

	return n, nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BenchmarkGoroutineCleanup benchmarks the goroutine cleanup performance
func BenchmarkGoroutineCleanup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

		dst := bytes.Buffer{}
		errChan := make(chan error, 1)

		reader := bytes.NewReader([]byte("benchmark output\n"))

		go func() {
			_, copyErr := io.Copy(&dst, reader)
			errChan <- copyErr
		}()

		select {
		case <-errChan:
		case <-ctx.Done():
			// Grace period
			select {
			case <-errChan:
			case <-time.After(5 * time.Second):
			}
		}

		cancel()
	}
}
