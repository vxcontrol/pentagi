package observability

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/version"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	otelloggernoop "go.opentelemetry.io/otel/log/noop"
	otelmetric "go.opentelemetry.io/otel/metric"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"
)

type SpanContextKey int

var (
	logSeverityKey = attribute.Key("log.severity")
	logMessageKey  = attribute.Key("log.message")
	spanComponent  = attribute.Key("span.component")
)

var ErrNotConfigured = errors.New("not configured")

const InstrumentationVersion = "1.0.0"

const (
	maximumCallerDepth int    = 25
	logrusPackageName  string = "github.com/sirupsen/logrus"
)

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified oteltrace.SpanKind = 0
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal oteltrace.SpanKind = 1
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer oteltrace.SpanKind = 2
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient oteltrace.SpanKind = 3
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker. Unlike
	// SpanKindClient and SpanKindServer, there is often no direct
	// relationship between this kind of Span and a SpanKindConsumer kind. A
	// SpanKindProducer Span will end once the message is accepted by the
	// message broker which might not overlap with the processing of that
	// message.
	SpanKindProducer oteltrace.SpanKind = 4
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker. Like
	// SpanKindProducer Spans, there is often no direct relationship between
	// this Span and the Span that produced the message.
	SpanKindConsumer oteltrace.SpanKind = 5
)

var Observer Observability

type Observability interface {
	Flush(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Meter
	Tracer
	Collector
	Langfuse
}

type Langfuse interface {
	NewObservation(context.Context, ...langfuse.ObservationContextOption) (context.Context, langfuse.Observation)
}

type Tracer interface {
	NewSpan(context.Context, oteltrace.SpanKind, string, ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span)
	NewSpanWithParent(
		context.Context,
		oteltrace.SpanKind,
		string,
		string,
		string,
		...oteltrace.SpanStartOption,
	) (context.Context, oteltrace.Span)
	SpanFromContext(ctx context.Context) oteltrace.Span
	SpanContextFromContext(ctx context.Context) oteltrace.SpanContext
}

type Meter interface {
	NewInt64Counter(string, ...otelmetric.Int64CounterOption) (otelmetric.Int64Counter, error)
	NewInt64UpDownCounter(string, ...otelmetric.Int64UpDownCounterOption) (otelmetric.Int64UpDownCounter, error)
	NewInt64Histogram(string, ...otelmetric.Int64HistogramOption) (otelmetric.Int64Histogram, error)
	NewInt64Gauge(string, ...otelmetric.Int64GaugeOption) (otelmetric.Int64Gauge, error)
	NewInt64ObservableCounter(string, ...otelmetric.Int64ObservableCounterOption) (otelmetric.Int64ObservableCounter, error)
	NewInt64ObservableUpDownCounter(string, ...otelmetric.Int64ObservableUpDownCounterOption) (otelmetric.Int64ObservableUpDownCounter, error)
	NewInt64ObservableGauge(string, ...otelmetric.Int64ObservableGaugeOption) (otelmetric.Int64ObservableGauge, error)
	NewFloat64Counter(string, ...otelmetric.Float64CounterOption) (otelmetric.Float64Counter, error)
	NewFloat64UpDownCounter(string, ...otelmetric.Float64UpDownCounterOption) (otelmetric.Float64UpDownCounter, error)
	NewFloat64Histogram(string, ...otelmetric.Float64HistogramOption) (otelmetric.Float64Histogram, error)
	NewFloat64Gauge(string, ...otelmetric.Float64GaugeOption) (otelmetric.Float64Gauge, error)
	NewFloat64ObservableCounter(string, ...otelmetric.Float64ObservableCounterOption) (otelmetric.Float64ObservableCounter, error)
	NewFloat64ObservableUpDownCounter(string, ...otelmetric.Float64ObservableUpDownCounterOption) (otelmetric.Float64ObservableUpDownCounter, error)
	NewFloat64ObservableGauge(string, ...otelmetric.Float64ObservableGaugeOption) (otelmetric.Float64ObservableGauge, error)
}

type Collector interface {
	StartProcessMetricCollect(attrs ...attribute.KeyValue) error
	StartGoRuntimeMetricCollect(attrs ...attribute.KeyValue) error
	StartDumperMetricCollect(stats Dumper, attrs ...attribute.KeyValue) error
}

type Dumper interface {
	DumpStats() (map[string]float64, error)
}

type observer struct {
	levels     []logrus.Level
	logger     otellog.Logger
	tracer     oteltrace.Tracer
	meter      otelmetric.Meter
	lfclient   LangfuseClient
	otelclient TelemetryClient
	observer   langfuse.Observer
}

func init() {
	InitObserver(context.Background(), nil, nil, []logrus.Level{})
}

func InitObserver(ctx context.Context, lfclient LangfuseClient, otelclient TelemetryClient, levels []logrus.Level) {
	if Observer != nil {
		Observer.Flush(ctx)
	}

	obs := &observer{
		levels:     levels,
		lfclient:   lfclient,
		otelclient: otelclient,
	}

	tname := version.GetBinaryName()
	tversion := InstrumentationVersion

	if lfclient != nil {
		obs.observer = lfclient.Observer()
	} else {
		obs.observer = langfuse.NewNoopObserver()
	}

	if otelclient != nil {
		provider := otelclient.Logger()
		global.SetLoggerProvider(provider)
		obs.logger = provider.Logger(tname, otellog.WithInstrumentationVersion(tversion))
	} else {
		obs.logger = otelloggernoop.NewLoggerProvider().Logger(tname)
	}

	if otelclient != nil {
		provider := otelclient.Tracer()
		otel.SetTracerProvider(provider)
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
				// TODO: add langfuse propagator
			),
		)
		obs.tracer = provider.Tracer(tname, oteltrace.WithInstrumentationVersion(tversion))
		logrus.AddHook(obs)
	} else {
		obs.tracer = oteltracenoop.NewTracerProvider().Tracer(tname)
	}

	if otelclient != nil {
		provider := otelclient.Meter()
		otel.SetMeterProvider(provider)
		obs.meter = provider.Meter(tname, otelmetric.WithInstrumentationVersion(tversion))
	} else {
		obs.meter = otelmetricnoop.NewMeterProvider().Meter(tname)
	}

	Observer = obs
}

