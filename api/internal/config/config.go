package config

import "github.com/caarlos0/env/v11"

type Config struct {
	NodeID         int64  `env:"NODE_ID" envDefault:"0"`
	DefaultDomain  string `env:"DEFAULT_DOMAIN,required"`
	Port           int    `env:"PORT,required"`
	RedirectorPort int    `env:"REDIRECTOR_PORT,required"`
	JwksURL        string `env:"JWKS_URL,required"`
	DatabaseURL    string `env:"DATABASE_URL,required"`

	Valkey ValkeyConfig `envPrefix:"VALKEY_"`
}

type ValkeyConfig struct {
	Host        string `env:"HOST,required"`
	Port        int    `env:"PORT,required"`
	RedirectKey string `env:"REDIRECT_KEY,required"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
