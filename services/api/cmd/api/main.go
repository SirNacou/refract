package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/application/queries"
	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/handlers"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/router"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/postgres"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/shortcode"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/validation"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize structured logger (JSON format)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	// Load configuration
	logger.Info("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Configuration loaded",
		slog.String("port", cfg.Port),
		slog.Int("allowed_domains", len(cfg.GetAllowedDomains())),
	)

	// Connect to database
	logger.Info("Connecting to database...")
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// Verify database connection
	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("Failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Database connected successfully")

	// Initialize Sqids generator
	logger.Info("Initializing short code generator...")
	generator, err := shortcode.NewSqidsGenerator(cfg.SqidsAlphabet, cfg.SqidsMinLength)
	if err != nil {
		logger.Error("Failed to initialize generator", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Short code generator initialized")

	// Initialize domain validator
	logger.Info("Initializing domain validator...")
	domainValidator := validation.NewWhitelistDomainValidator(cfg.GetAllowedDomains())
	logger.Info("Domain validator initialized",
		slog.Any("allowed_domains", cfg.GetAllowedDomains()),
	)

	// Initialize Snowflake ID generator
	logger.Info("Initializing Snowflake ID generator...")
	snowflakeGen, err := idgen.GetInstance(cfg.SnowflakeNodeID)
	if err != nil {
		logger.Error("Failed to initialize Snowflake generator", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Snowflake ID generator initialized",
		slog.Int64("node_id", cfg.SnowflakeNodeID),
	)

	// Initialize repository
	logger.Info("Initializing repository...")
	repo := postgres.NewPostgresURLRepository(dbPool)
	logger.Info("Repository initialized")

	// Initialize command handlers
	logger.Info("Initializing command handlers...")
	createURLHandler := commands.NewCreateURLHandler(repo, generator, domainValidator, snowflakeGen)
	logger.Info("Command handlers initialized")

	// Initialize query handlers
	logger.Info("Initializing query handlers...")
	getURLHandler := queries.NewGetURLHandler(repo)
	logger.Info("Query handlers initialized")

	// Initialize HTTP handlers
	logger.Info("Initializing HTTP handlers...")
	urlHandler := handlers.NewURLHandler(createURLHandler, getURLHandler, logger)
	healthHandler := handlers.NewHealthHandler(dbPool, logger)
	logger.Info("HTTP handlers initialized")

	// Setup router with middleware
	logger.Info("Setting up router...")
	httpHandler := router.NewRouter(router.Config{
		URLHandler:    urlHandler,
		HealthHandler: healthHandler,
		Logger:        logger,
		CORSOrigins:   cfg.GetCORSOrigins(),
	})
	logger.Info("Router configured")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Starting HTTP server", slog.String("addr", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server failed to start", slog.String("error", err.Error()))
		os.Exit(1)
	case sig := <-quit:
		logger.Info("Received shutdown signal", slog.String("signal", sig.String()))

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Info("Shutting down server gracefully...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown error", slog.String("error", err.Error()))
			os.Exit(1)
		}

		logger.Info("Server stopped gracefully")
	}
}
