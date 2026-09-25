package searchers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

func TestSploitus_Handle_PostsTheSearchAndFormatsTheExploits(t *testing.T) {
	up := &searchersUpstream{body: `{
		"exploits":[{
			"id":"CVE-2024-1234",
			"title":"Test Exploit for nginx",
			"type":"githubexploit",
			"href":"https://github.com/test/exploit",
			"score":9.8,
			"published":"2024-01-15",
			"language":"python",
			"source":"exploit code here"
		}],
		"exploits_total":42
	}`}
	p := newTestProxy(t, "sploitus.com", up)
	sp := NewSploitus(&config.Config{SploitusEnabled: true, ProxyURL: p.URL(), ExternalSSLCAPath: p.CACertPath()})

	got, err := sp.Handle(t.Context(), Request{Query: "nginx", ExploitType: "exploits", Sort: "date", MaxResults: 5})
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if sp.Engine() != database.SearchengineTypeSploitus {
		t.Errorf("Engine() = %q, want %q", sp.Engine(), database.SearchengineTypeSploitus)
	}

	if up.method != http.MethodPost || up.path != "/search" {
		t.Errorf("request = %s %s, want POST /search", up.method, up.path)
	}
	for header, want := range map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"Origin":       "https://sploitus.com",
		"Referer":      "https://sploitus.com/?query=nginx",
	} {
		if got := up.header.Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	if ua := up.header.Get("User-Agent"); !strings.Contains(ua, "Mozilla") {
		t.Errorf("User-Agent = %q, want to contain Mozilla", ua)
	}
	if want := `{"query":"nginx","type":"exploits","sort":"date","title":false,"offset":0}`; string(up.payload) != want {
		t.Errorf("request body = %s, want %s", up.payload, want)
	}

	for _, want := range []string{
		"# Sploitus Search Results",
		"**Query:** `nginx`",
		"**Total matches on Sploitus:** 42",
		"Test Exploit for nginx",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("result misses %q: %q", want, got)
		}
	}
}

func TestSploitus_Handle_NormalisesTypeSortAndLimit(t *testing.T) {
	exploits := make([]string, 30)
	for i := range exploits {
		exploits[i] = fmt.Sprintf(`{"id":"E-%d","title":"Exploit %d","href":"https://example.com/%d"}`, i, i, i)
	}
	up := &searchersUpstream{body: `{"exploits":[` + strings.Join(exploits, ",") + `],"exploits_total":30}`}
	p := newTestProxy(t, "sploitus.com", up)
	sp := NewSploitus(&config.Config{SploitusEnabled: true, ProxyURL: p.URL(), ExternalSSLCAPath: p.CACertPath()})

	for _, tt := range []struct {
		name        string
		req         Request
		wantPayload string
		wantShown   int
	}{
		{
			name:        "an unset type, sort and limit take the defaults",
			req:         Request{Query: "test"},
			wantPayload: `{"query":"test","type":"exploits","sort":"default","title":false,"offset":0}`,
			wantShown:   10,
		},
		{
			name:        "padded upper-case hints are normalised and an oversized limit falls back to ten",
			req:         Request{Query: "test", ExploitType: " Tools ", Sort: " DATE ", MaxResults: 100},
			wantPayload: `{"query":"test","type":"tools","sort":"date","title":false,"offset":0}`,
			wantShown:   10,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sp.Handle(t.Context(), tt.req)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if string(up.payload) != tt.wantPayload {
				t.Errorf("request body = %s, want %s", up.payload, tt.wantPayload)
			}
			if n := strings.Count(got, "### "); n != tt.wantShown {
				t.Errorf("Handle() showed %d exploits, want %d", n, tt.wantShown)
			}
		})
	}
}

func TestSploitus_Search_RetriesRateLimitsAndServerErrorsOnly(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "sploitus.com", up)
	sp := NewSploitus(&config.Config{SploitusEnabled: true, ProxyURL: p.URL(), ExternalSSLCAPath: p.CACertPath()})

	for _, tt := range []struct {
		name       string
		status     int
		retryable  bool
		errContain string
	}{
		{"a 499 rate limit is retryable", 499, true, "rate limit exceeded"},
		{"a 422 rate limit is retryable", http.StatusUnprocessableEntity, true, "rate limit exceeded"},
		{"a server error is retryable", http.StatusInternalServerError, true, "HTTP 500"},
		{"a refusal is fatal", http.StatusForbidden, false, "HTTP 403"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, "")

			got, err := sp.Handle(t.Context(), Request{Query: "test", ExploitType: "exploits"})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if got != "" || err == nil {
				t.Fatalf("Handle() = %q, %v; want no result and an error", got, err)
			}
			if IsRetryable(err) != tt.retryable || IsFatal(err) == tt.retryable {
				t.Errorf("Handle() error = %v, want retryable=%v", err, tt.retryable)
			}
			if !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("Handle() error = %v, expected to contain %q", err, tt.errContain)
			}
		})
	}
}

