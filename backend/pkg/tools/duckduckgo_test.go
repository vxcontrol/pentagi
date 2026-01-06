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

// TestDuckDuckGoBasicSearch tests basic search functionality with a simple query
// This is an integration test that makes real HTTP requests to DuckDuckGo
func TestDuckDuckGoBasicSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode - use unit tests instead")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	flowID := int64(1)
	taskID := int64(1)
	subtaskID := int64(1)

	ddg := NewDuckDuckGoTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"", // no proxy
		RegionUS,
		DuckDuckGoSafeSearchModerate,
		"",
		&MockSearchLogProvider{},
	)

	maxResults := 5
	searchAction := SearchAction{
		Query:      "what is PentAGI?",
		MaxResults: Int64(maxResults),
	}

	args, err := json.Marshal(searchAction)
	if err != nil {
		t.Fatalf("failed to marshal search action: %v", err)
	}

	result, err := ddg.Handle(ctx, "duckduckgo_search", args)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	if result == "No results found" {
		t.Fatal("expected to find results for simple query")
	}

	if pattern := "## URL"; strings.Count(result, pattern) != maxResults {
		t.Errorf("expected %d URLs, got %d", maxResults, strings.Count(result, pattern))
	}
	if pattern := "## Description"; strings.Count(result, pattern) != maxResults {
		t.Errorf("expected %d descriptions, got %d", maxResults, strings.Count(result, pattern))
	}
	for i := range maxResults {
		if !strings.Contains(result, fmt.Sprintf("# %d. ", i+1)) {
			t.Errorf("expected title %d, got %s", i+1, result)
		}
	}
}

// TestDuckDuckGoDisabled tests that disabled tool returns appropriate error
func TestDuckDuckGoDisabled(t *testing.T) {
	flowID := int64(4)
	taskID := int64(4)
	subtaskID := int64(4)

	ddg := NewDuckDuckGoTool(
		flowID,
		&taskID,
		&subtaskID,
		false, // disabled
		"",
		RegionUS,
		DuckDuckGoSafeSearchModerate,
		"",
		&MockSearchLogProvider{},
	)

	if ddg.IsAvailable() {
		t.Error("expected tool to be unavailable when disabled")
	}
}

// TestDuckDuckGoInvalidJSON tests handling of invalid JSON input
func TestDuckDuckGoInvalidJSON(t *testing.T) {
	ctx := context.Background()

	flowID := int64(5)
	taskID := int64(5)
	subtaskID := int64(5)

	ddg := NewDuckDuckGoTool(
		flowID,
		&taskID,
		&subtaskID,
		true,
		"",
		RegionUS,
		DuckDuckGoSafeSearchModerate,
		"",
		&MockSearchLogProvider{},
	)

	invalidJSON := json.RawMessage(`{invalid json}`)

	_, err := ddg.Handle(ctx, "duckduckgo_search", invalidJSON)
	if err == nil {
		t.Error("expected error when handling invalid JSON")
	}
}

// TestParseHTMLStructured tests structured parser
func TestDuckDuckGoParseHTMLStructured(t *testing.T) {
	ddg := &duckduckgo{}
	testdata := []struct {
		filename string
		expected int
	}{
		{filename: "ddg_result_golang_http_client.html", expected: 10},
		{filename: "ddg_result_site_github_golang.html", expected: 10},
		{filename: "ddg_result_owasp_vulnerabilities.html", expected: 10},
		{filename: "ddg_result_sql_injection.html", expected: 10},
		{filename: "ddg_result_docker_security.html", expected: 10},
	}

	for _, tt := range testdata {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			results, err := ddg.parseHTMLStructured(body)
			if err != nil {
				t.Fatalf("parseHTMLStructured failed: %v", err)
			}

			if len(results) != tt.expected {
				t.Fatalf("expected %d results, got %d", tt.expected, len(results))
			}

			// Verify results
			for i, r := range results {
				if r.Title == "" {
					t.Errorf("result %d should have title", i)
				}
				if r.URL == "" {
					t.Errorf("result %d should have URL", i)
				}
				if r.Description == "" {
					t.Errorf("result %d should have description", i)
				}
			}
		})
	}
}

