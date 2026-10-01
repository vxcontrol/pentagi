package observability

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	otelloggernoop "go.opentelemetry.io/otel/log/noop"
	otelmetric "go.opentelemetry.io/otel/metric"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"
	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type fakeLangfuseClient struct {
	shutdownCalled bool
	flushCalled    bool
	shutdownErr    error
	flushErr       error
}

func (f *fakeLangfuseClient) API() langfuse.Client        { return langfuse.Client{} }
func (f *fakeLangfuseClient) Observer() langfuse.Observer { return langfuse.NewNoopObserver() }
func (f *fakeLangfuseClient) Shutdown(context.Context) error {
	f.shutdownCalled = true
	return f.shutdownErr
}
func (f *fakeLangfuseClient) ForceFlush(context.Context) error {
	f.flushCalled = true
	return f.flushErr
}

type fakeTelemetryClient struct {
	shutdownCalled bool
	flushCalled    bool
	shutdownErr    error
	flushErr       error
}

func (f *fakeTelemetryClient) Logger() otellog.LoggerProvider {
	return otelloggernoop.NewLoggerProvider()
}
func (f *fakeTelemetryClient) Tracer() oteltrace.TracerProvider {
	return oteltracenoop.NewTracerProvider()
}
func (f *fakeTelemetryClient) Meter() otelmetric.MeterProvider {
	return otelmetricnoop.NewMeterProvider()
}
func (f *fakeTelemetryClient) Shutdown(context.Context) error {
	f.shutdownCalled = true
	return f.shutdownErr
}
func (f *fakeTelemetryClient) ForceFlush(context.Context) error {
	f.flushCalled = true
	return f.flushErr
}

func TestObserverShutdownDrainsBothClients(t *testing.T) {
	lf := &fakeLangfuseClient{}
	ot := &fakeTelemetryClient{}
	obs := &observer{lfclient: lf, otelclient: ot}

	if err := obs.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if !lf.shutdownCalled {
		t.Error("langfuse client was not shut down")
	}
	if !ot.shutdownCalled {
		t.Error("otel client was not shut down")
	}
}

func TestObserverFlushDrainsBothClients(t *testing.T) {
	lf := &fakeLangfuseClient{}
	ot := &fakeTelemetryClient{}
	obs := &observer{lfclient: lf, otelclient: ot}

	if err := obs.Flush(context.Background()); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if !lf.flushCalled {
		t.Error("langfuse client was not flushed")
	}
	if !ot.flushCalled {
		t.Error("otel client was not flushed")
	}
}

func TestObserverShutdownAggregatesErrorsAndDrainsBoth(t *testing.T) {
	lfErr := errors.New("langfuse boom")
	otErr := errors.New("otel boom")
	lf := &fakeLangfuseClient{shutdownErr: lfErr}
	ot := &fakeTelemetryClient{shutdownErr: otErr}
	obs := &observer{lfclient: lf, otelclient: ot}

	err := obs.Shutdown(context.Background())
	if !errors.Is(err, lfErr) || !errors.Is(err, otErr) {
		t.Fatalf("want both errors joined, got %v", err)
	}
	if !lf.shutdownCalled || !ot.shutdownCalled {
		t.Fatal("a client was skipped despite the other erroring")
	}
}

func TestObserverShutdownFlushNilClientsAreNoops(t *testing.T) {
	obs := &observer{}
	if err := obs.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown with no clients: %v", err)
	}
	if err := obs.Flush(context.Background()); err != nil {
		t.Errorf("Flush with no clients: %v", err)
	}
}

type recordingTraceCollector struct {
	coltracev1.UnimplementedTraceServiceServer
	spanBatches int32
}

func (r *recordingTraceCollector) Export(context.Context, *coltracev1.ExportTraceServiceRequest) (*coltracev1.ExportTraceServiceResponse, error) {
	atomic.AddInt32(&r.spanBatches, 1)
	return &coltracev1.ExportTraceServiceResponse{}, nil
}

// blockingLangfuse models the real langfuse leg: ForceFlush/Shutdown ignore the
// passed ctx and block on the observer's own machinery, so a pending batch against
// a down endpoint blocks past the drain budget.
type blockingLangfuse struct {
	release    chan struct{}
	flushCalls int32
}

func (b *blockingLangfuse) API() langfuse.Client        { return langfuse.Client{} }
func (b *blockingLangfuse) Observer() langfuse.Observer { return langfuse.NewNoopObserver() }
func (b *blockingLangfuse) ForceFlush(context.Context) error {
	atomic.AddInt32(&b.flushCalls, 1)
	<-b.release
	return nil
}
func (b *blockingLangfuse) Shutdown(context.Context) error {
	<-b.release
	return nil
}

func newRecordingOtel(t *testing.T) (*telemetryClient, *recordingTraceCollector) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	rec := &recordingTraceCollector{}
	srv := grpc.NewServer()
	collogsv1.RegisterLogsServiceServer(srv, fakeLogsCollector{})
	colmetricsv1.RegisterMetricsServiceServer(srv, fakeMetricsCollector{})
	coltracev1.RegisterTraceServiceServer(srv, rec)
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(srv.Stop)

	tc, err := NewTelemetryClient(context.Background(), &config.Config{TelemetryEndpoint: ln.Addr().String()})
	if err != nil {
		t.Fatalf("NewTelemetryClient: %v", err)
	}
	return tc.(*telemetryClient), rec
}

