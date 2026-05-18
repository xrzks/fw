package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Path       string   `toml:"path"`
	Commands   []string `toml:"commands"`
	Extensions []string `toml:"extensions"`
	Ignore     []string `toml:"ignore"`
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

	if gitignorePattern := loadGitignore(filepath.Dir(path)); len(gitignorePattern) > 0 {
		cfg.Ignore = append(cfg.Ignore, gitignorePattern...)
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

func loadGitignore(configDir string) []string {
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
