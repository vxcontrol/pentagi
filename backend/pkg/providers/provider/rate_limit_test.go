package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/genai"
)

// geminiQuotaError is the error the Gemini free tier returned on 2026-10-14 when a flow
// went over 15 requests per minute; the google.golang.org/genai client returns it by value.
var geminiQuotaError = genai.APIError{
	Code: 429,
	Message: "You exceeded your current quota, please check your plan and billing details. " +
		"For more information on this error, head to: https://ai.google.dev/gemini-api/docs/rate-limits.",
	Status: "RESOURCE_EXHAUSTED",
	Details: []map[string]any{
		{
			"@type": "type.googleapis.com/google.rpc.QuotaFailure",
			"violations": []any{map[string]any{
				"quotaMetric": "generativelanguage.googleapis.com/generate_content_free_tier_requests",
				"quotaId":     "GenerateRequestsPerMinutePerProjectPerModel-FreeTier",
				"quotaValue":  "15",
			}},
		},
		{"@type": "type.googleapis.com/google.rpc.RetryInfo", "retryDelay": "11s"},
	},
}

func TestRateLimit_IsTooManyRequestsError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"Gemini quota error", geminiQuotaError, true},
		{"Gemini quota error by pointer", &geminiQuotaError, true},
		{"Gemini quota error wrapped", fmt.Errorf("failed to get flow title: %w", geminiQuotaError), true},
		{"Gemini quota error as text", errors.New(geminiQuotaError.Error()), true},
		{"Gemini 429 with only the status", genai.APIError{Code: 429}, true},
		{"OpenAI 429", errors.New("API returned unexpected status code: 429: Rate limit reached for gpt-5.4-mini"), true},
		{"Anthropic 429", errors.New("API returned unexpected status code: 429: rate_limit_error: Number of requests has exceeded your rate limit"), true},
		{"Gemini bad request", genai.APIError{Code: 400, Message: "API key not valid", Status: "INVALID_ARGUMENT"}, false},
		{"Gemini server error", genai.APIError{Code: 500, Message: "Internal error", Status: "INTERNAL"}, false},
		{"Gemini overloaded", genai.APIError{Code: 503, Message: "The model is overloaded", Status: "UNAVAILABLE"}, false},
		{"Gemini overloaded as text", errors.New(genai.APIError{Code: 503, Status: "UNAVAILABLE"}.Error()), false},
		{"OpenAI server error", errors.New("API returned unexpected status code: 500: server error"), false},
		{"canceled", context.Canceled, false},
		{"deadline", fmt.Errorf("call: %w", context.DeadlineExceeded), false},
		{"nil", nil, false},
		{"429 inside a token count", errors.New("prompt is 4290 tokens, Error 4291 in tool call_429"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isTooManyRequestsError(tc.err); got != tc.want {
				t.Errorf("isTooManyRequestsError(%v) = %t, want %t", tc.err, got, tc.want)
			}
		})
	}
}
