-- +goose Up
-- +goose StatementBegin
ALTER TABLE jobs ADD COLUMN dlq_reason TEXT, ADD COLUMN failed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE jobs DROP COLUMN IF EXISTS dlq_reason, DROP COLUMN IF EXISTS failed_at;
-- +goose StatementEnd
