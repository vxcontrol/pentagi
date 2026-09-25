# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Interaction Rules

1. **English** for code, comments, documentation, commit messages and identifiers.
2. **Password Complexity Requirements**: For all password-related development (registration, password reset, API token generation, etc.), enforce the same policy in **both** backend and frontend — never rely on frontend validation alone. Source of truth, keep the two in sync: `backend/pkg/password/policy.go` (the API validator and the installer both call it) and `frontend/src/features/authentication/password-change-form.tsx` (zod schema). The policy:
   - Length 8–72 characters (72 **bytes**, the most bcrypt will hash — a longer value fails inside `bcrypt.GenerateFromPassword`, after validation).
   - A password is valid if it is **either** 16+ characters (any composition), **or** 8–15 characters containing at least 1 lowercase letter, 1 uppercase letter, 1 number, and 1 special character from `!@#$&*`.
3. **Markdown is prose, not wrapped to a width.** In a `.md` file every paragraph, list item and table row is one line, and the text is split into paragraphs where the thought changes, not where a column ends — the viewer wraps it. Hard wrapping at a fixed width belongs only to comments in code.

## Code Comments

The target: comment lines at 0–2% of a file of logic and about 18% of a file of type declarations; function bodies of hundreds of lines carry none, because the code says what it does.

**A comment earns its place only by carrying what the code cannot show**: why this decision and not the obvious one, what external constraint it works around, why the code exists at all, and — where it genuinely matters — the boundary conditions. It must help maintenance and must not cost readability.

1. **Never restate the code.** If the sentence follows from the line under it, delete it. A doc comment that only repeats an identifier's name is noise; delete it rather than reword it. A comment that repeats the `fmt.Errorf` or log message on the next line is noise too.

2. **Short and dense beats long and dense.** No filler that dilutes the point.

3. **English only.** No Russian, not even a quoted requirement.

4. **No `§`, no requirement ids like `(R37)`, no reference to a document that is not in this repository.** A planning document that is not tracked leaves every pointer into it dead. Do not just delete the token — rewrite the sentence so it stands alone, or delete the comment. "The spec says X" is not a reason; if the reason is real, state the reason. (`§` is a legitimate *string* separator in `backend/cmd/installer/wizard/models/types.go` and a legitimate reference to the CommonMark spec in `markdown-editor-marked.ts` — those are not comments.)

5. **No archaeology.** Git holds the history. Delete "it used to be", "no longer", "before this", "now that X arrives", "was an overgeneralisation". Keep the invariant the story was there to protect. This applies to test *names* as well as test comments.

6. **No meta-commentary** about why the enum has five members rather than four, and no narrative padding that sets a scene or argues with an imagined reviewer.

7. **Rationale lives in one place** — the production code. A `_test.go` gets at most one line saying what a case pins down, and only when the test name does not already say it.

8. **When you change a comment, verify it against the code**, including that the case it describes is still reachable where it sits. A shorter comment that is wrong is worse than the long one.

Keep, and shorten only where it stays unambiguous: package doc comments; the contract of an exported identifier (units, ownership, zero/nil semantics, what the caller must hold); locking, ordering and fail-closed rationale; load-bearing formatting facts; workarounds for a library, protocol, database or browser quirk. Never touch `TODO`/`FIXME`, `//nolint`, `// eslint-disable`, `//go:generate`, `//go:embed`, build tags, or the swaggo `// @...` annotations in `pkg/server/services/` — they are machine-read, not prose.

Two kinds of comment here are machine-read even though they look like prose: the triple-quoted descriptions in `schema.graphqls`, which reach `models_gen.go`, the frontend types and every API consumer; and the comment above a `-- name:` line in `backend/sqlc/models/*.sql`, which is copied verbatim into `pkg/database/*.sql.go`. Both need regeneration after an edit.

### The GraphQL schema

`backend/pkg/graph/schema.graphqls` is a **public** artifact: introspection is on by design, so every triple-quoted description in it ships to clients and bloats the spec. A `#` comment does not ship — the parser drops it, so no client sees it and regeneration leaves every output unchanged — but it is read by everyone who edits the file. The schema carries almost no prose.

