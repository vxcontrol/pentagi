# Database Package Documentation

## Overview

The `database` package is a core component of PentAGI that provides a robust, type-safe interface for interacting with PostgreSQL database operations. Built on top of [sqlc](https://sqlc.dev/), this package automatically generates Go code from SQL queries, ensuring compile-time safety and eliminating the need for manual ORM mapping.

PentAGI uses PostgreSQL with the [pgvector](https://github.com/pgvector/pgvector) extension to support vector embeddings for AI-powered semantic search and memory storage capabilities.

## Architecture

### Database Technology Stack

- **Database Engine**: PostgreSQL 15+ with pgvector extension
- **Code Generation**: sqlc for type-safe SQL-to-Go compilation
- **ORM Support**: GORM v1 for advanced operations and HTTP server handlers
- **Schema Management**: Database migrations located in `backend/migrations/`
- **Vector Operations**: pgvector extension for AI embeddings and semantic search

### Entity Relationship Model

The database follows PentAGI's hierarchical data model for penetration testing workflows:

```
Flow (Top-level workflow)
├── Task (Major testing phases)
│   └── SubTask (Specific agent assignments)
│       └── Action (Individual operations)
│           ├── Artifact (Output files/data)
│           └── Memory (Knowledge/observations)
└── Assistant (AI assistants for flows)
    └── AssistantLog (Assistant interaction logs)
```

Additional supporting entities include:
- **Container**: Docker containers for isolated execution
- **User**: System users with role-based access
- **MsgChain**: LLM conversation chains
- **ToolCall**: Function calls made by AI agents
- **Various Logs**: Comprehensive audit trail for all operations

## SQL Query Organization

The database package is built on a comprehensive set of SQL queries organized by entity type in the `backend/sqlc/models/` directory. Each file contains CRUD operations and specialized queries for its respective entity.

### Query File Structure

| File                 | Entity       | Purpose                            |
| -------------------- | ------------ | ---------------------------------- |
| `flows.sql`          | Flow         | Top-level workflow management and analytics |
| `tasks.sql`          | Task         | Task lifecycle and status tracking |
| `subtasks.sql`       | SubTask      | Agent assignment and execution     |
| `assistants.sql`     | Assistant    | AI assistant management            |
| `containers.sql`     | Container    | Docker environment tracking        |
| `users.sql`          | User         | User management and authentication |
| `roles.sql`          | Role         | Role-based access control          |
| `prompts.sql`        | Prompt       | User-defined prompt templates      |
| `providers.sql`      | Provider     | LLM provider configurations        |
| `msgchains.sql`      | MsgChain     | LLM conversation chains and usage stats |
| `toolcalls.sql`      | ToolCall     | AI function call tracking and analytics |
| `screenshots.sql`    | Screenshot   | Visual artifacts storage           |
| `analytics.sql`      | Analytics    | Flow execution time and hierarchy analytics |
| **Logging Entities** |              |                                    |
| `agentlogs.sql`      | AgentLog     | Inter-agent communication          |
| `assistantlogs.sql`  | AssistantLog | Human-assistant interactions       |
| `msglogs.sql`        | MsgLog       | General message logging            |
| `searchlogs.sql`     | SearchLog    | External search operations         |
| `termlogs.sql`       | TermLog      | Terminal command execution         |
| `vecstorelogs.sql`   | VecStoreLog  | Vector database operations         |

### Query Naming Conventions

sqlc queries follow consistent naming patterns:

```sql
-- CRUD Operations
-- name: Create[Entity] :one
-- name: Get[Entity] :one
-- name: Get[Entities] :many
-- name: Update[Entity] :one
-- name: Delete[Entity] :exec/:one

-- Scoped Operations
-- name: GetUser[Entity] :one
-- name: GetUser[Entities] :many
-- name: GetFlow[Entity] :one
-- name: GetFlow[Entities] :many

-- Specialized Queries
-- name: Get[Entity][Condition] :many
-- name: Update[Entity][Field] :one
```

### Security and Multi-tenancy Patterns

Most queries implement user-scoped access through JOIN operations:

```sql
-- Example: User-scoped flow access
-- name: GetUserFlow :one
SELECT f.*
FROM flows f
INNER JOIN users u ON f.user_id = u.id
WHERE f.id = $1 AND f.user_id = $2 AND f.deleted_at IS NULL;

-- Example: Flow-scoped task access
-- name: GetFlowTasks :many
SELECT t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id
WHERE t.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY t.created_at ASC;
```

### Soft Delete Implementation

Critical entities implement soft deletes to maintain audit trails:

```sql
-- Soft delete operation
-- name: DeleteFlow :one
UPDATE flows
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- All queries filter soft-deleted records
WHERE f.deleted_at IS NULL
```

### Logging Query Patterns

Logging entities follow consistent patterns for audit trails:

```sql
-- name: CreateAgentLog :one
INSERT INTO agentlogs (
  initiator,     -- AI agent that initiated the action
  executor,      -- AI agent that executed the action
  task,          -- Description of the task
  result,        -- JSON result of the operation
  flow_id,       -- Associated flow
  task_id,       -- Associated task (nullable)
  subtask_id     -- Associated subtask (nullable)
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- Hierarchical retrieval with security joins
-- name: GetFlowAgentLogs :many
SELECT al.*
FROM agentlogs al
INNER JOIN flows f ON al.flow_id = f.id
WHERE al.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY al.created_at ASC;
```

### Complex Query Examples

#### Message Chain Management

```sql
-- Get conversation chains for a specific task
-- name: GetTaskPrimaryMsgChains :many
SELECT mc.*
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
WHERE (mc.task_id = $1 OR s.task_id = $1) AND mc.type = 'primary_agent'
ORDER BY mc.created_at DESC;

-- Update conversation usage tracking with duration
-- name: UpdateMsgChainUsage :one
UPDATE msgchains
SET 
  usage_in = usage_in + $1, 
  usage_out = usage_out + $2,
  usage_cache_in = usage_cache_in + $3,
  usage_cache_out = usage_cache_out + $4,
  usage_cost_in = usage_cost_in + $5,
  usage_cost_out = usage_cost_out + $6,
  duration_seconds = duration_seconds + $7
WHERE id = $8
RETURNING *;

// Get usage statistics for a specific flow
-- name: GetFlowUsageStats :one
SELECT
  COALESCE(SUM(mc.usage_in), 0) AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0) AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0) AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0) AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0) AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0) AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE (mc.flow_id = $1 OR t.flow_id = $1) AND f.deleted_at IS NULL;
```

#### Container Management with Constraints

```sql
-- Upsert container with conflict resolution
-- name: CreateContainer :one
INSERT INTO containers (
  type, name, image, status, flow_id, local_id, local_dir
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT ON CONSTRAINT containers_local_id_unique
DO UPDATE SET
  type = EXCLUDED.type,
  name = EXCLUDED.name,
  image = EXCLUDED.image,
  status = EXCLUDED.status,
  flow_id = EXCLUDED.flow_id,
  local_dir = EXCLUDED.local_dir
RETURNING *;
```

#### Role-Based Access Control

```sql
-- Complex role aggregation
-- name: GetUser :one
SELECT
  u.*,
  r.name AS role_name,
  (
    SELECT ARRAY_AGG(p.name)
    FROM privileges p
    WHERE p.role_id = r.id
  ) AS privileges
FROM users u
INNER JOIN roles r ON u.role_id = r.id
WHERE u.id = $1;
```

## Code Generation with sqlc

### Configuration

The package uses sqlc for code generation with the following configuration (`sqlc/sqlc.yml`):

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: ["models/*.sql"]
    schema: ["../migrations/sql/*.sql"]
    gen:
      go:
        package: "database"
        out: "../pkg/database"
        sql_package: "database/sql"
        emit_interface: true
        emit_json_tags: true
    database:
      uri: ${DATABASE_URL}
```

### Generation Command

Code generation is performed using Docker to ensure consistency:

```bash
docker run --rm -v "$(pwd):/src" --network pentagi-network \
  -e DATABASE_URL='postgres://postgres:postgres@pgvector:5432/pentagidb?sslmode=disable' \
  -w /src sqlc/sqlc generate -f sqlc/sqlc.yml
```

This command:
1. Mounts the current directory into the container
2. Connects to the PentAGI database network
3. Uses the PostgreSQL database URL for schema introspection
4. Generates type-safe Go code from SQL queries

## Core Components

### 1. Database Interface (`db.go`)

Provides the foundational database transaction interface:

```go
type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Queries struct {
    db DBTX
}
```

**Key Features:**
- Generic database transaction interface
- Support for both direct database connections and transactions
- Thread-safe query execution
- Context-aware operations for timeout handling

### 2. Database Utilities (`database.go`)

Contains utility functions and GORM integration:

```go
// Null value converters
func StringToNullString(s string) sql.NullString
func NullStringToPtrString(s sql.NullString) *string
func Int64ToNullInt64(i *int64) sql.NullInt64
func NullInt64ToInt64(i sql.NullInt64) *int64
func TimeToNullTime(t time.Time) sql.NullTime

// GORM configuration
func NewGorm(dsn, dbType string) (*gorm.DB, error)
```

**Key Features:**
- Null value handling for optional database fields
- GORM integration with custom logging
- Connection pooling configuration
- OpenTelemetry observability integration

### 3. Query Interface (`querier.go`)

Auto-generated interface containing all database operations:

```go
type Querier interface {
    // Flow operations
    CreateFlow(ctx context.Context, arg CreateFlowParams) (Flow, error)
    GetFlows(ctx context.Context) ([]Flow, error)
    GetUserFlow(ctx context.Context, arg GetUserFlowParams) (Flow, error)
    UpdateFlowStatus(ctx context.Context, arg UpdateFlowStatusParams) (Flow, error)
    DeleteFlow(ctx context.Context, id int64) (Flow, error)

    // Task operations
    CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error)
    GetFlowTasks(ctx context.Context, flowID int64) ([]Task, error)
    UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) (Task, error)

    // ... 150+ additional methods
}
```

**Features:**
- Complete CRUD operations for all entities
- User-scoped queries for multi-tenancy
- Efficient joins with foreign key relationships
- Soft delete support for critical entities

### 4. Model Converters (`converter/converter.go`)

Converts database models to GraphQL schema types:

```go
func ConvertFlows(flows []database.Flow, containers []database.Container) []*model.Flow
func ConvertFlow(flow database.Flow, containers []database.Container) *model.Flow
func ConvertTasks(tasks []database.Task, subtasks []database.Subtask) []*model.Task
func ConvertAssistants(assistants []database.Assistant) []*model.Assistant
```

**Key Functions:**
- Transform database types to GraphQL models
- Handle relationship mapping (flows → tasks → subtasks)
- Null value processing for optional fields
- Aggregation of related entities

## Data Models

### Core Workflow Entities

#### Flow
Top-level penetration testing workflow:
- `id`, `title`, `status` (active/completed/failed)
- `model`, `model_provider_name`, `model_provider_type` for AI configuration
- `language` for localization
- `tool_call_id_template` for customizing tool call ID format
- `functions` as JSON for AI behavior
- `trace_id` for observability
- `user_id` for multi-tenancy
- Soft delete with `deleted_at`

**Note**: Prompts are no longer stored in flows. They are managed separately through the `prompts` table and loaded dynamically based on `PROMPT_TYPE`.

#### Task
Major phases within a flow:
- `id`, `flow_id`, `title`, `status` (pending/running/done/failed)
- `input` for task parameters
- `result` JSON for task outputs
- Creation and update timestamps

#### SubTask
Specific assignments for AI agents:
- `id`, `task_id`, `title`, `description`
- `status` (created/waiting/running/finished/failed)
- `result` and `context` JSON fields
- Agent type classification

### Supporting Entities

#### Container
Docker execution environments:
- `type` (primary/secondary), `name`, `image`
- `status` (starting/running/stopped)
- `local_id` for Docker integration
- `local_dir` for volume mapping

#### Assistant
AI assistants for interactive flows:
- `title`, `status`, `model`, `model_provider_name`, `model_provider_type`
- `language` for localization
- `tool_call_id_template` for customizing tool call ID format
- `functions` configuration as JSON
- `use_agents` flag for delegation behavior
- `msgchain_id` for conversation tracking
- Flow association and soft delete

**Note**: Prompts are managed separately through the `prompts` table, not stored in assistants.

#### Message Chains (MsgChain)
LLM conversation management and usage tracking:
- `type` (primary_agent/assistant/generator/refiner/reporter/etc.)
- `model`, `model_provider` for tracking
- **Token usage tracking**:
  - `usage_in`, `usage_out` - input/output tokens
  - `usage_cache_in`, `usage_cache_out` - cached tokens (for prompt caching)
  - `usage_cost_in`, `usage_cost_out` - cost tracking in currency units
- **Duration tracking**:
  - `duration_seconds` - pre-calculated execution duration (DOUBLE PRECISION, NOT NULL, DEFAULT 0.0)
  - Automatically incremented during updates using delta from backend
  - Provides fast analytics without real-time calculations
- `chain` JSON for conversation history
- Multi-level association (flow/task/subtask)
- Creation and update timestamps for temporal analysis

#### Provider
LLM provider configurations for multi-provider support:
- `type` - PROVIDER_TYPE enum (openai/anthropic/gemini/bedrock/ollama/custom)
- `name` - user-defined provider name
- `config` - JSON configuration for API keys and settings
- `user_id` - user ownership
- Soft delete with `deleted_at`
- Unique constraint on (name, user_id) for active providers

#### Prompt
Centralized prompt template management:
- `type` - PROMPT_TYPE enum (primary_agent/assistant/pentester/coder/etc.)
- `prompt` - template content
- `user_id` - user ownership
- Creation and update timestamps

### Logging Entities

The package provides comprehensive logging for all system operations:

- **AgentLog**: Inter-agent communication and delegation
- **AssistantLog**: Human-assistant interactions
- **MsgLog**: General message logging (thoughts/browser/terminal/file/search/advice/ask/input/done)
- **SearchLog**: External search operations (google/tavily/traversaal/browser/duckduckgo/perplexity/sploitus/searxng)
- **TermLog**: Terminal command execution (stdin/stdout/stderr)
- **ToolCall**: AI function calling with duration tracking
  - `duration_seconds` - pre-calculated execution duration (DOUBLE PRECISION, NOT NULL, DEFAULT 0.0)
  - Automatically incremented during status updates using delta from backend
  - Only counts completed toolcalls (finished/failed) in analytics
- **VecStoreLog**: Vector database operations

## LLM Usage Analytics

The database package provides comprehensive analytics for tracking LLM usage, costs, and performance across all levels of the workflow hierarchy. This enables detailed monitoring of AI resource consumption and cost optimization.

### Usage Tracking Fields

The `msgchains` table tracks six key metrics for each conversation:

| Field             | Type             | Description                              |
| ----------------- | ---------------- | ---------------------------------------- |
| `usage_in`        | BIGINT           | Input tokens consumed                    |
| `usage_out`       | BIGINT           | Output tokens generated                  |
| `usage_cache_in`  | BIGINT           | Cached input tokens (for prompt caching) |
| `usage_cache_out` | BIGINT           | Cached output tokens                     |
| `usage_cost_in`   | DOUBLE PRECISION | Input cost in currency units             |
| `usage_cost_out`  | DOUBLE PRECISION | Output cost in currency units            |

### Analytics Queries

#### 1. Hierarchical Usage Statistics

Get aggregated usage for specific entities:

```go
// Get total usage for a flow
stats, err := db.GetFlowUsageStats(ctx, flowID)