func (obs *observer) Flush(ctx context.Context) error {
	if obs.lfclient != nil {
		return obs.lfclient.ForceFlush(ctx)
	}

	if obs.otelclient != nil {
		return obs.otelclient.ForceFlush(ctx)
	}

	return nil
}

func (obs *observer) Shutdown(ctx context.Context) error {
	if obs.lfclient != nil {
		return obs.lfclient.Shutdown(ctx)
	}

	if obs.otelclient != nil {
		return obs.otelclient.Shutdown(ctx)
	}

	return nil
}

func (obs *observer) StartProcessMetricCollect(attrs ...attribute.KeyValue) error {
	if obs.meter == nil {
		return nil
	}

	attrs = append(attrs,
		semconv.ServiceNameKey.String(version.GetBinaryName()),
		semconv.ServiceVersionKey.String(version.GetBinaryVersion()),
	)
	return startProcessMetricCollect(obs.meter, attrs)
}

func (obs *observer) StartGoRuntimeMetricCollect(attrs ...attribute.KeyValue) error {
	if obs.meter == nil {
		return nil
	}

	attrs = append(attrs,
		semconv.ServiceNameKey.String(version.GetBinaryName()),
		semconv.ServiceVersionKey.String(version.GetBinaryVersion()),
	)
	return startGoRuntimeMetricCollect(obs.meter, attrs)
}

