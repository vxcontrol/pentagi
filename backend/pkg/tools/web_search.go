package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/tools/searchers"

	"github.com/sirupsen/logrus"
)

// ============================================================================
// web_search engine selection & fallback strategy
//
// This block is the SINGLE SOURCE OF TRUTH for how web_search chooses an engine.
// It is meant to be read and edited by hand: to re-prioritize engines, reorder a
// slice below; to add a newly-implemented engine, construct it in buildSearchEngines
// and drop its id into the relevant chains here. Nothing else in this file needs to
// change to change the behavior.
// ============================================================================

// Engine identifiers. These are aliases for the database.SearchengineType values so
// the strategy table reads cleanly AND the winning engine can be logged for
// attribution with no extra mapping. EngineInternal reuses the (otherwise unused)
// "browser" attribution value, so the internal analytics engine needs no new enum
// value and no DB migration.
const (
	EngineGoogle     = database.SearchengineTypeGoogle
	EngineDuckDuckGo = database.SearchengineTypeDuckduckgo
	EngineTavily     = database.SearchengineTypeTavily
	EngineFirecrawl  = database.SearchengineTypeFirecrawl
	EngineTraversaal = database.SearchengineTypeTraversaal
	EnginePerplexity = database.SearchengineTypePerplexity
	EngineSearxng    = database.SearchengineTypeSearxng
	EngineSploitus   = database.SearchengineTypeSploitus
	EngineInternal   = database.SearchengineTypeBrowser
)

// SearchMode is the INTENT the agent expresses — the shape of answer it needs. The
// agent chooses the mode; web_search chooses the concrete engine.
type SearchMode string

const (
	// ModeLinks — a ranked list of source links with titles/snippets ("what pages
	// exist"). Cheapest intent; pure index engines lead.
	ModeLinks SearchMode = "links"

	// ModeAnswer — a synthesized answer over live sources ("answer my question").
	// Engines that browse and summarize themselves lead; link engines are the floor.
	ModeAnswer SearchMode = "answer"

	// ModeResearch — deep, multi-source synthesis with reasoning ("research this").
	// Most expensive intent; the strongest analytic engine leads.
	ModeResearch SearchMode = "research"

	// ModeExploit — exploit / PoC / offensive-tool discovery. The exploit index
	// leads; universal analytic engines and then classic engines back it up.
	ModeExploit SearchMode = "exploit"
)

// defaultMode is used when the agent omits `mode` (or sends an unknown one). "answer"
// is the most generally useful intent and degrades to links on its own chain's tail.
const defaultMode = ModeAnswer

// fallbackStrategy maps each mode to its ordered engine chain. For a given mode the
// orchestrator walks the slice in order, SKIPS any engine whose IsAvailable() is
// false (e.g. no API key configured), and returns the first successful result. The
// chain is only "exhausted" — and an error surfaced to the LLM — after every
// AVAILABLE engine in the slice has failed.
//
// EngineInternal appears in the analytic chains but is disabled by default
// (IsAvailable() == cfg.WebSearchInternalEnabled), so it is skipped unless the
// operator opts in via WEB_SEARCH_INTERNAL_ENABLED=true.
var fallbackStrategy = map[SearchMode][]database.SearchengineType{
	// 1. Link discovery — cheap index engines first, ordered by breadth/consistency;
	//    analytic engines are a deep last resort so links never dead-ends.
	ModeLinks: {
		EngineGoogle, EngineDuckDuckGo, EngineSearxng, EngineFirecrawl,
		EngineTavily, EnginePerplexity, EngineTraversaal,
	},

	// 2. Answer with analysis — engines that navigate + summarize lead; the internal
	//    browser-analytics engine is a mid/late fallback; link engines are the floor.
	ModeAnswer: {
		EngineTavily, EngineFirecrawl, EnginePerplexity, EngineInternal, EngineTraversaal,
		EngineGoogle, EngineDuckDuckGo, EngineSearxng,
	},

	// 3. Deep research — strongest reasoning engine first, then the rest.
	ModeResearch: {
		EnginePerplexity, EngineTavily, EngineFirecrawl, EngineInternal, EngineTraversaal,
		EngineGoogle, EngineDuckDuckGo, EngineSearxng,
	},

	// 4. Exploit search — dedicated exploit index first, universal analytic engines
	//    next, classic engines last.
	ModeExploit: {
		EngineSploitus, EngineTavily, EngineFirecrawl, EnginePerplexity, EngineInternal,
		EngineTraversaal, EngineGoogle, EngineDuckDuckGo, EngineSearxng,
	},
}

