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

const testSearxngURL = "http://searxng.example.com"

func testSearxngConfig() *config.Config {
	return &config.Config{
		SearxngURL:        testSearxngURL,
		SearxngLanguage:   "en",
		SearxngCategories: "general",
		SearxngSafeSearch: "0",
		SearxngTimeRange:  "",
		SearxngTimeout:    30,
	}
}

func TestSearxngHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedUserAgent string
	var receivedQuery string
	var receivedFormat string
	var receivedLanguage string
	var receivedCategories string
	var receivedLimit string

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedUserAgent = r.Header.Get("User-Agent")

		query := r.URL.Query()
		receivedQuery = query.Get("q")
		receivedFormat = query.Get("format")
		receivedLanguage = query.Get("language")
		receivedCategories = query.Get("categories")
		receivedLimit = query.Get("limit")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query":"test query","results":[{"title":"Test Result","url":"https://example.com","content":"Test content","engine":"google"}]}`))
	})

	proxy, err := newTestProxy("searxng.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		SearxngURL:        testSearxngURL,
		SearxngLanguage:   "en",
		SearxngCategories: "general",
		SearxngSafeSearch: "0",
		SearxngTimeout:    30,
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
	}

	sx := NewSearxngTool(cfg, flowID, &taskID, &subtaskID, slp, nil)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := sx.Handle(
		ctx,
		SearxngToolName,
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
	if receivedMethod != http.MethodGet {
		t.Errorf("request method = %q, want GET", receivedMethod)
	}
	if receivedUserAgent != "PentAGI/1.0" {
		t.Errorf("User-Agent = %q, want PentAGI/1.0", receivedUserAgent)
	}
	if receivedQuery != "test query" {
		t.Errorf("query param q = %q, want %q", receivedQuery, "test query")
	}
	if receivedFormat != "json" {
		t.Errorf("query param format = %q, want json", receivedFormat)
	}
	if receivedLanguage != "en" {
		t.Errorf("query param language = %q, want en", receivedLanguage)
	}
	if receivedCategories != "general" {
		t.Errorf("query param categories = %q, want general", receivedCategories)
	}
	if receivedLimit != "5" {
		t.Errorf("query param limit = %q, want 5", receivedLimit)
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Searxng Search Results") {
		t.Errorf("result missing '# Searxng Search Results' section: %q", got)
	}
	if !strings.Contains(got, "Test Result") {
		t.Errorf("result missing expected text 'Test Result': %q", got)
	}
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("result missing expected URL 'https://example.com': %q", got)
	}
	if !strings.Contains(got, "Test content") {
		t.Errorf("result missing expected content 'Test content': %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeSearxng {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeSearxng)
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

func TestSearxngIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when URL is set",
			cfg:  testSearxngConfig(),
			want: true,
		},
		{
			name: "unavailable when URL is empty",
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
			sx := &searxng{cfg: tt.cfg}
			if got := sx.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearxngParseHTTPResponse_StatusAndDecodeErrors(t *testing.T) {
	sx := &searxng{flowID: 1}

	t.Run("status error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("")),
		}
		_, err := sx.parseHTTPResponse(resp, "test query")
		if err == nil || !strings.Contains(err.Error(), "unexpected status code") {
			t.Fatalf("expected status code error, got: %v", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{invalid json")),
		}
		_, err := sx.parseHTTPResponse(resp, "test query")
		if err == nil || !strings.Contains(err.Error(), "failed to decode response body") {
			t.Fatalf("expected decode error, got: %v", err)
		}
	})

	t.Run("successful response", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"query":"test","results":[{"title":"Title","url":"https://example.com","content":"Content"}]}`)),
		}
		result, err := sx.parseHTTPResponse(resp, "test")
		if err != nil {
			t.Fatalf("parseHTTPResponse() unexpected error: %v", err)
		}
		if !strings.Contains(result, "# Searxng Search Results") {
			t.Errorf("result missing header: %q", result)
		}
		if !strings.Contains(result, "Title") {
			t.Errorf("result missing title: %q", result)
		}
	})
}

func TestSearxngFormatResults_NoResults(t *testing.T) {
	sx := &searxng{flowID: 1}

	result := sx.formatResults([]SearxngResult{}, "test query")
	if !strings.Contains(result, "No Results Found") {
		t.Errorf("result missing 'No Results Found': %q", result)
	}
	if !strings.Contains(result, "test query") {
		t.Errorf("result missing query: %q", result)
	}
}

func TestSearxngFormatResults_WithResults(t *testing.T) {
	sx := &searxng{flowID: 1}

	results := []SearxngResult{
		{
			Title:         "Test Title",
			URL:           "https://example.com",
			Content:       "Test content",
			Author:        "Test Author",
			PublishedDate: "2024-01-01",
			Engine:        "google",
		},
	}

	result := sx.formatResults(results, "test query")

	if !strings.Contains(result, "# Searxng Search Results") {
		t.Errorf("result missing header: %q", result)
	}
	if !strings.Contains(result, "Test Title") {
		t.Errorf("result missing title: %q", result)
	}
	if !strings.Contains(result, "https://example.com") {
		t.Errorf("result missing URL: %q", result)
	}
	if !strings.Contains(result, "Test content") {
		t.Errorf("result missing content: %q", result)
	}
	if !strings.Contains(result, "Test Author") {
		t.Errorf("result missing author: %q", result)
	}
	if !strings.Contains(result, "2024-01-01") {
		t.Errorf("result missing published date: %q", result)
	}
	if !strings.Contains(result, "google") {
		t.Errorf("result missing engine: %q", result)
	}
}

func TestSearxngHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		sx := &searxng{cfg: testSearxngConfig()}
		_, err := sx.Handle(t.Context(), SearxngToolName, []byte("{"))
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

		proxy, err := newTestProxy("searxng.example.com", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		sx := &searxng{
			flowID: 1,
			cfg: &config.Config{
				SearxngURL:        testSearxngURL,
				ProxyURL:          proxy.URL(),
				ExternalSSLCAPath: proxy.CACertPath(),
				SearxngTimeout:    30,
			},
		}

		result, err := sx.Handle(
			context.Background(),
			SearxngToolName,
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
		if !strings.Contains(result, "failed to search in searxng") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}

func TestSearxngHandle_DefaultLimit(t *testing.T) {
	var receivedLimit string

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		receivedLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query":"test","results":[]}`))
	})

	proxy, err := newTestProxy("searxng.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	cfg := &config.Config{
		SearxngURL:        testSearxngURL,
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
		SearxngTimeout:    30,
	}

	sx := NewSearxngTool(cfg, 1, nil, nil, nil, nil)

	_, err = sx.Handle(
		t.Context(),
		SearxngToolName,
		[]byte(`{"query":"test","message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	if receivedLimit != "10" {
		t.Errorf("default limit = %q, want 10", receivedLimit)
	}
}

func TestSearxngHandle_TimeRange(t *testing.T) {
	var receivedTimeRange string

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		receivedTimeRange = r.URL.Query().Get("time_range")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query":"test","results":[]}`))
	})

	proxy, err := newTestProxy("searxng.example.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	cfg := &config.Config{
		SearxngURL:        testSearxngURL,
		SearxngTimeRange:  "day",
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
		SearxngTimeout:    30,
	}

	sx := NewSearxngTool(cfg, 1, nil, nil, nil, nil)

	_, err = sx.Handle(
		t.Context(),
		SearxngToolName,
		[]byte(`{"query":"test","max_results":5,"message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	if receivedTimeRange != "day" {
		t.Errorf("time_range = %q, want day", receivedTimeRange)
	}
}
