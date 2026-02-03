-- +goose Up
-- +goose StatementBegin
-- Add usage tracking fields to msgchains
ALTER TABLE msgchains ADD COLUMN usage_cache_in BIGINT NOT NULL DEFAULT 0;
ALTER TABLE msgchains ADD COLUMN usage_cache_out BIGINT NOT NULL DEFAULT 0;
ALTER TABLE msgchains ADD COLUMN usage_cost_in DOUBLE PRECISION NOT NULL DEFAULT 0.0;
ALTER TABLE msgchains ADD COLUMN usage_cost_out DOUBLE PRECISION NOT NULL DEFAULT 0.0;

-- Add duration tracking to msgchains (nullable first)
ALTER TABLE msgchains ADD COLUMN duration_seconds DOUBLE PRECISION NULL;

-- Calculate duration for existing msgchains records
UPDATE msgchains
SET duration_seconds = EXTRACT(EPOCH FROM (updated_at - created_at))
WHERE updated_at > created_at;

-- Set remaining NULL values to 0.0
UPDATE msgchains
SET duration_seconds = 0.0
WHERE duration_seconds IS NULL;

-- Make column NOT NULL with default
ALTER TABLE msgchains ALTER COLUMN duration_seconds SET NOT NULL;
ALTER TABLE msgchains ALTER COLUMN duration_seconds SET DEFAULT 0.0;

-- Add duration tracking to toolcalls (nullable first)
ALTER TABLE toolcalls ADD COLUMN duration_seconds DOUBLE PRECISION NULL;

-- Calculate duration for existing toolcalls records (finished and failed only)
UPDATE toolcalls
SET duration_seconds = EXTRACT(EPOCH FROM (updated_at - created_at))
WHERE updated_at > created_at AND status IN ('finished', 'failed');

-- Set remaining NULL values to 0.0
UPDATE toolcalls
SET duration_seconds = 0.0
WHERE duration_seconds IS NULL;

-- Make column NOT NULL with default
ALTER TABLE toolcalls ALTER COLUMN duration_seconds SET NOT NULL;
ALTER TABLE toolcalls ALTER COLUMN duration_seconds SET DEFAULT 0.0;

-- Add task and subtask references to termlogs for better hierarchical tracking
ALTER TABLE termlogs ADD COLUMN flow_id BIGINT NULL REFERENCES flows(id) ON DELETE CASCADE;
ALTER TABLE termlogs ADD COLUMN task_id BIGINT NULL REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE termlogs ADD COLUMN subtask_id BIGINT NULL REFERENCES subtasks(id) ON DELETE CASCADE;

-- Fill flow_id from related containers
UPDATE termlogs tl
SET flow_id = c.flow_id
FROM containers c
WHERE tl.container_id = c.id AND tl.flow_id IS NULL;

