package langfuse

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"pentagi/pkg/observability/langfuse/api"
	"pentagi/pkg/observability/langfuse/api/core"

	"github.com/sirupsen/logrus"
)

const (
	defaultQueueSize    = 100
	defaultSendInterval = 10 * time.Second
	defaultTimeout      = 20 * time.Second
)

type Observer interface {
	NewObservation(
		ctx context.Context,
		opts ...ObservationContextOption,
	) (context.Context, Observation)
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}

type enqueue interface {
	enqueue(span *api.OtelSpan)
	enqueueScore(score *api.CreateScoreRequest)
}

type observer struct {
	mx          *sync.Mutex
	wg          *sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	client      *Client
	project     string
	release     string
	environment string
	interval    time.Duration
	timeout     time.Duration
	queueSize   int
	queue       chan *api.OtelSpan
	flusher     chan chan error
}

func NewObserver(client *Client, opts ...ObserverOption) Observer {
	ctx, cancel := context.WithCancel(context.Background())
	o := &observer{
		mx:      &sync.Mutex{},
		wg:      &sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
		client:  client,
		flusher: make(chan chan error),
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.interval <= 0 || o.interval > 10*time.Minute {
		o.interval = defaultSendInterval
	}
	if o.timeout <= 0 || o.timeout > 2*time.Minute {
		o.timeout = defaultTimeout
	}
	if o.queueSize <= 0 || o.queueSize > 10000 {
		o.queueSize = defaultQueueSize
	}
	o.queue = make(chan *api.OtelSpan, o.queueSize)

	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		o.sender()
	}()

	return o
}

func (o *observer) NewObservation(
	ctx context.Context,
	opts ...ObservationContextOption,
) (context.Context, Observation) {
	var obsCtx ObservationContext
	for _, opt := range opts {
		opt(&obsCtx)
	}

	parentObsCtx, parentObsCtxFound := getObservationContext(ctx)

	if obsCtx.TraceID == "" { // wants to use parent trace id in general or create new one
		if parentObsCtxFound && parentObsCtx.TraceID != "" {
			obsCtx.TraceID = parentObsCtx.TraceID
			if obsCtx.ObservationID == "" { // wants to use parent observation id in general
				obsCtx.ObservationID = parentObsCtx.ObservationID
			}
		} else {
			obsCtx.TraceID = newTraceID()
		}
	}

	var mergedTraceCtx *TraceContext
	if parentObsCtxFound && parentObsCtx.TraceCtx != nil {
		mergedTraceCtx = parentObsCtx.TraceCtx
	}
	if obsCtx.TraceCtx != nil {
		mergedTraceCtx = mergeTraceContext(mergedTraceCtx, obsCtx.TraceCtx)
	} else {
		mergedTraceCtx = cloneTraceContext(mergedTraceCtx)
	}

	obs := &observation{
		obsCtx: observationContext{
			TraceID:       obsCtx.TraceID,
			ObservationID: obsCtx.ObservationID,
			TraceCtx:      mergedTraceCtx,
		},
		observer: o,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (o *observer) Shutdown(ctx context.Context) error {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil {
		return err
	}

	o.cancel()
	o.wg.Wait()
	close(o.flusher)

	return nil
}

func (o *observer) ForceFlush(ctx context.Context) error {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil {
		return err
	}

	ch := make(chan error)
	o.flusher <- ch

	return <-ch
}

func (o *observer) enqueue(span *api.OtelSpan) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil || span == nil {
		return
	}

	if o.release != "" {
		span.Attributes = append(span.Attributes, otelStringAttr(otelAttrRelease, o.release))
	}
	if o.environment != "" {
		span.Attributes = append(span.Attributes, otelStringAttr(otelAttrEnvironment, o.environment))
	}

	o.queue <- span
}

func (o *observer) enqueueScore(scoreReq *api.CreateScoreRequest) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if err := o.ctx.Err(); err != nil || scoreReq == nil || o.client == nil || o.client.Score == nil {
		return
	}

	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		ctx, cancel := context.WithTimeout(o.ctx, o.timeout)
		defer cancel()
		if _, err := o.client.Score.Create(ctx, scoreReq); err != nil && o.ctx.Err() == nil {
			logrus.WithContext(o.ctx).WithError(err).Error("failed to send score to Langfuse")
		}
	}()
}

