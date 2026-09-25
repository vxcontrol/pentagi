package searchers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func TestTraversaal_Handle_PostsTheQueryAndReturnsItsOutcome(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "api-ares.traversaal.ai", up)
	trav := NewTraversaal(&config.Config{TraversaalAPIKey: "test-key", ProxyURL: p.URL(), ExternalSSLCAPath: p.CACertPath()})
	if trav.Engine() != database.SearchengineTypeTraversaal {
		t.Errorf("Engine() = %q, want %q", trav.Engine(), database.SearchengineTypeTraversaal)
	}

	for _, tt := range []struct {
		name          string
		status        int
		body          string
		wantRetryable bool
	}{
		{
			name:   "an answer is returned with numbered links",
			status: http.StatusOK,
			body:   `{"data":{"response_text":"answer text","web_url":["https://a.com","https://b.com"]}}`,
		},
		{name: "an upstream 502 is retryable", status: http.StatusBadGateway, wantRetryable: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, tt.body)

			got, err := trav.Handle(t.Context(), Request{Query: "test query"})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if up.method != http.MethodPost || up.path != "/live/predict" {
				t.Errorf("request = %s %s, want POST /live/predict", up.method, up.path)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if key := up.header.Get("x-api-key"); key != "test-key" {
				t.Errorf("x-api-key = %q, want test-key", key)
			}
			if want := `{"query":"test query"}`; string(up.payload) != want {
				t.Errorf("request body = %s, want %s", up.payload, want)
			}

			if tt.wantRetryable {
				if got != "" || !IsRetryable(err) {
					t.Fatalf("Handle() = %q, %v; want no result and a RetryableError", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if want := "# Answer\n\nanswer text\n\n# Links\n\n1. https://a.com\n2. https://b.com\n"; got != want {
				t.Errorf("Handle() = %q, want %q", got, want)
			}
		})
	}
}

func TestTraversaal_ParseHTTPResponse_ClassifiesTheFailure(t *testing.T) {
	trav := &traversaal{}

	for _, tt := range []struct {
		name       string
		status     int
		body       string
		errContain string
		retryable  bool
	}{
		{"a server error is retryable", http.StatusInternalServerError, "", "unexpected status code: 500", true},
		{"a rejected key is fatal", http.StatusUnauthorized, "", "unexpected status code: 401", false},
		{"an undecodable body is fatal", http.StatusOK, "{invalid json", "failed to decode response body", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body))}

			_, err := trav.parseHTTPResponse(resp)

			if err == nil || !strings.Contains(err.Error(), tt.errContain) {
				t.Fatalf("parseHTTPResponse() error = %v, want it to contain %q", err, tt.errContain)
			}
			if IsRetryable(err) != tt.retryable || IsFatal(err) == tt.retryable {
				t.Errorf("parseHTTPResponse() error = %T, want retryable=%v", err, tt.retryable)
			}
		})
	}
}

func TestTraversaal_IsAvailable_RequiresAnAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when API key is set", &config.Config{TraversaalAPIKey: "test-key"}, true},
		{"unavailable when API key is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trav := &traversaal{cfg: tt.cfg}
			if got := trav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
