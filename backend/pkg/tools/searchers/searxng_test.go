package searchers

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func TestSearxng_Handle_SendsTheConfiguredQueryAndReturnsItsOutcome(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "searxng.example.com", up)
	const found = `{"query":"test query","results":[{"title":"Test Result","url":"https://example.com","content":"Test content","engine":"google"}]}`

	for _, tt := range []struct {
		name          string
		cfg           config.Config
		maxResults    int
		status        int
		wantQuery     url.Values
		wantRetryable bool
	}{
		{
			name:       "the configured settings and limit are sent",
			cfg:        config.Config{SearxngLanguage: "en", SearxngCategories: "general", SearxngSafeSearch: "0"},
			maxResults: 5,
			status:     http.StatusOK,
			wantQuery: url.Values{
				"q": {"test query"}, "format": {"json"}, "language": {"en"}, "categories": {"general"},
				"safesearch": {"0"}, "limit": {"5"},
			},
		},
		{
			name:   "no limit asks for ten",
			status: http.StatusOK,
			wantQuery: url.Values{
				"q": {"test query"}, "format": {"json"}, "language": {""}, "categories": {""},
				"safesearch": {""}, "limit": {"10"},
			},
		},
		{
			name:       "a time range is passed on",
			cfg:        config.Config{SearxngTimeRange: "day"},
			maxResults: 5,
			status:     http.StatusOK,
			wantQuery: url.Values{
				"q": {"test query"}, "format": {"json"}, "language": {""}, "categories": {""},
				"safesearch": {""}, "time_range": {"day"}, "limit": {"5"},
			},
		},
		{
			name:       "an upstream 502 is retryable",
			maxResults: 5,
			status:     http.StatusBadGateway,
			wantQuery: url.Values{
				"q": {"test query"}, "format": {"json"}, "language": {""}, "categories": {""},
				"safesearch": {""}, "limit": {"5"},
			},
			wantRetryable: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, found)
			tt.cfg.SearxngURL, tt.cfg.SearxngTimeout = "http://searxng.example.com", 30
			tt.cfg.ProxyURL, tt.cfg.ExternalSSLCAPath = p.URL(), p.CACertPath()
			sx := NewSearxng(&tt.cfg, nil)
			if sx.Engine() != database.SearchengineTypeSearxng {
				t.Errorf("Engine() = %q, want %q", sx.Engine(), database.SearchengineTypeSearxng)
			}

			got, err := sx.Handle(t.Context(), Request{Query: "test query", MaxResults: tt.maxResults})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if up.method != http.MethodGet || up.path != "/search" {
				t.Errorf("request = %s %s, want GET /search", up.method, up.path)
			}
			if ua := up.header.Get("User-Agent"); ua != "PentAGI/1.0" {
				t.Errorf("User-Agent = %q, want PentAGI/1.0", ua)
			}
			if !reflect.DeepEqual(up.query, tt.wantQuery) {
				t.Errorf("query = %v, want %v", up.query, tt.wantQuery)
			}

			if tt.wantRetryable {
				if got != "" || !IsRetryable(err) {
					t.Fatalf("Handle() = %q, %v; want no result and a RetryableError", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			for _, want := range []string{"# Searxng Search Results", "Test Result", "https://example.com", "Test content"} {
				if !strings.Contains(got, want) {
					t.Errorf("result misses %q: %q", want, got)
				}
			}
		})
	}
}

func TestSearxng_ParseHTTPResponse_ClassifiesTheAnswer(t *testing.T) {
	sx := &searxng{}

	for _, tt := range []struct {
		name       string
		status     int
		body       string
		errContain string // empty for a success
		retryable  bool
	}{
		{"a server error is retryable", http.StatusInternalServerError, "", "unexpected status code: 500", true},
		{"a client error is fatal", http.StatusNotFound, "", "unexpected status code: 404", false},
		{"an undecodable body is fatal", http.StatusOK, "{invalid json", "failed to decode response body", false},
		{"a result is formatted", http.StatusOK, `{"query":"test","results":[{"title":"Title","url":"https://example.com","content":"Content"}]}`, "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body))}

			result, err := sx.parseHTTPResponse(resp, "test")

			if tt.errContain == "" {
				if err != nil {
					t.Fatalf("parseHTTPResponse() unexpected error: %v", err)
				}
				if !strings.Contains(result, "# Searxng Search Results") || !strings.Contains(result, "Title") {
					t.Errorf("result misses the header or the title: %q", result)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.errContain) {
				t.Fatalf("parseHTTPResponse() error = %v, want it to contain %q", err, tt.errContain)
			}
			if IsRetryable(err) != tt.retryable || IsFatal(err) == tt.retryable {
				t.Errorf("parseHTTPResponse() error = %T, want retryable=%v", err, tt.retryable)
			}
		})
	}
}

func TestSearxng_FormatResults_ListsEveryFieldOrSaysNothingWasFound(t *testing.T) {
	sx := &searxng{}

	for _, tt := range []struct {
		name    string
		results []SearxngResult
		want    []string
	}{
		{"no results name the query", []SearxngResult{}, []string{"No Results Found", "test query"}},
		{
			name: "a result shows every field",
			results: []SearxngResult{{
				Title:         "Test Title",
				URL:           "https://example.com",
				Content:       "Test content",
				Author:        "Test Author",
				PublishedDate: "2024-01-01",
				Engine:        "google",
			}},
			want: []string{"# Searxng Search Results", "Test Title", "https://example.com", "Test content", "Test Author", "2024-01-01", "google"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := sx.formatResults(tt.results, "test query")
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("result misses %q: %q", want, result)
				}
			}
		})
	}
}

func TestSearxng_IsAvailable_RequiresABaseURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when URL is set", &config.Config{SearxngURL: "http://searxng.example.com"}, true},
		{"unavailable when URL is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
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
