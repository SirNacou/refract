package app

import (
	"context"
	"fmt"

	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/application/queries"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/cache"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/handlers"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/http/router"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/persistence/postgres"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/shortcode"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/validation"
	"github.com/jackc/pgx/v5/pgxpool"
)

// initDatabase connects to the database with helpful error messages
func (a *Application) initDatabase(ctx context.Context) error {
	dbURL := a.cfg.DatabaseURL()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w\n"+
			"  → Check DB_HOST=%s and DB_PORT=%d are correct\n"+
			"  → Ensure PostgreSQL is running\n"+
			"  → Verify DB_USER and DB_PASSWORD are valid",
			err, a.cfg.DBHost, a.cfg.DBPort)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("database connection failed health check: %w\n"+
			"  → Database may be starting up\n"+
			"  → Check if database '%s' exists\n"+
			"  → Verify user '%s' has access permissions",
			err, a.cfg.DBName, a.cfg.DBUser)
	}

	a.dbPool = pool
	return nil
}

// initCache connects to Valkey with graceful fallback to NoopCache
func (a *Application) initCache(ctx context.Context) {
	valkeyCache, err := cache.NewValkeyCache(
		a.cfg.ValkeyHost,
		a.cfg.ValkeyPort,
		a.cfg.ValkeyPassword,
		a.cfg.ValkeyDB,
		a.logger,
	)
	if err != nil {
		a.logger.Warn("Failed to connect to Valkey, using NoopCache",
			"error", err,
			"host", a.cfg.ValkeyHost,
			"port", a.cfg.ValkeyPort,
		)
		a.valkeyCache = cache.NewNoopCache()
		return
	}

	a.logger.Info("Connected to Valkey",
		"host", a.cfg.ValkeyHost,
		"port", a.cfg.ValkeyPort,
		"db", a.cfg.ValkeyDB,
	)
	a.valkeyCache = valkeyCache
}

// initServer builds the HTTP server with all dependencies
func (a *Application) initServer() error {
	// Initialize ID generator
	idGen, err := idgen.NewSnowflakeGenerator(a.cfg.SnowflakeNodeID)
	if err != nil {
		return fmt.Errorf("failed to create ID generator: %w\n"+
			"  → SNOWFLAKE_NODE_ID must be between 0 and 1023\n"+
			"  → Current value: %d",
			err, a.cfg.SnowflakeNodeID)
	}

	// Initialize short code generator
	shortCodeGen, err := shortcode.NewSqidsGenerator(
		a.cfg.SqidsAlphabet,
		a.cfg.SqidsMinLength,
	)
	if err != nil {
		return fmt.Errorf("failed to create short code generator: %w\n"+
			"  → Check SQIDS_ALPHABET contains valid characters\n"+
			"  → Check SQIDS_MIN_LENGTH is reasonable (1-20)",
			err)
	}

	// Initialize domain validator
	domainValidator := validation.NewWhitelistDomainValidator(
		a.cfg.GetAllowedDomains(),
	)

	// Initialize repository with cache decorator
	baseRepo := postgres.NewPostgresURLRepository(a.dbPool)
	repo := persistence.NewCachedURLRepository(baseRepo, a.valkeyCache, a.logger)

	// Initialize command/query handlers
	createURLHandler := commands.NewCreateURLHandler(
		repo,
		shortCodeGen,
		domainValidator,
		idGen,
	)
	getURLHandler := queries.NewGetURLHandler(repo)

	// Initialize HTTP handlers
	urlHandler := handlers.NewURLHandler(
		createURLHandler,
		getURLHandler,
		a.logger,
	)
	healthHandler := handlers.NewHealthHandler(a.dbPool, a.logger)

	// Setup router
	httpHandler := router.NewRouter(router.Config{
		URLHandler:    urlHandler,
		HealthHandler: healthHandler,
		Logger:        a.logger,
		CORSOrigins:   a.cfg.GetCORSOrigins(),
	})

	// Create server
	a.server = NewServer(a.cfg, httpHandler, a.logger)

	return nil
}
