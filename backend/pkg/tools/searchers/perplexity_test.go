package searchers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

const testPerplexityChatBody = `{
	"id":"test-id",
	"model":"sonar",
	"choices":[
		{
			"index":0,
			"finish_reason":"stop",
			"message":{"role":"assistant","content":"This is a test answer."}
		}
	],
	"search_results":[
		{"title":"Example","url":"https://example.com","date":"2026-09-01"},
		{"title":"Test","url":"https://test.com","date":"2026-09-02"}
	],
	"citations":["https://example.com","https://test.com"]
}`

func TestPerplexity_Handle_PostsAChatCompletionAndFormatsItsAnswer(t *testing.T) {
	up := &searchersUpstream{body: testPerplexityChatBody}
	p := newTestProxy(t, "api.perplexity.ai", up)

	for _, tt := range []struct {
		name string
		cfg  config.Config
		want map[string]any
	}{
		{
			name: "the configured model and context size are sent",
			cfg:  config.Config{PerplexityModel: "sonar-pro", PerplexityContextSize: "high"},
			want: map[string]any{
				"model":              "sonar-pro",
				"messages":           []any{map[string]any{"role": "user", "content": "test query"}},
				"web_search_options": map[string]any{"search_context_size": "high"},
			},
		},
		{
			name: "an unset model defaults to sonar and the context size defaults to low",
			want: map[string]any{
				"model":              "sonar",
				"messages":           []any{map[string]any{"role": "user", "content": "test query"}},
				"web_search_options": map[string]any{"search_context_size": "low"},
			},
		},
		{
			name: "a slashed gateway model name is sent verbatim",
			cfg:  config.Config{PerplexityModel: "openrouter/perplexity/sonar", PerplexityContextSize: "medium"},
			want: map[string]any{
				"model":              "openrouter/perplexity/sonar",
				"messages":           []any{map[string]any{"role": "user", "content": "test query"}},
				"web_search_options": map[string]any{"search_context_size": "medium"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.PerplexityAPIKey = "test-key"
			tt.cfg.ProxyURL, tt.cfg.ExternalSSLCAPath = p.URL(), p.CACertPath()
			px := NewPerplexity(&tt.cfg, nil)

			got, err := px.Handle(t.Context(), Request{Query: "test query"})
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if px.Engine() != database.SearchengineTypePerplexity {
				t.Errorf("Engine() = %q, want %q", px.Engine(), database.SearchengineTypePerplexity)
			}

			if up.method != http.MethodPost || up.path != "/chat/completions" {
				t.Errorf("request = %s %s, want POST /chat/completions", up.method, up.path)
			}
			if auth := up.header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Authorization = %q, want Bearer test-key", auth)
			}
			if ct := up.header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var sent map[string]any
			if err := json.Unmarshal(up.payload, &sent); err != nil || !reflect.DeepEqual(sent, tt.want) {
				t.Errorf("request body = %s, want %v", up.payload, tt.want)
			}

			for _, want := range []string{"# Answer", "This is a test answer.", "# Citations", "1. https://example.com"} {
				if !strings.Contains(got, want) {
					t.Errorf("result misses %q: %q", want, got)
				}
			}
		})
	}
}