**No triple-quoted descriptions anywhere.** Not on types, not on fields, not on operations, not on enum values.

**A comment on a type, when one is genuinely needed, is a single `#` line above it, with no trailing period.** It earns its place only by saying something a client cannot get from the name — for example that the bytes are served over REST and not here, or that every mutation of the shipped template is refused.

**Fields inside types and inputs are never commented.** The name and the type are the documentation; anything more belongs in `backend/docs/` or in the Go code.

**Queries, mutations and subscriptions are never commented.** Their names say what they do; if a name does not, rename it. A single `#` line may head a *group* of operations as a navigation marker, nothing more.

**Enum values are commented only where the name does not give it**, and then about what the value leads to or how the algorithm treats it — one `#` line each.

After editing the schema, regenerate: `gqlgen`, then `swag init` — the swagger document is built from the gqlgen-generated models — and then `pnpm run graphql:generate` in `frontend/`.

**Proving a comment-only change touched no code.** For anything larger than a few files, compare the comment-free token stream of each file against its previous revision rather than reading the diff: a ~30-line Go program over `go/scanner` for `.go`, and `ts.createPrinter({removeComments:true})` over `ts.createSourceFile(...)` for `.ts`/`.tsx` (the raw `ts.createScanner` returns `Unknown` inside JSX and is not a substitute). Prove the checker itself in both directions — plant a one-token change and confirm it fails, plant a comment-only change and confirm it passes — before trusting it. A dead reference inside a string literal (a `t.Errorf` format, a `describe(...)` title) is a code token and needs its own deliberate edit.

## Project Overview

**PentAGI** is an automated security testing platform powered by AI agents. A `primary_agent` drives a flow, a `generator`/`refiner` pair plans and re-plans its subtasks, and specialist agents — `pentester`, `coder`, `searcher`, `installer`, `memorist`, `adviser`, `reflector`, `enricher`, `reporter`, `summarizer`, `assistant`, `tool_call_fixer` — execute them against LLM providers, Docker-sandboxed tools and a pgvector memory store. The authoritative list of roles is the `MsgchainType` enum in `backend/pkg/database/models.go`, mirrored as `AgentType` in the GraphQL schema.

The application is a monorepo:

- **`backend/`** — Go REST + GraphQL API server
- **`frontend/`** — React + TypeScript web UI
- **`observability/`** — optional monitoring stack configs
- **`scripts/`** — version stamping, image entrypoint, helper scripts

## Build & Development Commands

### Backend (run from `backend/`)

```bash
go mod download                              # Install dependencies
go generate ./cmd/installer/files/           # REQUIRED once per fresh clone: fs.go, fs_test.go and fs/ in
                                             #   cmd/installer/files/ are generated and gitignored; the installer
                                             #   tests fail without them
go build -trimpath -o pentagi ./cmd/pentagi  # Quick compile; the binary reports version "ce"
go test ./...                                # Default test tier (see Testing)
golangci-lint run --timeout=5m               # Lint — config backend/.golangci.yml (v2 schema); CI pins v2.12.2
```

A binary whose reported version matters needs the ldflags CI passes: `-ldflags "-X pentagi/pkg/version.PackageVer=<ver> -X pentagi/pkg/version.PackageRev=<short-sha>"`.

### Frontend (run from `frontend/`)

```bash
nvm use                   # Node from .nvmrc — corepack does not manage Node
corepack enable           # once per Node install: activates the pnpm pinned in package.json
pnpm install
pnpm run dev              # Dev server on http://localhost:8000 (VITE_PORT / VITE_HOST / VITE_USE_HTTPS override it)
pnpm run build            # tsc -b && vite build
pnpm run typescript       # Typecheck only. Use THIS: the root tsconfig is `files: []`, so
                          #   `tsc --noEmit -p tsconfig.json` passes without checking anything
pnpm run lint             # ESLint, --max-warnings 0 over src, e2e and scripts
pnpm run prettier         # Format check (prettier:fix / lint:fix to write)
pnpm run test             # Vitest (test:watch, test:coverage)
pnpm run e2e              # Playwright, mock tier (see Testing)
pnpm run graphql:generate # Regenerate src/graphql/types.ts
```

