# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Core Interaction Rules

1. **Always use English** for all interactions, responses, explanations, and questions with users.
2. **Password Complexity Requirements**: For all password-related development (registration, password reset, API token generation, etc.), the following rules must be enforced:
   - Minimum 12 characters
   - Must contain at least 1 uppercase letter, 1 lowercase letter, 1 number, and 1 special character
   - Common weak passwords (e.g., `password`, `123456`) are prohibited
   - Both backend and frontend validation must be implemented; do not rely on frontend validation alone

## Project Overview

**PentAGI** is an automated security testing platform powered by AI agents. It runs autonomous penetration testing workflows using a multi-agent system (Researcher, Developer, Executor agents) that coordinates LLM providers, Docker-sandboxed tool execution, and a persistent vector memory store.

The application is a monorepo with:
- **`backend/`** — Go REST + GraphQL API server
- **`frontend/`** — React + TypeScript web UI
- **`observability/`** — Optional monitoring stack configs

## Build & Development Commands

### Backend (run from `backend/`)

```bash
go mod download                              # Install dependencies
go build -trimpath -o pentagi ./cmd/pentagi  # Build main binary
go test ./...                                # Run all tests
go test ./pkg/foo/... -v -run TestName       # Run specific test
golangci-lint run --timeout=5m               # Lint

# Code generation (run after schema changes)
go run github.com/99designs/gqlgen --config ./gqlgen/gqlgen.yml  # GraphQL resolvers
swag init -g ../../pkg/server/router.go -o pkg/server/docs/ --parseDependency --parseInternal --parseDepth 2 -d cmd/pentagi  # Swagger docs
```

### Frontend (run from `frontend/`)

```bash
npm ci                    # Install dependencies
npm run dev               # Dev server on http://localhost:8000
npm run build             # Production build
npm run lint              # ESLint check
npm run lint:fix          # ESLint auto-fix
npm run prettier          # Prettier check
npm run prettier:fix      # Prettier auto-format
npm run test              # Vitest
npm run test:coverage     # Coverage report
npm run graphql:generate  # Regenerate GraphQL types from schema
```

### Docker (run from repo root)

```bash
docker compose up -d                                                          # Start core services
docker compose -f docker-compose.yml -f docker-compose-observability.yml up -d  # + monitoring
docker compose -f docker-compose.yml -f docker-compose-langfuse.yml up -d       # + LLM analytics
docker compose -f docker-compose.yml -f docker-compose-graphiti.yml up -d       # + knowledge graph
docker build -t local/pentagi:latest .                                        # Build image
```

The full stack runs at `https://localhost:8443` when using Docker Compose. Copy `.env.example` to `.env` and fill in at minimum the database and at least one LLM provider key.

## Architecture

### Backend Package Structure

| Package | Role |
|---|---|
| `cmd/pentagi/` | Main entry point; initializes config, DB, server |
| `pkg/config/` | Environment-based config parsing |
| `pkg/server/` | Gin router, middleware, auth (JWT/OAuth2/API tokens), Swagger |
| `pkg/controller/` | Business logic for REST endpoints |
| `pkg/graph/` | gqlgen GraphQL schema (`schema.graphqls`) and resolvers |
| `pkg/database/` | GORM models, SQLC queries, goose migrations |
| `pkg/providers/` | LLM provider adapters (OpenAI, Anthropic, Gemini, Bedrock, Ollama, etc.) |
| `pkg/tools/` | Penetration testing tool integrations |
| `pkg/docker/` | Docker SDK wrapper for sandboxed container execution |
| `pkg/terminal/` | Terminal session and command execution management |
| `pkg/queue/` | Async task queue |
| `pkg/csum/` | Chain summarization for LLM context management |
| `pkg/graphiti/` | Knowledge graph (Neo4j via Graphiti) integration |
| `pkg/observability/` | OpenTelemetry tracing, metrics, structured logging |

Database migrations live in `backend/migrations/sql/` and run automatically via goose at startup.

### Frontend Structure

```
frontend/src/
├── app.tsx / main.tsx     # Entry points and router setup
├── pages/                 # Route-level page components
│   ├── flows/             # Flow management UI
│   └── settings/          # Provider, prompt, token settings
├── components/
│   ├── layouts/           # App shell layouts
│   └── ui/                # Base Radix UI components
├── graphql/               # Auto-generated Apollo types (do not edit)
├── hooks/                 # Custom React hooks
├── lib/                   # Apollo client, HTTP utilities
└── schemas/               # Zod validation schemas
```

State is managed primarily through Apollo Client (GraphQL) with real-time updates via GraphQL subscriptions over WebSocket.

### Data Flow

1. User creates a "flow" (penetration test) via the UI or REST API.
2. The backend queues the flow and spawns agent goroutines.
3. The Researcher agent gathers information; the Developer plans attack strategies; the Executor runs tools in isolated Docker containers.
4. Results, tool outputs, and LLM reasoning are stored in PostgreSQL (with pgvector for semantic search/memory).
5. Real-time progress is pushed to the frontend via GraphQL subscriptions.

### Authentication

- **Session cookies** for browser login (secure, httpOnly)
- **OAuth2** via Google and GitHub
- **Bearer tokens** (API tokens table) for programmatic API access

### Key Integrations

- **LLM Providers**: OpenAI, Anthropic, Gemini, AWS Bedrock, Ollama, DeepSeek, GLM, Kimi, Qwen, and custom HTTP endpoints — configured via environment variables or the Settings UI
- **Search**: DuckDuckGo, Google, Tavily, Traversaal, Perplexity, Searxng
- **Databases**: PostgreSQL + pgvector (required), Neo4j (optional, for knowledge graph)
- **Observability**: OpenTelemetry → VictoriaMetrics + Loki + Jaeger → Grafana; Langfuse for LLM analytics

### Adding a New LLM Provider

1. Create `backend/pkg/providers/<name>/<name>.go` implementing the `provider.Provider` interface.
2. Add a new `Provider<Name> ProviderType` constant and `DefaultProviderName<Name>` in `pkg/providers/provider/provider.go`.
3. Register the provider in `pkg/providers/providers.go` (`DefaultProviderConfig`, `NewProvider`, `buildProviderFromConfig`, `GetProvider`).
4. Add the new type to the `Valid()` whitelist in `pkg/server/models/providers.go` — **without this step, the REST API returns 422 Unprocessable Entity**.
5. Add the env var key to `pkg/config/config.go` (e.g., `<NAME>_API_KEY`, `<NAME>_SERVER_URL`).
6. Add the new `PROVIDER_TYPE` enum value via a goose migration in `backend/migrations/sql/`.
7. Add the provider icon in `frontend/src/components/icons/<name>.tsx` and register it in `frontend/src/components/icons/provider-icon.tsx`.
8. Update the GraphQL schema/types and frontend settings page if needed.

### Code Generation

When modifying `backend/pkg/graph/schema.graphqls`, re-run the gqlgen command to regenerate resolver stubs. When modifying REST handler annotations, re-run swag to update Swagger docs. When modifying `frontend/src/graphql/*.graphql` query files, re-run `npm run graphql:generate` to update TypeScript types.

### Utility Binaries

The backend contains helper binaries for development/testing:
- `cmd/ctester/` — tests container execution
- `cmd/ftester/` — tests LLM function/tool calling
- `cmd/etester/` — tests embedding providers
- `cmd/installer/` — interactive TUI wizard for guided deployment setup (configures `.env`, Docker Compose, DB, search engines, etc.)
