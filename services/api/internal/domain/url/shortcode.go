package url

import (
	"regexp"
	"strings"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
)

const (
	MinShortCodeLength = 3
	MaxShortCodeLength = 50
)

var (
	shortCodePattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	reservedWords    = map[string]struct{}{
		"admin": {}, "login": {}, "api": {}, // ...add more
	}
)

type ShortCode struct {
	value string
}

func NewShortCode(snowflakeID uint64) *ShortCode {
	return &ShortCode{idgen.EncodeBase62(snowflakeID)}
}

func NewCustomShortCode(s string) (*ShortCode, error) {
	s = strings.TrimSpace(s)
	if len(s) < MinShortCodeLength || len(s) > MaxShortCodeLength {
		return nil, ErrInvalidShortCode
	}
	if !shortCodePattern.MatchString(s) {
		return nil, ErrInvalidShortCode
	}
	if _, found := reservedWords[strings.ToLower(s)]; found {
		return nil, ErrInvalidShortCode
	}
	return &ShortCode{value: s}, nil
}

func (s ShortCode) String() string {
	return s.value
}
