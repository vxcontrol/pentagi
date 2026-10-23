package searchers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"

	customsearch "google.golang.org/api/customsearch/v1"
)

func TestGoogle_Handle_SendsTheKeysAndClampsTheResultCount(t *testing.T) {
	up := &searchersUpstream{body: `{"items":[{"title":"Doc","link":"https://example.com","snippet":"snip"}]}`}
	p := newTestProxy(t, "customsearch.googleapis.com", up)
	g := NewGoogle(&config.Config{
		GoogleAPIKey:      "test-key",
		GoogleCXKey:       "test-cx",
		GoogleLRKey:       "lang_en",
		ProxyURL:          p.URL(),
		ExternalSSLCAPath: p.CACertPath(),
	})
	if g.Engine() != database.SearchengineTypeGoogle {
		t.Errorf("Engine() = %q, want %q", g.Engine(), database.SearchengineTypeGoogle)
	}

	for _, tt := range []struct {
		name       string
		maxResults int
		wantNum    string
	}{
		{"a limit below ten is kept", 5, "5"},
		{"a limit of ten is kept", 10, "10"},
		{"a larger limit is cut to ten", 100, "10"},
		{"no limit means ten", 0, "10"},
		{"a negative limit means ten", -5, "10"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.Handle(t.Context(), Request{Query: "hello world", MaxResults: tt.maxResults})
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			if up.path != "/customsearch/v1" {
				t.Errorf("request path = %q, want /customsearch/v1", up.path)
			}
			// The proxy client replaces the transport that would carry option.WithAPIKey.
			for param, want := range map[string]string{
				"key": "test-key", "cx": "test-cx", "lr": "lang_en", "q": "hello world", "num": tt.wantNum,
			} {
				if got := up.query.Get(param); got != want {
					t.Errorf("query param %s = %q, want %q", param, got, want)
				}
			}
			if !strings.Contains(got, "# 1. Doc\n\n## URL\nhttps://example.com\n\n## Snippet\n\nsnip") {
				t.Errorf("Handle() result misses the item: %q", got)
			}
		})
	}
}

func TestGoogle_ClassifyGoogleError_TellsATransportFailureFromAnAPIRefusal(t *testing.T) {
	t.Run("a request that never got an answer is retryable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := NewGoogle(&config.Config{GoogleAPIKey: "test-key", GoogleCXKey: "test-cx"}).
			Handle(ctx, Request{Query: "q", MaxResults: 5})

		if got != "" || !IsRetryable(err) || !errors.Is(err, context.Canceled) {
			t.Errorf("Handle() = %q, %v; want no result and a RetryableError around the cancellation", got, err)
		}
	})

	t.Run("an api rejection is fatal", func(t *testing.T) {
		up := &searchersUpstream{status: http.StatusForbidden, body: `{"error":{"code":403,"message":"caller is not authorized"}}`}
		p := newTestProxy(t, "customsearch.googleapis.com", up)

		got, err := NewGoogle(&config.Config{
			GoogleAPIKey:      "test-key",
			GoogleCXKey:       "test-cx",
			ProxyURL:          p.URL(),
			ExternalSSLCAPath: p.CACertPath(),
		}).Handle(t.Context(), Request{Query: "q", MaxResults: 5})

		if got != "" || !IsFatal(err) || !strings.Contains(err.Error(), "HTTP 403") {
			t.Errorf("Handle() = %q, %v; want no result and a FatalError naming HTTP 403", got, err)
		}
	})
}

func TestGoogle_FormatResults_NumbersEachItemVerbatim(t *testing.T) {
	g := &google{}

	for _, tt := range []struct {
		name  string
		items []*customsearch.Result
		want  string
	}{
		{"no items render nothing", nil, ""},
		{
			name: "items are numbered with their link and snippet, special characters kept",
			items: []*customsearch.Result{
				{Title: "Go Programming Language", Link: "https://go.dev", Snippet: "Go is an open source programming language."},
				{Title: `Test & <Special> "Characters"`, Link: "https://example.com/path?q=test&lang=en", Snippet: `Content with special chars: <, >, &, "quotes"`},
			},
			want: "# 1. Go Programming Language\n\n## URL\nhttps://go.dev\n\n## Snippet\n\nGo is an open source programming language.\n\n" +
				"# 2. Test & <Special> \"Characters\"\n\n## URL\nhttps://example.com/path?q=test&lang=en\n\n## Snippet\n\nContent with special chars: <, >, &, \"quotes\"\n\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.formatResults(&customsearch.Search{Items: tt.items}); got != tt.want {
				t.Errorf("formatResults() = %q, want %q", got, tt.want)
			}
		})
	}
}