// Get total usage for a task
stats, err := db.GetTaskUsageStats(ctx, taskID)

// Get total usage for a subtask
stats, err := db.GetSubtaskUsageStats(ctx, subtaskID)

// Get usage for all flows (grouped by flow_id)
allStats, err := db.GetAllFlowsUsageStats(ctx)
```

Each query returns:
```go
type UsageStats struct {
    TotalUsageIn      int64   // Total input tokens
    TotalUsageOut     int64   // Total output tokens
    TotalUsageCacheIn int64   // Total cached input tokens
    TotalUsageCacheOut int64  // Total cached output tokens
    TotalUsageCostIn  float64 // Total input cost
    TotalUsageCostOut float64 // Total output cost
}
```

#### 2. Provider and Model Analytics

Track usage by LLM provider or specific model:

```sql
-- Get usage statistics grouped by provider
-- name: GetUsageStatsByProvider :many
SELECT
  mc.model_provider,
  COALESCE(SUM(mc.usage_in), 0) AS total_usage_in,
  COALESCE(SUM(mc.usage_out), 0) AS total_usage_out,
  COALESCE(SUM(mc.usage_cache_in), 0) AS total_usage_cache_in,
  COALESCE(SUM(mc.usage_cache_out), 0) AS total_usage_cache_out,
  COALESCE(SUM(mc.usage_cost_in), 0.0) AS total_usage_cost_in,
  COALESCE(SUM(mc.usage_cost_out), 0.0) AS total_usage_cost_out
