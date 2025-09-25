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

| File | Entity | Purpose |
|------|--------|---------|
| `flows.sql` | Flow | Top-level workflow management |
| `tasks.sql` | Task | Task lifecycle and status tracking |
| `subtasks.sql` | SubTask | Agent assignment and execution |
| `assistants.sql` | Assistant | AI assistant management |
| `containers.sql` | Container | Docker environment tracking |
| `users.sql` | User | User management and authentication |
| `roles.sql` | Role | Role-based access control |
| `prompts.sql` | Prompt | User-defined prompt templates |
| `msgchains.sql` | MsgChain | LLM conversation chains |
| `toolcalls.sql` | ToolCall | AI function call tracking |
| `screenshots.sql` | Screenshot | Visual artifacts storage |
| **Logging Entities** | | |
| `agentlogs.sql` | AgentLog | Inter-agent communication |
| `assistantlogs.sql` | AssistantLog | Human-assistant interactions |
| `msglogs.sql` | MsgLog | General message logging |
| `searchlogs.sql` | SearchLog | External search operations |
| `termlogs.sql` | TermLog | Terminal command execution |
| `vecstorelogs.sql` | VecStoreLog | Vector database operations |

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

-- Update conversation usage tracking
-- name: UpdateMsgChainUsage :one
UPDATE msgchains
SET usage_in = usage_in + $1, usage_out = usage_out + $2
WHERE id = $3
RETURNING *;
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
- `model`, `model_provider`, `language` for AI configuration
- `functions`, `prompts` as JSON for AI behavior
- `trace_id` for observability
- `user_id` for multi-tenancy
- Soft delete with `deleted_at`

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
- `title`, `status`, `model`, `model_provider`
- `language`, `functions`, `prompts` configuration
- `use_agents` flag for delegation behavior
- Flow association and soft delete

#### Message Chains (MsgChain)
LLM conversation management:
- `type` (primary_agent/assistant/etc.)
- `model`, `model_provider` for tracking
- `usage_in`, `usage_out` for token counting
- `chain` JSON for conversation history
- Multi-level association (flow/task/subtask)

### Logging Entities

The package provides comprehensive logging for all system operations:

- **AgentLog**: Inter-agent communication and delegation
- **AssistantLog**: Human-assistant interactions
- **MsgLog**: General message logging
- **SearchLog**: External search operations
- **TermLog**: Terminal command execution
- **ToolCall**: AI function calling
- **VecStoreLog**: Vector database operations

## Usage Patterns

### Basic Query Operations

```go
// Initialize queries
db := database.New(sqlConnection)

// Create a new flow
flow, err := db.CreateFlow(ctx, database.CreateFlowParams{
    Title:         "Security Assessment",
    Status:        "active",
    Model:         "gpt-4",
    ModelProvider: "openai",
    Language:      "en",
    Functions:     []byte(`{"tools": ["nmap", "sqlmap"]}`),
    Prompts:       []byte(`{"system": "You are a security expert"}`),
    UserID:        userID,
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

// Track tool calls
toolCall, err := db.CreateToolcall(ctx, database.CreateToolcallParams{
    CallID:    callID,
    Status:    "running",
    Name:      "nmap_scan",
    Args:      argsJSON,
    FlowID:    flowID,
    TaskID:    sql.NullInt64{Int64: taskID, Valid: true},
    SubtaskID: sql.NullInt64{Int64: subtaskID, Valid: true},
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
-- Common query patterns use composite indexes
CREATE INDEX idx_flows_user_status ON flows(user_id, status);
CREATE INDEX idx_tasks_flow_status ON tasks(flow_id, status);
```

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
