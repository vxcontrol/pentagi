package cloud

import (
	"context"
	"errors"
	"time"

	"github.com/vxcontrol/cloud/sdk"
)

// FailureReason classifies why a call to the update server did not produce an answer.
//
// The distinction matters to the user, not just to the log: "the server is unreachable"
// invites checking the network, "the daily quota is spent" invites coming back tomorrow,
// and "this request was refused" means retrying will never help. Collapsing all of them
// into one "update server unavailable" flag — as the installer used to — tells the user
// nothing they can act on.
type FailureReason string

const (
	// FailureNone is the zero value: no failure.
	FailureNone FailureReason = ""
	// FailureUnreachable covers everything that never reached a verdict: no network, no
	// DNS, a proxy that refused, a connection that died mid-flight.
	FailureUnreachable FailureReason = "unreachable"
	// FailureTimeout is a deadline: either the caller's, or the proof-of-work challenge
	// taking longer than the budget on slow hardware.
	FailureTimeout FailureReason = "timeout"
	// FailureRateLimited means too many requests in a rolling window. RetryAfter says when
	// the window reopens.
	FailureRateLimited FailureReason = "rate_limited"
	// FailureQuotaExceeded means the allowance for the period is spent. RetryAfter says
	// when it resets.
	FailureQuotaExceeded FailureReason = "quota_exceeded"
	// FailureForbidden means this client is not allowed to make this call at all — a
	// license tier without the endpoint, or a temporarily blocked client. Retrying does
	// not help; the license does.
	FailureForbidden FailureReason = "forbidden"
	// FailureNotFound means the server has no such package or version. It is an answer,
	// not an outage.
	FailureNotFound FailureReason = "not_found"
	// FailureRejected means the server refused the request as malformed. That is our bug,
	// not the user's, and retrying an identical request cannot succeed.
	FailureRejected FailureReason = "rejected"
	// FailureServerError means the server failed on its side. Worth retrying later.
	FailureServerError FailureReason = "server_error"
)

// Retryable reports whether waiting and trying again could plausibly succeed.
func (r FailureReason) Retryable() bool {
	switch r {
	case FailureUnreachable, FailureTimeout, FailureRateLimited,
		FailureQuotaExceeded, FailureServerError:
		return true
	default:
		return false
	}
}

// Failure is a classified call error. It wraps the original, so errors.Is and errors.As
// against the SDK sentinels keep working on it.
type Failure struct {
	Reason FailureReason
	// RetryAfter is the server-advertised cooldown, zero when the server did not say.
	RetryAfter time.Duration
	Err        error
}

func (f *Failure) Error() string { return f.Err.Error() }

func (f *Failure) Unwrap() error { return f.Err }

// Classify maps an SDK error onto a reason the interface can act on. It returns nil for a
// nil error, so it can wrap a call result directly.
//
// Ordering matters here: the typed errors are checked before the sentinels they wrap,
// because that is where the Retry-After cooldown lives.
func Classify(err error) *Failure {
	if err == nil {
		return nil
	}

	var quotaErr *sdk.QuotaError
	if errors.As(err, &quotaErr) {
		// A blocked scope is not an exhausted allowance — the tier never had access, so
		// there is nothing to wait for.
		if quotaErr.Scope == sdk.QuotaScopeBlocked {
			return &Failure{Reason: FailureForbidden, Err: err}
		}
		return &Failure{Reason: FailureQuotaExceeded, RetryAfter: sdk.RetryAfterOf(err), Err: err}
	}

	var rateLimitErr *sdk.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return &Failure{Reason: FailureRateLimited, RetryAfter: sdk.RetryAfterOf(err), Err: err}
	}

	switch {
	case errors.Is(err, sdk.ErrForbidden), errors.Is(err, sdk.ErrBlocked):
		return &Failure{Reason: FailureForbidden, Err: err}

	case errors.Is(err, sdk.ErrNotFound):
		return &Failure{Reason: FailureNotFound, Err: err}

	// A rejected request means we built it wrong. The replay and signature errors land
	// here too: both mean the server refused what we sent, and resending it is futile.
	case errors.Is(err, sdk.ErrBadRequest),
		errors.Is(err, sdk.ErrInvalidRequest),
		errors.Is(err, sdk.ErrInvalidConfiguration),
		errors.Is(err, sdk.ErrInvalidSignature),
		errors.Is(err, sdk.ErrReplayAttack):
		return &Failure{Reason: FailureRejected, Err: err}

	case errors.Is(err, sdk.ErrServerInternal), errors.Is(err, sdk.ErrBadGateway):
		return &Failure{Reason: FailureServerError, Err: err}

	// The proof-of-work budget is not retried by the SDK, so a client on slow hardware
	// has to be told what happened rather than shown a generic network error.
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, sdk.ErrExperimentTimeout),
		errors.Is(err, sdk.ErrPoWFailed):
		return &Failure{Reason: FailureTimeout, Err: err}

	default:
		return &Failure{Reason: FailureUnreachable, Err: err}
	}
}
