package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xrzks/fw/internal/watcher"
)

func main() {
	var commands []string
	var debounce int
	flag.IntVar(&debounce, "d", 500, "debounce delay in milliseconds")
	flag.Func("c", "command to run on file changes", func(val string) error {
		commands = append(commands, val)
		return nil
	})
	flag.Parse()

	path := "."
	args := flag.Args()
	if len(args) > 0 {
		path = args[0]
	}

	if err := watcher.Watch(path, commands, debounce); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
