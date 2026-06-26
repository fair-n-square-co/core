//go:build integration

package main

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver for goose
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1/ledgerpbconnect"
	coredb "github.com/fair-n-square-co/core/internal/core/db"
	"github.com/fair-n-square-co/core/internal/core/db/sqlc"
	"github.com/fair-n-square-co/core/internal/ledger/service"
)

const migrationsDir = "../../db/core/migrations"

// TestListFriends_RoundTrip is the acceptance test for FNS-87: a sample connect
// RPC is callable over HTTP and round-trips to a real Core DB. It spins a
// throwaway Postgres, applies the goose migrations, seeds a friendship, and
// calls ListFriends through the same mux the binary serves.
func TestListFriends_RoundTrip(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("core"),
		postgres.WithUsername("core"),
		postgres.WithPassword("core"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(pgContainer) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	runMigrations(t, dsn)

	pool, err := coredb.NewPool(ctx, coredb.DBConfig{
		ConnString:        dsn,
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	callerID, friendID := seedFriendship(t, ctx, pool)

	// Serve the production mux and call it with a real connect client.
	ts := httptest.NewServer(newMux(pool, slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(ts.Close)

	client := ledgerpbconnect.NewFriendServiceClient(http.DefaultClient, ts.URL)
	req := connect.NewRequest(&ledgerpb.ListFriendsRequest{})
	req.Header().Set("X-User-Id", callerID)

	resp, err := client.ListFriends(ctx, req)
	require.NoError(t, err)

	require.Len(t, resp.Msg.GetFriendships(), 1)
	got := resp.Msg.GetFriendships()[0]
	assert.Equal(t, friendID, got.GetFriendId())
	assert.Equal(t, ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_PENDING, got.GetStatus())
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(sqlDB, migrationsDir))
}

// seedFriendship inserts two users plus a pending friendship between them and
// returns (callerID, friendID) as canonical UUID strings, with caller == user_a.
func seedFriendship(t *testing.T, ctx context.Context, pool sqlc.DBTX) (string, string) {
	t.Helper()
	q := sqlc.New(pool)

	u1, err := q.CreateUser(ctx, "auth-subject-1")
	require.NoError(t, err)
	u2, err := q.CreateUser(ctx, "auth-subject-2")
	require.NoError(t, err)

	// The friendship CHECK requires user_a < user_b (byte order).
	low, high := u1, u2
	if bytes.Compare(u1.ID.Bytes[:], u2.ID.Bytes[:]) > 0 {
		low, high = u2, u1
	}

	_, err = q.CreateFriendship(ctx, sqlc.CreateFriendshipParams{
		UserA:         low.ID,
		UserB:         high.ID,
		Status:        service.FriendshipStatusPending,
		StatusActorID: low.ID,
	})
	require.NoError(t, err)

	return uuidString(t, low.ID), uuidString(t, high.ID)
}

func uuidString(t *testing.T, u pgtype.UUID) string {
	t.Helper()
	v, err := u.Value()
	require.NoError(t, err)
	s, ok := v.(string)
	require.True(t, ok)
	return s
}
