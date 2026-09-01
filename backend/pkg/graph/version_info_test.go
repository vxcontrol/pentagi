package graph

import (
	"testing"
	"time"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/server/update"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A status that has never heard from the service carries zero times and no
// version. Those have to leave as nulls: a formatted zero time is a date, and a
// date is a claim.
func TestVersionInfo_NothingKnownIsNull(t *testing.T) {
	t.Parallel()

	info := versionInfoFromStatus(update.Status{
		Current:  "2.1.0-93e99748",
		State:    update.StatePending,
		Strategy: "preview",
	})

	assert.Equal(t, "2.1.0-93e99748", info.Current)
	assert.Equal(t, model.UpdateStatePending, info.State)
	assert.Equal(t, "preview", info.Strategy)
	assert.Nil(t, info.Latest)
	assert.Nil(t, info.CheckedAt)
	assert.Nil(t, info.FailedAt)
}

func TestVersionInfo_AVerdictCarriesItsEvidence(t *testing.T) {
	t.Parallel()

	checked := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	failed := checked.Add(3 * time.Hour)

	info := versionInfoFromStatus(update.Status{
		Current:   "2.1.0",
		State:     update.StateUpdateAvailable,
		Latest:    "2.4.0",
		Strategy:  "stable",
		CheckedAt: checked,
		FailedAt:  failed,
	})

	assert.Equal(t, model.UpdateStateUpdateAvailable, info.State)
	require.NotNil(t, info.Latest)
	assert.Equal(t, "2.4.0", *info.Latest)
	require.NotNil(t, info.CheckedAt)
	assert.True(t, info.CheckedAt.Equal(checked))
	require.NotNil(t, info.FailedAt)
	assert.True(t, info.FailedAt.Equal(failed))
}

// Every state the service can be in has a name in the schema, and a state the
// schema has never heard of reads as "cannot say" rather than as a verdict.
func TestVersionInfo_EveryStateHasASchemaName(t *testing.T) {
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
		assert.True(t, expected.IsValid(), "state %q maps outside the schema", state)
	}
}
