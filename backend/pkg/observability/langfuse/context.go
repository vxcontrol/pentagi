package langfuse

import "context"

type ObservationContextKey int

var observationContextKey ObservationContextKey

type observationContext struct {
	TraceID       string `json:"trace_id"`
	ObservationID string `json:"observation_id"`
}

func getObservationContext(ctx context.Context) (observationContext, bool) {
	obsCtx, ok := ctx.Value(observationContextKey).(observationContext)
	return obsCtx, ok
}

func putObservationContext(ctx context.Context, obsCtx observationContext) context.Context {
	return context.WithValue(ctx, observationContextKey, obsCtx)
}
