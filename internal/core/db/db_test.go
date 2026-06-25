package db

import (
	"context"
	"testing"

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
