// Package repository owns data access for the ledger module, wrapping the
// sqlc-generated queries and translating between database (pgtype) values and
// the plain Go types the service layer consumes.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fair-n-square-co/core/internal/core/db/sqlc"
)

// Friendship is the repository-level view of a friendship row. Identifiers are
// canonical UUID strings so layers above never depend on pgtype.
type Friendship struct {
	ID     string
	UserA  string
	UserB  string
	Status string
}

// Repository provides ledger data access backed by the sqlc query layer.
type Repository struct {
	q *sqlc.Queries
}

// New builds a Repository over any sqlc.DBTX (e.g. a *pgxpool.Pool).
func New(db sqlc.DBTX) *Repository {
	return &Repository{q: sqlc.New(db)}
}

// ListFriendshipsForUser returns every friendship the given user participates
// in, regardless of status, newest first.
func (r *Repository) ListFriendshipsForUser(ctx context.Context, userID string) ([]Friendship, error) {
	uid, err := toPgUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	rows, err := r.q.ListFriendshipsForUser(ctx, sqlc.ListFriendshipsForUserParams{
		UserA:  uid,
		Status: pgtype.Text{}, // NULL -> no status filter
	})
	if err != nil {
		return nil, fmt.Errorf("list friendships: %w", err)
	}

	friendships := make([]Friendship, 0, len(rows))
	for _, row := range rows {
		friendships = append(friendships, Friendship{
			ID:     fromPgUUID(row.ID),
			UserA:  fromPgUUID(row.UserA),
			UserB:  fromPgUUID(row.UserB),
			Status: row.Status,
		})
	}
	return friendships, nil
}

// toPgUUID parses a canonical UUID string into a pgtype.UUID.
func toPgUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

// fromPgUUID renders a pgtype.UUID as its canonical string, or "" if invalid.
func fromPgUUID(u pgtype.UUID) string {
	v, err := u.Value()
	if err != nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
