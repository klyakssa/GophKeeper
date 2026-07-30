package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophkeeper/internal/domain/auth"
	"gophkeeper/internal/domain/secure"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *PostgresStorage) CreateUser(ctx context.Context, login, password string) (string, error) {
	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
	var userid string
	if err := p.DB.QueryRowContext(ctx, query, login, password).Scan(&userid); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", auth.ErrUserAlreadyExists
		}
		return "", err
	}
	return userid, nil
}

func (p *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*auth.User, error) {
	query := `SELECT id, login, password FROM users WHERE login = $1`
	row := p.DB.QueryRowxContext(ctx, query, login)
	user := &auth.User{}
	err := row.StructScan(user)
	return user, err
}

func (p *PostgresStorage) GetUserByID(ctx context.Context, userid string) (*auth.User, error) {
	query := `SELECT id, login, password FROM users WHERE id = $1`

	var user auth.User
	err := p.DB.GetContext(ctx, &user, query, userid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (p *PostgresStorage) CreateSecureData(ctx context.Context, req *secure.SecureDataCreate) error {
	query := `
    INSERT INTO secure_data (
		user_id,
        data_type,
        login,
        password_encrypted,
        text_data,
        binary_data,
        binary_mime_type,
        card_number_encrypted,
        card_holder,
        card_expiry_month,
        card_expiry_year,
        card_cvv_encrypted,
        card_type,
        metadata,
        created_at,
        updated_at
    ) VALUES (
		:user_id,
        :data_type,
        :login,
        :password_encrypted,
        :text_data,
        :binary_data,
        :binary_mime_type,
        :card_number_encrypted,
        :card_holder,
        :card_expiry_month,
        :card_expiry_year,
        :card_cvv_encrypted,
        :card_type,
        :metadata,
        NOW(),
        NOW()
    )
`

	_, err := p.DB.NamedExecContext(ctx, query, req)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetSecureData(ctx context.Context, userid string) ([]secure.SecureData, error) {
	query := `SELECT id, user_id, data_type, login, password_encrypted, text_data, binary_data, binary_mime_type, card_number_encrypted, card_holder, card_expiry_month, card_expiry_year, card_cvv_encrypted, card_type, metadata, created_at, updated_at FROM secure_data WHERE user_id = :userid AND deleted_at IS NULL`

	var secureData []secure.SecureData

	params := map[string]interface{}{
		"userid": userid,
	}

	rows, err := p.DB.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query secure data: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data secure.SecureData
		if err := rows.StructScan(&data); err != nil {
			return nil, fmt.Errorf("failed to scan secure data: %w", err)
		}
		secureData = append(secureData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return secureData, nil
}

func (p *PostgresStorage) DeleteSecureData(ctx context.Context, userid string, id int) error {
	query := `UPDATE secure_data SET deleted_at = NOW() WHERE id = :id AND user_id = :userid AND deleted_at IS NULL`
	_, err := p.DB.NamedExecContext(ctx, query, map[string]interface{}{"id": id, "userid": userid})
	return err
}

func (p *PostgresStorage) UpdateSecureData(ctx context.Context, req *secure.SecureDataUpdate) error {
	query := `UPDATE secure_data SET`

	var setParts []string
	args := map[string]interface{}{
		"id":     req.ID,
		"userid": req.UserID,
	}

	if req.DataType != nil {
		setParts = append(setParts, "data_type = :data_type")
		args["data_type"] = *req.DataType
	}
	if req.Login != nil {
		setParts = append(setParts, "login = :login")
		args["login"] = *req.Login
	}
	if req.PasswordEncrypted != nil {
		setParts = append(setParts, "password_encrypted = :password_encrypted")
		args["password_encrypted"] = req.PasswordEncrypted
	}
	if req.TextData != nil {
		setParts = append(setParts, "text_data = :text_data")
		args["text_data"] = *req.TextData
	}
	if req.BinaryData != nil {
		setParts = append(setParts, "binary_data = :binary_data")
		args["binary_data"] = req.BinaryData
	}
	if req.BinaryMimeType != nil {
		setParts = append(setParts, "binary_mime_type = :binary_mime_type")
		args["binary_mime_type"] = *req.BinaryMimeType
	}
	if req.CardNumberEncrypted != nil {
		setParts = append(setParts, "card_number_encrypted = :card_number_encrypted")
		args["card_number_encrypted"] = req.CardNumberEncrypted
	}
	if req.CardHolder != nil {
		setParts = append(setParts, "card_holder = :card_holder")
		args["card_holder"] = *req.CardHolder
	}
	if req.CardExpiryMonth != nil {
		setParts = append(setParts, "card_expiry_month = :card_expiry_month")
		args["card_expiry_month"] = *req.CardExpiryMonth
	}
	if req.CardExpiryYear != nil {
		setParts = append(setParts, "card_expiry_year = :card_expiry_year")
		args["card_expiry_year"] = *req.CardExpiryYear
	}
	if req.CardCvvEncrypted != nil {
		setParts = append(setParts, "card_cvv_encrypted = :card_cvv_encrypted")
		args["card_cvv_encrypted"] = req.CardCvvEncrypted
	}
	if req.CardType != nil {
		setParts = append(setParts, "card_type = :card_type")
		args["card_type"] = *req.CardType
	}
	if req.Metadata != nil {
		setParts = append(setParts, "metadata = :metadata")
		args["metadata"] = *req.Metadata
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")

	query += " " + strings.Join(setParts, ", ")
	query += " WHERE id = :id AND user_id = :userid AND deleted_at IS NULL"

	_, err := p.DB.NamedExecContext(ctx, query, args)
	return err
}