func TestPerplexity_Search_RetriesOnlyTransientFailures(t *testing.T) {
	up := &searchersUpstream{}
	p := newTestProxy(t, "api.perplexity.ai", up)
	px := NewPerplexity(&config.Config{
		PerplexityAPIKey:  "test-key",
		ProxyURL:          p.URL(),
		ExternalSSLCAPath: p.CACertPath(),
	}, nil)

	for _, tt := range []struct {
		name       string
		status     int
		body       string
		retryable  bool
		errContain string
	}{
		{"a rejected key is fatal and carries the api's reason", http.StatusUnauthorized, `{"error":{"message":"invalid api key","type":"auth"}}`, false, "API key is wrong: invalid api key"},
		{"a rejected request is fatal and carries the api's reason", http.StatusBadRequest, `{"error":{"message":"invalid request","type":"invalid_request"}}`, false, "request is invalid: invalid request"},
		{"a rate limit is retryable", http.StatusTooManyRequests, "", true, "too many"},
		{"a server error is retryable", http.StatusInternalServerError, "", true, "server"},
		{"a completion that carries no message is fatal", http.StatusOK, `{"id":"x","choices":[]}`, false, "carries no answer"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			up.reset(tt.status, tt.body)

			got, err := px.Handle(t.Context(), Request{Query: "q"})

			if up.hits != 1 {
				t.Fatalf("upstream saw %d requests, want 1", up.hits)
			}
			if got != "" || err == nil {
				t.Fatalf("Handle() = %q, %v; web_search would sell this as the answer", got, err)
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

// A retry after our own deadline spends the budget again on a request that is still going.
func TestPerplexity_Search_TreatsItsOwnDeadlineAsFatal(t *testing.T) {
	t.Parallel()

	p := newTestProxy(t, "api.perplexity.ai", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(3 * time.Second):
		case <-r.Context().Done():
		}
	}))
	cfg := &config.Config{
		PerplexityAPIKey:  "test-key",
		PerplexityTimeout: 1,
		ProxyURL:          p.URL(),
		ExternalSSLCAPath: p.CACertPath(),
	}

	_, err := NewPerplexity(cfg, nil).Handle(t.Context(), Request{Query: "test query"})

	if err == nil {
		t.Fatal("want an error when the budget runs out")
	}
	if IsRetryable(err) {
		t.Errorf("a deadline of our own must not be retried against the same engine: %v", err)
	}
	if !IsFatal(err) {
		t.Errorf("want the orchestrator to move on, got %v", err)
	}
}

func TestPerplexity_HandleErrorResponse_NamesEachRejectedStatus(t *testing.T) {
	px := &perplexity{}

	tests := []struct {
		name       string
		statusCode int
		errContain string
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, errContain: "invalid"},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, errContain: "API key"},
		{name: "forbidden", statusCode: http.StatusForbidden, errContain: "administrators"},
		{name: "not found", statusCode: http.StatusNotFound, errContain: "not be found"},
		{name: "method not allowed", statusCode: http.StatusMethodNotAllowed, errContain: "invalid method"},
		{name: "too many requests", statusCode: http.StatusTooManyRequests, errContain: "too many"},
		{name: "internal server error", statusCode: http.StatusInternalServerError, errContain: "server"},
		{name: "bad gateway", statusCode: http.StatusBadGateway, errContain: "server"},
		{name: "service unavailable", statusCode: http.StatusServiceUnavailable, errContain: "maintenance"},
		{name: "gateway timeout", statusCode: http.StatusGatewayTimeout, errContain: "maintenance"},
		{name: "unexpected", statusCode: http.StatusTeapot, errContain: "unexpected status code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := px.handleErrorResponse(tt.statusCode)
			if err == nil {
				t.Fatalf("handleErrorResponse(%d) = nil, want an error", tt.statusCode)
			}
			if !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("handleErrorResponse(%d) = %v, expected to contain %q", tt.statusCode, err, tt.errContain)
			}
		})
	}
}

func TestPerplexity_FormatResponse_AnswersWithNumberedCitations(t *testing.T) {
	px := &perplexity{}

	message := func(text string) chatResponse {
		return chatResponse{Choices: []chatChoice{{Message: chatMessage{Role: "assistant", Content: text}}}}
	}

	t.Run("empty output returns nothing for the caller to refuse", func(t *testing.T) {
		result := px.formatResponse(t.Context(), &chatResponse{}, "test query")
		if result != "" {
			t.Errorf("unexpected result for empty output: %q", result)
		}
	})

	t.Run("message without sources", func(t *testing.T) {
		resp := message("Go is a compiled language.")
		result := px.formatResponse(t.Context(), &resp, "what is Go")
		if !strings.Contains(result, "# Answer") {
			t.Error("result should contain '# Answer' heading")
		}
		if !strings.Contains(result, "Go is a compiled language.") {
			t.Error("result should contain the answer content")
		}
		if strings.Contains(result, "# Citations") {
			t.Error("result should NOT contain citations section when none provided")
		}
	})

	t.Run("search results become numbered citations", func(t *testing.T) {
		resp := message("Go is fast.")
		resp.SearchResults = []chatSearchResult{{URL: "https://go.dev"}, {URL: "https://example.com/go"}}
		result := px.formatResponse(t.Context(), &resp, "test")
		if !strings.Contains(result, "# Citations") {
			t.Error("result should contain '# Citations' heading")
		}
		if !strings.Contains(result, "1. https://go.dev") {
			t.Error("result should contain numbered citations")
		}
		if !strings.Contains(result, "2. https://example.com/go") {
			t.Error("result should contain second citation")
		}
	})

	t.Run("a flat citations list is numbered when there are no search results", func(t *testing.T) {
		resp := message("answer")
		resp.Citations = []string{"https://cited.example"}
		result := px.formatResponse(t.Context(), &resp, "query")
		if !strings.Contains(result, "1. https://cited.example") {
			t.Errorf("citation url missing from citations: %q", result)
		}
	})

	t.Run("a url reported in both fields is cited once", func(t *testing.T) {
		resp := message("answer")
		resp.SearchResults = []chatSearchResult{{URL: "https://dup.example"}}
		resp.Citations = []string{"https://dup.example"}
		result := px.formatResponse(t.Context(), &resp, "query")
		if strings.Count(result, "https://dup.example") != 1 {
			t.Errorf("duplicate url should be cited once: %q", result)
		}
	})

	t.Run("content from every choice is joined", func(t *testing.T) {
		resp := chatResponse{Choices: []chatChoice{
			{Message: chatMessage{Content: "first. "}},
			{Message: chatMessage{Content: "second."}},
		}}
		result := px.formatResponse(t.Context(), &resp, "query")
		if !strings.Contains(result, "first. second.") {
			t.Errorf("both choices' content should appear: %q", result)
		}
	})
}

