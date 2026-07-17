-- +goose Up
-- +goose StatementBegin
create table if not exists withdraw_history (
    number VARCHAR(1024) primary key,
    sum NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists withdraw_history;
-- +goose StatementEnd
