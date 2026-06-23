package api

import (
	"context"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FriendServer struct {
	ledgerpb.UnimplementedFriendServiceServer
}

func NewFriendServer() *FriendServer {
	return &FriendServer{}
}

func (f *FriendServer) ManageFriend(ctx context.Context, req *ledgerpb.ManageFriendRequest) (*ledgerpb.ManageFriendResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ManageFriend not implemented")
}

func (f *FriendServer) ListFriends(ctx context.Context, req *ledgerpb.ListFriendsRequest) (*ledgerpb.ListFriendsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListFriends not implemented")
}
