package config

import (
	"context"
	"errors"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Port        uint16 `env:"BROWSERFUL_PORT,default=8080"`
	SessionsDir string `env:"BROWSERFUL_SESSIONS_DIR,default=data/sessions"`
}

func Load() (*Config, error) {
	var cfg Config

	err := envconfig.Process(context.Background(), &cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.SessionsDir == "" {
		return errors.New("sessions dir cannot be empty")
	}

	if c.Port == 0 {
		return errors.New("port cannot be 0")
	}

	return nil
}
