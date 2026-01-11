package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/postgres"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/middleware"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := postgres.NewDBConnection(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Pool.Close()

	redis, err := cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer redis.Close()

	zitadelProvider, err := auth.NewZitadelProvider(ctx, cfg.Zitadel.URL, cfg.Zitadel.ClientID)
	if err != nil {
		log.Fatalf("Failed to create Zitadel provider: %v", err)
	}

	authMiddleware := middleware.NewAuthMiddleware(zitadelProvider)
	rateLimiter := middleware.NewRateLimiter(
		redis.Client(),
		&cfg.Security,
		slog.Default(),
	)

	router := server.NewRouter(authMiddleware,
		rateLimiter,
		&cfg.Security,
	)

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting API server on %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
