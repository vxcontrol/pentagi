package searchers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

// searchersDuckduckgoPage renders n result blocks the way html.duckduckgo.com lays them out.
func searchersDuckduckgoPage(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, `<div class="result results_links results_links_deep web-result">
	<div class="links_main links_deep result__body">
		<h2 class="result__title"><a rel="nofollow" class="result__a" href="https://example.com/%d">Result %d</a></h2>
		<a class="result__snippet" href="https://example.com/%d">Description %d</a>
		<div class="clear"></div>
	</div>
</div>
`, i, i, i, i)
	}
	return b.String()
}

func TestDuckduckgo_Handle_PostsTheFormAndReturnsAtMostMaxResults(t *testing.T) {
	up := &searchersUpstream{body: searchersDuckduckgoPage(12)}
	p := newTestProxy(t, "html.duckduckgo.com", up)
	ddg := NewDuckDuckGo(&config.Config{
		DuckDuckGoEnabled:    true,
		DuckDuckGoSafeSearch: DuckDuckGoSafeSearchModerate,
		ProxyURL:             p.URL(),
		ExternalSSLCAPath:    p.CACertPath(),
	})
	if ddg.Engine() != database.SearchengineTypeDuckduckgo {
		t.Errorf("Engine() = %q, want %q", ddg.Engine(), database.SearchengineTypeDuckduckgo)
	}

	for _, tt := range []struct {
		name       string
		maxResults int
		want       int
	}{
		{"a limit below ten is kept", 5, 5},
		{"a limit of ten is kept", 10, 10},
		{"a larger limit is cut to ten", 100, 10},
		{"no limit means ten", 0, 10},
		{"a negative limit means ten", -5, 10},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ddg.Handle(t.Context(), Request{Query: "test query", MaxResults: tt.maxResults})
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			if up.method != http.MethodPost || up.path != "/html/" {
				t.Errorf("request = %s %s, want POST /html/", up.method, up.path)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", ct)
			}
			if ua := up.header.Get("User-Agent"); !strings.Contains(ua, "Mozilla") {
				t.Errorf("User-Agent = %q, want to contain Mozilla", ua)
			}
			if accept := up.header.Get("Accept"); !strings.Contains(accept, "text/html") {
				t.Errorf("Accept = %q, want to contain text/html", accept)
			}
			form, err := url.ParseQuery(string(up.payload))
			if err != nil || form.Get("q") != "test query" || form.Get("kl") != "us-en" || form.Get("kp") != "0" {
				t.Errorf("request form = %q, want q=test query, kl=us-en, kp=0", up.payload)
			}

			if n := strings.Count(got, "## URL\n"); n != tt.want {
				t.Errorf("Handle() returned %d results, want %d:\n%s", n, tt.want, got)
			}
			if !strings.Contains(got, "# 1. Result 1\n\n## URL\nhttps://example.com/1\n\n## Description\n\nDescription 1") {
				t.Errorf("Handle() result misses the first result's title, URL and description:\n%s", got)
			}
		})
	}

	t.Run("a page without results says so", func(t *testing.T) {
		up.reset(http.StatusOK, "<div>No results</div>")
		if got, err := ddg.Handle(t.Context(), Request{Query: "test query", MaxResults: 5}); err != nil || got != "No results found" {
			t.Errorf("Handle() = %q, %v; want %q", got, err, "No results found")
		}
	})
}

func TestDuckduckgo_Handle_GivesUpFatallyAfterThreeRejectedAttempts(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
	}{
		{"a server error is fatal once retried", http.StatusInternalServerError},
		{"a client error is fatal once retried", http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up := &searchersUpstream{status: tt.status}
			p := newTestProxy(t, "html.duckduckgo.com", up)
			ddg := NewDuckDuckGo(&config.Config{
				DuckDuckGoEnabled: true,
				ProxyURL:          p.URL(),
				ExternalSSLCAPath: p.CACertPath(),
			})

			got, err := ddg.Handle(t.Context(), Request{Query: "test", MaxResults: 5})

			if got != "" || !IsFatal(err) {
				t.Fatalf("Handle() = %q, %v; want no result and a FatalError", got, err)
			}
			if want := fmt.Sprintf("unexpected status code: %d", tt.status); !strings.Contains(err.Error(), want) {
				t.Errorf("Handle() error = %v, want it to contain %q", err, want)
			}
			if up.hits != 3 {
				t.Errorf("upstream saw %d attempts, want 3", up.hits)
			}
		})
	}
}

