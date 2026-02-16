package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/api/internal/infrastructure/clickhouse"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/server"
	"github.com/SirNacou/refract/api/internal/infrastructure/snowflake"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Redirect Key: %s", cfg.Valkey.RedirectKey)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	err = snowflake.NewSnowflakeNode(cfg.NodeID)
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake ID: %v", err)
	}

	db, err := persistence.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	valkey, err := cache.NewCache(ctx, &cfg.Valkey)
	if err != nil {
		log.Fatalf("Failed to initialize Valkey: %v", err)
	}
	defer valkey.Close()

	clickhouseConn, err := clickhouse.NewClient(&cfg.ClickHouse)
	if err != nil {
		log.Fatalf("Failed to initialize Clickhouse: %v", err)
	}

	router, err := server.NewRouter(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	if err := router.SetUp(ctx, db, valkey, clickhouseConn); err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	// 2. Run the server in a goroutine so it doesn't block main
	go func() {
		if err := router.Run(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server forced to shutdown", "error", err)
		}
	}()

	slog.Info("Server started", "port", cfg.Port)

	// 3. Wait for the signal
	<-ctx.Done()
	slog.Info("Shutting down gracefully... Press Ctrl+C again to force")

	// 4. Trigger the router's shutdown logic
	// We create a second context with a timeout (e.g., 10 seconds)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := router.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exited properly")
}
