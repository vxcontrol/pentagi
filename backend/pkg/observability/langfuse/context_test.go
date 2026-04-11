package langfuse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservationContext_RoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	obsCtx := observationContext{
		TraceID:       "trace-abc",
		ObservationID: "obs-123",
	}

	ctx = putObservationContext(ctx, obsCtx)
	got, ok := getObservationContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "trace-abc", got.TraceID)
	assert.Equal(t, "obs-123", got.ObservationID)
}

func TestGetObservationContext_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, ok := getObservationContext(ctx)
	assert.False(t, ok)
}

func TestObservationContext_NestedOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// First level
	ctx1 := putObservationContext(ctx, observationContext{
		TraceID:       "trace-1",
		ObservationID: "obs-1",
	})

	// Second level overrides
	ctx2 := putObservationContext(ctx1, observationContext{
		TraceID:       "trace-2",
		ObservationID: "obs-2",
	})

	// Inner context sees the override
	got2, ok2 := getObservationContext(ctx2)
	require.True(t, ok2)
	assert.Equal(t, "trace-2", got2.TraceID)
	assert.Equal(t, "obs-2", got2.ObservationID)

	// Outer context still sees original
	got1, ok1 := getObservationContext(ctx1)
	require.True(t, ok1)
	assert.Equal(t, "trace-1", got1.TraceID)
	assert.Equal(t, "obs-1", got1.ObservationID)
}
