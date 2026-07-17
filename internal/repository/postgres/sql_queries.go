package postgres

import (
	"context"
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
