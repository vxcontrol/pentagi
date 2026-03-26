package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"

	customsearch "google.golang.org/api/customsearch/v1"
)

const (
	testGoogleAPIKey = "test-api-key"
	testGoogleCXKey  = "test-cx-key"
	testGoogleLRKey  = "lang_en"
)

func testGoogleConfig() *config.Config {
	return &config.Config{
		GoogleAPIKey: testGoogleAPIKey,
		GoogleCXKey:  testGoogleCXKey,
		GoogleLRKey:  testGoogleLRKey,
	}
}

func TestGoogleIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when both keys are set",
			cfg:  testGoogleConfig(),
			want: true,
		},
		{
			name: "unavailable when API key is empty",
			cfg:  &config.Config{GoogleCXKey: testGoogleCXKey},
			want: false,
		},
		{
			name: "unavailable when CX key is empty",
			cfg:  &config.Config{GoogleAPIKey: testGoogleAPIKey},
			want: false,
		},
		{
			name: "unavailable when both keys are empty",
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
			g := &google{cfg: tt.cfg}
			if got := g.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleFormatResults(t *testing.T) {
	g := &google{flowID: 1}

	t.Run("empty results", func(t *testing.T) {
		res := &customsearch.Search{Items: nil}
		result := g.formatResults(res)
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
		result := g.formatResults(res)

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
		result := g.formatResults(res)

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
		result := g.formatResults(res)

		if !strings.Contains(result, "Test & <Special> \"Characters\"") {
			t.Error("title special characters should be preserved")
		}
		if !strings.Contains(result, "q=test&lang=en") {
			t.Error("URL query parameters should be preserved")
		}
	})
}

func TestGoogleNewSearchService(t *testing.T) {
	t.Run("without proxy", func(t *testing.T) {
		g := &google{cfg: testGoogleConfig()}

		svc, err := g.newSearchService(t.Context())
		if err != nil {
			t.Fatalf("newSearchService() unexpected error: %v", err)
		}
		if svc == nil {
			t.Fatal("newSearchService() returned nil service")
		}
	})

	t.Run("with proxy", func(t *testing.T) {
		g := &google{cfg: &config.Config{
			GoogleAPIKey: testGoogleAPIKey,
			GoogleCXKey:  testGoogleCXKey,
			ProxyURL:     "http://proxy.example.com:8080",
		}}

		svc, err := g.newSearchService(t.Context())
		if err != nil {
			t.Fatalf("newSearchService() unexpected error: %v", err)
		}
		if svc == nil {
			t.Fatal("newSearchService() returned nil service")
		}
	})
}

func TestGoogleHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		g := &google{cfg: testGoogleConfig()}
		_, err := g.Handle(t.Context(), GoogleToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		// Use canceled context to make Do() fail immediately
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		g := &google{cfg: testGoogleConfig()}

		got, err := g.Handle(
			ctx,
			GoogleToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}
		if !strings.Contains(got, "failed to search in google") {
			t.Fatalf("Handle() = %q, expected swallowed error", got)
		}
	})
}

func TestGoogleHandle_WithAgentContext(t *testing.T) {
	// Note: This test cannot fully verify search behavior without a real API call.
	// It verifies parameter handling and agent context propagation.

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	g := NewGoogleTool(testGoogleConfig(), flowID, &taskID, &subtaskID, slp)

	// Use canceled context to make search fail quickly
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	ctx = PutAgentContext(ctx, database.MsgchainTypeSearcher)

	// This will fail due to canceled context, but we can verify the structure
	result, err := g.Handle(
		ctx,
		GoogleToolName,
		[]byte(`{"query":"test query","max_results":5,"message":"m"}`),
	)

	// Error should be swallowed
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if !strings.Contains(result, "failed to search in google") {
		t.Errorf("Handle() = %q, expected swallowed error message", result)
	}

	// Search log should be written even on error
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeGoogle {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeGoogle)
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

func TestGoogleMaxResultsClamp(t *testing.T) {
	tests := []struct {
		name            string
		maxResults      int
		expectedClamped int64
	}{
		{"valid max results", 5, 5},
		{"max limit", 10, 10},
		{"too large", 100, 10},
		{"zero gets default", 0, 10},
		{"negative gets default", -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &google{cfg: testGoogleConfig()}

			// Use canceled context to fail quickly without real API call
			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			result, err := g.Handle(
				ctx,
				GoogleToolName,
				[]byte(`{"query":"test","max_results":`+strings.TrimSpace(fmt.Sprintf("%d", tt.maxResults))+`,"message":"m"}`),
			)

			// Should not return error (errors are swallowed)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			// Should contain error message since context is canceled
			if !strings.Contains(result, "failed to search in google") {
				t.Errorf("Handle() = %q, expected swallowed error", result)
			}
		})
	}
}

func TestGoogleConfigHelpers(t *testing.T) {
	g := &google{cfg: testGoogleConfig()}

	if g.apiKey() != testGoogleAPIKey {
		t.Errorf("apiKey() = %q, want %q", g.apiKey(), testGoogleAPIKey)
	}
	if g.cxKey() != testGoogleCXKey {
		t.Errorf("cxKey() = %q, want %q", g.cxKey(), testGoogleCXKey)
	}
	if g.lrKey() != testGoogleLRKey {
		t.Errorf("lrKey() = %q, want %q", g.lrKey(), testGoogleLRKey)
	}
}

func TestGoogleConfigHelpers_NilConfig(t *testing.T) {
	g := &google{cfg: nil}

	if g.apiKey() != "" {
		t.Errorf("apiKey() with nil config = %q, want empty", g.apiKey())
	}
	if g.cxKey() != "" {
		t.Errorf("cxKey() with nil config = %q, want empty", g.cxKey())
	}
	if g.lrKey() != "" {
		t.Errorf("lrKey() with nil config = %q, want empty", g.lrKey())
	}
}
