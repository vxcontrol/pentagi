package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTavilySearchDoesNotMutateDefaultClient(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	tav := &tavily{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "http://127.0.0.1:19999", // non-empty to trigger new client path
	}

	// search will fail to connect (expected); the assertion is on global state.
	_, _ = tav.search(context.Background(), "test query", 5)

	if http.DefaultClient.Transport != originalTransport {
		t.Error("http.DefaultClient.Transport was mutated; search must create a new http.Client instead")
	}
}

func TestTavilySearchWithoutProxy(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	// Set a custom transport to avoid real network calls.
	http.DefaultClient.Transport = &http.Transport{}
	defer func() { http.DefaultClient.Transport = originalTransport }()

	tav := &tavily{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "", // no proxy -- uses DefaultClient as-is
	}

	// search() will fail to reach tavilyURL (expected).
	// The key check: DefaultClient.Transport is not replaced by search().
	transportBefore := http.DefaultClient.Transport
	_, _ = tav.search(context.Background(), "test query", 5)

	if http.DefaultClient.Transport != transportBefore {
		t.Error("http.DefaultClient.Transport was changed by search() when no proxy is configured")
	}
}

func TestTavilyIsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   bool
	}{
		{
			name:   "available when API key is set",
			apiKey: "test-key",
			want:   true,
		},
		{
			name:   "unavailable when API key is empty",
			apiKey: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tav := &tavily{apiKey: tt.apiKey}
			if got := tav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTavilyParseHTTPResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
			errContain: "invalid",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
			errContain: "API key",
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			wantErr:    true,
			errContain: "too many",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
			errContain: "server",
		},
		{
			name:       "unknown status code",
			statusCode: 418,
			wantErr:    true,
			errContain: fmt.Sprintf("%d", 418),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatalf("failed to get test server response: %v", err)
			}

			tav := &tavily{flowID: 1}
			_, parseErr := tav.parseHTTPResponse(context.Background(), resp)
			if !tt.wantErr {
				if parseErr != nil {
					t.Errorf("parseHTTPResponse() unexpected error: %v", parseErr)
				}
				return
			}
			if parseErr == nil {
				t.Fatal("parseHTTPResponse() expected error, got nil")
			}
			if !strings.Contains(parseErr.Error(), tt.errContain) {
				t.Errorf("parseHTTPResponse() error = %q, want to contain %q", parseErr.Error(), tt.errContain)
			}
		})
	}
}
