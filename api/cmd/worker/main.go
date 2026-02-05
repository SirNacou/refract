package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirNacou/refract/api/internal/config"
	ingestclicks "github.com/SirNacou/refract/api/internal/features/clicks/ingest_clicks"
	"github.com/SirNacou/refract/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/api/internal/infrastructure/clickhouse"
	"github.com/SirNacou/refract/api/internal/infrastructure/worker"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	valkey, err := cache.NewCache(ctx, &cfg.Valkey)
	if err != nil {
		log.Fatalf("Failed to initialize Valkey cache: %v", err)
	}
	defer valkey.Close()

	chClient, err := clickhouse.NewClient(&cfg.ClickHouse)
	if err != nil {
		log.Fatalf("Failed to initialize ClickHouse: %v", err)
	}
	defer chClient.Close()

	handler := ingestclicks.NewCommandHandler(chClient)

	clicksWorker, err := worker.NewClicksStreamWorker(ctx, valkey.Client(), handler, &cfg.Valkey)
	if err != nil {
		log.Fatalf("Failed to initialize Worker: %v", err)
	}

	err = clicksWorker.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start Worker: %v", err)
	}

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := clicksWorker.Stop(stopCtx); err != nil {
		log.Fatalf("Failed to stop Worker: %v", err)
	}
}
