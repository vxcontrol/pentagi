package observability

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"pentagi/pkg/config"

	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type fakeLogsCollector struct {
	collogsv1.UnimplementedLogsServiceServer
}

func (fakeLogsCollector) Export(context.Context, *collogsv1.ExportLogsServiceRequest) (*collogsv1.ExportLogsServiceResponse, error) {
	return &collogsv1.ExportLogsServiceResponse{}, nil
}

type fakeMetricsCollector struct {
	colmetricsv1.UnimplementedMetricsServiceServer
}

func (fakeMetricsCollector) Export(context.Context, *colmetricsv1.ExportMetricsServiceRequest) (*colmetricsv1.ExportMetricsServiceResponse, error) {
	return &colmetricsv1.ExportMetricsServiceResponse{}, nil
}

type fakeTraceCollector struct {
	coltracev1.UnimplementedTraceServiceServer
}

func (fakeTraceCollector) Export(context.Context, *coltracev1.ExportTraceServiceRequest) (*coltracev1.ExportTraceServiceResponse, error) {
	return &coltracev1.ExportTraceServiceResponse{}, nil
}

func TestNewTelemetryClient_EmptyEndpointReturnsErrNotConfigured(t *testing.T) {
	_, err := NewTelemetryClient(context.Background(), &config.Config{TelemetryEndpoint: ""})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestNewTelemetryClient_SuccessPathExportsAndShutsDown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	collogsv1.RegisterLogsServiceServer(srv, fakeLogsCollector{})
	colmetricsv1.RegisterMetricsServiceServer(srv, fakeMetricsCollector{})
	coltracev1.RegisterTraceServiceServer(srv, fakeTraceCollector{})
	go func() { _ = srv.Serve(ln) }()
	defer srv.Stop()

	client, err := NewTelemetryClient(context.Background(), &config.Config{TelemetryEndpoint: ln.Addr().String()})
	if err != nil {
		t.Fatalf("NewTelemetryClient: %v", err)
	}
	if client.Logger() == nil || client.Tracer() == nil || client.Meter() == nil {
		t.Fatal("a telemetry provider is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.ForceFlush(ctx); err != nil {
		t.Fatalf("ForceFlush: %v", err)
	}
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

// A set-but-unreachable collector must not stall startup: grpc.NewClient is
// non-blocking, so the client is returned immediately and connects in the
// background if the collector later comes up.
func TestNewTelemetryClient_UnreachableDoesNotBlockStartup(t *testing.T) {
	type result struct {
		client TelemetryClient
		err    error
	}
	cfg := &config.Config{TelemetryEndpoint: "127.0.0.1:1"} // nothing listening
	done := make(chan result, 1)
	go func() {
		c, e := NewTelemetryClient(context.Background(), cfg)
		done <- result{c, e}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("non-blocking client must not error on an unreachable collector: %v", res.err)
		}
		if res.client == nil {
			t.Fatal("expected a non-nil client")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = res.client.Shutdown(shutdownCtx)
	case <-time.After(3 * time.Second):
		t.Fatal("NewTelemetryClient blocked on an unreachable collector")
	}
}
