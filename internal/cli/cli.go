package cli

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"

	"github.com/xrzks/fw/internal/config"
	"github.com/xrzks/fw/internal/ignore"
	"github.com/xrzks/fw/internal/watcher"
)

func New() *cli.Command {
	return &cli.Command{
		Name:        "fw",
		Usage:       "watch files and run commands on change",
		Version:     "0.1.0",
		Description: "fw watches a file or directory for changes and runs the specified commands when a change is detected.",
		ArgsUsage:   "[path]",
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
			&cli.StringSliceFlag{
				Name:    "ignore",
				Aliases: []string{"i"},
				Usage:   "glob pattern to ignore (can be specified multiple times)",
			},
			&cli.BoolFlag{
				Name:  "no-gitignore",
				Usage: "disable automatic .gitignore loading",
			},
			&cli.BoolFlag{
				Name:    "fail-fast",
				Aliases: []string{"f"},
				Usage:   "stop running subsequent commands on first failure",
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

	logger := newLogger(cmd, cfg)

	if cfg != nil {
		logger.Debug("Config loaded", "path", cfg.Path, "commands", cfg.Commands, "extensions", cfg.Extensions, "ignore", cfg.Ignore, "no-gitignore", cfg.NoGitignore, "fail-fast", cfg.FailFast, "debug", cfg.Debug)
	} else {
		logger.Debug("No config file found, using defaults")
	}

	path := resolvePath(cmd, cfg)
	commands := resolveCommands(cmd, cfg)
	extensions := resolveExtensions(cmd, cfg)
	ignorer, ignorePatterns := resolveIgnore(cmd, cfg, logger)
	failFast := resolveFailFast(cmd, cfg)

	logger.Debug("Resolved config", "path", path, "commands", commands, "extensions", extensions, "ignore-patterns", ignorePatterns, "fail-fast", failFast)

	return watcher.Watch(ctx, watcher.WatchOptions{
		Path:       path,
		Commands:   commands,
		Extensions: extensions,
		Ignorer:    ignorer,
		Logger:     logger,
		FailFast:   failFast,
	})
}

func loadConfig(cmd *cli.Command) (*config.Config, error) {
	if configPath := cmd.String("config"); configPath != "" {
		return config.Load(configPath)
	}
	return config.FindAndLoad()
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
	var commands []string
	if cfg != nil {
		commands = append(commands, cfg.Commands...)
	}
	commands = append(commands, cmd.StringSlice("command")...)
	return commands
}

func resolveExtensions(cmd *cli.Command, cfg *config.Config) []string {
	var extensions []string
	if cfg != nil {
		extensions = append(extensions, cfg.Extensions...)
	}
	extensions = append(extensions, cmd.StringSlice("extension")...)
	return extensions
}

func resolveIgnore(cmd *cli.Command, cfg *config.Config, logger *log.Logger) (*ignore.Matcher, []string) {
	var patterns []string
	if cfg != nil {
		patterns = append(patterns, cfg.Ignore...)
	}
	noGitignore := cmd.Bool("no-gitignore")
	if !noGitignore && cfg != nil {
		noGitignore = cfg.NoGitignore
	}
	if !noGitignore {
		gitignoreDir := "."
		if cfg != nil && cfg.Path != "" {
			gitignoreDir = cfg.Path
		}
		gitignorePatterns := config.LoadGitignore(gitignoreDir)
		logger.Debugf("Loaded %d patterns from .gitignore in %s", len(gitignorePatterns), gitignoreDir)
		patterns = append(patterns, gitignorePatterns...)
	} else {
		logger.Debug(".gitignore loading disabled")
	}
	patterns = append(patterns, cmd.StringSlice("ignore")...)
	return ignore.New(patterns), patterns
}

func resolveFailFast(cmd *cli.Command, cfg *config.Config) bool {
	if cmd.IsSet("fail-fast") {
		return cmd.Bool("fail-fast")
	}
	if cfg != nil {
		return cfg.FailFast
	}
	return false
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
