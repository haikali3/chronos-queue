-- +goose Up
-- +goose StatementBegin
ALTER TABLE "jobs" ADD COLUMN visible_after TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "jobs" DROP COLUMN visible_after;
-- +goose StatementEnd