func (o *observer) flush(ctx context.Context, batch []*api.OtelSpan) error {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	if len(batch) == 0 {
		return nil
	}

	resourceAttrs := []*api.OtelAttribute{
		otelStringAttr("service.name", "pentagi"),
		otelStringAttr("telemetry.sdk.language", "go"),
		otelStringAttr("telemetry.sdk.name", "langfuse-observability-client-go"),
		otelStringAttr("telemetry.sdk.version", InstrumentationVersion),
	}
	if o.release != "" {
		resourceAttrs = append(resourceAttrs, otelStringAttr("service.version", o.release))
	}
	if o.environment != "" {
		resourceAttrs = append(resourceAttrs, otelStringAttr("deployment.environment", o.environment))
	}

	req := &api.OpentelemetryExportTracesRequest{
		ResourceSpans: []*api.OtelResourceSpan{
			{
				Resource: &api.OtelResource{
					Attributes: resourceAttrs,
				},
				ScopeSpans: []*api.OtelScopeSpan{
					{
						Scope: &api.OtelScope{
							Name:    getStringRef("langfuse-sdk"),
							Version: getStringRef(InstrumentationVersion),
						},
						Spans: batch,
					},
				},
			},
		},
	}

	_, err := o.client.Opentelemetry.Exporttraces(ctx, req)
	return err
}

// flushWithSplit flushes the batch and, if the request fails only because
// the serialized payload exceeded Langfuse's body-size limit (413), splits
// the batch in half and retries each half recursively instead of discarding
// every event in the batch on a single oversized flush. Only an event that
// still doesn't fit the limit on its own is ever dropped, and that is logged
// with its approximate size so the loss is visible instead of silent.
func (o *observer) flushWithSplit(ctx context.Context, batch []*api.OtelSpan) error {
	if len(batch) == 0 {
		return nil
	}

	err := o.flush(ctx, batch)
	if err == nil {
		return nil
	}

	var apiErr *core.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		// Not a body-size problem (network error, auth failure, a 5xx from
		// Langfuse itself, ...) - splitting would not help here and would
		// just multiply requests, so surface the error as-is.
		return err
	}

	if len(batch) == 1 {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"event_size_bytes": approxSpanSize(batch[0]),
		}).Error("dropping a single telemetry event that exceeds Langfuse's body size limit even alone")
		return nil
	}

	mid := len(batch) / 2
	errFirst := o.flushWithSplit(ctx, batch[:mid])
	errSecond := o.flushWithSplit(ctx, batch[mid:])
	return errors.Join(errFirst, errSecond)
}

// approxSpanSize returns the serialized size of a single span, or
// -1 if it can't be marshaled - used only for the drop-log message above, so
// a marshal failure there must not itself fail or panic.
func approxSpanSize(span *api.OtelSpan) int {
	data, err := json.Marshal(span)
	if err != nil {
		return -1
	}
	return len(data)
}

func (o *observer) sender() {
	batch := make([]*api.OtelSpan, 0, o.queueSize)
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case ch := <-o.flusher:
			ch <- o.flushWithSplit(o.ctx, batch)
			batch = batch[:0]
			ticker.Reset(o.interval)
		case <-ticker.C:
			if err := o.flushWithSplit(o.ctx, batch); err != nil {
				logrus.WithContext(o.ctx).WithError(err).Error("failed to flush events by interval")
			}
			batch = batch[:0]
			ticker.Reset(o.interval)
		case span := <-o.queue:
			batch = append(batch, span)
			if len(batch) >= o.queueSize {
				if err := o.flushWithSplit(o.ctx, batch); err != nil {
					logrus.WithContext(o.ctx).WithError(err).Error("failed to flush events by queue size")
				}
				batch = batch[:0]
				ticker.Reset(o.interval)
			}
		}
	}
}
