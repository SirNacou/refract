package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/SirNacou/refract/api/internal/infrastructure/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

type AuthMiddleware struct {
	cache *jwk.Cache
	api   huma.API
	url   string
}

func NewAuthMiddleware(ctx context.Context, jwksURL string) (*AuthMiddleware, error) {
	cache, err := jwk.NewCache(ctx, httprc.NewClient(httprc.WithHTTPClient(&http.Client{
		Timeout: 5 * time.Second,
	})))
	if err != nil {
		slog.Error("Failed to create JWK cache", slog.Any("error", err))
		return nil, err
	}

	err = cache.Register(ctx, jwksURL, jwk.WithMinInterval(15*time.Minute))
	if err != nil {
		slog.Error("Failed to register JWK URL", slog.Any("error", err))
		return nil, err
	}

	return &AuthMiddleware{
		cache: cache,
		url:   jwksURL,
	}, nil
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyset, err := am.cache.Lookup(r.Context(), am.url)
		if err != nil {
			slog.Error("Failed to fetch JWK set", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseHeader(r.Header, "Authorization", jwt.WithKeySet(keyset))
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to parse jwt", slog.Any("error", err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		slog.InfoContext(r.Context(), "Authenticated request", slog.Any("token", token.Keys()))

		ctx := auth.SetClaimsToContext(r.Context(), &auth.Claims{Token: token})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
