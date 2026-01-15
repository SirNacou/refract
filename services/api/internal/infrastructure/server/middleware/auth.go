package middleware

import (
	"context"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	TokenTypeKey contextKey = "token_type"
	ClaimsKey    contextKey = "claims"
)

type AuthMiddleware struct {
	Interceptor *middleware.Interceptor[*oauth.IntrospectionContext]
}

// NewAuthMiddleware creates a new authentication middleware using the OIDC verifier.
func NewAuthMiddleware(auth *auth.Auth) *AuthMiddleware {
	return &AuthMiddleware{
		Interceptor: middleware.New(auth),
	}
}

func (am *AuthMiddleware) RequireAuthorization(options ...authorization.CheckOption) func(next http.Handler) http.Handler {
	return am.Interceptor.RequireAuthorization(options...)
}

func (am *AuthMiddleware) Context(ctx context.Context) *oauth.IntrospectionContext {
	return am.Interceptor.Context(ctx)
}
