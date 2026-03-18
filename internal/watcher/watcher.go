package watcher

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	path string
}

func NewWatcher(path string) *Watcher {
	return &Watcher{
		path: path,
	}
}

func (w *Watcher) Watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(w.path)
	if err != nil {
		return fmt.Errorf("failed to watch path: %w", err)
	}

	fmt.Printf("Watching directory: %s\n", w.path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			fmt.Printf("Change detected: %s\n", event.Name)
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			log.Printf("Watcher error: %v", err)
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}
