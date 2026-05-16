package cli

import (
	"context"

	"github.com/urfave/cli/v3"
	"github.com/xrzks/fw/internal/watcher"
)

func New() *cli.Command {
	return &cli.Command{
		Name:        "fw",
		Usage:       "watch files and run commands on change",
		Version:     "0.1.0",
		Description: "fw watches a file or directory for changes and runs the specified commands when a change is detected.",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "command",
				Aliases: []string{"c"},
				Usage:   "command to run on file changes (can be specified multiple times)",
			},
			&cli.IntFlag{
				Name:    "debounce",
				Aliases: []string{"d"},
				Value:   500,
				Usage:   "debounce delay in milliseconds",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"D"},
				Usage:   "enable debug logging",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "path",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	var path string
	path = cmd.StringArg("path")
	if path == "" {
		path = "."
	}
	debounce := cmd.Int("debounce")
	commands := cmd.StringSlice("command")
	debug := cmd.Bool("debug")
	return watcher.Watch(ctx, path, commands, debounce, debug)
}
