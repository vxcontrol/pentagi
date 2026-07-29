# Database Layer

## Overview

PentAGI stores application state in PostgreSQL and uses the `vector` extension for agent memory and knowledge search. The database layer combines:

- **sqlc v1.27.0** for the main type-safe query API (`backend/pkg/database`);
- **GORM v1** (`github.com/jinzhu/gorm`) for HTTP server models and handlers (`backend/pkg/server/models`);
- **goose v3** for embedded, ordered schema migrations (`backend/migrations`);
- **lib/pq** for the shared `database/sql` pool used by sqlc and GORM;
- **pgxpool** for every `pgvector.Store` instance.

SQL queries live in `backend/sqlc/models`. Schema history lives in `backend/migrations/sql`.

This document describes how the product uses the database. It is not a substitute for the schema itself. Sources of truth, in descending order:

1. `backend/migrations/sql/*.sql` — schema and data migrations;
2. `backend/sqlc/models/*.sql` — application queries;
3. `backend/pkg/database/*.sql.go`, `models.go`, `querier.go` — generated Go API;
4. `backend/cmd/pentagi/main.go` and `backend/pkg/database/tenant.go` — connection, migration and tenant bootstrap behavior;
5. `backend/pkg/database/{database.go,converter,knowledge}` — helpers and higher-level database services.

Environment variables, PgBouncer and Supavisor setup are documented in [config.md](config.md); this file only covers the product behavior those settings enable.

## Runtime Architecture

### Startup sequence

`backend/cmd/pentagi/main.go` initializes PostgreSQL in this order:

1. Load and validate configuration (including `TENANT_ID`).
2. If `TENANT_ID` is set, create the tenant schema, validate shared extension placement and rewrite `DATABASE_URL` with the tenant `search_path` (`database.EnsureTenantSchema`).
3. Open one `*sql.DB` through `lib/pq`.
4. Configure the shared `database/sql` connection pool (`DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS`, one-hour max lifetime).
5. Verify that `current_schema()` resolves to the configured tenant schema (`database.VerifySearchPath`).
6. Build sqlc `Queries` and GORM on the same `*sql.DB`.
7. Create one shared `pgxpool.Pool` for every pgvector store and attach it to `cfg.PgxPool`.
8. Configure goose with the schema-qualified version table and run embedded migrations under a PostgreSQL advisory lock (`database.RunMigrations`).
9. Start controllers and the API server.

The process refuses to serve traffic after a tenant schema mismatch or a migration failure. The same tenant helpers are reused by utility binaries (`ftester`, `etester`) and by the installer's password-reset path when they talk to PostgreSQL.

### Database clients and pools

PentAGI opens two independent pools to the same PostgreSQL database:

| Pool | Configuration | Default | Consumers |
|---|---|---:|---|
| `database/sql` (`lib/pq`) | `DATABASE_MAX_OPEN_CONNS` / `DATABASE_MAX_IDLE_CONNS` | `25` / `5` | sqlc and GORM |
| `pgxpool.Pool` | `DATABASE_VECTOR_MAX_CONNS` | `10` | agent memory and knowledge pgvector stores |

GORM does not open another pool:

```go
queries := database.New(db)
orm, err := database.NewGorm(db, cfg.Debug)
```

During normal startup the shared pgx pool is stored in `cfg.PgxPool` and passed to vector stores with `pgvector.WithConn(cfg.PgxPool)`. Tool executors and the knowledge API therefore reuse the same pool. Their utility/test fallback uses `pgvector.WithConnectionURL(cfg.DatabaseURL)` only when `PgxPool` is nil.

### Connection budget

With the defaults, one PentAGI process can use up to 35 PostgreSQL connections: 25 through `database/sql` and 10 through `pgxpool`. This is a hard application budget, not a prediction of steady-state usage.

When several PentAGI instances share one database server, add both pool limits for every instance and leave capacity for PostgreSQL reserved connections, autovacuum, monitoring clients, platform services and administrative/migration sessions.

```sql
SELECT name, setting
FROM pg_settings
WHERE name IN ('max_connections', 'superuser_reserved_connections');

SELECT application_name, client_addr, state, count(*)
FROM pg_stat_activity
WHERE pid <> pg_backend_pid()
GROUP BY 1, 2, 3
ORDER BY count(*) DESC;
```

