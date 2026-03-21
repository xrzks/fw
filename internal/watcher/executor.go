package watcher

import (
	"fmt"
	"os/exec"
)

type Executor struct {
	command string
}

func NewExecutor(cmd string) *Executor {
	return &Executor{command: cmd}
}

func (e *Executor) Execute() error {
	if e.command == "" {
		return nil
	}

	fmt.Println("change detected, running: " + e.command)
	output, err := exec.Command("sh", "-c", e.command).CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}
