-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: CountFlows :one
SELECT count(*) FROM flows WHERE deleted_at IS NULL;

-- name: CountActiveFlows :one
SELECT count(*) FROM flows WHERE deleted_at IS NULL AND status IN ('created', 'running', 'waiting');

-- name: CountTasks :one
SELECT count(*) FROM tasks;

-- name: CountSubtasks :one
SELECT count(*) FROM subtasks;

-- name: CountContainers :one
SELECT count(*) FROM containers;

-- name: CountAssistants :one
SELECT count(*) FROM assistants WHERE deleted_at IS NULL;

-- name: CountFlowTemplates :one
SELECT count(*) FROM flow_templates;

-- name: CountPrompts :one
SELECT count(*) FROM prompts;

-- name: CountAPITokens :one
SELECT count(*) FROM api_tokens;

-- name: CountToolcalls :one
SELECT count(*) FROM toolcalls;

-- name: GetSummaryProviderCounts :many
-- Number of configured providers per type, for the instance summary.
SELECT type, count(*) AS total
FROM providers
WHERE deleted_at IS NULL
GROUP BY type
ORDER BY type;

-- name: GetSummaryUsageByProvider :many
-- Aggregate LLM usage per provider type, over the whole life of the instance,
-- for the instance summary.
SELECT
  model_provider,
  count(DISTINCT flow_id) AS flows,
  count(*) AS chains,
  COALESCE(SUM(usage_in), 0)::bigint AS tokens_in,
  COALESCE(SUM(usage_out), 0)::bigint AS tokens_out,
  COALESCE(SUM(usage_cache_in), 0)::bigint AS cache_in,
  COALESCE(SUM(usage_cache_out), 0)::bigint AS cache_out,
  COALESCE(SUM(usage_cost_in), 0.0)::double precision AS cost_in,
  COALESCE(SUM(usage_cost_out), 0.0)::double precision AS cost_out
FROM msgchains
GROUP BY model_provider
ORDER BY model_provider;
