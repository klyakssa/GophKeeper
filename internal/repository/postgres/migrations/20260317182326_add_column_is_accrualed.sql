-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
ADD COLUMN is_accrualed BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
DROP COLUMN is_accrualed;
-- +goose StatementEnd
