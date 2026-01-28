package config

import "github.com/caarlos0/env/v11"

type Config struct {
	NodeID        int64  `env:"NODE_ID" envDefault:"0"`
	DefaultDomain string `env:"DEFAULT_DOMAIN"`
	Port          int    `env:"PORT"`
	JwksURL       string `env:"JWKS_URL"`
	DatabaseURL   string `env:"DATABASE_URL"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
