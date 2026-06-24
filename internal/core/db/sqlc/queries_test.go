package sqlc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/fair-n-square-co/core/internal/core/db/sqlc/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

// scanType returns an AssignableToTypeOf matcher for a Scan destination of
// type T. Scan destinations are pointers into the (unexported) row struct
// inside the generated code, so their concrete values can't be known — we
// assert their types instead.
func scanType[T any]() gomock.Matcher {
	return gomock.AssignableToTypeOf(reflect.TypeFor[T]())
}

// Scan destination type signatures for each row shape.
var (
	userScan = []any{
		scanType[*pgtype.UUID](), scanType[*string](), scanType[*pgtype.Timestamptz](),
	}
	friendshipScan = []any{
		scanType[*pgtype.UUID](), scanType[*pgtype.UUID](), scanType[*pgtype.UUID](), scanType[*string](),
		scanType[*pgtype.UUID](), scanType[*pgtype.Timestamptz](), scanType[*pgtype.Timestamptz](),
	}
	friendEventScan = []any{
		scanType[*pgtype.UUID](), scanType[*pgtype.UUID](), scanType[*pgtype.UUID](), scanType[*string](),
		scanType[*pgtype.Timestamptz](),
	}
)

// scanReturn drives a mocked Scan: it assigns each provided value into the
// matching destination pointer positionally, simulating a successful row read.
func scanReturn(values ...any) func(dest ...any) error {
	return func(dest ...any) error {
		for i := range dest {
			if i < len(values) {
				reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(values[i]))
			}
		}
		return nil
	}
}

func TestNewAndWithTx(t *testing.T) {
	q := New(mocks.NewMockDBTX(gomock.NewController(t)))
	require.NotNil(t, q)

	// WithTx returns a fresh Queries bound to the transaction. A nil pgx.Tx is
	// fine here: WithTx only stores it.
	got := q.WithTx(nil)
	require.NotNil(t, got)
	assert.NotSame(t, q, got)
}

// rowMock wires a MockDBTX so QueryRow, matched against the exact SQL and
// args, returns the MockRow.
func rowMock(t *testing.T, ctx context.Context, sql string, args ...any) (*Queries, *mocks.MockRow) {
	ctrl := gomock.NewController(t)
	db := mocks.NewMockDBTX(ctrl)
	row := mocks.NewMockRow(ctrl)
	db.EXPECT().QueryRow(ctx, sql, args...).Return(row)
	return New(db), row
}

// --- QueryRow-based queries: success populates the row, error propagates. ---

