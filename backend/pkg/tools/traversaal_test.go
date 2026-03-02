package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTraversaalSearchDoesNotMutateDefaultClient(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	trav := &traversaal{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "http://127.0.0.1:19999", // non-empty to trigger new client path
	}

	// search will fail to connect (expected); the assertion is on global state.
	_, _ = trav.search(context.Background(), "test query")

	if http.DefaultClient.Transport != originalTransport {
		t.Error("http.DefaultClient.Transport was mutated; search must create a new http.Client instead")
	}
}

func TestTraversaalSearchWithoutProxy(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	// Temporarily set a custom transport so we don't hit real network.
	http.DefaultClient.Transport = &http.Transport{}
	defer func() { http.DefaultClient.Transport = originalTransport }()

	trav := &traversaal{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "", // no proxy -- uses DefaultClient as-is
	}

	// Will fail to reach traversaalURL (expected).
	// The key check: DefaultClient.Transport is not replaced by search().
	transportBefore := http.DefaultClient.Transport
	_, _ = trav.search(context.Background(), "test query")

	if http.DefaultClient.Transport != transportBefore {
		t.Error("http.DefaultClient.Transport was changed by search() when no proxy is configured")
	}
}

func TestTraversaalIsAvailable(t *testing.T) {
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
			trav := &traversaal{apiKey: tt.apiKey}
			if got := trav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraversaalParseHTTPResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		wantResult string
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			body:       `{"data":{"response_text":"answer text","web_url":["https://a.com","https://b.com"]}}`,
			wantErr:    false,
			wantResult: "answer text",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			body:       "",
			wantErr:    true,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.body != "" {
					w.Write([]byte(tt.body))
				}
			}))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatalf("failed to get test server response: %v", err)
			}

			trav := &traversaal{flowID: 1}
			result, parseErr := trav.parseHTTPResponse(resp)
			if tt.wantErr {
				if parseErr == nil {
					t.Fatal("parseHTTPResponse() expected error, got nil")
				}
				return
			}
			if parseErr != nil {
				t.Fatalf("parseHTTPResponse() unexpected error: %v", parseErr)
			}
			if !strings.Contains(result, tt.wantResult) {
				t.Errorf("parseHTTPResponse() result = %q, want to contain %q", result, tt.wantResult)
			}
		})
	}
}
