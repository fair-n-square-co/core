package service

import (
	"context"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
)

type FriendServer struct {
	ledgerpb.UnimplementedFriendServiceServer
}

func NewFriendServer() *FriendServer {
	return &FriendServer{}
}

func (f *FriendServer) ManageFriend(ctx context.Context, req *ledgerpb.ManageFriendRequest) (*ledgerpb.ManageFriendResponse, error) {
	return &ledgerpb.ManageFriendResponse{}, nil
}

func (f *FriendServer) ListFriends(ctx context.Context, req *ledgerpb.ListFriendsRequest) (*ledgerpb.ListFriendsResponse, error) {
	return &ledgerpb.ListFriendsResponse{}, nil
}
