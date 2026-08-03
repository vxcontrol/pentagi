// Package searchers holds the individual web-search primitives (google, duckduckgo,
// tavily, perplexity, traversaal, sploitus, searxng, firecrawl, and the optional
// internal browser-analytics engine) that the tools.web_search orchestrator drives.
//
// Each searcher wraps a single provider, keeps its own Langfuse observation, and
// returns the ORIGINAL error — classified as *RetryableError or *FatalError — so the
// orchestrator can decide whether to retry the same engine or fall back to the next
// one. Searchers never mask an error as a "failed to search…" string and never write
// to the database search-log; both concerns belong to the orchestrator.
//
// This package must not import pentagi/pkg/tools (the orchestrator imports this
// package). It depends only on neutral packages: config, database (for the
// SearchengineType attribution enum), system (HTTP client), and observability.
package searchers

import (
	"context"

	"pentagi/pkg/database"
)

// Request is the engine-agnostic, already-parsed input the orchestrator hands to
// every searcher. The lenient LLM JSON arg types (tools.Int64, tools.String, …) are
// unmarshaled ONCE by the orchestrator; searchers only ever see plain Go types.
type Request struct {
	// Query is the technical-channel search query, always in English. The
	// orchestrator guarantees it is non-empty.
	Query string

	// MaxResults is already clamped by the orchestrator to a sane range.
	MaxResults int

	// ExploitType and Sort are exploit-search hints. They are empty for every
	// non-exploit mode and honored only by engines that support them (sploitus).
	ExploitType string // "exploits" | "tools"
	Sort        string // "default" | "date" | "score"
}

// SummarizeHandler compresses oversized source content into a query-focused answer.
// It has the same underlying signature as tools.SummarizeHandler; the orchestrator
// passes the flow summarizer in via an explicit conversion.
type SummarizeHandler func(ctx context.Context, result string) (string, error)

// Searcher is a single search primitive. Implementations keep their own Langfuse
// observation, return the original (typed) error, and never touch the DB search-log.
type Searcher interface {
	// Engine is the stable identifier used for search-log attribution and for keying
	// the orchestrator's fallback table.
	Engine() database.SearchengineType

	// IsAvailable reports whether the engine is configured and usable. The
	// orchestrator skips unavailable engines before calling Handle.
	IsAvailable() bool

	// Handle runs one query. The caller has already parsed and normalized the
	// request, which is why this signature takes a Request rather than the raw
	// (name string, args json.RawMessage) pair the tools.Tool interface uses.
	Handle(ctx context.Context, req Request) (string, error)
}

// Link is a single discovered source. It is produced by link-oriented engines and
// consumed by the internal analytics engine when it needs URLs to fetch and
// summarize.
type Link struct {
	Title   string
	URL     string
	Snippet string
}

// PageFetcher fetches a single URL and returns its main content as markdown. It is
// implemented in the tools package on top of the existing browser tool / scraper and
// injected into the internal analytics engine, so this package need not depend on the
// tools package. Callers bound the returned content themselves.
type PageFetcher interface {
	FetchMarkdown(ctx context.Context, url string) (string, error)
}
