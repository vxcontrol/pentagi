package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pentagi/pkg/observability/langfuse/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RetryTestCase struct {
	description string

	giveAttempts    uint
	giveStatusCodes []int
	giveResponse    *InternalTestResponse

	wantResponse *InternalTestResponse
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
			giveResponse: &InternalTestResponse{
				Id: "1",
			},
			wantResponse: &InternalTestResponse{
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

			var response *InternalTestResponse
			_, err := caller.Call(
				context.Background(),
				&CallParams{
					URL:                server.URL,
					Method:             http.MethodGet,
					Request:            &InternalTestRequest{},
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
					expectedDurationMin := expectedRetryDurations[index-1] * 50 / 100
					expectedDurationMax := expectedRetryDurations[index-1] * 150 / 100
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

				request := new(InternalTestRequest)
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
// Values are calculated based off of `minRetryDelay * 2^i`.
var expectedRetryDurations = []time.Duration{
	1000 * time.Millisecond, // 500ms * 2^1 = 1000ms
	2000 * time.Millisecond, // 500ms * 2^2 = 2000ms
	4000 * time.Millisecond, // 500ms * 2^3 = 4000ms
	8000 * time.Millisecond, // 500ms * 2^4 = 8000ms
}

func TestRetryWithRequestBody(t *testing.T) {
	// This test verifies that POST requests with a body are properly retried.
	// The request body should be re-sent on each retry attempt.
	expectedBody := `{"id":"test-id"}`
	var requestBodies []string
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		requestBodies = append(requestBodies, string(bodyBytes))

		if requestCount == 1 {
			// First request - return retryable error
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Second request - return success
		w.WriteHeader(http.StatusOK)
		response := &InternalTestResponse{Id: "success"}
		bytes, _ := json.Marshal(response)
		w.Write(bytes)
	}))
	defer server.Close()

	caller := NewCaller(&CallerParams{
		Client: server.Client(),
	})

	var response *InternalTestResponse
	_, err := caller.Call(
		context.Background(),
		&CallParams{
			URL:                server.URL,
			Method:             http.MethodPost,
			Request:            &InternalTestRequest{Id: "test-id"},
			Response:           &response,
			MaxAttempts:        2,
			ResponseIsOptional: true,
		},
	)

	require.NoError(t, err)
	require.Equal(t, 2, requestCount, "Expected exactly 2 requests")
	require.Len(t, requestBodies, 2, "Expected 2 request bodies to be captured")

	// Both requests should have the same non-empty body
	assert.Equal(t, expectedBody, requestBodies[0], "First request body should match expected")
	assert.Equal(t, expectedBody, requestBodies[1], "Second request body should match expected (retry should re-send body)")
}

func TestRetryDelayTiming(t *testing.T) {
	tests := []struct {
		name            string
		headerName      string
		headerValueFunc func() string
		expectedMinMs   int64
		expectedMaxMs   int64
	}{
		{
			name:       "retry-after with seconds value",
			headerName: "retry-after",
			headerValueFunc: func() string {
				return "1"
			},
			expectedMinMs: 500,
			expectedMaxMs: 1500,
		},
		{
			name:       "retry-after with HTTP date",
			headerName: "retry-after",
			headerValueFunc: func() string {
				return time.Now().Add(3 * time.Second).Format(time.RFC1123)
			},
			expectedMinMs: 1500,
			expectedMaxMs: 4500,
		},
		{
			name:       "x-ratelimit-reset with future timestamp",
			headerName: "x-ratelimit-reset",
			headerValueFunc: func() string {
				return fmt.Sprintf("%d", time.Now().Add(3 * time.Second).Unix())
			},
			expectedMinMs: 1500,
			expectedMaxMs: 4500,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var timestamps []time.Time
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				timestamps = append(timestamps, time.Now())
				if len(timestamps) == 1 {
					// First request - return retryable error with header
					w.Header().Set(tt.headerName, tt.headerValueFunc())
					w.WriteHeader(http.StatusTooManyRequests)
				} else {
					// Second request - return success
					w.WriteHeader(http.StatusOK)
					response := &InternalTestResponse{Id: "success"}
					bytes, _ := json.Marshal(response)
					w.Write(bytes)
				}
			}))
			defer server.Close()

			caller := NewCaller(&CallerParams{
				Client: server.Client(),
			})

			var response *InternalTestResponse
			_, err := caller.Call(
				context.Background(),
				&CallParams{
					URL:                server.URL,
					Method:             http.MethodGet,
					Request:            &InternalTestRequest{},
					Response:           &response,
					MaxAttempts:        2,
					ResponseIsOptional: true,
				},
			)

			require.NoError(t, err)
			require.Len(t, timestamps, 2, "Expected exactly 2 requests")

			actualDelayMs := timestamps[1].Sub(timestamps[0]).Milliseconds()

			assert.GreaterOrEqual(t, actualDelayMs, tt.expectedMinMs,
				"Actual delay %dms should be >= expected min %dms", actualDelayMs, tt.expectedMinMs)
			assert.LessOrEqual(t, actualDelayMs, tt.expectedMaxMs,
				"Actual delay %dms should be <= expected max %dms", actualDelayMs, tt.expectedMaxMs)
		})
	}
}
