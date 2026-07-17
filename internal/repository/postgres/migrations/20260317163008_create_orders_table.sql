-- +goose Up
-- +goose StatementBegin
 CREATE TYPE order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED'
);
create table if not exists orders (
    number VARCHAR(1024) PRIMARY KEY,
    status order_status NOT NULL DEFAULT 'NEW',
    accrual INTEGER DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id INTEGER REFERENCES users(id) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists orders;
drop type order_status;
-- +goose StatementEnd