### Docker (run from repo root)

```bash
docker compose up -d                                                            # Core services
docker compose -f docker-compose.yml -f docker-compose-observability.yml up -d  # + monitoring
docker compose -f docker-compose.yml -f docker-compose-langfuse.yml up -d       # + LLM analytics
docker compose -f docker-compose.yml -f docker-compose-graphiti.yml up -d       # + knowledge graph
docker build -t local/pentagi:latest .                                          # Image, version "ce"
./scripts/build-image.sh                                                        # Image stamped the way
                                                                                #   the pipeline stamps it
```

`docker-compose.e2e.yml` is not part of that stack: it is the isolated Tier-2 e2e backend, started under its own compose project (see the header of the file).

The full stack runs at `https://localhost:8443`. Copy `.env.example` to `.env` and fill in at minimum the database and one LLM provider key.

## Testing

**Backend.** `go test ./...` is not the whole suite — one tier sits behind a build tag and is invisible without it:

```bash
PENTAGI_TEST_DSN=postgres://… go test -p 1 -tags postgres ./pkg/database/... ./pkg/timezone/... ./migrations/
```

The tier fails when its variable is unset. `-p 1` because each package migrates into its own schema of one database, while `pg_trgm` is installed once per database: a second package running beside the first fails with `operator class gin_trgm_ops does not exist`. Under a plain `go test ./...` the suites that need a live dependency skip themselves instead — `TEST_DATABASE_URL` for the statement-timeout suite, a reachable Docker daemon for `pkg/docker` — and `go test` without `-v` does not print skips, so a green run can mean they never ran.

**Frontend.** Vitest for units; Playwright for the rest, in three tiers chosen by `E2E_TIER`:

- `mock` (default) — builds the production bundle, serves it, answers every call from a cassette;
- `local` — real backend from `docker-compose.e2e.yml` with a mock LLM (`e2e/tools/run-local-tier.sh`);
- `stand` — a deployed stand, needs `E2E_BASE_URL` and credentials.

Visual snapshots run only inside the pinned Playwright container (`pnpm run e2e:visual`); never on the host. The authoritative verdict of a run is `e2e/test-results/.last-run.json`, not the exit code of a piped command.

**The written playbooks live in `frontend/docs/`** and are worth reading before a manual pass: `e2e.md` (harness, cassettes, tiers, traps), `live-testing.md` (driving the real UI) and `list-detail-pages.md`.

## Writing Tests

Where a test lives and what it is called decide whether anyone finds it again. Each rule below is here because breaking it happened: `pkg/providers` grew 43 test files for 17 source files, one per fix, and seven of them each loaded the model catalogues to check a single property.

**Layout — one test file per source file.**

1. Tests of `x.go` live in `x_test.go`. A test for a bug, a feature or a regression goes into the test file of the source it pins — never into a new file named after the fix, the feature or the ticket. What decides the file is the code the test exists to pin, not the code it happens to execute most, and not the package it was written in: a test of `pconfig.GetOptionsForType` belongs in `pkg/providers/pconfig`.
2. The only other names: `fixtures_test.go` — test doubles shared by more than one test file of the package, and a package's catalogue of realistic sample inputs with their builders (the message chains of `pkg/cast`); `x_ext_test.go` — tests that must be `package <name>_test` because of an import cycle, with one line saying which; `x_live_test.go` — the `//go:build postgres` tier; `x_linux_test.go` and other GOOS/GOARCH suffixes — filename build constraints, never renamed away; `main_test.go` for `TestMain` and the tests of `main.go`; `example_test.go` for `Example…` functions only.
3. A helper defined in one test file and used by another is a hidden dependency: deleting the first breaks the second. Shared doubles live in `fixtures_test.go`, one per role — read it before writing another fake embedder, log recorder or Docker client.

**Names.**

4. `Test<Component>_<Unit>_<Behaviour>`: the source file in CamelCase, the one function or method under test, and what must hold — `TestTerminal_ExecCommand_KeepsPartialOutputOnTimeout`. Then `go test -run 'TestTerminal_'` selects a component and `-run 'TestTerminal_ExecCommand'` a unit. The name promises nothing the assertions do not check. Subtests are lower-case sentences.