The stock Compose stack uses `vxcontrol/pgvector:latest`. Do not assume a specific PostgreSQL major version from the image tag; inspect `SHOW server_version` on the deployed database.

## Package Layout

| Path | Responsibility |
|---|---|
| `backend/migrations/migrations.go` | Embeds all goose migration SQL into the binary |
| `backend/migrations/sql/` | Authoritative ordered schema and data migration history |
| `backend/sqlc/sqlc.yml` | sqlc input, type overrides and output configuration |
| `backend/sqlc/models/` | Hand-written parameterized SQL queries |
| `backend/pkg/database/db.go` | Generated `DBTX`, `Queries`, `New` and `WithTx` |
| `backend/pkg/database/models.go` | Generated table models and PostgreSQL enum wrappers |
| `backend/pkg/database/querier.go` | Generated `Querier` interface covering all sqlc operations |
| `backend/pkg/database/*.sql.go` | Generated query implementations and result/parameter structs |
| `backend/pkg/database/database.go` | Null helpers, UTF-8 sanitization and shared GORM initialization |
| `backend/pkg/database/tenant.go` | Tenant schema bootstrap, DSN rewrite, search-path verification and advisory locks (shared by pentagi, ftester, etester, installer) |
| `backend/pkg/database/converter/` | Conversion from database rows to GraphQL models plus execution analytics calculations |
| `backend/pkg/database/knowledge/` | Knowledge-store business logic over sqlc, pgvector embeddings and GraphQL subscriptions |
| `backend/pkg/server/models/` | GORM v1 models used by REST/server services |
| `backend/cmd/pentagi/main.go` | Production pool creation, migration execution and dependency wiring |

sqlc and GORM are both active. Controllers and GraphQL paths primarily use generated sqlc queries; REST/server services continue to use GORM models for authentication, users, logs, settings, analytics, resources, flow files and related endpoints. Both clients share the same `*sql.DB`. GORM never owns schema creation.

## Schema and Data Model

### Workflow hierarchy

```text
users
  └── flows
      ├── tasks
      │   └── subtasks
      ├── containers
      ├── assistants
      ├── msgchains
      ├── toolcalls
      ├── screenshots
      └── operational logs
```

There are no relational `actions`, `artifacts` or `memories` tables. Individual operations are represented by tool calls and specialized log tables. Agent memory and knowledge documents are stored in the LangChain pgvector tables.

### Table groups

#### Identity and authorization

| Table | Key fields / notes |
|---|---|
| `users` | Local/OAuth identity (`type`, `mail`, `hash`, `password`, `provider`), `status`, `role_id`, `password_change_required` |
| `roles` | Built-in application roles (seeded in the initial migration) |
| `privileges` | Per-role permission names used by REST and GraphQL authorization; grants evolve through later privilege migrations (not RLS) |
| `api_tokens` | `token_id`, `user_id`, `role_id`, `ttl`, `status`, soft deletion via `deleted_at` |
| `user_preferences` | One JSONB preferences document per user, including favorite-flow state |

#### Workflow and interaction

| Table | Key fields / notes |
|---|---|
| `flows` | Status, title, model, provider name/type, language, functions JSON, `tool_call_id_template`, optional `trace_id`, soft deletion |
| `tasks` | Status, title, input, result; owned by `flow_id` |
| `subtasks` | Status, title, description, result, persisted `context`; owned by `task_id` |
| `containers` | Type (`primary`/`secondary`), name, image, status, optional Docker `local_id`/`local_dir` |
| `assistants` | Flow-scoped interactive assistants with model/provider/functions, `use_agents`, optional `msgchain_id`, soft deletion |
| `msgchains` | LLM chain JSON plus usage (`usage_in`/`out`, cache, cost) and `duration_seconds` |
| `toolcalls` | `call_id`, name, args JSON, result, status, `duration_seconds` |
| `flow_templates` | User-owned reusable flow descriptions (`title`, `text`) |

`flows` and `assistants` support soft deletion. Normal API deletion marks a flow's `deleted_at` and leaves child rows in place for audit/history; foreign-key cascades only run if the flow row is physically deleted.

#### Configuration and user content

