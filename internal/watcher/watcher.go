package watcher

import (
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

func Watch(path string, commands []string, debounceMs int) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch path: %w", err)
	}

	pathInfo, err := NewPathInfo(path)
	if err != nil {
		return err
	}
	pathInfo.PrintStatus()

	executor := NewExecutor(commands)
	debouncer := NewDebouncer(time.Duration(debounceMs)*time.Millisecond, func(event fsnotify.Event) {
		if err := executor.Execute(event); err != nil {
			log.Printf("Execution error: %v", err)
		}
	})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			debouncer.Trigger(event)
		case <-debouncer.getTimerChannel():
			debouncer.onTimerFire()
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			log.Printf("Watcher error: %v", err)
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}