// roundTripFunc adapts a function to http.RoundTripper for transport unit tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestGoogle_RoundTrip_AddsTheKeyWithoutTouchingTheCallersRequest(t *testing.T) {
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

// Subtests are keyed by accessor; each guards a nil config on its own.
func TestGoogle_KeyAccessorsAreEmptyForANilConfig(t *testing.T) {
	g := &google{cfg: nil}

	for _, tt := range []struct {
		name string
		key  func() string
	}{
		{"apiKey is empty", g.apiKey},
		{"cxKey is empty", g.cxKey},
		{"lrKey is empty", g.lrKey},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key(); got != "" {
				t.Errorf("got %q with a nil config, want empty", got)
			}
		})
	}
}

func TestGoogle_IsAvailable_RequiresBothKeys(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when both keys are set", &config.Config{GoogleAPIKey: "test-key", GoogleCXKey: "test-cx"}, true},
		{"unavailable when API key is empty", &config.Config{GoogleCXKey: "test-cx"}, false},
		{"unavailable when CX key is empty", &config.Config{GoogleAPIKey: "test-key"}, false},
		{"unavailable when both keys are empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
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

func TestGoogleEndpoint(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"empty uses Google default", "", ""},
		{"whitespace uses Google default", "   ", ""},
		{"base URL gets trailing slash", "https://search.example.com", "https://search.example.com/"},
		{"base URL with trailing slash", "https://search.example.com/", "https://search.example.com/"},
		{"full API URL is trimmed to base", "https://search.example.com/customsearch/v1", "https://search.example.com/"},
		{"full API URL with trailing slash", "https://search.example.com/customsearch/v1/", "https://search.example.com/"},
		{"base URL with path prefix", "https://example.com/cse", "https://example.com/cse/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &google{cfg: &config.Config{GoogleCSEURL: tt.url}}
			if got := g.endpoint(); got != tt.want {
				t.Errorf("endpoint() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("nil config", func(t *testing.T) {
		g := &google{}
		if got := g.endpoint(); got != "" {
			t.Errorf("endpoint() = %q, want empty", got)
		}
	})
}

func TestGoogleNewSearchService_DefaultBasePath(t *testing.T) {
	g := &google{cfg: testGoogleConfig()}

	svc, err := g.newSearchService(t.Context())
	if err != nil {
		t.Fatalf("newSearchService() unexpected error: %v", err)
	}
	if svc.BasePath != "https://customsearch.googleapis.com/" {
		t.Errorf("BasePath = %q, want Google's default", svc.BasePath)
	}
}

func TestGoogleHandle_CustomEndpoint(t *testing.T) {
	var gotPath, gotKey, gotCX, gotQ string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.URL.Query().Get("key")
		gotCX = r.URL.Query().Get("cx")
		gotQ = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"title":"Compat Doc","link":"https://compat.example.com","snippet":"from compatible endpoint"}]}`))
	}))
	defer srv.Close()

	// Both spellings of the setting should reach the same path.
	for _, setting := range []string{srv.URL, srv.URL + "/customsearch/v1"} {
		t.Run(setting, func(t *testing.T) {
			gotPath, gotKey, gotCX, gotQ = "", "", "", ""
			cfg := testGoogleConfig()
			cfg.GoogleCSEURL = setting
			g := NewGoogle(cfg)

			got, err := g.Handle(t.Context(), Request{Query: "hello world", MaxResults: 3})
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if gotPath != "/customsearch/v1" {
				t.Errorf("request path = %q, want /customsearch/v1", gotPath)
			}
			if gotKey != testGoogleAPIKey || gotCX != testGoogleCXKey || gotQ != "hello world" {
				t.Errorf("request params key=%q cx=%q q=%q, want key/cx from config and the query", gotKey, gotCX, gotQ)
			}
			if !strings.Contains(got, "https://compat.example.com") || !strings.Contains(got, "Compat Doc") {
				t.Errorf("result not parsed from custom endpoint: %q", got)
			}
		})
	}
}
