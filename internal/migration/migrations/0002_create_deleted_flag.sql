-- +goose Up
ALTER TABLE urls ADD COLUMN deleted_flag BOOL NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE urls DROP COLUMN deleted_flag;