// Subtests are keyed by parser: parseHTMLStructured, and parseHTMLRegex that parseHTMLResponse falls back to.
func TestDuckduckgo_ParseHTML_FindsEveryResultBlock(t *testing.T) {
	type page struct {
		name string
		body []byte
		want []searchResult
	}
	pages := []page{{
		name: "adjacent blocks stay apart",
		body: []byte(searchersDuckduckgoPage(2)),
		want: []searchResult{
			{Title: "Result 1", URL: "https://example.com/1", Description: "Description 1"},
			{Title: "Result 2", URL: "https://example.com/2", Description: "Description 2"},
		},
	}}
	for _, file := range []string{
		"ddg_result_golang_http_client.html",
		"ddg_result_site_github_golang.html",
		"ddg_result_owasp_vulnerabilities.html",
		"ddg_result_sql_injection.html",
		"ddg_result_docker_security.html",
	} {
		body, err := os.ReadFile(filepath.Join("testdata", file))
		if err != nil {
			t.Fatalf("failed to read test data: %v", err)
		}
		pages = append(pages, page{name: "ten results in " + file, body: body})
	}

	ddg := &duckduckgo{}
	for _, pg := range pages {
		for _, parser := range []struct {
			name  string
			parse func([]byte) ([]searchResult, error)
		}{
			{"structured parser", ddg.parseHTMLStructured},
			{"regex parser", ddg.parseHTMLRegex},
		} {
			t.Run(parser.name+" finds "+pg.name, func(t *testing.T) {
				got, err := parser.parse(pg.body)
				if err != nil {
					t.Fatalf("parse failed: %v", err)
				}
				if pg.want != nil {
					if !reflect.DeepEqual(got, pg.want) {
						t.Errorf("parsed %+v, want %+v", got, pg.want)
					}
					return
				}
				if len(got) != 10 {
					t.Fatalf("expected 10 results, got %d", len(got))
				}
				for i, r := range got {
					if r.Title == "" || r.URL == "" || r.Description == "" {
						t.Errorf("result %d has an empty field: %+v", i, r)
					}
				}
			})
		}
	}
}

func TestDuckduckgo_CleanText_StripsTagsAndDecodesEntities(t *testing.T) {
	ddg := &duckduckgo{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "html tags are dropped",
			input:    "This is <b>bold</b> text",
			expected: "This is bold text",
		},
		{
			name:     "a named apostrophe entity is decoded",
			input:    "Go&#x27;s http package",
			expected: "Go's http package",
		},
		{
			name:     "several entities are decoded",
			input:    "&quot;Hello&quot; &amp; &lt;goodbye&gt;",
			expected: "\"Hello\" & <goodbye>",
		},
		{
			name:     "numeric entities below 128 are decoded and others kept",
			input:    "&#x41;&#66; &#x263A;",
			expected: "AB &#x263A;",
		},
		{
			name:     "whitespace is collapsed",
			input:    "Multiple   spaces   and\n\nnewlines",
			expected: "Multiple spaces and newlines",
		},
		{
			name:     "tags and entities together",
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

func TestDuckduckgo_FormatSearchResults_NumbersResultsAndSeparatesThem(t *testing.T) {
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

func TestDuckduckgo_BuildFormData_MapsTheSettingsToFormFields(t *testing.T) {
	for _, tt := range []struct {
		name string
		cfg  config.Config
		want url.Values
	}{
		{
			name: "unset settings search the US without safe search or time range",
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {""}, "kl": {"us-en"}},
		},
		{
			name: "a region is passed on",
			cfg:  config.Config{DuckDuckGoRegion: RegionDE},
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {""}, "kl": {"de-de"}},
		},
		{
			name: "strict safe search is 1",
			cfg:  config.Config{DuckDuckGoSafeSearch: DuckDuckGoSafeSearchStrict},
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {""}, "kl": {"us-en"}, "kp": {"1"}},
		},
		{
			name: "moderate safe search is 0",
			cfg:  config.Config{DuckDuckGoSafeSearch: DuckDuckGoSafeSearchModerate},
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {""}, "kl": {"us-en"}, "kp": {"0"}},
		},
		{
			name: "safe search off is -1",
			cfg:  config.Config{DuckDuckGoSafeSearch: DuckDuckGoSafeSearchOff},
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {""}, "kl": {"us-en"}, "kp": {"-1"}},
		},
		{
			name: "a time range fills df",
			cfg:  config.Config{DuckDuckGoTimeRange: TimeRangeWeek},
			want: url.Values{"q": {"test query"}, "b": {""}, "df": {"w"}, "kl": {"us-en"}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := url.ParseQuery((&duckduckgo{cfg: &tt.cfg}).buildFormData("test query"))
			if err != nil {
				t.Fatalf("buildFormData() is not a form: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildFormData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuckduckgo_Region_DefaultsToTheUSWithoutAConfig(t *testing.T) {
	if got := (&duckduckgo{}).region(); got != "us-en" {
		t.Errorf("region() with no config = %q, want %q", got, "us-en")
	}
}

func TestDuckduckgo_IsAvailable_RequiresTheEnabledFlag(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when enabled", &config.Config{DuckDuckGoEnabled: true}, true},
		{"unavailable when disabled", &config.Config{DuckDuckGoEnabled: false}, false},
		{"unavailable when nil config", nil, false},
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
