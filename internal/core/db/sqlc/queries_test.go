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

// rowMock wires a MockDBTX so QueryRow returns the given MockRow.
func rowMock(t *testing.T) (*Queries, *mocks.MockRow) {
	ctrl := gomock.NewController(t)
	db := mocks.NewMockDBTX(ctrl)
	row := mocks.NewMockRow(ctrl)
	db.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).Return(row)
	return New(db), row
}

// --- QueryRow-based queries: success populates the row, error propagates. ---

func TestCreateFriendEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "requested", pgtype.Timestamptz{}))

		got, err := q.CreateFriendEvent(ctx, CreateFriendEventParams{Type: "requested"})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "requested", got.Type)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.CreateFriendEvent(ctx, CreateFriendEventParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateFriendship(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "pending", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.CreateFriendship(ctx, CreateFriendshipParams{Status: "pending"})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "pending", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.CreateFriendship(ctx, CreateFriendshipParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetFriendshipByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(7)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "accepted", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.GetFriendshipByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "accepted", got.Status)
	})

	t.Run("not found", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)

		_, err := q.GetFriendshipByID(ctx, testUUID(7))
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetFriendshipByPair(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(9)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "blocked", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.GetFriendshipByPair(ctx, GetFriendshipByPairParams{UserA: testUUID(2), UserB: testUUID(3)})
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.GetFriendshipByPair(ctx, GetFriendshipByPairParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestUpdateFriendshipStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(4)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(
			scanReturn(id, testUUID(2), testUUID(3), "cancelled", testUUID(2), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		got, err := q.UpdateFriendshipStatus(ctx, UpdateFriendshipStatusParams{ID: id, Status: "cancelled"})
		require.NoError(t, err)
		assert.Equal(t, "cancelled", got.Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.UpdateFriendshipStatus(ctx, UpdateFriendshipStatusParams{})
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(1)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(id, "auth-subject", pgtype.Timestamptz{}))

		got, err := q.CreateUser(ctx, "auth-subject")
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "auth-subject", got.AuthSubject)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.CreateUser(ctx, "auth-subject")
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestGetUserByAuthSubject(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(testUUID(1), "subject-a", pgtype.Timestamptz{}))

		got, err := q.GetUserByAuthSubject(ctx, "subject-a")
		require.NoError(t, err)
		assert.Equal(t, "subject-a", got.AuthSubject)
	})

	t.Run("not found", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)

		_, err := q.GetUserByAuthSubject(ctx, "missing")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		id := testUUID(5)
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(id, "subject-b", pgtype.Timestamptz{}))

		got, err := q.GetUserByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("scan error", func(t *testing.T) {
		q, row := rowMock(t)
		row.EXPECT().Scan(gomock.Any()).Return(errBoom)

		_, err := q.GetUserByID(ctx, testUUID(5))
		assert.ErrorIs(t, err, errBoom)
	})
}

// --- Query-based (:many) queries: cover query error, scan error, row error
// and the happy path. ---

// rowsMock wires a MockDBTX so Query returns the given MockRows (or an error).
func rowsMock(t *testing.T, queryErr error) (*Queries, *mocks.MockRows) {
	ctrl := gomock.NewController(t)
	db := mocks.NewMockDBTX(ctrl)
	rows := mocks.NewMockRows(ctrl)
	if queryErr != nil {
		db.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, queryErr)
	} else {
		db.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).Return(rows, nil)
	}
	return New(db), rows
}

func TestListEventsForFriendship(t *testing.T) {
	ctx := context.Background()
	fid := testUUID(2)

	t.Run("query error", func(t *testing.T) {
		q, _ := rowsMock(t, errBoom)
		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q, rows := rowsMock(t, nil)
		gomock.InOrder(
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(testUUID(1), fid, testUUID(3), "requested", pgtype.Timestamptz{})),
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(testUUID(2), fid, testUUID(3), "accepted", pgtype.Timestamptz{})),
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
		q, rows := rowsMock(t, nil)
		rows.EXPECT().Next().Return(true)
		rows.EXPECT().Scan(gomock.Any()).Return(errBoom)
		rows.EXPECT().Close()

		_, err := q.ListEventsForFriendship(ctx, fid)
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q, rows := rowsMock(t, nil)
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
	uid := testUUID(2)

	t.Run("query error", func(t *testing.T) {
		q, _ := rowsMock(t, errBoom)
		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("success with rows", func(t *testing.T) {
		q, rows := rowsMock(t, nil)
		gomock.InOrder(
			rows.EXPECT().Next().Return(true),
			rows.EXPECT().Scan(gomock.Any()).DoAndReturn(scanReturn(testUUID(1), uid, testUUID(3), "pending", uid, pgtype.Timestamptz{}, pgtype.Timestamptz{})),
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(nil),
		)
		rows.EXPECT().Close()

		got, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "pending", got[0].Status)
	})

	t.Run("scan error", func(t *testing.T) {
		q, rows := rowsMock(t, nil)
		rows.EXPECT().Next().Return(true)
		rows.EXPECT().Scan(gomock.Any()).Return(errBoom)
		rows.EXPECT().Close()

		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("rows error", func(t *testing.T) {
		q, rows := rowsMock(t, nil)
		gomock.InOrder(
			rows.EXPECT().Next().Return(false),
			rows.EXPECT().Err().Return(errBoom),
		)
		rows.EXPECT().Close()

		_, err := q.ListFriendshipsForUser(ctx, ListFriendshipsForUserParams{UserA: uid})
		assert.ErrorIs(t, err, errBoom)
	})
}