func bufferOneSpan(o *observer) {
	_, span := o.otelclient.Tracer().Tracer("drain-test").Start(context.Background(), "s")
	span.End()
}

// A healthy otel sink is delivered even while an unhealthy langfuse sink blocks.
func TestObserverDrain_DeliversOtelDespiteBlockedLangfuse(t *testing.T) {
	otel, rec := newRecordingOtel(t)
	lf := &blockingLangfuse{release: make(chan struct{})}
	t.Cleanup(func() { close(lf.release) })

	obs := &observer{lfclient: lf, otelclient: otel}
	bufferOneSpan(obs)

	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()
	_ = obs.Drain(ctx)

	if atomic.LoadInt32(&lf.flushCalls) == 0 {
		t.Fatal("precondition: langfuse ForceFlush should have been attempted")
	}
	if got := atomic.LoadInt32(&rec.spanBatches); got != 1 {
		t.Fatalf("otel span must be delivered despite a blocked langfuse; got %d export batches", got)
	}
}

func TestObserverDrain_ReturnsWithinDeadlineWhenSinkBlocks(t *testing.T) {
	otel, _ := newRecordingOtel(t)
	lf := &blockingLangfuse{release: make(chan struct{})}
	t.Cleanup(func() { close(lf.release) })

	obs := &observer{lfclient: lf, otelclient: otel}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := obs.Drain(ctx)

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Drain hung past its deadline: %v", elapsed)
	}
	if err == nil {
		t.Fatal("Drain should report the deadline was hit while a sink was blocked")
	}
}

type recordingLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (r *recordingLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, record := range records {
		r.records = append(r.records, record.Clone())
	}
	return nil
}
func (r *recordingLogExporter) Shutdown(context.Context) error   { return nil }
func (r *recordingLogExporter) ForceFlush(context.Context) error { return nil }

func TestObs_Fire_ExportsAFieldToTheLogRecordAndTheSpanEvent(t *testing.T) {
	type namedString string

	for name, tc := range map[string]struct {
		value any
		log   attribute.Value
		span  attribute.Value
	}{
		"a string": {"s", attribute.StringValue("s"), attribute.StringValue("s")},
		"an int slice": {[]int{1, 2},
			attribute.SliceValue(attribute.Int64Value(1), attribute.Int64Value(2)), attribute.IntSliceValue([]int{1, 2})},
		"an int array": {[2]int{1, 2},
			attribute.SliceValue(attribute.Int64Value(1), attribute.Int64Value(2)), attribute.IntSliceValue([]int{1, 2})},
		"a string array": {[2]string{"a", "b"},
			attribute.SliceValue(attribute.StringValue("a"), attribute.StringValue("b")), attribute.StringSliceValue([]string{"a", "b"})},
		"a slice of a named string type": {[]namedString{"a"},
			attribute.SliceValue(attribute.StringValue("a")), attribute.StringSliceValue([]string{"a"})},
	} {
		t.Run(name, func(t *testing.T) {
			logs := &recordingLogExporter{}
			spans := tracetest.NewSpanRecorder()
			obs := &observer{
				levels: logrus.AllLevels,
				logger: sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(logs))).Logger("test"),
				tracer: sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)).Tracer("test"),
			}
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			logger.AddHook(obs)

			logger.WithField("value", tc.value).Info("probe")

			if len(logs.records) != 1 {
				t.Fatalf("exported %d log records, want 1", len(logs.records))
			}
			var logged []attribute.Value
			logs.records[0].WalkAttributes(func(kv attribute.KeyValue) bool {
				if kv.Key == "value" {
					logged = append(logged, kv.Value)
				}
				return true
			})
			if len(logged) != 1 || !sameValue(logged[0], tc.log) {
				t.Errorf("log record attribute = %v, want %v", jsonValues(logged), jsonValues([]attribute.Value{tc.log}))
			}

			ended := spans.Ended()
			if len(ended) != 1 || len(ended[0].Events()) != 1 {
				t.Fatalf("want one ended span with one event, got %d spans", len(ended))
			}
			var evented []attribute.Value
			for _, kv := range ended[0].Events()[0].Attributes {
				if kv.Key == "log.value" {
					evented = append(evented, kv.Value)
				}
			}
			if len(evented) != 1 || !sameValue(evented[0], tc.span) {
				t.Errorf("span event attribute = %v, want %v", jsonValues(evented), jsonValues([]attribute.Value{tc.span}))
			}
		})
	}
}

func sameValue(a, b attribute.Value) bool {
	return jsonValues([]attribute.Value{a}) == jsonValues([]attribute.Value{b})
}

func jsonValues(values []attribute.Value) string {
	var out string
	for _, v := range values {
		b, _ := v.MarshalJSON()
		out += string(b)
	}
	return out
}
