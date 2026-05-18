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
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	path := resolvePath(cmd, cfg)
	commands := resolveCommands(cmd, cfg)
	extensions := resolveExtensions(cmd, cfg)
	logger := newLogger(cmd, cfg)

	return watcher.Watch(ctx, path, commands, extensions, logger)
}

func loadConfig(cmd *cli.Command) (*config.Config, error) {
	if configPath := cmd.String("config"); configPath != "" {
		return config.Load(configPath)
	}
	return config.Find()
}

func resolvePath(cmd *cli.Command, cfg *config.Config) string {
	if path := cmd.StringArg("path"); path != "" {
		return path
	}
	if cfg != nil && cfg.Path != "" {
		return cfg.Path
	}
	return "."
}

func resolveCommands(cmd *cli.Command, cfg *config.Config) []string {
	if commands := cmd.StringSlice("command"); len(commands) > 0 {
		return commands
	}
	if cfg != nil {
		return cfg.Commands
	}
	return nil
}

func resolveExtensions(cmd *cli.Command, cfg *config.Config) []string {
	if extensions := cmd.StringSlice("extension"); len(extensions) > 0 {
		return extensions
	}
	if cfg != nil {
		return cfg.Extensions
	}
	return nil
}

func newLogger(cmd *cli.Command, cfg *config.Config) *log.Logger {
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
	return logger
}
