package url

// IDGenerator generates unique identifiers for URLs
// This abstraction allows for different ID generation strategies
// (e.g., Snowflake, ULID, UUID) without affecting the domain layer
type IDGenerator interface {
	// Generate creates a new unique ID
	// Returns int64 for compatibility with Sqids encoding
	Generate() int64
}
