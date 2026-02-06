// internal/pkg/ptr/ptr.go
package ptr

import (
	"strings"
	"time"
)

// SafeVal returns the dereferenced value of p if it's not nil, otherwise returns the fallback.
// Use this when you need a specific default value that isn't necessarily the type's zero value.
//
// Example:
//
//	timeout := ptr.SafeVal(entity.updatedAt, 0)
func SafeVal[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}

// ToValue returns the dereferenced value of p or the type's zero value if p is nil.
// String returns "", Int returns 0, Bool returns false, etc.
//
// Use this for "shallow" conversions where a zero value is acceptable.
func ToValue[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// ToPtr wraps any literal or variable into a pointer.
// Primarily used for initializing optional fields in DTOs or unit tests
// where you cannot take the address of a constant directly.
//
// Example:
//
//	req := CreateBookingRequest{ Status: ptr.ToPtr("PENDING") }
func ToPtr[T any](v T) *T {
	return &v
}

// ParseString trims whitespace from a string pointer and performs a "nullify" check.
// If the result is an empty string, it returns nil to ensure database cleanliness
// by avoiding empty strings in optional columns.
func ParseString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if len(trimmed) == 0 {
		return nil
	}
	return &trimmed
}

// ParseTimeToUint64 converts a time.Time pointer to Unix Milliseconds (uint64).
// Safely handles nil pointers and zero-time values by returning nil.
//
// Useful for API responses or DB storage that requires numeric epoch timestamps.
func ParseTimeToUint64(value *time.Time) *uint64 {
	if value == nil || value.IsZero() {
		return nil
	}
	timeInUint64 := uint64(value.UnixMilli())
	return &timeInUint64
}
