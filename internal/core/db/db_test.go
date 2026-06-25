package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewPool_InvalidURL verifies that NewPool fails fast on a malformed
// connection string without needing a live database. The happy-path round-trip
// (real pool + Ping) is covered by the testcontainers integration test.
func TestNewPool_InvalidURL(t *testing.T) {
	pool, err := NewPool(context.Background(), DBConfig{ConnString: "://not-a-valid-dsn"})

	require.Error(t, err)
	assert.Nil(t, pool)
}

// TestPoolConfig_AppliesTuning verifies the non-zero tuning fields override the
// parsed pgx defaults.
func TestPoolConfig_AppliesTuning(t *testing.T) {
	cfg, err := poolConfig(DBConfig{
		ConnString:        "postgres://user:pass@localhost:5432/core",
		MaxConns:          15,
		MinConns:          3,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	})
	require.NoError(t, err)

	assert.Equal(t, int32(15), cfg.MaxConns)
	assert.Equal(t, int32(3), cfg.MinConns)
	assert.Equal(t, time.Hour, cfg.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, cfg.MaxConnIdleTime)
	assert.Equal(t, time.Minute, cfg.HealthCheckPeriod)
}

// TestPoolConfig_ZeroKeepsDefaults verifies that leaving the tuning fields at
// their zero value preserves pgx's own defaults rather than zeroing the pool.
func TestPoolConfig_ZeroKeepsDefaults(t *testing.T) {
	cfg, err := poolConfig(DBConfig{ConnString: "postgres://user:pass@localhost:5432/core"})
	require.NoError(t, err)

	assert.Positive(t, cfg.MaxConns)
}
