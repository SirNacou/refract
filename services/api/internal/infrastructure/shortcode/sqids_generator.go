package shortcode

import (
	"fmt"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/sqids/sqids-go"
)

// SqidsGenerator implements short code generation using Sqids
type SqidsGenerator struct {
	sqids *sqids.Sqids
}

// NewSqidsGenerator creates a new Sqids-based generator
func NewSqidsGenerator(alphabet string, minLength int) (*SqidsGenerator, error) {
	s, err := sqids.New(sqids.Options{
		Alphabet:  alphabet,
		MinLength: uint8(minLength),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sqids: %w", err)
	}

	return &SqidsGenerator{
		sqids: s,
	}, nil
}

// Generate creates a short code from an ID
func (g *SqidsGenerator) Generate(id int64) (url.ShortCode, error) {
	encoded, err := g.sqids.Encode([]uint64{uint64(id)})
	if err != nil {
		return url.ShortCode{}, fmt.Errorf("failed to encode ID: %w", err)
	}

	return url.NewShortCode(encoded)
}

// Decode extracts the ID from a short code
func (g *SqidsGenerator) Decode(code url.ShortCode) (int64, error) {
	decoded := g.sqids.Decode(code.String())
	if len(decoded) == 0 {
		return 0, fmt.Errorf("failed to decode short code")
	}

	return int64(decoded[0]), nil
}
