package url

import "errors"

var (
	ErrURLNotFound       = errors.New("url not found")
	ErrAliasAlreadyTaken = errors.New("alias already taken")
	ErrInvalidURL        = errors.New("invalid url")
	ErrInvalidExpiry     = errors.New("invalid expiry date")
	ErrMaliciousURL      = errors.New("malicious url")
	ErrInvalidShortCode  = errors.New("invalid short code")
)
