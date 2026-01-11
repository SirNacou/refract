package middleware

import "errors"

var (
    // ErrRateLimitExceeded is returned when user exceeds rate limit
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
    
    // ErrUserIDNotFound is returned when userID is missing from context
    ErrUserIDNotFound = errors.New("user ID not found in context")
)