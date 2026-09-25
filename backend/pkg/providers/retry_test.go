package providers

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// A local refusal or a vendor 4xx fails the same way again, except 408 and 429.
func TestRetry_IsDeterministicError_TellsARepeatableFailureFromATransientOne(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"reasoning cannot be turned off", &reasoning.ErrReasoningOffUnsupported{Model: "claude-fable-5-1"}, true},
		{"forced tool use with thinking", &reasoning.ErrForcedToolUseWithThinking{Model: "claude-fable-5-1"}, true},
		{"assistant prefill with thinking", &reasoning.ErrAssistantPrefillUnsupported{Model: "claude-fable-5-1"}, true},
		{"effort has no budget", &reasoning.ErrEffortHasNoBudget{Model: "claude-fable-5-1", Effort: "high"}, true},
		{"effort with tools", &reasoning.ErrEffortWithTools{Model: "gpt-5.5", Effort: "high"}, true},
		{"budget unsupported", &reasoning.ErrThinkingBudgetUnsupported{Model: "gpt-5.5"}, true},
		{"thinking requires stream", &reasoning.ErrThinkingRequiresStream{Model: "qwen3.6-35b-a3b"}, true},
		{"structured output unsupported", &llms.ErrStructuredOutputUnsupported{Model: "claude-fable-5-1"}, true},
		{"structured output conflict", &llms.ErrStructuredOutputConflict{}, true},
		{"structured output config", llms.ErrStructuredOutputConfig, true},
		{"provider content filter", llms.NewError(llms.ErrCodeContentFilter, "googleai", "prompt blocked"), true},
		{"wrapped provider content filter", fmt.Errorf("generator: %w", llms.NewError(llms.ErrCodeContentFilter, "googleai", "prompt blocked")), true},
		{"wrapped refusal", fmt.Errorf("call chain: %w", &reasoning.ErrReasoningOffUnsupported{Model: "x"}), true},
		{"vendor rejected the request", errors.New("API returned unexpected status code: 400: unsupported parameter"), true},
		{"vendor rejected the credentials", errors.New("API returned unexpected status code: 401: invalid api key"), true},
		{"vendor refused access", errors.New("API returned unexpected status code: 403: permission denied"), true},
		{"vendor has no such model", errors.New("API returned unexpected status code: 404: the model does not exist"), true},
		{"vendor rejected the payload", errors.New("API returned unexpected status code: 413: payload too large"), true},
		{"no error", nil, false},
		{"rate limited", errors.New("API returned unexpected status code: 429: too many requests"), false},
		{"request timed out", errors.New("API returned unexpected status code: 408: request timeout"), false},
		{"gateway is down", errors.New("API returned unexpected status code: 502: bad gateway"), false},
		// The one the LLM gateway in front of this deployment actually returns.
		{"gateway timed out", errors.New("API returned unexpected status code: 504: " +
			"<html>\r\n<head><title>504 Gateway Time-out</title></head>\r\n</html>"), false},
		{"vendor failed", errors.New("API returned unexpected status code: 500: internal server error"), false},
		{"connection dropped", errors.New("connection refused"), false},
		{"a digit run that spells a status", errors.New("API returned unexpected status code: 503: maximum 4000 tokens"), false},
		{"output stopped at the limit", &llms.ErrStructuredOutputValidation{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isDeterministicError(tt.err))
		})
	}
}

