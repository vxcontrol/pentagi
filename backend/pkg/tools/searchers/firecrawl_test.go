package searchers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func TestFirecrawl_Handle_PostsTheSearchAndReturnsItsOutcome(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "api.firecrawl.dev", up)
	fc := NewFirecrawl(&config.Config{
		FirecrawlAPIKey:   "test-key",
		ProxyURL:          p.URL(),
		ExternalSSLCAPath: p.CACertPath(),
	}, nil)
	if fc.Engine() != database.SearchengineTypeFirecrawl {
		t.Errorf("Engine() = %q, want %q", fc.Engine(), database.SearchengineTypeFirecrawl)
	}

	for _, tt := range []struct {
		name          string
		status        int
		body          string
		wantRetryable bool
	}{
		{
			name:   "a found page is listed with its link",
			status: http.StatusOK,
			body:   `{"success":true,"data":{"web":[{"title":"Doc","description":"short","url":"https://example.com","markdown":"long markdown content","metadata":{"title":"Doc","sourceURL":"https://example.com"}}]}}`,
		},
		{name: "an upstream 502 is retryable", status: http.StatusBadGateway, wantRetryable: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, tt.body)

			got, err := fc.Handle(t.Context(), Request{Query: "test query", MaxResults: 5})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if up.method != http.MethodPost || up.path != "/v2/search" {
				t.Errorf("request = %s %s, want POST /v2/search", up.method, up.path)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if auth := up.header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Authorization = %q, want Bearer test-key", auth)
			}
			if want := `{"query":"test query","limit":5,"scrapeOptions":{"formats":["markdown"],"onlyMainContent":true}}`; string(up.payload) != want {
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
			if !strings.Contains(got, "# Links") || !strings.Contains(got, "https://example.com") {
				t.Errorf("Handle() result misses the links section or the URL: %q", got)
			}
		})
	}
}

func TestFirecrawl_ParseHTTPResponse_ClassifiesEachStatus(t *testing.T) {
	fc := &firecrawl{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		errContain string // empty for a success
		retryable  bool
	}{
		{"successful response", http.StatusOK, `{"success":true,"data":{"web":[{"title":"A","description":"c","url":"https://a.com"}]}}`, "", false},
		{"success false with error message", http.StatusOK, `{"success":false,"error":"Request timed out"}`, "Request timed out", false},
		{"decode error", http.StatusOK, "{invalid json", "failed to decode response body", false},
		{"bad request", http.StatusBadRequest, "", "invalid", false},
		{"unauthorized", http.StatusUnauthorized, "", "API key", false},
		{"payment required", http.StatusPaymentRequired, "", "credits", false},
		{"forbidden", http.StatusForbidden, "", "administrators only", false},
		{"not found", http.StatusNotFound, "", "could not be found", false},
		{"method not allowed", http.StatusMethodNotAllowed, "", "invalid method", false},
		{"request timeout", http.StatusRequestTimeout, "", "timed out", true},
		{"too many requests", http.StatusTooManyRequests, "", "too many", true},
		{"internal server error", http.StatusInternalServerError, "", "server", true},
		{"bad gateway", http.StatusBadGateway, "", "server", true},
		{"service unavailable", http.StatusServiceUnavailable, "", "offline", true},
		{"gateway timeout", http.StatusGatewayTimeout, "", "offline", true},
		{"unknown status code", http.StatusTeapot, "", "unexpected status code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			result, err := fc.parseHTTPResponse(t.Context(), "test query", resp)

			if tt.errContain == "" {
				if err != nil {
					t.Errorf("parseHTTPResponse() unexpected error: %v", err)
				}
				if !strings.Contains(result, "# Links") {
					t.Errorf("parseHTTPResponse() result missing '# Links': %q", result)
				}
				return
			}

			if err == nil {
				t.Fatal("parseHTTPResponse() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("parseHTTPResponse() error = %q, want to contain %q", err.Error(), tt.errContain)
			}
			if IsRetryable(err) != tt.retryable || IsFatal(err) == tt.retryable {
				t.Errorf("parseHTTPResponse() error = %T, want retryable=%v", err, tt.retryable)
			}
		})
	}
}

