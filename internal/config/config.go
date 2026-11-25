package config

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/sethvargo/go-envconfig"
)

const defaultConfigFileName = "config.json"

type Config struct {
	Port            uint16   `env:"BROWSERFULL_PORT,default=8080"`
	AllowedOrigins  []string `env:"BROWSERFULL_ALLOWED_ORIGINS"`
	BrowserExecPath string   `env:"BROWSERFULL_BROWSER_EXECUTABLE_PATH"`
	Headed          bool     `env:"BROWSERFULL_DEFAULT_HEADED"`
	DataDir         string   `env:"BROWSERFULL_DATA_DIR,default=$HOME/.browserfull"`
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
