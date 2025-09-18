package repositories

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/Pro100x3mal/yp-gophermart.git/internal/configs"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg *configs.ServerConfig, logger *zap.Logger) (*DB, error) {
	if cfg.DatabaseURI == "" {
		return nil, errors.New("database URI is not set")
	}

	logger.Info("initializing database storage", zap.String("URI", cfg.DatabaseURI))

	logger.Debug("running database migrations")
	err := runMigrations(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}
	logger.Debug("database migrations completed")

	logger.Debug("connecting to database")
	pool, err := initPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Debug("connected to database")

	logger.Info("database storage initialized successfully")

	return &DB{
		pool: pool,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(cfg *configs.ServerConfig) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("failed to initialize a migration instance: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations to the DB: %w", err)
	}
	return nil
}

func initPool(ctx context.Context, cfg *configs.ServerConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URI %s: %w", cfg.DatabaseURI, err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxTimeout, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err = pool.Ping(ctxTimeout); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func (db *DB) Close() {
	db.pool.Close()
}
