package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	TokenTypeKey contextKey = "token_type"
	ClaimsKey    contextKey = "claims"
)

// AuthMiddleware provides JWT authentication using a generic OIDC verifier.
type AuthMiddleware struct {
	verifier *auth.OIDCVerifier
}

// NewAuthMiddleware creates a new authentication middleware using the OIDC verifier.
func NewAuthMiddleware(verifier *auth.OIDCVerifier) *AuthMiddleware {
	return &AuthMiddleware{verifier: verifier}
}

// RequireAuthentication returns a middleware that requires a valid JWT token.
func (m *AuthMiddleware) RequireAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteUnauthorized(w, r, "Authorization header required")
				return
			}

			// Expect "Bearer <token>" format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				errors.WriteUnauthorized(w, r, "Invalid Authorization header format")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				errors.WriteUnauthorized(w, r, "Empty token")
				return
			}

			// Verify the token
			claims, err := m.verifier.VerifyToken(r.Context(), tokenString)
			if err != nil {
				// Map specific errors to appropriate messages
				switch err {
				case auth.ErrTokenExpired:
					errors.WriteUnauthorized(w, r, "Token expired")
				case auth.ErrInvalidIssuer:
					errors.WriteUnauthorized(w, r, "Invalid token issuer")
				case auth.ErrInvalidAudience:
					errors.WriteUnauthorized(w, r, "Invalid token audience")
				case auth.ErrInvalidSignature:
					errors.WriteUnauthorized(w, r, "Invalid token signature")
				default:
					errors.WriteUnauthorized(w, r, "Invalid or expired JWT token")
				}
				return
			}

			// Set context values
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, TokenTypeKey, "jwt")
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			// Set email if present
			if claims.Email != "" {
				ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the context.
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// GetUserEmail extracts the user email from the context.
func GetUserEmail(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailKey).(string)
	return email
}

// GetTokenType extracts the token type from the context.
func GetTokenType(ctx context.Context) string {
	tokenType, _ := ctx.Value(TokenTypeKey).(string)
	return tokenType
}

// GetClaims extracts the JWT claims from the context.
func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims
}
