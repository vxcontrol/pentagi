package internal

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultRetryAttempts = 2
	minRetryDelay        = 1000 * time.Millisecond
	maxRetryDelay        = 60000 * time.Millisecond
)

// RetryOption adapts the behavior the *Retrier.
type RetryOption func(*retryOptions)

// RetryFunc is a retryable HTTP function call (i.e. *http.Client.Do).
type RetryFunc func(*http.Request) (*http.Response, error)

// WithMaxAttempts configures the maximum number of attempts
// of the *Retrier.
func WithMaxAttempts(attempts uint) RetryOption {
	return func(opts *retryOptions) {
		opts.attempts = attempts
	}
}

// Retrier retries failed requests a configurable number of times with an
// exponential back-off between each retry.
type Retrier struct {
	attempts uint
}

// NewRetrier constructs a new *Retrier with the given options, if any.
func NewRetrier(opts ...RetryOption) *Retrier {
	options := new(retryOptions)
	for _, opt := range opts {
		opt(options)
	}
	attempts := uint(defaultRetryAttempts)
	if options.attempts > 0 {
		attempts = options.attempts
	}
	return &Retrier{
		attempts: attempts,
	}
}

// Run issues the request and, upon failure, retries the request if possible.
//
// The request will be retried as long as the request is deemed retryable and the
// number of retry attempts has not grown larger than the configured retry limit.
func (r *Retrier) Run(
	fn RetryFunc,
	request *http.Request,
	errorDecoder ErrorDecoder,
	opts ...RetryOption,
) (*http.Response, error) {
	options := new(retryOptions)
	for _, opt := range opts {
		opt(options)
	}
	maxRetryAttempts := r.attempts
	if options.attempts > 0 {
		maxRetryAttempts = options.attempts
	}
	var (
		retryAttempt  uint
		previousError error
	)
	return r.run(
		fn,
		request,
		errorDecoder,
		maxRetryAttempts,
		retryAttempt,
		previousError,
	)
}

func (r *Retrier) run(
	fn RetryFunc,
	request *http.Request,
	errorDecoder ErrorDecoder,
	maxRetryAttempts uint,
	retryAttempt uint,
	previousError error,
) (*http.Response, error) {
	if retryAttempt >= maxRetryAttempts {
		return nil, previousError
	}

	// If the call has been cancelled, don't issue the request.
	if err := request.Context().Err(); err != nil {
		return nil, err
	}

	// Reset the request body for retries since the body may have already been read.
	if retryAttempt > 0 && request.GetBody != nil {
		requestBody, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		request.Body = requestBody
	}

	response, err := fn(request)
	if err != nil {
		return nil, err
	}

	if r.shouldRetry(response) {
		defer response.Body.Close()

		delay, err := r.retryDelay(response, retryAttempt)
		if err != nil {
			return nil, err
		}

		time.Sleep(delay)

		return r.run(
			fn,
			request,
			errorDecoder,
			maxRetryAttempts,
			retryAttempt + 1,
			decodeError(response, errorDecoder),
		)
	}

	return response, nil
}

// shouldRetry returns true if the request should be retried based on the given
// response status code.
func (r *Retrier) shouldRetry(response *http.Response) bool {
	return response.StatusCode == http.StatusTooManyRequests ||
		response.StatusCode == http.StatusRequestTimeout ||
		response.StatusCode >= http.StatusInternalServerError
}

// retryDelay calculates the delay time based on response headers,
// falling back to exponential backoff if no headers are present.
func (r *Retrier) retryDelay(response *http.Response, retryAttempt uint) (time.Duration, error) {
	// Check for Retry-After header first (RFC 7231), applying no jitter
	if retryAfter := response.Header.Get("Retry-After"); retryAfter != "" {
		// Parse as number of seconds...
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			delay := time.Duration(seconds) * time.Second
			if delay > 0 {
				if delay > maxRetryDelay {
					delay = maxRetryDelay
				}
				return delay, nil
			}
		}

		// ...or as an HTTP date; both are valid
		if retryTime, err := time.Parse(time.RFC1123, retryAfter); err == nil {
			delay := time.Until(retryTime)
			if delay > 0 {
				if delay > maxRetryDelay {
					delay = maxRetryDelay
				}
				return delay, nil
			}
		}
	}

	// Then check for industry-standard X-RateLimit-Reset header, applying positive jitter
	if rateLimitReset := response.Header.Get("X-RateLimit-Reset"); rateLimitReset != "" {
		if resetTimestamp, err := strconv.ParseInt(rateLimitReset, 10, 64); err == nil {
			// Assume Unix timestamp in seconds
			resetTime := time.Unix(resetTimestamp, 0)
			delay := time.Until(resetTime)
			if delay > 0 {
				if delay > maxRetryDelay {
					delay = maxRetryDelay
				}
				return r.addPositiveJitter(delay)
			}
		}
	}

	// Fall back to exponential backoff
	return r.exponentialBackoff(retryAttempt)
}

// exponentialBackoff calculates the delay time based on the retry attempt
// and applies symmetric jitter (±10% around the delay).
func (r *Retrier) exponentialBackoff(retryAttempt uint) (time.Duration, error) {
	if retryAttempt > 63 { // 2^63+ would overflow uint64
		retryAttempt = 63
	}

	delay := minRetryDelay << retryAttempt
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}

	return r.addSymmetricJitter(delay)
}

// addJitterWithRange applies jitter to the given delay.
// minPercent and maxPercent define the jitter range (e.g., 100, 120 for +0% to +20%).
func (r *Retrier) addJitterWithRange(delay time.Duration, minPercent, maxPercent int) (time.Duration, error) {
	jitterRange := big.NewInt(int64(delay * time.Duration(maxPercent - minPercent) / 100))
	jitter, err := rand.Int(rand.Reader, jitterRange)
	if err != nil {
		return 0, err
	}

	jitteredDelay := delay + time.Duration(jitter.Int64()) + delay * time.Duration(minPercent-100)/100
	if jitteredDelay < minRetryDelay {
		jitteredDelay = minRetryDelay
	}
	if jitteredDelay > maxRetryDelay {
		jitteredDelay = maxRetryDelay
	}
	return jitteredDelay, nil
}

// addPositiveJitter applies positive jitter to the given delay (100%-120% range).
func (r *Retrier) addPositiveJitter(delay time.Duration) (time.Duration, error) {
	return r.addJitterWithRange(delay, 100, 120)
}

// addSymmetricJitter applies symmetric jitter to the given delay (90%-110% range).
func (r *Retrier) addSymmetricJitter(delay time.Duration) (time.Duration, error) {
	return r.addJitterWithRange(delay, 90, 110)
}

type retryOptions struct {
	attempts uint
}
