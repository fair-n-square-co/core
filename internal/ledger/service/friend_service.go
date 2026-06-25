package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/fair-n-square-co/core/internal/ledger/repository"
)

// ErrInvalidUser is returned when a friend operation is attempted without a
// valid caller identity (a missing or malformed user id).
var ErrInvalidUser = errors.New("invalid user id")

// Repository is the data-access surface the friend service depends on. It is an
// interface so the service can be unit-tested with a generated mock.
//
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/repository.go -package=mocks . Repository
type Repository interface {
	ListFriendshipsForUser(ctx context.Context, userID string) ([]repository.Friendship, error)
}

// Friend is the service-level view of a friendship from one user's perspective:
// the friendship id, the *other* participant, and the current status.
type Friend struct {
	FriendshipID string
	FriendID     string
	Status       string
}

// FriendService holds the ledger module's friend business logic.
type FriendService struct {
	repo Repository
}

// NewFriendService constructs a FriendService backed by repo.
func NewFriendService(repo Repository) *FriendService {
	return &FriendService{repo: repo}
}

// ListFriends returns the friendships for userID, each resolved to the other
// participant from userID's point of view.
func (s *FriendService) ListFriends(ctx context.Context, userID string) ([]Friend, error) {
	// Validate the caller id here (not in the repository) so malformed input is
	// a caller error (4xx) rather than an internal one. Catches both the empty
	// and the malformed-UUID cases.
	if _, err := uuid.Parse(userID); err != nil {
		return nil, ErrInvalidUser
	}

	rows, err := s.repo.ListFriendshipsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	friends := make([]Friend, 0, len(rows))
	for _, row := range rows {
		other := row.UserB
		if row.UserA != userID {
			other = row.UserA
		}
		friends = append(friends, Friend{
			FriendshipID: row.ID,
			FriendID:     other,
			Status:       row.Status,
		})
	}
	return friends, nil
}
