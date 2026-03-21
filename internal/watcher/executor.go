package watcher

import (
	"fmt"
	"os/exec"

	"github.com/fsnotify/fsnotify"
)

type Executor struct {
	commands []string
}

func NewExecutor(commands []string) *Executor {
	return &Executor{commands: commands}
}

func (e *Executor) Execute(event fsnotify.Event) error {
	if len(e.commands) == 0 {
		return nil
	}

	fmt.Printf("change detected: %s (%v)\n", event.Name, event.Op)
	var lastErr error
	for _, cmd := range e.commands {
		fmt.Println("running: " + cmd)
		output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Println(err)
			lastErr = err
		}
		fmt.Println(string(output))
	}
	return lastErr
}
