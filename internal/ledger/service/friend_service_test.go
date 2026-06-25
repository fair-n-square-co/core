package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fair-n-square-co/core/internal/ledger/repository"
	"github.com/fair-n-square-co/core/internal/ledger/service"
	"github.com/fair-n-square-co/core/internal/ledger/service/mocks"
)

const (
	userA = "11111111-1111-1111-1111-111111111111"
	userB = "22222222-2222-2222-2222-222222222222"
)

func TestFriendService_ListFriends_EmptyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)

	svc := service.NewFriendService(repo)
	got, err := svc.ListFriends(context.Background(), "")

	require.ErrorIs(t, err, service.ErrInvalidUser)
	assert.Nil(t, got)
}

func TestFriendService_ListFriends_ResolvesOtherParticipant(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)

	// Two friendships: in the first the caller is user_a, in the second user_b.
	// In both cases the "other" participant should be resolved correctly.
	repo.EXPECT().ListFriendshipsForUser(gomock.Any(), userA).Return([]repository.Friendship{
		{ID: "f1", UserA: userA, UserB: userB, Status: "accepted"},
		{ID: "f2", UserA: userB, UserB: userA, Status: "pending"},
	}, nil)

	svc := service.NewFriendService(repo)
	got, err := svc.ListFriends(context.Background(), userA)

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, service.Friend{FriendshipID: "f1", FriendID: userB, Status: "accepted"}, got[0])
	assert.Equal(t, service.Friend{FriendshipID: "f2", FriendID: userB, Status: "pending"}, got[1])
}

func TestFriendService_ListFriends_PropagatesRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)

	wantErr := errors.New("db down")
	repo.EXPECT().ListFriendshipsForUser(gomock.Any(), userA).Return(nil, wantErr)

	svc := service.NewFriendService(repo)
	got, err := svc.ListFriends(context.Background(), userA)

	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, got)
}
