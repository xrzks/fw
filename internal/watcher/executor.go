package watcher

import (
	"fmt"
	"os/exec"
)

type Executor struct {
	commands []string
}

func NewExecutor(commands []string) *Executor {
	return &Executor{commands: commands}
}

func (e *Executor) Execute() error {
	if len(e.commands) == 0 {
		return nil
	}

	fmt.Println("change detected")
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
