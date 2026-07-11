package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
	"github.com/xrzks/fw/internal/config"
)

func TestNewCommand(t *testing.T) {
	cmd := New()

	if cmd.Name != "fw" {
		t.Errorf("expected command name 'fw', got %q", cmd.Name)
	}
	if cmd.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", cmd.Version)
	}
	if cmd.Action == nil {
		t.Error("expected action to be set")
	}
	if len(cmd.Flags) != 7 {
		t.Errorf("expected 7 flags, got %d", len(cmd.Flags))
	}
}

func TestRunFlags(t *testing.T) {
	cmd := New()

	found := map[string]bool{}
	for _, f := range cmd.Flags {
		for _, name := range f.Names() {
			found[name] = true
		}
	}

	for _, name := range []string{"command", "c", "config", "C", "debug", "D", "extension", "e", "ignore", "i"} {
		if !found[name] {
			t.Errorf("expected flag %q to be defined", name)
		}
	}
}

func TestRunHelp(t *testing.T) {
	cmd := New()

	ctx := context.Background()
	args := []string{"fw", "--help"}

	err := cmd.Run(ctx, args)
	if err != nil {
		t.Errorf("help should not error: %v", err)
	}
}

func TestRunWatchesThenCancel(t *testing.T) {
	cmd := New()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	args := []string{"fw", "."}

	err := cmd.Run(ctx, args)
	if err != nil {
		t.Errorf("expected clean shutdown on context cancel, got: %v", err)
	}
}

func TestHasStringArg(t *testing.T) {
	cmd := New()

	hasPathArg := false
	for _, arg := range cmd.Arguments {
		if a, ok := arg.(*cli.StringArg); ok && a.Name == "path" {
			hasPathArg = true
		}
	}
	if !hasPathArg {
		t.Error("expected 'path' string argument to be defined")
	}
}

func newTestCmd(t *testing.T, flags map[string][]string) *cli.Command {
	t.Helper()
	cmd := New()
	for name, values := range flags {
		for _, v := range values {
			if err := cmd.Set(name, v); err != nil {
				t.Fatalf("failed to set flag %s=%s: %v", name, v, err)
			}
		}
	}
	return cmd
}

func TestResolveCommandsMerge(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"command": {"go test"},
	})
	cfg := &config.Config{Commands: []string{"go build", "go vet"}}

	result := resolveCommands(cmd, cfg)

	if len(result) != 3 {
		t.Fatalf("expected 3 commands, got %d: %v", len(result), result)
	}
	if result[0] != "go build" || result[1] != "go vet" || result[2] != "go test" {
		t.Errorf("expected [go build go vet go test], got %v", result)
	}
}

func TestResolveCommandsConfigOnly(t *testing.T) {
	cmd := newTestCmd(t, nil)
	cfg := &config.Config{Commands: []string{"go build"}}

	result := resolveCommands(cmd, cfg)

	if len(result) != 1 || result[0] != "go build" {
		t.Errorf("expected [go build], got %v", result)
	}
}

func TestResolveCommandsCLIOnly(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"command": {"go test"},
	})

	result := resolveCommands(cmd, nil)

	if len(result) != 1 || result[0] != "go test" {
		t.Errorf("expected [go test], got %v", result)
	}
}

func TestResolveCommandsNone(t *testing.T) {
	cmd := newTestCmd(t, nil)

	result := resolveCommands(cmd, nil)

	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestResolveExtensionsMerge(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"extension": {".go"},
	})
	cfg := &config.Config{Extensions: []string{".ts", ".js"}}

	result := resolveExtensions(cmd, cfg)

	if len(result) != 3 {
		t.Fatalf("expected 3 extensions, got %d: %v", len(result), result)
	}
	if result[0] != ".ts" || result[1] != ".js" || result[2] != ".go" {
		t.Errorf("expected [.ts .js .go], got %v", result)
	}
}

func TestResolveExtensionsConfigOnly(t *testing.T) {
	cmd := newTestCmd(t, nil)
	cfg := &config.Config{Extensions: []string{".ts"}}

	result := resolveExtensions(cmd, cfg)

	if len(result) != 1 || result[0] != ".ts" {
		t.Errorf("expected [.ts], got %v", result)
	}
}

func TestResolveExtensionsCLIOnly(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"extension": {".go"},
	})

	result := resolveExtensions(cmd, nil)

	if len(result) != 1 || result[0] != ".go" {
		t.Errorf("expected [.go], got %v", result)
	}
}

func TestResolveExtensionsNone(t *testing.T) {
	cmd := newTestCmd(t, nil)

	result := resolveExtensions(cmd, nil)

	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestResolveIgnoreMerge(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"ignore": {"*.log"},
	})
	cfg := &config.Config{Ignore: []string{"node_modules", ".git"}}

	matcher, _ := resolveIgnore(cmd, cfg, log.New(os.Stderr))

	if !matcher.Match("node_modules") {
		t.Error("expected config ignore 'node_modules' to match")
	}
	if !matcher.Match("app.log") {
		t.Error("expected CLI ignore '*.log' to match")
	}
	if matcher.Match("main.go") {
		t.Error("expected 'main.go' to not match")
	}
}

func TestResolveIgnoreConfigOnly(t *testing.T) {
	cmd := newTestCmd(t, nil)
	cfg := &config.Config{Ignore: []string{"node_modules"}}

	matcher, _ := resolveIgnore(cmd, cfg, log.New(os.Stderr))

	if !matcher.Match("node_modules") {
		t.Error("expected config ignore to match")
	}
}

func TestResolveIgnoreCLIOnly(t *testing.T) {
	cmd := newTestCmd(t, map[string][]string{
		"ignore": {"*.log"},
	})

	matcher, _ := resolveIgnore(cmd, nil, log.New(os.Stderr))

	if !matcher.Match("app.log") {
		t.Error("expected CLI ignore to match")
	}
	if matcher.Match("node_modules") {
		t.Error("expected 'node_modules' to not match")
	}
}

func TestResolveIgnoreNone(t *testing.T) {
	cmd := newTestCmd(t, nil)

	matcher, _ := resolveIgnore(cmd, nil, log.New(os.Stderr))

	if matcher.Match("anything") {
		t.Error("expected nothing to match with no patterns")
	}
}
