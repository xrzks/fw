package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/logger"
)

func Watch(ctx context.Context, path string, commands []string, debounceMs int, debug bool) error {
	log := logger.New(debug)

	if debug {
		log.Debug("Starting watcher with path: %s, debounce: %dms, debug: %v", path, debounceMs, debug)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Failed to create watcher: %v", err)
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	if debug {
		log.Debug("Created fsnotify watcher")
	}

	err = watcher.Add(path)
	if err != nil {
		log.Error("Failed to watch path %q: %v", path, err)
		return fmt.Errorf("failed to watch path %q: %w", path, err)
	}

	if debug {
		log.Debug("Added path to watcher: %s", path)
	}

	pathInfo, err := NewPathInfo(path, log)
	if err != nil {
		log.Error("Failed to create path info: %v", err)
		return err
	}
	pathInfo.PrintStatus(debounceMs)

	executor := NewExecutor(ctx, commands, log)
	debouncer := NewDebouncer(ctx, time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		if debug {
			log.Debug("Triggering debouncer for event: %s (%v)", event.Name, event.Op)
		}
		if err := executor.Execute(event); err != nil {
			log.Error("Command failed: %v", err)
		}
	})
	defer debouncer.Stop()

	log.Info("Press Ctrl+C to stop")

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping watcher...")
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				log.Error("Watcher events channel closed unexpectedly")
				return fmt.Errorf("watcher events channel closed unexpectedly")
			}
			if debug {
				log.Debug("Received event: %s (%v)", event.Name, event.Op)
			}
			debouncer.Trigger(event)
		case <-debouncer.TimerChannel():
			debouncer.OnTimerFire()
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Error("Watcher errors channel closed unexpectedly")
				return fmt.Errorf("watcher errors channel closed unexpectedly")
			}
			log.Error("Watcher error: %v", err)
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}