// Rows script each attempt's error; once the script runs out, its last entry repeats.
func TestRetry_RetryTransient_AsksAgainOnlyAfterATransientFailure(t *testing.T) {
	t.Parallel()

	gateway := errors.New("API returned unexpected status code: 502: bad gateway")
	timeout := errors.New("API returned unexpected status code: 504: gateway timeout")
	rejected := errors.New("API returned unexpected status code: 400: bad request")
	cancelled := fmt.Errorf("request aborted: %w", context.Canceled)
	filtered := llms.NewError(llms.ErrCodeContentFilter, "googleai", "prompt blocked")

	tests := []struct {
		name      string
		delay     time.Duration
		errs      []error
		want      string
		wantCalls int
		wantIs    error
		wantMsg   string
	}{
		{
			name:      "a transient failure is asked again and the later answer returned",
			delay:     100 * time.Millisecond,
			errs:      []error{gateway, nil},
			want:      "ok",
			wantCalls: 2,
		},
		{
			name:      "a failure that persists gives up after three attempts, waiting between them",
			delay:     100 * time.Millisecond,
			errs:      []error{timeout},
			wantCalls: 3,
			wantIs:    timeout,
			wantMsg:   "gave up after 3 attempts",
		},
		{
			name:      "a rejected request is not asked again",
			delay:     2 * time.Second,
			errs:      []error{rejected},
			wantCalls: 1,
			wantIs:    rejected,
		},
		{
			name:      "a content filter is not asked again",
			delay:     2 * time.Second,
			errs:      []error{filtered},
			wantCalls: 1,
			wantIs:    filtered,
		},
		{
			name:      "a cancellation the call reports is not asked again while the context lives",
			delay:     2 * time.Second,
			errs:      []error{cancelled},
			wantCalls: 1,
			wantIs:    context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calledAt []time.Time
			got, err := retryTransient(context.Background(), tt.delay, func() (string, error) {
				calledAt = append(calledAt, time.Now())
				if err := tt.errs[min(len(calledAt), len(tt.errs))-1]; err != nil {
					return "", err
				}
				return "ok", nil
			})
			returnedAt := time.Now()

			assert.Equal(t, tt.want, got)
			assert.Len(t, calledAt, tt.wantCalls)
			if tt.wantIs == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantIs)
				assert.ErrorContains(t, err, tt.wantMsg)
			}

			for i := 1; i < len(calledAt); i++ {
				assert.GreaterOrEqual(t, calledAt[i].Sub(calledAt[i-1]), tt.delay,
					"attempt %d must wait out the backoff", i+1)
			}
			assert.Less(t, returnedAt.Sub(calledAt[len(calledAt)-1]), tt.delay,
				"nothing is left to wait for after the last attempt")
		})
	}
}

type retryCallExProvider struct {
	provider.Provider

	calls int
	err   error
}

func (f *retryCallExProvider) CallEx(
	context.Context, pconfig.ProviderOptionsType, []llms.MessageContent, streaming.Callback,
) (*llms.ContentResponse, error) {
	f.calls++

	return nil, f.err
}

func TestRetry_RetryTransient_SimpleChainNeverAsksAgainBeforeTheBackoff(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		cancelFirst bool
		cancelAfter time.Duration
		wantIs      error
		wantMsg     string
	}{
		{
			name:        "a gateway failure waits, and a cancellation during the wait returns at once",
			err:         errors.New("API returned unexpected status code: 502: bad gateway"),
			cancelAfter: 200 * time.Millisecond,
			wantIs:      context.Canceled,
		},
		{
			name:    "a rejected request is not asked again",
			err:     errors.New("API returned unexpected status code: 400: bad request"),
			wantMsg: "400: bad request",
		},
		{
			name:        "a cancelled context stops the loop",
			err:         context.Canceled,
			cancelFirst: true,
			wantIs:      context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prv := &retryCallExProvider{err: tt.err}
			fp := newFlowProvider()
			fp.Provider = prv

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancelFirst {
				cancel()
			}
			if tt.cancelAfter > 0 {
				time.AfterFunc(tt.cancelAfter, cancel)
			}

			start := time.Now()
			_, err := fp.performSimpleChain(ctx, nil, nil, pconfig.OptionsTypeSimple, "", "system", "user")
			elapsed := time.Since(start)

			require.Error(t, err)
			if tt.wantIs != nil {
				assert.ErrorIs(t, err, tt.wantIs)
			}
			assert.ErrorContains(t, err, tt.wantMsg)
			assert.Equal(t, 1, prv.calls, "the model must not be asked again before the backoff")
			assert.Less(t, elapsed, delayBetweenRetries, "returning must not serve the whole backoff")
		})
	}
}

func TestRetry_RetryTransient_SetupCallIsNotAskedAgainAfterALocalRefusal(t *testing.T) {
	prv := &fakeCallProvider{
		failTimes: 3,
		err:       &llms.ErrStructuredOutputUnsupported{Model: "claude-fable-5-1"},
	}

	_, err := callWithSetupRetries(context.Background(), prv, pconfig.OptionsTypeSimple, "prompt")

	require.Error(t, err)
	assert.ErrorAs(t, err, new(*llms.ErrStructuredOutputUnsupported))
	assert.Equal(t, 1, prv.callCount, "a refusal the library decided locally must not be asked again")
}
