package middleware

import (
	"context"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/errors"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	TokenTypeKey   contextKey = "token_type"
	AuthContextKey contextKey = "auth_context"
)

type AuthMiddleware struct {
	interceptor *middleware.Interceptor[*oauth.IntrospectionContext]
}

func NewAuthMiddleware(zitadel *auth.ZitadelProvider) *AuthMiddleware {
	mw := middleware.New(zitadel.AuthZ)
	return &AuthMiddleware{mw}
}

func (m *AuthMiddleware) RequireAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return m.interceptor.RequireAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := m.interceptor.Context(r.Context())

			if !authCtx.IsAuthorized() {
				errors.WriteUnauthorized(w, r, "Invalid or expired JWT token")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, authCtx.UserID())
			ctx = context.WithValue(ctx, TokenTypeKey, "jwt")
			ctx = context.WithValue(ctx, AuthContextKey, authCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return m.interceptor.RequireAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := m.interceptor.Context(r.Context())

			if !authCtx.IsGrantedRole(role) {
				errors.WriteForbidden(w, r, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

func GetTokenType(ctx context.Context) string {
	tokenType, _ := ctx.Value(TokenTypeKey).(string)
	return tokenType
}

func GetAuthContext(ctx context.Context) *oauth.IntrospectionContext {
	authCtx, _ := ctx.Value(AuthContextKey).(*oauth.IntrospectionContext)
	return authCtx
}
