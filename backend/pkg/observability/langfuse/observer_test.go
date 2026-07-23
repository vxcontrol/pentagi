package langfuse

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
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

// TestObserverFlushWithSplit_413SplitsInstead OfDroppingBatch reproduces the
// production failure: Langfuse rejects an oversized batch with 413 "Body
// exceeded ... limit". Previously the whole batch was discarded on any flush
// error; flushWithSplit must instead split the batch and retry each half so
// every event still lands, as long as no single event alone is oversized.
func TestObserverFlushWithSplit_413SplitsInsteadOfDroppingBatch(t *testing.T) {
	const maxEventsPerRequest = 2

	var (
		mu          sync.Mutex
		received    = map[string]bool{}
		postCount   int32
		rejectCount int32
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"proj","name":"test"}]}`))
			return
		}

		atomic.AddInt32(&postCount, 1)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var req api.IngestionBatchRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to unmarshal batch request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(req.Batch) > maxEventsPerRequest {
			atomic.AddInt32(&rejectCount, 1)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_, _ = w.Write([]byte("Body exceeded 4.5mb limit"))
			return
		}

		mu.Lock()
		for _, event := range req.Batch {
			if zero := event.GetIngestionEventZero(); zero != nil {
				received[zero.ID] = true
			}
		}
		mu.Unlock()

		_, _ = w.Write([]byte(`{"successes":[],"errors":[]}`))
	}))
	defer srv.Close()

	client, err := NewClient(
		WithBaseURL(srv.URL),
		WithPublicKey("pk"),
		WithSecretKey("sk"),
		WithProjectID("proj"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	o := NewObserver(client,
		WithSendInterval(10*time.Minute),
		WithSendTimeout(2*time.Second),
		WithQueueSize(100),
	).(*observer)

	const numEvents = 5
	ids := make([]string, 0, numEvents)
	for i := 0; i < numEvents; i++ {
		id := newSpanID()
		ids = append(ids, id)
		o.enqueue(&api.IngestionEvent{IngestionEventZero: &api.IngestionEventZero{
			ID:        id,
			Timestamp: getCurrentTimeString(),
		}})
	}

	for i := 0; i < 400; i++ {
		if len(o.queue) == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err := o.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil (splitting should recover from 413)", err)
	}
	_ = o.Shutdown(context.Background())

	if got := atomic.LoadInt32(&rejectCount); got == 0 {
		t.Fatal("expected at least one 413 rejection to exercise the split path")
	}
	if got := atomic.LoadInt32(&postCount); got <= 1 {
		t.Fatalf("expected multiple POST requests from splitting, got %d", got)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, id := range ids {
		if !received[id] {
			t.Errorf("event %s was never received by the sink - it was dropped instead of split", id)
		}
	}
}

// TestObserverFlushWithSplit_OversizedSingleEventIsDroppedNotRetriedForever
// checks that when even a single event alone still exceeds the body-size
// limit, flushWithSplit gives up after splitting down to it (returning nil,
// logging the drop) instead of recursing forever.
func TestObserverFlushWithSplit_OversizedSingleEventIsDroppedNotRetriedForever(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"proj","name":"test"}]}`))
			return
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte("Body exceeded 4.5mb limit"))
	}))
	defer srv.Close()

	client, err := NewClient(
		WithBaseURL(srv.URL),
		WithPublicKey("pk"),
		WithSecretKey("sk"),
		WithProjectID("proj"),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	o := NewObserver(client,
		WithSendInterval(10*time.Minute),
		WithSendTimeout(2*time.Second),
		WithQueueSize(100),
	).(*observer)

	batch := []*api.IngestionEvent{
		{IngestionEventZero: &api.IngestionEventZero{ID: newSpanID(), Timestamp: getCurrentTimeString()}},
		{IngestionEventZero: &api.IngestionEventZero{ID: newSpanID(), Timestamp: getCurrentTimeString()}},
	}

	done := make(chan error, 1)
	go func() {
		done <- o.flushWithSplit(context.Background(), batch)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("flushWithSplit() error = %v, want nil (an oversized single event must be dropped, not surfaced)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("flushWithSplit did not return - it likely recursed forever on an always-413 sink")
	}

	_ = o.Shutdown(context.Background())
}
