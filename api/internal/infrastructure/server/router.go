package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/features/urls"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/server/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	slogchi "github.com/samber/slog-chi"
)

type Router struct {
	api    huma.API
	router *chi.Mux
	port   int
}

func NewRouter(ctx context.Context, cfg *config.Config) (*Router, error) {
	router := chi.NewRouter()

	router.Use(slogchi.New(slog.Default()))
	router.Use(chiMw.Recoverer)
	router.Use(cors.AllowAll().Handler)

	humaCfg := huma.DefaultConfig("Refract API", "1.0.0")
	humaCfg.DocsPath = ""
	humaCfg.Servers = []*huma.Server{
		{URL: fmt.Sprintf("http://localhost:%d", cfg.Port)},
	}

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

	api := humachi.New(router, humaCfg)

	grp := huma.NewGroup(api, "/api")

	authMw, err := middleware.NewAuthMiddleware(ctx, cfg.JwksURL)
	if err != nil {
		return nil, err
	}

	grp.UseMiddleware(authMw.HandlerHuma)

	huma.Get(api, "/", func(ctx context.Context, i *struct {
		Authorication string `header:"Authorization"`
	}) (*struct{ Body string }, error) {
		return &struct{ Body string }{
			Body: "Welcome to the Refract API!",
		}, nil
	})
	return &Router{grp, router, cfg.Port}, nil
}

func (r *Router) SetUp(db *persistence.DB) (err error) {
	err = urls.NewModule(db).RegisterRoutes(r.api)

	return err
}

func (r *Router) Run() error {
	addr := fmt.Sprintf(":%d", r.port)

	return http.ListenAndServe(addr, r.router)
}
