package searchers

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrNotConfigured is returned when Handle is called on an engine whose
// IsAvailable() is false. The orchestrator treats it as "skip immediately, do not
// retry". It is defensive: the orchestrator already checks IsAvailable() first.
var ErrNotConfigured = errors.New("search engine is not configured")

// RetryableError wraps a transient failure that MAY succeed if the SAME engine is
// retried: HTTP 429, HTTP 5xx, and network/transport errors (*url.Error, timeouts,
// context deadline on the outbound request). RetryAfter is a hint (0 if unknown).
type RetryableError struct {
	Err        error
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

// FatalError wraps a definitive failure that will NOT succeed on retry of the same
// engine: HTTP 4xx (auth, bad request, not found, method), response-decode errors,
// and unknown/unexpected status codes. The orchestrator moves to the next engine
// immediately, without wasting a retry.
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string { return e.Err.Error() }
func (e *FatalError) Unwrap() error { return e.Err }

// Retryable wraps err as a RetryableError with an optional RetryAfter hint.
func Retryable(err error, retryAfter time.Duration) error {
	return &RetryableError{Err: err, RetryAfter: retryAfter}
}

// Fatal wraps err as a FatalError.
func Fatal(err error) error {
	return &FatalError{Err: err}
}

// ClassifyHTTPStatus maps a non-2xx HTTP status to the right typed error so every
// engine classifies identically. The engine supplies a human-readable message; the
// retry/fatal decision is centralized here.
//
//	429       -> RetryableError (rate limited; back off and retry the same engine)
//	5xx       -> RetryableError (upstream problem; may clear on retry)
//	4xx/other -> FatalError     (auth/bad-request/etc.; retrying the same engine is pointless)
func ClassifyHTTPStatus(status int, msg string) error {
	switch {
	case status == http.StatusTooManyRequests:
		return Retryable(fmt.Errorf("%s (HTTP 429)", msg), 0)
	case status >= 500:
		return Retryable(fmt.Errorf("%s (HTTP %d)", msg, status), 0)
	default:
		return Fatal(fmt.Errorf("%s (HTTP %d)", msg, status))
	}
}

// IsRetryable reports whether err (or anything it wraps) is a RetryableError.
func IsRetryable(err error) bool {
	var t *RetryableError
	return errors.As(err, &t)
}

// IsFatal reports whether err (or anything it wraps) is a FatalError.
func IsFatal(err error) bool {
	var t *FatalError
	return errors.As(err, &t)
}
