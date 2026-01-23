package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SirNacou/refract/services/analytics-processor/internal/config"
	"github.com/SirNacou/refract/services/analytics-processor/internal/geo"
	"github.com/SirNacou/refract/services/analytics-processor/internal/processor"
	"github.com/SirNacou/refract/services/analytics-processor/internal/redis"
	"github.com/SirNacou/refract/services/analytics-processor/internal/repository"
	"github.com/SirNacou/refract/services/analytics-processor/internal/useragent"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Setup logger from config
	logger := cfg.SetupLogger()
	slog.SetDefault(logger)

	logger.Info("analytics processor starting",
		"stream_key", cfg.REDIS_STREAM_KEY,
		"consumer_group", cfg.ANALYTICS_CONSUMER_GROUP,
		"consumer_name", cfg.ANALYTICS_CONSUMER_NAME,
		"batch_size", cfg.ANALYTICS_BATCH_SIZE,
	)

	// Create Redis valkey
	valkey, err := redis.NewValkeyClient(cfg)
	if err != nil {
		log.Fatalf("failed to initialize valkey client: %v", err)
	}
	defer valkey.Close()

	// Create database connection pool
	db, err := pgxpool.New(ctx, cfg.ANALYTICS_DATABASE_URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo, err := repository.NewPostgresRepository(db)
	if err != nil {
		log.Fatalf("failed to initialize repo: %v", err)
	}

	logger.Info("database connection established", "url", cfg.ANALYTICS_DATABASE_URL)

	// Create GeoIP lookup (optional fallback enrichment)
	geoLookup, err := geo.NewGeoLookup(cfg.GEOIP_DB_PATH)
	if err != nil {
		logger.Warn("failed to load GeoIP database, geo enrichment disabled",
			"path", cfg.GEOIP_DB_PATH,
			"error", err,
		)
		geoLookup = nil // continue without geo enrichment
	} else {
		defer geoLookup.Close()
		logger.Info("GeoIP database loaded", "path", cfg.GEOIP_DB_PATH)
	}

	uap := useragent.NewUserAgentParser()

	pro := processor.NewProcessor(valkey, repo, geoLookup, uap, cfg)

	// Context with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	// Run consumer loop
	logger.Info("starting consumer loop...")

	if err := pro.Run(ctx); err != nil {
		logger.Error("processor encountered an error", "error", err)
	}

	logger.Info("analytics processor stopped gracefully")
}
