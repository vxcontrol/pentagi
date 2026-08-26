package searchers

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func TestXquikHandle(t *testing.T) {
	var method string
	var apiKey string
	var query string
	var queryType string
	var limit string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/x/tweets/search", func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		apiKey = r.Header.Get("x-api-key")
		query = r.URL.Query().Get("q")
		queryType = r.URL.Query().Get("queryType")
		limit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tweets": [{
				"id": "123",
				"text": "New external exposure report",
				"createdAt": "2026-08-26T08:00:00Z",
				"likeCount": 12,
				"retweetCount": 3,
				"replyCount": 2,
				"quoteCount": 1,
				"viewCount": 400,
				"author": {"username": "security_team", "name": "Security Team"}
			}],
			"has_next_page": true,
			"next_cursor": "opaque"
		}`))
	})

	proxy, err := newTestProxy("xquik.com", mux)
	if err != nil {
		t.Fatalf("newTestProxy() error: %v", err)
	}
	defer proxy.Close()

	engine := NewXquik(&config.Config{
		XquikAPIKey:       "test-key",
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
	})
	got, err := engine.Handle(t.Context(), Request{Query: `"acme.example" breach`, MaxResults: 8})
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if method != http.MethodGet {
		t.Errorf("method = %q, want GET", method)
	}
	if apiKey != "test-key" {
		t.Errorf("x-api-key = %q, want test-key", apiKey)
	}
	if query != `"acme.example" breach` {
		t.Errorf("q = %q, want exact caller query", query)
	}
	if queryType != "Latest" {
		t.Errorf("queryType = %q, want Latest", queryType)
	}
	if limit != "8" {
		t.Errorf("limit = %q, want 8", limit)
	}
	if engine.Engine() != database.SearchengineTypeXquik {
		t.Errorf("Engine() = %q, want xquik", engine.Engine())
	}

	for _, fragment := range []string{
		"# X search results",
		"Security Team (@security_team)",
		"https://x.com/security_team/status/123",
		`"New external exposure report"`,
		"untrusted external data",
		"did not follow the cursor",
	} {
		if !strings.Contains(got, fragment) {
			t.Errorf("result missing %q: %s", fragment, got)
		}
	}
}

func TestXquikIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{name: "configured", cfg: &config.Config{XquikAPIKey: "key"}, want: true},
		{name: "blank key", cfg: &config.Config{XquikAPIKey: "  "}, want: false},
		{name: "nil config", cfg: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &xquik{cfg: tt.cfg}
			if got := engine.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXquikHandleStatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		retryAfter string
		retryable  bool
		wantDelay  time.Duration
	}{
		{name: "unauthorized", status: http.StatusUnauthorized},
		{name: "rate limited", status: http.StatusTooManyRequests, retryAfter: "3", retryable: true, wantDelay: 3 * time.Second},
		{name: "long retry capped", status: http.StatusTooManyRequests, retryAfter: "3600", retryable: true, wantDelay: xquikMaxRetryAfter},
		{name: "service unavailable", status: http.StatusServiceUnavailable, retryable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v1/x/tweets/search", func(w http.ResponseWriter, _ *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error":"upstream_error","message":"request unavailable"}`))
			})

			proxy, err := newTestProxy("xquik.com", mux)
			if err != nil {
				t.Fatalf("newTestProxy() error: %v", err)
			}
			defer proxy.Close()

			engine := NewXquik(&config.Config{
				XquikAPIKey:       "test-key",
				ProxyURL:          proxy.URL(),
				ExternalSSLCAPath: proxy.CACertPath(),
			})
			got, err := engine.Handle(t.Context(), Request{Query: "test", MaxResults: 1})
			if err == nil {
				t.Fatal("Handle() expected an error")
			}
			if got != "" {
				t.Errorf("Handle() result = %q, want empty", got)
			}
			if IsRetryable(err) != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", IsRetryable(err), tt.retryable)
			}
			if !tt.retryable && !IsFatal(err) {
				t.Errorf("error = %v, want FatalError", err)
			}
			var retryable *RetryableError
			if tt.retryable && errors.As(err, &retryable) && retryable.RetryAfter != tt.wantDelay {
				t.Errorf("RetryAfter = %v, want %v", retryable.RetryAfter, tt.wantDelay)
			}
		})
	}
}

func TestXquikHandleRejectsInvalidResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/x/tweets/search", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tweets":`))
	})

	proxy, err := newTestProxy("xquik.com", mux)
	if err != nil {
		t.Fatalf("newTestProxy() error: %v", err)
	}
	defer proxy.Close()

	engine := NewXquik(&config.Config{
		XquikAPIKey:       "test-key",
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
	})
	_, err = engine.Handle(t.Context(), Request{Query: "test", MaxResults: 1})
	if err == nil || !IsFatal(err) {
		t.Fatalf("Handle() error = %v, want FatalError", err)
	}
}
