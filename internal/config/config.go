package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

const defaultConfigFileName = "config.json"

const extensionsDirName = "extensions"

type Config struct {
	Port            uint16   `env:"BROWSERFULL_PORT,default=8080"`
	AllowedOrigins  []string `env:"BROWSERFULL_ALLOWED_ORIGINS"`
	BrowserExecPath string   `env:"BROWSERFULL_BROWSER_EXECUTABLE_PATH"`
	Headed          bool     `env:"BROWSERFULL_DEFAULT_HEADED"`
	Extensions      []string `env:"BROWSERFULL_DEFAULT_EXTENSIONS"`
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

	for _, name := range c.Extensions {
		_, err := c.ExtensionPath(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ConfigFilePath() string {
	return filepath.Join(c.DataDir, defaultConfigFileName)
}

// ExtensionPath returns the directory of an unpacked extension in the data
// dir's extensions library.
func (c *Config) ExtensionPath(name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid extension name '%s'", name)
	}

	dir := filepath.Join(c.DataDir, extensionsDirName, name)

	dirInfo, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("extension '%s' not found: %w", name, err)
	}

	if !dirInfo.IsDir() {
		return "", fmt.Errorf("extension '%s' is not a directory", name)
	}

	manifestPath := filepath.Join(dir, "manifest.json")

	_, err = os.Stat(manifestPath)
	if err != nil {
		return "", fmt.Errorf("extension '%s' has no manifest.json: %w", name, err)
	}

	return dir, nil
}
