package graph

import (
	"pentagi/pkg/graph/model"
	"pentagi/pkg/server/update"
	"pentagi/pkg/version"
)

// versionInfoFromStatus is the GraphQL shape of what the update service knows.
//
// Zero times become nulls rather than the year 1: a "checked at" of 0001-01-01 is
// a date the UI would happily format, and "checked 2025 years ago" is worse than
// "not checked yet". An empty version becomes null for the same reason — the
// badge treats "no version named" as a verdict of its own.
func versionInfoFromStatus(status update.Status) *model.VersionInfo {
	info := &model.VersionInfo{
		Build:    version.GetBuildRevision(),
		Current:  status.Current,
		State:    updateStateFromStatus(status.State),
		Strategy: status.Strategy,
	}
	if status.Latest != "" {
		latest := status.Latest
		info.Latest = &latest
	}
	if !status.CheckedAt.IsZero() {
		checkedAt := status.CheckedAt
		info.CheckedAt = &checkedAt
	}
	if !status.FailedAt.IsZero() {
		failedAt := status.FailedAt
		info.FailedAt = &failedAt
	}
	return info
}

// updateStateFromStatus maps the service's vocabulary onto the schema's.
//
// Anything unrecognised — including the zero value of a Status that never went
// through the service — reads as unknown rather than as any verdict: the schema
// enum is closed, and "cannot say" is the only answer that is never wrong.
func updateStateFromStatus(state update.State) model.UpdateState {
	switch state {
	case update.StateDisabled:
		return model.UpdateStateDisabled
	case update.StatePending:
		return model.UpdateStatePending
	case update.StateUnreachable:
		return model.UpdateStateUnreachable
	case update.StateUpToDate:
		return model.UpdateStateUpToDate
	case update.StateUpdateAvailable:
		return model.UpdateStateUpdateAvailable
	default:
		return model.UpdateStateUnknown
	}
}