-- For any remaining NULL flow_id (shouldn't happen due to CASCADE, but just in case)
-- Fill with the first available flow_id
UPDATE termlogs
SET flow_id = (SELECT id FROM flows ORDER BY id LIMIT 1)
WHERE flow_id IS NULL;

-- Delete orphaned records if any still have NULL flow_id (no flows exist)
-- This shouldn't happen in practice due to CASCADE DELETE
DELETE FROM termlogs WHERE flow_id IS NULL;

-- Now make flow_id NOT NULL
ALTER TABLE termlogs ALTER COLUMN flow_id SET NOT NULL;

-- Add task and subtask references to screenshots for better hierarchical tracking
-- Note: flow_id already exists as NOT NULL in screenshots table
ALTER TABLE screenshots ADD COLUMN task_id BIGINT NULL REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE screenshots ADD COLUMN subtask_id BIGINT NULL REFERENCES subtasks(id) ON DELETE CASCADE;

-- Create indexes for termlogs foreign keys
CREATE INDEX termlogs_flow_id_idx ON termlogs(flow_id);
CREATE INDEX termlogs_task_id_idx ON termlogs(task_id);
CREATE INDEX termlogs_subtask_id_idx ON termlogs(subtask_id);

-- Create indexes for screenshots foreign keys
CREATE INDEX screenshots_task_id_idx ON screenshots(task_id);
CREATE INDEX screenshots_subtask_id_idx ON screenshots(subtask_id);

-- Index for soft delete filtering on flows (used in all analytics queries)
-- Using partial index because we mostly query non-deleted flows
CREATE INDEX flows_deleted_at_idx ON flows(deleted_at) WHERE deleted_at IS NULL;

-- Index for time-based analytics queries
CREATE INDEX msgchains_created_at_idx ON msgchains(created_at);

-- Index for grouping by model provider
CREATE INDEX msgchains_model_provider_idx ON msgchains(model_provider);

-- Index for grouping by model
CREATE INDEX msgchains_model_idx ON msgchains(model);

-- Composite index for queries that group by both model and provider
CREATE INDEX msgchains_model_provider_composite_idx ON msgchains(model, model_provider);

-- Composite index for time-based queries with flow filtering
-- This helps queries that filter by created_at AND join with flows
CREATE INDEX msgchains_created_at_flow_id_idx ON msgchains(created_at, flow_id);

-- Composite index for type-based analytics with flow filtering
CREATE INDEX msgchains_type_flow_id_idx ON msgchains(type, flow_id);

-- ==================== Toolcalls Analytics Indexes ====================

-- Index for time-based toolcalls analytics queries
CREATE INDEX toolcalls_created_at_idx ON toolcalls(created_at);

-- Index for updated_at to help with duration calculations
CREATE INDEX toolcalls_updated_at_idx ON toolcalls(updated_at);

-- Composite index for time-based queries with flow filtering
CREATE INDEX toolcalls_created_at_flow_id_idx ON toolcalls(created_at, flow_id);

-- Composite index for function-based analytics with flow filtering
CREATE INDEX toolcalls_name_flow_id_idx ON toolcalls(name, flow_id);

-- Composite index for status and timestamps (for duration calculations)
CREATE INDEX toolcalls_status_updated_at_idx ON toolcalls(status, updated_at);

-- ==================== Flows Analytics Indexes ====================

-- Index for time-based flows analytics queries
CREATE INDEX flows_created_at_idx ON flows(created_at) WHERE deleted_at IS NULL;

-- Index for tasks time-based analytics
CREATE INDEX tasks_created_at_idx ON tasks(created_at);

-- Index for subtasks time-based analytics  
CREATE INDEX subtasks_created_at_idx ON subtasks(created_at);

-- Composite index for tasks with flow filtering
CREATE INDEX tasks_flow_id_created_at_idx ON tasks(flow_id, created_at);

-- Composite index for subtasks with task filtering
CREATE INDEX subtasks_task_id_created_at_idx ON subtasks(task_id, created_at);

-- Add usage privileges
INSERT INTO privileges (role_id, name) VALUES
    (1, 'usage.admin'),
    (1, 'usage.view'),
    (2, 'usage.view')
    ON CONFLICT DO NOTHING;

-- ==================== Assistants Analytics Indexes ====================

-- Partial index for soft delete filtering (used in almost all assistants queries)
CREATE INDEX assistants_deleted_at_idx ON assistants(deleted_at) WHERE deleted_at IS NULL;

-- Index for time-based queries and sorting
CREATE INDEX assistants_created_at_idx ON assistants(created_at);

-- Composite index for flow-scoped queries with soft delete filter
-- Optimizes: SELECT ... FROM assistants WHERE flow_id = $1 AND deleted_at IS NULL
CREATE INDEX assistants_flow_id_deleted_at_idx ON assistants(flow_id, deleted_at) WHERE deleted_at IS NULL;

-- Composite index for temporal analytics queries
-- Optimizes: GetFlowsStatsByDay* queries that join assistants with DATE(created_at) condition
CREATE INDEX assistants_flow_id_created_at_idx ON assistants(flow_id, created_at) WHERE deleted_at IS NULL;

-- ==================== Additional Analytics Indexes ====================

-- Composite index for subtasks filtering by task and status
-- Optimizes: GetTaskPlannedSubtasks, GetTaskCompletedSubtasks, analytics calculations
CREATE INDEX subtasks_task_id_status_idx ON subtasks(task_id, status);

-- Composite index for toolcalls filtering by flow and status
-- Optimizes: Analytics queries counting finished/failed toolcalls per flow
CREATE INDEX toolcalls_flow_id_status_idx ON toolcalls(flow_id, status);

-- Composite index for msgchains type-based analytics with hierarchy
-- Optimizes: Queries searching for specific msgchain types at task/subtask level
CREATE INDEX msgchains_type_task_id_subtask_id_idx ON msgchains(type, task_id, subtask_id);

-- Composite index for tasks with flow and status filtering
-- Optimizes: Flow-scoped task queries with status filtering
CREATE INDEX tasks_flow_id_status_idx ON tasks(flow_id, status);

-- Composite index for subtasks with status filtering (extended version)
-- Optimizes: Subtask analytics excluding created/waiting subtasks
CREATE INDEX subtasks_status_created_at_idx ON subtasks(status, created_at);

-- Composite index for toolcalls analytics by name and status
-- Optimizes: GetToolcallsStatsByFunction queries (filtering by status)
CREATE INDEX toolcalls_name_status_idx ON toolcalls(name, status);

-- Composite index for msgchains analytics by type and created_at
-- Optimizes: Time-based analytics grouped by msgchain type
CREATE INDEX msgchains_type_created_at_idx ON msgchains(type, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop termlogs indexes and columns
DROP INDEX IF EXISTS termlogs_flow_id_idx;
DROP INDEX IF EXISTS termlogs_task_id_idx;
DROP INDEX IF EXISTS termlogs_subtask_id_idx;
ALTER TABLE termlogs DROP COLUMN flow_id;
ALTER TABLE termlogs DROP COLUMN task_id;
ALTER TABLE termlogs DROP COLUMN subtask_id;

-- Drop screenshots indexes and columns
DROP INDEX IF EXISTS screenshots_task_id_idx;
DROP INDEX IF EXISTS screenshots_subtask_id_idx;
ALTER TABLE screenshots DROP COLUMN task_id;
ALTER TABLE screenshots DROP COLUMN subtask_id;

-- Drop msgchains usage tracking columns
ALTER TABLE msgchains DROP COLUMN usage_cache_in;
ALTER TABLE msgchains DROP COLUMN usage_cache_out;
ALTER TABLE msgchains DROP COLUMN usage_cost_in;
ALTER TABLE msgchains DROP COLUMN usage_cost_out;
ALTER TABLE msgchains DROP COLUMN duration_seconds;

-- Drop toolcalls duration tracking column
ALTER TABLE toolcalls DROP COLUMN duration_seconds;

-- Drop indexes
DROP INDEX IF EXISTS flows_deleted_at_idx;
DROP INDEX IF EXISTS msgchains_created_at_idx;
DROP INDEX IF EXISTS msgchains_model_provider_idx;
DROP INDEX IF EXISTS msgchains_model_idx;
DROP INDEX IF EXISTS msgchains_model_provider_composite_idx;
DROP INDEX IF EXISTS msgchains_created_at_flow_id_idx;
DROP INDEX IF EXISTS msgchains_type_flow_id_idx;

-- Drop toolcalls analytics indexes
DROP INDEX IF EXISTS toolcalls_created_at_idx;
DROP INDEX IF EXISTS toolcalls_updated_at_idx;
DROP INDEX IF EXISTS toolcalls_created_at_flow_id_idx;
DROP INDEX IF EXISTS toolcalls_name_flow_id_idx;
DROP INDEX IF EXISTS toolcalls_status_updated_at_idx;

-- Drop flows analytics indexes
DROP INDEX IF EXISTS flows_created_at_idx;
DROP INDEX IF EXISTS tasks_created_at_idx;
DROP INDEX IF EXISTS subtasks_created_at_idx;
DROP INDEX IF EXISTS tasks_flow_id_created_at_idx;
DROP INDEX IF EXISTS subtasks_task_id_created_at_idx;

-- Drop usage privileges
DELETE FROM privileges WHERE name IN ('usage.admin', 'usage.view');

-- Drop assistants analytics indexes
DROP INDEX IF EXISTS assistants_deleted_at_idx;
DROP INDEX IF EXISTS assistants_created_at_idx;
DROP INDEX IF EXISTS assistants_flow_id_deleted_at_idx;
DROP INDEX IF EXISTS assistants_flow_id_created_at_idx;

-- Drop additional analytics indexes
DROP INDEX IF EXISTS subtasks_task_id_status_idx;
DROP INDEX IF EXISTS toolcalls_flow_id_status_idx;
DROP INDEX IF EXISTS msgchains_type_task_id_subtask_id_idx;
DROP INDEX IF EXISTS tasks_flow_id_status_idx;
DROP INDEX IF EXISTS subtasks_status_created_at_idx;
DROP INDEX IF EXISTS toolcalls_name_status_idx;
DROP INDEX IF EXISTS msgchains_type_created_at_idx;
-- +goose StatementEnd
