-- +goose Up
-- +goose StatementBegin

ALTER TABLE toolcalls     SET (autovacuum_vacuum_scale_factor = 0.05, autovacuum_analyze_scale_factor = 0.05);
ALTER TABLE msglogs       SET (autovacuum_vacuum_scale_factor = 0.05, autovacuum_analyze_scale_factor = 0.05);
ALTER TABLE msgchains     SET (autovacuum_vacuum_scale_factor = 0.05, autovacuum_analyze_scale_factor = 0.05);
ALTER TABLE assistantlogs SET (autovacuum_vacuum_scale_factor = 0.05, autovacuum_analyze_scale_factor = 0.05);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE toolcalls     RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
ALTER TABLE msglogs       RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
ALTER TABLE msgchains     RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);
ALTER TABLE assistantlogs RESET (autovacuum_vacuum_scale_factor, autovacuum_analyze_scale_factor);

-- +goose StatementEnd
