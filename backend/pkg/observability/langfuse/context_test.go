package langfuse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_GetObservationContext_ReturnsTheInnermostPut(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, ok := getObservationContext(ctx)
	assert.False(t, ok, "a context without an observation must report none")

	outer := putObservationContext(ctx, observationContext{TraceID: "trace-1", ObservationID: "obs-1"})
	inner := putObservationContext(outer, observationContext{TraceID: "trace-2", ObservationID: "obs-2"})

	got, ok := getObservationContext(inner)
	require.True(t, ok)
	assert.Equal(t, observationContext{TraceID: "trace-2", ObservationID: "obs-2"}, got)

	got, ok = getObservationContext(outer)
	require.True(t, ok)
	assert.Equal(t, observationContext{TraceID: "trace-1", ObservationID: "obs-1"}, got, "an inner put must not change the outer context")
}