FROM msgchains mc
LEFT JOIN subtasks s ON mc.subtask_id = s.id
LEFT JOIN tasks t ON s.task_id = t.id OR mc.task_id = t.id
INNER JOIN flows f ON (mc.flow_id = f.id OR t.flow_id = f.id)
WHERE f.deleted_at IS NULL
GROUP BY mc.model_provider
ORDER BY mc.model_provider;

-- Get usage statistics grouped by model
-- name: GetUsageStatsByModel :many
-- Similar structure, GROUP BY mc.model, mc.model_provider
```

Usage example:
```go
// Analyze costs per provider
providerStats, err := db.GetUsageStatsByProvider(ctx)
for _, stat := range providerStats {
    totalCost := stat.TotalUsageCostIn + stat.TotalUsageCostOut
    fmt.Printf("Provider: %s, Total Cost: $%.2f\n", 
        stat.ModelProvider, totalCost)
}

// Compare model efficiency
modelStats, err := db.GetUsageStatsByModel(ctx)
```

#### 3. Agent Type Analytics

Track usage by agent type (primary_agent, assistant, pentester, coder, etc.):

```go
// Get usage by type across all flows
typeStats, err := db.GetUsageStatsByType(ctx)

// Get usage by type for a specific flow
flowTypeStats, err := db.GetUsageStatsByTypeForFlow(ctx, flowID)
```

This helps identify which agent types consume the most resources.

#### 4. Temporal Analytics

Analyze usage trends over time:

```go
// Last 7 days
weekStats, err := db.GetUsageStatsByDayLastWeek(ctx)

