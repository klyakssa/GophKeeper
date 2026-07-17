package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"gophkeeper/internal/config"
	"gophkeeper/internal/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var (
	ErrConnection         = errors.New("connection error")     // return when connection error
	ErrMigration          = errors.New("migration error")      // return when migration error
	ErrNoConnectionString = errors.New("no connection string") // return when no connection string
)

// PostgresStorage
type PostgresStorage struct {
	*sqlx.DB
	cfg *config.Config
	l   *logger.Logger
}

// NewPostgresStorage returns a new instance of PostgresStorage
func NewPostgresStorage(cfg *config.Config, l *logger.Logger) (*PostgresStorage, error) {
	c, err := connectPostgres(cfg)
	return &PostgresStorage{
			c,
			cfg,
			l,
		},
		err
}

// Close closes the database
func (s *PostgresStorage) Close() error {
	return s.DB.Close()
}

func connectPostgres(cfg *config.Config) (*sqlx.DB, error) {
	if cfg.PostDB.ConnectionString == "" {
		return nil, ErrNoConnectionString
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(cfg.PostDB.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.ConnConfig.ConnectTimeout = 5 * time.Second

	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	db := stdlib.OpenDBFromPool(dbpool)
	sqlxDB := sqlx.NewDb(db, "pgx")
	if err := applyMigrations(sqlxDB); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}
	return sqlxDB, nil
}

func applyMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
