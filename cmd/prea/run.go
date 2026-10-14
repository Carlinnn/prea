package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v3"
)

var profileFlag = &cli.StringFlag{
	Name:  "profile",
	Usage: "config profile: loads config/application-<profile>.yaml",
}

func runCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Build and run the project (the main package at the root)",
		Flags: []cli.Flag{profileFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(ctx, cmd)
		},
	}
}

// runServer builds the main package at the project root and runs it with
// the profile and env, until it exits. It runs the binary directly (not via
// go run) so it gets the signals and can shut down gracefully.
func runServer(ctx context.Context, cmd *cli.Command, env ...string) error {
	dir, err := os.MkdirTemp("", "prea-run")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "server")
	if err := buildServer(ctx, bin); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	server := exec.Command(bin)
	server.Env = append(os.Environ(), env...)
	if cmd.IsSet("profile") {
		server.Env = append(server.Env, "PREA_PROFILE="+cmd.String("profile")) // config.ProfileEnv
	}
	server.Stdin, server.Stdout, server.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := server.Start(); err != nil {
		return err
	}
	go func() {
		for sig := range signals {
			// Ctrl+C already reached the server through the terminal's process group.
			if sig != os.Interrupt {
				server.Process.Signal(sig)
			}
		}
	}()
	return server.Wait()
}

// buildServer builds the main package at the project root into bin.
func buildServer(ctx context.Context, bin string) error {
	if err := goRun(ctx, nil, "build", "-o", bin, "."); err != nil {
		return fmt.Errorf("build: %w", err)
	}
	return nil
}