**Before adding a test**, find the existing test of the same unit. The same setup with another input is a new row in its table, not a new function; the same input class with the same asserted outcome is a duplicate.

**Every assertion must be able to fail.** Each of these shipped here and stayed green with the behaviour it named removed:

5. A double standing in for a database, log or network write returns `ctx.Err()` once its context is done. One that ignores the context cannot tell a write on a cancelled context from a live one.
6. No expected value computed by the code under test or read from a production constant — a literal.
7. Test through the public entry — `Handle`, `Execute`, the Gin handler via `httptest`, `resolver.Mutation().X` — not the helper behind it, which stays green when the entry stops calling it.
8. Timing: lower bounds, or generous upper bounds as literals; a starved CI runner must not fail the test. Every blocking wait sits in a `select` with a timeout that fails with a reason. Parallel subtests do not share a double that mutates state.
9. `err != nil` does not say which error: use `errors.Is` or the distinguishing text.

**Proof and gates.**

10. A new or rewritten test is shown failing with its behaviour broken: mutate the production line its path takes, see `--- FAIL`, restore. Mutate in a throwaway copy of the tree — an interrupted script once left production code mutated in the working copy.
11. Run the package with `-race` — CI runs `go test ./... -race` — and without `-trimpath`, which breaks tests that open source files (`pkg/providers` parses `pkg/database/models.go`).
12. Renaming or moving a test file updates every document that names it, this one included, in the same change.
13. A test passes on any developer's machine, so it never depends on that machine's timezone, locale or installation. Build times in an explicit zone — `time.UTC`, or `Asia/Riyadh` (+03:00, no DST) where a non-zero offset matters — import `_ "time/tzdata"` wherever a zone is loaded by name, and never assert on what a local server happens to carry (its zone catalogue, its version). Prove a test that touches dates under `TZ=America/Los_Angeles`, `TZ=Asia/Shanghai` and `TZ=Pacific/Kiritimati`, and on the frontend also under `LC_ALL=zh_CN.UTF-8`; the Go test cache does not key on `TZ`, so run with `-count=1` or a zone-bound test comes back `(cached) ok`.
14. Coverage is not test strength: statement coverage of `pkg/csum` rose while 38% of its mutants still went unnoticed. A package that implements an algorithm over structured input — `pkg/cast` parsing, `pkg/csum` summarisation — is tested end to end on realistic inputs, step by step, with an exact assertion after every step, and its strength is measured by a mutation sweep, before and after any test is merged or deleted. Run that sweep one mutant at a time and cap each test binary's memory and time in its own process group: a mutated loop counter can allocate gigabytes in seconds, and killing `go test` leaves its test binary running.

## Architecture

### Backend Package Structure

| Package | Role |
|---|---|
| `cmd/pentagi/` | Entry point: config, tenant schema, DB pools, goose migrations, telemetry, Docker, providers, flow restore, HTTP server |
| `pkg/config/` | One env-tagged `Config`; validates `TENANT_ID` and derives the schema name. The widest dependency in the backend |
| `pkg/server/` | Gin engine and middleware. The work is in its subpackages: `services/` REST handlers, `models/` DTOs and their `Valid()` gates, `auth/` sessions + token JWTs + privilege checks, `oauth/`, `rdb/` filtering and paging, `response/`, `update/` |
| `pkg/controller/` | Flow/task/subtask agent runtime: spawns and stops workers, publishes agent, message, tool-call, terminal and screenshot logs to the subscription hub |
| `pkg/graph/` | gqlgen schema (`schema.graphqls`), resolvers, `model/` (generated) and `subscriptions/` (the pub/sub hub) |
| `pkg/database/` | sqlc-generated queries and models, tenant helpers, statement timeouts, and the legacy GORM handle. GORM *models* live in `pkg/server/models/` |
| `pkg/providers/` | LLM adapters. `provider/` holds the interface, the `ProviderType` constants and `AllProviderTypes`; `pconfig/` the per-agent model config; `openaicompat/` the shared OpenAI-shaped door; `embeddings/` a separate axis |
| `pkg/tools/` | The agent tool-call layer: schemas (`registry.go`), dispatcher (`executor.go`), and the tools — terminal, code, browser, memory, knowledge, `web_search` over `searchers/` |
| `pkg/docker/` | Docker SDK wrapper for the per-flow sandbox containers |
| `pkg/csum/` | Chain summarization for LLM context management, over the `chainAST` in `pkg/cast/` |
| `pkg/templates/` | Embedded agent prompts (`prompts/*.tmpl`) and the validator that checks a user-edited prompt still renders |
| `pkg/graphiti/` | HTTP client for the Graphiti knowledge-graph service. It does not speak Neo4j |
| `pkg/observability/` | OpenTelemetry traces/metrics/logs, `langfuse/`, `profiling/` |
| `pkg/flowfiles/`, `pkg/resources/` | Per-flow file tree and sandbox transfer; the user-level resource store |
| `pkg/schema/`, `pkg/password/`, `pkg/system/`, `pkg/timezone/`, `pkg/version/`, `pkg/terminal/` | Validation plumbing; the password policy; shared HTTP client; IANA name resolution; link-time version; console output for the `cmd/*tester` CLIs |

