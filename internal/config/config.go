package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Path        string   `toml:"path" json:"path"`
	Commands    []string `toml:"commands" json:"commands"`
	Extensions  []string `toml:"extensions" json:"extensions"`
	Ignore      []string `toml:"ignore" json:"ignore"`
	Debug       bool     `toml:"debug" json:"debug"`
	NoGitignore bool     `toml:"no-gitignore" json:"noGitignore"`
	FailFast    bool     `toml:"fail-fast" json:"failFast"`
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

func FindAndLoad() (*Config, error) {
	candidates := []string{"fw.toml", ".fw.toml"}
	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return Load(name)
		}
	}
	return nil, nil
}

func LoadGitignore(configDir string) []string {
	data, err := os.ReadFile(filepath.Join(configDir, ".gitignore"))
	if err != nil {
		return []string{}
	}

	var patterns []string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}
