package tools

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
)

func TestPerplexityIsAvailable(t *testing.T) {
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
			p := &perplexity{apiKey: tt.apiKey}
			if got := p.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPerplexityNewDefaults(t *testing.T) {
	tool := NewPerplexityTool(1, nil, nil, "key", "", "", "", 0, 0, 0, 0, nil, nil)
	p, ok := tool.(*perplexity)
	if !ok {
		t.Fatal("NewPerplexityTool did not return *perplexity")
	}

	if p.model != perplexityModel {
		t.Errorf("default model = %q, want %q", p.model, perplexityModel)
	}
	if p.temperature != perplexityTemperature {
		t.Errorf("default temperature = %v, want %v", p.temperature, perplexityTemperature)
	}
	if p.topP != perplexityTopP {
		t.Errorf("default topP = %v, want %v", p.topP, perplexityTopP)
	}
	if p.maxTokens != perplexityMaxTokens {
		t.Errorf("default maxTokens = %d, want %d", p.maxTokens, perplexityMaxTokens)
	}
	if p.timeout != perplexityTimeout {
		t.Errorf("default timeout = %v, want %v", p.timeout, perplexityTimeout)
	}
}

func TestPerplexityNewCustomValues(t *testing.T) {
	tool := NewPerplexityTool(
		1, nil, nil,
		"key", "http://proxy:8080",
		"sonar-pro", "high",
		0.7, 0.8,
		8000, 30*time.Second,
		nil, nil,
	)
	p, ok := tool.(*perplexity)
	if !ok {
		t.Fatal("NewPerplexityTool did not return *perplexity")
	}

	if p.model != "sonar-pro" {
		t.Errorf("model = %q, want %q", p.model, "sonar-pro")
	}
	if p.temperature != 0.7 {
		t.Errorf("temperature = %v, want 0.7", p.temperature)
	}
	if p.topP != 0.8 {
		t.Errorf("topP = %v, want 0.8", p.topP)
	}
	if p.maxTokens != 8000 {
		t.Errorf("maxTokens = %d, want 8000", p.maxTokens)
	}
	if p.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want 30s", p.timeout)
	}
	if p.proxyURL != "http://proxy:8080" {
		t.Errorf("proxyURL = %q, want %q", p.proxyURL, "http://proxy:8080")
	}
}

func TestPerplexityHandleErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errContain string
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			errContain: "invalid",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			errContain: "API key",
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			errContain: "administrators",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			errContain: "not be found",
		},
		{
			name:       "method not allowed",
			statusCode: http.StatusMethodNotAllowed,
			errContain: "invalid method",
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			errContain: "too many",
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			errContain: "server",
		},
		{
			name:       "bad gateway",
			statusCode: http.StatusBadGateway,
			errContain: "server",
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			errContain: "maintenance",
		},
		{
			name:       "gateway timeout",
			statusCode: http.StatusGatewayTimeout,
			errContain: "maintenance",
		},
		{
			name:       "unknown status code",
			statusCode: 418,
			errContain: fmt.Sprintf("%d", 418),
		},
	}

	p := &perplexity{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.handleErrorResponse(tt.statusCode)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContain)
			}
		})
	}
}

func TestPerplexityFormatResponse(t *testing.T) {
	p := &perplexity{flowID: 1}

	t.Run("empty choices returns fallback message", func(t *testing.T) {
		resp := &CompletionResponse{Choices: []Choice{}}
		result := p.formatResponse(t.Context(), resp, "test query")
		if result != "No response received from Perplexity API" {
			t.Errorf("unexpected result for empty choices: %q", result)
		}
	})

	t.Run("single choice without citations", func(t *testing.T) {
		resp := &CompletionResponse{
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "Go is a compiled language."}},
			},
		}
		result := p.formatResponse(t.Context(), resp, "what is Go")
		if !strings.Contains(result, "# Answer") {
			t.Error("result should contain '# Answer' heading")
		}
		if !strings.Contains(result, "Go is a compiled language.") {
			t.Error("result should contain the answer content")
		}
		if strings.Contains(result, "# Citations") {
			t.Error("result should NOT contain citations section when none provided")
		}
	})

	t.Run("single choice with citations", func(t *testing.T) {
		citations := []string{"https://go.dev", "https://example.com/go"}
		resp := &CompletionResponse{
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "Go is fast."}},
			},
			Citations: &citations,
		}
		result := p.formatResponse(t.Context(), resp, "test")
		if !strings.Contains(result, "# Citations") {
			t.Error("result should contain '# Citations' heading")
		}
		if !strings.Contains(result, "1. https://go.dev") {
			t.Error("result should contain numbered citations")
		}
		if !strings.Contains(result, "2. https://example.com/go") {
			t.Error("result should contain second citation")
		}
	})

	t.Run("nil citations pointer", func(t *testing.T) {
		resp := &CompletionResponse{
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "answer"}},
			},
			Citations: nil,
		}
		result := p.formatResponse(t.Context(), resp, "query")
		if strings.Contains(result, "# Citations") {
			t.Error("result should NOT contain citations when pointer is nil")
		}
	})

	t.Run("empty citations slice", func(t *testing.T) {
		emptyCitations := []string{}
		resp := &CompletionResponse{
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "answer"}},
			},
			Citations: &emptyCitations,
		}
		result := p.formatResponse(t.Context(), resp, "query")
		if strings.Contains(result, "# Citations") {
			t.Error("result should NOT contain citations when slice is empty")
		}
	})
}