Migrations are `backend/migrations/sql/*.sql`, **embedded into the binary** and run by goose at startup — a new migration needs a rebuild, not just a restart. A failure aborts the boot.

### Frontend Structure

```
frontend/src/
├── main.tsx / app.tsx   # Entry point; Apollo and the context providers wrap the router tree
├── pages/               # Route components, one directory per route group (flows, settings, …)
├── features/<domain>/   # Domain code: authentication, flows, knowledges, resources, templates
├── providers/           # Context providers — where the queries and subscriptions live
├── components/ui/       # Design-system primitives (shadcn registry; about half wrap Radix)
├── components/shared/   # Composites used by more than one area or by the app shell
├── components/layouts/  # App shell layouts
├── graphql/types.ts     # Generated. Never edit by hand — it carries no banner saying so
├── hooks/  lib/  styles/  models/  types/  test-utils/
└── (outside src) e2e/ docs/ scripts/ eslint-rules/   # Playwright suite, testing playbooks,
                                                    #   build helpers, and one local ESLint rule
```

Placement rule, as the code applies it: `components/ui` and `components/shared` never import `features/*` (ESLint rejects it); `components/ui` reaches the app only through the generic hooks in `hooks/`; `components/shared` holds composites that more than one area renders and takes its domain data through props (the version panel is the one that reads a domain provider); `features/<domain>` is where data coupling belongs — providers, generated types and REST calls — and a feature is normally consumed by its own domain's pages.

Server state comes from **Apollo Client 4** (note the v4 import paths, `@apollo/client/react`). `lib/apollo.ts` splits subscriptions onto a `graphql-ws` WebSocket link. Components seldom call Apollo directly: the queries and subscriptions live in `src/providers/*`, which `app.tsx` mounts around the router — add a new read there first. Not everything is GraphQL: account, resource and flow-file writes go over REST through `@/lib/axios`.

### Data Flow

1. A user creates a "flow" — the UI calls the GraphQL `createFlow` mutation; REST is the programmatic equivalent.
2. `CreateFlow` (`pkg/controller/flows.go`) reserves the row and builds the worker asynchronously; there is no queue, and a build failure marks the flow `failed` and says so on its message log.
3. The `primary_agent` decomposes the flow into tasks; the `generator` produces each task's subtasks and the `refiner` re-plans after every completed one. Subtasks run one at a time and delegate to the specialist agents.
4. Those agents call tools — terminal, file, browser, search — inside per-flow Docker containers.
5. Results, tool output and LLM reasoning land in PostgreSQL (pgvector for semantic memory).
6. Progress reaches the UI through GraphQL subscriptions fed by `pkg/graph/subscriptions`.
7. Beside flows there is an **assistant**: an interactive chat attached to a flow, sharing its containers and provider settings but with its own log tables.

### Authentication

