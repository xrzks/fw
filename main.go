package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xrzks/fw/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
