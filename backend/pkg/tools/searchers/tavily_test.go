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

func TestTavily_Handle_PostsTheSearchAndReturnsItsOutcome(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "api.tavily.com", up)
	tav := NewTavily(&config.Config{TavilyAPIKey: "test-key", ProxyURL: p.URL(), ExternalSSLCAPath: p.CACertPath()}, nil)
	if tav.Engine() != database.SearchengineTypeTavily {
		t.Errorf("Engine() = %q, want %q", tav.Engine(), database.SearchengineTypeTavily)
	}

	for _, tt := range []struct {
		name          string
		status        int
		body          string
		wantRetryable bool
	}{
		{
			name:   "an answer is returned with its links",
			status: http.StatusOK,
			body:   `{"answer":"final answer","query":"test query","response_time":0.1,"results":[{"title":"Doc","url":"https://example.com","content":"short","raw_content":"long raw content","score":0.9}]}`,
		},
		{name: "an upstream 502 is retryable", status: http.StatusBadGateway, wantRetryable: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, tt.body)

			got, err := tav.Handle(t.Context(), Request{Query: "test query", MaxResults: 5})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if up.method != http.MethodPost || up.path != "/search" {
				t.Errorf("request = %s %s, want POST /search", up.method, up.path)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			for _, want := range []string{`"query":"test query"`, `"api_key":"test-key"`, `"max_results":5`} {
				if !strings.Contains(string(up.payload), want) {
					t.Errorf("request body = %s, want it to contain %s", up.payload, want)
				}
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
			for _, want := range []string{"# Answer", "# Links", "final answer", "https://example.com"} {
				if !strings.Contains(got, want) {
					t.Errorf("result misses %q: %q", want, got)
				}
			}
		})
	}
}

func TestTavily_ParseHTTPResponse_ClassifiesEachStatus(t *testing.T) {
	tav := &tavily{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		errContain string // empty for a success
		retryable  bool
	}{
		{"successful response", http.StatusOK, `{"answer":"ok","query":"q","response_time":0.1,"results":[{"title":"A","url":"https://a.com","content":"c","score":0.3}]}`, "", false},
		{"decode error", http.StatusOK, "{invalid json", "failed to decode response body", false},
		{"bad request", http.StatusBadRequest, "", "invalid", false},
		{"unauthorized", http.StatusUnauthorized, "", "API key", false},
		{"forbidden", http.StatusForbidden, "", "administrators only", false},
		{"not found", http.StatusNotFound, "", "could not be found", false},
		{"method not allowed", http.StatusMethodNotAllowed, "", "invalid method", false},
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
			result, err := tav.parseHTTPResponse(t.Context(), resp)

			if tt.errContain == "" {
				if err != nil {
					t.Errorf("parseHTTPResponse() unexpected error: %v", err)
				}
				if !strings.Contains(result, "# Answer") {
					t.Errorf("parseHTTPResponse() result missing '# Answer': %q", result)
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

func TestTavily_BuildTavilyResult_SummarizesOrInlinesTheRawContent(t *testing.T) {
	t.Run("uses summarizer when raw content exists", func(t *testing.T) {
		tav := &tavily{
			summarizer: func(ctx context.Context, prompt string) (string, error) {
				if !strings.Contains(prompt, "<raw_content") {
					t.Fatalf("summarizer prompt must include raw content, got: %q", prompt)
				}
				if !strings.Contains(prompt, "test query") {
					t.Fatalf("summarizer prompt must include query, got: %q", prompt)
				}
				return "short summary", nil
			},
		}

		raw := "very long raw content"
		out := tav.buildTavilyResult(t.Context(), &tavilySearchResult{
			Answer: "answer",
			Query:  "test query",
			Results: []tavilyResult{
				{
					Title:      "Title",
					URL:        "https://example.com",
					Content:    "content",
					RawContent: &raw,
					Score:      0.5,
				},
			},
		})

		if !strings.Contains(out, "### Summarized Content") {
			t.Errorf("buildTavilyResult() missing '### Summarized Content', got: %q", out)
		}
		if !strings.Contains(out, "short summary") {
			t.Errorf("buildTavilyResult() missing 'short summary', got: %q", out)
		}
	})

	t.Run("falls back to raw content when no summarizer", func(t *testing.T) {
		tav := &tavily{}

		raw := "very long raw content"
		out := tav.buildTavilyResult(t.Context(), &tavilySearchResult{
			Answer: "answer",
			Query:  "test query",
			Results: []tavilyResult{
				{
					Title:      "Title",
					URL:        "https://example.com",
					Content:    "content",
					RawContent: &raw,
					Score:      0.5,
				},
			},
		})

		if !strings.Contains(out, "### Raw content for") {
			t.Errorf("buildTavilyResult() missing '### Raw content for', got: %q", out)
		}
		if !strings.Contains(out, "very long raw content") {
			t.Errorf("buildTavilyResult() missing raw content, got: %q", out)
		}
	})

	t.Run("no raw content sections when raw content is nil", func(t *testing.T) {
		tav := &tavily{}

		out := tav.buildTavilyResult(t.Context(), &tavilySearchResult{
			Answer: "answer",
			Query:  "test query",
			Results: []tavilyResult{
				{
					Title:   "Title",
					URL:     "https://example.com",
					Content: "content",
					Score:   0.5,
				},
			},
		})

		if strings.Contains(out, "### Raw content for") {
			t.Errorf("buildTavilyResult() should not have raw content section, got: %q", out)
		}
		if strings.Contains(out, "### Summarized Content") {
			t.Errorf("buildTavilyResult() should not have summarized content section, got: %q", out)
		}
	})
}

func TestTavily_IsAvailable_RequiresAnAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when API key is set", &config.Config{TavilyAPIKey: "test-key"}, true},
		{"unavailable when API key is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tav := &tavily{cfg: tt.cfg}
			if got := tav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
