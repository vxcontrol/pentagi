package internal

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"
)

const (
	defaultRetryAttempts = 2
	minRetryDelay        = 500 * time.Millisecond
	maxRetryDelay        = 5000 * time.Millisecond
)

// RetryOption adapts the behavior the *Retrier.
type RetryOption func(*retryOptions)

// RetryFunc is a retriable HTTP function call (i.e. *http.Client.Do).
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
// The request will be retried as long as the request is deemed retriable and the
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

	response, err := fn(request)
	if err != nil {
		return nil, err
	}

	if r.shouldRetry(response) {
		defer response.Body.Close()

		delay, err := r.retryDelay(retryAttempt)
		if err != nil {
			return nil, err
		}

		time.Sleep(delay)

		return r.run(
			fn,
			request,
			errorDecoder,
			maxRetryAttempts,
			retryAttempt+1,
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

// retryDelay calculates the delay time in milliseconds based on the retry attempt.
func (r *Retrier) retryDelay(retryAttempt uint) (time.Duration, error) {
	// Apply exponential backoff.
	delay := minRetryDelay + minRetryDelay*time.Duration(retryAttempt*retryAttempt)

	// Do not allow the number to exceed maxRetryDelay.
	if delay > maxRetryDelay {
		delay = maxRetryDelay
	}

	// Apply some itter by randomizing the value in the range of 75%-100%.
	max := big.NewInt(int64(delay / 4))
	jitter, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	delay -= time.Duration(jitter.Int64())

	// Never sleep less than the base sleep seconds.
	if delay < minRetryDelay {
		delay = minRetryDelay
	}

	return delay, nil
}

type retryOptions struct {
	attempts uint
}
