package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeGo puts a go command on the PATH that records its arguments and
// PREA_PROFILE, and returns what the last call recorded.
func fakeGo(t *testing.T) func() string {
	t.Helper()
	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "go.log")
	script := "#!/bin/sh\necho \"$* | profile=$PREA_PROFILE\" > " + log + "\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PREA_PROFILE", "dev") // tests must override it
	return func() string {
		b, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(b))
	}
}

func TestBuildAndTests(t *testing.T) {
	last := fakeGo(t)
	newProject(t) // module example.com/app
	for args, want := range map[string]string{
		"build":                        "build -o bin/app . | profile=dev",
		"build -o out/server":          "build -o out/server . | profile=dev",
		"build -trimpath":              "build -o bin/app -trimpath . | profile=dev",
		"tests":                        "test ./... | profile=tests",
		"tests -run TestCreate -v":     "test -run TestCreate -v ./... | profile=tests",
		"tests -race ./preas/user/...": "test -race ./preas/user/... | profile=tests",
	} {
		if err := run(strings.Fields(args)...); err != nil {
			t.Fatalf("prea %s: %v", args, err)
		}
		if got := last(); got != want {
			t.Errorf("prea %s ran go %q, want %q", args, got, want)
		}
	}
}
