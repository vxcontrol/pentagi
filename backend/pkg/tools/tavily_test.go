package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

const testTavilyAPIKey = "test-key"

func testTavilyConfig() *config.Config {
	return &config.Config{TavilyAPIKey: testTavilyAPIKey}
}

func TestTavilyHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedContentType string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"answer":"final answer","query":"test query","response_time":0.1,"results":[{"title":"Doc","url":"https://example.com","content":"short","raw_content":"long raw content","score":0.9}]}`))
	})

	proxy, err := newTestProxy("api.tavily.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		TavilyAPIKey:        testTavilyAPIKey,
		ProxyURL:            proxy.URL(),
		ExternalSSLCAPath:   proxy.CACertPath(),
		ExternalSSLInsecure: false,
	}

	tav := NewTavilyTool(cfg, flowID, &taskID, &subtaskID, slp, nil)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := tav.Handle(
		ctx,
		TavilyToolName,
		[]byte(`{"query":"test query","max_results":5,"message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
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
	if !strings.Contains(string(receivedBody), `"query":"test query"`) {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"api_key":"test-key"`) {
		t.Errorf("request body = %q, expected to contain api_key", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"max_results":5`) {
		t.Errorf("request body = %q, expected to contain max_results", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Answer") {
		t.Errorf("result missing '# Answer' section: %q", got)
	}
	if !strings.Contains(got, "# Links") {
		t.Errorf("result missing '# Links' section: %q", got)
	}
	if !strings.Contains(got, "final answer") {
		t.Errorf("result missing expected text 'final answer': %q", got)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("result missing expected URL 'https://example.com': %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeTavily {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeTavily)
	}
	if slp.query != "test query" {
		t.Errorf("logged query = %q, want %q", slp.query, "test query")
	}
	if slp.parentType != database.MsgchainTypeSearcher {
		t.Errorf("parent agent type = %q, want %q", slp.parentType, database.MsgchainTypeSearcher)
	}
	if slp.currType != database.MsgchainTypeSearcher {
		t.Errorf("current agent type = %q, want %q", slp.currType, database.MsgchainTypeSearcher)
	}
	if slp.taskID == nil || *slp.taskID != taskID {
		t.Errorf("task ID = %v, want %d", slp.taskID, taskID)
	}
	if slp.subtaskID == nil || *slp.subtaskID != subtaskID {
		t.Errorf("subtask ID = %v, want %d", slp.subtaskID, subtaskID)
	}
}

func TestTavilyIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when API key is set",
			cfg:  testTavilyConfig(),
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
			tav := &tavily{cfg: tt.cfg}
			if got := tav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTavilyParseHTTPResponse_StatusAndDecodeErrors(t *testing.T) {
	tav := &tavily{flowID: 1}

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
			body:       `{"answer":"ok","query":"q","response_time":0.1,"results":[{"title":"A","url":"https://a.com","content":"c","score":0.3}]}`,
			wantErr:    false,
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
			result, err := tav.parseHTTPResponse(t.Context(), resp)

			if !tt.wantErr {
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
		})
	}
}

func TestTavilyBuildResult_WithSummarizer(t *testing.T) {
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

func TestTavilyHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		tav := &tavily{cfg: testTavilyConfig()}
		_, err := tav.Handle(t.Context(), TavilyToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		var seenRequest bool
		mockMux := http.NewServeMux()
		mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
			seenRequest = true
			w.WriteHeader(http.StatusBadGateway)
		})

		proxy, err := newTestProxy("api.tavily.com", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		tav := &tavily{
			flowID: 1,
			cfg: &config.Config{
				TavilyAPIKey:        testTavilyAPIKey,
				ProxyURL:            proxy.URL(),
				ExternalSSLCAPath:   proxy.CACertPath(),
				ExternalSSLInsecure: false,
			},
		}

		result, err := tav.Handle(
			t.Context(),
			TavilyToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}

		// Verify mock handler was called (request was intercepted)
		if !seenRequest {
			t.Error("request was not intercepted by proxy - mock handler was not called")
		}

		// Verify error was swallowed and returned as string
		if !strings.Contains(result, "failed to search in tavily") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}