func TestSploitus_FormatSploitusResults_RendersEachTypeUpToTheLimit(t *testing.T) {
	thirty := sploitusResponse{Exploits: make([]sploitusExploit, 30), ExploitsTotal: 30}
	for i := range thirty.Exploits {
		thirty.Exploits[i] = sploitusExploit{ID: fmt.Sprintf("TEST-%d", i), Title: fmt.Sprintf("Test %d", i), Href: "https://example.com"}
	}

	tests := []struct {
		name        string
		query       string
		exploitType string
		limit       int
		response    sploitusResponse
		expected    []string
		wantItems   int
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
			wantItems: 2,
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
			wantItems: 2,
		},
		{
			name:        "empty results",
			query:       "nonexistent",
			exploitType: "exploits",
			limit:       10,
			response:    sploitusResponse{Exploits: []sploitusExploit{}},
			expected:    []string{"# Sploitus Search Results", "**Query:** `nonexistent`", "No exploits were found"},
		},
		{name: "a limit below the matches cuts them", query: "test", exploitType: "exploits", limit: 5, response: thirty, wantItems: 5},
		{name: "a limit of ten shows ten", query: "test", exploitType: "exploits", limit: 10, response: thirty, wantItems: 10},
		{name: "a limit above the matches shows them all", query: "test", exploitType: "exploits", limit: 100, response: thirty, wantItems: 30},
		{name: "no limit shows ten", query: "test", exploitType: "exploits", limit: 0, response: thirty, wantItems: 10},
		{name: "a negative limit shows ten", query: "test", exploitType: "exploits", limit: -5, response: thirty, wantItems: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSploitusResults(tt.query, tt.exploitType, tt.limit, tt.response)

			for _, expectedStr := range tt.expected {
				if !strings.Contains(result, expectedStr) {
					t.Errorf("expected result to contain %q\nGot:\n%s", expectedStr, result)
				}
			}
			if n := strings.Count(result, "### "); n != tt.wantItems {
				t.Errorf("rendered %d items, want %d", n, tt.wantItems)
			}
		})
	}
}

func TestSploitus_FormatSploitusResults_StaysUnder80KB(t *testing.T) {
	t.Run("source truncation at 50KB", func(t *testing.T) {
		resp := sploitusResponse{
			Exploits: []sploitusExploit{
				{
					ID:     "TEST-1",
					Title:  "Test with large source",
					Href:   "https://example.com",
					Source: strings.Repeat("A", 60*1024),
				},
			},
			ExploitsTotal: 1,
		}

		result := formatSploitusResults("test", "exploits", 10, resp)

		if !strings.Contains(result, "source truncated, exceeded 50 KB limit") {
			t.Error("expected source truncation message for 60 KB source")
		}
		if len(result) > 80*1024 {
			t.Errorf("result size %d exceeds 80 KB limit", len(result))
		}
	})

	t.Run("total size limit at 80KB", func(t *testing.T) {
		results := make([]sploitusExploit, 100)
		for i := range results {
			results[i] = sploitusExploit{
				ID:     fmt.Sprintf("TEST-%d", i),
				Title:  fmt.Sprintf("Test Result %d", i),
				Href:   "https://example.com",
				Source: strings.Repeat("X", 5000),
			}
		}

		result := formatSploitusResults("test", "exploits", 100, sploitusResponse{Exploits: results, ExploitsTotal: 100})

		if len(result) > 80*1024 {
			t.Errorf("result size %d exceeds 80 KB hard limit", len(result))
		}
		if !strings.Contains(result, "Results truncated") {
			t.Error("expected truncation warning when hitting 80 KB limit")
		}
		if count := strings.Count(result, "### "); count >= 100 {
			t.Errorf("expected fewer than 100 results due to size limit, got %d", count)
		}
	})
}

func TestSploitus_IsAvailable_RequiresTheEnabledFlag(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when enabled", &config.Config{SploitusEnabled: true}, true},
		{"unavailable when disabled", &config.Config{SploitusEnabled: false}, false},
		{"unavailable when nil config", nil, false},
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
