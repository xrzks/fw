package watcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/xrzks/fw/internal/ignore"
)

const debounceMs = 500

type WatchOptions struct {
	Path       string
	Commands   []string
	Extensions []string
	Ignorer    *ignore.Matcher
	Logger     *log.Logger
	FailFast   bool

	normalizedExtensions map[string]bool
}

func Watch(ctx context.Context, opts WatchOptions) error {
	path := opts.Path
	commands := opts.Commands
	extensions := opts.Extensions
	ignorer := opts.Ignorer
	logger := opts.Logger
	failFast := opts.FailFast

	normalizedExts := normalizeExtensions(extensions)
	logger.Debugf("Starting watcher with path: %s, debounce: %dms", path, debounceMs)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

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

	var running atomic.Bool
	debouncer := NewDebouncer(ctx, time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		logger.Debugf("Debounce settled, executing callback for: %s (%v)", event.Name, event.Op)
		running.Store(true)
		defer running.Store(false)
		if err := runCommands(ctx, commands, event, failFast); err != nil {
			logger.Errorf("Command failed: %v", err)
		}
	})
	defer debouncer.Stop()

	done := make(chan struct{})
	errCh := make(chan error, 1)
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
				if !running.Load() {
					logger.Debugf("Received event: %s (%v)", event.Name, event.Op)
				}

				eventRel, err := filepath.Rel(path, event.Name)
				if err != nil {
					logger.Warnf("Failed to compute relative path for %q: %v", event.Name, err)
					eventRel = event.Name
				}

				if event.Has(fsnotify.Create) {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
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

				if !matchExtension(event.Name, normalizedExts) {
					logger.Debugf("Skipping event: extension filter did not match %s", event.Name)
					continue
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
				watcher.Close()
				errCh <- err
				return
			}
		}
	}()

	<-done
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
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

func normalizeExtensions(extensions []string) map[string]bool {
	extSet := make(map[string]bool, len(extensions))
	for _, e := range extensions {
		ext := strings.ToLower(e)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = true
	}
	return extSet
}

func matchExtension(name string, extSet map[string]bool) bool {
	if len(extSet) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	return extSet[ext]
}

func runCommands(ctx context.Context, commands []string, event fsnotify.Event, failFast bool) error {
	if len(commands) == 0 {
		fmt.Printf("\n--- change detected: %s (%v) --- (no commands configured)\n", event.Name, event.Op)
		return nil
	}

	fmt.Printf("\n--- change detected: %s (%v) ---\n", event.Name, event.Op)
	var lastErr error
	for i, cmd := range commands {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("[%d/%d] running: %s\n", i+1, len(commands), cmd)

		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Printf("[%d/%d] failed: %v\n", i+1, len(commands), err)
			if failFast {
				return fmt.Errorf("command %q: %w", cmd, err)
			}
			lastErr = fmt.Errorf("command %q: %w", cmd, err)
		} else {
			fmt.Printf("[%d/%d] done\n", i+1, len(commands))
		}
	}
	return lastErr
}