// linkEngineOrder is the priority order of link-discovery engines. It is used to feed
// the internal analytics engine the URLs to read.
var linkEngineOrder = []database.SearchengineType{
	EngineGoogle, EngineDuckDuckGo, EngineSearxng, EngineFirecrawl,
}

const (
	// maxEngineAttempts is the per-engine try budget: 1 initial call plus retries for
	// the retryable error classes (429 / 5xx / transport). Fatal errors never retry.
	maxEngineAttempts = 2
	// retryBackoffBase is multiplied by the attempt number for a small linear backoff.
	retryBackoffBase = 300 * time.Millisecond

	// webSearchDefaultResults is used when the agent omits max_results.
	webSearchDefaultResults = 5
	// webSearchMaxResults is the upper clamp (sploitus supports up to 25).
	webSearchMaxResults = 25
)

// ============================================================================
// Orchestrator
// ============================================================================

type webSearch struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider

	// engines is the availability-agnostic registry keyed by engine id, built once at
	// construction. Availability is (re)checked per call so config edits are honored.
	engines map[database.SearchengineType]searchers.Searcher
}

// NewWebSearchTool builds the web_search orchestrator. summarizer and fetcher are the
// flow summarizer and a browser-backed page fetcher (may be nil, which only disables
// the internal analytics engine).
func NewWebSearchTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
	summarizer SummarizeHandler,
	fetcher searchers.PageFetcher,
) Tool {
	return &webSearch{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
		engines:   buildSearchEngines(cfg, summarizer, fetcher),
	}
}

// buildSearchEngines constructs every searcher from cfg. To add an engine: construct
// it here and reference it in fallbackStrategy above.
func buildSearchEngines(
	cfg *config.Config,
	summarizer SummarizeHandler,
	fetcher searchers.PageFetcher,
) map[database.SearchengineType]searchers.Searcher {
	// tools.SummarizeHandler and searchers.SummarizeHandler share an underlying type;
	// the explicit conversion bridges the two packages without an import cycle.
	sum := searchers.SummarizeHandler(summarizer)

	engines := map[database.SearchengineType]searchers.Searcher{
		EngineGoogle:     searchers.NewGoogle(cfg),
		EngineDuckDuckGo: searchers.NewDuckDuckGo(cfg),
		EngineTavily:     searchers.NewTavily(cfg, sum),
		EngineFirecrawl:  searchers.NewFirecrawl(cfg, sum),
		EngineTraversaal: searchers.NewTraversaal(cfg),
		EnginePerplexity: searchers.NewPerplexity(cfg, sum),
		EngineSearxng:    searchers.NewSearxng(cfg, sum),
		EngineSploitus:   searchers.NewSploitus(cfg),
	}

	// The internal analytics engine discovers URLs with the link engines (in priority
	// order), reads them via the browser fetcher, and summarizes the result.
	engines[EngineInternal] = NewInternalEngine(cfg, summarizer, fetcher)

	return engines
}

// NewInternalEngine constructs the internal browser-analytics engine with its link
// searchers wired in the production priority order (linkEngineOrder). fetcher must be a
// browser-backed page fetcher (see NewBrowserPageFetcher); a nil fetcher (or a disabled
// WEB_SEARCH_INTERNAL_ENABLED) yields an engine that reports IsAvailable()==false. This
// is the single place the internal engine is assembled, reused by the orchestrator and
// exposed for the function tester.
func NewInternalEngine(cfg *config.Config, summarizer SummarizeHandler, fetcher searchers.PageFetcher) searchers.Searcher {
	sum := searchers.SummarizeHandler(summarizer)
	byID := map[database.SearchengineType]searchers.Searcher{
		EngineGoogle:     searchers.NewGoogle(cfg),
		EngineDuckDuckGo: searchers.NewDuckDuckGo(cfg),
		EngineSearxng:    searchers.NewSearxng(cfg, sum),
		EngineFirecrawl:  searchers.NewFirecrawl(cfg, sum),
	}
	links := make([]searchers.Searcher, 0, len(linkEngineOrder))
	for _, id := range linkEngineOrder {
		if e, ok := byID[id]; ok {
			links = append(links, e)
		}
	}
	return searchers.NewInternal(cfg, fetcher, links, sum)
}

