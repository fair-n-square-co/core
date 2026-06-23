package service

// FriendshipStatus values mirror the CHECK constraint on friendship.status.
// Stored as text in the database; defined here as constants for validation in the
// service and api layers.
const (
	FriendshipStatusPending   = "pending"
	FriendshipStatusAccepted  = "accepted"
	FriendshipStatusRejected  = "rejected"
	FriendshipStatusCancelled = "cancelled"
	FriendshipStatusBlocked   = "blocked"
)

// FriendEventType values mirror the CHECK constraint on friend_event.type.
const (
	FriendEventRequested = "requested"
	FriendEventAccepted  = "accepted"
	FriendEventRejected  = "rejected"
	FriendEventCancelled = "cancelled"
	FriendEventBlocked   = "blocked"
	FriendEventUnblocked = "unblocked"
)

// validFriendshipStatuses is the set of allowed friendship.status values.
var validFriendshipStatuses = map[string]struct{}{
	FriendshipStatusPending:   {},
	FriendshipStatusAccepted:  {},
	FriendshipStatusRejected:  {},
	FriendshipStatusCancelled: {},
	FriendshipStatusBlocked:   {},
}

// validFriendEventTypes is the set of allowed friend_event.type values.
var validFriendEventTypes = map[string]struct{}{
	FriendEventRequested: {},
	FriendEventAccepted:  {},
	FriendEventRejected:  {},
	FriendEventCancelled: {},
	FriendEventBlocked:   {},
	FriendEventUnblocked: {},
}

// IsValidFriendshipStatus reports whether s is an allowed friendship status.
func IsValidFriendshipStatus(s string) bool {
	_, ok := validFriendshipStatuses[s]
	return ok
}

// IsValidFriendEventType reports whether t is an allowed friend event type.
func IsValidFriendEventType(t string) bool {
	_, ok := validFriendEventTypes[t]
	return ok
}