- **Session cookies** for browser login — cookie store keyed by `COOKIE_SIGNING_SALT` mixed with `TENANT_ID`; `HttpOnly` always, `Secure` only over HTTPS. Changing the salt invalidates every session.
- **OAuth2** via Google and GitHub.
- **Bearer tokens** (API tokens table) for programmatic access.

All three pass through `pkg/server/auth/auth_middleware.go`, which offers `AuthUserRequired`, `AuthTokenRequired` and `TryAuth`. Authorization is separate: `auth.PrivilegesRequired("<area>.<action>")` wraps individual routes in `pkg/server/router.go`. Repeated failed logins are throttled.

### Key Integrations

- **LLM Providers**: OpenAI, Anthropic, Gemini, AWS Bedrock, Ollama, DeepSeek, GLM, Kimi, Qwen, MiniMax, Mistral, xAI, and custom OpenAI-compatible endpoints. The authoritative list is `providerRegistry` in `backend/pkg/providers/registry.go` — check it before enumerating providers anywhere else.
- **Search**: DuckDuckGo, Google, Tavily, Firecrawl, Traversaal, Perplexity, Searxng, Sploitus and an internal browser-backed engine, all behind the single `web_search` tool. Authoritative list: `buildSearchEngines` in `backend/pkg/tools/web_search.go`.
- **Databases**: PostgreSQL + pgvector (required); Neo4j (optional) behind the Graphiti service, reached over HTTP.
- **Observability**: OpenTelemetry → VictoriaMetrics + Loki + Jaeger → Grafana; Langfuse for LLM analytics.

## Code Generation

The output of the generators below is committed — regenerate and commit in the same change. When a codegen input changes, `.github/workflows/ci.yml` regenerates `frontend/src/graphql/types.ts` and fails if the committed file differs; it does not check the gqlgen, sqlc or swag output, so a stale one compiles and merges unnoticed.

| Change | Command (from) | Writes |
|---|---|---|
| `backend/pkg/graph/schema.graphqls` | `go run github.com/99designs/gqlgen --config ./gqlgen/gqlgen.yml` (`backend/`) | `pkg/graph/generated.go`, `pkg/graph/model/`, resolver stubs in `pkg/graph/schema.resolvers.go` |
| `backend/sqlc/models/*.sql`, a migration | `sqlc generate` (`backend/sqlc/`) — needs a **live** PostgreSQL in `DATABASE_URL` | `backend/pkg/database/` |
| REST handler annotations | `swag init -g ../../pkg/server/router.go -o pkg/server/docs/ --parseDependency --parseInternal --parseDepth 2 -d cmd/pentagi` (`backend/`) | `pkg/server/docs/` |
| `frontend/graphql-schema.graphql` (the operations file) or the backend schema | `pnpm run graphql:generate` (`frontend/`) | `frontend/src/graphql/types.ts` |

A backend schema change therefore needs **both** gqlgen and the frontend codegen.

## Adding a New LLM Provider

Order matters — types before the door, because the door's constructor names them. Copy the last provider added (`mistral`, `xai`) rather than an old one.

