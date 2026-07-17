-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
ALTER COLUMN accrual TYPE NUMERIC(10, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
ALTER COLUMN accrual TYPE INTEGER;
-- +goose StatementEnd
