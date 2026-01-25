package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
)

type contextKey string

const (
	userContextKey contextKey = "user"
)

type User struct {
	ID    string
	Name  string
	Email string
}

type AuthMiddeware struct {
	auth *auth.Auth
}

func NewAuthMiddleware(auth *auth.Auth) *AuthMiddeware {
	return &AuthMiddeware{
		auth: auth,
	}
}

func (am *AuthMiddeware) RequireAuthorization() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwt := r.Header.Get("Authorization")
			if jwt == "" || !strings.HasPrefix(jwt, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			jwt = strings.TrimPrefix(jwt, "Bearer ")

			ac := am.auth.WithJWT(jwt)

			u, err := ac.Get()
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()

			ctx = context.WithValue(ctx, userContextKey, User{
				ID:    u.Id,
				Name:  u.Name,
				Email: u.Email,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (user User, ok bool) {
	user, ok = ctx.Value(userContextKey).(User)
	return user, ok
}

func SetUserToContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
