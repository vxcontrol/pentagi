package tools

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

const testPerplexityAPIKey = "test-key"

func testPerplexityConfig() *config.Config {
	return &config.Config{
		PerplexityAPIKey:      testPerplexityAPIKey,
		PerplexityModel:       "sonar",
		PerplexityContextSize: "high",
	}
}

func TestPerplexityHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedAuth string
	var receivedContentType string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedAuth = r.Header.Get("Authorization")
		receivedContentType = r.Header.Get("Content-Type")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id":"test-id",
			"model":"sonar",
			"created":1234567890,
			"object":"chat.completion",
			"choices":[{
				"index":0,
				"finish_reason":"stop",
				"message":{"role":"assistant","content":"This is a test answer."}
			}],
			"usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30},
			"citations":["https://example.com","https://test.com"]
		}`))
	})

	proxy, err := newTestProxy("api.perplexity.ai", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		PerplexityAPIKey:      testPerplexityAPIKey,
		PerplexityModel:       "sonar",
		PerplexityContextSize: "high",
		ProxyURL:              proxy.URL(),
		ExternalSSLCAPath:     proxy.CACertPath(),
	}

	px := NewPerplexityTool(cfg, flowID, &taskID, &subtaskID, slp, nil)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := px.Handle(
		ctx,
		PerplexityToolName,
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
	if receivedAuth != "Bearer "+testPerplexityAPIKey {
		t.Errorf("Authorization = %q, want Bearer %s", receivedAuth, testPerplexityAPIKey)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", receivedContentType)
	}
	if !strings.Contains(string(receivedBody), `"model":"sonar"`) {
		t.Errorf("request body = %q, expected to contain model", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"content":"test query"`) {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"search_context_size":"high"`) {
		t.Errorf("request body = %q, expected to contain context size", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Answer") {
		t.Errorf("result missing '# Answer' section: %q", got)
	}
	if !strings.Contains(got, "This is a test answer.") {
		t.Errorf("result missing expected text 'This is a test answer.': %q", got)
	}
	if !strings.Contains(got, "# Citations") {
		t.Errorf("result missing '# Citations' section: %q", got)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("result missing expected citation 'https://example.com': %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypePerplexity {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypePerplexity)
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

func TestPerplexityIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when API key is set",
			cfg:  testPerplexityConfig(),
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
			px := &perplexity{cfg: tt.cfg}
			if got := px.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPerplexityHandleErrorResponse(t *testing.T) {
	px := &perplexity{flowID: 1}

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
			errContain: "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := px.handleErrorResponse(tt.statusCode)
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
	px := &perplexity{flowID: 1}

	t.Run("empty choices returns fallback message", func(t *testing.T) {
		resp := &CompletionResponse{Choices: []Choice{}}
		result := px.formatResponse(t.Context(), resp, "test query")
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
		result := px.formatResponse(t.Context(), resp, "what is Go")
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
		result := px.formatResponse(t.Context(), resp, "test")
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
		result := px.formatResponse(t.Context(), resp, "query")
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
		result := px.formatResponse(t.Context(), resp, "query")
		if strings.Contains(result, "# Citations") {
			t.Error("result should NOT contain citations when slice is empty")
		}
	})
}

func TestPerplexityGetSummarizePrompt(t *testing.T) {
	px := &perplexity{}

	t.Run("prompt without citations", func(t *testing.T) {
		prompt, err := px.getSummarizePrompt("test query", "some content", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "test query") {
			t.Error("prompt should contain the query")
		}
		if !strings.Contains(prompt, "some content") {
			t.Error("prompt should contain the content")
		}
		if strings.Contains(prompt, "</citations>") {
			t.Error("prompt should NOT contain closing </citations> tag when nil")
		}
	})

	t.Run("prompt with citations", func(t *testing.T) {
		citations := []string{"https://a.com", "https://b.com"}
		prompt, err := px.getSummarizePrompt("query", "content", &citations)
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

func TestPerplexityHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		px := &perplexity{cfg: testPerplexityConfig()}
		_, err := px.Handle(t.Context(), PerplexityToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		var seenRequest bool
		mockMux := http.NewServeMux()
		mockMux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
			seenRequest = true
			w.WriteHeader(http.StatusBadGateway)
		})

		proxy, err := newTestProxy("api.perplexity.ai", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		px := &perplexity{
			flowID: 1,
			cfg: &config.Config{
				PerplexityAPIKey:  testPerplexityAPIKey,
				ProxyURL:          proxy.URL(),
				ExternalSSLCAPath: proxy.CACertPath(),
			},
		}

		result, err := px.Handle(
			t.Context(),
			PerplexityToolName,
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
		if !strings.Contains(result, "failed to search in perplexity") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}

func TestPerplexityHandle_StatusCodeErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errContain string
	}{
		{"unauthorized", http.StatusUnauthorized, "API key"},
		{"server error", http.StatusInternalServerError, "server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/chat/completions", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			proxy, err := newTestProxy("api.perplexity.ai", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			px := &perplexity{
				flowID: 1,
				cfg: &config.Config{
					PerplexityAPIKey:  testPerplexityAPIKey,
					ProxyURL:          proxy.URL(),
					ExternalSSLCAPath: proxy.CACertPath(),
				},
			}

			result, err := px.Handle(
				t.Context(),
				PerplexityToolName,
				[]byte(`{"query":"test","max_results":5,"message":"m"}`),
			)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			// Error should be swallowed and returned as string
			if !strings.Contains(result, "failed to search in perplexity") {
				t.Errorf("Handle() = %q, expected swallowed error", result)
			}
			if !strings.Contains(result, tt.errContain) {
				t.Errorf("Handle() = %q, expected to contain %q", result, tt.errContain)
			}
		})
	}
}

func TestPerplexityDefaultValues(t *testing.T) {
	px := &perplexity{cfg: &config.Config{}}

	if px.model() != perplexityModel {
		t.Errorf("default model = %q, want %q", px.model(), perplexityModel)
	}
	if px.temperature() != perplexityTemperature {
		t.Errorf("default temperature = %v, want %v", px.temperature(), perplexityTemperature)
	}
	if px.topP() != perplexityTopP {
		t.Errorf("default topP = %v, want %v", px.topP(), perplexityTopP)
	}
	if px.maxTokens() != perplexityMaxTokens {
		t.Errorf("default maxTokens = %d, want %d", px.maxTokens(), perplexityMaxTokens)
	}
	if px.timeout() != perplexityTimeout {
		t.Errorf("default timeout = %v, want %v", px.timeout(), perplexityTimeout)
	}
}

func TestPerplexityCustomModel(t *testing.T) {
	px := &perplexity{cfg: &config.Config{PerplexityModel: "sonar-pro"}}

	if px.model() != "sonar-pro" {
		t.Errorf("model = %q, want sonar-pro", px.model())
	}
}
