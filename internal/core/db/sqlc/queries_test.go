package sqlc

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func testUUID(b byte) pgtype.UUID {
	var u pgtype.UUID
	u.Valid = true
	for i := range u.Bytes {
		u.Bytes[i] = b
	}
	return u
}

func TestNewAndWithTx(t *testing.T) {
	q := New(&fakeDBTX{})
	require.NotNil(t, q)

	// WithTx returns a fresh Queries bound to the transaction. Passing a nil
	// pgx.Tx is fine here: WithTx only stores it.
	got := q.WithTx(nil)
	require.NotNil(t, got)
	assert.NotSame(t, q, got)
}

// rowDB builds a Queries whose QueryRow returns a fakeRow driven by scan.
func rowDB(scan func(dest ...any) error) *Queries {
	return New(&fakeDBTX{
		queryRowFn: func(_ context.Context, _ string, _ ...interface{}) pgx.Row {
			return fakeRow{scan: scan}
		},
	})
}

func TestCreateFriendEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q := rowDB(scanValues(id, testUUID(2), testUUID(3), "requested", pgtype.Timestamptz{}))
		got, err := q.CreateFriendEvent(ctx, CreateFriendEventParams{Type: "requested"})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "requested", got.Type)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.CreateFriendEvent(ctx, CreateFriendEventParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateFriendship(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q := rowDB(scanValues(id, testUUID(2), testUUID(3), "pending", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))
		got, err := q.CreateFriendship(ctx, CreateFriendshipParams{Status: "pending"})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "pending", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.CreateFriendship(ctx, CreateFriendshipParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetFriendshipByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(7)
		q := rowDB(scanValues(id, testUUID(2), testUUID(3), "accepted", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))
		got, err := q.GetFriendshipByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "accepted", got.Status)
	})

	t.Run("not found", func(t *testing.T) {
		q := rowDB(scanErr(pgx.ErrNoRows))
		_, err := q.GetFriendshipByID(ctx, testUUID(7))
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetFriendshipByPair(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(9)
		q := rowDB(scanValues(id, testUUID(2), testUUID(3), "blocked", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))
		got, err := q.GetFriendshipByPair(ctx, GetFriendshipByPairParams{UserA: testUUID(2), UserB: testUUID(3)})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.GetFriendshipByPair(ctx, GetFriendshipByPairParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUpdateFriendshipStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(4)
		q := rowDB(scanValues(id, testUUID(2), testUUID(3), "cancelled", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))
		got, err := q.UpdateFriendshipStatus(ctx, UpdateFriendshipStatusParams{ID: id, Status: "cancelled"})
		require.NoError(t, err)
		assert.Equal(t, "cancelled", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.UpdateFriendshipStatus(ctx, UpdateFriendshipStatusParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q := rowDB(scanValues(id, "auth-subject", pgtype.Timestamptz{}))
		got, err := q.CreateUser(ctx, "auth-subject")
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "auth-subject", got.AuthSubject)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.CreateUser(ctx, "auth-subject")
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetUserByAuthSubject(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		q := rowDB(scanValues(testUUID(1), "subject-a", pgtype.Timestamptz{}))
		got, err := q.GetUserByAuthSubject(ctx, "subject-a")
		require.NoError(t, err)
		assert.Equal(t, "subject-a", got.AuthSubject)
	})

	t.Run("not found", func(t *testing.T) {
		q := rowDB(scanErr(pgx.ErrNoRows))
		_, err := q.GetUserByAuthSubject(ctx, "missing")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(5)
		q := rowDB(scanValues(id, "subject-b", pgtype.Timestamptz{}))
		got, err := q.GetUserByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q := rowDB(scanErr(errBoom))
		_, err := q.GetUserByID(ctx, testUUID(5))
		assert.ErrorIs(t, err, errBoom)
	})
}

// --- Query-based (:many) queries: cover query error, scan error, row error
// and the happy path. ---

func TestListEventsForFriendship(t *testing.T) {
	ctx := context.Background()
	fid := testUUID(2)

	t.Run("query error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return nil, errBoom
		}})
		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{scans: []func(dest ...any) error{
				scanValues(testUUID(1), fid, testUUID(3), "requested", pgtype.Timestamptz{}),
				scanValues(testUUID(2), fid, testUUID(3), "accepted", pgtype.Timestamptz{}),
			}}, nil
		}})
		got, err := q.ListEventsForFriendship(ctx, fid)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "requested", got[0].Type)
		assert.Equal(t, "accepted", got[1].Type)
	})

	t.Run("scan error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{scans: []func(dest ...any) error{scanErr(errBoom)}}, nil
		}})
		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{err: errBoom}, nil
		}})
		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestListFriendshipsForUser(t *testing.T) {
	ctx := context.Background()
	uid := testUUID(2)

	t.Run("query error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return nil, errBoom
		}})
		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{scans: []func(dest ...any) error{
				scanValues(testUUID(1), uid, testUUID(3), "pending", uid, pgtype.Timestamptz{}, pgtype.Timestamptz{}),
			}}, nil
		}})
		got, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "pending", got[0].Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{scans: []func(dest ...any) error{scanErr(errBoom)}}, nil
		}})
		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q := New(&fakeDBTX{queryFn: func(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
			return &fakeRows{err: errBoom}, nil
		}})
		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})
}
