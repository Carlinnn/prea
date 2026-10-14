package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// prea dev on a new project: it serves, a change to a handler restarts it
// with the change, a build error keeps the running version, and the end of
// the context stops the server.
func TestDev(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a project several times")
	}
	framework, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range map[string]string{"GOFLAGS": "-mod=mod", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off"} {
		t.Setenv(k, v)
	}
	t.Chdir(t.TempDir())
	if err := run("new", "shop", "--local", framework); err != nil {
		t.Fatal(err)
	}
	t.Chdir("shop")
	url := fmt.Sprintf("http://127.0.0.1:%d/health", usePort(t))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- dev(ctx, "dev", 100*time.Millisecond) }()
	defer func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(30 * time.Second):
			t.Error("dev must stop when the context ends")
		}
	}()

	body := func() string {
		res, err := http.Get(url)
		if err != nil {
			return ""
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return string(b)
	}
	eventually := func(want, why string) {
		t.Helper()
		for range 600 { // builds take a few seconds
			if strings.Contains(body(), want) {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatalf("%s: never answered %q (last: %q)", why, want, body())
	}

	eventually(`"ok"`, "dev must build and serve the project")

	handler := filepath.Join("preas", "health", "handler.go")
	src, _ := os.ReadFile(handler)
	write(t, handler, strings.Replace(string(src), `"ok"`, `"changed"`, 1))
	eventually(`"changed"`, "a change to a handler must restart the server with it")

	write(t, filepath.Join("preas", "health", "broken.go"), "package health\n\nfunc {\n")
	time.Sleep(3 * time.Second) // time to notice, rebuild and fail
	if got := body(); !strings.Contains(got, `"changed"`) {
		t.Fatalf("a build error must keep the running version: %q", got)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	done <- nil // for the deferred check
	if got := body(); got != "" {
		t.Fatalf("the server must stop with dev: %q", got)
	}
}
