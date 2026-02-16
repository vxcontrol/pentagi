# Pentagi Neo4j Server

GraphQL server (Apollo Server + Express) that queries Neo4j and exposes pentest metrics for the frontend. Supports real-time updates via WebSocket subscriptions.

## Setup

```bash
cd frontend/neo4j-server
cp .env.example .env
# Edit .env with your Neo4j credentials
npm install
```

## Run

### Standalone

```bash
npm start          # production
npm run dev        # with auto-reload (--watch)
```

### Together with frontend (recommended)

From the `frontend/` directory:

```bash
npm run dev        # starts both Vite and neo4j-server via concurrently
npm run dev:neo4j  # starts only neo4j-server
```

Server listens on `http://localhost:4000` (or `PORT` from `.env`).
Apollo Explorer is available at `http://localhost:4000/graphql`.

## Environment variables

| Variable           | Default                  | Description                             |
| ------------------ | ------------------------ | --------------------------------------- |
| `PORT`             | `4000`                   | Server listen port                      |
| `NEO4J_URI`        | `bolt://localhost:7687`  | Neo4j connection URI                    |
| `NEO4J_USER`       | `neo4j`                  | Neo4j username                          |
| `NEO4J_PASSWORD`   | `neo4j`                  | Neo4j password                          |
| `NEO4J_DATABASE`   | `neo4j`                  | Neo4j database name                     |
| `NEO4J_ENCRYPTED`  | _(auto from URI)_        | Force encryption (`true` / `false`)     |
| `POLL_INTERVAL_MS` | `5000`                   | Subscription polling interval (ms)      |

## Endpoints

| Path       | Protocol       | Description                                   |
| ---------- | -------------- | --------------------------------------------- |
| `/graphql` | HTTP (POST)    | GraphQL queries and mutations                 |
| `/graphql` | WebSocket (WS) | GraphQL subscriptions (real-time updates)     |
| `/health`  | HTTP (GET)     | Health check (`{ status: "ok", neo4j: "…" }`) |

## GraphQL schema

### Queries

All queries require `groupId: String!` parameter (flow ID).

| Query                    | Returns                    | Description                          |
| ------------------------ | -------------------------- | ------------------------------------ |
| `health`                 | `HealthStatus`             | Health check                         |
| `dashboard`              | `PentestSummary`           | Pentest summary stats                |
| `attackSurface`          | `[AttackSurfaceEntity]`    | Entity counts by type                |
| `credentialsStatus`      | `[CredentialStatus]`       | Credential status breakdown          |
| `accessDetails`          | `[AccessRecord]`           | Valid access details                 |
| `hostsWithServices`      | `[HostWithServices]`       | Hosts with their services            |
| `openPorts`              | `[OpenPort]`               | Open ports with services             |
| `vulnerabilitySeverity`  | `[VulnerabilitySeverity]`  | Vulnerability severity breakdown     |
| `allCves`                | `[CveRecord]`              | All discovered CVEs                  |
| `exploitAttempts`        | `[ExploitAttempt]`         | Exploit attempts and their status    |
| `toolUsage`              | `[ToolUsage]`              | Tool execution counts                |
| `toolEffectiveness`      | `[ToolEffectiveness]`      | Tool effectiveness with discoveries  |
| `artifacts`              | `[Artifact]`               | Produced artifacts                   |
| `mainAttackChain`        | `GraphData`                | Main attack chain graph              |
| `fullAttackChain`        | `GraphData`                | Full attack chain graph              |
| `infrastructureGraph`    | `GraphData`                | Infrastructure-only graph            |
| `accessChainGraph`       | `GraphData`                | Access chain graph                   |
| `shortestPathGraph`      | `GraphData`                | Shortest path graph                  |
| `attackPathStats`        | `AttackPathStats`          | Attack path statistics               |

### Subscriptions (WebSocket)

Real-time updates pushed to clients at the configured polling interval.

| Subscription                     | Returns                   |
| -------------------------------- | ------------------------- |
| `dashboardUpdated(groupId)`      | `PentestSummary`          |
| `attackSurfaceUpdated(groupId)`  | `[AttackSurfaceEntity]`   |
| `mainAttackChainUpdated(groupId)`| `GraphData`               |
| `fullAttackChainUpdated(groupId)`| `GraphData`               |
| `attackPathStatsUpdated(groupId)`| `AttackPathStats`         |
| `vulnerabilitySeverityUpdated(groupId)` | `[VulnerabilitySeverity]` |

## Docker

The server is bundled with the frontend in a single Docker image (see `Dockerfile.fe` in the repo root). Nginx serves the static frontend and proxies `/graphql` to this server.

```bash
docker build -f Dockerfile.fe -t pentagi-frontend .
docker run -p 3000:3000 \
  -e NEO4J_URI=bolt://neo4j-host:7687 \
  -e NEO4J_USER=neo4j \
  -e NEO4J_PASSWORD=secret \
  -e NEO4J_DATABASE=graphiti \
  pentagi-frontend
```

Additional Docker environment variables:

| Variable     | Default | Description              |
| ------------ | ------- | ------------------------ |
| `NGINX_PORT` | `3000`  | Nginx listen port        |

All neo4j-server variables from the table above are also supported.
