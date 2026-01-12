package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Redis
	REDIS_HOST     string `env:"REDIS_HOST" envDefault:"localhost"`
	REDIS_PORT     string `env:"REDIS_PORT" envDefault:"6379"`
	REDIS_PASSWORD string `env:"REDIS_PASSWORD" envDefault:""`
	REDIS_DB       int    `env:"REDIS_DB" envDefault:"0"`

	// Stream
	REDIS_STREAM_KEY string `env:"REDIS_STREAM_KEY" envDefault:"refract:click_events"`

	// Consumer
	ANALYTICS_CONSUMER_GROUP string `env:"ANALYTICS_CONSUMER_GROUP" envDefault:"analytics-processor"`
	ANALYTICS_CONSUMER_NAME  string `env:"ANALYTICS_CONSUMER_NAME"`
	ANALYTICS_BATCH_SIZE     int64  `env:"ANALYTICS_BATCH_SIZE" envDefault:"100"`
	ANALYTICS_BLOCK_MS       int64  `env:"ANALYTICS_BLOCK_MS" envDefault:"1000"`
	ANALYTICS_STREAM_START   string `env:"ANALYTICS_STREAM_START" envDefault:"$"`

	// Retry backoff
	ANALYTICS_RETRY_MIN_BACKOFF_MS int64 `env:"ANALYTICS_RETRY_MIN_BACKOFF_MS" envDefault:"250"`
	ANALYTICS_RETRY_MAX_BACKOFF_MS int64 `env:"ANALYTICS_RETRY_MAX_BACKOFF_MS" envDefault:"5000"`

	// Database
	ANALYTICS_DATABASE_URL string `env:"ANALYTICS_DATABASE_URL" envDefault:"postgres://refract:password@localhost:5432/url_shortener?sslmode=disable"`

	// Logging
	LOG_LEVEL  string `env:"LOG_LEVEL" envDefault:"info"`
	LOG_FORMAT string `env:"LOG_FORMAT" envDefault:"json"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	// Default consumer name to hostname if not set
	if c.ANALYTICS_CONSUMER_NAME == "" {
		hostname, err := os.Hostname()
		if err != nil {
			c.ANALYTICS_CONSUMER_NAME = "processor-1"
		} else {
			c.ANALYTICS_CONSUMER_NAME = hostname
		}
	}

	return &c, nil
}

func (c *Config) GetRedisAddress() string {
	return c.REDIS_HOST + ":" + c.REDIS_PORT
}

func (c *Config) GetRedisURL() string {
	if c.REDIS_PASSWORD == "" {
		return fmt.Sprintf("redis://%s:%s/%d", c.REDIS_HOST, c.REDIS_PORT, c.REDIS_DB)
	}
	return fmt.Sprintf("redis://:%s@%s:%s/%d", c.REDIS_PASSWORD, c.REDIS_HOST, c.REDIS_PORT, c.REDIS_DB)
}
