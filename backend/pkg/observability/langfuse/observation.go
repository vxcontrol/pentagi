package langfuse

import (
	"context"
	"fmt"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/sirupsen/logrus"
)

type Observation interface {
	ID() string
	TraceID() string
	String() string
	Log(ctx context.Context, message string)
	Score(opts ...ScoreOption)
	Event(opts ...EventStartOption) Event
	Span(opts ...SpanStartOption) Span
	Generation(opts ...GenerationStartOption) Generation
}

type observation struct {
	obsCtx   observationContext
	observer enqueue
}

func (o *observation) ID() string {
	return o.obsCtx.ObservationID
}

func (o *observation) TraceID() string {
	return o.obsCtx.TraceID
}

func (o *observation) String() string {
	return fmt.Sprintf("Trace(%s) Observation(%s)", o.obsCtx.TraceID, o.obsCtx.ObservationID)
}

func (o *observation) Log(ctx context.Context, message string) {
	logID := newSpanID()
	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"langfuse_trace_id":       o.obsCtx.TraceID,
		"langfuse_observation_id": o.obsCtx.ObservationID,
		"langfuse_log_id":         logID,
	}).Info(message)

	obsLog := api.NewIngestionEventFromIngestionEventSeven(&api.IngestionEventSeven{
		Id:        logID,
		Timestamp: getCurrentTimeString(),
		Type:      ingestionPutLog,
		Body: &api.SdkLogBody{
			Log: message,
		},
	})

	o.observer.enqueue(obsLog)
}

func (o *observation) Score(opts ...ScoreOption) {
	opts = append(opts,
		withScoreTraceID(o.obsCtx.TraceID),
		withScoreParentObservationID(o.obsCtx.ObservationID),
	)
	newScore(o.observer, opts...)
}

func (o *observation) Event(opts ...EventStartOption) Event {
	opts = append(opts,
		withEventTraceID(o.obsCtx.TraceID),
		withEventParentObservationID(o.obsCtx.ObservationID),
	)
	return newEvent(o.observer, opts...)
}

func (o *observation) Span(opts ...SpanStartOption) Span {
	opts = append(opts,
		withSpanTraceID(o.obsCtx.TraceID),
		withSpanParentObservationID(o.obsCtx.ObservationID),
	)
	return newSpan(o.observer, opts...)
}

func (o *observation) Generation(opts ...GenerationStartOption) Generation {
	opts = append(opts,
		withGenerationTraceID(o.obsCtx.TraceID),
		withGenerationParentObservationID(o.obsCtx.ObservationID),
	)
	return newGeneration(o.observer, opts...)
}

type ObservationInfo struct {
	TraceID             string `json:"trace_id"`
	ObservationID       string `json:"observation_id"`
	ParentObservationID string `json:"parent_observation_id"`
}

type ObservationContext struct {
	TraceID       string
	TraceCtx      *TraceContext
	ObservationID string
}

type ObservationContextOption func(*ObservationContext)

func WithObservationTraceID(traceID string) ObservationContextOption {
	return func(o *ObservationContext) {
		o.TraceID = traceID
	}
}

func WithObservationID(observationID string) ObservationContextOption {
	return func(o *ObservationContext) {
		o.ObservationID = observationID
	}
}

func WithObservationTraceContext(opts ...TraceContextOption) ObservationContextOption {
	traceCtx := &TraceContext{
		Timestamp: getCurrentTimeRef(),
		Version:   getStringRef(firstVersion),
	}
	for _, opt := range opts {
		opt(traceCtx)
	}

	return func(o *ObservationContext) {
		o.TraceCtx = traceCtx
	}
}
