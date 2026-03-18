package cli

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/xrzks/fw/internal/watcher"
)

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:      "fw",
		Usage:     "file watcher",
		ArgsUsage: "[directory]",
		Action:    NewWatchCommand,
	}
}

func NewWatchCommand(ctx context.Context, c *cli.Command) error {
	var dir string
	if c.Args().Len() > 0 {
		dir = c.Args().First()
	} else {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	w := watcher.NewWatcher(dir)
	return w.Watch()
}
