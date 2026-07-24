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

const testFirecrawlAPIKey = "test-key"

func testFirecrawlConfig() *config.Config {
	return &config.Config{FirecrawlAPIKey: testFirecrawlAPIKey}
}

func TestFirecrawlHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedContentType string
	var receivedAuthorization string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/v2/search", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedAuthorization = r.Header.Get("Authorization")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"data":{"web":[{"title":"Doc","description":"short","url":"https://example.com","markdown":"long markdown content","metadata":{"title":"Doc","sourceURL":"https://example.com"}}]}}`))
	})

	proxy, err := newTestProxy("api.firecrawl.dev", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	cfg := &config.Config{
		FirecrawlAPIKey:     testFirecrawlAPIKey,
		ProxyURL:            proxy.URL(),
		ExternalSSLCAPath:   proxy.CACertPath(),
		ExternalSSLInsecure: false,
	}

	fc := NewFirecrawl(cfg, nil)

	got, err := fc.Handle(
		t.Context(),
		Request{Query: "test query", MaxResults: 5},
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if fc.Engine() != database.SearchengineTypeFirecrawl {
		t.Errorf("Engine() = %q, want %q", fc.Engine(), database.SearchengineTypeFirecrawl)
	}

	// Verify mock handler was called
	if !seenRequest {
		t.Fatal("request was not intercepted by proxy - mock handler was not called")
	}

	// Verify request was built correctly
	if receivedMethod != http.MethodPost {
		t.Errorf("request method = %q, want POST", receivedMethod)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", receivedContentType)
	}
	if receivedAuthorization != "Bearer "+testFirecrawlAPIKey {
		t.Errorf("Authorization = %q, want %q", receivedAuthorization, "Bearer "+testFirecrawlAPIKey)
	}
	if !strings.Contains(string(receivedBody), `"query":"test query"`) {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"limit":5`) {
		t.Errorf("request body = %q, expected to contain limit", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"markdown"`) {
		t.Errorf("request body = %q, expected to contain markdown format", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Links") {
		t.Errorf("result missing '# Links' section: %q", got)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("result missing expected URL 'https://example.com': %q", got)
	}
}

func TestFirecrawlIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when API key is set",
			cfg:  testFirecrawlConfig(),
			want: true,
		},
		{
			name: "unavailable when API key is empty",
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
			fc := &firecrawl{cfg: tt.cfg}
			if got := fc.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirecrawlSearchURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "default when URL is empty",
			cfg:  &config.Config{},
			want: "https://api.firecrawl.dev/v2/search",
		},
		{
			name: "custom self-hosted URL",
			cfg:  &config.Config{FirecrawlAPIURL: "https://firecrawl.internal:3002"},
			want: "https://firecrawl.internal:3002/v2/search",
		},
		{
			name: "custom URL with trailing slash",
			cfg:  &config.Config{FirecrawlAPIURL: "https://firecrawl.internal/"},
			want: "https://firecrawl.internal/v2/search",
		},
		{
			name: "nil config falls back to default",
			cfg:  nil,
			want: "https://api.firecrawl.dev/v2/search",
		},
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

func TestFirecrawlParseHTTPResponse_StatusAndDecodeErrors(t *testing.T) {
	fc := &firecrawl{}

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			body:       `{"success":true,"data":{"web":[{"title":"A","description":"c","url":"https://a.com"}]}}`,
			wantErr:    false,
		},
		{
			name:       "success false with error message",
			statusCode: http.StatusOK,
			body:       `{"success":false,"error":"Request timed out"}`,
			wantErr:    true,
			errContain: "Request timed out",
		},
		{
			name:       "decode error",
			statusCode: http.StatusOK,
			body:       "{invalid json",
			wantErr:    true,
			errContain: "failed to decode response body",
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       "",
			wantErr:    true,
			errContain: "invalid",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       "",
			wantErr:    true,
			errContain: "API key",
		},
		{
			name:       "payment required",
			statusCode: http.StatusPaymentRequired,
			body:       "",
			wantErr:    true,
			errContain: "credits",
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			body:       "",
			wantErr:    true,
			errContain: "administrators only",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       "",
			wantErr:    true,
			errContain: "could not be found",
		},
		{
			name:       "method not allowed",
			statusCode: http.StatusMethodNotAllowed,
			body:       "",
			wantErr:    true,
			errContain: "invalid method",
		},
		{
			name:       "request timeout",
			statusCode: http.StatusRequestTimeout,
			body:       "",
			wantErr:    true,
			errContain: "timed out",
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			body:       "",
			wantErr:    true,
			errContain: "too many",
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			body:       "",
			wantErr:    true,
			errContain: "server",
		},
		{
			name:       "bad gateway",
			statusCode: http.StatusBadGateway,
			body:       "",
			wantErr:    true,
			errContain: "server",
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       "",
			wantErr:    true,
			errContain: "offline",
		},
		{
			name:       "gateway timeout",
			statusCode: http.StatusGatewayTimeout,
			body:       "",
			wantErr:    true,
			errContain: "offline",
		},
		{
			name:       "unknown status code",
			statusCode: 418,
			body:       "",
			wantErr:    true,
			errContain: "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.body)),
			}
			result, err := fc.parseHTTPResponse(t.Context(), "test query", resp)

			if !tt.wantErr {
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
		})
	}
}

func TestFirecrawlBuildResult_WithSummarizer(t *testing.T) {
	t.Run("uses summarizer when markdown content exists", func(t *testing.T) {
		fc := &firecrawl{
			summarizer: func(ctx context.Context, prompt string) (string, error) {
				if !strings.Contains(prompt, "<raw_content") {
					t.Fatalf("summarizer prompt must include raw content, got: %q", prompt)
				}
				// The user query must be rendered so the model can answer it.
				if !strings.Contains(prompt, `USER QUERY: "test query"`) {
					t.Fatalf("summarizer prompt must include the user query, got: %q", prompt)
				}
				// Source IDs must be one-based to line up with the numbered links.
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

func TestFirecrawlHandle_ReturnsTypedError(t *testing.T) {
	t.Run("upstream 502 returns a retryable error", func(t *testing.T) {
		var seenRequest bool
		mockMux := http.NewServeMux()
		mockMux.HandleFunc("/v2/search", func(w http.ResponseWriter, r *http.Request) {
			seenRequest = true
			w.WriteHeader(http.StatusBadGateway)
		})

		proxy, err := newTestProxy("api.firecrawl.dev", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		fc := &firecrawl{
			cfg: &config.Config{
				FirecrawlAPIKey:     testFirecrawlAPIKey,
				ProxyURL:            proxy.URL(),
				ExternalSSLCAPath:   proxy.CACertPath(),
				ExternalSSLInsecure: false,
			},
		}

		result, err := fc.Handle(
			t.Context(),
			Request{Query: "q", MaxResults: 5},
		)

		// Verify mock handler was called (request was intercepted)
		if !seenRequest {
			t.Error("request was not intercepted by proxy - mock handler was not called")
		}

		// The error is now surfaced (not swallowed) and classified. A 502 is retryable.
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
