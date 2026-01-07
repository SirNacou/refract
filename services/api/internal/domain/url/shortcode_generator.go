package url

// ShortCodeGenerator defines the interface for short code generation
type ShortCodeGenerator interface {
	// Generate creates a short code from an ID
	Generate(id int64) (ShortCode, error)

	// Decode extracts the ID from a short code
	Decode(code ShortCode) (int64, error)
}
