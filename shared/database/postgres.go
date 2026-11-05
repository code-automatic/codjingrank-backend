// Package database provides shared database connectivity and pooling.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates a new pgxpool connection pool for PostgreSQL using
// the provided database URL. It pings the database to verify connectivity
// before returning the pool.
//
// The databaseURL must be a valid PostgreSQL connection string (e.g.,
// "postgres://user:password@localhost:5432/dbname?sslmode=disable").
//
// An error is returned if the pool cannot be created or if the database
// is unreachable.
func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL must not be empty")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the database is reachable and responsive.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
