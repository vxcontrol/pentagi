package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pentagi/pkg/observability/langfuse/api/core"
)

type RetryTestCase struct {
	description string

	giveAttempts    uint
	giveStatusCodes []int
	giveResponse    *Response

	wantResponse *Response
	wantError    *core.APIError
}

func TestRetrier(t *testing.T) {
	tests := []*RetryTestCase{
		{
			description:  "retry request succeeds after multiple failures",
			giveAttempts: 3,
			giveStatusCodes: []int{
				http.StatusServiceUnavailable,
				http.StatusServiceUnavailable,
				http.StatusOK,
			},
			giveResponse: &Response{
				Id: "1",
			},
			wantResponse: &Response{
				Id: "1",
			},
		},
		{
			description:  "retry request fails if MaxAttempts is exceeded",
			giveAttempts: 3,
			giveStatusCodes: []int{
				http.StatusRequestTimeout,
				http.StatusRequestTimeout,
				http.StatusRequestTimeout,
				http.StatusOK,
			},
			wantError: &core.APIError{
				StatusCode: http.StatusRequestTimeout,
			},
		},
		{
			description:  "retry durations increase exponentially and stay within the min and max delay values",
			giveAttempts: 4,
			giveStatusCodes: []int{
				http.StatusServiceUnavailable,
				http.StatusServiceUnavailable,
				http.StatusServiceUnavailable,
				http.StatusOK,
			},
		},
		{
			description:     "retry does not occur on status code 404",
			giveAttempts:    2,
			giveStatusCodes: []int{http.StatusNotFound, http.StatusOK},
			wantError: &core.APIError{
				StatusCode: http.StatusNotFound,
			},
		},
		{
			description:     "retries occur on status code 429",
			giveAttempts:    2,
			giveStatusCodes: []int{http.StatusTooManyRequests, http.StatusOK},
		},
		{
			description:     "retries occur on status code 408",
			giveAttempts:    2,
			giveStatusCodes: []int{http.StatusRequestTimeout, http.StatusOK},
		},
		{
			description:     "retries occur on status code 500",
			giveAttempts:    2,
			giveStatusCodes: []int{http.StatusInternalServerError, http.StatusOK},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			var (
				test   = tc
				server = newTestRetryServer(t, test)
				client = server.Client()
			)

			t.Parallel()

			caller := NewCaller(
				&CallerParams{
					Client: client,
				},
			)

			var response *Response
			err := caller.Call(
				context.Background(),
				&CallParams{
					URL:                server.URL,
					Method:             http.MethodGet,
					Request:            &Request{},
					Response:           &response,
					MaxAttempts:        test.giveAttempts,
					ResponseIsOptional: true,
				},
			)

			if test.wantError != nil {
				require.IsType(t, err, &core.APIError{})
				expectedErrorCode := test.wantError.StatusCode
				actualErrorCode := err.(*core.APIError).StatusCode
				assert.Equal(t, expectedErrorCode, actualErrorCode)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantResponse, response)
		})
	}
}

// newTestRetryServer returns a new *httptest.Server configured with the
// given test parameters, suitable for testing retries.
func newTestRetryServer(t *testing.T, tc *RetryTestCase) *httptest.Server {
	var index int
	timestamps := make([]time.Time, 0, len(tc.giveStatusCodes))

	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				timestamps = append(timestamps, time.Now())
				if index > 0 && index < len(expectedRetryDurations) {
					// Ensure that the duration between retries increases exponentially,
					// and that it is within the minimum and maximum retry delay values.
					actualDuration := timestamps[index].Sub(timestamps[index-1])
					expectedDurationMin := expectedRetryDurations[index-1] * 75 / 100
					expectedDurationMax := expectedRetryDurations[index-1] * 125 / 100
					assert.True(
						t,
						actualDuration >= expectedDurationMin && actualDuration <= expectedDurationMax,
						"expected duration to be in range [%v, %v], got %v",
						expectedDurationMin,
						expectedDurationMax,
						actualDuration,
					)
					assert.LessOrEqual(
						t,
						actualDuration,
						maxRetryDelay,
						"expected duration to be less than the maxRetryDelay (%v), got %v",
						maxRetryDelay,
						actualDuration,
					)
					assert.GreaterOrEqual(
						t,
						actualDuration,
						minRetryDelay,
						"expected duration to be greater than the minRetryDelay (%v), got %v",
						minRetryDelay,
						actualDuration,
					)
				}

				request := new(Request)
				bytes, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(bytes, request))
				require.LessOrEqual(t, index, len(tc.giveStatusCodes))

				statusCode := tc.giveStatusCodes[index]
				w.WriteHeader(statusCode)

				if tc.giveResponse != nil && statusCode == http.StatusOK {
					bytes, err = json.Marshal(tc.giveResponse)
					require.NoError(t, err)
					_, err = w.Write(bytes)
					require.NoError(t, err)
				}

				index++
			},
		),
	)
}

// expectedRetryDurations holds an array of calculated retry durations,
// where the index of the array should correspond to the retry attempt.
//
// Values are calculated based off of `minRetryDelay + minRetryDelay*i*i`, with
// a max and min value of 5000ms and 500ms respectively.
var expectedRetryDurations = []time.Duration{
	500 * time.Millisecond,
	1000 * time.Millisecond,
	2500 * time.Millisecond,
	5000 * time.Millisecond,
	5000 * time.Millisecond,
}
