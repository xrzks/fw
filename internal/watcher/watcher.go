package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

const debounceMs = 500

func Watch(ctx context.Context, path string, commands []string, logger *log.Logger) error {
	logger.Debugf("Starting watcher with path: %s, debounce: %dms", path, debounceMs)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	logger.Debug("Created fsnotify watcher")

	if err := watcher.Add(path); err != nil {
		return fmt.Errorf("failed to watch path %q: %w", path, err)
	}

	logger.Debugf("Added path to watcher: %s", path)

	pt, err := getPathType(path)
	if err != nil {
		return err
	}
	logger.Infof("Watching %s: %s", pt, path)

	debouncer := NewDebouncer(ctx, time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		logger.Debugf("Triggering debouncer for event: %s (%v)", event.Name, event.Op)
		if err := runCommands(ctx, commands, event, logger); err != nil {
			logger.Errorf("Command failed: %v", err)
		}
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping watcher...")
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Debugf("Received event: %s (%v)", event.Name, event.Op)
				debouncer.Trigger(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("Watcher error: %v", err)
				return
			}
		}
	}()

	<-done
	debouncer.Stop()
	return nil
}

func getPathType(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}
	if info.Mode().IsRegular() {
		return "file", nil
	}
	return "directory", nil
}

func runCommands(ctx context.Context, commands []string, event fsnotify.Event, logger *log.Logger) error {
	logger.Debugf("Change detected: %s (%v)", event.Name, event.Op)
	if len(commands) == 0 {
		fmt.Printf("\n--- change detected: %s (%v) ---\n", event.Name, event.Op)
		logger.Info("No commands configured to run for this change")
		return nil
	}

	var lastErr error
	for i, cmd := range commands {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logger.Debugf("Running command %d/%d: %s", i+1, len(commands), cmd)
		fmt.Printf("[%d/%d] running: %s\n", i+1, len(commands), cmd)

		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			logger.Errorf("Command %d/%d failed: %v", i+1, len(commands), err)
			fmt.Printf("[%d/%d] failed: %v\n", i+1, len(commands), err)
			lastErr = fmt.Errorf("command %q: %w", cmd, err)
		} else {
			logger.Debugf("Command %d/%d completed successfully", i+1, len(commands))
			fmt.Printf("[%d/%d] done\n", i+1, len(commands))
		}
	}
	return lastErr
}