| Table | Key fields / notes |
|---|---|
| `providers` | User-owned LLM provider configs (`type`, `name`, `config` JSON), soft deletion |
| `prompts` | User-owned prompt templates keyed by `PROMPT_TYPE` |
| `user_resources` | Uploaded file/directory metadata (`hash`, `name`, `path`, `size`, `is_dir`) |

Flow and assistant rows contain model/provider selection and runtime function configuration. Prompt templates themselves are stored in `prompts`, not in current flow or assistant rows.

#### Logs and artifacts

| Table | Key fields / notes |
|---|---|
| `agentlogs` | Agent-to-agent delegation (`initiator`, `executor`, task/result text) |
| `assistantlogs` | Assistant messages with optional `thinking` and `result_format` |
| `msglogs` | General flow messages with optional `thinking` and `result_format` |
| `searchlogs` | Search engine calls (`engine`, query/result) |
| `termlogs` | Terminal stdin/stdout/stderr; **requires** `container_id` and `flow_id` |
| `vecstorelogs` | Vector-store ops (`action`, filter JSON, query/result) |
| `screenshots` | Screenshot metadata (`name`, `url`); **requires** `flow_id` |

Several log/artifact tables carry nullable `task_id` and `subtask_id` in addition to a required `flow_id`, allowing flow-, task- and subtask-level retrieval.

#### Vector knowledge and memory

| Table | Purpose |
|---|---|
| `langchain_pg_collection` | Logical pgvector collections (`name`, `cmetadata`, `uuid`) |
| `langchain_pg_embedding` | Document text, vector embedding and JSON `cmetadata` |

PentAGI uses the collection named `langchain`. Ownership and association are represented in `cmetadata` fields such as `user_id`, `flow_id`, `task_id`, `subtask_id`, `doc_type`, `question`, `description`, `guide_type`, `answer_type`, `code_lang`, chunk sizing and a `manual` flag. See `backend/pkg/database/knowledge`.

The knowledge query API excludes `doc_type = 'memory'` from user-managed knowledge listings/searches. GraphQL and REST flow-deletion paths issue an explicit best-effort deletion of memory rows (`DeleteFlowMemoryDocuments`); this is application behavior, not a database trigger or foreign-key cascade. The knowledge migration intentionally does not drop the LangChain tables on downgrade because they may contain production data managed by the vector store.

### Enums

PostgreSQL enums are migrated explicitly and generated as Go string types in `models.go`. Current values:

| Enum | Values |
|---|---|
| `FLOW_STATUS` / `TASK_STATUS` / `SUBTASK_STATUS` / `ASSISTANT_STATUS` | `created`, `running`, `waiting`, `finished`, `failed` |
| `CONTAINER_STATUS` | `starting`, `running`, `stopped`, `deleted`, `failed` |
| `CONTAINER_TYPE` | `primary`, `secondary` |
| `TOOLCALL_STATUS` | `received`, `running`, `finished`, `failed` |
| `TOKEN_STATUS` | `active`, `revoked` |
| `USER_STATUS` | `created`, `active`, `blocked` |
| `USER_TYPE` | `local`, `oauth` |
| `MSGCHAIN_TYPE` | `primary_agent`, `reporter`, `generator`, `refiner`, `reflector`, `enricher`, `adviser`, `coder`, `memorist`, `searcher`, `installer`, `pentester`, `summarizer`, `tool_call_fixer`, `assistant` |
| `MSGLOG_TYPE` | `answer`, `report`, `thoughts`, `browser`, `terminal`, `file`, `search`, `advice`, `ask`, `input`, `done` |
| `MSGLOG_RESULT_FORMAT` | `plain`, `markdown`, `terminal` |
| `TERMLOG_TYPE` | `stdin`, `stdout`, `stderr` |
| `VECSTORE_ACTION_TYPE` | `retrieve`, `store` |
| `PROVIDER_TYPE` | `openai`, `anthropic`, `gemini`, `bedrock`, `ollama`, `custom`, `deepseek`, `glm`, `kimi`, `qwen`, `minimax` |
| `SEARCHENGINE_TYPE` | `google`, `tavily`, `firecrawl`, `traversaal`, `browser`, `duckduckgo`, `perplexity`, `searxng`, `sploitus` |
| `PROMPT_TYPE` | Agent/system prompt keys from `primary_agent` through `task_assignment_wrapper` (full list in `models.go`) |

