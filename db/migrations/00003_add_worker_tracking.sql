-- +goose Up
-- +goose StatementBegin
ALTER TABLE jobs ADD COLUMN claimed_by TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE jobs DROP COLUMN claimed_by;
-- +goose StatementEnd
