package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Port int `env:"PORT"`
}

func LoadConfig() (*Config, error) {
	c, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
