package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xrzks/fw/internal/watcher"
)

func main() {
	var cmd string
	var debounce int
	flag.StringVar(&cmd, "c", "", "command to run on file changes")
	flag.IntVar(&debounce, "d", 500, "debounce delay in milliseconds")
	flag.Parse()

	path := "."
	args := flag.Args()
	if len(args) > 0 {
		path = args[0]
	}

	if err := watcher.Watch(path, cmd, debounce); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
