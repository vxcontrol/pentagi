package tools

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
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
		result := p.formatResponse(context.Background(), resp, "test query")
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
		result := p.formatResponse(context.Background(), resp, "what is Go")
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
		result := p.formatResponse(context.Background(), resp, "test")
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
		result := p.formatResponse(context.Background(), resp, "query")
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
		result := p.formatResponse(context.Background(), resp, "query")
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

// unusedPort returns a TCP address with a port that is guaranteed to be unused.
// It binds to port 0 (OS-assigned), captures the address, then closes the listener.
func unusedPort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find unused port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestPerplexitySearchWithInvalidKeyReturnsError(t *testing.T) {
	// search() sends to the hardcoded perplexityURL const with a fake API key.
	// Depending on network environment:
	//   - Online: API returns 401 -> handleErrorResponse -> "API key is wrong"
	//   - Offline: connection fails -> "failed to send request"
	// Either way, an error must be returned. HTTP status-code-specific error
	// handling is covered separately by TestPerplexityHandleErrorResponse.
	p := &perplexity{
		flowID:  1,
		apiKey:  "test-key-invalid",
		model:   "sonar",
		timeout: 5 * time.Second,
	}

	_, err := p.search(context.Background(), "test query")
	if err == nil {
		t.Fatal("search() with invalid API key should return error")
	}
	// Accept either network error or API rejection.
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to send request") &&
		!strings.Contains(errMsg, "API key") &&
		!strings.Contains(errMsg, "status code") {
		t.Errorf("unexpected error = %q; want network or auth error", errMsg)
	}
}

func TestPerplexitySearchCreatesNewClientWithProxy(t *testing.T) {
	// Verify that search() with a proxy URL does not panic and returns
	// a clean error. The proxy transport is created correctly but the
	// hardcoded perplexityURL is unreachable through the non-existent proxy.
	addr := unusedPort(t)
	p := &perplexity{
		flowID:   1,
		apiKey:   "test-key",
		model:    "sonar",
		timeout:  2 * time.Second,
		proxyURL: "http://" + addr, // guaranteed unused port
	}

	_, err := p.search(context.Background(), "test query")
	if err == nil {
		t.Fatal("search() should return error when proxy is unreachable")
	}
	if !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("error = %q, want to contain 'failed to send request'", err.Error())
	}
}
