package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
	"github.com/xrzks/fw/internal/config"
	"github.com/xrzks/fw/internal/watcher"
)

func New() *cli.Command {
	return &cli.Command{
		Name:        "fw",
		Usage:       "watch files and run commands on change",
		Version:     "0.1.0",
		Description: "fw watches a file or directory for changes and runs the specified commands when a change is detected.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"C"},
				Usage:   "path to config file (fw.toml or .fw.toml are auto-detected)",
			},
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
			&cli.StringSliceFlag{
				Name:    "extension",
				Aliases: []string{"e"},
				Usage:   "file extension to watch (can be specified multiple times)",
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
	var cfg *config.Config

	if configPath := cmd.String("config"); configPath != "" {
		var err error
		cfg, err = config.Load(configPath)
		if err != nil {
			return err
		}
	} else {
		var err error
		cfg, err = config.Find()
		if err != nil {
			return err
		}
	}

	path := cmd.StringArg("path")
	if path == "" {
		if cfg != nil && cfg.Path != "" {
			path = cfg.Path
		} else {
			path = "."
		}
	}

	commands := cmd.StringSlice("command")
	if len(commands) == 0 && cfg != nil {
		commands = cfg.Commands
	}

	extensions := cmd.StringSlice("extension")
	if len(extensions) == 0 && cfg != nil {
		extensions = cfg.Extensions
	}

	debug := cmd.Bool("debug")
	if !debug && cfg != nil {
		debug = cfg.Debug
	}

	logger := log.New(os.Stderr)
	if debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	return watcher.Watch(ctx, path, commands, extensions, logger)
}
