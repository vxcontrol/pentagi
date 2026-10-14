-- +goose Up
CREATE TABLE flow_input_updates (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  flow_id BIGINT NOT NULL REFERENCES flows(id) ON DELETE CASCADE,
  task_id BIGINT REFERENCES tasks(id) ON DELETE SET NULL,
  input TEXT NOT NULL CHECK (length(btrim(input)) > 0),
  status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'completed', 'failed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX flow_input_updates_pending_idx ON flow_input_updates (flow_id, id) WHERE status = 'queued';
CREATE UNIQUE INDEX flow_input_updates_task_idx ON flow_input_updates (task_id) WHERE task_id IS NOT NULL;

-- +goose Down
DROP TABLE flow_input_updates;
