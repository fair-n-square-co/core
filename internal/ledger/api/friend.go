package api

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1/ledgerpbconnect"
	"github.com/fair-n-square-co/core/internal/ledger/service"
)

// userIDHeader carries the caller's user id. This is a temporary seam: real
// caller identity will come from the validated auth token in FNS-96. Until
// then the header lets the sample RPC be exercised end-to-end.
const userIDHeader = "X-User-Id"

// FriendService is the slice of the ledger service the handler depends on.
type FriendService interface {
	ListFriends(ctx context.Context, userID string) ([]service.Friend, error)
}

// FriendServer implements the connect FriendService handler. Methods not yet
// implemented fall through to UnimplementedFriendServiceHandler.
type FriendServer struct {
	ledgerpbconnect.UnimplementedFriendServiceHandler
	svc FriendService
}

// NewFriendServer constructs a FriendServer backed by svc.
func NewFriendServer(svc FriendService) *FriendServer {
	return &FriendServer{svc: svc}
}

// ListFriends returns the caller's friendships. The caller is identified by the
// userIDHeader (see its doc comment).
func (f *FriendServer) ListFriends(
	ctx context.Context,
	req *connect.Request[ledgerpb.ListFriendsRequest],
) (*connect.Response[ledgerpb.ListFriendsResponse], error) {
	userID := req.Header().Get(userIDHeader)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing "+userIDHeader+" header"))
	}

	friends, err := f.svc.ListFriends(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUser) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &ledgerpb.ListFriendsResponse{
		Friendships: make([]*ledgerpb.Friend, 0, len(friends)),
	}
	for _, fr := range friends {
		resp.Friendships = append(resp.Friendships, &ledgerpb.Friend{
			Id:       fr.FriendshipID,
			FriendId: fr.FriendID,
			Status:   toProtoStatus(fr.Status),
		})
	}
	return connect.NewResponse(resp), nil
}

// toProtoStatus maps a stored friendship status to the proto enum. Statuses
// without a proto counterpart (rejected, cancelled) map to UNSPECIFIED.
func toProtoStatus(status string) ledgerpb.FriendshipStatus {
	switch status {
	case service.FriendshipStatusPending:
		return ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_PENDING
	case service.FriendshipStatusAccepted:
		return ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_ACCEPTED
	case service.FriendshipStatusBlocked:
		return ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_BLOCKED
	default:
		return ledgerpb.FriendshipStatus_FRIENDSHIP_STATUS_UNSPECIFIED
	}
}
