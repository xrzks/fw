package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/logger"
)

func Watch(ctx context.Context, path string, commands []string, debounceMs int, debug bool) error {
	log := logger.New(debug)

	log.Debug("Starting watcher with path: %s, debounce: %dms, debug: %v", path, debounceMs, debug)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Close()
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	log.Debug("Created fsnotify watcher")

	if err := watcher.Add(path); err != nil {
		log.Close()
		return fmt.Errorf("failed to watch path %q: %w", path, err)
	}

	log.Debug("Added path to watcher: %s", path)

	if err := printPathStatus(path, debounceMs, log); err != nil {
		log.Close()
		return err
	}

	debouncer := NewDebouncer(ctx, time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		log.Debug("Triggering debouncer for event: %s (%v)", event.Name, event.Op)
		if err := runCommands(ctx, commands, event, log); err != nil {
			log.Error("Command failed: %v", err)
		}
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping watcher...")
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Debug("Received event: %s (%v)", event.Name, event.Op)
				debouncer.Trigger(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Watcher error: %v", err)
				return
			}
		}
	}()

	<-done
	debouncer.Stop()
	log.Close()
	return nil
}

func printPathStatus(path string, debounceMs int, log *logger.Logger) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}
	pathType := "directory"
	if info.Mode().IsRegular() {
		pathType = "file"
	}
	log.Info("Watching %s: %s (debounce: %dms)", pathType, path, debounceMs)
	return nil
}

func runCommands(ctx context.Context, commands []string, event fsnotify.Event, log *logger.Logger) error {
	log.Debug("Change detected: %s (%v)", event.Name, event.Op)
	if len(commands) == 0 {
		fmt.Printf("\n--- change detected: %s (%v) ---\n", event.Name, event.Op)
		log.Info("No commands configured to run for this change")
		return nil
	}

	var lastErr error
	for i, cmd := range commands {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Debug("Running command %d/%d: %s", i+1, len(commands), cmd)
		fmt.Printf("[%d/%d] running: %s\n", i+1, len(commands), cmd)

		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			log.Error("Command %d/%d failed: %v", i+1, len(commands), err)
			fmt.Printf("[%d/%d] failed: %v\n", i+1, len(commands), err)
			lastErr = fmt.Errorf("command %q: %w", cmd, err)
		} else {
			log.Debug("Command %d/%d completed successfully", i+1, len(commands))
			fmt.Printf("[%d/%d] done\n", i+1, len(commands))
		}
	}
	return lastErr
}