func (obs *observer) StartDumperMetricCollect(stats Dumper, attrs ...attribute.KeyValue) error {
	if obs.meter == nil {
		return nil
	}

	attrs = append(attrs,
		semconv.ServiceNameKey.String(version.GetBinaryName()),
		semconv.ServiceVersionKey.String(version.GetBinaryVersion()),
	)
	return startDumperMetricCollect(stats, obs.meter, attrs)
}

func (obs *observer) NewInt64Counter(
	name string, options ...otelmetric.Int64CounterOption,
) (otelmetric.Int64Counter, error) {
	return obs.meter.Int64Counter(name, options...)
}

func (obs *observer) NewInt64UpDownCounter(
	name string, options ...otelmetric.Int64UpDownCounterOption,
) (otelmetric.Int64UpDownCounter, error) {
	return obs.meter.Int64UpDownCounter(name, options...)
}

func (obs *observer) NewInt64Histogram(
	name string, options ...otelmetric.Int64HistogramOption,
) (otelmetric.Int64Histogram, error) {
	return obs.meter.Int64Histogram(name, options...)
}

func (obs *observer) NewInt64Gauge(
	name string, options ...otelmetric.Int64GaugeOption,
) (otelmetric.Int64Gauge, error) {
	return obs.meter.Int64Gauge(name, options...)
}

func (obs *observer) NewInt64ObservableCounter(
	name string, options ...otelmetric.Int64ObservableCounterOption,
) (otelmetric.Int64ObservableCounter, error) {
	return obs.meter.Int64ObservableCounter(name, options...)
}

func (obs *observer) NewInt64ObservableUpDownCounter(
	name string, options ...otelmetric.Int64ObservableUpDownCounterOption,
) (otelmetric.Int64ObservableUpDownCounter, error) {
	return obs.meter.Int64ObservableUpDownCounter(name, options...)
}

func (obs *observer) NewInt64ObservableGauge(
	name string, options ...otelmetric.Int64ObservableGaugeOption,
) (otelmetric.Int64ObservableGauge, error) {
	return obs.meter.Int64ObservableGauge(name, options...)
}

func (obs *observer) NewFloat64Counter(
	name string, options ...otelmetric.Float64CounterOption,
) (otelmetric.Float64Counter, error) {
	return obs.meter.Float64Counter(name, options...)
}

func (obs *observer) NewFloat64UpDownCounter(
	name string, options ...otelmetric.Float64UpDownCounterOption,
) (otelmetric.Float64UpDownCounter, error) {
	return obs.meter.Float64UpDownCounter(name, options...)
}

func (obs *observer) NewFloat64Histogram(
	name string, options ...otelmetric.Float64HistogramOption,
) (otelmetric.Float64Histogram, error) {
	return obs.meter.Float64Histogram(name, options...)
}

func (obs *observer) NewFloat64Gauge(
	name string, options ...otelmetric.Float64GaugeOption,
) (otelmetric.Float64Gauge, error) {
	return obs.meter.Float64Gauge(name, options...)
}

func (obs *observer) NewFloat64ObservableCounter(
	name string, options ...otelmetric.Float64ObservableCounterOption,
) (otelmetric.Float64ObservableCounter, error) {
	return obs.meter.Float64ObservableCounter(name, options...)
}

func (obs *observer) NewFloat64ObservableUpDownCounter(
	name string, options ...otelmetric.Float64ObservableUpDownCounterOption,
) (otelmetric.Float64ObservableUpDownCounter, error) {
	return obs.meter.Float64ObservableUpDownCounter(name, options...)
}

func (obs *observer) NewFloat64ObservableGauge(
	name string, options ...otelmetric.Float64ObservableGaugeOption,
) (otelmetric.Float64ObservableGauge, error) {
	return obs.meter.Float64ObservableGauge(name, options...)
}

func (obs *observer) NewObservation(
	ctx context.Context, options ...langfuse.ObservationContextOption,
) (context.Context, langfuse.Observation) {
	return obs.observer.NewObservation(ctx, options...)
}

