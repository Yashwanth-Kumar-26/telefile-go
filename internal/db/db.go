package db

import (
	"context"
	"fmt"

	"github.com/Yashwanth-Kumar-26/telefile/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// Connect establishes connection to the database
func Connect(cfg *config.Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5

	ctx := context.Background()
	Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

// MustConnect panics if connection fails
func MustConnect(cfg *config.Config) {
	if err := Connect(cfg); err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
}