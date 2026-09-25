package cloud

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/vxcontrol/cloud/sdk"
)

func TestErrors_Classify_MapsSDKErrorsToActionableReasons(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantReason     FailureReason
		wantRetryAfter time.Duration
	}{
		{name: "no error", err: nil},
		{name: "forbidden", err: sdk.ErrForbidden, wantReason: FailureForbidden},
		{name: "blocked client", err: sdk.ErrBlocked, wantReason: FailureForbidden},
		{
			// A tier without access has no reset to wait for, so it must not read as "try again later".
			name:       "quota blocked is a permission problem, not a wait",
			err:        &sdk.QuotaError{Err: sdk.ErrQuotaBlocked, Scope: sdk.QuotaScopeBlocked},
			wantReason: FailureForbidden,
		},
		{
			name: "daily quota carries its reset",
			err: &sdk.QuotaError{
				Err: sdk.ErrQuotaExceededDaily, Scope: sdk.QuotaScopeDaily, RetryAfter: 3 * time.Hour,
			},
			wantReason:     FailureQuotaExceeded,
			wantRetryAfter: 3 * time.Hour,
		},
		{
			name: "rate limit carries its window",
			err: &sdk.RateLimitError{
				Err: sdk.ErrTooManyRequestsRPH, Scope: sdk.RateLimitScopeRPH, RetryAfter: 42 * time.Second,
			},
			wantReason:     FailureRateLimited,
			wantRetryAfter: 42 * time.Second,
		},
		{name: "not found is an answer", err: sdk.ErrNotFound, wantReason: FailureNotFound},
		{name: "malformed request is ours to fix", err: sdk.ErrBadRequest, wantReason: FailureRejected},
		{name: "server side failure", err: sdk.ErrServerInternal, wantReason: FailureServerError},
		{name: "bad gateway", err: sdk.ErrBadGateway, wantReason: FailureServerError},
		// The SDK does not retry a timed-out challenge, so this is a budget, not a broken network.
		{name: "proof of work budget", err: sdk.ErrExperimentTimeout, wantReason: FailureTimeout},
		{name: "caller deadline", err: context.DeadlineExceeded, wantReason: FailureTimeout},
		{name: "anything else is a reachability problem", err: errors.New("dial tcp: no route to host"), wantReason: FailureUnreachable},
		{name: "wrapped errors are still classified", err: fmt.Errorf("update check failed: %w", sdk.ErrForbidden), wantReason: FailureForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failure := Classify(tt.err)

			if tt.err == nil {
				if failure != nil {
					t.Fatalf("Classify(nil) = %v, want nil", failure)
				}
				return
			}

			if failure == nil {
				t.Fatalf("Classify(%v) = nil, want reason %q", tt.err, tt.wantReason)
			}
			if failure.Reason != tt.wantReason {
				t.Errorf("reason = %q, want %q", failure.Reason, tt.wantReason)
			}
			if failure.RetryAfter != tt.wantRetryAfter {
				t.Errorf("retry after = %v, want %v", failure.RetryAfter, tt.wantRetryAfter)
			}
			if !errors.Is(failure, tt.err) {
				t.Errorf("the original error is no longer reachable through the classification")
			}
		})
	}
}

// Waiting out a forbidden endpoint wastes the user's time; hiding a quota reset wastes the license.
func TestErrors_Retryable_MatchesWhatWaitingCanFix(t *testing.T) {
	retryable := map[FailureReason]bool{
		FailureUnreachable:   true,
		FailureTimeout:       true,
		FailureRateLimited:   true,
		FailureQuotaExceeded: true,
		FailureServerError:   true,
		FailureForbidden:     false,
		FailureNotFound:      false,
		FailureRejected:      false,
		FailureNone:          false,
	}

	for reason, want := range retryable {
		if got := reason.Retryable(); got != want {
			t.Errorf("%q.Retryable() = %t, want %t", reason, got, want)
		}
	}
}
