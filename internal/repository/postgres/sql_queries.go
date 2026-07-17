package postgres

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/internal/domain/auth"
)

func (p *PostgresStorage) CreateUser(ctx context.Context, login, password string) (string, error) {
	query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
	var userid string
	err := p.DB.QueryRowContext(ctx, query, login, password).Scan(&userid)
	return userid, err
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
