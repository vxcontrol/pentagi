package searchers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

const (
	internalDefaultMaxSites     = 5
	internalDefaultMaxSiteBytes = 10 * 1024 // 10 KB of markdown per site
	internalLinkDiscoveryLimit  = 8         // how many candidate links to ask the link engine for
)

// urlPattern extracts candidate source URLs from a link engine's markdown output.
// The link engines format URLs in several shapes ("* URL https://…", "## URL\nhttps://…",
// "[https://…](https://…)"), so a permissive http(s) matcher is the most robust way to
// recover them without coupling this engine to each engine's formatting.
var urlPattern = regexp.MustCompile(`https?://[^\s)\]<>"']+`)

// internalEngine is the optional, opt-in browser-analytics engine. Instead of calling
// a paid analytic API, it discovers links with the first available link searcher,
// fetches each page's main-content markdown through the browser scraper (bounded per
// site), and asks the summarizer to synthesize a query-focused answer. It is disabled
// by default (WEB_SEARCH_INTERNAL_ENABLED) because scraping + summarizing many pages
// can cost more than a purpose-built third-party analytic call.
type internalEngine struct {
	cfg        *config.Config
	fetcher    PageFetcher
	links      []Searcher // link-discovery engines, in orchestrator priority order
	summarizer SummarizeHandler
}

// NewInternal builds the internal analytics engine. fetcher and summarizer are
// injected by the orchestrator (the fetcher wraps the browser tool). links are the
// link-oriented searchers used to discover URLs to read.
func NewInternal(
	cfg *config.Config,
	fetcher PageFetcher,
	links []Searcher,
	summarizer SummarizeHandler,
) Searcher {
	return &internalEngine{
		cfg:        cfg,
		fetcher:    fetcher,
		links:      links,
		summarizer: summarizer,
	}
}

func (e *internalEngine) Engine() database.SearchengineType {
	// The internal engine's work is browser-driven, so it attributes its search-log
	// entry to the existing (otherwise unused) "browser" engine value — no new enum
	// value or DB migration is required.
	return database.SearchengineTypeBrowser
}

func (e *internalEngine) IsAvailable() bool {
	if e.cfg == nil || !e.cfg.WebSearchInternalEnabled {
		return false
	}
	if e.fetcher == nil || e.summarizer == nil {
		return false
	}
	for _, l := range e.links {
		if l != nil && l.IsAvailable() {
			return true
		}
	}
	return false
}

func (e *internalEngine) Handle(ctx context.Context, req Request) (string, error) {
	if !e.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine": "internal",
		"query":  req.Query[:min(len(req.Query), 1000)],
	})

	result, err := e.analyze(ctx, req)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine": "internal",
				"query":  req.Query,
				"error":  err.Error(),
			}),
		)
		obs.LogErrorOrCancel(logger, err, "internal analytics engine failed")
		return "", err
	}

	return result, nil
}

