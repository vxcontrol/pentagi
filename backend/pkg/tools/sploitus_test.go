package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSploitusExploitsSearch tests exploit search functionality with a CVE query
// This is an integration test that makes real HTTP requests to Sploitus
func TestSploitusExploitsSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode - use unit tests instead")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	flowID := int64(1)
	taskID := int64(1)
	subtaskID := int64(1)

	sploitus := NewSploitusTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"", // no proxy
		&MockSearchLogProvider{},
	)

	maxResults := 5
	searchAction := SploitusAction{
		Query:       "nginx",
		ExploitType: "exploits",
		Sort:        "date",
		MaxResults:  Int64(maxResults),
	}

	args, err := json.Marshal(searchAction)
	if err != nil {
		t.Fatalf("failed to marshal search action: %v", err)
	}

	result, err := sploitus.Handle(ctx, "sploitus_search", args)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	if strings.Contains(result, "No exploits were found") {
		t.Fatal("expected to find exploits for nginx query")
	}

	// Check for expected structure
	if !strings.Contains(result, "# Sploitus Search Results") {
		t.Error("expected results header")
	}

	if !strings.Contains(result, "**Query:** `nginx`") {
		t.Error("expected query in results")
	}

	if !strings.Contains(result, "**Type:** exploits") {
		t.Error("expected type in results")
	}
}

// TestSploitusToolsSearch tests tools search functionality
// This is an integration test that makes real HTTP requests to Sploitus
func TestSploitusToolsSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode - use unit tests instead")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	flowID := int64(2)
	taskID := int64(2)
	subtaskID := int64(2)

	sploitus := NewSploitusTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"", // no proxy
		&MockSearchLogProvider{},
	)

	maxResults := 5
	searchAction := SploitusAction{
		Query:       "metasploit",
		ExploitType: "tools",
		Sort:        "default",
		MaxResults:  Int64(maxResults),
	}

	args, err := json.Marshal(searchAction)
	if err != nil {
		t.Fatalf("failed to marshal search action: %v", err)
	}

	result, err := sploitus.Handle(ctx, "sploitus_search", args)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	if strings.Contains(result, "No security tools were found") {
		t.Fatal("expected to find tools for metasploit query")
	}

	// Check for expected structure
	if !strings.Contains(result, "# Sploitus Search Results") {
		t.Error("expected results header")
	}

	if !strings.Contains(result, "**Query:** `metasploit`") {
		t.Error("expected query in results")
	}

	if !strings.Contains(result, "**Type:** tools") {
		t.Error("expected type in results")
	}

	if !strings.Contains(result, "## Security Tools") {
		t.Error("expected security tools section")
	}
}

// TestSploitusDisabled tests that disabled tool returns appropriate status
func TestSploitusDisabled(t *testing.T) {
	flowID := int64(3)
	taskID := int64(3)
	subtaskID := int64(3)

	sploitus := NewSploitusTool(
		flowID,
		&taskID,
		&subtaskID,
		false, // disabled
		"",
		&MockSearchLogProvider{},
	)

	if sploitus.IsAvailable() {
		t.Error("expected tool to be unavailable when disabled")
	}
}

// TestSploitusInvalidJSON tests handling of invalid JSON input
func TestSploitusInvalidJSON(t *testing.T) {
	ctx := context.Background()

	flowID := int64(4)
	taskID := int64(4)
	subtaskID := int64(4)

	sploitus := NewSploitusTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"",
		&MockSearchLogProvider{},
	)

	invalidJSON := json.RawMessage(`{invalid json}`)

	_, err := sploitus.Handle(ctx, "sploitus_search", invalidJSON)
	if err == nil {
		t.Error("expected error when handling invalid JSON")
	}
}

// TestSploitusParseExploitsResponse tests parsing of exploits response from test data
func TestSploitusParseExploitsResponse(t *testing.T) {
	testdata := []struct {
		filename    string
		expectedMin int
		exploitType string
	}{
		{filename: "sploitus_result_cve_2026.json", expectedMin: 5, exploitType: "exploits"},
		{filename: "sploitus_result_nginx.json", expectedMin: 5, exploitType: "exploits"},
	}

	for _, tt := range testdata {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			var resp sploitusResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(resp.Exploits) < tt.expectedMin {
				t.Fatalf("expected at least %d results, got %d", tt.expectedMin, len(resp.Exploits))
			}

			if resp.ExploitsTotal == 0 {
				t.Error("expected non-zero total count")
			}

			// Verify result structure
			for i, item := range resp.Exploits {
				if item.Title == "" {
					t.Errorf("result %d should have title", i)
				}
				if item.Href == "" {
					t.Errorf("result %d should have href", i)
				}
				if item.ID == "" {
					t.Errorf("result %d should have id", i)
				}
				if item.Type == "" {
					t.Errorf("result %d should have type", i)
				}
			}
		})
	}
}

// TestSploitusParseToolsResponse tests parsing of tools response from test data
func TestSploitusParseToolsResponse(t *testing.T) {
	testdata := []struct {
		filename    string
		expectedMin int
		exploitType string
	}{
		{filename: "sploitus_result_metasploit.json", expectedMin: 5, exploitType: "tools"},
		{filename: "sploitus_result_nmap.json", expectedMin: 5, exploitType: "tools"},
	}

	for _, tt := range testdata {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			var resp sploitusResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if len(resp.Exploits) < tt.expectedMin {
				t.Fatalf("expected at least %d results, got %d", tt.expectedMin, len(resp.Exploits))
			}

			if resp.ExploitsTotal == 0 {
				t.Error("expected non-zero total count")
			}

			// Verify result structure for tools
			for i, item := range resp.Exploits {
				if item.Title == "" {
					t.Errorf("result %d should have title", i)
				}
				if item.Href == "" {
					t.Errorf("result %d should have href", i)
				}
				if item.ID == "" {
					t.Errorf("result %d should have id", i)
				}
				if item.Type == "" {
					t.Errorf("result %d should have type", i)
				}
				// Tools should have download field
				if item.Download == "" {
					t.Logf("warning: result %d missing download field", i)
				}
			}
		})
	}
}

// TestSploitusFormatResults tests the formatting of results
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

// TestSploitusDefaultValues tests default values for optional parameters
func TestSploitusDefaultValues(t *testing.T) {
	ctx := context.Background()

	flowID := int64(5)
	taskID := int64(5)
	subtaskID := int64(5)

	sploitus := NewSploitusTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"",
		&MockSearchLogProvider{},
	)

	// Test with minimal action (only query)
	searchAction := SploitusAction{
		Query: "test",
	}

	args, err := json.Marshal(searchAction)
	if err != nil {
		t.Fatalf("failed to marshal search action: %v", err)
	}

	// This will fail in actual API call, but we're testing parameter handling
	// The Handle method should set defaults: exploitType="exploits", sort="default", maxResults=10
	_, _ = sploitus.Handle(ctx, "sploitus_search", args)

	// The test passes if no panic occurs during parameter processing
}

// TestSploitusSizeLimits tests hard size limits (50 KB source, 80 KB total)
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
		if strings.Contains(result, "source truncated, exceeded 50 KB limit") {
			// Good, truncation message is present
		} else {
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

// TestSploitusMaxResultsClamp tests that max results are clamped to valid range
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
