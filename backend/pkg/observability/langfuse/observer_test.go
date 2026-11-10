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

func observerEvent(id string) *api.OtelSpan {
	return &api.OtelSpan{
		TraceID: newTraceID(),
		SpanID:  id,
		Name:    getStringRef("test"),
	}
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
	srv := observerSink(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/otel/v1/traces" {
			t.Errorf("unexpected ingestion path %q, want /api/public/otel/v1/traces", r.URL.Path)
		}
		if got := r.Header.Get(IngestionVersionHeader); got != IngestionVersion {
			t.Errorf("header %s = %q, want %q", IngestionVersionHeader, got, IngestionVersion)
		}
		hits.Add(1)
		_, _ = w.Write([]byte(`{}`))
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
		var req api.OpentelemetryExportTracesRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to unmarshal OTLP traces request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var spans []*api.OtelSpan
		for _, rs := range req.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				spans = append(spans, ss.Spans...)
			}
		}

		if len(spans) > maxEventsPerRequest {
			rejectCount.Add(1)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_, _ = w.Write([]byte("Body exceeded 4.5mb limit"))
			return
		}

		mu.Lock()
		for _, span := range spans {
			if spanID, ok := span.SpanID.(string); ok {
				received[spanID] = true
			}
		}
		mu.Unlock()

		_, _ = w.Write([]byte(`{}`))
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
		done <- o.flushWithSplit(context.Background(), []*api.OtelSpan{observerEvent(newSpanID()), observerEvent(newSpanID())})
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

