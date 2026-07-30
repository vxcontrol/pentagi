# PentAGI Configuration Guide

This document serves as a comprehensive guide to the configuration system in PentAGI, primarily aimed at developers. It details all available configuration options, their purposes, default values, and how they're used throughout the application.

## Table of Contents

- [PentAGI Configuration Guide](#pentagi-configuration-guide)
  - [Table of Contents](#table-of-contents)
  - [Configuration Basics](#configuration-basics)
    - [Current Web Settings Coverage](#current-web-settings-coverage)
    - [Still Server-Managed](#still-server-managed)
  - [General Settings](#general-settings)
    - [Multi-Instance Deployment (`TENANT_ID`)](#multi-instance-deployment-tenant_id)
      - [What the application namespaces automatically](#what-the-application-namespaces-automatically)
      - [What stays the operator's responsibility](#what-stays-the-operators-responsibility)
      - [Deployment topologies](#deployment-topologies)
      - [Extensions installed outside `public` (`DATABASE_EXTENSIONS_SCHEMA`)](#extensions-installed-outside-public-database_extensions_schema)
      - [Multi-tenant PostgreSQL access through PgBouncer](#multi-tenant-postgresql-access-through-pgbouncer)
      - [Multi-tenant PostgreSQL through Supabase's Supavisor pooler (`DATABASE_SEARCH_PATH_VIA_OPTIONS`)](#multi-tenant-postgresql-through-supabases-supavisor-pooler-database_search_path_via_options)
    - [Usage Details](#usage-details)
  - [Docker Settings](#docker-settings)
    - [Worker Docker Access (`DOCKER_INSIDE_*`)](#worker-docker-access-docker_inside_)
    - [Usage Details](#usage-details-1)
  - [Server Settings](#server-settings)
    - [Usage Details](#usage-details-2)
  - [Frontend Settings](#frontend-settings)
    - [Usage Details](#usage-details-3)
  - [Authentication Settings](#authentication-settings)
    - [Usage Details](#usage-details-4)
  - [Web Scraper Settings](#web-scraper-settings)
    - [Usage Details](#usage-details-5)
  - [LLM Provider Settings](#llm-provider-settings)
    - [OpenAI](#openai)
    - [Anthropic](#anthropic)
    - [Ollama LLM Provider](#ollama-llm-provider)
    - [Google AI (Gemini) LLM Provider](#google-ai-gemini-llm-provider)
    - [AWS Bedrock LLM Provider](#aws-bedrock-llm-provider)
    - [DeepSeek LLM Provider](#deepseek-llm-provider)
    - [GLM LLM Provider](#glm-llm-provider)
    - [Kimi LLM Provider](#kimi-llm-provider)
    - [Qwen LLM Provider](#qwen-llm-provider)
    - [MiniMax LLM Provider](#minimax-llm-provider)
    - [Custom LLM Provider](#custom-llm-provider)
    - [Usage Details](#usage-details-6)
  - [Embedding Settings](#embedding-settings)
    - [Usage Details](#usage-details-7)
  - [Summarizer Settings](#summarizer-settings)
    - [Usage Details and Impact on System Behavior](#usage-details-and-impact-on-system-behavior)
      - [Core Summarization Strategies and Their Parameters](#core-summarization-strategies-and-their-parameters)
      - [Deep Dive: Parameter Impact and Recommendations](#deep-dive-parameter-impact-and-recommendations)
    - [Summarization Effects on Agent Behavior](#summarization-effects-on-agent-behavior)
    - [Implementation Details](#implementation-details)
    - [Recommended Settings for Different Use Cases](#recommended-settings-for-different-use-cases)
  - [Assistant Settings](#assistant-settings)
    - [Usage Details](#usage-details-8)
    - [Recommended Assistant Settings for Different Use Cases](#recommended-assistant-settings-for-different-use-cases)
  - [Functions Configuration](#functions-configuration)
    - [DisableFunction Structure](#disablefunction-structure)
    - [ExternalFunction Structure](#externalfunction-structure)
    - [Usage Details](#usage-details-9)
    - [Example Configuration](#example-configuration)
    - [Security Considerations](#security-considerations)
    - [Built-in Functions Reference](#built-in-functions-reference)
  - [Search Engine Settings](#search-engine-settings)
    - [DuckDuckGo Search](#duckduckgo-search)
    - [Sploitus Search](#sploitus-search)
    - [Google Search](#google-search)
    - [Traversaal Search](#traversaal-search)
    - [Tavily Search](#tavily-search)
    - [Firecrawl Search](#firecrawl-search)
    - [Perplexity Search](#perplexity-search)
    - [Searxng Search](#searxng-search)
    - [Internal Analytics Engine](#internal-analytics-engine)
    - [Usage Details](#usage-details-10)
  - [Network and Proxy Settings](#network-and-proxy-settings)
    - [Usage Details](#usage-details-11)
  - [Graphiti Knowledge Graph Settings](#graphiti-knowledge-graph-settings)
    - [PentAGI Configuration Boundary](#pentagi-configuration-boundary)
    - [Client Lifecycle and Failure Behavior](#client-lifecycle-and-failure-behavior)
    - [Data Flow, Search, and Tenancy](#data-flow-search-and-tenancy)
    - [Deployment Ownership](#deployment-ownership)
  - [Agent Supervision Settings](#agent-supervision-settings)
    - [Usage Details](#usage-details-13)
    - [Supervision System Integration](#supervision-system-integration)
    - [Recommended Settings](#recommended-settings)
  - [Observability Settings](#observability-settings)
    - [Telemetry](#telemetry)
    - [Langfuse](#langfuse)
    - [Usage Details](#usage-details-14)

## Configuration Basics

PentAGI uses environment variables for configuration, with support for `.env` files through the `godotenv` package. The configuration is defined in the `Config` struct in `pkg/config/config.go` and is loaded using the `NewConfig()` function.

```go
func NewConfig() (*Config, error) {
    godotenv.Load()

    var config Config
    if err := env.ParseWithOptions(&config, env.Options{
        RequiredIfNoDef: false,
        FuncMap: map[reflect.Type]env.ParserFunc{
            reflect.TypeOf(&url.URL{}): func(s string) (interface{}, error) {
                if s == "" {
                    return nil, nil
                }
                return url.Parse(s)
            },
        },
    }); err != nil {
        return nil, err
    }

    return &config, nil
}
```

This function automatically loads environment variables from a `.env` file if present, then parses them into the `Config` struct using the `env` package from `github.com/caarlos0/env/v10`.

### Current Web Settings Coverage

The running PentAGI instance already exposes several settings areas in the web UI:

- **Settings -> Providers**: Manage user-defined provider profiles, per-agent model and runtime options, and provider test actions for provider types supported by the running server.
- **Settings -> Prompts**: Manage system, human, and tool prompt templates.
- **Settings -> PentAGI API**: Create, revoke, and delete PentAGI API tokens.
- **Other UI-managed preferences**: Favorite flows are stored as user preferences, and theme selection is handled client-side from the main sidebar/profile controls.

These web-console features do not replace the environment variables in this guide for provider credentials, endpoints, or external integrations.

### Still Server-Managed

The environment variables documented below remain the source of truth for configuration that is not currently editable from the web console:

- **LLM credentials and connection settings**: API keys, base URLs, auth modes, and provider-specific connection settings for OpenAI, Anthropic, Bedrock, Ollama, custom providers, and similar backends; config-path settings apply only where supported, such as `OLLAMA_SERVER_CONFIG_PATH`, `LLM_SERVER_CONFIG_PATH`, and `BEDROCK_CONFIG_PATH`.
- **Search provider credentials and options**: DuckDuckGo, Google, Tavily, Traversaal, Perplexity, Searxng, Sploitus, and related search configuration.
- **Third-party integrations**: Langfuse, Graphiti, and other external observability or knowledge services.
- **MCP server management**: MCP settings are not currently exposed as a live web-console feature.

## General Settings

These settings control basic application behavior and are foundational for the system's operation.

| Option           | Environment Variable        | Default Value                                                                | Description                                                              |
| ---------------- | --------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| DatabaseURL      | `DATABASE_URL`              | `postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable` | Connection string for the PostgreSQL database with pgvector extension    |
| DatabaseExtensionsSchema | `DATABASE_EXTENSIONS_SCHEMA` | `public`                                                              | Schema every tenant's search_path must include for shared extensions (`vector`, `pg_trgm`) to resolve. Only used when `TenantID` is set. Override for databases that install extensions elsewhere by convention, e.g. Supabase uses `extensions`. See [Extensions installed outside `public`](#extensions-installed-outside-public-database_extensions_schema). |
| DatabaseSearchPathViaOptions | `DATABASE_SEARCH_PATH_VIA_OPTIONS` | `false`                                                          | Sends the tenant search_path as `options=--search_path=<value>` instead of a bare `search_path` parameter. Only used when `TenantID` is set. For poolers that forward `options` but drop an unrecognized bare `search_path` (e.g. some Supabase Supavisor versions). See [Multi-tenant PostgreSQL through Supabase's Supavisor pooler](#multi-tenant-postgresql-through-supabases-supavisor-pooler-database_search_path_via_options). |
| DBMaxOpenConns   | `DATABASE_MAX_OPEN_CONNS`   | `25`                                                                         | Maximum open connections in the shared `sql.DB` pool (sqlc + GORM combined). See [database.md §Connection Pooling](database.md#connection-pooling). |
| DBMaxIdleConns   | `DATABASE_MAX_IDLE_CONNS`   | `5`                                                                          | Maximum idle connections kept open between requests                      |
| DBVectorMaxConns | `DATABASE_VECTOR_MAX_CONNS` | `10`                                                                         | Maximum connections in the shared `pgxpool` for all pgvector stores      |
| Debug            | `DEBUG`                     | `false`                                                                      | Enables debug mode with additional logging                               |
| DataDir          | `DATA_DIR`                  | `./data`                                                                     | Directory for storing persistent data                                    |
| AskUser          | `ASK_USER`                  | `false`                                                                      | When enabled, requires explicit user confirmation for certain operations |
| TenantID         | `TENANT_ID`                 | *(empty)*                                                                    | Namespaces every externally-visible artifact this instance creates so several PentAGI instances can share one host and one set of backing services. Empty = single-instance behavior, unchanged. See [Multi-Instance Deployment](#multi-instance-deployment-tenant_id). |
| DockerPortsBase  | `DOCKER_PORTS_BASE`         | `0` (means `28000`)                                                          | First host port for per-flow sandbox port publishing; each instance owns `[base, base+2000)` |
| InstallationID   | `INSTALLATION_ID`           | *(none)*                                                                     | Unique installation identifier for PentAGI Cloud API communication       |
| LicenseKey       | `LICENSE_KEY`               | *(none)*                                                                     | License key for PentAGI Cloud API authentication and feature activation  |

### Multi-Instance Deployment (`TENANT_ID`)

`TENANT_ID` namespaces the artifacts a PentAGI instance creates in **shared external services**, so that several independent installations can use one PostgreSQL server, one worker node (Docker daemon), one Neo4j/Graphiti and one Langfuse without colliding.

**Primary use case: several management instances, shared resources.** The typical deployment puts each PentAGI management backend on its own server, while the heavy shared resources — the worker node where sandbox containers run, and the database — are common. `TENANT_ID` is what keeps those instances from writing over each other.

Running several management instances on **one** server is also possible (for example behind an nginx reverse proxy), but that is not what the installer or the stock `docker-compose.yml` are built for — see *Deployment topologies* below.

**When empty (the default) nothing changes.** Every helper degrades to an identity function: container names stay `pentagi-terminal-<id>`, the database schema stays `public`, Graphiti group ids stay `flow-<id>`, and the session cookie stays `auth`. This is enforced in one place (`backend/pkg/config/tenant.go`) and covered by tests.

**Validation.** `TENANT_ID` must match `^[a-z][a-z0-9_]{0,31}$` — the intersection of the constraints imposed by every consumer:

| Consumer | Constraint |
| --- | --- |
| Docker object name | `[a-zA-Z0-9][a-zA-Z0-9_.-]*` |
| PostgreSQL identifier | ≤ 63 bytes; unquoted-safe as `[a-z_][a-z0-9_]*` |
| Graphiti / Neo4j group id | parsed on a hyphen boundary, so no hyphens |

Hyphens are excluded deliberately: they separate the tenant from the rest of a group id or object name. An invalid value **aborts startup** rather than being normalised, because collapsing two distinct tenants onto one namespace is exactly the collision tenancy exists to prevent.

#### What the application namespaces automatically

| Area | Effect |
| --- | --- |
| PostgreSQL | A schema named after the tenant is created on boot and `search_path` is set to `<tenant>,<DATABASE_EXTENSIONS_SCHEMA>`; the DSN is rewritten once so sqlc, GORM, goose and the pgvector pool all follow. Extensions (`vector`, `pg_trgm`) stay shared in `DATABASE_EXTENSIONS_SCHEMA` (default `public`). |
| Worker containers | Sandbox container names become `<tenant>-pentagi-terminal-<flow>`; the per-flow volume and the container hostname derive from that name automatically. Both are labelled `pentagi.tenant`, so daemon-wide sweeps can filter by owner. |
| Host ports | Per-flow sandbox ports are allocated from `DOCKER_PORTS_BASE` (default `28000`), giving each instance the window `[base, base+2000)`. |
| Knowledge graph | Graphiti/Neo4j group ids become `<tenant>-flow-<id>`. The GraphQL API contract is unchanged — clients still send `flow-<id>` and the server rebuilds the namespaced key internally. |
| Auth | Cookie and JWT keys are derived from `COOKIE_SIGNING_SALT` **plus** the tenant, and the session cookie is renamed, so a session or API token minted by one instance is rejected by another. |
| Telemetry | OTel resources carry `service.instance.id` and, with a tenant set, `tenant_id`; Langfuse traces carry the native `environment` field plus a `tenant:<id>` tag and tenant-prefixed trace/session names. |

#### What stays the operator's responsibility

`TENANT_ID` does **not** rewrite infrastructure-level settings. The following must be given distinct values per instance by whoever provisions it:

| Setting | Why it is not derived from `TENANT_ID` |
| --- | --- |
| `DATA_DIR` | The data directory is chosen by the operator. Flow file caches, the container `/work` bind mount, screenshots and `installation_id` all live under it, and two instances pointed at the same directory **will overwrite each other's flow data**. Give each instance its own path or its own volume. |
| `DOCKER_NETWORK` | Instances may legitimately share one Docker network; the application never renames it. Set it explicitly if you want separate networks. |
| Published ports | `PENTAGI_LISTEN_PORT`, `PGVECTOR_LISTEN_PORT`, `SCRAPER_LISTEN_PORT`, `PPROF_ADDR` and `DOCKER_PORTS_BASE` are host-level resources, not string namespaces. |
| `INSTALLATION_ID` | Unique per installation, or left empty to be generated once and cached in `DATA_DIR`. |
| Database privileges | The configured user needs `CREATE SCHEMA`, and on first boot `CREATE EXTENSION` — unless an administrator pre-installed `vector` and `pg_trgm` into `DATABASE_EXTENSIONS_SCHEMA` (default `public`). |

The values actually in effect are written to the startup log under `Instance identity` (`tenant_id`, `data_dir`, `schema`, `installation_id`), which is the quickest way to confirm two instances are not sharing something they should not.

`DOCKER_INSIDE` is orthogonal to tenancy and may be enabled alongside it.

#### Deployment topologies

| Topology | Supported by |
| --- | --- |
| One instance per server, shared PostgreSQL and/or worker node | The installer and the stock `docker-compose.yml`, with `TENANT_ID` set per server. This is the intended setup. |
| Several instances on one server | Manual configuration only. The stock `docker-compose.yml` uses fixed `container_name` and network names, so copying it verbatim will not start a second stack. Adapt it to your network and infrastructure — for example, distinct container names behind a shared nginx — and give each instance its own `DATA_DIR` and ports. |

The installer provisions **one** instance per server; it does not manage several side by side. Note that it does scope the sandbox resources it cleans up — worker containers and volumes are matched by the tenant prefix — so a purge run for one management instance will not remove another's sandboxes from a shared worker node.

**Upgrading an existing deployment.** Leave `TENANT_ID` empty. The instance keeps using `public` and its current data directory; no migration is required. Setting `TENANT_ID` on an existing installation points it at a **new, empty schema** — the data in `public` is not migrated and will appear to be gone. Do not point an existing `public` deployment at a `search_path` that lists another tenant's schema.

#### Extensions installed outside `public` (`DATABASE_EXTENSIONS_SCHEMA`)

`ensureTenantSchema` requires `vector` and `pg_trgm` to already live in (or be creatable in) one schema common to every tenant, because it must be part of every tenant's `search_path`. That schema defaults to `public`, which is where a stock PostgreSQL/`docker-compose.yml` install keeps them.

Managed providers do not always follow that convention. **Supabase** (cloud and self-hosted) installs its bundled extensions into a dedicated `extensions` schema instead, and its default roles get `extensions` added to their `search_path` for exactly that reason. Pointing `DATABASE_URL` at such a database with `TENANT_ID` set fails fast on boot:

```
Tenant schema initialization failed: extension "vector" is installed in schema "extensions",
but multi-tenant mode requires it in "public" so every tenant can reach it; either run
ALTER EXTENSION vector SET SCHEMA public, or set DATABASE_EXTENSIONS_SCHEMA=extensions
to match where it already lives
```

Set `DATABASE_EXTENSIONS_SCHEMA=extensions` (or whatever schema the error reports) instead of moving the extension with `ALTER EXTENSION ... SET SCHEMA` — the schema only needs to be part of the `search_path` PentAGI computes (`<tenant>,<DATABASE_EXTENSIONS_SCHEMA>`), moving a provider-managed extension out of its documented location is unnecessary and risks breaking whatever else that provider expects to find it there.

#### Multi-tenant PostgreSQL access through PgBouncer

`DATABASE_URL` can point at a PgBouncer instance instead of PostgreSQL directly, but three things have to be true, independent of each other:

1. **`pool_mode = session` on the PgBouncer side.** PentAGI holds a `pg_advisory_lock`/`pg_advisory_unlock` pair on one dedicated connection across the whole tenant-bootstrap + migration sequence (`backend/pkg/database/tenant.go`), and `pgx` (used by the pgvector pool) caches server-side prepared statements by default. Both break silently under `transaction`/`statement` pooling, because PgBouncer is then free to hand the client a different backend connection between statements. This requirement is unrelated to tenancy — it applies even with `TENANT_ID` empty.
2. **`ignore_startup_parameters = search_path` in `pgbouncer.ini`.** With a tenant configured, PentAGI's own DSN already carries `?search_path=<tenant>,<DATABASE_EXTENSIONS_SCHEMA>` (see the table above). PgBouncer validates startup parameters from the client against a small built-in allowlist and rejects anything else with `unsupported startup parameter: search_path` unless it is explicitly ignored. Ignoring it does **not** apply the value — it only stops PgBouncer from rejecting the connection — so this step alone is not sufficient; see the next point.
3. **A `connect_query` per tenant in PgBouncer's `[databases]` section**, since PentAGI's own `search_path` startup parameter is ignored per point 2 above. `connect_query` runs on PgBouncer's own connection to PostgreSQL before any client statement, so it is not subject to the client-facing startup-parameter allowlist and works regardless of pool mode:

   ```ini
   [databases]
   pentagi_testing = host=pgvector port=5432 dbname=pentagidb pool_mode=session connect_query='SET search_path TO testing,public'
   pentagi_acme    = host=pgvector port=5432 dbname=pentagidb pool_mode=session connect_query='SET search_path TO acme,public'
   ```

   Point each instance's `DATABASE_URL` at its own virtual database name (`pentagi_testing`, `pentagi_acme`, ...) rather than the shared `pentagidb` — PgBouncer pools per `(user, dbname)` pair, so distinct virtual names are what keeps the tenants' pools, and therefore their `connect_query`, apart.

   (`track_extra_parameters = search_path` is PgBouncer's other mechanism for this, but it only works when PostgreSQL reports `search_path` changes back to the client, which requires PostgreSQL 18+ or Citus 12+ — not an option against the PostgreSQL 16/17 that ships in `docker-compose.yml`.)

#### Multi-tenant PostgreSQL through Supabase's Supavisor pooler (`DATABASE_SEARCH_PATH_VIA_OPTIONS`)

Supabase's shared/self-hosted pooler (Supavisor) is not PgBouncer and none of its `[databases]`/`connect_query` configuration exists for it, so the PgBouncer recipe above does not apply. Reports on whether Supavisor forwards a tenant's `search_path` at all are inconsistent — see [supabase/supavisor#206](https://github.com/supabase/supavisor/issues/206) — and depend on the Supavisor version (a parsing fix landed in [PR #768](https://github.com/supabase/supavisor/pull/768)).

If bypassing the pooler entirely (connecting straight to the underlying PostgreSQL, or to a Supabase project's "Direct connection"/IPv4-add-on string) is not an option, set:

```
DATABASE_SEARCH_PATH_VIA_OPTIONS=true
```

This sends the tenant's search_path as `options=--search_path=<tenant>,<DATABASE_EXTENSIONS_SCHEMA>` instead of a bare `search_path=` parameter. Some poolers forward the `options` startup parameter through to the real backend while silently dropping an unrecognized bare `search_path` — this is exactly the workaround reported to work against some Supavisor versions. **It is not guaranteed** — verify it actually took effect by checking that the app starts (`verifySearchPath` fails fast with a clear error if it did not) rather than assuming success from the flag alone.

This flag changes nothing for a direct PostgreSQL connection or a PgBouncer setup already following the recipe above; both accept `search_path` and `options` equally, so there is no reason to enable it outside a Supavisor-fronted deployment.

### Usage Details

- **DatabaseURL**: This is a critical setting used throughout the application for all database connections. It is used to:
  - Initialize the single shared `sql.DB` connection pool in `main.go` (used by both sqlc `Queries` and GORM)
  - Seed the shared `pgxpool.Pool` for all pgvector stores (agent memory + knowledge API)

  For connection pool sizing and operational monitoring commands see [database.md §Connection Pooling](database.md#connection-pooling).

```go
// In main.go — one pool shared by sqlc Queries and GORM
db, err := sql.Open("postgres", cfg.DatabaseURL)
db.SetMaxOpenConns(cfg.DBMaxOpenConns)
db.SetMaxIdleConns(cfg.DBMaxIdleConns)
db.SetConnMaxLifetime(time.Hour)

queries := database.New(db)
orm, err := database.NewGorm(db)   // GORM wraps the same *sql.DB

// Shared pgxpool for all pgvector stores
pgPoolConfig, _ := pgxpool.ParseConfig(cfg.DatabaseURL)
pgPoolConfig.MaxConns = int32(cfg.DBVectorMaxConns)
pgPool, _ := pgxpool.NewWithConfig(ctx, pgPoolConfig)
cfg.PgxPool = pgPool               // passed to tools and router
```

- **Debug**: Controls debug mode throughout the application, enabling additional logging and development features:
  - Activates detailed logging in the router setup
  - Can enable development endpoints and tools

```go
// In router.go for enabling debug mode
if cfg.Debug {
    // Enable debug features
}
```

- **DataDir**: Specifies where PentAGI stores persistent data. This is used across multiple components:
  - In `docker/client.go` for container volume mapping
  - For screenshots storage in `services.NewScreenshotService`
  - In tools for file operations and data persistence
  - In Docker container management for mapping volumes

```go
// In docker/client.go
dataDir, err := filepath.Abs(cfg.DataDir)

// In router.go for screenshot service
screenshotService := services.NewScreenshotService(orm, cfg.DataDir)

// In tools.go for various tools
dataDir: fte.cfg.DataDir
```

- **AskUser**: A safety feature that, when enabled, requires explicit user confirmation before executing potentially destructive operations:
  - Used in tools to prompt for confirmation before executing commands
  - Serves as a safeguard for sensitive operations

```go
// In tools.go
if fte.cfg.AskUser {
    // Prompt user for confirmation before executing
}
```

- **InstallationID**: A unique identifier for the PentAGI installation used for cloud API communication:
  - Generated automatically during installation or can be manually set
  - Required for certain cloud-based features and integrations

```go
// Used in cloud SDK initialization
if cfg.InstallationID != "" {
    // Initialize cloud API client with installation ID
}
```

- **LicenseKey**: Authentication key for PentAGI Cloud API and premium feature activation:
  - Validates license and enables licensed features
  - Required for enterprise features and support
  - Used for authentication with PentAGI Cloud services

```go
// Used in cloud SDK initialization
if cfg.LicenseKey != "" {
    // Validate license and activate premium features
}
```

## Docker Settings

These settings control how PentAGI interacts with Docker, which is used for terminal isolation and executing commands in a controlled environment. They're crucial for the security and functionality of tool execution.

| Option                       | Environment Variable               | Default Value          | Description |
| ---------------------------- | ---------------------------------- | ---------------------- | ----------- |
| DockerInside                 | `DOCKER_INSIDE`                    | `false`                | Set to `true` if PentAGI runs inside Docker and needs to access the host Docker daemon. |
| DockerNetAdmin               | `DOCKER_NET_ADMIN`                 | `false`                | Set to `true` to grant the primary container NET_ADMIN capability for advanced networking. |
| DockerSocket                 | `DOCKER_SOCKET`                    | *(none)*               | Path to Docker socket for container management |
| DockerInsideHost             | `DOCKER_INSIDE_HOST`               | *(none)*               | Docker daemon endpoint given to worker containers; also disables host-socket autodetection. See [Worker Docker Access](#worker-docker-access-docker_inside_) |
| DockerInsideTLSVerify        | `DOCKER_INSIDE_TLS_VERIFY`         | *(none)*               | TLS verification for the worker container's Docker connection |
| DockerInsideCertPath         | `DOCKER_INSIDE_CERT_PATH`          | *(none)*               | TLS certificate directory **on the worker node**, mounted read-only into worker containers |
| DockerNetwork                | `DOCKER_NETWORK`                   | *(none)*               | Docker network name for bridge mode, or `host` for host network mode. See network modes below. |
| DockerPublicIP               | `DOCKER_PUBLIC_IP`                 | `0.0.0.0`              | Public IP address for Docker containers' port bindings (bridge mode only) |
| DockerWorkDir                | `DOCKER_WORK_DIR`                  | *(none)*               | Custom working directory inside Docker containers |
| DockerDefaultImage           | `DOCKER_DEFAULT_IMAGE`             | `debian:latest`        | Default Docker image for containers when specific images fail |
| DockerDefaultImageForPentest | `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` | `vxcontrol/kali-linux` | Default Docker image for penetration testing tasks |
| TerminalToolTimeout          | `TERMINAL_TOOL_TIMEOUT`            | `1200`                 | Default execution timeout in seconds applied when an agent requests `timeout=0` or a negative value. Accepted range: `1`–`10800` (3 hours). Values `<= 0` or above `10800` are clamped to the 3-hour maximum. Negative values are treated identically to `0`. |

### Worker Docker Access (`DOCKER_INSIDE_*`)

`DOCKER_HOST`, `DOCKER_TLS_VERIFY` and `DOCKER_CERT_PATH` tell **PentAGI itself** which daemon to create worker containers on. The `DOCKER_INSIDE_*` trio is the mirror image: it tells a **worker container** which daemon *it* may talk to. Both sets are independent — sandboxes can be pointed at a different daemon than the one that spawned them.

They are read only when `DOCKER_INSIDE=true`; with it disabled the sandbox gets no Docker access and no Docker configuration whatsoever.

**How the socket is chosen.** With `DOCKER_INSIDE=true`:

| `DOCKER_SOCKET` | `DOCKER_INSIDE_HOST` | Result |
| --- | --- | --- |
| set | any | That socket is bind-mounted at `/var/run/docker.sock` — historical behaviour, an explicit socket always wins. |
| empty | set | **Nothing is mounted.** The sandbox reaches Docker over `DOCKER_INSIDE_HOST` instead. |
| empty | empty | The host socket is autodetected and mounted — historical behaviour. |

The middle row is the point of the feature: mounting the host socket into a sandbox gives an autonomous agent control of the daemon running PentAGI itself, including every other flow's containers. Designating a separate endpoint — a DinD sidecar, a remote daemon, a socket proxy — keeps that authority out of the sandbox.

**What is injected.** Every non-empty `DOCKER_INSIDE_*` value is passed into the container as an environment variable with the `_INSIDE_` segment removed, so the Docker CLI inside picks it up with no extra configuration:

| Configured | Seen inside the worker container |
| --- | --- |
| `DOCKER_INSIDE_HOST=tcp://dind:2376` | `DOCKER_HOST=tcp://dind:2376` |
| `DOCKER_INSIDE_TLS_VERIFY=1` | `DOCKER_TLS_VERIFY=1` |
| `DOCKER_INSIDE_CERT_PATH=/certs/client` | `DOCKER_CERT_PATH=/certs/client` |

Empty values are omitted rather than injected blank.

When `DOCKER_INSIDE_CERT_PATH` is set, that directory is additionally bind-mounted **read-only at the same path** inside the container, so the injected `DOCKER_CERT_PATH` resolves unchanged. The path is resolved on the **worker node** — the machine whose daemon creates sandboxes — which may not be the machine running PentAGI or the installer. For that reason the installer does not verify it exists.

**Example — sandboxes use a DinD sidecar over mutual TLS:**

```bash
DOCKER_INSIDE=true
DOCKER_SOCKET=                      # leave empty so the host socket is not mounted
DOCKER_INSIDE_HOST=tcp://dind:2376
DOCKER_INSIDE_TLS_VERIFY=1
DOCKER_INSIDE_CERT_PATH=/certs/client
```

### Usage Details

The Docker settings are primarily used in `pkg/docker/client.go` which implements the Docker client interface used throughout the application. This client is responsible for creating, managing, and executing commands in Docker containers:

- **DockerInside**: Signals whether PentAGI is running inside a Docker container itself, which affects how volumes and sockets are mounted:
  ```go
  inside := cfg.DockerInside
  ```

- **DockerSocket**: Specifies the path to the Docker socket, which is crucial for container management:
  ```go
  if cfg.DockerSocket != "" {
      socket = cfg.DockerSocket
  }
  ```

- **TerminalToolTimeout**: Sets the default execution timeout for terminal tool commands when the tool call uses `timeout=0` or a negative value:
  ```go
  term := NewTerminalTool(
      flowID,
      taskID,
      subtaskID,
      containerID,
      containerLID,
      dockerClient,
      termLogProvider,
      time.Duration(cfg.TerminalToolTimeout)*time.Second,
  )
  ```

  The value is clamped inside the terminal tool: values `<= 0` or above `10800` s (3 hours) are silently raised/capped to the 3-hour maximum — agents always receive a finite timeout. Negative values are accepted at the environment level and treated identically to `0` (both resolve to the 3-hour ceiling). Explicit `timeout` values provided by the tool call override this default when they are within the `1`–`10800` s range.

- **DockerNetwork**: Controls the network isolation mode for containers. Supports two modes:
  
  **Bridge Mode** (custom network name, e.g., `pentagi-network`):
  - Containers run in an isolated bridge network
  - Port forwarding maps container ports to host ports
  - Enhanced security through network isolation
  - Recommended for most deployments
  
  **Host Mode** (special value: `host`):
  - Containers share the host's network stack directly
  - No port forwarding - services bind directly to host interfaces
  - Required for advanced network testing (raw packets, custom protocols)
  - Reduced isolation - use with caution
  
  ```go
  network := cfg.DockerNetwork

  // Host network mode
  if dc.network == "host" {
      hostConfig.NetworkMode = container.NetworkMode("host")
      // No port bindings needed
  } else if dc.network != "" {
      // Bridge mode with custom network
      networkingConfig = &network.NetworkingConfig{
          EndpointsConfig: map[string]*network.EndpointSettings{
              dc.network: {},
          },
      }
      // Port bindings are configured
  }
  ```

- **DockerPublicIP**: Defines the IP address to bind container ports to, making services accessible:
  ```go
  publicIP := cfg.DockerPublicIP

  // Used when setting up port bindings
  hostConfig.PortBindings[natPort] = []nat.PortBinding{
      {
          HostIP:   dc.publicIP,
          HostPort: fmt.Sprintf("%d", port),
      },
  }
  ```

- **DockerWorkDir**: Provides a custom working directory path to use inside containers:
  ```go
  hostDir := getHostDataDir(ctx, cli, dataDir, cfg.DockerWorkDir)
  ```

- **DockerDefaultImage**: Specifies the fallback image to use when requested images aren't available:
  ```go
  defImage := strings.ToLower(cfg.DockerDefaultImage)
  if defImage == "" {
      defImage = defaultImage
  }
  ```

This client is used by the tools executor to run commands in isolated containers, providing a secure environment for AI agents to execute terminal commands.

## Server Settings

These settings control the HTTP and GraphQL server that forms the backend API of PentAGI.

| Option       | Environment Variable | Default Value | Description                      |
| ------------ | -------------------- | ------------- | -------------------------------- |
| ServerPort   | `SERVER_PORT`        | `8080`        | Port for the HTTP server         |
| ServerHost   | `SERVER_HOST`        | `0.0.0.0`     | Host address for the HTTP server |
| ServerUseSSL | `SERVER_USE_SSL`     | `false`       | Enable SSL for the HTTP server   |
| ServerSSLKey | `SERVER_SSL_KEY`     | *(none)*      | Path to SSL key file             |
| ServerSSLCrt | `SERVER_SSL_CRT`     | *(none)*      | Path to SSL certificate file     |

### Usage Details

These settings are used in `main.go` to configure and start the HTTP server:

```go
// Build the listen address from host and port
listen := net.JoinHostPort(cfg.ServerHost, strconv.Itoa(cfg.ServerPort))

// Conditionally use TLS based on SSL configuration
if cfg.ServerUseSSL && cfg.ServerSSLCrt != "" && cfg.ServerSSLKey != "" {
    err = r.RunTLS(listen, cfg.ServerSSLCrt, cfg.ServerSSLKey)
} else {
    err = r.Run(listen)
}
```

The settings determine:
- The IP address and port the server listens on
- Whether to use HTTPS (SSL/TLS) for secure connections
- The location of the SSL certificate and key files (when SSL is enabled)

These configurations are crucial for production deployments where proper server binding and secure communication are required.

## Frontend Settings

These settings control how the server serves frontend assets and handles Cross-Origin Resource Sharing (CORS) for API requests from browsers.

| Option      | Environment Variable | Default Value | Description                                                              |
| ----------- | -------------------- | ------------- | ------------------------------------------------------------------------ |
| StaticURL   | `STATIC_URL`         | *(none)*      | URL to serve static frontend assets from (enables reverse proxy mode)    |
| StaticDir   | `STATIC_DIR`         | `./fe`        | Directory containing frontend static files (used when not in proxy mode) |
| CorsOrigins | `CORS_ORIGINS`       | `*`           | Allowed origins for CORS requests (comma-separated)                      |

### Usage Details

The frontend settings are extensively used in `pkg/server/router.go` for configuring how the application serves the frontend:

- **StaticURL**: When set, enables reverse proxy mode where static assets are served from an external URL:
  ```go
  if cfg.StaticURL != nil && cfg.StaticURL.Scheme != "" && cfg.StaticURL.Host != "" {
      // Set up reverse proxy for static assets
      router.NoRoute(func(c *gin.Context) {
          req := c.Request.Clone(c.Request.Context())
          req.URL.Scheme = cfg.StaticURL.Scheme
          req.URL.Host = cfg.StaticURL.Host
          // ...
      })
  }
  ```

- **StaticDir**: When StaticURL is not set, specifies the local directory containing static frontend assets:
  ```go
  // Serve static files from local directory
  router.Use(static.Serve("/", static.LocalFile(cfg.StaticDir, true)))

  // Also used for finding index.html for SPA routes
  indexPath := filepath.Join(cfg.StaticDir, "index.html")
  ```

- **CorsOrigins**: Configures CORS policy for the API, controlling which origins can make requests:
  ```go
  // In GraphQL service initialization
  graphqlService := services.NewGraphqlService(db, baseURL, cfg.CorsOrigins, providers, controller, subscriptions)

  // In CORS middleware configuration
  if !slices.Contains(cfg.CorsOrigins, "*") {
      config.AllowCredentials = true
  }
  config.AllowOrigins = cfg.CorsOrigins
  ```

These settings are essential for:
- Supporting different deployment architectures (single server vs. separate frontend/backend)
- Enabling proper SPA routing for frontend applications
- Configuring security policies for cross-origin requests

## Authentication Settings

These settings control authentication mechanisms, including cookie-based sessions and OAuth providers for user login.

| Option                  | Environment Variable         | Default Value | Description                                            |
| ----------------------- | ---------------------------- | ------------- | ------------------------------------------------------ |
| CookieSigningSalt       | `COOKIE_SIGNING_SALT`        | *(none)*      | Salt for signing and securing cookies used in sessions |
| PublicURL               | `PUBLIC_URL`                 | *(none)*      | Public origin/base URL used to build OAuth callback URLs such as `/api/v1/auth/login-callback` |
| OAuthGoogleClientID     | `OAUTH_GOOGLE_CLIENT_ID`     | *(none)*      | Google OAuth client ID for authentication              |
| OAuthGoogleClientSecret | `OAUTH_GOOGLE_CLIENT_SECRET` | *(none)*      | Google OAuth client secret                             |
| OAuthGithubClientID     | `OAUTH_GITHUB_CLIENT_ID`     | *(none)*      | GitHub OAuth client ID for authentication              |
| OAuthGithubClientSecret | `OAUTH_GITHUB_CLIENT_SECRET` | *(none)*      | GitHub OAuth client secret                             |

### Usage Details

The authentication settings are used in `pkg/server/router.go` to set up authentication middleware and OAuth providers:

- **CookieSigningSalt**: Used to secure cookies for session management:
  ```go
  // Used in auth middleware for authentication checks
  authMiddleware := auth.NewAuthMiddleware(baseURL, cfg.CookieSigningSalt, tokenCache, userCache)

  // Used for cookie store creation
  cookieStore := cookie.NewStore(auth.MakeCookieStoreKey(cfg.CookieSigningSalt)...)
  router.Use(sessions.Sessions("auth", cookieStore))
  ```

- **PublicURL**: The public origin/base URL for OAuth callback endpoints, crucial for redirects after authentication:
  ```go
  publicURL, err := url.Parse(cfg.PublicURL)
  ```

  The router builds the login callback path under the API base URL:
  ```go
  oauthLoginCallbackURL := "/auth/login-callback"

  publicURL, err := url.Parse(cfg.PublicURL)
  if err == nil {
      publicURL.Path = path.Join(baseURL, oauthLoginCallbackURL)
  }
  ```

  In the default deployment, configure your OAuth providers with:
  - **Homepage URL**: `PUBLIC_URL`
  - **Authorization callback URL / Redirect URI**: `${PUBLIC_URL}/api/v1/auth/login-callback`

  Example:
  ```bash
  PUBLIC_URL=https://pentagi.example.com
  OAUTH_GITHUB_CLIENT_ID=your_github_client_id
  OAUTH_GITHUB_CLIENT_SECRET=your_github_client_secret
  OAUTH_GOOGLE_CLIENT_ID=your_google_client_id
  OAUTH_GOOGLE_CLIENT_SECRET=your_google_client_secret
  ```

- **OAuth Provider Settings**: Used to configure authentication with Google and GitHub:
  ```go
  // Google OAuth setup
  if publicURL != nil && cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
      googleClient := oauth.NewGoogleOAuthClient(
          cfg.OAuthGoogleClientID,
          cfg.OAuthGoogleClientSecret,
          publicURL.String(),
      )
      // ...
  }

  // GitHub OAuth setup
  if publicURL != nil && cfg.OAuthGithubClientID != "" && cfg.OAuthGithubClientSecret != "" {
      githubClient := oauth.NewGithubOAuthClient(
          cfg.OAuthGithubClientID,
          cfg.OAuthGithubClientSecret,
          publicURL.String(),
      )
      // ...
  }
  ```

  Google and GitHub both use the same PentAGI login callback endpoint. `PUBLIC_URL` should be the externally reachable base URL only, without an extra path suffix. If the URL configured in the provider console does not exactly match the generated callback URL, authentication will fail with a redirect URI mismatch error.

These settings are essential for:
- Secure user authentication and session management
- Supporting social login through OAuth providers
- Enabling proper redirects in the authentication flow

## Web Scraper Settings

These settings control the web scraper service used for browsing websites and taking screenshots, which allows AI agents to interact with web content.

| Option            | Environment Variable  | Default Value | Description                                               |
| ----------------- | --------------------- | ------------- | --------------------------------------------------------- |
| ScraperPublicURL  | `SCRAPER_PUBLIC_URL`  | *(none)*      | Public URL for accessing the scraper service from clients |
| ScraperPrivateURL | `SCRAPER_PRIVATE_URL` | *(none)*      | Private URL for internal scraper service access           |

### Usage Details

The scraper settings are extensively used in the tools executor to provide web browsing capabilities to AI agents:

```go
// In various tool functions in pkg/tools/tools.go
browseTool = &functions.BrowseFunc{
    scPrvURL: fte.cfg.ScraperPrivateURL,
    scPubURL: fte.cfg.ScraperPublicURL,
    // ...
}

screenshotTool = &functions.ScreenshotFunc{
    scPrvURL: fte.cfg.ScraperPrivateURL,
    scPubURL: fte.cfg.ScraperPublicURL,
    // ...
}
```

These URLs serve different purposes:
- **ScraperPublicURL**: Used when generating URLs that will be accessed by the client (browser)
- **ScraperPrivateURL**: Used for internal communication between the backend and the scraper service

The scraper settings enable critical functionality:
- Web browsing capabilities for AI agents
- Screenshot capturing for web content analysis
- Web information gathering for research tasks

## LLM Provider Settings

These settings control the integration with various Large Language Model (LLM) providers, including OpenAI, Anthropic, and custom providers.

### OpenAI

| Option          | Environment Variable | Default Value               | Description                        |
| --------------- | -------------------- | --------------------------- | ---------------------------------- |
| OpenAIKey       | `OPEN_AI_KEY`        | *(none)*                    | API key for OpenAI services        |
| OpenAIServerURL | `OPEN_AI_SERVER_URL` | `https://api.openai.com/v1` | Server URL for OpenAI API requests |

### Anthropic

| Option             | Environment Variable   | Default Value                  | Description                           |
| ------------------ | ---------------------- | ------------------------------ | ------------------------------------- |
| AnthropicAPIKey    | `ANTHROPIC_API_KEY`    | *(none)*                       | API key for Anthropic Claude services |
| AnthropicServerURL | `ANTHROPIC_SERVER_URL` | `https://api.anthropic.com/v1` | Server URL for Anthropic API requests |

**Note on Google Vertex AI**: PentAGI does not currently expose a dedicated Vertex AI configuration path for Anthropic Claude in `.env`. The variables above target the direct Anthropic API. To run Claude through a non-Anthropic-hosted backend, use one of:

- **AWS Bedrock**: see the [AWS Bedrock LLM Provider](#aws-bedrock-llm-provider) section below and configure the `BEDROCK_*` variables.
- **OpenAI-compatible gateway in front of Vertex AI**: expose Vertex AI through a proxy or gateway that translates requests into the Chat Completions format while preserving the chat and tool-call behavior PentAGI requires, then configure it as a [custom LLM provider](#custom-llm-provider) (`LLM_SERVER_URL`, `LLM_SERVER_KEY`, `LLM_SERVER_MODEL`). Reliability of this path depends on the gateway you choose.

There is no `VERTEX_API_KEY` or `GOOGLE_APPLICATION_CREDENTIALS` variable wired into PentAGI's provider initialization today.

### Ollama LLM Provider

| Option                        | Environment Variable                | Default Value        | Description                                                       |
| ----------------------------- | ----------------------------------- | -------------------- | ----------------------------------------------------------------- |
| OllamaServerURL               | `OLLAMA_SERVER_URL`                 | *(none)*             | Ollama server URL (local or cloud https://ollama.com)             |
| OllamaServerAPIKey            | `OLLAMA_SERVER_API_KEY`             | *(none)*             | Ollama Cloud API key (optional, required for https://ollama.com)  |
| OllamaServerModel             | `OLLAMA_SERVER_MODEL`               | *(none)*             | Default model to use for inference                                |
| OllamaServerConfig            | `OLLAMA_SERVER_CONFIG_PATH`         | *(none)*             | Path to config file for Ollama provider options                   |
| OllamaServerPullModelsTimeout | `OLLAMA_SERVER_PULL_MODELS_TIMEOUT` | `600`                | Timeout in seconds for model downloads                            |
| OllamaServerPullModelsEnabled | `OLLAMA_SERVER_PULL_MODELS_ENABLED` | `false`              | Automatically download required models on startup                 |
| OllamaServerLoadModelsEnabled | `OLLAMA_SERVER_LOAD_MODELS_ENABLED` | `false`              | Load available models list from server API                        |

**Deployment Scenarios**: 
- **Local Server**: Set `OLLAMA_SERVER_URL` to local endpoint (e.g., `http://ollama-server:11434`), leave `OLLAMA_SERVER_API_KEY` empty
- **Ollama Cloud**: Set `OLLAMA_SERVER_URL=https://ollama.com` and provide `OLLAMA_SERVER_API_KEY` from https://ollama.com/settings/keys

**Note:** When `OllamaServerLoadModelsEnabled=false`, only the default model is available. Enable this to see all installed models in the UI.

### Google AI (Gemini) LLM Provider

| Option          | Environment Variable | Default Value                               | Description                           |
| --------------- | -------------------- | ------------------------------------------- | ------------------------------------- |
| GeminiAPIKey    | `GEMINI_API_KEY`     | *(none)*                                    | API key for Google AI Gemini services |
| GeminiServerURL | `GEMINI_SERVER_URL`  | `https://generativelanguage.googleapis.com` | Server URL for Gemini API requests    |

### AWS Bedrock LLM Provider

| Option              | Environment Variable        | Default Value | Description                                                                                                              |
| ------------------- | --------------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------ |
| BedrockRegion       | `BEDROCK_REGION`            | `us-east-1`   | AWS region for Bedrock service                                                                                           |
| BedrockDefaultAuth  | `BEDROCK_DEFAULT_AUTH`      | `false`       | Use default AWS SDK credential chain (environment variables, EC2 role, ~/.aws/credentials) - highest priority            |
| BedrockBearerToken  | `BEDROCK_BEARER_TOKEN`      | *(none)*      | Bearer token for authentication - takes priority over static credentials                                                 |
| BedrockAccessKey    | `BEDROCK_ACCESS_KEY_ID`     | *(none)*      | AWS access key ID for static credentials authentication                                                                  |
| BedrockSecretKey    | `BEDROCK_SECRET_ACCESS_KEY` | *(none)*      | AWS secret access key for static credentials authentication                                                              |
| BedrockSessionToken | `BEDROCK_SESSION_TOKEN`     | *(none)*      | AWS session token for temporary credentials (optional, used with static credentials for STS/assumed roles)               |
| BedrockServerURL    | `BEDROCK_SERVER_URL`        | *(none)*      | Optional custom endpoint URL for Bedrock service (VPC endpoints, local testing)                                          |
| BedrockConfig       | `BEDROCK_CONFIG_PATH`       | *(none)*      | Path to a custom YAML config that replaces the built-in Bedrock per-agent config (model assignments, prices)             |

**Authentication Priority**: `BedrockDefaultAuth` (highest) → `BedrockBearerToken` → `BedrockAccessKey`+`BedrockSecretKey` (lowest)

### DeepSeek LLM Provider

| Option            | Environment Variable  | Default Value              | Description                                              |
| ----------------- | --------------------- | -------------------------- | -------------------------------------------------------- |
| DeepSeekAPIKey    | `DEEPSEEK_API_KEY`    | *(none)*                   | DeepSeek API key for authentication                      |
| DeepSeekServerURL | `DEEPSEEK_SERVER_URL` | `https://api.deepseek.com` | DeepSeek API endpoint URL                                |
| DeepSeekProvider  | `DEEPSEEK_PROVIDER`   | *(none)*                   | Provider name prefix for LiteLLM integration (optional)  |

**LiteLLM Integration**: Set `DEEPSEEK_PROVIDER=deepseek` to enable model prefixing (e.g., `deepseek/deepseek-v4-flash`) when using LiteLLM proxy with default PentAGI configs.

### GLM LLM Provider

| Option         | Environment Variable | Default Value                  | Description                                              |
| -------------- | -------------------- | ------------------------------ | -------------------------------------------------------- |
| GLMAPIKey      | `GLM_API_KEY`        | *(none)*                       | GLM API key for authentication                           |
| GLMServerURL   | `GLM_SERVER_URL`     | `https://api.z.ai/api/paas/v4` | GLM API endpoint URL (international)                     |
| GLMProvider    | `GLM_PROVIDER`       | *(none)*                       | Provider name prefix for LiteLLM integration (optional)  |

**Alternative Endpoints**:
- International: `https://api.z.ai/api/paas/v4` (default)
- China: `https://open.bigmodel.cn/api/paas/v4`
- Coding-specific: `https://api.z.ai/api/coding/paas/v4`

**LiteLLM Integration**: Set `GLM_PROVIDER=zai` to enable model prefixing (e.g., `zai/glm-4`) when using LiteLLM proxy with default PentAGI configs.

### Kimi LLM Provider

| Option          | Environment Variable | Default Value                 | Description                                              |
| --------------- | -------------------- | ----------------------------- | -------------------------------------------------------- |
| KimiAPIKey      | `KIMI_API_KEY`       | *(none)*                      | Kimi API key for authentication                          |
| KimiServerURL   | `KIMI_SERVER_URL`    | `https://api.moonshot.ai/v1`  | Kimi API endpoint URL (international)                    |
| KimiProvider    | `KIMI_PROVIDER`      | *(none)*                      | Provider name prefix for LiteLLM integration (optional)  |

**Alternative Endpoints**:
- International: `https://api.moonshot.ai/v1` (default)
- China: `https://api.moonshot.cn/v1`

**LiteLLM Integration**: Set `KIMI_PROVIDER=moonshot` to enable model prefixing (e.g., `moonshot/kimi-k2.5`) when using LiteLLM proxy with default PentAGI configs.

### Qwen LLM Provider

| Option          | Environment Variable | Default Value                                          | Description                                              |
| --------------- | -------------------- | ------------------------------------------------------ | -------------------------------------------------------- |
| QwenAPIKey      | `QWEN_API_KEY`       | *(none)*                                               | Qwen API key for authentication                          |
| QwenServerURL   | `QWEN_SERVER_URL`    | `https://dashscope-us.aliyuncs.com/compatible-mode/v1` | Qwen API endpoint URL (international)                    |
| QwenProvider    | `QWEN_PROVIDER`      | *(none)*                                               | Provider name prefix for LiteLLM integration (optional)  |

**Alternative Endpoints**:
- US: `https://dashscope-us.aliyuncs.com/compatible-mode/v1` (default)
- Singapore: `https://dashscope-intl.aliyuncs.com/compatible-mode/v1`
- China: `https://dashscope.aliyuncs.com/compatible-mode/v1`

**LiteLLM Integration**: Set `QWEN_PROVIDER=dashscope` to enable model prefixing (e.g., `dashscope/qwen-plus`) when using LiteLLM proxy with default PentAGI configs.

### MiniMax LLM Provider

| Option           | Environment Variable | Default Value               | Description                                             |
| ---------------- | -------------------- | --------------------------- | ------------------------------------------------------- |
| MiniMaxAPIKey    | `MINIMAX_API_KEY`    | *(none)*                    | MiniMax API key for authentication                      |
| MiniMaxServerURL | `MINIMAX_SERVER_URL` | `https://api.minimax.io/v1` | MiniMax API endpoint URL                                |
| MiniMaxProvider  | `MINIMAX_PROVIDER`   | *(none)*                    | Provider name prefix for LiteLLM integration (optional) |

**LiteLLM Integration**: Set `MINIMAX_PROVIDER=minimax` to enable model prefixing (e.g., `minimax/MiniMax-M3`) when using LiteLLM proxy with default PentAGI configs.

### Custom LLM Provider

| Option                     | Environment Variable            | Default Value | Description                                                                  |
| -------------------------- | ------------------------------- | ------------- | ---------------------------------------------------------------------------- |
| LLMServerURL               | `LLM_SERVER_URL`                | *(none)*      | Server URL for custom LLM provider                                           |
| LLMServerKey               | `LLM_SERVER_KEY`                | *(none)*      | API key for custom LLM provider                                              |
| LLMServerModel             | `LLM_SERVER_MODEL`              | *(none)*      | Model name for custom LLM provider                                           |
| LLMServerConfig            | `LLM_SERVER_CONFIG_PATH`        | *(none)*      | Path to config file for custom LLM provider options                          |
| LLMServerProvider          | `LLM_SERVER_PROVIDER`           | *(none)*      | Provider name prefix for model names (useful for LiteLLM proxy)              |
| LLMServerLegacyReasoning   | `LLM_SERVER_LEGACY_REASONING`   | `false`       | Controls reasoning format in API requests                                    |
| LLMServerPreserveReasoning | `LLM_SERVER_PRESERVE_REASONING` | `false`       | Preserve reasoning content in multi-turn conversations (required by some providers) |

### Usage Details

The LLM provider settings are used in `pkg/providers` modules to initialize and configure the appropriate language model providers:

- **OpenAI Settings**: Used in `pkg/providers/openai/openai.go` to create the OpenAI client:
  ```go
  baseURL := cfg.OpenAIServerURL

  client, err := openai.New(
      openai.WithToken(cfg.OpenAIKey),
      openai.WithModel(OpenAIAgentModel),
      openai.WithBaseURL(baseURL),
      // ...
  )
  ```

- **Anthropic Settings**: Used in `pkg/providers/anthropic/anthropic.go` to create the Anthropic client:
  ```go
  baseURL := cfg.AnthropicServerURL

  client, err := anthropic.New(
      anthropic.WithToken(cfg.AnthropicAPIKey),
      anthropic.WithBaseURL(baseURL),
      // ...
  )
  ```

- **Ollama Settings**: Used in `pkg/providers/ollama/ollama.go` to create the Ollama client:
  ```go
  serverURL := cfg.OllamaServerURL

  client, err := ollama.New(
      ollama.WithServerURL(serverURL),
      ollama.WithHTTPClient(httpClient),
      ollama.WithModel(OllamaAgentModel),
      ollama.WithPullModel(),
  )

  // Load provider options from config file if specified
  if cfg.OllamaServerConfig != "" {
      configData, err := os.ReadFile(cfg.OllamaServerConfig)
      providerConfig, err := BuildProviderConfig(cfg, configData)
      // ...
  }
  ```

- **Gemini Settings**: Used in `pkg/providers/gemini/gemini.go` to create the Google AI client:
  ```go
  opts := []googleai.Option{
      googleai.WithRest(),
      googleai.WithAPIKey(cfg.GeminiAPIKey),
      googleai.WithEndpoint(cfg.GeminiServerURL),
      googleai.WithDefaultModel(GeminiAgentModel),
  }

  client, err := googleai.New(context.Background(), opts...)
  ```

- **Bedrock Settings**: Used in `pkg/providers/bedrock/bedrock.go` to create the AWS Bedrock client:
  ```go
  opts := []func(*bconfig.LoadOptions) error{
      bconfig.WithRegion(cfg.BedrockRegion),
      bconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
          cfg.BedrockAccessKey,
          cfg.BedrockSecretKey,
          cfg.BedrockSessionToken,
      )),
  }

  if cfg.BedrockServerURL != "" {
      opts = append(opts, bconfig.WithBaseEndpoint(cfg.BedrockServerURL))
  }

  bcfg, err := bconfig.LoadDefaultConfig(context.Background(), opts...)
  bclient := bedrockruntime.NewFromConfig(bcfg)

  client, err := bedrock.New(
      bedrock.WithClient(bclient),
      bedrock.WithModel(BedrockAgentModel),
      bedrock.WithConverseAPI(),
  )
  ```

  The `BedrockSessionToken` is optional and only required when using temporary AWS credentials (e.g., from STS, assumed roles, or MFA-enabled IAM users). For permanent IAM user credentials, leave this field empty.

- **Custom LLM Settings**: Used in `pkg/providers/custom/custom.go` to create a custom LLM client:
  ```go
  baseKey := cfg.LLMServerKey
  baseURL := cfg.LLMServerURL
  baseModel := cfg.LLMServerModel

  client, err := openai.New(
      openai.WithToken(baseKey),
      openai.WithModel(baseModel),
      openai.WithBaseURL(baseURL),
      // ...
  )

  // Load provider options from config file if specified
  if cfg.LLMServerConfig != "" {
      providerConfig, err := LoadConfig(cfg.LLMServerConfig, simple)
      // ...
  }
  ```

- **LLMServerLegacyReasoning**: Controls the reasoning format used in API requests to custom LLM providers:
  ```go
  // Used in custom provider to determine reasoning format
  if cfg.LLMServerLegacyReasoning {
      // Uses legacy string-based reasoning_effort parameter
  } else {
      // Uses modern structured reasoning object with max_tokens
  }
  ```
  - `false` (default): Uses modern format where reasoning is sent as a structured object with `max_tokens` parameter
  - `true`: Uses legacy format with string-based `reasoning_effort` parameter

This setting is important when working with different LLM providers as they may expect different reasoning formats in their API requests. If you encounter reasoning-related errors with custom providers, try changing this setting.

- **LLMServerPreserveReasoning**: Controls whether reasoning content is preserved and sent back in multi-turn conversations:
  ```go
  // Used in custom provider to preserve reasoning content
  if cfg.LLMServerPreserveReasoning {
      // Preserves and returns reasoning_content in assistant messages
  }
  ```
  - `false` (default): Reasoning content is not preserved in conversation history
  - `true`: Reasoning content is preserved and sent in subsequent API calls

This setting is required by some LLM providers (e.g., Moonshot) that return errors like "thinking is enabled but reasoning_content is missing in assistant tool call message" when reasoning content is not included in multi-turn conversations. Enable this setting if your provider requires reasoning content to be preserved across conversation turns.

The provider registration is managed in `pkg/providers/providers.go`:

```go
// Provider registration based on available credentials
if cfg.OpenAIKey != "" {
    p, err := openai.New(cfg, defaultConfigs[provider.ProviderOpenAI])
    if err != nil {
        return nil, fmt.Errorf("failed to create openai provider: %w", err)
    }
    providers[provider.DefaultProviderNameOpenAI] = p
}

if cfg.AnthropicAPIKey != "" {
    p, err := anthropic.New(cfg, defaultConfigs[provider.ProviderAnthropic])
    if err != nil {
        return nil, fmt.Errorf("failed to create anthropic provider: %w", err)
    }
    providers[provider.DefaultProviderNameAnthropic] = p
}

if cfg.GeminiAPIKey != "" {
    p, err := gemini.New(cfg, defaultConfigs[provider.ProviderGemini])
    if err != nil {
        return nil, fmt.Errorf("failed to create gemini provider: %w", err)
    }
    providers[provider.DefaultProviderNameGemini] = p
}

if cfg.BedrockAccessKey != "" && cfg.BedrockSecretKey != "" {
    p, err := bedrock.New(cfg, defaultConfigs[provider.ProviderBedrock])
    if err != nil {
        return nil, fmt.Errorf("failed to create bedrock provider: %w", err)
    }
    providers[provider.DefaultProviderNameBedrock] = p
}

if cfg.OllamaServerURL != "" {
    p, err := ollama.New(cfg, defaultConfigs[provider.ProviderOllama])
    if err != nil {
        return nil, fmt.Errorf("failed to create ollama provider: %w", err)
    }
    providers[provider.DefaultProviderNameOllama] = p
}

if cfg.LLMServerURL != "" && (cfg.LLMServerModel != "" || cfg.LLMServerConfig != "") {
    p, err := custom.New(cfg, defaultConfigs[provider.ProviderCustom])
    if err != nil {
        return nil, fmt.Errorf("failed to create custom provider: %w", err)
    }
    providers[provider.DefaultProviderNameCustom] = p
}
```

These settings are critical for:
- Connecting to various LLM providers for AI capabilities
- Supporting multiple model options for different tasks
- Enabling custom or self-hosted LLM solutions
- Configuring specific model behaviors and parameters

## Embedding Settings

These settings control the vector embedding service used for semantic search and similarity matching, which is fundamental for PentAGI's intelligent search capabilities.

| Option                 | Environment Variable        | Default Value | Description                                                                |
| ---------------------- | --------------------------- | ------------- | -------------------------------------------------------------------------- |
| EmbeddingURL           | `EMBEDDING_URL`             | *(none)*      | Server URL for embedding provider (overrides provider-specific URLs)       |
| EmbeddingKey           | `EMBEDDING_KEY`             | *(none)*      | API key for embedding provider (overrides provider-specific keys)          |
| EmbeddingModel         | `EMBEDDING_MODEL`           | *(none)*      | Model name for embedding generation                                        |
| EmbeddingStripNewLines | `EMBEDDING_STRIP_NEW_LINES` | `true`        | Whether to strip newlines before embedding (improves quality)              |
| EmbeddingBatchSize     | `EMBEDDING_BATCH_SIZE`      | `512`         | Batch size for embedding operations (affects memory usage and performance) |
| EmbeddingProvider      | `EMBEDDING_PROVIDER`        | `openai`      | Provider for embeddings (openai, ollama, mistral, jina, huggingface)       |
| EmbeddingMaxTextBytes  | `EMBEDDING_MAX_TEXT_BYTES`  | `8192`        | Maximum byte size of text sent to the embedding model per document. Acts as a byte-level proxy for token limits (e.g. 8192 tokens for OpenAI models). When a stored document exceeds this limit the heavy content field (Guide/Answer/Code) is truncated to fit before computing the vector; the full original text is always preserved in the database. Reduce if your model has a smaller context window. |

### Usage Details

The embedding settings are extensively used in `pkg/providers/embeddings/embedder.go` to configure the vector embedding service:

- **EmbeddingProvider**: Determines which embedding provider to use:
  ```go
  switch cfg.EmbeddingProvider {
  case "openai":
      return newOpenAIEmbedder(ctx, cfg)
  case "ollama":
      return newOllamaEmbedder(ctx, cfg)
  case "mistral":
      return newMistralEmbedder(ctx, cfg)
  case "jina":
      return newJinaEmbedder(ctx, cfg)
  case "huggingface":
      return newHuggingFaceEmbedder(ctx, cfg)
  default:
      return &embedder{nil}, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
  }
  ```

- **Provider-specific configurations**: Used to configure each embedding provider with appropriate options:
  ```go
  // Example for OpenAI embeddings
  if cfg.EmbeddingURL != "" {
      opts = append(opts, openai.WithBaseURL(cfg.EmbeddingURL))
  } else if cfg.OpenAIServerURL != "" {
      opts = append(opts, openai.WithBaseURL(cfg.OpenAIServerURL))
  }

  if cfg.EmbeddingKey != "" {
      opts = append(opts, openai.WithToken(cfg.EmbeddingKey))
  } else if cfg.OpenAIKey != "" {
      opts = append(opts, openai.WithToken(cfg.OpenAIKey))
  }

  if cfg.EmbeddingModel != "" {
      opts = append(opts, openai.WithEmbeddingModel(cfg.EmbeddingModel))
  }
  ```

- **Embedding behavior configuration**: Controls how text is processed for embeddings:
  ```go
  embeddings.WithStripNewLines(cfg.EmbeddingStripNewLines),
  embeddings.WithBatchSize(cfg.EmbeddingBatchSize),
  ```

These settings are essential for:
- Configuring semantic search capabilities
- Determining which embedding model to use
- Optimizing embedding performance and quality
- Supporting multiple embedding providers for flexibility

## Summarizer Settings

These settings control the text summarization behavior used for condensing long conversations and improving context management in AI interactions. The summarization system is a critical component that allows PentAGI to maintain coherent, long-running conversations while managing token usage effectively.

| Option                   | Environment Variable             | Default Value | Description                                                |
| ------------------------ | -------------------------------- | ------------- | ---------------------------------------------------------- |
| SummarizerPreserveLast   | `SUMMARIZER_PRESERVE_LAST`       | `true`        | Preserve the last message in summarization                 |
| SummarizerUseQA          | `SUMMARIZER_USE_QA`              | `true`        | Use question-answer format for summarization               |
| SummarizerSumHumanInQA   | `SUMMARIZER_SUM_MSG_HUMAN_IN_QA` | `false`       | Include human messages in QA summaries                     |
| SummarizerLastSecBytes   | `SUMMARIZER_LAST_SEC_BYTES`      | `51200`       | Bytes to preserve from the last section (50KB)             |
| SummarizerMaxBPBytes     | `SUMMARIZER_MAX_BP_BYTES`        | `16384`       | Maximum bytes for bullet points summarization (16KB)       |
| SummarizerMaxQASections  | `SUMMARIZER_MAX_QA_SECTIONS`     | `10`          | Maximum QA sections to include                             |
| SummarizerMaxQABytes     | `SUMMARIZER_MAX_QA_BYTES`        | `65536`       | Maximum bytes for QA summarization (64KB)                  |
| SummarizerKeepQASections | `SUMMARIZER_KEEP_QA_SECTIONS`    | `1`           | Number of recent QA sections to keep without summarization |

### Usage Details and Impact on System Behavior

The summarizer settings map directly to the `SummarizerConfig` structure that controls the chain summarization algorithm in `pkg/csum`. These settings work together to implement a sophisticated, multi-strategy approach to managing conversation context:

#### Core Summarization Strategies and Their Parameters

1. **Section Summarization** - Always active, ensures all older sections (except the last one) consist of a single summarized pair
   - No specific parameters control this as it's a fundamental part of the algorithm
   - Prevents unbounded growth by consolidating completed conversation sections

2. **Last Section Management** (`SummarizerPreserveLast` and `SummarizerLastSecBytes`)
   - Controls how the current/active conversation section is managed
   - When `SummarizerPreserveLast = true`, older messages within the last section will be summarized when the section exceeds `SummarizerLastSecBytes` bytes
   - A reserve space of 25% is automatically maintained to accommodate new messages without triggering frequent re-summarization
   - Individual oversized pairs are summarized separately if they exceed `SummarizerMaxBPBytes`

3. **QA Pair Summarization** (`SummarizerUseQA`, `SummarizerMaxQASections`, `SummarizerMaxQABytes`, `SummarizerSumHumanInQA`)
   - When `SummarizerUseQA = true`, creates larger summarization units focused on question-answer patterns
   - Preserves the most recent `SummarizerMaxQASections` sections as long as they don't exceed `SummarizerMaxQABytes` total
   - If `SummarizerSumHumanInQA = true`, human messages are also summarized; otherwise, they're preserved verbatim

#### Deep Dive: Parameter Impact and Recommendations

**`SummarizerPreserveLast`** (Default: `true`)
- **Purpose**: Controls whether the last (active) section has size management applied
- **Impact**: When enabled, prevents the active conversation from growing indefinitely
- **When to adjust**:
  - Enable (default) for production systems and long-running conversations
  - Disable only for debugging or when you need to preserve the complete conversation history regardless of size

**`SummarizerLastSecBytes`** (Default: `51200` - 50KB)
- **Purpose**: Maximum byte size for the last (active) section before summarization begins
- **Impact**: Directly controls how much conversation history is preserved verbatim in the active section
- **When to adjust**:
  - Increase for models with larger context windows to maintain more conversation detail
  - Decrease for models with smaller context to prevent token limits from being exceeded
  - Balance with `SummarizerMaxBPBytes` to ensure coherent summarization

**`SummarizerMaxBPBytes`** (Default: `16384` - 16KB)
- **Purpose**: Maximum byte size for individual body pairs (typically AI responses)
- **Impact**: Controls when individual large responses get summarized, even if the overall section is under limit
- **When to adjust**:
  - Increase if your use case involves long but important AI responses that shouldn't be summarized
  - Decrease if you want more aggressive summarization of lengthy responses

**`SummarizerUseQA`** (Default: `true`)
- **Purpose**: Enables question-answer style summarization that creates more cohesive summaries
- **Impact**: When enabled, creates a new first section with a summary of older interactions, preserving recent sections
- **When to adjust**:
  - Enable (default) for more coherent, organized summaries focused on main topics
  - Disable if you prefer simpler, section-by-section summarization without cross-section analysis

**`SummarizerMaxQASections`** (Default: `10`)
- **Purpose**: Maximum number of recent sections to preserve when using QA-style summarization
- **Impact**: Directly controls how many conversation turns remain intact after QA summarization
- **When to adjust**:
  - Increase to preserve more recent conversation context (at the cost of token usage)
  - Decrease to create more compact conversation histories, focusing on only the very recent exchanges

**`SummarizerMaxQABytes`** (Default: `65536` - 64KB)
- **Purpose**: Maximum total byte size for preserved sections in QA-style summarization
- **Impact**: Sets an upper bound on memory used by preserved sections, regardless of section count
- **When to adjust**:
  - Increase for models with larger context windows or when detailed context is essential
  - Decrease for smaller context models or when prioritizing efficiency over context preservation

**`SummarizerSumHumanInQA`** (Default: `false`)
- **Purpose**: Controls whether human messages are summarized in QA-style summarization
- **Impact**: When false, human messages are preserved verbatim; when true, they are also summarized
- **When to adjust**:
  - Keep disabled (default) to preserve the exact wording of user queries
  - Enable only when human messages are very verbose and token efficiency is critical

**`SummarizerKeepQASections`** (Default: `1`)
- **Purpose**: Controls the number of recent QA sections to keep without summarization
- **Impact**: Directly controls how many recent conversation turns are preserved verbatim
- **When to adjust**:
  - Increase to preserve more recent conversation context
  - Decrease to create more compact conversation histories, focusing on only the very recent exchanges

### Summarization Effects on Agent Behavior

The summarization settings have significant effects on agent behavior:

1. **Context Retention vs. Token Efficiency**
   - More aggressive summarization (smaller byte limits) reduces token usage but may lose context details
   - More permissive settings (larger byte limits) preserve more context but increase token consumption

2. **Conversation Coherence**
   - Appropriate summarization helps the agent maintain a coherent understanding of the conversation
   - Over-aggressive summarization may cause the agent to lose important details or previous instructions
   - Under-aggressive summarization may lead to context overflow in longer conversations

3. **Response Quality**
   - QA-style summarization (`SummarizerUseQA = true`) typically improves response quality for complex tasks
   - Preserving human messages (`SummarizerSumHumanInQA = false`) helps maintain alignment with user intent
   - Appropriate `SummarizerMaxBPBytes` prevents loss of detailed information from complex AI responses

### Implementation Details

The summarizer settings are used in `pkg/providers/providers.go` to configure the summarization behavior:

```go
summarizer := provider.SummarizerSettings{
    PreserveLast:  cfg.SummarizerPreserveLast,
    UseQA:         cfg.SummarizerUseQA,
    SummHumanInQA: cfg.SummarizerSumHumanInQA,
    LastSecBytes:  cfg.SummarizerLastSecBytes,
    MaxBPBytes:    cfg.SummarizerMaxBPBytes,
    MaxQASections: cfg.SummarizerMaxQASections,
    MaxQABytes:    cfg.SummarizerMaxQABytes,
}
```

These settings are passed to various components through the chain summarization system:

```go
// In csum/chain_summary.go
func NewSummarizer(config SummarizerConfig) Summarizer {
    if config.PreserveLast {
        if config.LastSecBytes <= 0 {
            config.LastSecBytes = maxLastSectionByteSize
        }
    }

    if config.UseQA {
        if config.MaxQASections <= 0 {
            config.MaxQASections = maxQAPairSections
        }
        if config.MaxQABytes <= 0 {
            config.MaxQABytes = maxQAPairByteSize
        }
    }

    if config.MaxBPBytes <= 0 {
        config.MaxBPBytes = maxSingleBodyPairByteSize
    }

    return &summarizer{config: config}
}
```

### Recommended Settings for Different Use Cases

1. **Long-running Assistant Conversations**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 51200 (50KB)
   SummarizerMaxBPBytes: 16384 (16KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 10
   SummarizerMaxQABytes: 65536 (64KB)
   SummarizerSumHumanInQA: false
   SummarizerKeepQASections: 1
   ```
   The default settings are optimized for assistant-style conversations. They maintain a good balance between context retention and token efficiency.

2. **Technical Problem-Solving with Large Context Models**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 81920 (80KB)
   SummarizerMaxBPBytes: 32768 (32KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 15
   SummarizerMaxQABytes: 102400 (100KB)
   SummarizerSumHumanInQA: false
   SummarizerKeepQASections: 1
   ```
   Increased limits to preserve more technical details when using models with large context windows (e.g., GPT-4).

3. **Limited Context Models**
   ```
   SummarizerPreserveLast: true
   SummarizerLastSecBytes: 25600 (25KB)
   SummarizerMaxBPBytes: 8192 (8KB)
   SummarizerUseQA: true
   SummarizerMaxQASections: 5
   SummarizerMaxQABytes: 32768 (32KB)
   SummarizerSumHumanInQA: true
   SummarizerKeepQASections: 1
   ```
   More aggressive summarization for models with smaller context windows (e.g., smaller or older LLMs).

4. **Debugging or Analysis (Maximum Context Preservation)**
   ```
   SummarizerPreserveLast: false
   SummarizerUseQA: false
   SummarizerKeepQASections: 0
   ```
   Disables active summarization to preserve the complete conversation history for debugging purposes. Note that this can lead to context overflow in long conversations.

## Assistant Settings

These settings control the behavior of the AI assistant functionality, including whether to use multi-agent delegation and assistant-specific summarization settings.

| Option                            | Environment Variable                    | Default Value | Description                                                             |
| --------------------------------- | --------------------------------------- | ------------- | ----------------------------------------------------------------------- |
| AssistantUseAgents                | `ASSISTANT_USE_AGENTS`                  | `false`       | Controls the default value for agent usage when creating new assistants |
| AssistantSummarizerPreserveLast   | `ASSISTANT_SUMMARIZER_PRESERVE_LAST`    | `true`        | Whether to preserve all messages in the assistant's last section        |
| AssistantSummarizerLastSecBytes   | `ASSISTANT_SUMMARIZER_LAST_SEC_BYTES`   | `76800`       | Maximum byte size for assistant's last section (75KB)                   |
| AssistantSummarizerMaxBPBytes     | `ASSISTANT_SUMMARIZER_MAX_BP_BYTES`     | `16384`       | Maximum byte size for a single body pair in assistant context (16KB)    |
| AssistantSummarizerMaxQASections  | `ASSISTANT_SUMMARIZER_MAX_QA_SECTIONS`  | `7`           | Maximum QA sections to preserve in assistant context                    |
| AssistantSummarizerMaxQABytes     | `ASSISTANT_SUMMARIZER_MAX_QA_BYTES`     | `76800`       | Maximum byte size for assistant's QA sections (75KB)                    |
| AssistantSummarizerKeepQASections | `ASSISTANT_SUMMARIZER_KEEP_QA_SECTIONS` | `3`           | Number of recent QA sections to preserve without summarization          |

### Usage Details

The assistant settings are used to configure the behavior of the AI assistant and its context management:

- **AssistantUseAgents**: Controls the default state of the "Use Agents" toggle when creating new assistants in the UI:
  ```go
  // This setting affects the initial state when creating assistants
  // Users can always override this by toggling the "Use Agents" button in the UI
  ```
  - `false` (default): New assistants are created with agent delegation disabled by default
  - `true`: New assistants are created with agent delegation enabled by default

- **Assistant Summarizer Settings**: These provide dedicated summarization configuration for assistant instances, typically allowing for more memory retention compared to the global settings:
  ```go
  // Assistant summarizer configuration provides more context retention
  // compared to global settings, preserving more recent conversation history
  // while still ensuring efficient token usage
  ```

The assistant summarizer configuration is designed to provide more memory for context retention compared to the global settings, preserving more recent conversation history while still ensuring efficient token usage.

### Recommended Assistant Settings for Different Use Cases

1. **Standard Assistant Conversations**
   ```
   AssistantUseAgents: false
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 76800 (75KB)
   AssistantSummarizerMaxBPBytes: 16384 (16KB)
   AssistantSummarizerMaxQASections: 7
   AssistantSummarizerMaxQABytes: 76800 (75KB)
   AssistantSummarizerKeepQASections: 3
   ```
   The default settings provide a balance between context retention and performance for typical assistant interactions.

2. **Multi-Agent Assistant Workflows**
   ```
   AssistantUseAgents: true
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 102400 (100KB)
   AssistantSummarizerMaxBPBytes: 32768 (32KB)
   AssistantSummarizerMaxQASections: 10
   AssistantSummarizerMaxQABytes: 102400 (100KB)
   AssistantSummarizerKeepQASections: 5
   ```
   Enhanced settings for complex workflows that benefit from agent delegation with increased context preservation.

3. **Resource-Constrained Assistant**
   ```
   AssistantUseAgents: false
   AssistantSummarizerPreserveLast: true
   AssistantSummarizerLastSecBytes: 51200 (50KB)
   AssistantSummarizerMaxBPBytes: 16384 (16KB)
   AssistantSummarizerMaxQASections: 5
   AssistantSummarizerMaxQABytes: 51200 (50KB)
   AssistantSummarizerKeepQASections: 2
   ```
   More conservative settings for environments with limited resources or smaller context models.

## Functions Configuration

These settings control which tools are available to AI agents and allow adding custom external functions. The Functions API enables fine-grained control over agent capabilities by selectively disabling built-in tools or extending functionality with custom integrations.

For provider-neutral OSINT enrichment ideas that could be exposed through external functions, see [OSINT Integration Scenarios for PentAGI Agents](../../examples/proposals/osint-integration-scenarios.md).

| Field     | Type                 | Description                                                     |
| --------- | -------------------- | --------------------------------------------------------------- |
| token     | string (optional)    | API token for authenticating external function calls            |
| disabled  | DisableFunction[]    | List of built-in functions to disable for specific agent types  |
| functions | ExternalFunction[]   | List of custom external functions to add to agent capabilities  |

### DisableFunction Structure

Allows disabling specific built-in functions for certain agent contexts, providing security and control over agent capabilities.

| Field   | Type     | Description                                                                    |
| ------- | -------- | ------------------------------------------------------------------------------ |
| name    | string   | Name of the built-in function to disable (e.g., `terminal`, `browser`, `file`) |
| context | string[] | Agent contexts where the function should be disabled (optional)                |

**Available Agent Contexts**: `agent`, `adviser`, `coder`, `searcher`, `generator`, `memorist`, `enricher`, `reporter`, `assistant`

When `context` is empty or omitted, the function is disabled for all agents.

### ExternalFunction Structure

Allows adding custom external functions that agents can call via HTTP endpoints, enabling integration with external tools and services.

| Field   | Type          | Description                                                        |
| ------- | ------------- | ------------------------------------------------------------------ |
| name    | string        | Name of the custom function (must be unique)                       |
| url     | string        | HTTP(S) URL endpoint for the function                              |
| timeout | int64         | Timeout in seconds for function execution (default: 60)            |
| context | string[]      | Agent contexts where the function is available (optional)          |
| schema  | Schema object | JSON schema defining function parameters and description (OpenAI format) |

**Available Agent Contexts**: Same as DisableFunction (`agent`, `adviser`, `coder`, `searcher`, `generator`, `memorist`, `enricher`, `reporter`, `assistant`)

When `context` is empty or omitted, the function is available to all agents.

### Usage Details

The Functions configuration is typically provided when creating a flow through the API:

```go
// Example from pkg/tools/tools.go
type Functions struct {
    Token    *string            `json:"token,omitempty"`
    Disabled []DisableFunction  `json:"disabled,omitempty"`
    Function []ExternalFunction `json:"functions,omitempty"`
}
```

These settings are used in `pkg/tools/tools.go` to configure available tools for each agent type:

- **Token**: Used for authenticating requests to external function endpoints:
  ```go
  // The token is passed in the Authorization header when calling external functions
  req.Header.Set("Authorization", "Bearer " + *functions.Token)
  ```

- **Disabled**: Filters out built-in functions for specific agent contexts:
  ```go
  // Check if function is disabled for current agent context
  for _, disabled := range functions.Disabled {
      if disabled.Name == functionName && 
         (len(disabled.Context) == 0 || contains(disabled.Context, agentType)) {
          // Skip this function
      }
  }
  ```

- **Functions**: Adds custom external functions to agent capabilities:
  ```go
  // Register external functions as available tools
  for _, externalFunc := range functions.Function {
      if len(externalFunc.Context) == 0 || contains(externalFunc.Context, agentType) {
          definitions = append(definitions, externalFunc.Schema)
          handlers[externalFunc.Name] = createExternalHandler(externalFunc)
      }
  }
  ```

### Example Configuration

```json
{
  "token": "secret-api-token-for-external-functions",
  "disabled": [
    {
      "name": "terminal",
      "context": ["searcher", "enricher"]
    },
    {
      "name": "browser",
      "context": ["memorist"]
    },
    {
      "name": "file"
    }
  ],
  "functions": [
    {
      "name": "custom_vulnerability_scan",
      "url": "https://scanner.example.com/api/v1/scan",
      "timeout": 120,
      "context": ["pentester", "coder"],
      "schema": {
        "type": "function",
        "function": {
          "name": "custom_vulnerability_scan",
          "description": "Perform a custom vulnerability scan on the target",
          "parameters": {
            "type": "object",
            "properties": {
              "target": {
                "type": "string",
                "description": "Target IP address or domain to scan"
              },
              "scan_type": {
                "type": "string",
                "enum": ["quick", "full", "stealth"],
                "description": "Type of scan to perform"
              }
            },
            "required": ["target"]
          }
        }
      }
    },
    {
      "name": "query_threat_intelligence",
      "url": "https://threatintel.example.com/api/query",
      "timeout": 30,
      "context": ["searcher", "adviser"],
      "schema": {
        "type": "function",
        "function": {
          "name": "query_threat_intelligence",
          "description": "Query threat intelligence database for IoCs and TTPs",
          "parameters": {
            "type": "object",
            "properties": {
              "indicator": {
                "type": "string",
                "description": "IP, domain, hash, or other indicator to search"
              },
              "indicator_type": {
                "type": "string",
                "enum": ["ip", "domain", "hash", "url"],
                "description": "Type of indicator"
              }
            },
            "required": ["indicator", "indicator_type"]
          }
        }
      }
    }
  ]
}
```

### Security Considerations

- **Token Security**: Store the `token` value securely and use HTTPS endpoints for external functions
- **Function Validation**: External functions should validate all inputs and return structured error messages
- **Timeout Configuration**: Set appropriate timeouts to prevent long-running operations from blocking agents
- **Context Restriction**: Use the `context` field to limit which agents can access sensitive functions
- **URL Validation**: Ensure external function URLs are trusted and properly secured

### Built-in Functions Reference

Common built-in functions that can be disabled:

- `terminal` - Execute shell commands in containers
- `file` - Read and write files in containers
- `browser` - Browse websites and take screenshots
- `search_in_memory` - Search vector memory store
- `search_guide` - Search knowledge guides
- `search_answer` - Search for answers
- `search_code` - Search code repositories
- `store_guide` - Store knowledge guides
- `store_answer` - Store answers
- `store_code` - Store code snippets
- `google` - Google Search
- `duckduckgo` - DuckDuckGo Search
- `tavily` - Tavily Search
- `firecrawl` - Firecrawl Search
- `traversaal` - Traversaal Search
- `perplexity` - Perplexity Search
- `searxng` - SearXNG Search
- `sploitus` - Sploitus Exploit Search
- `graphiti_search` - Graphiti Knowledge Graph Search

The specific functions available depend on the agent type and system configuration.

## Search Engine Settings

These settings control the integration with various search engines used for web search capabilities, providing AI agents with up-to-date information from the internet.

### DuckDuckGo Search

| Option               | Environment Variable    | Default Value | Description                                                                |
| -------------------- | ----------------------- | ------------- | -------------------------------------------------------------------------- |
| DuckDuckGoEnabled    | `DUCKDUCKGO_ENABLED`    | `true`        | Enable or disable DuckDuckGo Search engine                                 |
| DuckDuckGoRegion     | `DUCKDUCKGO_REGION`     | *(none)*      | Region code for search results (e.g., `us-en`, `uk-en`, `cn-zh`)           |
| DuckDuckGoSafeSearch | `DUCKDUCKGO_SAFESEARCH` | *(none)*      | Safe search filter (`off`, `moderate`, `strict`)                           |
| DuckDuckGoTimeRange  | `DUCKDUCKGO_TIME_RANGE` | *(none)*      | Time range for search results (`d`: day, `w`: week, `m`: month, `y`: year) |

### Sploitus Search

| Option          | Environment Variable | Default Value | Description                                                 |
| --------------- | -------------------- | ------------- | ----------------------------------------------------------- |
| SploitusEnabled | `SPLOITUS_ENABLED`   | `true`        | Enable or disable Sploitus exploit and vulnerability search |

### Google Search

| Option       | Environment Variable | Default Value | Description                                              |
| ------------ | -------------------- | ------------- | -------------------------------------------------------- |
| GoogleAPIKey | `GOOGLE_API_KEY`     | *(none)*      | API key for Google Search                                |
| GoogleCXKey  | `GOOGLE_CX_KEY`      | *(none)*      | Custom Search Engine ID for Google Search                |
| GoogleLRKey  | `GOOGLE_LR_KEY`      | `lang_en`     | Language restriction for Google Search (e.g., `lang_en`) |

### Traversaal Search

| Option           | Environment Variable | Default Value | Description                          |
| ---------------- | -------------------- | ------------- | ------------------------------------ |
| TraversaalAPIKey | `TRAVERSAAL_API_KEY` | *(none)*      | API key for Traversaal search engine |

### Tavily Search

| Option       | Environment Variable | Default Value | Description                      |
| ------------ | -------------------- | ------------- | -------------------------------- |
| TavilyAPIKey | `TAVILY_API_KEY`     | *(none)*      | API key for Tavily search engine |

### Firecrawl Search

| Option          | Environment Variable | Default Value               | Description                                                                    |
| --------------- | -------------------- | --------------------------- | ------------------------------------------------------------------------------ |
| FirecrawlAPIKey | `FIRECRAWL_API_KEY`  | *(none)*                    | API key for Firecrawl search engine                                            |
| FirecrawlAPIURL | `FIRECRAWL_API_URL`  | `https://api.firecrawl.dev` | Base URL for the Firecrawl API (override to point at a self-hosted deployment) |

### Perplexity Search

| Option                | Environment Variable      | Default Value | Description                                                  |
| --------------------- | ------------------------- | ------------- | ------------------------------------------------------------ |
| PerplexityAPIKey      | `PERPLEXITY_API_KEY`      | *(none)*      | API key for Perplexity search engine                         |
| PerplexityModel       | `PERPLEXITY_MODEL`        | `sonar-pro`   | Model to use for Perplexity search                           |
| PerplexityContextSize | `PERPLEXITY_CONTEXT_SIZE` | `low`         | Context size for Perplexity search (`low`, `medium`, `high`) |

### Searxng Search

| Option            | Environment Variable | Default Value | Description                                                         |
| ----------------- | -------------------- | ------------- | ------------------------------------------------------------------- |
| SearxngURL        | `SEARXNG_URL`        | *(none)*      | Base URL for Searxng meta search engine instance                    |
| SearxngCategories | `SEARXNG_CATEGORIES` | `general`     | Search categories to use (e.g., `general`, `news`, `web`)           |
| SearxngLanguage   | `SEARXNG_LANGUAGE`   | *(none)*      | Language filter for search results (e.g., `en`, `ch`)               |
| SearxngSafeSearch | `SEARXNG_SAFESEARCH` | `0`           | Safe search filter level (`0` = none, `1` = moderate, `2` = strict) |
| SearxngTimeRange  | `SEARXNG_TIME_RANGE` | *(none)*      | Time range filter (e.g., `day`, `month`, `year`)                    |
| SearxngTimeout    | `SEARXNG_TIMEOUT`    | *(none)*      | Request timeout in seconds for Searxng API calls                    |

### Internal Analytics Engine

An optional, opt-in fallback engine for the `web_search` tool's analytic modes (`answer`/`research`). When enabled, it discovers links via the first available link engine, fetches each page's main-content markdown through the browser scraper, and asks the summarizer to synthesize a query-focused answer — without a paid analytic API. Off by default because per-page scraping and summarization can cost more than a purpose-built third-party analytic call. Requires a configured scraper and at least one available link engine (e.g. DuckDuckGo, Google).

| Option                        | Environment Variable                 | Default Value | Description                                                   |
| ----------------------------- | ------------------------------------- | -------------- | -------------------------------------------------------------- |
| WebSearchInternalEnabled      | `WEB_SEARCH_INTERNAL_ENABLED`         | `false`        | Enable or disable the internal browser-analytics engine        |
| WebSearchInternalMaxSites     | `WEB_SEARCH_INTERNAL_MAX_SITES`       | `5`            | Maximum number of pages to fetch and summarize per query       |
| WebSearchInternalMaxSiteBytes | `WEB_SEARCH_INTERNAL_MAX_SITE_BYTES`  | `10240`        | Maximum markdown bytes read from each page before truncation   |

### Usage Details

The search engine settings are used in `pkg/tools/tools.go` to configure various search providers that AI agents can use:

```go
// Google Search configuration
googleSearch: &functions.GoogleSearchFunc{
    apiKey:    fte.cfg.GoogleAPIKey,
    cxKey:     fte.cfg.GoogleCXKey,
    lrKey:     fte.cfg.GoogleLRKey,
    proxyURL:  fte.cfg.ProxyURL,
},

// Traversaal Search configuration
traversaalSearch: &functions.TraversaalSearchFunc{
    apiKey:    fte.cfg.TraversaalAPIKey,
    proxyURL:  fte.cfg.ProxyURL,
},

// Tavily Search configuration
tavilySearch: &functions.TavilySearchFunc{
    apiKey:     fte.cfg.TavilyAPIKey,
    proxyURL:   fte.cfg.ProxyURL,
    summarizer: cfg.Summarizer,
},

// Firecrawl Search configuration (FirecrawlAPIKey / FirecrawlAPIURL read from fte.cfg)
firecrawl := NewFirecrawlTool(
    fte.cfg,
    fte.flowID,
    cfg.TaskID,
    cfg.SubtaskID,
    fte.slp,
    cfg.Summarizer,
)

// Perplexity Search configuration
perplexitySearch: &functions.PerplexitySearchFunc{
    apiKey:      fte.cfg.PerplexityAPIKey,
    proxyURL:    fte.cfg.ProxyURL,
    model:       fte.cfg.PerplexityModel,
    contextSize: fte.cfg.PerplexityContextSize,
    summarizer:  cfg.Summarizer,
},

// Sploitus Search configuration
sploitus := NewSploitusTool(
    fte.flowID,
    cfg.TaskID,
    cfg.SubtaskID,
    fte.cfg.SploitusEnabled,
    fte.cfg.ProxyURL,
    fte.slp,
)

// Searxng Search configuration
searxng := NewSearxngTool(
    fte.flowID,
    cfg.TaskID,
    cfg.SubtaskID,
    fte.cfg.SearxngURL,
    fte.cfg.SearxngCategories,
    fte.cfg.SearxngLanguage,
    fte.cfg.SearxngSafeSearch,
    fte.cfg.SearxngTimeRange,
    fte.cfg.ProxyURL,
    fte.cfg.SearxngTimeout,
    fte.slp,
    cfg.Summarizer,
)
```

These settings enable:
- Access to multiple search engines for diverse information sources
- Configuration of search parameters like language, context size, and time range
- Integration of search capabilities into the AI agent's toolset
- Web information gathering with different search strategies
- Security research through Sploitus, providing access to exploit databases and CVE information
- Meta-search capabilities through Searxng, aggregating results from multiple search engines

Having multiple search engine options ensures redundancy and provides different search algorithms for varied information needs. Sploitus is specifically designed for security research, providing comprehensive exploit and vulnerability information essential for penetration testing. Searxng is particularly useful as it provides aggregated results from multiple search engines while offering enhanced privacy and customization options.

## Network and Proxy Settings

These settings control HTTP proxy, SSL configuration, and network timeouts for outbound connections, which are important for network security and access control.

| Option              | Environment Variable    | Default Value | Description                                                      |
| ------------------- | ----------------------- | ------------- | ---------------------------------------------------------------- |
| ProxyURL            | `PROXY_URL`             | *(none)*      | URL for HTTP proxy (e.g., `http://user:pass@proxy:8080`)         |
| ExternalSSLCAPath   | `EXTERNAL_SSL_CA_PATH`  | *(none)*      | Path to trusted CA certificate for external LLM SSL connections  |
| ExternalSSLInsecure | `EXTERNAL_SSL_INSECURE` | `false`       | Skip SSL certificate verification for external connections       |
| HTTPClientTimeout   | `HTTP_CLIENT_TIMEOUT`   | `600`         | Timeout in seconds for external API calls (0 = no timeout)       |

### Usage Details

The proxy settings are used in various places to configure HTTP clients for external API calls:

```go
// Example from openai.go, anthropic.go, and other provider files
if cfg.ProxyURL != "" {
    httpClient = &http.Client{
        Transport: &http.Transport{
            Proxy: func(req *http.Request) (*url.URL, error) {
                return url.Parse(cfg.ProxyURL)
            },
        },
    }
}
```

The proxy URL is also passed to various tools that make external requests:

```go
// In tools.go for search tools
googleSearch: &functions.GoogleSearchFunc{
    apiKey:    fte.cfg.GoogleAPIKey,
    cxKey:     fte.cfg.GoogleCXKey,
    lrKey:     fte.cfg.GoogleLRKey,
    proxyURL:  fte.cfg.ProxyURL,
},
```

The proxy setting is essential for:
- Routing all outbound API requests through a controlled proxy
- Implementing network-level security policies
- Enabling access to external services from restricted networks
- Monitoring and auditing external API usage

The SSL settings provide additional security configuration:

- **ExternalSSLCAPath**: Specifies a custom CA certificate for validating SSL connections to external services:
  ```go
  // Used in provider initialization to configure custom CA certificates
  if cfg.ExternalSSLCAPath != "" {
      caCert, err := os.ReadFile(cfg.ExternalSSLCAPath)
      // Configure TLS with custom CA
  }
  ```
  This is useful when connecting to LLM providers with self-signed certificates or internal CAs.

- **ExternalSSLInsecure**: Allows skipping SSL certificate verification:
  ```go
  // Used in HTTP client configuration
  if cfg.ExternalSSLInsecure {
      tlsConfig.InsecureSkipVerify = true
  }
  ```
  **Warning**: Only use this in development or trusted environments. Skipping certificate verification exposes connections to man-in-the-middle attacks.

- **HTTPClientTimeout**: Sets the timeout for all external HTTP requests (LLM providers, search engines, etc.):
  ```go
  // Used in pkg/system/utils.go for HTTP client configuration
  timeout := defaultHTTPClientTimeout
  if cfg.HTTPClientTimeout > 0 {
      timeout = time.Duration(cfg.HTTPClientTimeout) * time.Second
  }
  
  httpClient := &http.Client{
      Timeout: timeout,
  }
  ```
  The default value of 600 seconds (10 minutes) is suitable for most LLM API calls, including long-running operations. Setting this to 0 disables the timeout (not recommended in production), while very low values may cause legitimate requests to fail. This setting affects:
  - All LLM provider API calls (OpenAI, Anthropic, Bedrock, etc.)
  - Search engine requests (Google, Tavily, Perplexity, etc.)
  - External tool integrations
  - Embedding generation requests

  Adjust this value based on your network conditions and the complexity of operations being performed.

## Graphiti Knowledge Graph Settings

Graphiti is an optional beta integration. This section documents the PentAGI-side configuration boundary and lifecycle; the operator-facing deployment, provider, ingestion, extraction, and Neo4j settings are maintained in [README.md — Knowledge Graph Integration](../../README.md#knowledge-graph-integration-graphiti).

### PentAGI Configuration Boundary

`pkg/config.Config` intentionally owns only the settings required by the PentAGI process:

| Field | Environment Variable | Default | Responsibility |
| --- | --- | --- | --- |
| `GraphitiEnabled` | `GRAPHITI_ENABLED` | `false` | Operator intent to enable the integration |
| `GraphitiURL` | `GRAPHITI_URL` | *(empty)* | Graphiti API base URL; embedded deployments use `http://graphiti:8000` |
| `GraphitiTimeout` | `GRAPHITI_TIMEOUT` | `30` seconds | Timeout used for Graphiti client requests and storage contexts |

All other variables with `GRAPHITI_*`, `NEO4J_*`, provider, embedding, ingest, extraction, anchor, or logging names configure `docker-compose-graphiti.yml` or the Graphiti process. They are not parsed into the Go `Config` struct. Keep that boundary explicit when adding settings: a value needed by PentAGI belongs in `pkg/config/config.go`; a value consumed only by the sidecar belongs in `.env.example` plus the compose mapping.

The provider controller applies an additional URL guard:

```go
graphitiClient, err := graphiti.NewClient(
    cfg.GraphitiURL,
    time.Duration(cfg.GraphitiTimeout)*time.Second,
    cfg.GraphitiEnabled && cfg.GraphitiURL != "",
)
```

Consequently, `GRAPHITI_ENABLED=true` with an empty URL still creates a disabled wrapper.

### Client Lifecycle and Failure Behavior

`pkg/graphiti/client.go` wraps `graphiti-go-client` and performs a synchronous health check during provider initialization. It makes three attempts with a two-second backoff. A final failure is returned to `pkg/providers/providers.go`, which logs a warning and substitutes a disabled client so the rest of PentAGI can start.

The disabled wrapper has deliberate asymmetric behavior:

- `AddMessages` is a no-op, allowing normal flow execution to continue.
- Search methods return `graphiti is not enabled`; disabled clients are normally excluded from agent tool registration and prompts through `IsEnabled`.
- Storage calls use `GraphitiTimeout`, log failures with their group ID, and return the error to the caller without making Graphiti a required persistence layer.

Do not change initialization failures into fatal PentAGI startup errors without treating that as a deployment-contract change.

### Data Flow, Search, and Tenancy

Provider performers render `backend/pkg/templates/graphiti/*.tmpl` and enqueue agent responses and tool executions with observation metadata. The sidecar applies ingest policy and performs asynchronous extraction; successful submission does not imply that the graph is immediately searchable.

Enabled agents receive the `graphiti_search` tool. Its seven modes map to temporal-window, entity-relationship, diverse-result, episode-context, successful-tool, recent-context, and entity-by-label client methods. Transport and server failures are handled as degradable tool failures where possible, while malformed arguments and validation errors remain hard errors so the LLM can repair its call.

Every request is partitioned by `Config.GroupID(flowID)`:

- without `TENANT_ID`: `flow-<id>`;
- with `TENANT_ID`: `<tenant>-flow-<id>`.

`Config.ParseGroupID` rejects identifiers belonging to another tenant. Keep `GroupID` and `ParseGroupID` exact inverses, and never accept a client-supplied namespace in place of a server-derived group ID. The stock deployment also sets `GRAPHITI_SEARCH_SCOPE=flowid`; changing it to global search bypasses this retrieval boundary inside Graphiti and is unsafe for shared deployments.

### Deployment Ownership

The bundled stack is separate from `docker-compose.yml`. The installer distinguishes disabled, external, and embedded deployments using `GRAPHITI_ENABLED` and `GRAPHITI_URL`; `http://graphiti:8000` is the embedded endpoint. It starts the Graphiti stack before PentAGI so the startup health check can succeed.

Installer assets include `docker-compose-graphiti.yml` and the `graphiti` preset directory sourced from `examples/graphiti`. The target directory is checked, repaired, and removed with the Graphiti stack. Compose mounts it through `GRAPHITI_CONFIG_PATH` and `GRAPHITI_CONFIG_DIR`; the fallback target `configs` prevents a newer compose file paired with an older `.env` from masking the image's built-in `llm_configs` with an empty host directory.

Models and call parameters belong to the provider YAML presets, selected by `GRAPHITI_LLM_CLIENT_TYPE`; they are not fields in the PentAGI Go configuration. `GRAPHITI_MODEL_NAME` is obsolete. Refer operators to the README instead of duplicating the sidecar's tuning reference here.

## Agent Supervision Settings

These settings control the agent supervision system, including execution monitoring and tool call limits for different agent types.

| Option                         | Environment Variable                | Default Value | Description                                                            |
| ------------------------------ | ----------------------------------- | ------------- | ---------------------------------------------------------------------- |
| ExecutionMonitorEnabled        | `EXECUTION_MONITOR_ENABLED`         | `false`       | Enable automatic execution monitoring (mentor/adviser supervision)     |
| ExecutionMonitorSameToolLimit  | `EXECUTION_MONITOR_SAME_TOOL_LIMIT` | `5`           | Threshold for consecutive identical tool calls before mentor review    |
| ExecutionMonitorTotalToolLimit | `EXECUTION_MONITOR_TOTAL_TOOL_LIMIT`| `10`          | Threshold for total tool calls before mentor review                    |
| MaxGeneralAgentToolCalls       | `MAX_GENERAL_AGENT_TOOL_CALLS`      | `100`         | Maximum tool calls for general agents (Assistant, Primary, Pentester, Coder, Installer) |
| MaxLimitedAgentToolCalls       | `MAX_LIMITED_AGENT_TOOL_CALLS`      | `20`          | Maximum tool calls for limited agents (Searcher, Enricher, etc.)       |
| AgentPlanningStepEnabled       | `AGENT_PLANNING_STEP_ENABLED`       | `false`       | Enable automatic task planning for specialist agents                   |

### Usage Details

The agent supervision settings are used in `pkg/providers/providers.go` and `pkg/providers/performer.go` to configure supervision mechanisms:

- **ExecutionMonitorEnabled**: Controls whether execution monitoring (mentor) is active:
  ```go
  buildMonitor: func() *executionMonitor {
      return &executionMonitor{
          enabled:        pc.cfg.ExecutionMonitorEnabled,
          sameThreshold:  pc.cfg.ExecutionMonitorSameToolLimit,
          totalThreshold: pc.cfg.ExecutionMonitorTotalToolLimit,
      }
  }
  ```

- **ExecutionMonitorSameToolLimit**: Sets the threshold for identical consecutive tool calls:
  ```go
  // In executionMonitor.shouldInvokeMentor
  if emd.sameToolCount >= emd.sameThreshold {
      // Invoke mentor (adviser agent) for execution review
  }
  ```
  When an agent calls the same tool this many times consecutively, the execution monitor automatically invokes the mentor (adviser agent) to analyze progress and provide guidance.

- **ExecutionMonitorTotalToolLimit**: Sets the threshold for total tool calls:
  ```go
  // In executionMonitor.shouldInvokeMentor
  if emd.totalCallCount >= emd.totalThreshold {
      // Invoke mentor (adviser agent) for execution review
  }
  ```
  When an agent makes this many total tool calls since the last mentor review, the execution monitor automatically invokes the mentor to prevent inefficient loops and provide strategic guidance.

- **MaxGeneralAgentToolCalls**: Maximum iterations for general-purpose agents with full capabilities:
  ```go
  // In performAgentChain
  switch optAgentType {
  case pconfig.OptionsTypeAssistant, pconfig.OptionsTypePrimaryAgent,
      pconfig.OptionsTypePentester, pconfig.OptionsTypeCoder, pconfig.OptionsTypeInstaller:
      if fp.maxGACallsLimit <= 0 {
          maxCallsLimit = maxGeneralAgentChainIterations // fallback: 100
      } else {
          maxCallsLimit = max(fp.maxGACallsLimit, maxAgentShutdownIterations*2)
      }
  }
  ```
  General agents (Assistant, Primary Agent, Pentester, Coder, Installer) are designed for complex, multi-step workflows and have a higher tool call limit to complete sophisticated tasks.

- **MaxLimitedAgentToolCalls**: Maximum iterations for specialized, limited-scope agents:
  ```go
  // In performAgentChain
  default:
      if fp.maxLACallsLimit <= 0 {
          maxCallsLimit = maxLimitedAgentChainIterations // fallback: 20
      } else {
          maxCallsLimit = max(fp.maxLACallsLimit, maxAgentShutdownIterations*2)
      }
  }
  ```
  Limited agents (Searcher, Enricher, Memorist, Generator, Reporter, Adviser, Reflector, Planner) are designed for focused, specific tasks and have a lower tool call limit to ensure efficient execution.

- **AgentPlanningStepEnabled**: Controls automatic task planning for specialist agents:
  ```go
  // In flowProvider initialization
  planning: pc.cfg.AgentPlanningStepEnabled
  
  // Used when invoking specialist agents
  if fp.planning {
      // Generate execution plan via planner before specialist execution
  }
  ```
  When enabled, the planner (adviser in planning mode) generates a structured 3-7 step execution plan before specialist agents (Pentester, Coder, Installer) begin their work, improving task completion rates and preventing scope creep.

These settings enable:
- **Automatic supervision**: Mentor reviews execution patterns to detect loops and inefficiencies
- **Graceful termination**: Reflector guides agents to proper completion when approaching limits
- **Differentiated capabilities**: General agents have more autonomy for complex workflows
- **Efficient execution**: Limited agents stay focused on their specific scope
- **Strategic planning**: Automatic task decomposition for better execution quality

### Supervision System Integration

The supervision settings work together as a comprehensive system:

1. **Execution Monitoring** (via ExecutionMonitor settings):
   - Detects repetitive patterns (same tool called N times)
   - Detects excessive exploration (total tools called N times)
   - Automatically invokes mentor for guidance and correction
   - Resets counters after mentor review

2. **Tool Call Limits** (via MaxGeneralAgentToolCalls and MaxLimitedAgentToolCalls):
   - Prevents runaway executions with hard limits
   - Invokes reflector for graceful termination near limit
   - Different limits for different agent capabilities
   - Ensures system stability and resource efficiency

3. **Task Planning** (via AgentPlanningStepEnabled):
   - Generates structured execution plans before specialist work
   - Prevents scope creep and maintains focus
   - Improves success rates for complex tasks
   - Provides clear verification points

### Recommended Settings

1. **Production Environment**:
   ```
   ExecutionMonitorEnabled: false
   ExecutionMonitorSameToolLimit: 5
   ExecutionMonitorTotalToolLimit: 10
   MaxGeneralAgentToolCalls: 100
   MaxLimitedAgentToolCalls: 20
   AgentPlanningStepEnabled: false
   ```
   Default settings provide stable execution without beta features.

2. **High-Complexity Workflows**:
   ```
   ExecutionMonitorEnabled: false
   ExecutionMonitorSameToolLimit: 7
   ExecutionMonitorTotalToolLimit: 15
   MaxGeneralAgentToolCalls: 150
   MaxLimitedAgentToolCalls: 30
   AgentPlanningStepEnabled: false
   ```
   Increased limits for tasks requiring extensive exploration and iteration.

3. **Resource-Constrained Environment**:
   ```
   ExecutionMonitorEnabled: false
   ExecutionMonitorSameToolLimit: 3
   ExecutionMonitorTotalToolLimit: 7
   MaxGeneralAgentToolCalls: 50
   MaxLimitedAgentToolCalls: 15
   AgentPlanningStepEnabled: false
   ```
   Tighter limits to reduce resource usage.

4. **Debugging Mode**:
   ```
   ExecutionMonitorEnabled: false
   MaxGeneralAgentToolCalls: 200
   MaxLimitedAgentToolCalls: 50
   AgentPlanningStepEnabled: false
   ```
   Disabled supervision for debugging to observe natural agent behavior.

## Observability Settings

These settings control the observability and monitoring capabilities, including telemetry and trace collection for system performance and debugging.

### Telemetry

| Option            | Environment Variable | Default Value | Description                                |
| ----------------- | -------------------- | ------------- | ------------------------------------------ |
| TelemetryEndpoint | `OTEL_HOST`          | *(none)*      | Endpoint for OpenTelemetry data collection |

### Langfuse

| Option            | Environment Variable  | Default Value | Description                 |
| ----------------- | --------------------- | ------------- | --------------------------- |
| LangfuseBaseURL   | `LANGFUSE_BASE_URL`   | *(none)*      | Base URL for Langfuse API   |
| LangfuseProjectID | `LANGFUSE_PROJECT_ID` | *(none)*      | Project ID for Langfuse     |
| LangfusePublicKey | `LANGFUSE_PUBLIC_KEY` | *(none)*      | Public key for Langfuse API |
| LangfuseSecretKey | `LANGFUSE_SECRET_KEY` | *(none)*      | Secret key for Langfuse API |

### Usage Details

The observability settings are used in `main.go` and the observability package to initialize monitoring systems:

- **Telemetry Configuration**: Sets up OpenTelemetry for metrics, logs, and traces:
  ```go
  // Check if telemetry is configured
  if cfg.TelemetryEndpoint == "" {
      return nil, ErrNotConfigured
  }

  // Create telemetry client with endpoint
  otelclient, err := obs.NewTelemetryClient(ctx, cfg)
  ```

- **Langfuse Configuration**: Configures Langfuse for LLM operation monitoring:
  ```go
  // Check if Langfuse is configured
  if cfg.LangfuseBaseURL == "" {
      return nil, ErrNotConfigured
  }

  // Configure Langfuse client
  langfuse.WithBaseURL(cfg.LangfuseBaseURL),
  langfuse.WithPublicKey(cfg.LangfusePublicKey),
  langfuse.WithSecretKey(cfg.LangfuseSecretKey),
  langfuse.WithProjectID(cfg.LangfuseProjectID),
  ```

- **Integration in Application**: Used in `main.go` to initialize observability:
  ```go
  lfclient, err := obs.NewLangfuseClient(ctx, cfg)
  if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
      log.Fatalf("Unable to create langfuse client: %v\n", err)
  }

  otelclient, err := obs.NewTelemetryClient(ctx, cfg)
  if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
      log.Fatalf("Unable to create telemetry client: %v\n", err)
  }

  obs.InitObserver(ctx, lfclient, otelclient, []logrus.Level{
      logrus.DebugLevel,
      logrus.InfoLevel,
      logrus.WarnLevel,
      logrus.ErrorLevel,
  })
  ```

These settings enable:
- Comprehensive monitoring of system performance
- LLM-specific metrics collection via Langfuse
- Tracing of requests through the system
- Centralized logging for troubleshooting
- Performance optimization based on collected metrics
