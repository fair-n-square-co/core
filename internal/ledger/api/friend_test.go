package api_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/fair-n-square-co/core/internal/ledger/api"
	"github.com/fair-n-square-co/core/internal/ledger/repository"
	"github.com/fair-n-square-co/core/internal/ledger/service"
	"github.com/fair-n-square-co/core/internal/ledger/service/mocks"
)

const (
	userA = "11111111-1111-1111-1111-111111111111"
	userB = "22222222-2222-2222-2222-222222222222"
)

func newServer(t *testing.T) (*api.FriendServer, *mocks.MockRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepository(ctrl)
	svc := service.NewFriendService(repo)
	return api.NewFriendServer(svc), repo
}

func TestFriendServer_ListFriends_MissingHeader(t *testing.T) {
	srv, _ := newServer(t)

	_, err := srv.ListFriends(context.Background(), connect.NewRequest(&ledgerpb.ListFriendsRequest{}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
}

func TestFriendServer_ListFriends_MapsResponse(t *testing.T) {
	srv, repo := newServer(t)
	repo.EXPECT().ListFriendshipsForUser(gomock.Any(), userA).Return([]repository.Friendship{
		{ID: "f1", UserA: userA, UserB: userB, Status: service.FriendshipStatusAccepted},
	}, nil)

	req := connect.NewRequest(&ledgerpb.ListFriendsRequest{})
	req.Header().Set("X-User-Id", userA)

	resp, err := srv.ListFriends(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, resp.Msg.GetFriendships(), 1)
	friend := resp.Msg.GetFriendships()[0]
	assert.Equal(t, "f1", friend.GetId())
	assert.Equal(t, userB, friend.GetFriendId())
	assert.Equal(t, ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED, friend.GetStatus())
}

// TestFriendServer_ManageFriend_Unimplemented confirms the not-yet-built RPC
// reports CodeUnimplemented via the embedded handler rather than panicking.
func TestFriendServer_ManageFriend_Unimplemented(t *testing.T) {
	srv, _ := newServer(t)

	_, err := srv.ManageFriend(context.Background(), connect.NewRequest(&ledgerpb.ManageFriendRequest{}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}