// Last 30 days
monthStats, err := db.GetUsageStatsByDayLastMonth(ctx)

// Last 90 days
quarterStats, err := db.GetUsageStatsByDayLast3Months(ctx)
```

Each query returns daily aggregates:
```go
type DailyUsageStats struct {
    Date              time.Time
    TotalUsageIn      int64
    TotalUsageOut     int64
    TotalUsageCacheIn int64
    TotalUsageCacheOut int64
    TotalUsageCostIn  float64
    TotalUsageCostOut float64
}
```

### Usage Tracking Implementation

When making LLM API calls, update usage metrics with duration:

```go
// After receiving LLM response
startTime := time.Now()
// ... make LLM API call ...
durationDelta := time.Since(startTime).Seconds()

_, err := db.UpdateMsgChainUsage(ctx, database.UpdateMsgChainUsageParams{
    UsageIn:         response.Usage.PromptTokens,
    UsageOut:        response.Usage.CompletionTokens,
    UsageCacheIn:    response.Usage.PromptCacheTokens,
    UsageCacheOut:   response.Usage.CompletionCacheTokens,
    UsageCostIn:     calculateCost(response.Usage.PromptTokens, inputRate),
    UsageCostOut:    calculateCost(response.Usage.CompletionTokens, outputRate),
    DurationSeconds: durationDelta,
    ID:             msgChainID,
})
```

### Performance Considerations

All analytics queries are optimized with appropriate indexes:

- **Soft delete filtering**: `flows_deleted_at_idx` - partial index for active flows only
- **Time-based queries**: `msgchains_created_at_idx` - for temporal filtering
- **Provider analytics**: `msgchains_model_provider_idx` - for grouping by provider
- **Model analytics**: `msgchains_model_provider_composite_idx` - composite index
- **Type analytics**: `msgchains_type_flow_id_idx` - for flow-scoped type queries

These indexes ensure fast query execution even with millions of message chain records.

### Analytics-Specific Indexes

Additional indexes optimized for analytics queries:

**Assistants Analytics:**
- `assistants_deleted_at_idx` - Partial index for soft delete filtering (WHERE deleted_at IS NULL)
- `assistants_created_at_idx` - Temporal queries and sorting by creation date
- `assistants_flow_id_deleted_at_idx` - Flow-scoped queries with soft delete (GetFlowAssistants)
- `assistants_flow_id_created_at_idx` - Temporal analytics by flow (GetFlowsStatsByDay*)

**Subtasks Analytics:**
- `subtasks_task_id_status_idx` - Task-scoped queries with status filtering
- `subtasks_status_created_at_idx` - Execution time analytics (excludes created/waiting)

**Toolcalls Analytics:**
- `toolcalls_flow_id_status_idx` - Flow-scoped completed toolcalls counting
- `toolcalls_name_status_idx` - Function-based analytics with status filtering

**MsgChains Analytics:**
- `msgchains_type_task_id_subtask_id_idx` - Hierarchical msgchain lookup by type
- `msgchains_type_created_at_idx` - Temporal analytics grouped by msgchain type

**Tasks Analytics:**
- `tasks_flow_id_status_idx` - Flow-scoped task queries with status filtering

### Cost Optimization Strategies

Use analytics data to optimize LLM costs:

1. **Identify expensive flows**: `GetAllFlowsUsageStats()` to find high-cost workflows
2. **Compare providers**: `GetUsageStatsByProvider()` to choose cost-effective providers
3. **Optimize agent types**: `GetUsageStatsByType()` to reduce token usage per agent
4. **Monitor trends**: Temporal queries to detect unusual spikes in usage
5. **Cache effectiveness**: Compare `usage_cache_in` vs `usage_in` to measure prompt caching benefits

Example cost analysis:
```go
// Calculate cache savings
stats, _ := db.GetFlowUsageStats(ctx, flowID)
regularTokens := stats.TotalUsageIn + stats.TotalUsageOut
cachedTokens := stats.TotalUsageCacheIn + stats.TotalUsageCacheOut
cacheRatio := float64(cachedTokens) / float64(regularTokens+cachedTokens)
savings := stats.TotalUsageCostIn * (cacheRatio * 0.9) // Assuming 90% cache discount

