package searchers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

// fakeFetcher scripts FetchMarkdown responses by URL.
type fakeFetcher struct {
	pages map[string]string
	errs  map[string]error
	calls int
}

func (f *fakeFetcher) FetchMarkdown(_ context.Context, url string) (string, error) {
	f.calls++
	if err, ok := f.errs[url]; ok {
		return "", err
	}
	return f.pages[url], nil
}

// fakeLinkSearcher returns a fixed markdown blob (with URLs) as a link engine would.
type fakeLinkSearcher struct {
	available bool
	output    string
	err       error
}

func (f *fakeLinkSearcher) Engine() database.SearchengineType { return database.SearchengineTypeGoogle }
func (f *fakeLinkSearcher) IsAvailable() bool                 { return f.available }
func (f *fakeLinkSearcher) Handle(_ context.Context, _ Request) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.output, nil
}

func enabledInternalCfg() *config.Config {
	return &config.Config{
		WebSearchInternalEnabled:      true,
		WebSearchInternalMaxSites:     3,
		WebSearchInternalMaxSiteBytes: 20,
	}
}

func TestInternal_IsAvailable_NeedsTheFlagAFetcherAndAnAvailableLinkEngine(t *testing.T) {
	sum := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })
	link := &fakeLinkSearcher{available: true}

	for _, tt := range []struct {
		name    string
		cfg     *config.Config
		fetcher PageFetcher
		links   []Searcher
		want    bool
	}{
		{"disabled by config", &config.Config{WebSearchInternalEnabled: false}, &fakeFetcher{}, []Searcher{link}, false},
		{"no fetcher", enabledInternalCfg(), nil, []Searcher{link}, false},
		{"no available link engine", enabledInternalCfg(), &fakeFetcher{}, []Searcher{&fakeLinkSearcher{available: false}}, false},
		{"fully configured", enabledInternalCfg(), &fakeFetcher{}, []Searcher{link}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInternal(tt.cfg, tt.fetcher, tt.links, sum).IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}

	if got := NewInternal(enabledInternalCfg(), &fakeFetcher{}, []Searcher{link}, sum).Engine(); got != database.SearchengineTypeBrowser {
		t.Errorf("Engine() = %q, want browser (attribution)", got)
	}
}

func TestInternal_Analyze_SummarizesBoundedPagesAndListsTheirSources(t *testing.T) {
	link := &fakeLinkSearcher{available: true, output: "1. https://a.example\n2. https://b.example\n"}
	fetcher := &fakeFetcher{pages: map[string]string{
		"https://a.example": strings.Repeat("A", 100), // longer than MaxSiteBytes=20
		"https://b.example": "short B content",
	}}

	var captured string
	sum := SummarizeHandler(func(_ context.Context, prompt string) (string, error) {
		captured = prompt
		return "SYNTHESIZED ANSWER", nil
	})

	e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{link}, sum)
	got, err := e.Handle(context.Background(), Request{Query: "what is A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "SYNTHESIZED ANSWER") {
		t.Errorf("result = %q, want it to contain the summarizer output", got)
	}
	// Numbered in fetch order, matching the [Source N] ids the summarizer was given.
	if !strings.Contains(got, "### Sources") {
		t.Errorf("result = %q, want an appended '### Sources' section", got)
	}
	if !strings.Contains(got, "1. [a.example](https://a.example)") ||
		!strings.Contains(got, "2. [b.example](https://b.example)") {
		t.Errorf("result = %q, want the fetched URLs as numbered markdown links in fetch order", got)
	}
	if strings.Contains(captured, strings.Repeat("A", 21)) {
		t.Error("page A markdown was not truncated to the per-site byte limit")
	}
	if !strings.Contains(captured, "https://a.example") || !strings.Contains(captured, "https://b.example") {
		t.Error("summarizer prompt should reference both source URLs")
	}
}

