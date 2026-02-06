// Package uid provides utilities for generating globally unique identifiers.
// It leverages UUID v7 for time-ordered sorting and falls back to v4 if necessary.
package uid

import "github.com/google/uuid"

// NewUUID generates a unique identifier using the UUID v7 standard.
//
// UUID v7 is preferred as it is time-ordered (lexicographically sortable),
// making it highly efficient for database primary keys and indexing.
// If v7 generation fails, it falls back to a random UUID v4 string.
func NewUUID() string {
	// UUID V7 includes a timestamp, which improves DB locality and performance.
	newID, err := uuid.NewV7()
	if err != nil {
		// Fallback to V4 (Random) to ensure an ID is always returned.
		return uuid.New().String()
	}
	return newID.String()
}

// NewEventID generates a unique identifier specifically for event tracking.
//
// It currently uses the NewUUID standard. Separating this function allows
// for future changes in event ID formats (e.g., ULID or KSUID) without
// breaking the core UUID generation logic.
func NewEventID() string {
	return NewUUID()
}
