package langfuse

import (
	"context"

	"pentagi/pkg/observability/langfuse/api"
)

type noopObserver struct{}

func NewNoopObserver() Observer {
	return &noopObserver{}
}

func (o *noopObserver) NewObservation(
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

	obs := &observation{
		obsCtx: observationContext{
			TraceID:       obsCtx.TraceID,
			ObservationID: obsCtx.ObservationID,
		},
		observer: o,
	}

	return putObservationContext(ctx, obs.obsCtx), obs
}

func (o *noopObserver) Shutdown(ctx context.Context) error {
	return nil
}

func (o *noopObserver) ForceFlush(ctx context.Context) error {
	return nil
}

func (o *noopObserver) enqueue(event *api.IngestionEvent) {
}
