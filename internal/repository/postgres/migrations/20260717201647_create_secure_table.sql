-- +goose Up
-- +goose StatementBegin
CREATE TABLE secure_data (
    id BIGSERIAL PRIMARY KEY,
    data_type VARCHAR(50) NOT NULL CHECK (data_type IN ('credentials', 'text', 'binary', 'card')),
    
    login VARCHAR(255),
    password_encrypted BYTEA,
    
    text_data TEXT,
    
    binary_data BYTEA,
    binary_mime_type VARCHAR(100),
    
    card_number_encrypted BYTEA,
    card_holder VARCHAR(255),
    card_expiry_month SMALLINT CHECK (card_expiry_month BETWEEN 1 AND 12),
    card_expiry_year SMALLINT CHECK (card_expiry_year >= 2000),
    card_cvv_encrypted BYTEA,
    card_type VARCHAR(50)
    
    metadata VARCHAR(1024),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secure_data;
-- +goose StatementEnd
