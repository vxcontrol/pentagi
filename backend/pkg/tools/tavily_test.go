package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/database"
)

func TestTavilySearchDoesNotMutateDefaultClient(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	tav := &tavily{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "http://127.0.0.1:19999", // non-empty to trigger new client path
	}

	// search will fail to connect (expected); the assertion is on global state.
	_, _ = tav.search(t.Context(), "test query", 5)

	if http.DefaultClient.Transport != originalTransport {
		t.Error("http.DefaultClient.Transport was mutated; search must create a new http.Client instead")
	}
}

func TestTavilySearchWithoutProxy(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = originalTransport }()
	var seenRequest bool
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seenRequest = true
		if req.Method != http.MethodPost {
			t.Fatalf("request method = %s, want POST", req.Method)
		}
		if req.URL.String() != tavilyURL {
			t.Fatalf("request url = %s, want %s", req.URL.String(), tavilyURL)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		bodyStr := string(body)
		if !strings.Contains(bodyStr, `"query":"test query"`) {
			t.Fatalf("request body = %q, expected query field", bodyStr)
		}
		if !strings.Contains(bodyStr, `"api_key":"test-key"`) {
			t.Fatalf("request body = %q, expected api key field", bodyStr)
		}
		if !strings.Contains(bodyStr, `"max_results":5`) {
			t.Fatalf("request body = %q, expected max_results field", bodyStr)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"answer":"final answer","query":"test query","response_time":0.1,"results":[{"title":"Doc","url":"https://example.com","content":"short","raw_content":"long raw content","score":0.9}]}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	tav := &tavily{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "", // no proxy -- uses DefaultClient as-is
	}

	got, err := tav.search(t.Context(), "test query", 5)
	if err != nil {
		t.Fatalf("search() unexpected error: %v", err)
	}
	if !seenRequest {
		t.Fatal("custom transport was not used")
	}
	if !strings.Contains(got, "final answer") || !strings.Contains(got, "https://example.com") {
		t.Fatalf("search() unexpected result: %q", got)
	}
}

func TestTavilyIsAvailable(t *testing.T) {
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
			tav := &tavily{apiKey: tt.apiKey}
			if got := tav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTavilyParseHTTPResponse(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		tav := &tavily{}
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"answer":"ok","query":"q","response_time":0.1,"results":[{"title":"A","url":"https://a.com","content":"c","score":0.3}]}`,
			)),
		}
		got, err := tav.parseHTTPResponse(t.Context(), resp)
		if err != nil {
			t.Fatalf("parseHTTPResponse() unexpected error: %v", err)
		}
		if !strings.Contains(got, "# Answer") || !strings.Contains(got, "ok") {
			t.Fatalf("parseHTTPResponse() unexpected result: %q", got)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		tav := &tavily{}
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{invalid json")),
		}
		_, err := tav.parseHTTPResponse(t.Context(), resp)
		if err == nil || !strings.Contains(err.Error(), "failed to decode response body") {
			t.Fatalf("expected decode error, got: %v", err)
		}
	})

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
			errContain: "invalid",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
			errContain: "API key",
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			wantErr:    true,
			errContain: "too many",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
			errContain: "server",
		},
		{
			name:       "unknown status code",
			statusCode: 418,
			wantErr:    true,
			errContain: fmt.Sprintf("%d", 418),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tav := &tavily{flowID: 1}
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader("")),
			}
			_, parseErr := tav.parseHTTPResponse(t.Context(), resp)
			if !tt.wantErr {
				if parseErr != nil {
					t.Errorf("parseHTTPResponse() unexpected error: %v", parseErr)
				}
				return
			}
			if parseErr == nil {
				t.Fatal("parseHTTPResponse() expected error, got nil")
			}
			if !strings.Contains(parseErr.Error(), tt.errContain) {
				t.Errorf("parseHTTPResponse() error = %q, want to contain %q", parseErr.Error(), tt.errContain)
			}
		})
	}
}

func TestTavilyBuildResult_UsesSummarizerWhenRawContentExists(t *testing.T) {
	tav := &tavily{
		summarizer: func(ctx context.Context, result string) (string, error) {
			if !strings.Contains(result, "<raw_content") {
				t.Fatalf("summarizer prompt must include raw content, got: %q", result)
			}
			return "short summary", nil
		},
	}

	raw := "very long raw content"
	out := tav.buildTavilyResult(t.Context(), &tavilySearchResult{
		Answer: "answer",
		Query:  "query",
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

	if !strings.Contains(out, "### Summarized Content") || !strings.Contains(out, "short summary") {
		t.Fatalf("buildTavilyResult() expected summarized content, got: %q", out)
	}
}

func TestTavilyHandle_ValidationAndBehavior(t *testing.T) {
	tav := &tavily{apiKey: "test-key"}

	t.Run("invalid json", func(t *testing.T) {
		_, err := tav.Handle(t.Context(), TavilyToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		originalTransport := http.DefaultClient.Transport
		defer func() { http.DefaultClient.Transport = originalTransport }()
		http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		})

		got, err := tav.Handle(
			t.Context(),
			TavilyToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}
		if !strings.Contains(got, "failed to search in tavily") {
			t.Fatalf("Handle() = %q, expected swallowed error", got)
		}
	})
}

func TestTavilyHandle_Success_WritesSearchLogWithAgentContext(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = originalTransport }()
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"answer":"ok","query":"q","response_time":0.1,"results":[{"title":"A","url":"https://a.com","content":"c","score":0.3}]}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	taskID := int64(11)
	subtaskID := int64(22)
	slp := &searchLogProviderMock{}
	tav := &tavily{
		flowID:    1,
		taskID:    &taskID,
		subtaskID: &subtaskID,
		apiKey:    "test-key",
		slp:       slp,
	}

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := tav.Handle(
		ctx,
		TavilyToolName,
		[]byte(`{"query":"q","max_results":5,"message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if !strings.Contains(got, "ok") {
		t.Fatalf("Handle() result = %q, expected answer", got)
	}
	if slp.calls != 1 {
		t.Fatalf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeTavily {
		t.Fatalf("engine = %q, want %q", slp.engine, database.SearchengineTypeTavily)
	}
}
