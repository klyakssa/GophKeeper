-- +goose Up
CREATE INDEX idx_secure_data_user_id ON secure_data(user_id);

-- +goose Down
DROP INDEX idx_secure_data_user_id;