fmt.Printf("Cache effectiveness: %.1f%%\n", cacheRatio*100)
fmt.Printf("Estimated savings: $%.2f\n", savings)
```

## Flows and Structure Analytics

The database package provides comprehensive analytics for tracking flow structure, execution metrics, and assistant usage across the workflow hierarchy.

### Flow Structure Queries

#### 1. Flow-Level Statistics

Get structural metrics for specific flows:

```go
// Get structure stats for a flow
stats, err := db.GetFlowStats(ctx, flowID)
// Returns: total_tasks_count, total_subtasks_count, total_assistants_count

// Get total stats for all user's flows
allStats, err := db.GetUserTotalFlowsStats(ctx, userID)
// Returns: total_flows_count, total_tasks_count, total_subtasks_count, total_assistants_count
```

Each query returns:
```go
type FlowStats struct {
    TotalTasksCount      int64
    TotalSubtasksCount   int64
    TotalAssistantsCount int64
}

type FlowsStats struct {
    TotalFlowsCount      int64
    TotalTasksCount      int64
    TotalSubtasksCount   int64
    TotalAssistantsCount int64
}
```

#### 2. Temporal Flow Statistics

Track flow creation and structure over time:

```sql
-- Get flows stats by day for the last week
-- name: GetFlowsStatsByDayLastWeek :many
SELECT
  DATE(f.created_at) AS date,
  COALESCE(COUNT(DISTINCT f.id), 0)::bigint AS total_flows_count,
  COALESCE(COUNT(DISTINCT t.id), 0)::bigint AS total_tasks_count,
  COALESCE(COUNT(DISTINCT s.id), 0)::bigint AS total_subtasks_count,
  COALESCE(COUNT(DISTINCT a.id), 0)::bigint AS total_assistants_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
LEFT JOIN subtasks s ON t.id = s.task_id
LEFT JOIN assistants a ON f.id = a.flow_id AND a.deleted_at IS NULL
WHERE f.created_at >= NOW() - INTERVAL '7 days' 
  AND f.deleted_at IS NULL AND f.user_id = $1
GROUP BY DATE(f.created_at)
ORDER BY date DESC;
```

Usage example:
```go
// Analyze flow trends
weekStats, err := db.GetFlowsStatsByDayLastWeek(ctx, userID)
for _, stat := range weekStats {
    fmt.Printf("Date: %s, Flows: %d, Tasks: %d, Subtasks: %d, Assistants: %d\n",
        stat.Date, stat.TotalFlowsCount, stat.TotalTasksCount, 
        stat.TotalSubtasksCount, stat.TotalAssistantsCount)
}

// Available for different periods
monthStats, err := db.GetFlowsStatsByDayLastMonth(ctx, userID)
quarterStats, err := db.GetFlowsStatsByDayLast3Months(ctx, userID)
```

### Flow Execution Time Analytics

Track actual execution time and tool usage across the flow hierarchy using pre-calculated duration metrics.

#### Analytics Queries (`analytics.sql`)

```sql
-- name: GetFlowsForPeriodLastWeek :many
-- Get flow IDs created in the last week for analytics
SELECT id, title
FROM flows
WHERE created_at >= NOW() - INTERVAL '7 days' 
  AND deleted_at IS NULL AND user_id = $1
ORDER BY created_at DESC;

-- name: GetTasksForFlow :many
-- Get all tasks for a flow
SELECT id, title, created_at, updated_at
FROM tasks
WHERE flow_id = $1
ORDER BY id ASC;

-- name: GetSubtasksForTasks :many
-- Get all subtasks for multiple tasks
SELECT id, task_id, title, status, created_at, updated_at
FROM subtasks
WHERE task_id = ANY(@task_ids::BIGINT[])
ORDER BY id ASC;

-- name: GetMsgchainsForFlow :many
-- Get all msgchains for a flow (including task and subtask level)
SELECT id, type, flow_id, task_id, subtask_id, duration_seconds, created_at, updated_at
FROM msgchains
WHERE flow_id = $1
ORDER BY created_at ASC;

-- name: GetToolcallsForFlow :many
-- Get all toolcalls for a flow
SELECT tc.id, tc.status, tc.flow_id, tc.task_id, tc.subtask_id, 
       tc.duration_seconds, tc.created_at, tc.updated_at
FROM toolcalls tc
LEFT JOIN tasks t ON tc.task_id = t.id
LEFT JOIN subtasks s ON tc.subtask_id = s.id
INNER JOIN flows f ON tc.flow_id = f.id
WHERE tc.flow_id = $1 AND f.deleted_at IS NULL
  AND (tc.task_id IS NULL OR t.id IS NOT NULL)
  AND (tc.subtask_id IS NULL OR s.id IS NOT NULL)
ORDER BY tc.created_at ASC;

-- name: GetAssistantsCountForFlow :one
-- Get total count of assistants for a specific flow
SELECT COALESCE(COUNT(id), 0)::bigint AS total_assistants_count
FROM assistants
WHERE flow_id = $1 AND deleted_at IS NULL;
```

Usage example:
```go
// Get execution statistics for flows in a period
flows, _ := db.GetFlowsForPeriodLastWeek(ctx, userID)

for _, flow := range flows {
    // Get hierarchical data
    tasks, _ := db.GetTasksForFlow(ctx, flow.ID)
    
    // Collect task IDs
    taskIDs := make([]int64, len(tasks))
    for i, task := range tasks {
        taskIDs[i] = task.ID
    }
    
    // Get all subtasks for these tasks
    subtasks, _ := db.GetSubtasksForTasks(ctx, taskIDs)
    
    // Get msgchains and toolcalls
    msgchains, _ := db.GetMsgchainsForFlow(ctx, flow.ID)
    toolcalls, _ := db.GetToolcallsForFlow(ctx, flow.ID)
    
    // Get assistants count
    assistantsCount, _ := db.GetAssistantsCountForFlow(ctx, flow.ID)
    
    // Build execution stats using converter functions
    stats := converter.BuildFlowExecutionStats(
        flow.ID, flow.Title, tasks, subtasks, msgchains, toolcalls, 
        int(assistantsCount),
    )
    
    fmt.Printf("Flow: %s, Duration: %.2fs, Toolcalls: %d, Assistants: %d\n",
        stats.FlowTitle, stats.TotalDurationSeconds, 
        stats.TotalToolcallsCount, stats.TotalAssistantsCount)
}
```

### Assistant Usage Tracking

The database tracks assistant usage across flows:

```go
// Get assistant count for a flow
count, err := db.GetAssistantsCountForFlow(ctx, flowID)

