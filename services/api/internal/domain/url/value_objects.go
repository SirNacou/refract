package url

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	shortCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
)

// ShortCode represents a validated short code
type ShortCode struct {
	value string
}

// NewShortCode creates a new ShortCode after validation
func NewShortCode(s string) (ShortCode, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return ShortCode{}, NewValidationError("INVALID_SHORT_CODE", "Short code cannot be empty")
	}

	if !shortCodeRegex.MatchString(s) {
		return ShortCode{}, NewValidationError(
			"INVALID_SHORT_CODE",
			"Short code must be 4-20 alphanumeric characters",
		)
	}

	return ShortCode{value: s}, nil
}

// String returns the short code value
func (s ShortCode) String() string {
	return s.value
}

// OriginalURL represents a validated URL
type OriginalURL struct {
	value string
}

// NewOriginalURL creates a new OriginalURL after validation
func NewOriginalURL(s string) (OriginalURL, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return OriginalURL{}, NewValidationError("INVALID_URL", "URL cannot be empty")
	}

	parsedURL, err := url.Parse(s)
	if err != nil {
		return OriginalURL{}, NewValidationError("INVALID_URL", "URL format is invalid")
	}

	if parsedURL.Scheme == "" {
		return OriginalURL{}, NewValidationError(
			"INVALID_URL",
			"URL must include a scheme (http:// or https://)",
		)
	}

	if parsedURL.Host == "" {
		return OriginalURL{}, NewValidationError(
			"INVALID_URL",
			"URL must include a host",
		)
	}

	return OriginalURL{value: s}, nil
}

// String returns the URL value
func (o OriginalURL) String() string {
	return o.value
}

// Domain represents a validated domain name
type Domain struct {
	value string
}

// NewDomain creates a new Domain after validation
func NewDomain(s string) (Domain, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "" {
		return Domain{}, NewValidationError("INVALID_DOMAIN", "Domain cannot be empty")
	}

	// Basic format check: must contain at least one dot and no spaces
	if !strings.Contains(s, ".") || strings.Contains(s, " ") {
		return Domain{}, NewValidationError(
			"INVALID_DOMAIN",
			"Domain format is invalid",
		)
	}

	return Domain{value: s}, nil
}

// String returns the domain value
func (d Domain) String() string {
	return d.value
}
