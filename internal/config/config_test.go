package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "fw.toml")

	content := `
path = "./src"
commands = ["npm run build", "npm test"]
extensions = [".js", ".ts"]
debug = true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Path != "./src" {
		t.Errorf("expected path './src', got %q", cfg.Path)
	}
	if len(cfg.Commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cfg.Commands))
	}
	if cfg.Commands[0] != "npm run build" || cfg.Commands[1] != "npm test" {
		t.Errorf("unexpected commands: %v", cfg.Commands)
	}
	if len(cfg.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(cfg.Extensions))
	}
	if !cfg.Debug {
		t.Error("expected debug to be true")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("nonexistent.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "fw.toml")

	if err := os.WriteFile(cfgPath, []byte("not valid toml {{{{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestLoadEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "fw.toml")

	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Path != "" {
		t.Errorf("expected empty path, got %q", cfg.Path)
	}
	if len(cfg.Commands) != 0 {
		t.Errorf("expected no commands, got %d", len(cfg.Commands))
	}
	if cfg.Debug {
		t.Error("expected debug to be false")
	}
}

func TestFindFwToml(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "fw.toml")

	if err := os.WriteFile(cfgPath, []byte(`path = "./src"`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	cfg, err := Find()
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be found")
	}
	if cfg.Path != "./src" {
		t.Errorf("expected path './src', got %q", cfg.Path)
	}
}

func TestFindHiddenFwToml(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".fw.toml")

	if err := os.WriteFile(cfgPath, []byte(`path = "./lib"`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	cfg, err := Find()
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be found")
	}
	if cfg.Path != "./lib" {
		t.Errorf("expected path './lib', got %q", cfg.Path)
	}
}

func TestFindNoConfig(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	cfg, err := Find()
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config when no file exists")
	}
}

func TestFindFwTomlPreferredOverHidden(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "fw.toml"), []byte(`path = "./explicit"`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".fw.toml"), []byte(`path = "./hidden"`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	cfg, err := Find()
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if cfg.Path != "./explicit" {
		t.Errorf("expected fw.toml to take precedence, got %q", cfg.Path)
	}
}
