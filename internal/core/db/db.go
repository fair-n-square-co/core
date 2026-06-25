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
type DBConfig struct {
	ConnString string
}

// NewPool opens a pgx connection pool and verifies connectivity with a ping.
// The returned *pgxpool.Pool satisfies the generated sqlc DBTX interface, so it
// can be injected directly into module repositories.
func NewPool(ctx context.Context, c DBConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, c.ConnString)
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
