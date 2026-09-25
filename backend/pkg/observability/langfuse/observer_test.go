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

// observerSink serves the project list NewClient validates against and hands every ingestion POST to ingest.
func observerSink(t *testing.T, ingest http.HandlerFunc) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"proj","name":"test"}]}`))
			return
		}
		ingest(w, r)
	}))
	t.Cleanup(srv.Close)

	return srv
}

// observerNew never flushes on its own, so a batch is sent only by the call under test.
func observerNew(t *testing.T, url string) *observer {
	t.Helper()

	client, err := NewClient(WithBaseURL(url), WithPublicKey("pk"), WithSecretKey("sk"), WithProjectID("proj"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return NewObserver(client,
		WithSendInterval(10*time.Minute),
		WithSendTimeout(2*time.Second),
		WithQueueSize(100),
	).(*observer)
}

func observerEvent(id string) *api.IngestionEvent {
	return &api.IngestionEvent{IngestionEventZero: &api.IngestionEventZero{ID: id, Timestamp: getCurrentTimeString()}}
}

// observerWaitBatched waits until the sender has moved every queued event into its in-memory batch.
func observerWaitBatched(t *testing.T, o *observer) {
	t.Helper()

	for range 400 {
		if len(o.queue) == 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("sender never consumed the queued events into its batch")
}

// observerCountingSink accepts every batch and counts the requests.
func observerCountingSink(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var hits atomic.Int32
	srv := observerSink(t, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"successes":[],"errors":[]}`))
	})
	return srv, &hits
}

// Shutdown cancels without flushing, so the process must ForceFlush first or the last batch is lost.
func TestObserver_Shutdown_DropsTheBufferedBatch(t *testing.T) {
	srv, hits := observerCountingSink(t)
	o := observerNew(t, srv.URL)
	o.enqueue(observerEvent(newSpanID()))
	observerWaitBatched(t, o)

	_ = o.Shutdown(context.Background())

	if got := hits.Load(); got != 0 {
		t.Fatalf("Shutdown must not flush; sink received %d requests", got)
	}
}

func TestObserver_ForceFlush_DrainsTheBufferedBatch(t *testing.T) {
	srv, hits := observerCountingSink(t)
	o := observerNew(t, srv.URL)
	o.enqueue(observerEvent(newSpanID()))
	observerWaitBatched(t, o)

	_ = o.ForceFlush(context.Background())

	if got := hits.Load(); got != 1 {
		t.Fatalf("ForceFlush must send the batch exactly once; sink received %d requests", got)
	}
	_ = o.Shutdown(context.Background())
}

// Langfuse answers an oversized batch with 413; every event must still land once the batch is split.
func TestObserver_FlushWithSplit_SplitsA413InsteadOfDroppingTheBatch(t *testing.T) {
	const maxEventsPerRequest = 2

	var (
		mu          sync.Mutex
		received    = map[string]bool{}
		postCount   atomic.Int32
		rejectCount atomic.Int32
	)

	srv := observerSink(t, func(w http.ResponseWriter, r *http.Request) {
		postCount.Add(1)

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
			rejectCount.Add(1)
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
	})

	o := observerNew(t, srv.URL)

	ids := make([]string, 0, 5)
	for range 5 {
		id := newSpanID()
		ids = append(ids, id)
		o.enqueue(observerEvent(id))
	}
	observerWaitBatched(t, o)

	if err := o.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil (splitting should recover from 413)", err)
	}
	_ = o.Shutdown(context.Background())

	if rejectCount.Load() == 0 {
		t.Fatal("expected at least one 413 rejection to exercise the split path")
	}
	if got := postCount.Load(); got <= 1 {
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

// A single event that alone exceeds the limit is dropped, not split and retried forever.
func TestObserver_FlushWithSplit_DropsAnOversizedSingleEvent(t *testing.T) {
	srv := observerSink(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte("Body exceeded 4.5mb limit"))
	})
	o := observerNew(t, srv.URL)

	done := make(chan error, 1)
	go func() {
		done <- o.flushWithSplit(context.Background(), []*api.IngestionEvent{observerEvent(newSpanID()), observerEvent(newSpanID())})
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