Never add an enum value only in Go code. Add or replace the PostgreSQL enum in a goose migration, regenerate sqlc, and update backend validation where applicable. Provider and search-engine additions have additional project steps documented in `CLAUDE.md`.

### Data lifecycle and integrity

Primary keys use PostgreSQL identity columns. Foreign keys define ownership and use `ON DELETE CASCADE` for records that have no meaning without their parent, including user-owned settings/content and flow-owned execution data. Cascades apply only to physical deletion; ordinary flow, assistant, provider and API-token deletion paths use `deleted_at` where their schema supports soft deletion.

The shared `update_modified_column()` trigger maintains `updated_at` for mutable entities such as flows, tasks, subtasks, containers, tool calls, message chains, assistants, providers, API tokens, preferences, templates and resources. Generated models represent database-default timestamps with `sql.NullTime`, so callers should not assume a non-null Go `time.Time` before insertion/returning.

Important integrity constraints include unique user mail/hash values, one prompt type per user, one preferences row per user, active provider names per user, API-token identifiers, resource paths per user and non-empty template/resource text fields. JSON configuration and chain payloads use PostgreSQL `JSON`; preferences use `JSONB` with a GIN index; optional relationships use nullable SQL columns and generated `sql.Null*` wrappers.

## Application Scoping and Deployment Tenancy

PentAGI has two distinct isolation layers. They solve different problems and must not be confused.

### User scoping inside one PentAGI instance

Rows such as flows, providers, prompts, templates, resources and preferences carry `user_id` directly or are reached through a flow owned by a user. User-facing handlers select user-scoped sqlc methods such as `GetUserFlow` / `GetUserFlows`; admin paths may intentionally use unscoped variants such as `GetFlow` / `GetFlows`.

This access control is implemented by application queries and privilege checks. PentAGI does not rely on PostgreSQL row-level security for its own tables.

### Instance scoping with `TENANT_ID`

`TENANT_ID` isolates independent PentAGI installations that share a PostgreSQL database. It creates one PostgreSQL schema per instance (`public` when empty, otherwise the tenant name). For a tenant, the effective search path is `<tenant>,<DATABASE_EXTENSIONS_SCHEMA>`.

Tenant bootstrap (`backend/pkg/database/tenant.go`):

- validates `TENANT_ID`;
- creates the tenant schema under the `pentagi-tenant-bootstrap` advisory lock;
- ensures `vector` and `pg_trgm` exist in the configured shared extension schema;
- refuses to move provider-managed extensions automatically;
- rewrites `DATABASE_URL` once, before sqlc, GORM, goose or pgxpool consume it;
- verifies `current_schema()` and aborts on mismatch.

The goose version table is schema-qualified as `<schema>.goose_db_version`. This prevents a new tenant from reading `public.goose_db_version`, incorrectly deciding that migrations are already applied, and starting with an empty schema.

