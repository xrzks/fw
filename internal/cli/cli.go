package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
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
&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"D"},
				Usage:   "enable debug logging",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "path",
				UsageText: "The path to watch",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	path := cmd.StringArg("path")
	if path == "" {
		path = "."
	}

	logger := log.New(os.Stderr)
	if cmd.Bool("debug") {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	return watcher.Watch(ctx, path, cmd.StringSlice("command"), logger)
}
