package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func testDuckDuckGoConfig() *config.Config {
	return &config.Config{
		DuckDuckGoEnabled:    true,
		DuckDuckGoRegion:     RegionUS,
		DuckDuckGoSafeSearch: DuckDuckGoSafeSearchModerate,
		DuckDuckGoTimeRange:  "",
	}
}

func TestDuckDuckGoHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedContentType string
	var receivedUserAgent string
	var receivedAccept string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/html/", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedUserAgent = r.Header.Get("User-Agent")
		receivedAccept = r.Header.Get("Accept")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		// Serve a simple mock HTML response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<div class="result results_links results_links_deep web-result">
				<div class="links_main links_deep result__body">
					<h2 class="result__title">
						<a rel="nofollow" class="result__a" href="https://example.com/test">Test Result Title</a>
					</h2>
					<a class="result__snippet" href="https://example.com/test">This is a test description</a>
					<div class="clear"></div>
				</div>
			</div>
		`))
	})

	proxy, err := newTestProxy("html.duckduckgo.com", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		DuckDuckGoEnabled:    true,
		DuckDuckGoRegion:     RegionUS,
		DuckDuckGoSafeSearch: DuckDuckGoSafeSearchModerate,
		ProxyURL:             proxy.URL(),
		ExternalSSLCAPath:    proxy.CACertPath(),
	}

	ddg := NewDuckDuckGoTool(cfg, flowID, &taskID, &subtaskID, slp)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := ddg.Handle(
		ctx,
		DuckDuckGoToolName,
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
	if receivedContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", receivedContentType)
	}
	if !strings.Contains(receivedUserAgent, "Mozilla") {
		t.Errorf("User-Agent = %q, want to contain Mozilla", receivedUserAgent)
	}
	if !strings.Contains(receivedAccept, "text/html") {
		t.Errorf("Accept = %q, want to contain text/html", receivedAccept)
	}
	if !strings.Contains(string(receivedBody), "q=test+query") {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}
	if !strings.Contains(string(receivedBody), "kl=us-en") {
		t.Errorf("request body = %q, expected to contain region", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# 1. Test Result Title") {
		t.Errorf("result missing expected title: %q", got)
	}
	if !strings.Contains(got, "https://example.com/test") {
		t.Errorf("result missing expected URL: %q", got)
	}
	if !strings.Contains(got, "This is a test description") {
		t.Errorf("result missing expected description: %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeDuckduckgo {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeDuckduckgo)
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

func TestDuckDuckGoIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when enabled",
			cfg:  testDuckDuckGoConfig(),
			want: true,
		},
		{
			name: "unavailable when disabled",
			cfg:  &config.Config{DuckDuckGoEnabled: false},
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
			ddg := &duckduckgo{cfg: tt.cfg}
			if got := ddg.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuckDuckGoHandle_ValidationAndSwallowedError(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		ddg := &duckduckgo{cfg: testDuckDuckGoConfig()}
		_, err := ddg.Handle(t.Context(), DuckDuckGoToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		var seenRequest bool
		mockMux := http.NewServeMux()
		mockMux.HandleFunc("/html/", func(w http.ResponseWriter, r *http.Request) {
			seenRequest = true
			w.WriteHeader(http.StatusBadGateway)
		})

		proxy, err := newTestProxy("html.duckduckgo.com", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		ddg := &duckduckgo{
			flowID: 1,
			cfg: &config.Config{
				DuckDuckGoEnabled: true,
				ProxyURL:          proxy.URL(),
				ExternalSSLCAPath: proxy.CACertPath(),
			},
		}

		result, err := ddg.Handle(
			t.Context(),
			DuckDuckGoToolName,
			[]byte(`{"query":"test","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}

		// Verify mock handler was called (request was intercepted)
		if !seenRequest {
			t.Error("request was not intercepted by proxy - mock handler was not called")
		}

		// Verify error was swallowed and returned as string
		if !strings.Contains(result, "failed to search in DuckDuckGo") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}

func TestDuckDuckGoHandle_StatusCodeErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"server error", http.StatusInternalServerError},
		{"not found", http.StatusNotFound},
		{"forbidden", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMux := http.NewServeMux()
			mockMux.HandleFunc("/html/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			proxy, err := newTestProxy("html.duckduckgo.com", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			ddg := &duckduckgo{
				flowID: 1,
				cfg: &config.Config{
					DuckDuckGoEnabled: true,
					ProxyURL:          proxy.URL(),
					ExternalSSLCAPath: proxy.CACertPath(),
				},
			}

			result, err := ddg.Handle(
				t.Context(),
				DuckDuckGoToolName,
				[]byte(`{"query":"test","max_results":5,"message":"m"}`),
			)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			// Error should be swallowed and returned as string
			if !strings.Contains(result, "failed to search in DuckDuckGo") {
				t.Errorf("Handle() = %q, expected swallowed error", result)
			}
			if !strings.Contains(result, "unexpected status code") {
				t.Errorf("Handle() = %q, expected status code error", result)
			}
		})
	}
}

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