For the complete multi-instance deployment contract (non-database resources, validation regex, PgBouncer `connect_query`, Supavisor `DATABASE_SEARCH_PATH_VIA_OPTIONS`, Supabase `DATABASE_EXTENSIONS_SCHEMA=extensions`), see [config.md](config.md#multi-instance-deployment-tenant_id).

## Migrations

Migration files are stored in `backend/migrations/sql` and embedded by `backend/migrations/migrations.go`:

```go
//go:embed sql/*.sql
var EmbedMigrations embed.FS
```

At startup goose uses that filesystem and runs `goose.Up`. Migrations use goose `Up`/`Down` annotations and numeric filename prefixes. The current migration head is determined by the newest migration file; do not hard-code it in application logic.

PentAGI serializes:

1. tenant schema/extension bootstrap with the `pentagi-tenant-bootstrap` advisory lock;
2. migrations with a schema-specific `pentagi-migrations-<schema>` advisory lock.

The lock is held on a dedicated `*sql.Conn` because PostgreSQL session advisory locks belong to a physical connection, not to a `*sql.DB` pool.

### Adding a migration

1. Create a uniquely ordered SQL file in `backend/migrations/sql`.
2. Add `-- +goose Up` and, where safe, a reversible `Down`.
3. Use `-- +goose StatementBegin`/`StatementEnd` for multi-statement units.
4. Preserve tenant compatibility: unqualified application objects must be created in the active tenant schema, while extensions remain in the shared extension schema.
5. Regenerate sqlc if the schema or query types changed.
6. Test both a fresh database and an upgrade from the previous migration head.

Do not edit an already released migration. Add a new migration.

## sqlc Query Layer

### Configuration

`backend/sqlc/sqlc.yml` reads queries from `models/*.sql` and schema from `../migrations/sql/*.sql`, generating package `database` into `../pkg/database` with `emit_interface` and `emit_json_tags`. Notable overrides:

- `pg_catalog.numeric` → `float64`;
- nullable `vector` / `pg_catalog.vector` → `string` (queries cast vector literals explicitly).

Generated files start with `Code generated by sqlc. DO NOT EDIT.` Edit SQL or migrations and regenerate instead.

### Query files

| Query file | Product area | Named queries |
|---|---|---|
| `flows.sql` | Flow CRUD, soft deletion, provider/model updates, flow statistics | `GetFlows`, `GetUserFlows`, `GetFlow`, `GetUserFlow`, `CreateFlow`, `UpdateFlow*`, `DeleteFlow`, `GetFlowStats`, `GetUserTotalFlowsStats`, `GetFlowsStatsByDayLast{Week,Month,3Months}` |
| `tasks.sql` | Task lifecycle and hierarchy-scoped retrieval | `GetFlowTasks`, `GetUserFlowTasks`, `GetFlowTask`, `GetUserFlowTask`, `GetTask`, `CreateTask`, `UpdateTaskStatus/Result/FinishedResult/FailedResult` |
| `subtasks.sql` | Subtask lifecycle, context, planned/completed filters | `GetFlowSubtasks`, `GetFlowTaskSubtasks`, `GetUserFlow*`, `GetTaskSubtasks`, `GetTaskPlannedSubtasks`, `GetTaskCompletedSubtasks`, `GetSubtask`, `GetFlowSubtask`, `CreateSubtask`, `UpdateSubtask*`, `DeleteSubtask(s)` |
| `containers.sql` | Flow container lookup and status | `GetContainers`, `GetUserContainers`, `GetRunningContainers`, `GetFlowContainers`, `GetFlowPrimaryContainer`, `GetUserFlowContainers`, `CreateContainer`, `UpdateContainer*` |
| `assistants.sql` | User/admin assistant access and settings | `GetFlowAssistants`, `GetUserFlowAssistants`, `GetFlowAssistant`, `GetUserFlowAssistant`, `GetAssistant`, `GetAssistantUseAgents`, `CreateAssistant`, `UpdateAssistant*`, `DeleteAssistant` |
| `msgchains.sql` | Conversation chains plus usage analytics | Chain CRUD/lookup by hierarchy/type; `UpdateMsgChainUsage`; aggregates by flow/task/subtask/provider/model/type/day/user |
| `toolcalls.sql` | Tool-call lifecycle and analytics | Hierarchy CRUD/status updates; aggregates by flow/task/subtask/function/day/user |
| `analytics.sql` | Period-based flow selection and hierarchy batches for execution analytics | `GetFlowsForPeriodLast{Week,Month,3Months}`, `GetTasksForFlow`, `GetSubtasksForTasks`, `GetMsgchainsForFlow`, `GetToolcallsForFlow`, `GetAssistantsCountForFlow` |
| `screenshots.sql` | Screenshot retrieval/creation | Flow/user/task/subtask getters + `CreateScreenshot` |
| `agentlogs.sql` | Agent delegation logs | Flow/user/task/subtask getters + `CreateAgentLog` |
| `assistantlogs.sql` | Assistant message/result/thinking logs | Create/update/delete plus flow/user getters |
| `msglogs.sql` | General message logs | Create/update plus flow/user/task/subtask getters |
| `searchlogs.sql` | Search operation logs | Flow/user/task/subtask getters + `CreateSearchLog` |
| `termlogs.sql` | Terminal logs | Container and hierarchy-scoped getters + `CreateTermLog` |
| `vecstorelogs.sql` | Vector-store audit logs | Flow/user/task/subtask getters + `CreateVectorStoreLog` |
| `users.sql` | User identity and administration | `GetUsers`, `GetUser`, `GetUserByHash`, `CreateUser`, `UpdateUser*`, `DeleteUser` |
| `roles.sql` | Roles and privileges | `GetRoles`, `GetRole`, `GetRoleByName` |
| `api_tokens.sql` | Token lifecycle | Admin and user-scoped create/update/soft-delete/list |
| `user_preferences.sql` | Preferences and favorite flows | CRUD/upsert + `AddFavoriteFlow` / `DeleteFavoriteFlow` |
| `providers.sql` | Provider configuration | Admin and user-scoped CRUD/soft-delete, lookup by type/name |
| `prompts.sql` | Prompt templates by `PROMPT_TYPE` | Admin and user-scoped CRUD, lookup/update by type |
| `flow_templates.sql` | Flow templates | User-owned CRUD |
| `resources.sql` | `user_resources` trees (**read-only** sqlc; create/update/delete go through GORM REST services) | Root/dir/recursive/all lookups for one user or all users; lookup by ID(s) |
| `knowledge.sql` | LangChain pgvector documents | Admin/user get/list/update/delete, cosine search, insert, `DeleteFlowMemoryDocuments` |

The generated `Querier` interface in `backend/pkg/database/querier.go` currently exposes 251 methods matching the named SQL queries. Do not hand-edit it.

### Regeneration

From `backend/`, with the PentAGI PostgreSQL network available:

```bash
docker run --rm \
  -v "$(pwd):/src" \
  -w /src \
  --network pentagi-network \
  -e DATABASE_URL='postgres://postgres:postgres@pgvector:5432/pentagidb?sslmode=disable' \
  sqlc/sqlc:1.27.0 generate -f sqlc/sqlc.yml
```

Use credentials matching the target database. Then review generated changes in `pkg/database`, especially `models.go`, `querier.go` and the affected `*.sql.go` file.

### Query conventions

```sql
-- name: GetUserFlow :one
SELECT *
FROM flows
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;
```

Established conventions:

- `Create*`, `Get*`, `Update*`, `Delete*` for basic operations;
- `GetUser*` for user-owned paths;
- `GetFlow*`, `GetTask*`, `GetSubtask*` for hierarchy-scoped paths;
- admin/unscoped methods only where authorization is enforced by the caller;
- filter `deleted_at IS NULL` when querying soft-deletable entities;
- use foreign keys and `ON DELETE CASCADE` for owned child records;
- use `sqlc.arg(...)` for repeated or named parameters.

Parameterized queries protect values from SQL injection. Dynamic identifiers cannot be parameterized and must be validated and quoted separately.

### Transactions

`database.Queries` accepts the `DBTX` interface, so the same generated methods work with `*sql.DB` and `*sql.Tx`:

```go
tx, err := sqlDB.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

qtx := queries.WithTx(tx)
if _, err := qtx.CreateTask(ctx, taskParams); err != nil {
    return err
}
if _, err := qtx.CreateSubtask(ctx, subtaskParams); err != nil {
    return err
}

return tx.Commit()
```

Keep transactions short and pass the caller's context through every query.

## Helpers (`database.go`)

Hand-written helpers used across controllers and GraphQL resolvers:

| Helper | Purpose |
|---|---|
| `NullStringToPtrString` / `PtrStringToNullString` / `StringToNullString` | Nullable string bridging |
| `Int64ToNullInt64` / `Uint64ToNullInt64` / `NullInt64ToInt64` | Nullable integer bridging |
| `TimeToNullTime` / `PtrTimeToNullTime` | Nullable timestamp bridging |
| `SanitizeUTF8` | Strip NUL bytes and replace invalid UTF-8 before storing untrusted tool output |
| `NewGorm` / `GormLogger` | Shared GORM init and optional SQL logging when `DEBUG=true` |

## Converter Package

`backend/pkg/database/converter` maps sqlc rows to GraphQL models and computes execution analytics:

- entity converters for flows, containers/terminals, tasks, subtasks, assistants, screenshots, terminal/message/agent/search/vector/tool-call/assistant logs, prompts, preferences, API tokens, flow templates, user resources;
- provider/model converters between database provider configs, internal `pconfig` structures and GraphQL agent configs (including reasoning mode and call options);
- usage/toolcall/flow stats converters that adapt the typed sqlc analytics rows to GraphQL stats models;
- `BuildFlowExecutionStats` and related helpers in `analytics.go`, which combine tasks, subtasks, message-chain durations, tool-call counts and assistant activity, including overlap compensation for concurrent subtasks.

Analytics calculation behavior belongs in `analytics.go` and its tests. Avoid copying entire generated result structs into documentation; sqlc changes them when query aliases change.

## Knowledge Package

`backend/pkg/database/knowledge` implements the GraphQL knowledge API on top of `knowledge.sql`, an optional LangChain `VectorStore` and an embedding provider:

- admin reads have no `user_id` filter; user-scoped reads filter `cmetadata ->> 'user_id'`;
- writes always record the acting `userID` in metadata and publish scoped subscription events;
- create/update/search require a configured embedder/store; list/get/delete still work without embeddings;
- text sent to the embedding model is truncated to `maxEmbeddingBytes` (default 8192), while the full original text is stored in the database;
- in-memory filters refine SQL results by flow/task/subtask, doc type and related metadata.

## GORM Integration

`database.NewGorm(db, debug)` wraps the existing `*sql.DB`, installs `GormLogger` and enables GORM SQL logging only when `debug` is true. GORM v1 remains in use by server models and handlers; new type-safe database operations should prefer sqlc unless they need existing GORM model behavior.

GORM `AutoMigrate` is not used. Goose migrations are the only supported schema management path. Because sqlc and GORM share one pool, do not call `gorm.Open` elsewhere with the same `DATABASE_URL` without accounting for another pool.

Representative GORM model surfaces in `backend/pkg/server/models/` include users/roles/privileges/preferences, flows/tasks/subtasks/assistants, providers/prompts, API tokens, resources, knowledge request/response DTOs, settings, flow files and the various log entity models used by REST handlers. In particular, `user_resources` mutations are GORM-only; sqlc covers hierarchy listing and ID lookup. API token GraphQL conversion also uses the hand-written `APITokenWithSecret` helper in `backend/pkg/database/api_token_with_secret.go`.

## Vector Operations

PentAGI requires:

- `vector` for embeddings and cosine-distance search;
- `pg_trgm` for GIN trigram indexes on message/log text.

Tenant bootstrap guarantees that both extensions are reachable from every tenant. On a stock deployment they live in `public`; on Supabase they commonly live in `extensions`.

`knowledge.sql` supports admin and user-scoped document retrieval, metadata-only and full document updates, insertion with precomputed embeddings, cosine-similarity search using `<=>`, ownership filtering through `cmetadata ->> 'user_id'`, and cleanup of flow memory documents. Embedding arguments are PostgreSQL vector literals (`[f1,f2,...]`); metadata arguments must be valid JSON text.

The query layer currently filters by collection and metadata. No approximate HNSW/IVFFlat vector index is created by the migrations, so evaluate an appropriate pgvector index before assuming similarity searches will scale linearly to a large corpus.

## Analytics

Usage and execution analytics are derived from relational data rather than a separate warehouse.

### LLM usage (`msgchains`)

Stored fields: input/output tokens, cache input/output tokens, input/output cost, accumulated `duration_seconds`, model, provider and chain type.

Aggregates in `msgchains.sql`: per flow/task/subtask, all flows, by provider/model/type, by type or model-agents for a flow, daily windows (week/month/3 months) and per-user totals.

### Tool-call analytics (`toolcalls`)

Aggregates in `toolcalls.sql`: per flow/task/subtask, all flows, by function (global and per flow), daily windows and per-user totals. Converters distinguish agent tools from ordinary tools where needed for GraphQL presentation.

### Flow counts (`flows`)

`GetFlowStats`, `GetUserTotalFlowsStats` and daily flow-count windows for week/month/3 months.

### Flow execution analytics

`analytics.sql` selects flows for a period and loads hierarchy batches. `converter.BuildFlowExecutionStats` turns that into execution duration models (task/subtask timing, generator/refiner contributions, finished tool-call counts, assistant message-chain time).

## Indexing and Performance

PostgreSQL automatically indexes primary keys and unique constraints. It does **not** automatically index foreign-key columns; PentAGI migrations create the needed foreign-key and query-pattern indexes explicitly.

Current migrations include:

- ownership/hierarchy indexes (`user_id`, `flow_id`, `task_id`, `subtask_id`);
- partial indexes for active soft-deletable rows;
- provider/model/type/time indexes for analytics;
- GIN indexes for JSON preferences;
- trigram GIN indexes for message, result and thinking text;
- path-prefix indexes for user resources.

Large text B-tree indexes on task input/result and subtask description/result were deliberately removed by later migrations. Use full-text or trigram indexing for a concrete query pattern instead of restoring broad B-tree indexes.

Before adding an index: capture the real query and expected cardinality, run `EXPLAIN (ANALYZE, BUFFERS)` on representative data, account for write amplification and index size, add it through a migration, and confirm the generated schema still passes sqlc.

## Observability and Troubleshooting

### Query logging

Set `DEBUG=true` to enable the custom GORM logger. sqlc/libpq queries are not automatically printed by that logger; use PostgreSQL logging, tracing around the caller or a database proxy when those queries need inspection.

Never log `DATABASE_URL`, provider configs, token values or arbitrary SQL arguments containing credentials.

### Tenant schema mismatch

```text
search_path resolved to schema "public", expected "acme"
```

The connection or pooler ignored the tenant search path. Do not bypass the check. Use a direct connection, configure PgBouncer as documented in [config.md](config.md#multi-tenant-postgresql-access-through-pgbouncer), or try `DATABASE_SEARCH_PATH_VIA_OPTIONS=true` for a compatible Supavisor version.

### Extension schema mismatch

```text
extension "vector" is installed in schema "extensions",
but multi-tenant mode requires it in "<DATABASE_EXTENSIONS_SCHEMA>"
```

Set `DATABASE_EXTENSIONS_SCHEMA` to the existing shared extension schema. Do not move provider-managed extensions unless the database operator explicitly requires it. The effective default is `public`.

### No migrations run in a new tenant

Every tenant must have its own `<schema>.goose_db_version`. If a newly created tenant reports the public migration version but has no application tables, verify that the running binary includes schema-qualified goose table handling from `cmd/pentagi/main.go`.

### Constraint and scan errors

- Foreign-key errors usually mean the parent flow/task/subtask was not created or belongs to a different scoped path.
- `sql.ErrNoRows` is expected for missing user-scoped data and should normally be translated to a not-found/access-denied result by the service layer.
- Nullable columns use `sql.Null*` values in generated models; use helpers in `pkg/database/database.go` where they make call sites clearer.
- `SanitizeUTF8` must be applied to untrusted tool output before insert into PostgreSQL text fields.

## Development Checklist

### Adding or changing a query

1. Edit the appropriate file in `backend/sqlc/models`.
2. Keep user/admin scoping explicit in the query name and SQL.
3. Regenerate sqlc.
4. Review generated diffs; never hand-edit them.
5. Add tests at the service, converter or database boundary appropriate to the behavior.
6. Run:

```bash
cd backend
go test ./pkg/database/... ./pkg/server/... ./pkg/graph/...
go vet ./pkg/database/...
```

### Adding a table or field

1. Add a new goose migration.
2. Add explicit indexes for the actual foreign-key/query patterns.
3. Add or update sqlc queries.
4. Regenerate sqlc.
5. Update GraphQL/GORM converters only if the field crosses those boundaries.
6. Test fresh install, upgrade and tenant-schema startup.

### Review points

- Does every user-facing path enforce ownership?
- Is an unscoped/admin query intentionally authorized by its caller?
- Are soft-deleted rows filtered where expected?
- Does the migration work in a non-`public` tenant schema?
- Are shared extensions referenced through the configured search path?
- Does the change fit within the two-pool connection budget?
- Are generated files free of manual edits?

## Related Documentation

- [config.md](config.md) — database environment variables, tenancy, PgBouncer and Supavisor configuration
- [README.md](../../README.md) — deployment and development commands
- `backend/sqlc/sqlc.yml` — sqlc generation settings
- `backend/migrations/sql` — authoritative schema history
- `backend/pkg/database` — generated API and database helpers
