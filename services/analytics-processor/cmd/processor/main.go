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
	"github.com/SirNacou/refract/services/analytics-processor/internal/redis"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

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

	// Create stream consumer
	streamConsumer := consumer.NewStreamConsumer(
		client,
		logger,
		cfg.REDIS_STREAM_KEY,
		cfg.ANALYTICS_CONSUMER_GROUP,
		cfg.ANALYTICS_CONSUMER_NAME,
		cfg.ANALYTICS_BATCH_SIZE,
		cfg.ANALYTICS_BLOCK_MS,
		cfg.ANALYTICS_STREAM_START,
		cfg.ANALYTICS_RETRY_MIN_BACKOFF_MS,
		cfg.ANALYTICS_RETRY_MAX_BACKOFF_MS,
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

	// Placeholder handler (T041 will replace this with DB insert)
	handler := func(ctx context.Context, events []consumer.ClickEvent) error {
		logger.Info("processing batch (placeholder)",
			"count", len(events),
			"first_event_id", events[0].EventID,
			"first_url_id", events[0].URLID,
		)
		// TODO (T041): Insert events into TimescaleDB
		return nil
	}

	// Run consumer loop
	logger.Info("starting consumer loop...")
	if err := streamConsumer.Run(ctx, handler); err != nil && err != context.Canceled {
		log.Fatalf("consumer loop failed: %v", err)
	}

	logger.Info("analytics processor stopped gracefully")
}
