package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type Router struct {
	api huma.API
	cfg *config.Config
	srv *http.Server
}

func NewRouter(cfg *config.Config) (*Router, error) {
	router := chi.NewRouter()

	router.Use(
		chiMw.RealIP,
		slogchi.New(slog.Default()),
		chiMw.Recoverer,
		cors.AllowAll().Handler,
	)

	router.Get("/docs", handleDocs)

	huma.DefaultArrayNullable = false

	api := humachi.New(router, getHumaCfg(cfg))

	return &Router{
		api,
		cfg,
		&http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}}, nil
}

func (r *Router) SetUp(ctx context.Context, db *persistence.DB, valkey valkeyaside.CacheAsideClient) (err error) {

	grp := huma.NewGroup(r.api, "/api")

	authMw, err := middleware.NewAuthMiddleware(ctx, grp, r.cfg.JwksURL)
	if err != nil {
		return err
	}

	grp.UseMiddleware(authMw.HandlerHuma)

	if err = urls.NewModule(db, valkey, r.cfg).RegisterRoutes(grp); err != nil {
		return err
	}

	return nil
}

func (r *Router) Run() error {
	return r.srv.ListenAndServe()
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.srv.Shutdown(ctx)
}

func getHumaCfg(cfg *config.Config) huma.Config {
	humaCfg := huma.DefaultConfig("Refract API", "1.0.0")
	humaCfg.DocsPath = ""
	humaCfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	humaCfg.Servers = []*huma.Server{
		{URL: fmt.Sprintf("http://localhost:%d", cfg.Port)},
	}
	return humaCfg
}

func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(`<!doctype html>
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

	if err != nil {
		slog.Info("Failed to write docs response", slog.Any("error", err))
	}
}