// TestParseHTMLRegex tests regex fallback parser
func TestDuckDuckGoParseHTMLRegex(t *testing.T) {
	ddg := &duckduckgo{}
	testdata := []struct {
		filename string
		expected int
	}{
		{filename: "ddg_result_golang_http_client.html", expected: 10},
		{filename: "ddg_result_site_github_golang.html", expected: 10},
		{filename: "ddg_result_owasp_vulnerabilities.html", expected: 10},
		{filename: "ddg_result_sql_injection.html", expected: 10},
		{filename: "ddg_result_docker_security.html", expected: 10},
	}

	for _, tt := range testdata {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()

			body, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			results, err := ddg.parseHTMLRegex(body)
			if err != nil {
				t.Fatalf("parseHTMLRegex failed: %v", err)
			}

			if len(results) != tt.expected {
				t.Fatalf("expected %d results, got %d", tt.expected, len(results))
			}

			// Verify results
			for i, r := range results {
				if r.Title == "" {
					t.Errorf("result %d should have title", i)
				}
				if r.URL == "" {
					t.Errorf("result %d should have URL", i)
				}
				if r.Description == "" {
					t.Errorf("result %d should have description", i)
				}
			}
		})
	}
}

// TestParseHTMLRegex_BlockBoundaries tests that regex parser correctly identifies block boundaries
func TestDuckDuckGoParseHTMLRegex_BlockBoundaries(t *testing.T) {
	// Sample HTML with multiple result blocks
	html := `
		<div class="result results_links results_links_deep web-result ">
			<div class="links_main links_deep result__body">
				<h2 class="result__title">
					<a rel="nofollow" class="result__a" href="https://example1.com">Example 1</a>
				</h2>
				<a class="result__snippet" href="https://example1.com">First result description</a>
				<div class="clear"></div>
			</div>
		</div>
		<div class="result results_links results_links_deep web-result ">
			<div class="links_main links_deep result__body">
				<h2 class="result__title">
					<a rel="nofollow" class="result__a" href="https://example2.com">Example 2</a>
				</h2>
				<a class="result__snippet" href="https://example2.com">Second result description</a>
				<div class="clear"></div>
			</div>
		</div>
	`

	ddg := &duckduckgo{}
	results, err := ddg.parseHTMLRegex([]byte(html))
	if err != nil {
		t.Fatalf("parseHTMLRegex failed: %v", err)
	}

	// Should find exactly 2 results
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify first result
	if len(results) > 0 {
		if results[0].Title != "Example 1" {
			t.Errorf("first result title = %q, want %q", results[0].Title, "Example 1")
		}
		if results[0].URL != "https://example1.com" {
			t.Errorf("first result URL = %q, want %q", results[0].URL, "https://example1.com")
		}
		if results[0].Description != "First result description" {
			t.Errorf("first result description = %q, want %q", results[0].Description, "First result description")
		}
	}

	// Verify second result
	if len(results) > 1 {
		if results[1].Title != "Example 2" {
			t.Errorf("second result title = %q, want %q", results[1].Title, "Example 2")
		}
		if results[1].URL != "https://example2.com" {
			t.Errorf("second result URL = %q, want %q", results[1].URL, "https://example2.com")
		}
		if results[1].Description != "Second result description" {
			t.Errorf("second result description = %q, want %q", results[1].Description, "Second result description")
		}
	}
}

// TestCleanText tests HTML entity decoding and text cleaning
func TestDuckDuckGoCleanText(t *testing.T) {
	ddg := &duckduckgo{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML tags",
			input:    "This is <b>bold</b> text",
			expected: "This is bold text",
		},
		{
			name:     "HTML entities",
			input:    "Go&#x27;s http package",
			expected: "Go's http package",
		},
		{
			name:     "Multiple entities",
			input:    "&quot;Hello&quot; &amp; &lt;goodbye&gt;",
			expected: "\"Hello\" & <goodbye>",
		},
		{
			name:     "Whitespace normalization",
			input:    "Multiple   spaces   and\n\nnewlines",
			expected: "Multiple spaces and newlines",
		},
		{
			name:     "Complex HTML",
			input:    "The <b>http</b> package&#x27;s Transport &amp; Server",
			expected: "The http package's Transport & Server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ddg.cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
