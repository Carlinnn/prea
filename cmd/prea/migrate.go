package main

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
)

// migrateCommand runs the project in inspection mode: the server built with
// PREA_MIGRATE="<command> [args]" runs the command with its own config and
// database (core.App.Run) instead of serving. The arguments are checked
// there, in one place.
func migrateCommand() *cli.Command {
	sub := func(name, args, usage string) *cli.Command {
		return &cli.Command{
			Name:      name,
			Usage:     usage,
			ArgsUsage: args,
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return runServer(ctx, cmd, "PREA_MIGRATE="+strings.Join(append([]string{name}, cmd.Args().Slice()...), " "))
			},
		}
	}
	return &cli.Command{
		Name:  "migrate",
		Usage: "Manage the database migrations (database.migrations in the config)",
		Flags: []cli.Flag{profileFlag},
		Commands: []*cli.Command{
			sub("create", "<name>", "create an empty migration with the next version"),
			sub("up", "[version]", "apply the pending migrations, or up to version"),
			sub("down", "[n]", "undo the last n applied migrations (default 1)"),
			sub("status", "", "list the migrations: applied or pending"),
			sub("validate", "", "check the migration files without touching the database"),
		},
	}
}
