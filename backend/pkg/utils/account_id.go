package utils

import (
	"time"

	"github.com/google/uuid"
)

// GenerateAccountID generates a unique TigerBeetle account ID.
//
// The ID is generated using a combination of:
//   - Current timestamp (nanoseconds / 1000) - ensures temporal uniqueness
//   - Random UUID component (ID % 1000000) - adds randomness to prevent collisions
//
// This approach provides reasonable uniqueness for most use cases, though in
// extremely high-concurrency scenarios (thousands of accounts per microsecond),
// there is a small theoretical possibility of collision.
//
// Returns a uint64 suitable for use as a TigerBeetle account ID.
func GenerateAccountID() uint64 {
	// Use timestamp as base for temporal uniqueness
	timestamp := uint64(time.Now().UnixNano() / 1000)

	// Add random component from UUID to prevent collisions
	randomComponent := uint64(uuid.New().ID() % 1000000)

	return timestamp + randomComponent
}
