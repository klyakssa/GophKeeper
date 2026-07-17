-- +goose Up
-- +goose StatementBegin
create table if not exists balances (
    current INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER REFERENCES users(id) NOT NULL
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists balances;
-- +goose StatementEnd