// Get all assistants for a flow
assistants, err := db.GetFlowAssistants(ctx, flowID)

// User-scoped assistant access
userAssistants, err := db.GetUserFlowAssistants(ctx, database.GetUserFlowAssistantsParams{
    FlowID: flowID,
    UserID: userID,
})
```

Assistant metrics help understand:
- **Interactive flow usage**: Flows with high assistant counts indicate heavy user interaction
- **Delegation patterns**: Assistants with `use_agents` flag show delegation behavior
- **Resource allocation**: Track assistant-to-flow ratio for capacity planning

## Usage Patterns

### Basic Query Operations

```go
// Initialize queries
db := database.New(sqlConnection)

// Create a new flow
flow, err := db.CreateFlow(ctx, database.CreateFlowParams{
    Title:              "Security Assessment",
    Status:             "active",
    Model:              "gpt-4",
    ModelProviderName:  "my-openai",
    ModelProviderType:  "openai",
    Language:           "en",
    ToolCallIDTemplate: "call_{r:24:x}",
    Functions:          []byte(`{"tools": ["nmap", "sqlmap"]}`),
    UserID:             userID,
})

// Retrieve user's flows
flows, err := db.GetUserFlows(ctx, userID)

// Update flow status
updatedFlow, err := db.UpdateFlowStatus(ctx, database.UpdateFlowStatusParams{
    Status: "completed",
    ID:     flowID,
})
```

### Transaction Support

```go
tx, err := sqlDB.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

queries := db.WithTx(tx)

// Perform multiple operations atomically
task, err := queries.CreateTask(ctx, taskParams)
if err != nil {
    return err
}

subtask, err := queries.CreateSubtask(ctx, subtaskParams)
if err != nil {
    return err
}

return tx.Commit()
```

### User-Scoped Operations

Most queries include user-scoped variants for multi-tenancy:

```go
// Admin access - all flows
allFlows, err := db.GetFlows(ctx)

// User access - only user's flows
userFlows, err := db.GetUserFlows(ctx, userID)

