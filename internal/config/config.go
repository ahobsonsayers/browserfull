package config

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/sethvargo/go-envconfig"
)

const defaultConfigFileName = "config.json"

type Config struct {
	Port           uint16   `env:"BROWSERFUL_PORT,default=8080"`
	DataDir        string   `env:"BROWSERFUL_DATA_DIR,default=$HOME/.browserful"`
	AllowedOrigins []string `env:"BROWSERFUL_ALLOWED_ORIGINS"`
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
	if c.Port == 0 {
		return errors.New("port cannot be 0")
	}

	if c.DataDir == "" {
		return errors.New("data dir cannot be empty")
	}

	return nil
}

func (c *Config) ConfigFilePath() string {
	return filepath.Join(c.DataDir, defaultConfigFileName)
}
