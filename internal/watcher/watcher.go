package watcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
)

func Watch(path string, cmd string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		return fmt.Errorf("failed to watch path: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	pathType := "directory"
	if info.Mode().IsRegular() {
		pathType = "file"
	}

	fmt.Printf("Watching %s: %s\n", pathType, path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}
			fmt.Printf("Change detected: %s\n", event.Name)
			if cmd != "" {
				fmt.Println("running: " + cmd)
				output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			log.Printf("Watcher error: %v", err)
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}
