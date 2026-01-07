package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Database
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"refract"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	// Server
	Port string `env:"PORT" envDefault:"8080"`

	// Application
	SqidsAlphabet  string `env:"SQIDS_ALPHABET" envDefault:"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`
	SqidsMinLength int    `env:"SQIDS_MIN_LENGTH" envDefault:"6"`

	// Snowflake ID Configuration
	SnowflakeNodeID int64 `env:"SNOWFLAKE_NODE_ID" envDefault:"0"`

	// Domain Configuration
	AllowedDomains string `env:"ALLOWED_DOMAINS" envDefault:"short.link"`
	DefaultDomain  string `env:"DEFAULT_DOMAIN" envDefault:"short.link"`

	// CORS Configuration
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate at least one allowed domain exists
	if strings.TrimSpace(cfg.AllowedDomains) == "" {
		return nil, fmt.Errorf("at least one allowed domain must be configured")
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// GetAllowedDomains returns the list of allowed domains
func (c *Config) GetAllowedDomains() []string {
	domains := strings.Split(c.AllowedDomains, ",")
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		if trimmed := strings.TrimSpace(domain); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetCORSOrigins returns the list of allowed CORS origins
func (c *Config) GetCORSOrigins() []string {
	if strings.TrimSpace(c.CORSOrigins) == "" {
		return []string{}
	}
	origins := strings.Split(c.CORSOrigins, ",")
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
