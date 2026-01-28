package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/server"
	"github.com/SirNacou/refract/api/internal/infrastructure/snowflake"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	err = snowflake.NewSnowflakeNode(cfg.NodeID)
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake ID: %v", err)
	}

	db, err := persistence.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	router, err := server.NewRouter(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize router: %v", err)
	}

	if err := router.SetUp(db); err != nil {
		log.Fatalf("Failed to set up routes: %v", err)
	}

	log.Printf("Starting server on port %d", cfg.Port)

	if err := router.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	slog.Info("Server shut down")
}
