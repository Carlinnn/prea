package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

var projectName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// frameworkModules are the modules a project needs: the root (prea.Run)
// and the ones it requires. --local points each one at a clone.
var frameworkModules = []string{"", "config", "core", "database", "di", "logger", "security"}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:      "new",
		Usage:     "Create a project: main.go, preas/, config/ with the profiles, tests/ and a health preá",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "module", Usage: "Go module path of the project, the start of its imports (default: the name), e.g. github.com/you/shop"},
			&cli.BoolFlag{Name: "with-db", Usage: "configure a database in the profiles and generate a CRUD preá (item)"},
			&cli.StringFlag{Name: "local", Usage: "use the framework from a local clone instead of downloading it (to develop the framework)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.NArg() != 1 {
				return errors.New("expected exactly one argument: <name>")
			}
			name := cmd.Args().First()
			if !projectName.MatchString(name) {
				return fmt.Errorf("invalid name %q: start with a letter, then letters, digits, - or _", name)
			}
			return createProject(ctx, name, cmp.Or(cmd.String("module"), name), cmd.Bool("with-db"), cmd.String("local"))
		},
	}
}

// createProject renders templates/new into name and resolves the framework.
// On failure it leaves nothing behind.
func createProject(ctx context.Context, name, module string, withDB bool, local string) (err error) {
	if entries, err := os.ReadDir(name); err == nil && len(entries) > 0 {
		return fmt.Errorf("%s already exists and is not empty", name)
	}
	if err := os.MkdirAll(name, 0o755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(name)
		}
	}()

	data := templateData{Name: name, Module: module, Framework: frameworkModule, WithDB: withDB}
	err = fs.WalkDir(templatesFS, "templates/new", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(name, "templates/new/"), ".tmpl")
		return render(filepath.Join(data.Name, filepath.FromSlash(rel)), name, data)
	})
	if err != nil {
		return err
	}

	steps := [][]string{{"get", frameworkModule + "@latest"}, {"mod", "tidy"}}
	if local != "" {
		if err := useLocal(filepath.Join(name, "go.mod"), local); err != nil {
			return err
		}
		steps = steps[1:]
	}
	if withDB {
		if err := inDir(name, func() error { return generatePrea("item") }); err != nil {
			return err
		}
	}
	for _, args := range steps {
		if err := inDir(name, func() error { return goRun(ctx, nil, args...) }); err != nil {
			return fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
		}
	}

	fmt.Printf("project %s created: cd %s && cp .env.example .env && prea dev\n", name, name)
	return nil
}

// useLocal requires the framework modules from a clone at dir.
func useLocal(gomod, dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, "prea.go")); err != nil {
		return fmt.Errorf("--local %s: not a clone of the framework (no prea.go)", dir)
	}
	lines := "\nrequire " + frameworkModule + " v0.0.0\n"
	for _, m := range frameworkModules {
		lines += fmt.Sprintf("\nreplace %s => %s\n", strings.TrimSuffix(frameworkModule+"/"+m, "/"), filepath.Join(dir, m))
	}
	f, err := os.OpenFile(gomod, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = f.WriteString(lines)
	return errors.Join(err, f.Close())
}

// inDir runs fn with dir as the current directory, as the generator expects.
func inDir(dir string, fn func() error) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer os.Chdir(wd)
	return fn()
}
