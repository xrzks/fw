package cli

import (
	"context"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

func TestNewCommand(t *testing.T) {
	cmd := New()

	if cmd.Name != "fw" {
		t.Errorf("expected command name 'fw', got %q", cmd.Name)
	}
	if cmd.Version != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", cmd.Version)
	}
	if cmd.Action == nil {
		t.Error("expected action to be set")
	}
	if len(cmd.Flags) != 4 {
		t.Errorf("expected 4 flags, got %d", len(cmd.Flags))
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

	for _, name := range []string{"command", "c", "config", "C", "debug", "D", "extension", "e"} {
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
