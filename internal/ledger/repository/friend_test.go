package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListFriendshipsForUser_InvalidUUID verifies the repository rejects a
// malformed user id before issuing a query. The happy-path round-trip against a
// real database is covered by the testcontainers integration test.
func TestListFriendshipsForUser_InvalidUUID(t *testing.T) {
	repo := New(nil) // db is never touched: parsing fails first

	got, err := repo.ListFriendshipsForUser(context.Background(), "not-a-uuid")

	require.Error(t, err)
	assert.Nil(t, got)
}
