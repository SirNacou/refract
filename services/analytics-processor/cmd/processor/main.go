package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SirNacou/refract/services/analytics-processor/internal/config"
	"github.com/SirNacou/refract/services/analytics-processor/internal/consumer"
	"github.com/SirNacou/refract/services/analytics-processor/internal/geo"
	"github.com/SirNacou/refract/services/analytics-processor/internal/redis"
	"github.com/SirNacou/refract/services/analytics-processor/internal/repository"
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

	// Create Redis client
	client, err := redis.NewValkeyClient(cfg)
	if err != nil {
		log.Fatalf("failed to initialize valkey client: %v", err)
	}
	defer client.Close()

	// Create database connection pool
	db, err := pgxpool.New(ctx, cfg.ANALYTICS_DATABASE_URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

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

	// Create repository
	repo := repository.NewTimescaleRepository(db, geoLookup)

	// Create stream consumer
	streamConsumer := consumer.NewStreamConsumer(
		client,
		cfg,
	)

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

	// Ensure consumer group exists
	if err := streamConsumer.EnsureGroup(ctx); err != nil {
		log.Fatalf("failed to ensure consumer group: %v", err)
	}

	// Handler: batch insert click events into TimescaleDB
	handler := func(ctx context.Context, events []consumer.ClickEvent) error {
		return repo.InsertClickEvents(ctx, events)
	}

	// Run consumer loop
	logger.Info("starting consumer loop...")
	if err := streamConsumer.Run(ctx, handler); err != nil && err != context.Canceled {
		log.Fatalf("consumer loop failed: %v", err)
	}

	logger.Info("analytics processor stopped gracefully")
}