func (obs *observer) NewSpan(ctx context.Context, kind oteltrace.SpanKind,
	component string, opts ...oteltrace.SpanStartOption,
) (context.Context, oteltrace.Span) {
	if ctx == nil {
		// TODO: here should use default context
		ctx = context.TODO()
	}

	opts = append(opts,
		oteltrace.WithSpanKind(kind),
		oteltrace.WithAttributes(spanComponent.String(component)),
	)

	return obs.tracer.Start(ctx, component, opts...)
}

func (obs *observer) NewSpanWithParent(ctx context.Context, kind oteltrace.SpanKind,
	component, traceID, pspanID string, opts ...oteltrace.SpanStartOption,
) (context.Context, oteltrace.Span) {
	if ctx == nil {
		// TODO: here should use default context
		ctx = context.TODO()
	}

	var (
		err error
		tid oteltrace.TraceID
		sid oteltrace.SpanID
	)
	tid, err = oteltrace.TraceIDFromHex(traceID)
	if err != nil {
		_, _ = rand.Read(tid[:])
	}
	sid, err = oteltrace.SpanIDFromHex(pspanID)
	if err != nil {
		sid = oteltrace.SpanID{}
	}
	ctx = oteltrace.ContextWithRemoteSpanContext(ctx,
		oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
			TraceID: tid,
			SpanID:  sid,
		}),
	)

	return obs.tracer.Start(
		ctx,
		component,
		oteltrace.WithSpanKind(kind),
		oteltrace.WithAttributes(spanComponent.String(component)),
	)
}

func (obs *observer) SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func (obs *observer) SpanContextFromContext(ctx context.Context) oteltrace.SpanContext {
	return oteltrace.SpanContextFromContext(ctx)
}

func (obs *observer) makeAttrs(entry *logrus.Entry, span oteltrace.Span) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(entry.Data)+2+3)

	attrs = append(attrs, logSeverityKey.String(levelString(entry.Level)))
	attrs = append(attrs, logMessageKey.String(entry.Message))

	if entry.Caller != nil {
		if entry.Caller.Function != "" {
			attrs = append(attrs, semconv.CodeFunctionKey.String(entry.Caller.Function))
		}
		if entry.Caller.File != "" {
			attrs = append(attrs, semconv.CodeFilepathKey.String(entry.Caller.File))
			attrs = append(attrs, semconv.CodeLineNumberKey.Int(entry.Caller.Line))
		}
	}

	opts := []oteltrace.EventOption{}
	if entry.Level <= logrus.ErrorLevel {
		span.SetStatus(codes.Error, entry.Message)
		stack := strings.Join(getStackTrace(), "\n")
		opts = append(opts, oteltrace.WithAttributes(semconv.ExceptionStacktraceKey.String(stack)))
	}

	for k, v := range entry.Data {
		if k == "error" {
			switch val := v.(type) {
			case error:
				span.RecordError(val, opts...)
			case fmt.Stringer:
				attrs = append(attrs, semconv.ExceptionTypeKey.String(reflect.TypeOf(val).String()))
				attrs = append(attrs, semconv.ExceptionMessageKey.String(val.String()))
				span.RecordError(errors.New(val.String()), opts...)
			case nil:
				span.RecordError(fmt.Errorf("unknown or empty error type: nil"), opts...)
			default:
				attrs = append(attrs, semconv.ExceptionTypeKey.String(reflect.TypeOf(val).String()))
				span.RecordError(fmt.Errorf("unknown exception: %v", v), opts...)
			}
			continue
		}

		attrs = append(attrs, attributeKeyValue("log."+k, v))
	}

	return attrs
}

