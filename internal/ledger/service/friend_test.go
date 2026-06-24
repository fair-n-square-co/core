package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFriendshipStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{name: "pending is valid", status: FriendshipStatusPending, want: true},
		{name: "accepted is valid", status: FriendshipStatusAccepted, want: true},
		{name: "rejected is valid", status: FriendshipStatusRejected, want: true},
		{name: "cancelled is valid", status: FriendshipStatusCancelled, want: true},
		{name: "blocked is valid", status: FriendshipStatusBlocked, want: true},
		{name: "unknown value is invalid", status: "unknown", want: false},
		{name: "empty string is invalid", status: "", want: false},
		{name: "event type is not a status", status: FriendEventRequested, want: false},
		{name: "case sensitive", status: "Pending", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidFriendshipStatus(tt.status))
		})
	}
}

func TestIsValidFriendEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		want      bool
	}{
		{name: "requested is valid", eventType: FriendEventRequested, want: true},
		{name: "accepted is valid", eventType: FriendEventAccepted, want: true},
		{name: "rejected is valid", eventType: FriendEventRejected, want: true},
		{name: "cancelled is valid", eventType: FriendEventCancelled, want: true},
		{name: "blocked is valid", eventType: FriendEventBlocked, want: true},
		{name: "unblocked is valid", eventType: FriendEventUnblocked, want: true},
		{name: "unknown value is invalid", eventType: "unknown", want: false},
		{name: "empty string is invalid", eventType: "", want: false},
		{name: "pending status is not an event type", eventType: FriendshipStatusPending, want: false},
		{name: "case sensitive", eventType: "Requested", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidFriendEventType(tt.eventType))
		})
	}
}

// TestConstantValues guards the string values, since they must mirror the
// database CHECK constraints exactly.
func TestConstantValues(t *testing.T) {
	assert.Equal(t, "pending", FriendshipStatusPending)
	assert.Equal(t, "accepted", FriendshipStatusAccepted)
	assert.Equal(t, "rejected", FriendshipStatusRejected)
	assert.Equal(t, "cancelled", FriendshipStatusCancelled)
	assert.Equal(t, "blocked", FriendshipStatusBlocked)

	assert.Equal(t, "requested", FriendEventRequested)
	assert.Equal(t, "accepted", FriendEventAccepted)
	assert.Equal(t, "rejected", FriendEventRejected)
	assert.Equal(t, "cancelled", FriendEventCancelled)
	assert.Equal(t, "blocked", FriendEventBlocked)
	assert.Equal(t, "unblocked", FriendEventUnblocked)
}