func TestPerplexityGetSummarizePrompt(t *testing.T) {
	p := &perplexity{}

	t.Run("prompt without citations", func(t *testing.T) {
		prompt, err := p.getSummarizePrompt("test query", "some content", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "test query") {
			t.Error("prompt should contain the query")
		}
		if !strings.Contains(prompt, "some content") {
			t.Error("prompt should contain the content")
		}
		// The template instructions always mention <citations> as a description,
		// but the actual citations data block (</citations>) should be absent.
		if strings.Contains(prompt, "</citations>") {
			t.Error("prompt should NOT contain closing </citations> tag when nil")
		}
	})

	t.Run("prompt with citations", func(t *testing.T) {
		citations := []string{"https://a.com", "https://b.com"}
		prompt, err := p.getSummarizePrompt("query", "content", &citations)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "<citations>") {
			t.Error("prompt should contain citations block")
		}
		if !strings.Contains(prompt, "https://a.com") {
			t.Error("prompt should contain first citation")
		}
	})
}

func readRequestBody(t *testing.T, req *http.Request) string {
	t.Helper()
	b, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	return string(b)
}

func newHTTPJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestPerplexitySearch_RequestBuildAndParse(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	var seenRequest bool
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seenRequest = true
		if req.Method != http.MethodPost {
			t.Fatalf("request method = %s, want POST", req.Method)
		}
		if req.URL.String() != perplexityURL {
			t.Fatalf("request url = %s, want %s", req.URL.String(), perplexityURL)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer test-key")
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		body := readRequestBody(t, req)
		if !strings.Contains(body, `"model":"sonar"`) {
			t.Fatalf("request body = %q, expected model", body)
		}
		if !strings.Contains(body, `"search_context_size":"high"`) {
			t.Fatalf("request body = %q, expected context size", body)
		}
		if !strings.Contains(body, `"content":"test query"`) {
			t.Fatalf("request body = %q, expected query content", body)
		}

		return newHTTPJSONResponse(http.StatusOK, `{
		  "id":"id",
		  "model":"sonar",
		  "created":1,
		  "object":"chat.completion",
		  "choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"answer"}}],
		  "usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2},
		  "citations":["https://a.com"]
		}`), nil
	})

	p := &perplexity{
		flowID:      1,
		apiKey:      "test-key",
		model:       "sonar",
		contextSize: "high",
		timeout:     5 * time.Second,
	}

	got, err := p.search(t.Context(), "test query")
	if err != nil {
		t.Fatalf("search() unexpected error: %v", err)
	}
	if !seenRequest {
		t.Fatal("custom default transport was not used")
	}
	if !strings.Contains(got, "# Answer") || !strings.Contains(got, "https://a.com") {
		t.Fatalf("search() unexpected response format: %q", got)
	}
}

func TestPerplexitySearch_TransportError(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	p := &perplexity{
		flowID:  1,
		apiKey:  "test-key",
		model:   "sonar",
		timeout: 2 * time.Second,
	}

	_, err := p.search(t.Context(), "test query")
	if err == nil {
		t.Fatal("search() should return error on transport failure")
	}
	if !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("error = %q, want to contain 'failed to send request'", err.Error())
	}
}

func TestPerplexityHandle_ValidationAndBehavior(t *testing.T) {
	p := &perplexity{apiKey: "test-key", model: "sonar", timeout: 2 * time.Second}

	t.Run("invalid json", func(t *testing.T) {
		_, err := p.Handle(t.Context(), PerplexityToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		originalTransport := http.DefaultTransport
		defer func() { http.DefaultTransport = originalTransport }()
		http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		})

		got, err := p.Handle(
			t.Context(),
			PerplexityToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}
		if !strings.Contains(got, "failed to search in perplexity") {
			t.Fatalf("Handle() = %q, expected swallowed error message", got)
		}
	})
}

func TestPerplexityHandle_Success_WritesSearchLogWithAgentContext(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return newHTTPJSONResponse(http.StatusOK, `{
		  "id":"id",
		  "model":"sonar",
		  "created":1,
		  "object":"chat.completion",
		  "choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"ok"}}],
		  "usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
		}`), nil
	})

	taskID := int64(101)
	subtaskID := int64(202)
	slp := &searchLogProviderMock{}
	p := &perplexity{
		flowID:    1,
		taskID:    &taskID,
		subtaskID: &subtaskID,
		apiKey:    "test-key",
		model:     "sonar",
		timeout:   2 * time.Second,
		slp:       slp,
	}

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := p.Handle(
		ctx,
		PerplexityToolName,
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
	if slp.engine != database.SearchengineTypePerplexity {
		t.Fatalf("engine = %q, want %q", slp.engine, database.SearchengineTypePerplexity)
	}
}
