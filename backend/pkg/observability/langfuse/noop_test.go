package langfuse

import (
	"context"
	"regexp"
	"testing"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopObserver_ImplementsObserver(t *testing.T) {
	t.Parallel()

	// Compile-time interface check independent of constructor behavior
	var _ Observer = (*noopObserver)(nil)

	// Runtime check that constructor returns non-nil
	obs := NewNoopObserver()
	require.NotNil(t, obs)
}

func TestNoopObserver_NewObservation_NewTrace(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	newCtx, observation := obs.NewObservation(ctx)

	// Should generate a new trace ID - validate both length and hex format
	traceID := observation.TraceID()
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 32, "trace ID should be 32 characters")
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), traceID, "trace ID must be lowercase hex")

	// Context should contain the observation
	obsCtx, ok := getObservationContext(newCtx)
	require.True(t, ok)
	assert.Equal(t, traceID, obsCtx.TraceID)
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

func TestNoopObserver_NewObservation_ExplicitObsIDNoParent(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Explicit observation ID with no parent context
	_, observation := obs.NewObservation(ctx, WithObservationID("my-obs"))
	traceID := observation.TraceID()
	assert.NotEmpty(t, traceID, "must generate new trace ID")
	assert.Len(t, traceID, 32)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), traceID, "trace ID must be lowercase hex")
	assert.Equal(t, "my-obs", observation.ID())
}

func TestNoopObserver_NewObservation_ExplicitTraceDoesNotInheritObsID(t *testing.T) {
	t.Parallel()

	obs := NewNoopObserver()
	ctx := context.Background()

	// Set parent with both trace and observation IDs
	ctx = putObservationContext(ctx, observationContext{
		TraceID:       "parent-trace",
		ObservationID: "parent-obs",
	})

	// Explicit trace ID skips the inheritance block entirely
	_, observation := obs.NewObservation(ctx, WithObservationTraceID("my-trace"))
	assert.Equal(t, "my-trace", observation.TraceID())
	assert.Empty(t, observation.ID(), "must NOT inherit parent observation ID when trace is explicit")
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

	// Safe type assertion - test should fail gracefully if type changes
	obs, ok := NewNoopObserver().(*noopObserver)
	require.True(t, ok, "NewNoopObserver must return *noopObserver")

	// enqueue should be a no-op and not panic
	assert.NotPanics(t, func() {
		obs.enqueue(nil)
		obs.enqueue(&api.IngestionEvent{})
	})
}