func TestInternal_Analyze_StopsFetchingAtMaxSites(t *testing.T) {
	cfg := enabledInternalCfg()
	cfg.WebSearchInternalMaxSites = 1
	link := &fakeLinkSearcher{available: true, output: "https://a.example https://b.example https://c.example"}
	fetcher := &fakeFetcher{pages: map[string]string{
		"https://a.example": "A", "https://b.example": "B", "https://c.example": "C",
	}}
	sum := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })

	e := NewInternal(cfg, fetcher, []Searcher{link}, sum)
	if _, err := e.Handle(context.Background(), Request{Query: "q"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetcher.calls != 1 {
		t.Errorf("fetched %d pages, want 1 (MaxSites cap)", fetcher.calls)
	}
}

func TestInternal_Analyze_RetriesOnlyTransientFailures(t *testing.T) {
	fetchErr := errors.New("page unreachable")
	llmErr := errors.New("llm down")
	linkErr := Fatal(errors.New("link engine broke"))
	echo := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })

	for _, tt := range []struct {
		name      string
		link      *fakeLinkSearcher
		fetcher   *fakeFetcher
		sum       SummarizeHandler
		retryable bool
		cause     error
		reason    string
	}{
		{
			name:      "every page failing to fetch is retryable",
			link:      &fakeLinkSearcher{available: true, output: "https://a.example https://b.example"},
			fetcher:   &fakeFetcher{errs: map[string]error{"https://a.example": fetchErr, "https://b.example": fetchErr}},
			sum:       echo,
			retryable: true,
			cause:     fetchErr,
			reason:    "all pages failed to fetch",
		},
		{
			name:      "a failed summary is retryable",
			link:      &fakeLinkSearcher{available: true, output: "https://a.example"},
			fetcher:   &fakeFetcher{pages: map[string]string{"https://a.example": "content"}},
			sum:       func(context.Context, string) (string, error) { return "", llmErr },
			retryable: true,
			cause:     llmErr,
			reason:    "summarization failed",
		},
		{
			name:    "the link engine's own error is passed on",
			link:    &fakeLinkSearcher{available: true, err: linkErr},
			fetcher: &fakeFetcher{},
			sum:     echo,
			cause:   linkErr,
		},
		{
			name:    "links without a url are fatal",
			link:    &fakeLinkSearcher{available: true, output: "nothing to follow here"},
			fetcher: &fakeFetcher{},
			sum:     echo,
			reason:  "no usable URLs",
		},
		{
			name:    "pages without content are fatal",
			link:    &fakeLinkSearcher{available: true, output: "https://a.example"},
			fetcher: &fakeFetcher{pages: map[string]string{"https://a.example": "  \n"}},
			sum:     echo,
			reason:  "no readable content",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := NewInternal(enabledInternalCfg(), tt.fetcher, []Searcher{tt.link}, tt.sum)

			got, err := e.Handle(t.Context(), Request{Query: "q"})

			if got != "" || err == nil {
				t.Fatalf("Handle() = %q, %v; want no result and an error", got, err)
			}
			if IsRetryable(err) != tt.retryable || IsFatal(err) == tt.retryable {
				t.Errorf("Handle() error = %v, want retryable=%v", err, tt.retryable)
			}
			if tt.cause != nil && !errors.Is(err, tt.cause) {
				t.Errorf("Handle() error = %v, want it to carry %v", err, tt.cause)
			}
			if !strings.Contains(err.Error(), tt.reason) {
				t.Errorf("Handle() error = %v, want it to say %q", err, tt.reason)
			}
		})
	}
}

func TestInternal_AppendSources_ListsEachSourceHostAsANumberedLink(t *testing.T) {
	if got := appendSources("answer", nil); got != "answer" {
		t.Errorf("appendSources with no URLs = %q, want the answer unchanged", got)
	}

	got := appendSources("answer\n\n", []string{
		"https://github.com/vxcontrol/pentagi",
		"https://pentagi.com/get-started",
	})
	want := "answer\n\n### Sources\n\n" +
		"1. [github.com](https://github.com/vxcontrol/pentagi)\n" +
		"2. [pentagi.com](https://pentagi.com/get-started)\n"
	if got != want {
		t.Errorf("appendSources = %q, want %q", got, want)
	}
}

func TestInternal_DedupeURLs_DropsTrailingPunctuationAndRepeats(t *testing.T) {
	in := []string{"https://a.com", "https://a.com", "https://b.com.", "https://a.com,"}
	got := dedupeURLs(in)
	want := []string{"https://a.com", "https://b.com"}
	if len(got) != len(want) {
		t.Fatalf("dedupeURLs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dedupeURLs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
