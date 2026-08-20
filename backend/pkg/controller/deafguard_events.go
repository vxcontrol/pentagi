package controller

import (
	"context"
	"sync/atomic"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/tools/deafguard"
)

// deafGuardEventMaxCommandLen bounds the event command payload so the
// live GraphQL subscription does not fan out unbounded agent input.
// Matches the truncation used by the Deaf Guard's structured logger.
const deafGuardEventMaxCommandLen = 200

// deafGuardEventSeq is a process-wide monotonic counter for Deaf Guard
// event IDs. Events are not persisted, so the ID only needs to be unique
// within the lifetime of the process — enough for the frontend to use it
// as a stable React key and to dedupe in-flight pairs. Kept at package
// scope so concurrent publishers across different flows never collide.
var deafGuardEventSeq atomic.Int64

// FlowDeafGuardEventWorker is the per-flow publisher for Deaf Guard
// classification events. It implements tools.DeafGuardEventProvider, but
// intentionally is not imported against that name here to avoid a
// circular dependency between the controller and tools packages.
type FlowDeafGuardEventWorker interface {
	Publish(ctx context.Context, result *deafguard.ClassificationResult)
}

type flowDeafGuardEventWorker struct {
	flowID int64
	pub    subscriptions.FlowPublisher
}

// NewFlowDeafGuardEventWorker constructs a publisher bound to the given
// flow. The returned worker is safe to invoke from any goroutine — the
// underlying subscription channel handles concurrent sends.
func NewFlowDeafGuardEventWorker(flowID int64, pub subscriptions.FlowPublisher) FlowDeafGuardEventWorker {
	return &flowDeafGuardEventWorker{
		flowID: flowID,
		pub:    pub,
	}
}

// Publish converts a ClassificationResult into the GraphQL model and fans
// it out to any live subscribers for this flow. A nil result is a no-op.
// The publish call itself is non-blocking on the caller's side — the
// underlying Channel drops the send if a subscriber is slow, matching the
// existing terminalLogAdded semantics.
func (w *flowDeafGuardEventWorker) Publish(ctx context.Context, result *deafguard.ClassificationResult) {
	if result == nil || w == nil || w.pub == nil {
		return
	}

	id := deafGuardEventSeq.Add(1)
	event := convertDeafGuardEvent(id, w.flowID, result)
	w.pub.DeafGuardEventAdded(ctx, event)
}

// convertDeafGuardEvent projects an in-memory ClassificationResult onto
// the GraphQL model. All typed fields on ClassificationResult are
// stringified here — the GraphQL schema intentionally uses String
// (rather than dedicated enums) so that new categories, actions, or
// risk levels added to the Go enum do not force a gqlgen regeneration
// on every rule-pack update.
func convertDeafGuardEvent(
	id, flowID int64,
	result *deafguard.ClassificationResult,
) *model.DeafGuardEvent {
	return &model.DeafGuardEvent{
		ID:     id,
		FlowID: flowID,
		// Timestamp is Unix-seconds from ClassificationResult (int64). gqlgen
		// maps GraphQL Int to Go int — on all PentAGI target platforms int is
		// 64-bit, so the cast is lossless for any realistic timestamp.
		Timestamp: int(result.Timestamp),
		Command:   truncateCommand(result.Command, deafGuardEventMaxCommandLen),
		Category:  string(result.Category),
		Tier:      result.Tier,
		Risk:      string(result.Risk),
		Action:    string(result.Action),
		Allowed:   result.Allowed,
		Mode:      string(result.Mode),
		Reason:    result.Reason,
	}
}

// truncateCommand returns the first maxLen runes of s followed by an
// ellipsis when truncation occurred. Rune-safe to match the Deaf Guard's
// own log-truncation helper; the two are kept separate so a change to
// one does not silently change the other.
func truncateCommand(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
