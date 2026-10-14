package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:  "prea",
		Usage: "Preá framework CLI",
		Commands: []*cli.Command{
			runCommand(),
			buildCommand(),
			testsCommand(),
			generateCommand(),
			migrateCommand(),
		},
	}
}
