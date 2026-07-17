-- +goose Up
-- +goose StatementBegin
ALTER TABLE balances
ALTER COLUMN current TYPE NUMERIC(10, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE balances
ALTER COLUMN current TYPE INTEGER;
-- +goose StatementEnd
