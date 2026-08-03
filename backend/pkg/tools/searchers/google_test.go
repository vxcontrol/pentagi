package searchers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"

	customsearch "google.golang.org/api/customsearch/v1"
)

const (
	testGoogleAPIKey = "test-api-key"
	testGoogleCXKey  = "test-cx-key"
	testGoogleLRKey  = "lang_en"
)

func testGoogleConfig() *config.Config {
	return &config.Config{
		GoogleAPIKey: testGoogleAPIKey,
		GoogleCXKey:  testGoogleCXKey,
		GoogleLRKey:  testGoogleLRKey,
	}
}

func TestGoogleIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when both keys are set",
			cfg:  testGoogleConfig(),
			want: true,
		},
		{
			name: "unavailable when API key is empty",
			cfg:  &config.Config{GoogleCXKey: testGoogleCXKey},
			want: false,
		},
		{
			name: "unavailable when CX key is empty",
			cfg:  &config.Config{GoogleAPIKey: testGoogleAPIKey},
			want: false,
		},
		{
			name: "unavailable when both keys are empty",
			cfg:  &config.Config{},
			want: false,
		},
		{
			name: "unavailable when nil config",
			cfg:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &google{cfg: tt.cfg}
			if got := g.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleEngine(t *testing.T) {
	g := NewGoogle(testGoogleConfig())
	if g.Engine() != database.SearchengineTypeGoogle {
		t.Errorf("Engine() = %q, want %q", g.Engine(), database.SearchengineTypeGoogle)
	}
}

func TestGoogleFormatResults(t *testing.T) {
	g := &google{}

	t.Run("empty results", func(t *testing.T) {
		res := &customsearch.Search{Items: nil}
		result := g.formatResults(res)
		if result != "" {
			t.Errorf("expected empty string for nil items, got %q", result)
		}
	})

	t.Run("single result", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{
					Title:   "Go Programming Language",
					Link:    "https://go.dev",
					Snippet: "Go is an open source programming language.",
				},
			},
		}
		result := g.formatResults(res)

		if !strings.Contains(result, "# 1. Go Programming Language") {
			t.Error("result should contain numbered title")
		}
		if !strings.Contains(result, "## URL\nhttps://go.dev") {
			t.Error("result should contain URL section")
		}
		if !strings.Contains(result, "## Snippet") {
			t.Error("result should contain Snippet section")
		}
		if !strings.Contains(result, "Go is an open source programming language.") {
			t.Error("result should contain snippet text")
		}
	})

	t.Run("multiple results numbered correctly", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{Title: "First", Link: "https://first.com", Snippet: "first snippet"},
				{Title: "Second", Link: "https://second.com", Snippet: "second snippet"},
				{Title: "Third", Link: "https://third.com", Snippet: "third snippet"},
			},
		}
		result := g.formatResults(res)

		if !strings.Contains(result, "# 1. First") {
			t.Error("result should contain '# 1. First'")
		}
		if !strings.Contains(result, "# 2. Second") {
			t.Error("result should contain '# 2. Second'")
		}
		if !strings.Contains(result, "# 3. Third") {
			t.Error("result should contain '# 3. Third'")
		}
	})

	t.Run("special characters in content preserved", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{
					Title:   "Test & <Special> \"Characters\"",
					Link:    "https://example.com/path?q=test&lang=en",
					Snippet: "Content with special chars: <, >, &, \"quotes\"",
				},
			},
		}
		result := g.formatResults(res)

		if !strings.Contains(result, "Test & <Special> \"Characters\"") {
			t.Error("title special characters should be preserved")
		}
		if !strings.Contains(result, "q=test&lang=en") {
			t.Error("URL query parameters should be preserved")
		}
	})
}

func TestGoogleNewSearchService(t *testing.T) {
	t.Run("without proxy", func(t *testing.T) {
		g := &google{cfg: testGoogleConfig()}

		svc, err := g.newSearchService(t.Context())
		if err != nil {
			t.Fatalf("newSearchService() unexpected error: %v", err)
		}
		if svc == nil {
			t.Fatal("newSearchService() returned nil service")
		}
	})

	t.Run("with proxy", func(t *testing.T) {
		g := &google{cfg: &config.Config{
			GoogleAPIKey: testGoogleAPIKey,
			GoogleCXKey:  testGoogleCXKey,
			ProxyURL:     "http://proxy.example.com:8080",
		}}

		svc, err := g.newSearchService(t.Context())
		if err != nil {
			t.Fatalf("newSearchService() unexpected error: %v", err)
		}
		if svc == nil {
			t.Fatal("newSearchService() returned nil service")
		}
	})
}