// IsAvailable reports whether at least one underlying engine is configured. If none
// are, web_search is not registered (matching the old per-engine behavior).
func (w *webSearch) IsAvailable() bool {
	for _, e := range w.engines {
		if e != nil && e.IsAvailable() {
			return true
		}
	}
	return false
}

func (w *webSearch) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action WebSearchAction
	if err := json.Unmarshal(args, &action); err != nil {
		// Malformed arguments ARE fixable by the tool-call arg-fixer, so this is the
		// one case web_search hard-fails (non-nil error) — matching every other tool.
		return "", fmt.Errorf("failed to unmarshal %s arguments: %w", name, err)
	}

	query := strings.TrimSpace(action.Query)
	if query == "" {
		return "", fmt.Errorf("%s: 'query' is required and cannot be empty", name)
	}

	mode := normalizeMode(action.Mode.String())
	req := searchers.Request{
		Query:       query,
		MaxResults:  clampResults(action.MaxResults.Int()),
		ExploitType: strings.TrimSpace(action.ExploitType.String()),
		Sort:        strings.TrimSpace(action.Sort.String()),
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(w.flowID, w.taskID, w.subtaskID, logrus.Fields{
		"tool":  name,
		"mode":  string(mode),
		"query": query[:min(len(query), 1000)],
	}))

	chain := fallbackStrategy[mode]
	attempted := 0
	var lastErr error
	var lastEngine database.SearchengineType

	for _, engineID := range chain {
		engine, ok := w.engines[engineID]
		if !ok || engine == nil || !engine.IsAvailable() {
			continue
		}

		attempted++
		result, err := w.runWithRetry(ctx, engine, req)
		if err == nil {
			w.logSearch(ctx, engineID, query, result)
			observation.Event(
				langfuse.WithEventName("web_search resolved"),
				langfuse.WithEventInput(query),
				langfuse.WithEventLevel(langfuse.ObservationLevelDefault),
				langfuse.WithEventMetadata(langfuse.Metadata{
					"tool_name": WebSearchToolName,
					"mode":      string(mode),
					"engine":    string(engineID),
					"attempts":  attempted,
				}),
			)
			logger.WithFields(logrus.Fields{"engine": engineID, "attempts": attempted}).
				Info("web_search resolved")
			return result, nil
		}

		lastErr, lastEngine = err, engineID
		logger.WithError(err).WithField("engine", engineID).
			Warn("web_search engine failed, falling back")
	}

	// The chain is exhausted. This is NOT a malformed-arguments problem, so degrade
	// gracefully (message + nil error) rather than burning agent retries — matching
	// the terminal/browser/graphiti convention.
	if attempted == 0 {
		msg := fmt.Sprintf(
			"web_search: no search engine is configured for mode '%s'. "+
				"Ask the operator to configure at least one provider "+
				"(e.g. TAVILY_API_KEY, GOOGLE_API_KEY + GOOGLE_CX_KEY, or DUCKDUCKGO_ENABLED).",
			mode,
		)
		observation.Event(
			langfuse.WithEventName("web_search unavailable"),
			langfuse.WithEventInput(query),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": WebSearchToolName,
				"mode":      string(mode),
				"reason":    "no engine configured for mode",
			}),
		)
		logger.Warn("web_search has no configured engine for this mode")
		return msg, nil
	}

	w.logSearch(ctx, lastEngine, query, "web_search failed: "+errString(lastErr))
	observation.Event(
		langfuse.WithEventName("web_search exhausted"),
		langfuse.WithEventInput(query),
		langfuse.WithEventStatus(errString(lastErr)),
		langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
		langfuse.WithEventMetadata(langfuse.Metadata{
			"tool_name":   WebSearchToolName,
			"mode":        string(mode),
			"attempts":    attempted,
			"last_engine": string(lastEngine),
			"error":       errString(lastErr),
		}),
	)
	logger.WithError(lastErr).Warn("web_search exhausted all configured engines")

	return fmt.Sprintf(
		"web_search: all %d configured engine(s) failed for mode '%s'. Last error (%s): %s. "+
			"Rephrase the query or try a different mode.",
		attempted, mode, lastEngine, errString(lastErr),
	), nil
}

