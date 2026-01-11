package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the API service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Zitadel  ZitadelConfig
	Worker   WorkerConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port    int    `env:"API_PORT" envDefault:"8080"`
	BaseURL string `env:"API_BASE_URL" envDefault:"http://localhost:8080"`
	Host    string `env:"API_HOST" envDefault:"0.0.0.0"`
}

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	Host            string        `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port            int           `env:"POSTGRES_PORT" envDefault:"5432"`
	User            string        `env:"POSTGRES_USER" envDefault:"refract"`
	Password        string        `env:"POSTGRES_PASSWORD" envDefault:"password"`
	Database        string        `env:"POSTGRES_DB" envDefault:"url_shortener"`
	SSLMode         string        `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"POSTGRES_CONN_MAX_IDLE_TIME" envDefault:"5m"`

	// DatabaseURL takes precedence if set (for DATABASE_URL env var)
	DatabaseURL string `env:"DATABASE_URL"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`

	// Cache configuration
	CacheTTL        time.Duration `env:"REDIS_CACHE_TTL" envDefault:"1h"`
	MaxRetries      int           `env:"REDIS_MAX_RETRIES" envDefault:"3"`
	PoolSize        int           `env:"REDIS_POOL_SIZE" envDefault:"10"`
	MinIdleConns    int           `env:"REDIS_MIN_IDLE_CONNS" envDefault:"2"`
	ConnMaxIdleTime time.Duration `env:"REDIS_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}

// ZitadelConfig holds Zitadel OIDC provider configuration
type ZitadelConfig struct {
	URL          string `env:"ZITADEL_URL" envDefault:"https://zitadel.nacou.uk"`
	ClientID     string `env:"ZITADEL_CLIENT_ID"`
	ClientSecret string `env:"ZITADEL_CLIENT_SECRET"`
	Issuer       string `env:"ZITADEL_ISSUER" envDefault:"https://zitadel.nacou.uk"`

	// JWT validation
	JWTIssuer string `env:"JWT_ISSUER" envDefault:"https://zitadel.nacou.uk"`
}

// WorkerConfig holds Snowflake ID generator configuration
type WorkerConfig struct {
	// WorkerID must be unique per API service instance (0-1023)
	// API service should use range 0-63
	// Redirector service uses range 64-127
	WorkerID int `env:"WORKER_ID" envDefault:"0"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// Safe Browsing API key for malicious URL detection
	SafeBrowsingAPIKey string `env:"SAFE_BROWSING_API_KEY"`

	// Rate limiting
	RateLimitPerUser   int           `env:"RATE_LIMIT_PER_USER" envDefault:"100"`
	RateLimitPerAPIKey int           `env:"RATE_LIMIT_PER_API_KEY" envDefault:"1000"`
	RateLimitWindow    time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1h"`

	// CORS
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173,http://localhost:3000"`
	CORSAllowedMethods []string `env:"CORS_ALLOWED_METHODS" envSeparator:"," envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	CORSAllowedHeaders []string `env:"CORS_ALLOWED_HEADERS" envSeparator:"," envDefault:"Authorization,Content-Type,X-API-Key"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"` // json or text
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks if configuration values are valid
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.BaseURL == "" {
		return fmt.Errorf("API_BASE_URL is required")
	}

	// Validate database configuration
	if c.Database.DatabaseURL == "" {
		// If DATABASE_URL not set, validate individual fields
		if c.Database.Host == "" {
			return fmt.Errorf("POSTGRES_HOST is required")
		}
		if c.Database.User == "" {
			return fmt.Errorf("POSTGRES_USER is required")
		}
		if c.Database.Database == "" {
			return fmt.Errorf("POSTGRES_DB is required")
		}
	}

	// Validate Redis configuration
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}

	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid Redis port: %d (must be 1-65535)", c.Redis.Port)
	}

	// Validate Zitadel configuration
	if c.Zitadel.URL == "" {
		return fmt.Errorf("ZITADEL_URL is required")
	}

	if c.Zitadel.ClientID == "" {
		return fmt.Errorf("ZITADEL_CLIENT_ID is required")
	}

	// ClientSecret is optional (not needed for JWT validation, only for token introspection)
	// if c.Zitadel.ClientSecret == "" {
	// 	return fmt.Errorf("ZITADEL_CLIENT_SECRET is required")
	// }

	if c.Zitadel.Issuer == "" {
		return fmt.Errorf("ZITADEL_ISSUER is required")
	}

	// Validate worker ID (must be 0-1023 for Snowflake IDs)
	if c.Worker.WorkerID < 0 || c.Worker.WorkerID > 1023 {
		return fmt.Errorf("invalid WORKER_ID: %d (must be 0-1023)", c.Worker.WorkerID)
	}

	// API service should use worker IDs 0-63 (warning, not error)
	if c.Worker.WorkerID > 63 {
		// Note: This is just a convention, not a hard requirement
		// Redirector uses 64-127, but API can technically use any 0-1023
	}

	// Validate logging configuration
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid LOG_LEVEL: %s (must be debug, info, warn, error, or fatal)", c.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("invalid LOG_FORMAT: %s (must be json or text)", c.Logging.Format)
	}

	return nil
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDatabaseDSN() string {
	// If DATABASE_URL is set, use it directly
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	// Otherwise, construct DSN from individual fields
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address in host:port format
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Address returns the server address in host:port format
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