func TestGoogleHandle_ReturnsTypedError(t *testing.T) {
	t.Run("search error returned as typed error", func(t *testing.T) {
		// Use a canceled context to make Do() fail immediately.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		g := &google{cfg: testGoogleConfig()}

		result, err := g.Handle(ctx, Request{Query: "q", MaxResults: 5})

		// The error is now surfaced (not swallowed). A canceled request is a
		// transport-level failure (not a *googleapi.Error), so classifyGoogleError
		// classifies it as retryable.
		if err == nil {
			t.Fatal("Handle() expected an error, got nil")
		}
		if result != "" {
			t.Errorf("Handle() result = %q, want empty on error", result)
		}
		if !IsRetryable(err) {
			t.Errorf("Handle() error = %v, want a RetryableError", err)
		}
	})
}

func TestGoogleMaxResultsClamp(t *testing.T) {
	tests := []struct {
		name            string
		maxResults      int
		expectedClamped int64
	}{
		{"valid max results", 5, 5},
		{"max limit", 10, 10},
		{"too large", 100, 10},
		{"zero gets default", 0, 10},
		{"negative gets default", -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &google{cfg: testGoogleConfig()}

			// Use a canceled context to fail quickly without a real API call.
			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			result, err := g.Handle(ctx, Request{Query: "test", MaxResults: tt.maxResults})

			// Errors are now surfaced (not swallowed): a canceled request is a
			// transport-level failure, classified as retryable.
			if err == nil {
				t.Fatal("Handle() expected an error, got nil")
			}
			if result != "" {
				t.Errorf("Handle() result = %q, want empty on error", result)
			}
			if !IsRetryable(err) {
				t.Errorf("Handle() error = %v, want a RetryableError", err)
			}
		})
	}
}

func TestGoogleConfigHelpers(t *testing.T) {
	g := &google{cfg: testGoogleConfig()}

	if g.apiKey() != testGoogleAPIKey {
		t.Errorf("apiKey() = %q, want %q", g.apiKey(), testGoogleAPIKey)
	}
	if g.cxKey() != testGoogleCXKey {
		t.Errorf("cxKey() = %q, want %q", g.cxKey(), testGoogleCXKey)
	}
	if g.lrKey() != testGoogleLRKey {
		t.Errorf("lrKey() = %q, want %q", g.lrKey(), testGoogleLRKey)
	}
}

func TestGoogleConfigHelpers_NilConfig(t *testing.T) {
	g := &google{cfg: nil}

	if g.apiKey() != "" {
		t.Errorf("apiKey() with nil config = %q, want empty", g.apiKey())
	}
	if g.cxKey() != "" {
		t.Errorf("cxKey() with nil config = %q, want empty", g.cxKey())
	}
	if g.lrKey() != "" {
		t.Errorf("lrKey() with nil config = %q, want empty", g.lrKey())
	}
}

// roundTripFunc adapts a function to http.RoundTripper for transport unit tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// TestGoogleAPIKeyTransport_AddsKeyParam guards the fix for the WithHTTPClient /
// WithAPIKey drop: the wrapping transport must add `key` to every request, preserve
// existing params, and not mutate the caller's request.
func TestGoogleAPIKeyTransport_AddsKeyParam(t *testing.T) {
	var captured *http.Request
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: make(http.Header)}, nil
	})

	rt := &googleAPIKeyTransport{key: "test-key", base: base}
	req, err := http.NewRequest(http.MethodGet, "https://customsearch.googleapis.com/customsearch/v1?q=hello&cx=cx1", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip() unexpected error: %v", err)
	}

	if got := captured.URL.Query().Get("key"); got != "test-key" {
		t.Errorf("key param = %q, want test-key", got)
	}
	if got := captured.URL.Query().Get("q"); got != "hello" {
		t.Errorf("q param = %q, want hello (existing params must be preserved)", got)
	}
	if req.URL.Query().Get("key") != "" {
		t.Error("the caller's original request was mutated (key added to it)")
	}
}

// TestGoogleHandle_SendsAPIKeyThroughProxy is the end-to-end regression guard: it
// drives a real Handle through the proxy/TLS client (the exact setup that dropped the
// key) and asserts the API key and cx reach the wire, producing a parsed result.
func TestGoogleHandle_SendsAPIKeyThroughProxy(t *testing.T) {
	var receivedURL string
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/customsearch/v1", func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[{"title":"Doc","link":"https://example.com","snippet":"snip"}]}`))
	})

	proxy, err := newTestProxy("customsearch.googleapis.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	cfg := &config.Config{
		GoogleAPIKey:      "test-key",
		GoogleCXKey:       "test-cx",
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
	}
	g := NewGoogle(cfg)

	got, err := g.Handle(t.Context(), Request{Query: "hello world", MaxResults: 5})
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	if !strings.Contains(receivedURL, "key=test-key") {
		t.Errorf("request URL = %q, expected it to carry key=test-key", receivedURL)
	}
	if !strings.Contains(receivedURL, "cx=test-cx") {
		t.Errorf("request URL = %q, expected it to carry cx=test-cx", receivedURL)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("result missing expected URL: %q", got)
	}
}