1. **Types.** In `pkg/providers/provider/provider.go`: the `Provider<Name>` and `DefaultProviderName<Name>` constants, the entry in `AllProviderTypes`, and the classification in `ReasoningProvider()`.
2. **Config.** `<NAME>_API_KEY` / `<NAME>_SERVER_URL` / `<NAME>_PROVIDER` in `pkg/config/config.go` (the third carries the gateway's model prefix), mirrored in `.env.example`, `docker-compose.yml` and the `clearConfigEnv` list in `config_test.go`. Add every credential field to `GetSecretPatterns()` — values listed there are masked out of logs and agent transcripts.
3. **Migration.** A goose migration that rebuilds the `PROVIDER_TYPE` enum with the new value and re-points `providers.type`, `flows.model_provider_type` and `assistants.model_provider_type`; Postgres will not extend that enum in place. Then regenerate `pkg/database/` with sqlc.
4. **Catalogue and door.** `pkg/providers/<name>/` with `config.yml`, `models.yml` and a `<name>.go` exporting `DefaultProviderConfig`, `BuildProviderConfig`, `DefaultModels` and `New`. An OpenAI-shaped vendor does not implement the interface by hand — `New` returns `openaicompat.New(openaicompat.Spec{…})`.
5. **Three tables, three edits**: `providerRegistry` (`registry.go`), `bundledCatalogs` (`catalog.go`), `doorEndpoints` (`routing.go`). Miss one and a guard test in `pkg/providers` fails.
6. **Installer.** The screen id, menu row, screen construction and strings under `cmd/installer/wizard/`. Nothing fails the build if you skip it; the wizard simply never offers the provider.
7. **GraphQL.** The `ProviderType` enum value plus fields in `ProvidersModelsList`, `ProvidersReadinessStatus` and `DefaultProvidersConfig`; re-run gqlgen and extend the per-type switches in `schema.resolvers.go`.
8. **Frontend.** `providerLabels` in `pages/settings/settings-providers.tsx` (a `Record<ProviderType, …>`, so it will not compile without the new key), the icon in `components/icons/` registered in `provider-icon.tsx`, the fields in `frontend/graphql-schema.graphql`, then `pnpm run graphql:generate`, and the e2e cassette in `e2e/mocks/cassettes/settings-providers.ts`.
9. **Docs.** The section every other provider has in `README.md`, `backend/docs/config.md` and `backend/docs/database.md`. `TestCatalog_ReadmeProviderSectionsMatchTheShippedConfig` in `pkg/providers/catalog_test.go` reads `README.md` — a provider or model missing from it turns that test red.

Embeddings are a separate axis (`EMBEDDING_PROVIDER`, `pkg/providers/embeddings/`): adding a chat door does not add an embedder.

## Adding a New Search Engine

Engines are primitives under `backend/pkg/tools/searchers/`, orchestrated by the single `web_search` tool. Agents never call an engine directly — they call `web_search` with an intent `mode`.

1. `searchers/<name>.go` implementing `searchers.Searcher`: constructor, `IsAvailable()`, `Engine()`, and a `Handle(ctx, Request)` returning **typed** errors (`searchers.Retryable` / `searchers.Fatal` / `searchers.ErrNotConfigured`, or `searchers.ClassifyHTTPStatus`). Never swallow an error into a result string. `searchers` must not import `pkg/tools`.
2. Config fields in `pkg/config/config.go`, plus `.env.example`, `docker-compose.yml`, the `envDefault` assertions in `config_test.go`, and `GetSecretPatterns()` for every credential.
3. Construct it in `buildSearchEngines` and place its id in the `fallbackStrategy` chains in `web_search.go` — the only place per-mode priority lives. An id in a chain with no entry in the engines map is skipped silently at runtime, not at compile time.
4. Attribution, if the engine needs a **new** `SearchengineType`: a goose migration rebuilding the enum, the constant in `pkg/database/models.go`, **and** the value in both the constant block and the `Valid()` switch in `pkg/server/models/searchlogs.go`. Reusing an existing value needs none of it.
5. `<name>_test.go` in `searchers/` (the shared MITM proxy harness is in `fixtures_test.go`), and orchestrator coverage in `web_search_test.go` if behaviour changes.
6. The installer wizard (`cmd/installer/wizard/`: the config field, the `GetVar`/`SetVar` pair, the descriptions and masked/critical lists, the labels, the form field) and `cmd/ftester` if you want to drive the engine in isolation.
7. The prose that enumerates engines: `pkg/templates/prompts/searcher.tmpl` (what the agent is told exists), `backend/docs/config.md`, and `README.md`.

No frontend change is needed: it treats `SearchLog.engine` as an opaque string.

## Utility Binaries

- `cmd/ctester/` — provider conformance tester: runs the case suite against one provider (`-type`), filtered by agent role and group, and writes a report. `-check-routing` checks catalogue routing only.
- `cmd/ftester/` — exercises LLM function/tool calling, including one search engine at a time.
- `cmd/etester/` — embedding tester and pgvector store maintenance. `flush` and `reindex` act on the real database in `DATABASE_URL` and are destructive.
- `cmd/installer/` — interactive TUI wizard for guided deployment (writes `.env`, compose files, DB and search settings). Its embedded files are generated: see `go generate` above.
