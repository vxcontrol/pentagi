package searchers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestErrors_ClassifyHTTPStatus_RetriesRateLimitsAndServerErrorsOnly(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		wantRetry bool
		wantFatal bool
	}{
		{name: "429 is retryable", status: http.StatusTooManyRequests, wantRetry: true},
		{name: "500 is retryable", status: http.StatusInternalServerError, wantRetry: true},
		{name: "502 is retryable", status: http.StatusBadGateway, wantRetry: true},
		{name: "503 is retryable", status: http.StatusServiceUnavailable, wantRetry: true},
		{name: "504 is retryable", status: http.StatusGatewayTimeout, wantRetry: true},
		{name: "400 is fatal", status: http.StatusBadRequest, wantFatal: true},
		{name: "401 is fatal", status: http.StatusUnauthorized, wantFatal: true},
		{name: "403 is fatal", status: http.StatusForbidden, wantFatal: true},
		{name: "404 is fatal", status: http.StatusNotFound, wantFatal: true},
		{name: "418 is fatal", status: http.StatusTeapot, wantFatal: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ClassifyHTTPStatus(tt.status, "boom")
			if err == nil {
				t.Fatal("ClassifyHTTPStatus returned nil")
			}
			if IsRetryable(err) != tt.wantRetry {
				t.Errorf("IsRetryable(%d) = %v, want %v", tt.status, IsRetryable(err), tt.wantRetry)
			}
			if IsFatal(err) != tt.wantFatal {
				t.Errorf("IsFatal(%d) = %v, want %v", tt.status, IsFatal(err), tt.wantFatal)
			}
		})
	}
}

func TestErrors_Retryable_KeepsItsKindCauseAndDelay(t *testing.T) {
	base := errors.New("root cause")

	r := Retryable(fmt.Errorf("wrap: %w", base), 5*time.Second)
	if !IsRetryable(r) {
		t.Error("Retryable() value should satisfy IsRetryable")
	}
	if IsFatal(r) {
		t.Error("Retryable() value should not satisfy IsFatal")
	}
	if !errors.Is(r, base) {
		t.Error("Retryable() should preserve the wrapped chain for errors.Is")
	}
	var re *RetryableError
	if !errors.As(r, &re) || re.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter not preserved, got %+v", re)
	}
}

func TestErrors_Fatal_KeepsItsKindAndCause(t *testing.T) {
	base := errors.New("root cause")

	f := Fatal(fmt.Errorf("wrap: %w", base))
	if !IsFatal(f) {
		t.Error("Fatal() value should satisfy IsFatal")
	}
	if IsRetryable(f) {
		t.Error("Fatal() value should not satisfy IsRetryable")
	}
	if !errors.Is(f, base) {
		t.Error("Fatal() should preserve the wrapped chain for errors.Is")
	}
}

func TestErrors_ErrNotConfigured_EveryEngineRefusesWithoutConfig(t *testing.T) {
	for _, tt := range []struct {
		name   string
		engine Searcher
	}{
		{"duckduckgo refuses", NewDuckDuckGo(nil)},
		{"firecrawl refuses", NewFirecrawl(nil, nil)},
		{"google refuses", NewGoogle(nil)},
		{"the internal engine refuses", NewInternal(nil, nil, nil, nil)},
		{"perplexity refuses", NewPerplexity(nil, nil)},
		{"searxng refuses", NewSearxng(nil, nil)},
		{"sploitus refuses", NewSploitus(nil)},
		{"tavily refuses", NewTavily(nil, nil)},
		{"traversaal refuses", NewTraversaal(nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.engine.Handle(t.Context(), Request{Query: "q", MaxResults: 5})
			if !errors.Is(err, ErrNotConfigured) || got != "" {
				t.Fatalf("Handle() = %q, %v; want no result and ErrNotConfigured", got, err)
			}
			// The orchestrator retries only a RetryableError; an unconfigured engine is passed over at once.
			if IsRetryable(err) || IsFatal(err) {
				t.Errorf("ErrNotConfigured must be neither retryable nor fatal, got %T", err)
			}
		})
	}
}
