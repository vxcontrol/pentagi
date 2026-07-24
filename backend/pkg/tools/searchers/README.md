# `searchers` — web search primitives

This package holds the individual web-search engines that the `web_search` orchestrator (`pkg/tools/web_search.go`) drives. Each engine wraps a single provider and nothing more: it does one HTTP call (or, for the internal engine, a browser + summarize pass), formats the result as markdown, and returns.

## Contract

Every engine implements `Searcher`:

```go
type Searcher interface {
	Engine() database.SearchengineType          // stable id, also used for search-log attribution
	IsAvailable() bool                          // configured & usable?
	Handle(ctx context.Context, req Request) (string, error)
}
```

Two rules make the orchestrator possible:

1. **Original, unmasked errors.** A searcher NEVER swallows a failure into a `"failed to search…"` string. It returns the real error, classified as one of:
   - `*RetryableError` — HTTP 429, HTTP 5xx, or transport/network failures. The orchestrator may retry the **same** engine.
   - `*FatalError` — HTTP 4xx (auth/bad-request/etc.), response-decode errors, unknown status. The orchestrator moves to the **next** engine immediately.
   - `ErrNotConfigured` — `Handle` was called while `IsAvailable()` is false.

   Use `ClassifyHTTPStatus(status, msg)` for status-based classification, or `Retryable(err, retryAfter)` / `Fatal(err)` directly. `IsRetryable(err)` / `IsFatal(err)` inspect the class.

2. **No side effects beyond Langfuse.** Searchers keep their Langfuse observation (the "wrapper") but do **not** write the DB search-log — the orchestrator writes exactly one search-log row per `web_search` call, attributing it to the engine that won.

`Request` is the already-parsed, engine-agnostic input (`Query`, `MaxResults`, and the exploit-only `ExploitType`/`Sort`). The lenient LLM JSON arg types stay in the parent `tools` package; searchers see plain Go types.

## Import boundary

`searchers` must **not** import `pentagi/pkg/tools` (the orchestrator imports this package). It depends only on neutral packages: `config`, `database` (for the `SearchengineType` attribution enum), `system` (HTTP client), and `observability`.

## Adding a new engine

1. Create `<name>.go` implementing `Searcher` with a `New<Name>(cfg, …)` constructor, `IsAvailable()`, `Engine()`, and a `Handle` that returns typed errors.
2. Add the engine's config field(s) to `pkg/config/config.go` (+ `.env.example`, `docker-compose.yml`, `config_test.go`).
3. Construct it in `buildSearchEngines` (`pkg/tools/web_search.go`) and place its id in the relevant `fallbackStrategy` chains — that table is the only place engine priority lives.
4. If the engine needs a **new** attribution value (not one of the existing `SearchengineType`s), add a goose migration under `backend/migrations/sql/`, a `SearchengineType<Name>` constant in `pkg/database/models.go`, and reconcile `pkg/server/models/searchlogs.go`. Reusing an existing value (as the internal engine reuses `browser`) needs no migration.
5. Add `<name>_test.go`. The shared MITM proxy harness (`newTestProxy`) lives in `proxy_test.go`.

## Current engines

`google`, `duckduckgo`, `tavily`, `perplexity`, `traversaal`, `sploitus`, `searxng`, `firecrawl`, and `internal` (the opt-in browser-analytics engine, off by default via `WEB_SEARCH_INTERNAL_ENABLED`).