func (obs *observer) makeRecord(entry *logrus.Entry, span oteltrace.Span) otellog.Record {
	var record otellog.Record

	record.SetBody(logValue(entry.Message))
	record.SetObservedTimestamp(time.Now().UTC())
	record.SetTimestamp(entry.Time)
	record.SetSeverity(logSeverity(entry.Level))
	record.SetSeverityText(levelString(entry.Level))

	spanCtx := span.SpanContext()
	attrs := make([]otellog.KeyValue, 0, len(entry.Data)+5)
	attrs = append(attrs, otellog.String("trace.id", spanCtx.TraceID().String()))
	attrs = append(attrs, otellog.String("span.id", spanCtx.SpanID().String()))

	if entry.Caller != nil {
		if entry.Caller.Function != "" {
			attrs = append(attrs, otellog.String(string(semconv.CodeFunctionKey), entry.Caller.Function))
		}
		if entry.Caller.File != "" {
			attrs = append(attrs, otellog.String(string(semconv.CodeFilepathKey), entry.Caller.File))
			attrs = append(attrs, otellog.Int64(string(semconv.CodeLineNumberKey), int64(entry.Caller.Line)))
		}
	}

	for k, v := range entry.Data {
		if k == "error" {
			attrs = append(attrs, otellog.KeyValue{
				Key:   string(semconv.ExceptionStacktraceKey),
				Value: logValue(getStackTrace()),
			})
			switch val := v.(type) {
			case error:
				attrs = append(attrs, otellog.String(string(semconv.ExceptionTypeKey), reflect.TypeOf(val).String()))
				attrs = append(attrs, otellog.String(string(semconv.ExceptionMessageKey), val.Error()))
			case fmt.Stringer:
				attrs = append(attrs, otellog.String(string(semconv.ExceptionTypeKey), reflect.TypeOf(val).String()))
				attrs = append(attrs, otellog.String(string(semconv.ExceptionMessageKey), val.String()))
			case nil:
				attrs = append(attrs, otellog.String(string(semconv.ExceptionTypeKey), "empty error type: nil"))
			default:
				attrs = append(attrs, otellog.String(string(semconv.ExceptionTypeKey), reflect.TypeOf(val).String()))
				if errorData, err := json.Marshal(val); err == nil {
					attrs = append(attrs, otellog.String(string(semconv.ExceptionMessageKey), string(errorData)))
				}
			}
			continue
		}

		attrs = append(attrs, otellog.KeyValue{Key: k, Value: logValue(v)})
	}

	record.AddAttributes(attrs...)

	return record
}

// Fire is a logrus hook that is fired on a new log entry.
func (obs *observer) Fire(entry *logrus.Entry) error {
	if obs == nil {
		return nil
	}

	ctx := entry.Context
	if ctx == nil {
		ctx = context.Background()
	}

	span := oteltrace.SpanFromContext(ctx)
	if !span.IsRecording() {
		component := "internal"
		if op, ok := entry.Data["component"]; ok {
			component = op.(string)
		}
		if obs.tracer == nil {
			return nil
		}
		// case when span was closing by timeout, we need to create a new span with parent info
		traceID := span.SpanContext().TraceID()
		spanID := span.SpanContext().SpanID()
		_, span = obs.NewSpanWithParent(
			ctx,
			oteltrace.SpanKindInternal,
			component,
			traceID.String(),
			spanID.String(),
			oteltrace.WithLinks(oteltrace.Link{
				SpanContext: span.SpanContext(),
				Attributes: []attribute.KeyValue{
					attribute.String("relationship", "inheriting"),
					attribute.String("reason", "span was closed by timeout"),
				},
			}),
		)
		defer span.End()
	}

	span.AddEvent("log", oteltrace.WithAttributes(obs.makeAttrs(entry, span)...))

	obs.logger.Emit(ctx, obs.makeRecord(entry, span))

	return nil
}

func (obs *observer) Levels() []logrus.Level {
	if obs == nil {
		return []logrus.Level{}
	}

	return obs.levels
}

func levelString(lvl logrus.Level) string {
	s := lvl.String()
	if s == "warning" {
		s = "warn"
	}
	return strings.ToUpper(s)
}

