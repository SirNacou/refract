package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/SirNacou/refract/services/api/internal/application"
	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/auth"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/postgres"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/safebrowsing"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/server/middleware"
	"github.com/SirNacou/refract/services/api/migrations"
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	dbPool, err := postgres.NewPool(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// Run migrations before starting server
	if cfg.Database.RunMigrations {
		if err := postgres.RunMigrations(
			dbPool,
			migrations.PostgresFS,
			"postgres",
		); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	store := postgres.NewSQLStore(dbPool)

	cache, err := cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer cache.Close()

	sb, err := safebrowsing.NewGoogleSafeBrowsing(cfg.Security.SafeBrowsingAPIKey, cfg.Redis.GetRedisAddr())
	if err != nil {
		log.Fatalf("Failed to initialize SafeBrowsing: %v", err)
	}

	generator, err := idgen.NewSnowflakeGenerator(int64(cfg.Worker.WorkerID))
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake generator: %v", err)
	}

	// Create application service with all dependencies
	app := application.NewApplication(generator, sb, store, cache)

	authZ := auth.NewAuth(&cfg.Appwrite)
	authMw := middleware.NewAuthMiddleware(authZ)

	rateLimiter := middleware.NewRateLimiter(
		cache,
		&cfg.Security,
	)
	loggingMiddleware := middleware.NewLoggingMiddleware(slog.Default())

	router := server.NewRouter(
		authMw,
		rateLimiter,
		loggingMiddleware,
		&cfg.Security,
		app,
	)

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting API server on %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
