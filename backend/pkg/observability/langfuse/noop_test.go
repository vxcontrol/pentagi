package langfuse

import (
	"context"
	"testing"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopObserver_ImplementsObserver(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	require.NotNil(t, obs)

	// Verify it satisfies the Observer interface at compile time
	var _ Observer = obs
}

func TestNoopObserver_NewObservation_NewTrace(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	newCtx, observation := obs.NewObservation(ctx)

	// Should generate a new trace ID
	assert.NotEmpty(t, observation.TraceID())
	assert.Len(t, observation.TraceID(), 32, "trace ID should be 32 hex chars")

	// Context should contain the observation
	obsCtx, ok := getObservationContext(newCtx)
	require.True(t, ok)
	assert.Equal(t, observation.TraceID(), obsCtx.TraceID)
}

func TestNoopObserver_NewObservation_InheritsParentTrace(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Create parent observation
	ctx, parentObs := obs.NewObservation(ctx)
	parentTraceID := parentObs.TraceID()

	// Create child observation -- should inherit parent trace ID
	_, childObs := obs.NewObservation(ctx)
	assert.Equal(t, parentTraceID, childObs.TraceID(), "child must inherit parent trace ID")
}

func TestNoopObserver_NewObservation_ExplicitTraceID(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Set parent context with a different trace
	ctx = putObservationContext(ctx, observationContext{
		TraceID:       "parent-trace",
		ObservationID: "parent-obs",
	})

	// Explicit trace ID should override parent
	_, observation := obs.NewObservation(ctx, WithObservationTraceID("explicit-trace"))
	assert.Equal(t, "explicit-trace", observation.TraceID())
}

func TestNoopObserver_NewObservation_InheritsParentObservationID(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Set parent context
	ctx = putObservationContext(ctx, observationContext{
		TraceID:       "shared-trace",
		ObservationID: "parent-obs-id",
	})

	// Without explicit observation ID, should inherit parent
	_, observation := obs.NewObservation(ctx)
	assert.Equal(t, "shared-trace", observation.TraceID())
	assert.Equal(t, "parent-obs-id", observation.ID())
}

func TestNoopObserver_NewObservation_ExplicitObservationID(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Set parent context
	ctx = putObservationContext(ctx, observationContext{
		TraceID:       "shared-trace",
		ObservationID: "parent-obs-id",
	})

	// Explicit observation ID should override parent
	_, observation := obs.NewObservation(ctx, WithObservationID("my-obs"))
	assert.Equal(t, "shared-trace", observation.TraceID())
	assert.Equal(t, "my-obs", observation.ID())
}

func TestNoopObserver_Shutdown(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	err := obs.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestNoopObserver_ForceFlush(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	err := obs.ForceFlush(context.Background())
	assert.NoError(t, err)
}

func TestNoopObserver_Enqueue_NoPanic(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver().(*noopObserver)
	// enqueue should be a no-op and not panic
	assert.NotPanics(t, func() {
		obs.enqueue(nil)
		obs.enqueue(&api.IngestionEvent{})
	})
}
