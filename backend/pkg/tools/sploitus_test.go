package tools

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func testSploitusConfig() *config.Config {
	return &config.Config{SploitusEnabled: true}
}

func TestSploitusHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedContentType string
	var receivedAccept string
	var receivedOrigin string
	var receivedReferer string
	var receivedUserAgent string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedAccept = r.Header.Get("Accept")
		receivedOrigin = r.Header.Get("Origin")
		receivedReferer = r.Header.Get("Referer")
		receivedUserAgent = r.Header.Get("User-Agent")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"exploits":[
				{
					"id":"CVE-2024-1234",
					"title":"Test Exploit for nginx",
					"type":"githubexploit",
					"href":"https://github.com/test/exploit",
					"score":9.8,
					"published":"2024-01-15",
					"language":"python",
					"source":"exploit code here"
				}
			],
			"exploits_total":42
		}`))
	})

	proxy, err := newTestProxy("sploitus.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		SploitusEnabled:   true,
		ProxyURL:          proxy.URL(),
		ExternalSSLCAPath: proxy.CACertPath(),
	}

	sp := NewSploitusTool(cfg, flowID, &taskID, &subtaskID, slp)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := sp.Handle(
		ctx,
		SploitusToolName,
		[]byte(`{"query":"nginx","exploit_type":"exploits","sort":"date","max_results":5}`),
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
	if receivedAccept != "application/json" {
		t.Errorf("Accept = %q, want application/json", receivedAccept)
	}
	if receivedOrigin != "https://sploitus.com" {
		t.Errorf("Origin = %q, want https://sploitus.com", receivedOrigin)
	}
	if !strings.Contains(receivedReferer, "sploitus.com") {
		t.Errorf("Referer = %q, want to contain sploitus.com", receivedReferer)
	}
	if !strings.Contains(receivedUserAgent, "Mozilla") {
		t.Errorf("User-Agent = %q, want to contain Mozilla", receivedUserAgent)
	}
	if !strings.Contains(string(receivedBody), `"query":"nginx"`) {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"type":"exploits"`) {
		t.Errorf("request body = %q, expected to contain type", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), `"sort":"date"`) {
		t.Errorf("request body = %q, expected to contain sort", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Sploitus Search Results") {
		t.Errorf("result missing '# Sploitus Search Results' section: %q", got)
	}
	if !strings.Contains(got, "**Query:** `nginx`") {
		t.Errorf("result missing expected query: %q", got)
	}
	if !strings.Contains(got, "**Total matches on Sploitus:** 42") {
		t.Errorf("result missing expected total: %q", got)
	}
	if !strings.Contains(got, "Test Exploit for nginx") {
		t.Errorf("result missing expected title: %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeSploitus {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeSploitus)
	}
	if slp.query != "nginx" {
		t.Errorf("logged query = %q, want %q", slp.query, "nginx")
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

func TestSploitusIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when enabled",
			cfg:  testSploitusConfig(),
			want: true,
		},
		{
			name: "unavailable when disabled",
			cfg:  &config.Config{SploitusEnabled: false},
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
			sp := &sploitus{cfg: tt.cfg}
			if got := sp.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSploitusHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		sp := &sploitus{cfg: testSploitusConfig()}
		_, err := sp.Handle(t.Context(), SploitusToolName, []byte("{"))
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

		proxy, err := newTestProxy("sploitus.com", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		sp := &sploitus{
			flowID: 1,
			cfg: &config.Config{
				SploitusEnabled:   true,
				ProxyURL:          proxy.URL(),
				ExternalSSLCAPath: proxy.CACertPath(),
			},
		}

		result, err := sp.Handle(
			t.Context(),
			SploitusToolName,
			[]byte(`{"query":"test","exploit_type":"exploits"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}

		// Verify mock handler was called (request was intercepted)
		if !seenRequest {
			t.Error("request was not intercepted by proxy - mock handler was not called")
		}

		// Verify error was swallowed and returned as string
		if !strings.Contains(result, "failed to search in Sploitus") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}

func TestSploitusHandle_StatusCodeErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errContain string
	}{
		{"rate limit 499", 499, "rate limit exceeded"},
		{"rate limit 422", 422, "rate limit exceeded"},
		{"server error", http.StatusInternalServerError, "HTTP 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			proxy, err := newTestProxy("sploitus.com", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			sp := &sploitus{
				flowID: 1,
				cfg: &config.Config{
					SploitusEnabled:   true,
					ProxyURL:          proxy.URL(),
					ExternalSSLCAPath: proxy.CACertPath(),
				},
			}

			result, err := sp.Handle(
				t.Context(),
				SploitusToolName,
				[]byte(`{"query":"test","exploit_type":"exploits"}`),
			)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			// Error should be swallowed and returned as string
			if !strings.Contains(result, "failed to search in Sploitus") {
				t.Errorf("Handle() = %q, expected swallowed error", result)
			}
			if !strings.Contains(result, tt.errContain) {
				t.Errorf("Handle() = %q, expected to contain %q", result, tt.errContain)
			}
		})
	}
}

func TestSploitusFormatResults(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		exploitType string
		limit       int
		response    sploitusResponse
		expected    []string
	}{
		{
			name:        "exploits formatting",
			query:       "CVE-2026",
			exploitType: "exploits",
			limit:       2,
			response: sploitusResponse{
				Exploits: []sploitusExploit{
					{
						ID:        "TEST-001",
						Title:     "Test Exploit 1",
						Type:      "githubexploit",
						Href:      "https://example.com/exploit1",
						Score:     9.8,
						Published: "2026-01-15",
						Language:  "python",
					},
					{
						ID:        "TEST-002",
						Title:     "Test Exploit 2",
						Type:      "packetstorm",
						Href:      "https://example.com/exploit2",
						Score:     7.5,
						Published: "2026-01-20",
					},
				},
				ExploitsTotal: 100,
			},
			expected: []string{
				"# Sploitus Search Results",
				"**Query:** `CVE-2026`",
				"**Type:** exploits",
				"**Total matches on Sploitus:** 100",
				"## Exploits (showing up to 2)",
				"### 1. Test Exploit 1",
				"**URL:** https://example.com/exploit1",
				"**CVSS Score:** 9.8",
				"**Type:** githubexploit",
				"**Published:** 2026-01-15",
				"**Language:** python",
				"### 2. Test Exploit 2",
				"**CVSS Score:** 7.5",
			},
		},
		{
			name:        "tools formatting",
			query:       "nmap",
			exploitType: "tools",
			limit:       2,
			response: sploitusResponse{
				Exploits: []sploitusExploit{
					{
						ID:       "TOOL-001",
						Title:    "Nmap Tool 1",
						Type:     "kitploit",
						Href:     "https://example.com/tool1",
						Download: "https://github.com/tool1",
					},
					{
						ID:       "TOOL-002",
						Title:    "Nmap Tool 2",
						Type:     "n0where",
						Href:     "https://example.com/tool2",
						Download: "https://github.com/tool2",
					},
				},
				ExploitsTotal: 200,
			},
			expected: []string{
				"# Sploitus Search Results",
				"**Query:** `nmap`",
				"**Type:** tools",
				"**Total matches on Sploitus:** 200",
				"## Security Tools (showing up to 2)",
				"### 1. Nmap Tool 1",
				"**URL:** https://example.com/tool1",
				"**Download:** https://github.com/tool1",
				"**Source Type:** kitploit",
				"### 2. Nmap Tool 2",
				"**Download:** https://github.com/tool2",
			},
		},
		{
			name:        "empty results",
			query:       "nonexistent",
			exploitType: "exploits",
			limit:       10,
			response: sploitusResponse{
				Exploits:      []sploitusExploit{},
				ExploitsTotal: 0,
			},
			expected: []string{
				"# Sploitus Search Results",
				"**Query:** `nonexistent`",
				"No exploits were found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSploitusResults(tt.query, tt.exploitType, tt.limit, tt.response)

			for _, expectedStr := range tt.expected {
				if !strings.Contains(result, expectedStr) {
					t.Errorf("expected result to contain %q\nGot:\n%s", expectedStr, result)
				}
			}
		})
	}
}

func TestSploitusDefaultValues(t *testing.T) {
	mockMux := http.NewServeMux()
	var receivedBody []byte
	mockMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exploits":[],"exploits_total":0}`))
	})

	proxy, err := newTestProxy("sploitus.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	sp := &sploitus{
		flowID: 1,
		cfg: &config.Config{
			SploitusEnabled:   true,
			ProxyURL:          proxy.URL(),
			ExternalSSLCAPath: proxy.CACertPath(),
		},
	}

	// Test with minimal action (only query, no type/sort/maxResults)
	_, err = sp.Handle(
		t.Context(),
		SploitusToolName,
		[]byte(`{"query":"test"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	// Verify defaults were applied
	bodyStr := string(receivedBody)
	if !strings.Contains(bodyStr, `"type":"exploits"`) {
		t.Errorf("expected default type 'exploits', got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, `"sort":"default"`) {
		t.Errorf("expected default sort 'default', got: %s", bodyStr)
	}
}

func TestSploitusSizeLimits(t *testing.T) {
	t.Run("source truncation at 50KB", func(t *testing.T) {
		// Create a large source (60 KB)
		largeSource := strings.Repeat("A", 60*1024)

		resp := sploitusResponse{
			Exploits: []sploitusExploit{
				{
					ID:     "TEST-1",
					Title:  "Test with large source",
					Href:   "https://example.com",
					Source: largeSource,
				},
			},
			ExploitsTotal: 1,
		}

		result := formatSploitusResults("test", "exploits", 10, resp)

		// Check that source was truncated
		if !strings.Contains(result, "source truncated, exceeded 50 KB limit") {
			t.Error("expected source truncation message for 60 KB source")
		}

		// Verify result doesn't contain the full 60 KB
		if len(result) > 80*1024 {
			t.Errorf("result size %d exceeds 80 KB limit", len(result))
		}
	})

	t.Run("total size limit at 80KB", func(t *testing.T) {
		// Create many results to exceed 80 KB total
		results := make([]sploitusExploit, 100)
		for i := range results {
			results[i] = sploitusExploit{
				ID:     fmt.Sprintf("TEST-%d", i),
				Title:  fmt.Sprintf("Test Result %d", i),
				Href:   "https://example.com",
				Source: strings.Repeat("X", 5000), // 5 KB each
			}
		}

		resp := sploitusResponse{
			Exploits:      results,
			ExploitsTotal: 100,
		}

		result := formatSploitusResults("test", "exploits", 100, resp)

		// Result should be under 80 KB
		if len(result) > 80*1024 {
			t.Errorf("result size %d exceeds 80 KB hard limit", len(result))
		}

		// Should have truncation warning
		if !strings.Contains(result, "Results truncated") {
			t.Error("expected truncation warning when hitting 80 KB limit")
		}

		// Should not show all 100 results
		count := strings.Count(result, "### ")
		if count >= 100 {
			t.Errorf("expected fewer than 100 results due to size limit, got %d", count)
		}
	})
}

func TestSploitusMaxResultsClamp(t *testing.T) {
	tests := []struct {
		name          string
		maxResults    int
		expectedCount int
	}{
		{"valid max results", 10, 10},
		{"valid smaller", 5, 5},
		{"too large", 100, 30}, // Should limit to available results (30)
		{"zero gets default", 0, defaultSploitusLimit},
		{"negative gets default", -5, defaultSploitusLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response with 30 results
			resp := sploitusResponse{
				Exploits:      make([]sploitusExploit, 30),
				ExploitsTotal: 30,
			}

			// Fill with dummy data
			for i := range resp.Exploits {
				resp.Exploits[i] = sploitusExploit{
					ID:    fmt.Sprintf("TEST-%d", i),
					Title: fmt.Sprintf("Test %d", i),
					Href:  "https://example.com",
				}
			}

			result := formatSploitusResults("test", "exploits", tt.maxResults, resp)

			// Count how many results are shown (### is used for each result title)
			count := strings.Count(result, "### ")

			if count != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, count)
			}
		})
	}
}