// User-scoped flow access with validation
flow, err := db.GetUserFlow(ctx, database.GetUserFlowParams{
    ID:     flowID,
    UserID: userID,
})
```

## Integration with PentAGI

### GraphQL API Integration

The database package integrates with PentAGI's GraphQL API through the converter package:

```go
// In GraphQL resolvers
func (r *queryResolver) Flows(ctx context.Context) ([]*model.Flow, error) {
    userID := auth.GetUserID(ctx)

    // Fetch from database
    flows, err := r.DB.GetUserFlows(ctx, userID)
    if err != nil {
        return nil, err
    }

    containers, err := r.DB.GetUserContainers(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Convert to GraphQL models
    return converter.ConvertFlows(flows, containers), nil
}
```

### AI Agent Integration

The package supports AI agent operations through specialized queries:

```go
// Log agent interactions
agentLog, err := db.CreateAgentLog(ctx, database.CreateAgentLogParams{
    Initiator: "pentester",
    Executor:  "researcher",
    Task:      "Analyze target application",
    Result:    resultJSON,
    FlowID:    flowID,
    TaskID:    sql.NullInt64{Int64: taskID, Valid: true},
})

// Track tool calls with duration updates
toolCall, err := db.CreateToolcall(ctx, database.CreateToolcallParams{
    CallID:    callID,
    Status:    "received",
    Name:      "nmap_scan",
    Args:      argsJSON,
    FlowID:    flowID,
    TaskID:    sql.NullInt64{Int64: taskID, Valid: true},
    SubtaskID: sql.NullInt64{Int64: subtaskID, Valid: true},
})

// Update status with duration delta
startTime := time.Now()
// ... execute toolcall ...
durationDelta := time.Since(startTime).Seconds()

_, err = db.UpdateToolcallFinishedResult(ctx, database.UpdateToolcallFinishedResultParams{
    Result:          resultJSON,
    DurationSeconds: durationDelta,
    ID:              toolCall.ID,
})
```

### Vector Database Operations

For AI memory and semantic search:

```go
// Log vector operations
vecLog, err := db.CreateVectorStoreLog(ctx, database.CreateVectorStoreLogParams{
    Initiator: "memorist",
    Executor:  "vector_db",
    Filter:    "vulnerability_data",
    Query:     "SQL injection techniques",
    Action:    "search",
    Result:    resultsJSON,
    FlowID:    flowID,
})
```

## Best Practices

### Error Handling

Always handle database errors appropriately:

```go
flow, err := db.GetUserFlow(ctx, params)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("flow not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}
```

### Context Usage

Use context for timeout and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

flows, err := db.GetFlows(ctx)
```

### Null Value Handling

Use provided utilities for null values:

```go
// Converting optional strings
description := database.StringToNullString(optionalDesc)

// Converting back to pointers
descPtr := database.NullStringToPtrString(task.Description)
```

## Security Considerations

### Multi-tenancy

All user-facing operations use user-scoped queries to prevent unauthorized access:

- `GetUserFlows()` instead of `GetFlows()`
- `GetUserFlowTasks()` instead of `GetFlowTasks()`
- User ID validation in all operations

### Soft Deletes

Critical entities use soft deletes to maintain audit trails:

```sql
-- Flows and assistants are soft deleted
UPDATE flows SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1

-- Most queries automatically filter soft-deleted records
WHERE f.deleted_at IS NULL
```

### SQL Injection Prevention

sqlc generates parameterized queries that prevent SQL injection:

```sql
-- Safe parameterized query
SELECT * FROM flows WHERE user_id = $1 AND id = $2
```

## Performance Considerations

### Query Optimization

The database package is designed with performance in mind:

**Indexed Queries**: All foreign key relationships and frequently queried fields are properly indexed:
```sql
-- Primary keys and foreign keys are automatically indexed
-- Common query patterns use indexes for filtering and grouping

-- Flow indexes
CREATE INDEX flows_status_idx ON flows(status);
CREATE INDEX flows_title_idx ON flows(title);
CREATE INDEX flows_language_idx ON flows(language);
CREATE INDEX flows_model_provider_name_idx ON flows(model_provider_name);
CREATE INDEX flows_model_provider_type_idx ON flows(model_provider_type);
CREATE INDEX flows_user_id_idx ON flows(user_id);
CREATE INDEX flows_trace_id_idx ON flows(trace_id);
CREATE INDEX flows_deleted_at_idx ON flows(deleted_at) WHERE deleted_at IS NULL;

-- Task indexes  
CREATE INDEX tasks_status_idx ON tasks(status);
CREATE INDEX tasks_title_idx ON tasks(title);
CREATE INDEX tasks_flow_id_idx ON tasks(flow_id);

-- Subtask indexes
CREATE INDEX subtasks_status_idx ON subtasks(status);
CREATE INDEX subtasks_title_idx ON subtasks(title);
CREATE INDEX subtasks_task_id_idx ON subtasks(task_id);

-- MsgChain indexes for analytics and duration tracking
CREATE INDEX msgchains_type_idx ON msgchains(type);
CREATE INDEX msgchains_flow_id_idx ON msgchains(flow_id);
CREATE INDEX msgchains_task_id_idx ON msgchains(task_id);
CREATE INDEX msgchains_subtask_id_idx ON msgchains(subtask_id);
CREATE INDEX msgchains_created_at_idx ON msgchains(created_at);
CREATE INDEX msgchains_model_provider_idx ON msgchains(model_provider);
CREATE INDEX msgchains_model_idx ON msgchains(model);
CREATE INDEX msgchains_model_provider_composite_idx ON msgchains(model, model_provider);
CREATE INDEX msgchains_created_at_flow_id_idx ON msgchains(created_at, flow_id);
CREATE INDEX msgchains_type_flow_id_idx ON msgchains(type, flow_id);

-- Toolcalls indexes for analytics and duration tracking
CREATE INDEX toolcalls_flow_id_idx ON toolcalls(flow_id);
CREATE INDEX toolcalls_task_id_idx ON toolcalls(task_id);
CREATE INDEX toolcalls_subtask_id_idx ON toolcalls(subtask_id);
CREATE INDEX toolcalls_status_idx ON toolcalls(status);
CREATE INDEX toolcalls_name_idx ON toolcalls(name);
CREATE INDEX toolcalls_created_at_idx ON toolcalls(created_at);
CREATE INDEX toolcalls_call_id_idx ON toolcalls(call_id);

-- Assistants indexes for analytics
CREATE INDEX assistants_flow_id_idx ON assistants(flow_id);
CREATE INDEX assistants_deleted_at_idx ON assistants(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX assistants_created_at_idx ON assistants(created_at);
CREATE INDEX assistants_flow_id_deleted_at_idx ON assistants(flow_id, deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX assistants_flow_id_created_at_idx ON assistants(flow_id, created_at) WHERE deleted_at IS NULL;

-- Additional analytics indexes
CREATE INDEX subtasks_task_id_status_idx ON subtasks(task_id, status);
CREATE INDEX subtasks_status_created_at_idx ON subtasks(status, created_at);
CREATE INDEX toolcalls_flow_id_status_idx ON toolcalls(flow_id, status);
CREATE INDEX toolcalls_name_status_idx ON toolcalls(name, status);
CREATE INDEX msgchains_type_task_id_subtask_id_idx ON msgchains(type, task_id, subtask_id);
CREATE INDEX msgchains_type_created_at_idx ON msgchains(type, created_at);
CREATE INDEX tasks_flow_id_status_idx ON tasks(flow_id, status);

-- Provider indexes
CREATE INDEX providers_user_id_idx ON providers(user_id);
CREATE INDEX providers_type_idx ON providers(type);
CREATE INDEX providers_name_user_id_idx ON providers(name, user_id);
CREATE UNIQUE INDEX providers_name_user_id_unique ON providers(name, user_id) WHERE deleted_at IS NULL;
```

**Note**: Some indexes on large text fields (tasks.input, tasks.result, subtasks.description, subtasks.result) have been removed to improve write performance. These fields should use full-text search when needed.

**Efficient Joins**: User-scoped queries use INNER JOINs to leverage PostgreSQL query planner:
```sql
-- Efficient user-scoped access with proper join order
SELECT t.*
FROM tasks t
INNER JOIN flows f ON t.flow_id = f.id  -- Fast foreign key join
WHERE f.user_id = $1 AND f.deleted_at IS NULL;
```

**Batch Operations**: Use transaction batching for bulk operations:
```go
tx, err := db.BeginTx(ctx, nil)
defer tx.Rollback()

queries := database.New(tx)
for _, item := range items {
    if _, err := queries.CreateSubtask(ctx, item); err != nil {
        return err
    }
}
return tx.Commit()
```

### Connection Pooling

The package provides optimized connection pooling through GORM:
```go
func NewGorm(dsn, dbType string) (*gorm.DB, error) {
    db, err := gorm.Open(dbType, dsn)
    if err != nil {
        return nil, err
    }

    // Optimized connection settings
    db.DB().SetMaxIdleConns(5)
    db.DB().SetMaxOpenConns(20)
    db.DB().SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

### Vector Operations

For pgvector operations, consider:
- **Batch embedding inserts** for better performance
- **Appropriate vector dimensions** (typically 512-1536)
- **Index configuration** for similarity searches

## Debugging and Troubleshooting

### Query Logging

Enable query logging for debugging:
```go
// GORM logger captures all SQL operations
db.SetLogger(&GormLogger{})
db.LogMode(true)
```

**Log Output Example**:
```
INFO[0000] SELECT * FROM flows WHERE user_id = '1' AND deleted_at IS NULL  component=pentagi-gorm duration=2.5ms rows_returned=3
```

### Common Issues and Solutions

#### 1. Foreign Key Constraint Violations

**Error**: `pq: insert or update on table "tasks" violates foreign key constraint`

**Solution**: Ensure parent entities exist before creating child entities:
```go
// Verify flow exists and user has access
flow, err := db.GetUserFlow(ctx, database.GetUserFlowParams{
    ID:     flowID,
    UserID: userID,
})
if err != nil {
    return fmt.Errorf("invalid flow: %w", err)
}

// Now safe to create task
task, err := db.CreateTask(ctx, taskParams)
```

#### 2. Soft Delete Issues

**Error**: Records not appearing in queries after "deletion"

**Solution**: Check soft delete filters in custom queries:
```sql
-- Always include soft delete filter
WHERE f.deleted_at IS NULL
```

#### 3. Null Value Handling

**Error**: `sql: Scan error on column index 2: unsupported Scan`

**Solution**: Use proper null value converters:
```go
// When creating
description := database.StringToNullString(optionalDesc)

// When reading
descPtr := database.NullStringToPtrString(row.Description)
```

### Query Performance Analysis

Use PostgreSQL's EXPLAIN for performance analysis:
```sql
-- Analyze query performance
EXPLAIN ANALYZE SELECT f.*, COUNT(t.id) as task_count
FROM flows f
LEFT JOIN tasks t ON f.id = t.flow_id
WHERE f.user_id = $1 AND f.deleted_at IS NULL
GROUP BY f.id;
```

## Extending the Database Package

### Adding New Entities

1. **Create migration**: Add schema in `backend/migrations/sql/`
2. **Create SQL queries**: Add `.sql` file in `backend/sqlc/models/`
3. **Regenerate code**: Run sqlc generation command
4. **Add converters**: Update `converter/converter.go` for GraphQL integration

**Example New Entity**:
```sql
-- backend/sqlc/models/vulnerabilities.sql

-- name: CreateVulnerability :one
INSERT INTO vulnerabilities (
  title, severity, description, flow_id
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetFlowVulnerabilities :many
SELECT v.*
FROM vulnerabilities v
INNER JOIN flows f ON v.flow_id = f.id
WHERE v.flow_id = $1 AND f.deleted_at IS NULL
ORDER BY v.severity DESC, v.created_at DESC;
```

### Custom Query Patterns

Follow established patterns for consistency:

```sql
-- Pattern: User-scoped access
-- name: GetUser[Entity] :one/:many
SELECT [entity].*
FROM [entity] [alias]
INNER JOIN flows f ON [alias].flow_id = f.id
INNER JOIN users u ON f.user_id = u.id
WHERE [conditions] AND f.user_id = $user_id AND f.deleted_at IS NULL;

-- Pattern: Hierarchical retrieval
-- name: Get[Parent][Children] :many
SELECT [child].*
FROM [child] [child_alias]
INNER JOIN [parent] [parent_alias] ON [child_alias].[parent_id] = [parent_alias].id
WHERE [parent_alias].id = $1 AND [filters];
```

### Integration Testing

Test database operations with real PostgreSQL:
```go
func TestCreateFlow(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    queries := database.New(db)

    // Test operation
    flow, err := queries.CreateFlow(ctx, database.CreateFlowParams{
        Title:         "Test Flow",
        Status:        "active",
        ModelProvider: "openai",
        UserID:        1,
    })

    assert.NoError(t, err)
    assert.Equal(t, "Test Flow", flow.Title)
}
```

## Security Guidelines

### Input Validation

Always validate inputs before database operations:
```go
func validateFlowInput(params CreateFlowParams) error {
    if len(params.Title) > 255 {
        return fmt.Errorf("title too long")
    }
    if !isValidStatus(params.Status) {
        return fmt.Errorf("invalid status")
    }
    return nil
}
```

### Access Control

Implement consistent access control patterns:
```go
// Always verify user ownership
flow, err := db.GetUserFlow(ctx, database.GetUserFlowParams{
    ID:     flowID,
    UserID: currentUserID,
})
if err != nil {
    return fmt.Errorf("access denied or flow not found")
}
```

### Audit Logging

Use logging entities for security audit trails:
```go
// Log sensitive operations
_, err = db.CreateAgentLog(ctx, database.CreateAgentLogParams{
    Initiator: "system",
    Executor:  "user_action",
    Task:      "flow_deletion",
    Result:    []byte(fmt.Sprintf(`{"flow_id": %d, "user_id": %d}`, flowID, userID)),
    FlowID:    flowID,
})
```

## Conclusion

The database package provides a robust, secure, and performant foundation for PentAGI's data layer. By leveraging sqlc for code generation, implementing consistent security patterns, and maintaining comprehensive audit trails, it ensures reliable operation of the autonomous penetration testing system.

Key benefits:
- **Type Safety**: Compile-time verification of SQL queries
- **Performance**: Optimized queries with proper indexing
- **Security**: Multi-tenancy and soft delete support
- **Observability**: Comprehensive logging and tracing
- **Maintainability**: Consistent patterns and generated code

For developers working with this package, follow the established patterns for security, performance, and maintainability to ensure smooth integration with the broader PentAGI ecosystem.

This documentation provides a comprehensive overview of the database package's architecture, functionality, and integration within the PentAGI system.
