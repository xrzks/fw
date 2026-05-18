package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Path       string   `toml:"path"`
	Commands   []string `toml:"commands"`
	Extensions []string `toml:"extensions"`
	Debug      bool     `toml:"debug"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func Find() (*Config, error) {
	candidates := []string{"fw.toml", ".fw.toml"}
	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return Load(name)
		}
	}
	return nil, nil
}
