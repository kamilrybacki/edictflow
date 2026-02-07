package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds configuration for the connection pool
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultPoolConfig returns sensible defaults for production use
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

// NewPool creates a new PostgreSQL connection pool with default configuration
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return NewPoolWithConfig(ctx, databaseURL, DefaultPoolConfig())
}

// NewPoolWithConfig creates a new PostgreSQL connection pool with custom configuration
func NewPoolWithConfig(ctx context.Context, databaseURL string, poolConfig PoolConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Apply pool configuration
	config.MaxConns = poolConfig.MaxConns
	config.MinConns = poolConfig.MinConns
	config.MaxConnLifetime = poolConfig.MaxConnLifetime
	config.MaxConnIdleTime = poolConfig.MaxConnIdleTime
	config.HealthCheckPeriod = poolConfig.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