func attributeKeyValue(key string, value interface{}) attribute.KeyValue {
	switch value := value.(type) {
	case nil:
		return attribute.String(key, "<nil>")
	case string:
		return attribute.String(key, value)
	case int:
		return attribute.Int(key, value)
	case int64:
		return attribute.Int64(key, value)
	case uint64:
		return attribute.Int64(key, int64(value))
	case float64:
		return attribute.Float64(key, value)
	case bool:
		return attribute.Bool(key, value)
	case fmt.Stringer:
		return attribute.String(key, value.String())
	}

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Array:
		rv = rv.Slice(0, rv.Len())
		fallthrough
	case reflect.Slice:
		switch reflect.TypeOf(value).Elem().Kind() {
		case reflect.Bool:
			return attribute.BoolSlice(key, rv.Interface().([]bool))
		case reflect.Int:
			return attribute.IntSlice(key, rv.Interface().([]int))
		case reflect.Int64:
			return attribute.Int64Slice(key, rv.Interface().([]int64))
		case reflect.Float64:
			return attribute.Float64Slice(key, rv.Interface().([]float64))
		case reflect.String:
			return attribute.StringSlice(key, rv.Interface().([]string))
		default:
			return attribute.KeyValue{Key: attribute.Key(key)}
		}
	case reflect.Bool:
		return attribute.Bool(key, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64(key, rv.Int())
	case reflect.Float64:
		return attribute.Float64(key, rv.Float())
	case reflect.String:
		return attribute.String(key, rv.String())
	}
	if b, err := json.Marshal(value); b != nil && err == nil {
		return attribute.String(key, string(b))
	}
	return attribute.String(key, fmt.Sprint(value))
}

func logValue(value interface{}) otellog.Value {
	switch value := value.(type) {
	case nil:
		return otellog.StringValue("<nil>")
	case string:
		return otellog.StringValue(value)
	case int:
		return otellog.IntValue(value)
	case int64:
		return otellog.Int64Value(value)
	case uint64:
		return otellog.Int64Value(int64(value))
	case float64:
		return otellog.Float64Value(value)
	case bool:
		return otellog.BoolValue(value)
	case fmt.Stringer:
		return otellog.StringValue(value.String())
	}

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Array:
		rv = rv.Slice(0, rv.Len())
		fallthrough
	case reflect.Slice:
		values := make([]otellog.Value, rv.Len())
		for i := range values {
			values[i] = logValue(rv.Index(i).Interface())
		}
		return otellog.SliceValue(values...)
	case reflect.Bool:
		return otellog.BoolValue(rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return otellog.Int64Value(rv.Int())
	case reflect.Float64:
		return otellog.Float64Value(rv.Float())
	case reflect.String:
		return otellog.StringValue(rv.String())
	}
	if b, err := json.Marshal(value); err == nil {
		return otellog.StringValue(string(b))
	}
	return otellog.StringValue(fmt.Sprint(value))
}

func logSeverity(lvl logrus.Level) otellog.Severity {
	switch lvl {
	case logrus.PanicLevel, logrus.FatalLevel:
		return otellog.SeverityFatal
	case logrus.ErrorLevel:
		return otellog.SeverityError
	case logrus.WarnLevel:
		return otellog.SeverityWarn
	case logrus.InfoLevel:
		return otellog.SeverityInfo
	case logrus.DebugLevel:
		return otellog.SeverityDebug
	case logrus.TraceLevel:
		return otellog.SeverityTrace
	default:
		return otellog.SeverityUndefined
	}
}

func getStackTrace() []string {
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	depth = 0
	logrusPkgDepth := 0
	stack := make([]string, 0, depth)
	for f, again := frames.Next(); again; f, again = frames.Next() {
		depth++
		if getPackageName(f.Function) == logrusPackageName {
			logrusPkgDepth = depth
		}

		fileName := filepath.Base(f.File)
		stack = append(stack, fmt.Sprintf("%s(...) %s:%d", f.Function, fileName, f.Line))
	}

	return stack[logrusPkgDepth:]
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