func TestPerplexity_GetSummarizePrompt_ListsCitationsOnlyWhenThereAreSome(t *testing.T) {
	px := &perplexity{}

	t.Run("prompt without citations", func(t *testing.T) {
		prompt, err := px.getSummarizePrompt("test query", "some content", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "test query") {
			t.Error("prompt should contain the query")
		}
		if !strings.Contains(prompt, "some content") {
			t.Error("prompt should contain the content")
		}
		if strings.Contains(prompt, "</citations>") {
			t.Error("prompt should NOT contain closing </citations> tag when there are none")
		}
	})

	t.Run("prompt with citations", func(t *testing.T) {
		prompt, err := px.getSummarizePrompt("query", "content", []string{"https://a.com", "https://b.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(prompt, "<citations>") {
			t.Error("prompt should contain citations block")
		}
		if !strings.Contains(prompt, "https://a.com") {
			t.Error("prompt should contain first citation")
		}
	})
}

func TestPerplexity_Model_DefaultsToSonarAndKeepsTheConfiguredValue(t *testing.T) {
	for _, tt := range []struct {
		name       string
		configured string
		want       string
	}{
		{"an unset model defaults to sonar", "", "sonar"},
		{"a bare sonar name is kept", "sonar", "sonar"},
		{"a bare sonar-pro name is kept", "sonar-pro", "sonar-pro"},
		{"a bare sonar-reasoning name is kept", "sonar-reasoning", "sonar-reasoning"},
		{"a padded name is kept without the padding", "  sonar-pro  ", "sonar-pro"},
		{"a slashed gateway model name is passed through verbatim", "openrouter/perplexity/sonar", "openrouter/perplexity/sonar"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := &perplexity{cfg: &config.Config{PerplexityModel: tt.configured}}
			if got := p.model(); got != tt.want {
				t.Errorf("model() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("a nil config defaults to sonar", func(t *testing.T) {
		p := &perplexity{}
		if got := p.model(); got != "sonar" {
			t.Errorf("model() = %q, want sonar", got)
		}
	})
}

func TestPerplexity_ReadErrorDetail_KeepsTheAPIsReasonOrOneShortLine(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
		want string
	}{
		{
			name: "the api's own message is preferred",
			body: `{"error":{"message":"validation failed: model \"sonar-pro\" is not supported","type":"invalid_request","code":400}}`,
			want: `validation failed: model "sonar-pro" is not supported`,
		},
		{name: "a plain body is the detail", body: "upstream exploded", want: "upstream exploded"},
		{name: "an empty body has no detail", body: "", want: ""},
		{
			name: "an html error page is left out",
			body: "<!DOCTYPE html><html><head><title>502 Bad Gateway</title></head><body>" +
				strings.Repeat("<p>cloudflare</p>", 500) + "</body></html>",
			want: "",
		},
		{
			name: "plain text keeps its first line, cut at 200 runes",
			body: strings.Repeat("x", 5000) + "\nsecond line",
			want: strings.Repeat("x", 200) + "…",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := readErrorDetail(strings.NewReader(tt.body)); got != tt.want {
				t.Errorf("readErrorDetail() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A web-grounded completion can outlive a plain one, and an operator has to be able to move the budget without a rebuild.
func TestPerplexity_Timeout_FollowsTheConfiguredValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		cfg  *config.Config
		want time.Duration
	}{
		{"an unset or zero budget is two minutes", &config.Config{}, 120 * time.Second},
		{"a negative budget is two minutes", &config.Config{PerplexityTimeout: -5}, 120 * time.Second},
		{"a configured budget is kept", &config.Config{PerplexityTimeout: 45}, 45 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			px := &perplexity{cfg: tc.cfg}
			if got := px.timeout(); got != tc.want {
				t.Errorf("timeout() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPerplexity_IsAvailable_RequiresAnAPIKey(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"available when API key is set", &config.Config{PerplexityAPIKey: "test-key"}, true},
		{"unavailable when API key is empty", &config.Config{}, false},
		{"unavailable when nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			px := &perplexity{cfg: tt.cfg}
			if got := px.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
