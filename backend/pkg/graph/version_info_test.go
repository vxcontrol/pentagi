package graph

import (
	"testing"
	"time"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/server/update"

	"github.com/stretchr/testify/assert"
)

func TestVersionInfo_VersionInfoFromStatus_NullsWhatIsNotKnown(t *testing.T) {
	t.Parallel()

	checked := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	failed := checked.Add(3 * time.Hour)
	latest := "2.4.0"

	for _, tt := range []struct {
		name   string
		status update.Status
		want   model.VersionInfo
	}{
		{
			// A formatted zero time is a date, and a date is a claim.
			name:   "nothing heard from the service leaves nulls",
			status: update.Status{Current: "2.1.0-93e99748", State: update.StatePending, Strategy: "preview"},
			want:   model.VersionInfo{Current: "2.1.0-93e99748", State: model.UpdateStatePending, Strategy: "preview"},
		},
		{
			name: "a verdict carries its evidence",
			status: update.Status{
				Current:   "2.1.0",
				State:     update.StateUpdateAvailable,
				Latest:    "2.4.0",
				Strategy:  "stable",
				CheckedAt: checked,
				FailedAt:  failed,
			},
			want: model.VersionInfo{
				Current:   "2.1.0",
				State:     model.UpdateStateUpdateAvailable,
				Latest:    &latest,
				Strategy:  "stable",
				CheckedAt: &checked,
				FailedAt:  &failed,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := versionInfoFromStatus(tt.status)
			got.Build = "" // set from the link-time version, not from the status

			assert.Equal(t, &tt.want, got)
		})
	}
}

// A state the schema has never heard of reads as "cannot say" rather than as a verdict.
func TestVersionInfo_UpdateStateFromStatus_NamesEveryStateInTheSchema(t *testing.T) {
	t.Parallel()

	for state, expected := range map[update.State]model.UpdateState{
		update.StateDisabled:        model.UpdateStateDisabled,
		update.StatePending:         model.UpdateStatePending,
		update.StateUnreachable:     model.UpdateStateUnreachable,
		update.StateUnknown:         model.UpdateStateUnknown,
		update.StateUpToDate:        model.UpdateStateUpToDate,
		update.StateUpdateAvailable: model.UpdateStateUpdateAvailable,
		update.State(""):            model.UpdateStateUnknown,
		update.State("surprise"):    model.UpdateStateUnknown,
	} {
		assert.Equal(t, expected, updateStateFromStatus(state), "state %q", state)
	}
}
