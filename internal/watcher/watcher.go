package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/ignore"
)

const debounceMs = 500

func Watch(ctx context.Context, path string, commands []string, extensions []string, ignorer *ignore.Matcher, logger *log.Logger) error {
	logger.Debugf("Starting watcher with path: %s, debounce: %dms", path, debounceMs)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	logger.Debug("Created fsnotify watcher")

	pt, err := getPathType(path)
	if err != nil {
		return err
	}

	var watchedCount int
	if pt == "file" {
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("failed to watch path %q: %w", path, err)
		}
		watchedCount = 1
		logger.Infof("Watching %s: %s", pt, path)
	} else {
		watchedCount, err = addWatchDir(watcher, path, ignorer, logger)
		if err != nil {
			return fmt.Errorf("failed to watch path %q: %w", path, err)
		}
		logger.Infof("Watching directory: %s (%d directories)", path, watchedCount)
	}

	debouncer := NewDebouncer(ctx, time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		logger.Debugf("Debounce settled, executing callback for: %s (%v)", event.Name, event.Op)
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
				if event.Has(fsnotify.Create) {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						eventRel, err := filepath.Rel(path, event.Name)
						if err != nil {
							logger.Warnf("Failed to compute relative path for %q: %v", event.Name, err)
							eventRel = event.Name
						}
						if ignorer.Match(eventRel) {
							logger.Debugf("Skipping ignored new directory %q", eventRel)
							continue
						}
						count, err := addWatchDir(watcher, event.Name, ignorer, logger)
						if err != nil {
							logger.Errorf("Failed to watch new directory %q: %v", event.Name, err)
						} else {
							logger.Debugf("Watching new directory: %s (%d subdirectories)", event.Name, count)
						}
					}
				}
				if !matchExtension(event.Name, extensions) {
					logger.Debugf("Skipping event: extension filter did not match %s", event.Name)
					continue
				}
				eventRel, err := filepath.Rel(path, event.Name)
				if err != nil {
					logger.Warnf("Failed to compute relative path for %q: %v", event.Name, err)
					eventRel = event.Name
				}
				if ignorer.Match(eventRel) {
					logger.Debugf("Skipping event: ignore pattern matched %s", eventRel)
					continue
				}
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

func addWatchDir(watcher *fsnotify.Watcher, root string, ignorer *ignore.Matcher, logger *log.Logger) (int, error) {
	var count int
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Debugf("Skipping inaccessible path %q: %v", path, err)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr == nil && rel != "." && ignorer.Match(rel) {
			logger.Debugf("Skipping ignored directory %q", rel)
			return filepath.SkipDir
		}
		if err := watcher.Add(path); err != nil {
			logger.Debugf("Skipping unwatchable directory %q: %v", path, err)
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func matchExtension(name string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	ext := filepath.Ext(name)
	for _, e := range extensions {
		if strings.EqualFold(ext, e) || strings.EqualFold(ext, "."+e) {
			return true
		}
	}
	return false
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