func TestDuckDuckGoParseHTMLRegex_BlockBoundaries(t *testing.T) {
	// Sample HTML with multiple result blocks
	htmlContent := `
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
	results, err := ddg.parseHTMLRegex([]byte(htmlContent))
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

func TestDuckDuckGoFormatResults(t *testing.T) {
	ddg := &duckduckgo{}

	t.Run("empty results", func(t *testing.T) {
		result := ddg.formatSearchResults([]searchResult{})
		if result != "" {
			t.Errorf("expected empty string for no results, got %q", result)
		}
	})

	t.Run("single result", func(t *testing.T) {
		results := []searchResult{
			{
				Title:       "Go Programming",
				URL:         "https://go.dev",
				Description: "Go is a programming language",
			},
		}
		result := ddg.formatSearchResults(results)

		if !strings.Contains(result, "# 1. Go Programming") {
			t.Error("result should contain numbered title")
		}
		if !strings.Contains(result, "## URL\nhttps://go.dev") {
			t.Error("result should contain URL section")
		}
		if !strings.Contains(result, "## Description") {
			t.Error("result should contain Description section")
		}
		if strings.Contains(result, "---") {
			t.Error("result should NOT contain separator for single result")
		}
	})

	t.Run("multiple results", func(t *testing.T) {
		results := []searchResult{
			{Title: "First", URL: "https://first.com", Description: "first desc"},
			{Title: "Second", URL: "https://second.com", Description: "second desc"},
		}
		result := ddg.formatSearchResults(results)

		if !strings.Contains(result, "# 1. First") {
			t.Error("result should contain first title")
		}
		if !strings.Contains(result, "# 2. Second") {
			t.Error("result should contain second title")
		}
		if !strings.Contains(result, "---") {
			t.Error("result should contain separator between results")
		}
	})
}

func TestDuckDuckGoMaxResultsClamp(t *testing.T) {
	tests := []struct {
		name       string
		maxResults int
		wantClamp  int
	}{
		{"valid max results", 5, 5},
		{"max limit", 10, 10},
		{"too large", 100, 10},
		{"zero gets default", 0, 10},
		{"negative gets default", -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMux := http.NewServeMux()
			var receivedQuery string
			mockMux.HandleFunc("/html/", func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				receivedQuery = string(body)
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<div>No results</div>`))
			})

			proxy, err := newTestProxy("html.duckduckgo.com", mockMux)
			if err != nil {
				t.Fatalf("failed to create proxy: %v", err)
			}
			defer proxy.Close()

			ddg := &duckduckgo{
				flowID: 1,
				cfg: &config.Config{
					DuckDuckGoEnabled: true,
					ProxyURL:          proxy.URL(),
					ExternalSSLCAPath: proxy.CACertPath(),
				},
			}

			_, err = ddg.Handle(
				t.Context(),
				DuckDuckGoToolName,
				[]byte(fmt.Sprintf(`{"query":"test","max_results":%d,"message":"m"}`, tt.maxResults)),
			)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			// Verify request was made (proxy captured it)
			if !strings.Contains(receivedQuery, "q=test") {
				t.Errorf("request not captured or query missing: %q", receivedQuery)
			}
		})
	}
}

func TestDuckDuckGoSafeSearchMapping(t *testing.T) {
	tests := []struct {
		name       string
		safeSearch string
		want       string
	}{
		{"strict", DuckDuckGoSafeSearchStrict, "1"},
		{"moderate", DuckDuckGoSafeSearchModerate, "0"},
		{"off", DuckDuckGoSafeSearchOff, "-1"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ddg := &duckduckgo{cfg: &config.Config{DuckDuckGoSafeSearch: tt.safeSearch}}
			if got := ddg.safeSearch(); got != tt.want {
				t.Errorf("safeSearch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDuckDuckGoRegionDefault(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{"custom region", &config.Config{DuckDuckGoRegion: RegionDE}, RegionDE},
		{"empty defaults to US", &config.Config{DuckDuckGoRegion: ""}, RegionUS},
		{"nil config defaults to US", nil, RegionUS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ddg := &duckduckgo{cfg: tt.cfg}
			if got := ddg.region(); got != tt.want {
				t.Errorf("region() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDuckDuckGoTimeRange(t *testing.T) {
	tests := []struct {
		name      string
		timeRange string
		want      string
	}{
		{"day", TimeRangeDay, TimeRangeDay},
		{"week", TimeRangeWeek, TimeRangeWeek},
		{"month", TimeRangeMonth, TimeRangeMonth},
		{"year", TimeRangeYear, TimeRangeYear},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ddg := &duckduckgo{cfg: &config.Config{DuckDuckGoTimeRange: tt.timeRange}}
			if got := ddg.timeRange(); got != tt.want {
				t.Errorf("timeRange() = %q, want %q", got, tt.want)
			}
		})
	}
}