func TestObserver_V4Observations_SingleExportOnEndAndPropagatesTraceContext(t *testing.T) {
	var (
		mu    sync.Mutex
		spans []*api.OtelSpan
	)
	srv := observerSink(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var req api.OpentelemetryExportTracesRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to unmarshal OTLP traces request: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		for _, rs := range req.ResourceSpans {
			for _, ss := range rs.ScopeSpans {
				spans = append(spans, ss.Spans...)
			}
		}
		mu.Unlock()
		_, _ = w.Write([]byte(`{}`))
	})

	client, err := NewClient(WithBaseURL(srv.URL), WithPublicKey("pk"), WithSecretKey("sk"), WithProjectID("proj"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	o := NewObserver(client,
		WithRelease("1.2.3"),
		WithEnvironment("staging"),
		WithSendInterval(10*time.Minute),
		WithSendTimeout(2*time.Second),
		WithQueueSize(100),
	).(*observer)
	defer func() { _ = o.Shutdown(context.Background()) }()

	ctx, rootObs := o.NewObservation(context.Background(),
		WithObservationTraceID("trace-v4"),
		WithObservationTraceContext(
			WithTraceName("flow-1"),
			WithTraceUserID("alice"),
			WithTraceSessionID("sess-1"),
			WithTraceTags([]string{"controller", "flow"}),
			WithTraceInput("initial user input"),
			WithTraceMetadata(Metadata{"flow_id": 1}),
		),
	)

	rootSpan := rootObs.Span(WithSpanName("root-span"))
	ctx, childObs := rootSpan.Observation(ctx)
	gen := childObs.Generation(
		WithGenerationName("llm-call"),
		WithGenerationModel("gpt-4o"),
		WithGenerationInput("prompt"),
	)

	// Before End() is called, nothing should be queued (immutable v4 single-export semantics).
	if len(o.queue) != 0 {
		t.Fatalf("open observations must not be queued before End(); got %d queued items", len(o.queue))
	}

	gen.End(
		WithGenerationOutput("response"),
		WithGenerationUsage(&GenerationUsage{Input: 10, Output: 5, Unit: GenerationUsageUnitTokens}),
	)
	// Calling End() a second time must be a no-op.
	gen.End(WithGenerationOutput("duplicate"))

	rootSpan.End(WithSpanOutput("done"))
	observerWaitBatched(t, o)
	if err := o.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(spans) != 2 {
		t.Fatalf("expected exactly 2 exported spans (1 generation, 1 root span), got %d", len(spans))
	}

	attrMap := func(s *api.OtelSpan) map[string]any {
		m := make(map[string]any, len(s.Attributes))
		for _, a := range s.Attributes {
			if a.Key == nil || a.Value == nil {
				continue
			}
			switch {
			case a.Value.StringValue != nil:
				m[*a.Key] = *a.Value.StringValue
			case a.Value.IntValue != nil:
				m[*a.Key] = *a.Value.IntValue
			case a.Value.DoubleValue != nil:
				m[*a.Key] = *a.Value.DoubleValue
			case a.Value.BoolValue != nil:
				m[*a.Key] = *a.Value.BoolValue
			}
		}
		return m
	}

	genAttrs := attrMap(spans[0])
	rootAttrs := attrMap(spans[1])

	// Both root and child generation must carry propagated trace context attributes.
	for i, attrs := range []map[string]any{genAttrs, rootAttrs} {
		if attrs[otelAttrTraceName] != "flow-1" {
			t.Errorf("span[%d] %s = %v, want flow-1", i, otelAttrTraceName, attrs[otelAttrTraceName])
		}
		if attrs[otelAttrUserID] != "alice" {
			t.Errorf("span[%d] %s = %v, want alice", i, otelAttrUserID, attrs[otelAttrUserID])
		}
		if attrs[otelAttrSessionID] != "sess-1" {
			t.Errorf("span[%d] %s = %v, want sess-1", i, otelAttrSessionID, attrs[otelAttrSessionID])
		}
		if attrs[otelAttrRelease] != "1.2.3" {
			t.Errorf("span[%d] %s = %v, want 1.2.3", i, otelAttrRelease, attrs[otelAttrRelease])
		}
		if attrs[otelAttrEnvironment] != "staging" {
			t.Errorf("span[%d] %s = %v, want staging", i, otelAttrEnvironment, attrs[otelAttrEnvironment])
		}
		if attrs[otelAttrTraceMetadataPrefix+"flow_id"] != 1 {
			t.Errorf("span[%d] trace metadata flow_id = %v, want 1", i, attrs[otelAttrTraceMetadataPrefix+"flow_id"])
		}
	}

	// Root span inherits trace input when its own input is unset.
	if rootAttrs[otelAttrObservationInput] != "initial user input" {
		t.Errorf("root span input = %v, want 'initial user input'", rootAttrs[otelAttrObservationInput])
	}
	if genAttrs[otelAttrObservationType] != "generation" {
		t.Errorf("generation observation type = %v, want generation", genAttrs[otelAttrObservationType])
	}
	if genAttrs[otelAttrObservationModelName] != "gpt-4o" {
		t.Errorf("generation model = %v, want gpt-4o", genAttrs[otelAttrObservationModelName])
	}
}

func TestObserver_AllObservationTypesAndScoresV4(t *testing.T) {
	var mu sync.Mutex
	var spans []*api.OtelSpan
	var scores []*api.CreateScoreRequest

	srv := observerSink(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(IngestionVersionHeader); got != IngestionVersion {
			t.Errorf("header %s = %q, want %q", IngestionVersionHeader, got, IngestionVersion)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		switch r.URL.Path {
		case "/api/public/otel/v1/traces":
			var req api.OpentelemetryExportTracesRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("unmarshal otel request: %v", err)
			}
			mu.Lock()
			for _, rs := range req.ResourceSpans {
				for _, ss := range rs.ScopeSpans {
					spans = append(spans, ss.Spans...)
				}
			}
			mu.Unlock()
			_, _ = w.Write([]byte(`{}`))
		case "/api/public/scores":
			var req api.CreateScoreRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("unmarshal score request: %v", err)
			}
			mu.Lock()
			scores = append(scores, &req)
			mu.Unlock()
			_, _ = w.Write([]byte(`{"id":"score-1"}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	client, err := NewClient(WithBaseURL(srv.URL), WithPublicKey("pk"), WithSecretKey("sk"), WithProjectID("proj"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	obs := NewObserver(client,
		WithRelease("1.2.3"),
		WithEnvironment("test-env"),
		WithSendInterval(10*time.Minute),
		WithSendTimeout(2*time.Second),
		WithQueueSize(100),
	).(*observer)
	defer func() { _ = obs.Shutdown(context.Background()) }()

	traceID := newTraceID()
	ctx, rootObs := obs.NewObservation(context.Background(),
		WithObservationTraceID(traceID),
		WithObservationTraceContext(
			WithTraceName("v4-trace"),
			WithTraceUserID("user-1"),
			WithTraceSessionID("session-1"),
			WithTraceTags([]string{"v4", "unit"}),
			WithTraceInput(map[string]any{"task": "verify langfuse v4"}),
			WithTraceMetadata(Metadata{"flow_id": 42, "tenant": "test"}),
			WithTracePublic(),
		),
	)

	rootSpan := rootObs.Span(
		WithSpanName("flow-root-span"),
		WithSpanInput(map[string]any{"task": "verify langfuse v4"}),
		WithSpanMetadata(Metadata{"component": "controller"}),
	)
	ctx, spanObs := rootSpan.Observation(ctx)

	agent := spanObs.Agent(
		WithAgentName("pentester-agent"),
		WithAgentInput("run recon"),
		WithAgentMetadata(Metadata{"agent_type": "pentester"}),
	)
	ctx, agentObs := agent.Observation(ctx)

	temp := 0.2
	maxTok := 1024
	inCost := 0.0015
	outCost := 0.0030
	gen := agentObs.Generation(
		WithGenerationName("openai-generation"),
		WithGenerationModel("gpt-4o-mini"),
		WithGenerationModelParameters(&ModelParameters{
			Temperature: &temp,
			MaxTokens:   &maxTok,
		}),
		WithGenerationInput([]map[string]string{{"role": "user", "content": "ping"}}),
		WithGenerationPromptName("pentester-system"),
		WithGenerationPromptVersion(3),
	)
	gen.End(
		WithGenerationOutput(map[string]string{"role": "assistant", "content": "pong"}),
		WithGenerationUsage(&GenerationUsage{
			Input:      150,
			Output:     50,
			InputCost:  &inCost,
			OutputCost: &outCost,
			Unit:       GenerationUsageUnitTokens,
		}),
		WithGenerationStatus("ok"),
	)

	tool := agentObs.Tool(WithToolName("terminal-tool"), WithToolInput(map[string]string{"cmd": "whoami"}))
	tool.End(WithToolOutput("root"), WithToolStatus("completed"))

	emb := agentObs.Embedding(WithEmbeddingName("memory-embedding"), WithEmbeddingModel("text-embedding-3-small"), WithEmbeddingInput("recon summary"))
	emb.End(WithEmbeddingOutput([]float64{0.1, 0.2, 0.3}), WithEmbeddingStatus("embedded"))

	agentObs.Event(WithEventName("checkpoint-event"), WithEventInput("checkpoint reached"), WithEventOutput("saved"), WithEventMetadata(Metadata{"step": 1}))

	chain := agentObs.Chain(WithChainName("summarizer-chain"), WithChainInput("long conversation"))
	chain.End(WithChainOutput("short summary"), WithChainStatus("summarized"))

	retriever := agentObs.Retriever(WithRetrieverName("vector-retriever"), WithRetrieverInput("query vector"))
	retriever.End(WithRetrieverOutput([]string{"doc-1", "doc-2"}), WithRetrieverStatus("retrieved"))

	evaluator := agentObs.Evaluator(WithEvaluatorName("safety-evaluator"), WithEvaluatorInput("candidate output"))
	evaluator.End(WithEvaluatorOutput(map[string]any{"passed": true}), WithEvaluatorStatus("evaluated"))

	guardrail := agentObs.Guardrail(WithGuardrailName("scope-guardrail"), WithGuardrailInput("target ip"))
	guardrail.End(WithGuardrailOutput("allowed"), WithGuardrailLevel(ObservationLevelWarning), WithGuardrailStatus("warned"))

	agentObs.Score(WithScoreName("quality"), WithScoreFloatValue(0.95), WithScoreComment("score verification"))
	agentObs.Score(WithScoreName("verdict"), WithScoreStringValue("pass"), WithScoreComment("categorical score verification"))

	agent.End(WithAgentOutput("recon complete"), WithAgentStatus("done"))
	rootSpan.End(WithSpanOutput(map[string]any{"result": "success"}), WithSpanStatus("completed"))

	observerWaitBatched(t, obs)
	if err := obs.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(spans) != 10 {
		t.Fatalf("expected 10 exported OTLP spans, got %d", len(spans))
	}
	if len(scores) != 2 {
		t.Fatalf("expected 2 exported scores, got %d", len(scores))
	}

	seenTypes := make(map[string]bool)
	for _, s := range spans {
		m := make(map[string]any, len(s.Attributes))
		for _, a := range s.Attributes {
			if a.Key != nil && a.Value != nil && a.Value.StringValue != nil {
				m[*a.Key] = *a.Value.StringValue
			}
		}
		obsType, _ := m[otelAttrObservationType].(string)
		seenTypes[obsType] = true
		if m[otelAttrUserID] != "user-1" {
			t.Errorf("span %s (%s) %s = %v, want user-1", s.SpanID, obsType, otelAttrUserID, m[otelAttrUserID])
		}
		if m[otelAttrSessionID] != "session-1" {
			t.Errorf("span %s (%s) %s = %v, want session-1", s.SpanID, obsType, otelAttrSessionID, m[otelAttrSessionID])
		}
		if m[otelAttrEnvironment] != "test-env" {
			t.Errorf("span %s (%s) %s = %v, want test-env", s.SpanID, obsType, otelAttrEnvironment, m[otelAttrEnvironment])
		}
	}

	for _, wantType := range []string{"span", "agent", "generation", "tool", "embedding", "event", "chain", "retriever", "evaluator", "guardrail"} {
		if !seenTypes[wantType] {
			t.Errorf("missing observation type %q in exported spans", wantType)
		}
	}
}
