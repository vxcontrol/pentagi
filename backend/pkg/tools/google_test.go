package tools

import (
	"context"
	"strings"
	"testing"

	customsearch "google.golang.org/api/customsearch/v1"
)

func TestGoogleIsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		cxKey  string
		want   bool
	}{
		{
			name:   "available when both keys are set",
			apiKey: "test-api-key",
			cxKey:  "test-cx-key",
			want:   true,
		},
		{
			name:   "unavailable when API key is empty",
			apiKey: "",
			cxKey:  "test-cx-key",
			want:   false,
		},
		{
			name:   "unavailable when CX key is empty",
			apiKey: "test-api-key",
			cxKey:  "",
			want:   false,
		},
		{
			name:   "unavailable when both keys are empty",
			apiKey: "",
			cxKey:  "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &google{apiKey: tt.apiKey, cxKey: tt.cxKey}
			if got := g.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleParseSearchResult(t *testing.T) {
	g := &google{flowID: 1}

	t.Run("empty results", func(t *testing.T) {
		res := &customsearch.Search{Items: nil}
		result := g.parseGoogleSearchResult(res)
		if result != "" {
			t.Errorf("expected empty string for nil items, got %q", result)
		}
	})

	t.Run("single result", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{
					Title:   "Go Programming Language",
					Link:    "https://go.dev",
					Snippet: "Go is an open source programming language.",
				},
			},
		}
		result := g.parseGoogleSearchResult(res)

		if !strings.Contains(result, "# 1. Go Programming Language") {
			t.Error("result should contain numbered title")
		}
		if !strings.Contains(result, "## URL\nhttps://go.dev") {
			t.Error("result should contain URL section")
		}
		if !strings.Contains(result, "## Snippet") {
			t.Error("result should contain Snippet section")
		}
		if !strings.Contains(result, "Go is an open source programming language.") {
			t.Error("result should contain snippet text")
		}
	})

	t.Run("multiple results numbered correctly", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{Title: "First", Link: "https://first.com", Snippet: "first snippet"},
				{Title: "Second", Link: "https://second.com", Snippet: "second snippet"},
				{Title: "Third", Link: "https://third.com", Snippet: "third snippet"},
			},
		}
		result := g.parseGoogleSearchResult(res)

		if !strings.Contains(result, "# 1. First") {
			t.Error("result should contain '# 1. First'")
		}
		if !strings.Contains(result, "# 2. Second") {
			t.Error("result should contain '# 2. Second'")
		}
		if !strings.Contains(result, "# 3. Third") {
			t.Error("result should contain '# 3. Third'")
		}
	})

	t.Run("special characters in content preserved", func(t *testing.T) {
		res := &customsearch.Search{
			Items: []*customsearch.Result{
				{
					Title:   "Test & <Special> \"Characters\"",
					Link:    "https://example.com/path?q=test&lang=en",
					Snippet: "Content with special chars: <, >, &, \"quotes\"",
				},
			},
		}
		result := g.parseGoogleSearchResult(res)

		if !strings.Contains(result, "Test & <Special> \"Characters\"") {
			t.Error("title special characters should be preserved")
		}
		if !strings.Contains(result, "q=test&lang=en") {
			t.Error("URL query parameters should be preserved")
		}
	})
}

func TestGoogleNewSearchServiceWithoutProxy(t *testing.T) {
	g := &google{
		apiKey:   "test-api-key",
		cxKey:    "test-cx-key",
		proxyURL: "",
	}

	// newSearchService should succeed with valid API key (even if fake).
	// The Google API client library accepts any string as the API key at
	// construction time; validation happens on actual API calls.
	svc, err := g.newSearchService(t.Context())
	if err != nil {
		t.Fatalf("newSearchService() unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("newSearchService() returned nil service")
	}
}

func TestGoogleNewSearchServiceWithProxy(t *testing.T) {
	g := &google{
		apiKey:   "test-api-key",
		cxKey:    "test-cx-key",
		proxyURL: "http://proxy.example.com:8080",
	}

	// newSearchService constructs opts with the proxy HTTP client, but the
	// current implementation passes a hardcoded option.WithAPIKey(g.apiKey)
	// to customsearch.NewService instead of opts... (see google.go:141).
	// This test verifies the service is created without error; it does NOT
	// verify that the proxy is actually applied to the underlying HTTP client,
	// because that requires an integration test with real network traffic.
	svc, err := g.newSearchService(t.Context())
	if err != nil {
		t.Fatalf("newSearchService() unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("newSearchService() returned nil service")
	}
}

func TestGoogleHandle_ValidationAndBehavior(t *testing.T) {
	g := &google{
		flowID: 1,
		apiKey: "test-api-key",
		cxKey:  "test-cx-key",
	}

	t.Run("invalid json", func(t *testing.T) {
		_, err := g.Handle(t.Context(), GoogleToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		// Use canceled context to make Do() fail immediately and avoid depending
		// on real network/API state in this unit test.
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := g.Handle(
			ctx,
			GoogleToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}
		if !strings.Contains(got, "failed to call tool google to search in google results") {
			t.Fatalf("Handle() = %q, expected swallowed error", got)
		}
	})
}