// runWithRetry runs one engine, retrying only retryable error classes on the SAME
// engine. Fatal errors (and ErrNotConfigured) return immediately so the caller moves
// to the next engine without wasting a round-trip.
func (w *webSearch) runWithRetry(ctx context.Context, engine searchers.Searcher, req searchers.Request) (string, error) {
	var err error
	for attempt := 1; attempt <= maxEngineAttempts; attempt++ {
		var result string
		result, err = engine.Handle(ctx, req)
		if err == nil {
			return result, nil
		}
		if !searchers.IsRetryable(err) || attempt == maxEngineAttempts {
			return "", err
		}

		delay := retryBackoffBase * time.Duration(attempt)
		if ra := retryAfter(err); ra > 0 {
			delay = ra
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}
	}
	return "", err
}

func (w *webSearch) logSearch(ctx context.Context, engine database.SearchengineType, query, result string) {
	agentCtx, ok := GetAgentContext(ctx)
	if !ok {
		return
	}
	_, _ = w.slp.PutLog(
		ctx,
		agentCtx.ParentAgentType,
		agentCtx.CurrentAgentType,
		engine,
		query,
		result,
		w.taskID,
		w.subtaskID,
	)
}

func normalizeMode(mode string) SearchMode {
	switch SearchMode(strings.ToLower(strings.TrimSpace(mode))) {
	case ModeLinks:
		return ModeLinks
	case ModeAnswer:
		return ModeAnswer
	case ModeResearch:
		return ModeResearch
	case ModeExploit:
		return ModeExploit
	default:
		return defaultMode
	}
}

func clampResults(n int) int {
	if n < 1 {
		return webSearchDefaultResults
	}
	if n > webSearchMaxResults {
		return webSearchMaxResults
	}
	return n
}

func retryAfter(err error) time.Duration {
	var re *searchers.RetryableError
	if errors.As(err, &re) {
		return re.RetryAfter
	}
	return 0
}

func errString(err error) string {
	if err == nil {
		return "unknown error"
	}
	return err.Error()
}

// ============================================================================
// Browser-backed page fetcher for the internal analytics engine
// ============================================================================

// browserMDReader is the subset of the browser tool the internal analytics engine
// needs. The browser tool's concrete type satisfies it.
type browserMDReader interface {
	ContentMD(ctx context.Context, url string) (string, string, error)
	IsAvailable() bool
}

// browserPageFetcher adapts the browser tool to searchers.PageFetcher.
type browserPageFetcher struct {
	browser browserMDReader
}

func (f *browserPageFetcher) FetchMarkdown(ctx context.Context, url string) (string, error) {
	md, _, err := f.browser.ContentMD(ctx, url)
	return md, err
}

// newBrowserPageFetcher builds a page fetcher backed by the browser tool/scraper.
// Returns nil when the scraper is not configured, which disables the internal
// analytics engine (its IsAvailable() then reports false).
func newBrowserPageFetcher(fte *flowToolsExecutor, taskID, subtaskID *int64) searchers.PageFetcher {
	return NewBrowserPageFetcher(fte.flowID, taskID, subtaskID, fte.cfg, fte.scp)
}

// NewBrowserPageFetcher builds a browser/scraper-backed page fetcher for the internal
// analytics engine. It returns nil when the scraper is not configured (the internal
// engine then reports IsAvailable()==false). Exposed so out-of-flow callers such as the
// function tester can construct the internal engine.
func NewBrowserPageFetcher(
	flowID int64,
	taskID, subtaskID *int64,
	cfg *config.Config,
	scp ScreenshotProvider,
) searchers.PageFetcher {
	b := NewBrowserTool(
		flowID, taskID, subtaskID,
		cfg.DataDir, cfg.ScraperPrivateURL, cfg.ScraperPublicURL,
		scp,
	)
	reader, ok := b.(browserMDReader)
	if !ok || !reader.IsAvailable() {
		return nil
	}
	return &browserPageFetcher{browser: reader}
}

// buildWebSearch constructs the web_search orchestrator for an executor context,
// wiring the browser-backed page fetcher. Shared by GetAssistantExecutor and
// GetSearcherExecutor so the two call sites never drift.
func buildWebSearch(fte *flowToolsExecutor, taskID, subtaskID *int64, summarizer SummarizeHandler) Tool {
	fetcher := newBrowserPageFetcher(fte, taskID, subtaskID)
	return NewWebSearchTool(fte.cfg, fte.flowID, taskID, subtaskID, fte.slp, summarizer, fetcher)
}