func TestCreateFriendEvent(t *testing.T) {
	ctx := context.Background()
	params := CreateFriendEventParams{FriendshipID: testUUID(2), ActorID: testUUID(3), Type: "requested"}

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t, ctx, createFriendEvent, params.FriendshipID, params.ActorID, params.Type)
		row.EXPECT().Scan(friendEventScan...).DoAndReturn(
			scanReturn(id, params.FriendshipID, params.ActorID, params.Type, pgtype.Timestamptz{}))

		got, err := q.CreateFriendEvent(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "requested", got.Type)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, createFriendEvent, params.FriendshipID, params.ActorID, params.Type)
		row.EXPECT().Scan(friendEventScan...).Return(errBoom)

		_, err := q.CreateFriendEvent(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateFriendship(t *testing.T) {
	ctx := context.Background()
	params := CreateFriendshipParams{UserA: testUUID(2), UserB: testUUID(3), Status: "pending", StatusActorID: testUUID(2)}

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t, ctx, createFriendship, params.UserA, params.UserB, params.Status, params.StatusActorID)
		row.EXPECT().Scan(friendshipScan...).DoAndReturn(
			scanReturn(id, params.UserA, params.UserB, params.Status, params.StatusActorID, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.CreateFriendship(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "pending", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, createFriendship, params.UserA, params.UserB, params.Status, params.StatusActorID)
		row.EXPECT().Scan(friendshipScan...).Return(errBoom)

		_, err := q.CreateFriendship(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetFriendshipByID(t *testing.T) {
	ctx := context.Background()
	id := testUUID(7)

	t.Run("success", func(t *testing.T) {
		q, row := rowMock(t, ctx, getFriendshipByID, id)
		row.EXPECT().Scan(friendshipScan...).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "accepted", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.GetFriendshipByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "accepted", got.Status)
	})

	t.Run("not found", func(t *testing.T) {
		q, row := rowMock(t, ctx, getFriendshipByID, id)
		row.EXPECT().Scan(friendshipScan...).Return(pgx.ErrNoRows)

		_, err := q.GetFriendshipByID(ctx, id)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetFriendshipByPair(t *testing.T) {
	ctx := context.Background()
	params := GetFriendshipByPairParams{UserA: testUUID(2), UserB: testUUID(3)}

	t.Run("success", func(t *testing.T) {
		id := testUUID(9)
		q, row := rowMock(t, ctx, getFriendshipByPair, params.UserA, params.UserB)
		row.EXPECT().Scan(friendshipScan...).DoAndReturn(
			scanReturn(id, params.UserA, params.UserB, "blocked", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.GetFriendshipByPair(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, getFriendshipByPair, params.UserA, params.UserB)
		row.EXPECT().Scan(friendshipScan...).Return(errBoom)

		_, err := q.GetFriendshipByPair(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUpdateFriendshipStatus(t *testing.T) {
	ctx := context.Background()
	params := UpdateFriendshipStatusParams{ID: testUUID(4), Status: "cancelled", StatusActorID: testUUID(2)}

	t.Run("success", func(t *testing.T) {
		q, row := rowMock(t, ctx, updateFriendshipStatus, params.ID, params.Status, params.StatusActorID)
		row.EXPECT().Scan(friendshipScan...).DoAndReturn(
			scanReturn(params.ID, testUUID(2), testUUID(3), params.Status, params.StatusActorID, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.UpdateFriendshipStatus(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, updateFriendshipStatus, params.ID, params.Status, params.StatusActorID)
		row.EXPECT().Scan(friendshipScan...).Return(errBoom)

		_, err := q.UpdateFriendshipStatus(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	const authSubject = "auth-subject"

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t, ctx, createUser, authSubject)
		row.EXPECT().Scan(userScan...).DoAndReturn(scanReturn(id, authSubject, pgtype.Timestamptz{}))

		got, err := q.CreateUser(ctx, authSubject)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, authSubject, got.AuthSubject)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, createUser, authSubject)
		row.EXPECT().Scan(userScan...).Return(errBoom)

		_, err := q.CreateUser(ctx, authSubject)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetUserByAuthSubject(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		q, row := rowMock(t, ctx, getUserByAuthSubject, "subject-a")
		row.EXPECT().Scan(userScan...).DoAndReturn(scanReturn(testUUID(1), "subject-a", pgtype.Timestamptz{}))

		got, err := q.GetUserByAuthSubject(ctx, "subject-a")
		require.NoError(t, err)
		assert.Equal(t, "subject-a", got.AuthSubject)
	})

	t.Run("not found", func(t *testing.T) {
		q, row := rowMock(t, ctx, getUserByAuthSubject, "missing")
		row.EXPECT().Scan(userScan...).Return(pgx.ErrNoRows)

		_, err := q.GetUserByAuthSubject(ctx, "missing")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	id := testUUID(5)

	t.Run("success", func(t *testing.T) {
		q, row := rowMock(t, ctx, getUserByID, id)
		row.EXPECT().Scan(userScan...).DoAndReturn(scanReturn(id, "subject-b", pgtype.Timestamptz{}))

		got, err := q.GetUserByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t, ctx, getUserByID, id)
		row.EXPECT().Scan(userScan...).Return(errBoom)

		_, err := q.GetUserByID(ctx, id)
		assert.ErrorIs(t, err, errBoom)
	})
}

// --- Query-based (:many) queries: cover query error, scan error, row error
// and the happy path. ---

// rowsMock wires a MockDBTX so Query, matched against the exact SQL and args,
// returns the MockRows (or an error).
func rowsMock(t *testing.T, ctx context.Context, queryErr error, sql string, args ...any) (*Queries, *mocks.MockRows) {
	ctrl := gomock.NewController(t)
	db := mocks.NewMockDBTX(ctrl)
	rows := mocks.NewMockRows(ctrl)
	if queryErr != nil {
		db.EXPECT().Query(ctx, sql, args...).Return(nil, queryErr)
	} else {
		db.EXPECT().Query(ctx, sql, args...).Return(rows, nil)
	}
	return New(db), rows
}

func TestListEventsForFriendship(t *testing.T) {
	ctx := context.Background()
	fid := testUUID(2)

	t.Run("query error", func(t *testing.T) {
		q, _ := rowsMock(t, ctx, errBoom, listEventsForFriendship, fid)
		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listEventsForFriendship, fid)
		gomock.InOrder(
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(friendEventScan...).DoAndReturn(scanReturn(testUUID(1), fid, testUUID(3), "requested", pgtype.Timestamptz{})),
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(friendEventScan...).DoAndReturn(scanReturn(testUUID(2), fid, testUUID(3), "accepted", pgtype.Timestamptz{})),
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(nil),
		)
		rows.EXPECT().Close()

		got, err := q.ListEventsForFriendship(ctx, fid)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "requested", got[0].Type)
		assert.Equal(t, "accepted", got[1].Type)
	})

	t.Run("scan error", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listEventsForFriendship, fid)
		rows.EXPECT().Next().Return(true)
		rows.EXPECT().Scan(friendEventScan...).Return(errBoom)
		rows.EXPECT().Close()

		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listEventsForFriendship, fid)
		gomock.InOrder(
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(errBoom),
		)
		rows.EXPECT().Close()

		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestListFriendshipsForUser(t *testing.T) {
	ctx := context.Background()
	params := ListFriendshipsForUserParams{UserA: testUUID(2)}

	t.Run("query error", func(t *testing.T) {
		q, _ := rowsMock(t, ctx, errBoom, listFriendshipsForUser, params.UserA, params.Status)
		_, err := q.ListFriendshipsForUser(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listFriendshipsForUser, params.UserA, params.Status)
		gomock.InOrder(
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(friendshipScan...).DoAndReturn(scanReturn(testUUID(1), params.UserA, testUUID(3), "pending", params.UserA, pgtype.Timestamptz{}, pgtype.Timestamptz{})),
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(nil),
		)
		rows.EXPECT().Close()

		got, err := q.ListFriendshipsForUser(ctx, params)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "pending", got[0].Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listFriendshipsForUser, params.UserA, params.Status)
		rows.EXPECT().Next().Return(true)
		rows.EXPECT().Scan(friendshipScan...).Return(errBoom)
		rows.EXPECT().Close()

		_, err := q.ListFriendshipsForUser(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q, rows := rowsMock(t, ctx, nil, listFriendshipsForUser, params.UserA, params.Status)
		gomock.InOrder(
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(errBoom),
		)
		rows.EXPECT().Close()

		_, err := q.ListFriendshipsForUser(ctx, params)
		assert.ErrorIs(t, err, errBoom)
	})
}
