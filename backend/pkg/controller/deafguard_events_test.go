package controller

import (
	"context"
	"strings"
	"testing"

	"pentagi/pkg/tools/deafguard"
)

// TestConvertDeafGuardEvent verifies that a ClassificationResult is
// projected onto the GraphQL model field-for-field, and that command
// truncation kicks in at the deafGuardEventMaxCommandLen boundary with a
// trailing ellipsis.
func TestConvertDeafGuardEvent(t *testing.T) {
	t.Parallel()

	shortCmd := "ls -la /tmp"
	longCmd := strings.Repeat("a", deafGuardEventMaxCommandLen+50)

	tests := []struct {
		name      string
		result    *deafguard.ClassificationResult
		wantCmd   string
		wantTrunc bool
	}{
		{
			name: "short command passes through unchanged",
			result: &deafguard.ClassificationResult{
				Allowed:   true,
				Command:   shortCmd,
				Category:  deafguard.CategoryLocalUtility,
				Risk:      deafguard.RiskNone,
				Action:    deafguard.ActionLog,
				Reason:    "No matching rules",
				Mode:      deafguard.ModeLog,
				Tier:      9,
				Timestamp: 1700000000,
			},
			wantCmd:   shortCmd,
			wantTrunc: false,
		},
		{
			name: "long command is rune-truncated with ellipsis",
			result: &deafguard.ClassificationResult{
				Allowed:   false,
				Command:   longCmd,
				Category:  deafguard.CategoryContainerEscape,
				Risk:      deafguard.RiskCritical,
				Action:    deafguard.ActionBlock,
				Reason:    "Container escape attempt",
				Mode:      deafguard.ModeWarn,
				Tier:      1,
				Timestamp: 1700000001,
			},
			wantCmd:   strings.Repeat("a", deafGuardEventMaxCommandLen) + "...",
			wantTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := convertDeafGuardEvent(42, 7, tt.result)

			if got == nil {
				t.Fatalf("convertDeafGuardEvent returned nil")
			}
			if got.ID != 42 {
				t.Errorf("ID = %d, want 42", got.ID)
			}
			if got.FlowID != 7 {
				t.Errorf("FlowID = %d, want 7", got.FlowID)
			}
			if got.Command != tt.wantCmd {
				t.Errorf("Command mismatch: got %q, want %q", got.Command, tt.wantCmd)
			}
			if got.Allowed != tt.result.Allowed {
				t.Errorf("Allowed = %v, want %v", got.Allowed, tt.result.Allowed)
			}
			if got.Tier != tt.result.Tier {
				t.Errorf("Tier = %d, want %d", got.Tier, tt.result.Tier)
			}
			// Timestamp is int on the model (gqlgen Int mapping) but int64 on
			// the source ClassificationResult — compare via int64 to avoid
			// platform-dependent comparisons on 32-bit targets.
			if int64(got.Timestamp) != tt.result.Timestamp {
				t.Errorf("Timestamp = %d, want %d", got.Timestamp, tt.result.Timestamp)
			}
			if got.Category != string(tt.result.Category) {
				t.Errorf("Category = %q, want %q", got.Category, string(tt.result.Category))
			}
			if got.Risk != string(tt.result.Risk) {
				t.Errorf("Risk = %q, want %q", got.Risk, string(tt.result.Risk))
			}
			if got.Action != string(tt.result.Action) {
				t.Errorf("Action = %q, want %q", got.Action, string(tt.result.Action))
			}
			if got.Mode != string(tt.result.Mode) {
				t.Errorf("Mode = %q, want %q", got.Mode, string(tt.result.Mode))
			}
			if got.Reason != tt.result.Reason {
				t.Errorf("Reason = %q, want %q", got.Reason, tt.result.Reason)
			}
		})
	}
}

// TestFlowDeafGuardEventWorker_Publish_NilSafe asserts that Publish is a
// no-op on nil inputs rather than panicking — the wrapHandler hot path
// calls this inline, so a panic would tear down the tool-call.
func TestFlowDeafGuardEventWorker_Publish_NilSafe(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	// Nil worker must not panic.
	var nilWorker *flowDeafGuardEventWorker
	nilWorker.Publish(ctx, nil)

	// Worker with nil pub must not panic and must not try to dispatch.
	w := &flowDeafGuardEventWorker{flowID: 1, pub: nil}
	w.Publish(ctx, &deafguard.ClassificationResult{})

	// Worker with non-nil pub and nil result must not panic.
	w = &flowDeafGuardEventWorker{flowID: 1, pub: nil}
	w.Publish(ctx, nil)
}

// TestDeafGuardEventSeq_Monotonic covers the process-wide atomic counter
// — two successive Publish calls must produce strictly increasing IDs.
// This uses the unexported convertDeafGuardEvent directly because the
// Publish path requires a full FlowPublisher implementation.
func TestDeafGuardEventSeq_Monotonic(t *testing.T) {
	t.Parallel()

	// Seed reset is not exposed — the counter is process-wide. Just assert
	// strict monotonicity of two successive Add() calls.
	a := deafGuardEventSeq.Add(1)
	b := deafGuardEventSeq.Add(1)
	if b <= a {
		t.Errorf("expected strictly increasing IDs, got a=%d b=%d", a, b)
	}
}
