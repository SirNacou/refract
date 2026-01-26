package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/rs/cors"
)

type Router struct {
	api    huma.API
	router *chi.Mux
	port   int
}

func NewRouter(port int) *Router {
	router := chi.NewRouter()

	router.Use(cors.AllowAll().Handler)

	humaCfg := huma.DefaultConfig("Refract API", "1.0.0")
	humaCfg.DocsPath = ""
	humaCfg.Servers = []*huma.Server{
		{URL: fmt.Sprintf("http://localhost:%d", port)},
	}

	api := humachi.New(router, humaCfg)

	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		bearer := ctx.Header("Authorization")
		slog.Info("Incoming request", slog.String("method", ctx.Method()), slog.String("path", ctx.URL().Path), slog.String("auth", bearer))

		if bearer == "" {
			slog.Warn("Missing Authorization header")
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized request")
			return
		}

		var tokenString string
		fmt.Sscanf(bearer, "Bearer %s", &tokenString)

		jwksURL := "http://localhost:3000/api/auth/jwks"
		cache, err := jwk.NewCache(ctx.Context(), http.DefaultClient)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		err = cache.Register(ctx.Context(), jwksURL, jwk.WithMinInterval(15*time.Minute))
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		_, err = cache.Refresh(ctx.Context(), jwksURL)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		keyF, err := keyfunc.NewDefaultCtx(ctx.Context(), []string{jwksURL})
		if err != nil {
			slog.Error("Failed to create JWKs from URL", slog.String("url", jwksURL), slog.AnyValue(err))
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized request")
			return
		}

		token, err := jwt.Parse(tokenString, keyF.KeyfuncCtx(ctx.Context()))

		if err != nil || !token.Valid {
			slog.Error("Failed to parse or validate JWT", slog.Any("error", err))
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized request")
			return
		}

		slog.Info("Successfully authenticated request", slog.Any("claims", token.Claims))

		next(ctx)
	})

	router.Get("/docs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!doctype html>
	<html>
	  <head>
	    <title>API Reference</title>
	    <meta charset="utf-8" />
	    <meta
	      name="viewport"
	      content="width=device-width, initial-scale=1" />
	  </head>
	  <body>
	    <script
	      id="api-reference"
	      data-url="/openapi.json"></script>
	    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	  </body>
	</html>`))
	}))

	huma.Get(api, "/", func(ctx context.Context, i *struct {
		Authorication string `header:"Authorization"`
	}) (*struct{ Body string }, error) {
		slog.Info(fmt.Sprintf("token: %v", i.Authorication))

		return &struct{ Body string }{
			Body: "Welcome to the Refract API!",
		}, nil
	})
	return &Router{api, router, port}
}

func (r *Router) Handler() huma.API {
	return r.api
}

func (r *Router) Run() error {
	addr := fmt.Sprintf(":%d", r.port)

	return http.ListenAndServe(addr, r.router)
}
