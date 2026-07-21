package observability

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/version"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultLogInterval    = time.Second * 30
	DefaultLogTimeout     = time.Second * 10
	DefaultMetricInterval = time.Second * 30
	DefaultMetricTimeout  = time.Second * 10
	DefaultTraceInterval  = time.Second * 30
	DefaultTraceTimeout   = time.Second * 10
)

type TelemetryClient interface {
	Logger() otellog.LoggerProvider
	Tracer() oteltrace.TracerProvider
	Meter() otelmetric.MeterProvider
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}

type telemetryClient struct {
	conn   *grpc.ClientConn
	logger *sdklog.LoggerProvider
	tracer *sdktrace.TracerProvider
	meter  *sdkmetric.MeterProvider
}

func (c *telemetryClient) Logger() otellog.LoggerProvider {
	return c.logger
}

func (c *telemetryClient) Tracer() oteltrace.TracerProvider {
	return c.tracer
}

func (c *telemetryClient) Meter() otelmetric.MeterProvider {
	return c.meter
}

func (c *telemetryClient) Shutdown(ctx context.Context) error {
	var errs []error
	if err := c.logger.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown logger: %w", err))
	}
	if err := c.meter.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown meter: %w", err))
	}
	if err := c.tracer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown tracer: %w", err))
	}
	// Always close the connection, even if a provider shutdown failed above, so a
	// stalled flush can't leak the grpc conn.
	if err := c.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close telemetry connection: %w", err))
	}
	return errors.Join(errs...)
}

func (c *telemetryClient) ForceFlush(ctx context.Context) error {
	var errs []error
	if err := c.logger.ForceFlush(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to force flush logger: %w", err))
	}
	if err := c.meter.ForceFlush(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to force flush meter: %w", err))
	}
	if err := c.tracer.ForceFlush(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to force flush tracer: %w", err))
	}
	return errors.Join(errs...)
}

func NewTelemetryClient(ctx context.Context, cfg *config.Config) (TelemetryClient, error) {
	if cfg.TelemetryEndpoint == "" {
		return nil, fmt.Errorf("telemetry endpoint is not set: %w", ErrNotConfigured)
	}

	// grpc.NewClient is non-blocking: it never dials during startup, so a
	// set-but-unreachable collector can't stall main(), and the connection is
	// established (and re-established) lazily in the background — a collector that
	// comes up after the app does connects on its own, without a restart.
	conn, err := grpc.NewClient(
		cfg.TelemetryEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry connection: %w", err)
	}

	// Build all three exporters before creating any provider. Exporter creation
	// is the only remaining error path, and at that point no batch/reader
	// goroutine has started yet, so closing the connection is a complete teardown.
	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}
	spanExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create tracer exporter: %w", err)
	}

	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(
			logExporter,
			sdklog.WithExportInterval(DefaultLogInterval),
			sdklog.WithExportTimeout(DefaultLogTimeout),
		)),
		sdklog.WithResource(newResource()),
	)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			metricExporter,
			sdkmetric.WithInterval(DefaultMetricInterval),
			sdkmetric.WithTimeout(DefaultMetricTimeout),
		)),
		sdkmetric.WithResource(newResource()),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(
			spanExporter,
			sdktrace.WithBatchTimeout(DefaultTraceInterval),
			sdktrace.WithExportTimeout(DefaultTraceTimeout),
		)),
		sdktrace.WithResource(newResource()),
	)

	return &telemetryClient{
		conn:   conn,
		logger: logProvider,
		meter:  meterProvider,
		tracer: tracerProvider,
	}, nil
}

func newResource(opts ...attribute.KeyValue) *resource.Resource {
	var env = "production"
	if version.IsDevelopMode() {
		env = "development"
	}

	service := version.GetBinaryName()
	verRev := strings.Split(version.GetBinaryVersion(), "-")
	version := strings.TrimPrefix(verRev[0], "v")

	opts = append(opts,
		semconv.ServiceName(service),
		semconv.ServiceVersion(version),
		attribute.String("environment", env),
	)

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		opts...,
	)
}
