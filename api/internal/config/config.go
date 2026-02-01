package config

import "github.com/caarlos0/env/v11"

type Config struct {
	NodeID         int64  `env:"NODE_ID" envDefault:"0"`
	DefaultBaseURL string `env:"DEFAULT_BASE_URL,required"`
	Port           int    `env:"PORT,required"`
	RedirectorPort int    `env:"REDIRECTOR_PORT,required"`
	JwksURL        string `env:"JWKS_URL,required"`
	DatabaseURL    string `env:"DATABASE_URL,required"`

	Valkey ValkeyConfig `envPrefix:"VALKEY_"`

	ClickHouse ClickHouseConfig `envPrefix:"CLICKHOUSE_"`
}

type ValkeyConfig struct {
	Host            string `env:"HOST,required"`
	Port            int    `env:"PORT,required"`
	RedirectKey     string `env:"REDIRECT_KEY,required"`
	ClicksStreamKey string `env:"CLICKS_STREAM_KEY,required"`
	ReadGroup       string `env:"READ_GROUP,required"`
	Consumer        string `env:"CONSUMER,required"`
	BatchSize       int    `env:"BATCH_SIZE" envDefault:"100"`
}

type ClickHouseConfig struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	Database string `env:"DATABASE_NAME,required"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
