package service

// RelationshipStatus values mirror the CHECK constraint on relationship.status.
// Stored as text in the database; defined here as constants for validation in the
// service and api layers.
const (
	RelationshipStatusPending   = "pending"
	RelationshipStatusAccepted  = "accepted"
	RelationshipStatusRejected  = "rejected"
	RelationshipStatusCancelled = "cancelled"
	RelationshipStatusBlocked   = "blocked"
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

// validRelationshipStatuses is the set of allowed relationship.status values.
var validRelationshipStatuses = map[string]struct{}{
	RelationshipStatusPending:   {},
	RelationshipStatusAccepted:  {},
	RelationshipStatusRejected:  {},
	RelationshipStatusCancelled: {},
	RelationshipStatusBlocked:   {},
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

// IsValidRelationshipStatus reports whether s is an allowed relationship status.
func IsValidRelationshipStatus(s string) bool {
	_, ok := validRelationshipStatuses[s]
	return ok
}

// IsValidFriendEventType reports whether t is an allowed friend event type.
func IsValidFriendEventType(t string) bool {
	_, ok := validFriendEventTypes[t]
	return ok
}
