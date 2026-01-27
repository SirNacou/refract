package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/infrastructure/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))

	slog.Debug("Test")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	router, err := server.NewRouter(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	log.Printf("Starting server on port %d", cfg.Port)

	if err := router.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	slog.Info("Server shut down")
}
