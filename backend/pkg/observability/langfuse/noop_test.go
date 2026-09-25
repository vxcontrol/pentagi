package langfuse

import (
	"context"
	"testing"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoop_NoopObserver_LifecycleCallsDoNothing(t *testing.T) {
	t.Parallel()

	obs, ok := NewNoopObserver().(*noopObserver)
	require.True(t, ok, "NewNoopObserver must return *noopObserver")

	assert.NoError(t, obs.Shutdown(context.Background()))
	assert.NoError(t, obs.ForceFlush(context.Background()))
	assert.NotPanics(t, func() {
		obs.enqueue(nil)
		obs.enqueue(&api.IngestionEvent{})
	})
}

func TestNoop_NewObservation_InheritsOrGeneratesIDs(t *testing.T) {
	t.Parallel()

	parent := &observationContext{TraceID: "parent-trace", ObservationID: "parent-obs"}

	tests := []struct {
		name      string
		parent    *observationContext
		opts      []ObservationContextOption
		wantTrace string
		wantID    string
	}{
		{name: "no parent and no options start a new trace"},
		{name: "a parent lends its trace and observation", parent: parent, wantTrace: "parent-trace", wantID: "parent-obs"},
		{name: "an explicit observation id keeps the parent trace", parent: parent, opts: []ObservationContextOption{WithObservationID("my-obs")}, wantTrace: "parent-trace", wantID: "my-obs"},
		{name: "an explicit observation id without a parent starts a new trace", opts: []ObservationContextOption{WithObservationID("my-obs")}, wantID: "my-obs"},
		{name: "an explicit trace inherits nothing from the parent", parent: parent, opts: []ObservationContextOption{WithObservationTraceID("my-trace")}, wantTrace: "my-trace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.parent != nil {
				ctx = putObservationContext(ctx, *tt.parent)
			}

			newCtx, observation := NewNoopObserver().NewObservation(ctx, tt.opts...)

			if tt.wantTrace == "" {
				assert.Regexp(t, `^[0-9a-f]{32}$`, observation.TraceID(), "a new trace id is 32 lowercase hex characters")
			} else {
				assert.Equal(t, tt.wantTrace, observation.TraceID())
			}
			assert.Equal(t, tt.wantID, observation.ID())

			carried, ok := getObservationContext(newCtx)
			require.True(t, ok, "the returned context must carry the observation")
			assert.Equal(t, observationContext{TraceID: observation.TraceID(), ObservationID: observation.ID()}, carried)
		})
	}
}
