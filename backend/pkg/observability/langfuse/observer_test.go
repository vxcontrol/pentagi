package langfuse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/observability/langfuse/api"
)

// Shutdown cancels without flushing, so the process shutdown path must ForceFlush
// (Observer.Flush) before Shutdown or the last batch of observations is dropped.
// These pin both halves so a change to either method surfaces here.
func TestObserverShutdownDropsBatch_ForceFlushDrains(t *testing.T) {
	// GET serves the project list that NewClient validates against; POST is the
	// ingestion batch endpoint whose hits mean the batch was actually drained.
	newSink := func() (*httptest.Server, *int32) {
		var ingestHits int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte(`{"data":[{"id":"proj","name":"test"}]}`))
				return
			}
			atomic.AddInt32(&ingestHits, 1)
			_, _ = w.Write([]byte(`{"successes":[],"errors":[]}`))
		}))
		return srv, &ingestHits
	}

	makeObserver := func(t *testing.T, url string) *observer {
		t.Helper()
		client, err := NewClient(
			WithBaseURL(url),
			WithPublicKey("pk"),
			WithSecretKey("sk"),
			WithProjectID("proj"),
		)
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		// Huge interval + queue so nothing auto-flushes: the batch accumulates and
		// is drained only by the method under test.
		return NewObserver(client,
			WithSendInterval(10*time.Minute),
			WithSendTimeout(2*time.Second),
			WithQueueSize(100),
		).(*observer)
	}

	newEvent := func() *api.IngestionEvent {
		return &api.IngestionEvent{IngestionEventZero: &api.IngestionEventZero{
			ID:        newSpanID(),
			Timestamp: getCurrentTimeString(),
		}}
	}

	// Block until the sender goroutine has pulled the enqueued event out of the
	// buffered queue and into its in-memory batch, so the subsequent action acts
	// on a non-empty batch deterministically.
	waitBatched := func(t *testing.T, o *observer) {
		t.Helper()
		for i := 0; i < 400; i++ {
			if len(o.queue) == 0 {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatal("sender never consumed the queued event into its batch")
	}

	t.Run("Shutdown alone drops the buffered batch", func(t *testing.T) {
		srv, hits := newSink()
		defer srv.Close()
		o := makeObserver(t, srv.URL)
		o.enqueue(newEvent())
		waitBatched(t, o)

		_ = o.Shutdown(context.Background())

		if got := atomic.LoadInt32(hits); got != 0 {
			t.Fatalf("Shutdown must not flush; sink received %d requests", got)
		}
	})

	t.Run("ForceFlush drains the buffered batch", func(t *testing.T) {
		srv, hits := newSink()
		defer srv.Close()
		o := makeObserver(t, srv.URL)
		o.enqueue(newEvent())
		waitBatched(t, o)

		_ = o.ForceFlush(context.Background())

		if got := atomic.LoadInt32(hits); got != 1 {
			t.Fatalf("ForceFlush must send the batch exactly once; sink received %d requests", got)
		}
		_ = o.Shutdown(context.Background())
	})
}
