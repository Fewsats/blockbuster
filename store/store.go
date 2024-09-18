package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/fewsats/blockbuster/store/sqlc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DefaultQueryTimeout = time.Minute
	DefaultLimit        = 20
	MaxLimit            = 100
	DefaultOffset       = 0
)

func DefaultContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultQueryTimeout)
}

type Store struct {
	cfg     *Config
	db      *sql.DB
	queries *sqlc.Queries
	logger  *slog.Logger
}

func calculateLimitOffset(limit, offset int32) (int32, int32, error) {
	if limit > MaxLimit {
		return 0, 0, fmt.Errorf("limit exceeds the maximum allowed value of %d", MaxLimit)
	}

	if limit < 0 || offset < 0 {
		return 0, 0, fmt.Errorf("limit and offset must be non-negative")
	}

	if limit == 0 {
		limit = DefaultLimit
	}

	return limit, offset, nil
}

func runMigrations(logger *slog.Logger, cfg *Config) error {
	db, err := sql.Open("sqlite3", cfg.ConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+cfg.MigrationsPath,
		"sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	start := time.Now().UTC()
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	version, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return err
	}

	logger.Info(
		"DB migrations applied",
		"version", version,
		"time", time.Since(start),
	)

	return nil
}

func NewStore(logger *slog.Logger, cfg *Config) (*Store, error) {
	logger.Info(
		"Creating new store",
		"connection_string", cfg.DSN(true),
	)

	db, err := sql.Open("sqlite3", cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(int(cfg.MaxOpenConnections))

	if !cfg.SkipMigrations {
		if err := runMigrations(logger, cfg); err != nil {
			return nil, fmt.Errorf("unable to run migrations: %v", err)
		}
	}

	queries := sqlc.New(db)

	store := &Store{
		cfg:     cfg,
		db:      db,
		queries: queries,
		logger:  logger,
	}

	return store, nil
}


func (s *Store) Close() error {
	return s.db.Close()
}
