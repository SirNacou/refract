package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	// Start health server in a separate goroutine
	healthServer := &http.Server{
		Addr: fmt.Sprintf(":%v", cfg.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			} else {
				http.NotFound(w, r)
			}
		}),
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		log.Println("Health server starting on :8080")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	})

	err = clicksWorker.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start Worker: %v", err)
	}

	<-ctx.Done()

	log.Println("Shutting down worker and health server...")

	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := clicksWorker.Stop(stopCtx); err != nil {
		log.Fatalf("Failed to stop Worker: %v", err)
	}

	// Shutdown health server
	healthShutdownCtx, healthCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer healthCancel()
	if err := healthServer.Shutdown(healthShutdownCtx); err != nil {
		log.Printf("Error shutting down health server: %v", err)
	}

	wg.Wait()
}