func TestFirecrawl_BuildFirecrawlResult_SummarizesOrInlinesTheMarkdown(t *testing.T) {
	t.Run("uses summarizer when markdown content exists", func(t *testing.T) {
		fc := &firecrawl{
			summarizer: func(ctx context.Context, prompt string) (string, error) {
				if !strings.Contains(prompt, "<raw_content") {
					t.Fatalf("summarizer prompt must include raw content, got: %q", prompt)
				}
				if !strings.Contains(prompt, `USER QUERY: "test query"`) {
					t.Fatalf("summarizer prompt must include the user query, got: %q", prompt)
				}
				// One-based ids line up with the numbered links.
				if !strings.Contains(prompt, `id="1"`) {
					t.Fatalf("summarizer prompt must use one-based source ids, got: %q", prompt)
				}
				if !strings.Contains(prompt, `title="Title"`) || !strings.Contains(prompt, `url="https://example.com"`) {
					t.Fatalf("summarizer prompt must carry resolved title/url, got: %q", prompt)
				}
				return "short summary", nil
			},
		}

		out := fc.buildFirecrawlResult(t.Context(), "test query", &firecrawlSearchResult{
			Success: true,
			Data: firecrawlData{
				Web: []firecrawlResult{
					{
						Title:       "Title",
						Description: "content",
						URL:         "https://example.com",
						Markdown:    "very long markdown content",
					},
				},
			},
		})

		if !strings.Contains(out, "### Summarized Content") {
			t.Errorf("buildFirecrawlResult() missing '### Summarized Content', got: %q", out)
		}
		if !strings.Contains(out, "short summary") {
			t.Errorf("buildFirecrawlResult() missing 'short summary', got: %q", out)
		}
	})

	t.Run("falls back to markdown content when no summarizer", func(t *testing.T) {
		fc := &firecrawl{}

		out := fc.buildFirecrawlResult(t.Context(), "test query", &firecrawlSearchResult{
			Success: true,
			Data: firecrawlData{
				Web: []firecrawlResult{
					{
						Title:       "Title",
						Description: "content",
						URL:         "https://example.com",
						Markdown:    "very long markdown content",
					},
				},
			},
		})

		if !strings.Contains(out, "### Raw content for") {
			t.Errorf("buildFirecrawlResult() missing '### Raw content for', got: %q", out)
		}
		if !strings.Contains(out, "very long markdown content") {
			t.Errorf("buildFirecrawlResult() missing markdown content, got: %q", out)
		}
	})

	t.Run("no content sections when markdown is empty", func(t *testing.T) {
		fc := &firecrawl{}

		out := fc.buildFirecrawlResult(t.Context(), "test query", &firecrawlSearchResult{
			Success: true,
			Data: firecrawlData{
				Web: []firecrawlResult{
					{
						Title:       "Title",
						Description: "content",
						URL:         "https://example.com",
					},
				},
			},
		})

		if strings.Contains(out, "### Raw content for") {
			t.Errorf("buildFirecrawlResult() should not have raw content section, got: %q", out)
		}
		if strings.Contains(out, "### Summarized Content") {
			t.Errorf("buildFirecrawlResult() should not have summarized content section, got: %q", out)
		}
	})

	t.Run("falls back to metadata url and title", func(t *testing.T) {
		fc := &firecrawl{}

		out := fc.buildFirecrawlResult(t.Context(), "test query", &firecrawlSearchResult{
			Success: true,
			Data: firecrawlData{
				Web: []firecrawlResult{
					{
						Description: "content",
						Metadata: &firecrawlMetadata{
							Title:     "Meta Title",
							SourceURL: "https://meta.example.com",
						},
					},
				},
			},
		})

		if !strings.Contains(out, "Meta Title") {
			t.Errorf("buildFirecrawlResult() missing metadata title, got: %q", out)
		}
		if !strings.Contains(out, "https://meta.example.com") {
			t.Errorf("buildFirecrawlResult() missing metadata source URL, got: %q", out)
		}
	})
}

func TestFirecrawl_SearchURL_DefaultsAndTrimsTheBaseURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{"default when URL is empty", &config.Config{}, "https://api.firecrawl.dev/v2/search"},
		{"custom self-hosted URL", &config.Config{FirecrawlAPIURL: "https://firecrawl.internal:3002"}, "https://firecrawl.internal:3002/v2/search"},
		{"custom URL with trailing slash", &config.Config{FirecrawlAPIURL: "https://firecrawl.internal/"}, "https://firecrawl.internal/v2/search"},
		{"nil config falls back to default", nil, "https://api.firecrawl.dev/v2/search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &firecrawl{cfg: tt.cfg}
			if got := fc.searchURL(); got != tt.want {
				t.Errorf("searchURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirecrawl_IsAvailable_RequiresAnAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when API key is set", &config.Config{FirecrawlAPIKey: "test-key"}, true},
		{"unavailable when API key is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &firecrawl{cfg: tt.cfg}
			if got := fc.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
