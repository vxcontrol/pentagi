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

func TestInternalIsAvailable(t *testing.T) {
	fetcher := &fakeFetcher{}
	sum := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })
	link := &fakeLinkSearcher{available: true}

	t.Run("disabled by config", func(t *testing.T) {
		e := NewInternal(&config.Config{WebSearchInternalEnabled: false}, fetcher, []Searcher{link}, sum)
		if e.IsAvailable() {
			t.Error("internal engine must be unavailable when WEB_SEARCH_INTERNAL_ENABLED=false")
		}
	})
	t.Run("no fetcher", func(t *testing.T) {
		e := NewInternal(enabledInternalCfg(), nil, []Searcher{link}, sum)
		if e.IsAvailable() {
			t.Error("internal engine must be unavailable without a page fetcher")
		}
	})
	t.Run("no available link engine", func(t *testing.T) {
		e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{&fakeLinkSearcher{available: false}}, sum)
		if e.IsAvailable() {
			t.Error("internal engine must be unavailable when no link engine is available")
		}
	})
	t.Run("fully configured", func(t *testing.T) {
		e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{link}, sum)
		if !e.IsAvailable() {
			t.Error("internal engine should be available when enabled + fetcher + link engine present")
		}
		if e.Engine() != database.SearchengineTypeBrowser {
			t.Errorf("Engine() = %q, want browser (attribution)", e.Engine())
		}
	})
}

func TestInternalAnalyzeSummarizesBoundedContent(t *testing.T) {
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
	// The fetched source URLs must be appended as a numbered Sources section, in fetch
	// order, matching the [Source N] ids the summarizer was given.
	if !strings.Contains(got, "### Sources") {
		t.Errorf("result = %q, want an appended '### Sources' section", got)
	}
	if !strings.Contains(got, "1. [a.example](https://a.example)") ||
		!strings.Contains(got, "2. [b.example](https://b.example)") {
		t.Errorf("result = %q, want the fetched URLs as numbered markdown links in fetch order", got)
	}
	// Per-site markdown must be truncated to MaxSiteBytes (20).
	if strings.Contains(captured, strings.Repeat("A", 21)) {
		t.Error("page A markdown was not truncated to the per-site byte limit")
	}
	if !strings.Contains(captured, "https://a.example") || !strings.Contains(captured, "https://b.example") {
		t.Error("summarizer prompt should reference both source URLs")
	}
}

func TestAppendSources(t *testing.T) {
	if got := appendSources("answer", nil); got != "answer" {
		t.Errorf("appendSources with no URLs = %q, want the answer unchanged", got)
	}

	// Each source is rendered as a numbered markdown link whose text is the URL host.
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

func TestInternalRespectsMaxSites(t *testing.T) {
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

func TestInternalAllPagesFailIsRetryable(t *testing.T) {
	link := &fakeLinkSearcher{available: true, output: "https://a.example https://b.example"}
	fetcher := &fakeFetcher{errs: map[string]error{
		"https://a.example": errors.New("boom"),
		"https://b.example": errors.New("boom"),
	}}
	sum := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })

	e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{link}, sum)
	_, err := e.Handle(context.Background(), Request{Query: "q"})
	if err == nil {
		t.Fatal("expected an error when every page fails to fetch")
	}
	if !IsRetryable(err) {
		t.Errorf("error = %v, want retryable when all pages fail", err)
	}
}

func TestInternalSummarizerFailureIsRetryable(t *testing.T) {
	link := &fakeLinkSearcher{available: true, output: "https://a.example"}
	fetcher := &fakeFetcher{pages: map[string]string{"https://a.example": "content"}}
	sum := SummarizeHandler(func(_ context.Context, _ string) (string, error) { return "", errors.New("llm down") })

	e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{link}, sum)
	_, err := e.Handle(context.Background(), Request{Query: "q"})
	if err == nil || !IsRetryable(err) {
		t.Fatalf("expected a retryable error on summarizer failure, got %v", err)
	}
}

func TestInternalLinkDiscoveryErrorPropagates(t *testing.T) {
	link := &fakeLinkSearcher{available: true, err: Fatal(errors.New("link engine broke"))}
	fetcher := &fakeFetcher{}
	sum := SummarizeHandler(func(_ context.Context, s string) (string, error) { return s, nil })

	e := NewInternal(enabledInternalCfg(), fetcher, []Searcher{link}, sum)
	_, err := e.Handle(context.Background(), Request{Query: "q"})
	if err == nil || !IsFatal(err) {
		t.Fatalf("expected the link engine's fatal error to propagate, got %v", err)
	}
}

func TestDedupeURLs(t *testing.T) {
	in := []string{"https://a.com", "https://a.com", "https://b.com.", "https://a.com,"}
	got := dedupeURLs(in)
	// Trailing punctuation stripped, duplicates removed.
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
