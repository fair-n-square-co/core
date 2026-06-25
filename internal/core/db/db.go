package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingTimeout = 5 * time.Second

// DBConfig holds database connection settings. It is composed into the
// application config and passed to NewPool after loading.
//
// The pool-tuning fields are optional: any field left at its zero value falls
// back to the pgx pool default, so a bare ConnString still yields a working
// pool. Sensible defaults are supplied via the embedded YAML config (see
// cmd/core/config) and can be overridden per environment.
type DBConfig struct {
	ConnString string

	// MaxConns caps the total number of connections in the pool.
	MaxConns int32
	// MinConns is the number of idle connections the pool keeps warm.
	MinConns int32
	// MaxConnLifetime is how long a connection may live before it is recycled.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime is how long an idle connection is kept before it is closed.
	MaxConnIdleTime time.Duration
	// HealthCheckPeriod is how often the pool checks the health of idle connections.
	HealthCheckPeriod time.Duration
}

// poolConfig parses the connection string and applies the non-zero tuning
// overrides from c. It is split out from NewPool so the configuration wiring can
// be unit tested without a live database.
func poolConfig(c DBConfig) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(c.ConnString)
	if err != nil {
		return nil, fmt.Errorf("parse db conn string: %w", err)
	}
	if c.MaxConns > 0 {
		cfg.MaxConns = c.MaxConns
	}
	if c.MinConns > 0 {
		cfg.MinConns = c.MinConns
	}
	if c.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = c.MaxConnLifetime
	}
	if c.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = c.MaxConnIdleTime
	}
	if c.HealthCheckPeriod > 0 {
		cfg.HealthCheckPeriod = c.HealthCheckPeriod
	}
	return cfg, nil
}

// NewPool opens a pgx connection pool with the configured tuning and verifies
// connectivity with a ping. The returned *pgxpool.Pool satisfies the generated
// sqlc DBTX interface, so it can be injected directly into module repositories.
func NewPool(ctx context.Context, c DBConfig) (*pgxpool.Pool, error) {
	cfg, err := poolConfig(c)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}
