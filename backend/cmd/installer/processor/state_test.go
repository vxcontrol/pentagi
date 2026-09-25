package processor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pins sendStarted and sendCompletion through each operation that fans out; subtests are keyed by operation.
func TestState_AnOperationIsAnnouncedOnceUnderTheStackAskedFor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		run     func(*processor, context.Context, ProductStack, *operationState) error
		stack   ProductStack
		wantErr string
	}{
		{name: "an update of every compose stack", run: (*processor).update, stack: ProductStackCompose},
		{name: "an update of everything", run: (*processor).update, stack: ProductStackAll},
		{name: "an update of a single stack", run: (*processor).update, stack: ProductStackObservability},
		{name: "a download of every compose stack", run: (*processor).download, stack: ProductStackCompose},
		{name: "a removal of every compose stack", run: (*processor).remove, stack: ProductStackCompose},
		{name: "a start of every compose stack", run: (*processor).start, stack: ProductStackCompose},
		{name: "a stop of every compose stack", run: (*processor).stop, stack: ProductStackCompose},
		{name: "an update whose first stack fails", run: (*processor).update, stack: ProductStackCompose,
			wantErr: "failed to update stack: compose exploded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := processorHarnessWith(t, func(c *mockCheckConfig) {
				c.PentagiIsUpToDate, c.GraphitiIsUpToDate, c.ObservabilityIsUpToDate = false, false, false
			})
			if tc.wantErr != "" {
				h.compose.setError("updateStack", errors.New("compose exploded"))
			}

			state := testOperationState(t)
			err := tc.run(h.p, t.Context(), tc.stack, state)

			var started []ProcessorStartedMsg
			var completed []ProcessorCompletionMsg
			for _, msg := range state.msgs {
				switch msg := msg.(type) {
				case ProcessorStartedMsg:
					started = append(started, msg)
				case ProcessorCompletionMsg:
					completed = append(completed, msg)
				}
			}
			require.Len(t, started, 1, "started announced %d times", len(started))
			require.Len(t, completed, 1, "the screen was told the operation finished %d times", len(completed))
			assert.Equal(t, tc.stack, started[0].Stack)
			assert.Equal(t, tc.stack, completed[0].Stack, "the completion must name the stack asked about, not a nested one")

			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.NoError(t, completed[0].Error)
				return
			}
			require.EqualError(t, err, tc.wantErr)
			assert.EqualError(t, completed[0].Error, tc.wantErr, "the one completion the screen sees must carry the failure")
		})
	}
}
