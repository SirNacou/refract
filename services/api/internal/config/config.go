package config

import (
	"fmt"

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
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}
