package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/logger"
)

type Executor struct {
	ctx      context.Context
	commands []string
	logger   *logger.Logger
}

func NewExecutor(ctx context.Context, commands []string, logger *logger.Logger) *Executor {
	return &Executor{
		ctx:      ctx,
		commands: commands,
		logger:   logger,
	}
}

func (e *Executor) Execute(event fsnotify.Event) error {
	select {
	case <-e.ctx.Done():
		return e.ctx.Err()
	default:
	}

	e.logger.Debug("Change detected: %s (%v)", event.Name, event.Op)
	if len(e.commands) == 0 {
		fmt.Printf("\n--- change detected: %s (%v) ---\n", event.Name, event.Op)
		e.logger.Info("No commands configured to run for this change")
		return nil
	}

	var lastErr error
	for i, cmd := range e.commands {
		e.logger.Debug("Running command %d/%d: %s", i+1, len(e.commands), cmd)
		fmt.Printf("[%d/%d] running: %s\n", i+1, len(e.commands), cmd)

		select {
		case <-e.ctx.Done():
			return e.ctx.Err()
		default:
		}

		command := exec.CommandContext(e.ctx, "sh", "-c", cmd)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		err := command.Run()
		if err != nil {
			e.logger.Error("Command %d/%d failed: %v", i+1, len(e.commands), err)
			fmt.Printf("[%d/%d] failed: %v\n", i+1, len(e.commands), err)
			lastErr = fmt.Errorf("command %q: %w", cmd, err)
		} else {
			e.logger.Debug("Command %d/%d completed successfully", i+1, len(e.commands))
			fmt.Printf("[%d/%d] done\n", i+1, len(e.commands))
		}
	}

	return lastErr
}
