package config

import "github.com/caarlos0/env/v11"

type Config struct {
	NodeID        int64  `env:"NODE_ID" envDefault:"0"`
	DefaultDomain string `env:"DEFAULT_DOMAIN" required:"true"`
	Port          int    `env:"PORT" required:"true"`
	JwksURL       string `env:"JWKS_URL" required:"true"`
	DatabaseURL   string `env:"DATABASE_URL" required:"true"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