func (e *internalEngine) analyze(ctx context.Context, req Request) (string, error) {
	urls, out, err := e.discoverURLs(ctx, req)
	if err != nil {
		return "", err
	}
	if len(urls) == 0 {
		return "", Fatal(fmt.Errorf("internal engine: link discovery returned no usable URLs"))
	}

	maxSites := e.cfg.WebSearchInternalMaxSites
	if maxSites <= 0 {
		maxSites = internalDefaultMaxSites
	}
	maxBytes := e.cfg.WebSearchInternalMaxSiteBytes
	if maxBytes <= 0 {
		maxBytes = internalDefaultMaxSiteBytes
	}

	var (
		blocks    []string
		usedURLs  []string
		fetched   int
		lastFetch error
	)
	out = strings.TrimSpace(out)
	if out != "" {
		blocks = append(blocks, fmt.Sprintf("<search-results>\n%s\n</search-results>", out))
	}
	for _, u := range urls {
		if fetched >= maxSites {
			break
		}
		md, ferr := e.fetcher.FetchMarkdown(ctx, u)
		if ferr != nil {
			// Best-effort: one page failing is not fatal; remember the last error in
			// case every page fails.
			lastFetch = ferr
			continue
		}
		md = strings.TrimSpace(md)
		if md == "" {
			continue
		}
		if len(md) > maxBytes {
			md = md[:maxBytes]
		}
		// The source id (1-based, in fetch order) is what the summarizer cites as
		// [Source N]; usedURLs tracks the matching URL so it can be listed at the end.
		blocks = append(blocks, fmt.Sprintf("<source id=\"%d\" url=\"%s\">\n%s\n</source>", fetched+1, u, md))
		usedURLs = append(usedURLs, u)
		fetched++
	}

	if fetched == 0 {
		// Nothing could be read. If the pages errored (vs. returned empty), surface a
		// retryable error so the orchestrator can fall back to another engine.
		if lastFetch != nil {
			return "", Retryable(fmt.Errorf("internal engine: all pages failed to fetch: %w", lastFetch), 0)
		}
		return "", Fatal(fmt.Errorf("internal engine: no readable content found"))
	}

	prompt := e.buildPrompt(req.Query, blocks)
	answer, serr := e.summarizer(ctx, prompt)
	if serr != nil {
		return "", Retryable(fmt.Errorf("internal engine: summarization failed: %w", serr), 0)
	}

	return appendSources(answer, usedURLs), nil
}

// appendSources appends the list of source URLs the answer was synthesized from, so the
// [Source N] citations in the summary can be traced back to the page they came from. The
// numbering matches the source ids fed to the summarizer (1-based, in fetch order).
func appendSources(answer string, urls []string) string {
	if len(urls) == 0 {
		return answer
	}
	var sb strings.Builder
	sb.WriteString(strings.TrimRight(answer, "\n"))
	sb.WriteString("\n\n### Sources\n\n")
	for i, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, u))
			continue
		}
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, parsedURL.Hostname(), u))
	}
	return sb.String()
}

// discoverURLs asks the first available link searcher for candidate URLs and extracts
// them from its markdown output. A discovery error is propagated as-is (already typed).
func (e *internalEngine) discoverURLs(ctx context.Context, req Request) ([]string, string, error) {
	linkReq := Request{Query: req.Query, MaxResults: internalLinkDiscoveryLimit}

	for _, l := range e.links {
		if l == nil || !l.IsAvailable() {
			continue
		}
		out, err := l.Handle(ctx, linkReq)
		if err != nil {
			// Propagate the typed error from the link engine; the orchestrator decides
			// retry-vs-fallback for the internal engine as a whole.
			return nil, "", err
		}
		return dedupeURLs(urlPattern.FindAllString(out, -1)), out, nil
	}

	return nil, "", Fatal(fmt.Errorf("internal engine: no link-discovery engine is available"))
}

func (e *internalEngine) buildPrompt(query string, blocks []string) string {
	var sb strings.Builder
	sb.WriteString("<instructions>\n")
	sb.WriteString("TASK: Answer the user query using ONLY the web sources below.\n\n")
	sb.WriteString(fmt.Sprintf("USER QUERY: %q\n\n", query))
	sb.WriteString("REQUIREMENTS:\n")
	sb.WriteString("1. Give a direct, comprehensive answer to the user query.\n")
	sb.WriteString("2. Preserve critical facts, numbers, commands, code snippets, and technical details.\n")
	sb.WriteString("3. Remove any non-essential details that are not relevant to the user query.\n")
	sb.WriteString("4. Cite sources as [Source #] where # is the source id.\n")
	sb.WriteString("5. If the sources do not answer the query, say so explicitly.\n")
	sb.WriteString("6. Use `###` headers to separate paragraphs.\n")
	sb.WriteString("</instructions>\n\n")
	for _, b := range blocks {
		sb.WriteString(b)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func dedupeURLs(urls []string) []string {
	seen := make(map[string]struct{}, len(urls))
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		u = strings.TrimRight(u, ".,;")
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
