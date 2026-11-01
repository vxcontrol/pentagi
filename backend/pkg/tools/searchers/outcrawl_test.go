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

func TestOutcrawl_Handle_PostsTheSearchAndReturnsItsOutcome(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "api.outcrawl.ai", up)
	oc := NewOutcrawl(&config.Config{
		OutcrawlAPIKey:    "test-key",
		ProxyURL:          p.URL(),
		ExternalSSLCAPath: p.CACertPath(),
	}, nil)
	if oc.Engine() != database.SearchengineTypeOutcrawl {
		t.Errorf("Engine() = %q, want %q", oc.Engine(), database.SearchengineTypeOutcrawl)
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
			body:   `{"hits":[{"url":"https://example.com","title":"Doc","snippet":"short","rank":1,"markdown":"long markdown content"}],"usage":{"credits":"3"}}`,
		},
		{name: "an upstream 502 is retryable", status: http.StatusBadGateway, wantRetryable: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, tt.body)

			got, err := oc.Handle(t.Context(), Request{Query: "test query", MaxResults: 5})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if up.method != http.MethodPost || up.path != "/v1/search" {
				t.Errorf("request = %s %s, want POST /v1/search", up.method, up.path)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if auth := up.header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Authorization = %q, want Bearer test-key", auth)
			}
			if want := `{"query":"test query","limit":5,"scrape":true,"formats":["markdown"]}`; string(up.payload) != want {
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
			if !strings.Contains(got, "long markdown content") {
				t.Errorf("Handle() result misses the page markdown: %q", got)
			}
		})
	}
}

func TestOutcrawl_ParseHTTPResponse_ClassifiesEachStatus(t *testing.T) {
	oc := &outcrawl{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		errContain string // empty for a success
		retryable  bool
	}{
		{"successful response", http.StatusOK, `{"hits":[{"title":"A","snippet":"c","url":"https://a.com","rank":1}]}`, "", false},
		{"decode error", http.StatusOK, "{invalid json", "failed to decode response body", false},
		{"bad request", http.StatusBadRequest, "", "invalid", false},
		{"unauthorized keeps the API message", http.StatusUnauthorized,
			`{"code":"unauthorized","message":"authentication failed: no api key presented"}`, "no api key presented", false},
		{"payment required", http.StatusPaymentRequired, "", "credits", false},
		{"forbidden", http.StatusForbidden, "", "not allowed", false},
		{"not found", http.StatusNotFound, "", "could not be found", false},
		{"request timeout", http.StatusRequestTimeout, "", "timed out", true},
		{"too many requests", http.StatusTooManyRequests, "", "too many", true},
		{"internal server error", http.StatusInternalServerError, "", "server", true},
		{"bad gateway", http.StatusBadGateway, "", "server", true},
		{"service unavailable", http.StatusServiceUnavailable, "", "unavailable", true},
		{"gateway timeout", http.StatusGatewayTimeout, "", "unavailable", true},
		{"unknown status code", http.StatusTeapot, "", "unexpected status code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			result, err := oc.parseHTTPResponse(t.Context(), "test query", resp)

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

func TestOutcrawl_BuildOutcrawlResult_SummarizesOrInlinesTheMarkdown(t *testing.T) {
	hits := []outcrawlHit{
		{Title: "No markdown", Snippet: "first", URL: "https://first.example.com", Rank: 1},
		{Title: "Title", Snippet: "content", URL: "https://example.com", Rank: 2, Markdown: "very long markdown content"},
	}

	t.Run("uses summarizer when markdown content exists", func(t *testing.T) {
		oc := &outcrawl{
			summarizer: func(ctx context.Context, prompt string) (string, error) {
				if !strings.Contains(prompt, `USER QUERY: "test query"`) {
					t.Fatalf("summarizer prompt must include the user query, got: %q", prompt)
				}
				// The page with markdown is the second link, so its source id is 2.
				if !strings.Contains(prompt, `<raw_content id="2" title="Title" url="https://example.com">`) {
					t.Fatalf("summarizer prompt must keep one-based ids aligned with the links, got: %q", prompt)
				}
				if strings.Contains(prompt, "first.example.com") {
					t.Fatalf("summarizer prompt must skip results without markdown, got: %q", prompt)
				}
				return "short summary", nil
			},
		}

		out := oc.buildOutcrawlResult(t.Context(), "test query", &outcrawlSearchResult{Hits: hits})

		if !strings.Contains(out, "### Summarized Content\n\nshort summary") {
			t.Errorf("buildOutcrawlResult() missing the summary, got: %q", out)
		}
		if !strings.Contains(out, "## 1. No markdown") || !strings.Contains(out, "## 2. Title") {
			t.Errorf("buildOutcrawlResult() missing numbered links, got: %q", out)
		}
	})

	t.Run("falls back to markdown content when no summarizer", func(t *testing.T) {
		oc := &outcrawl{}

		out := oc.buildOutcrawlResult(t.Context(), "test query", &outcrawlSearchResult{Hits: hits})

		if !strings.Contains(out, "### Raw content for 2. Title\n\nvery long markdown content") {
			t.Errorf("buildOutcrawlResult() missing raw markdown content, got: %q", out)
		}
	})

	t.Run("raw content is bounded per page", func(t *testing.T) {
		oc := &outcrawl{}
		long := strings.Repeat("a", maxRawContentLength+100)

		out := oc.buildOutcrawlResult(t.Context(), "test query", &outcrawlSearchResult{
			Hits: []outcrawlHit{{Title: "Long", URL: "https://long.example.com", Markdown: long}},
		})

		if strings.Contains(out, long) || !strings.Contains(out, long[:maxRawContentLength]) {
			t.Errorf("buildOutcrawlResult() must cut markdown at %d characters", maxRawContentLength)
		}
	})

	t.Run("no content sections when markdown is empty", func(t *testing.T) {
		oc := &outcrawl{}

		out := oc.buildOutcrawlResult(t.Context(), "test query", &outcrawlSearchResult{Hits: hits[:1]})

		if strings.Contains(out, "### Raw content for") || strings.Contains(out, "### Summarized Content") {
			t.Errorf("buildOutcrawlResult() should not have content sections, got: %q", out)
		}
		if !strings.Contains(out, "### Short content\n\nfirst") {
			t.Errorf("buildOutcrawlResult() missing the snippet, got: %q", out)
		}
	})
}

func TestOutcrawl_IsAvailable_RequiresAnAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when API key is set", &config.Config{OutcrawlAPIKey: "test-key"}, true},
		{"unavailable when API key is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oc := &outcrawl{cfg: tt.cfg}
			if got := oc.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
