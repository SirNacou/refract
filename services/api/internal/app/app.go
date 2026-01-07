package app

import (
	"context"
	"log/slog"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Application encapsulates all application dependencies and lifecycle
type Application struct {
	cfg    *config.Config
	logger *slog.Logger

	// Infrastructure
	dbPool *pgxpool.Pool
	server *Server
}

// New creates and initializes the application with all dependencies
func New(cfg *config.Config, logger *slog.Logger) (*Application, error) {
	logger.Info("Starting Refract API",
		slog.String("port", cfg.Port),
		slog.Int64("node_id", cfg.SnowflakeNodeID),
	)

	app := &Application{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize dependencies in order
	if err := app.initDatabase(context.Background()); err != nil {
		return nil, err
	}

	if err := app.initServer(); err != nil {
		app.cleanup() // Clean up partial initialization
		return nil, err
	}

	logger.Info("Application ready",
		slog.Int("allowed_domains", len(cfg.GetAllowedDomains())),
		slog.Int("cors_origins", len(cfg.GetCORSOrigins())),
	)

	return app, nil
}

// Run starts the application and blocks until shutdown
func (a *Application) Run() error {
	return a.server.Run()
}

// Shutdown gracefully stops the application
func (a *Application) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// cleanup releases resources (called on init failure)
func (a *Application) cleanup() {
	if a.dbPool != nil {
		a.dbPool.Close()
	}
}
