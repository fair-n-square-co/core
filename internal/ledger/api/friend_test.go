package api

import (
	"context"
	"testing"

	ledgerpb "github.com/fair-n-square-co/apis/gen/pkg/fairnsquare/service/ledger/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewFriendServer(t *testing.T) {
	srv := NewFriendServer()
	require.NotNil(t, srv)
}

// TestFriendServer_Unimplemented verifies that the not-yet-implemented RPCs
// return a gRPC Unimplemented status rather than panicking or returning a
// nil error.
func TestFriendServer_Unimplemented(t *testing.T) {
	srv := NewFriendServer()
	ctx := context.Background()

	tests := []struct {
		name string
		call func() (any, error)
	}{
		{
			name: "ManageFriend",
			call: func() (any, error) {
				return srv.ManageFriend(ctx, &ledgerpb.ManageFriendRequest{})
			},
		},
		{
			name: "ListFriends",
			call: func() (any, error) {
				return srv.ListFriends(ctx, &ledgerpb.ListFriendsRequest{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.call()

			require.Error(t, err)
			assert.Nil(t, resp)

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, codes.Unimplemented, st.Code())
		})
	}
}
