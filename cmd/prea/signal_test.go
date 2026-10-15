//go:build unix

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// A signal sent only to the CLI, as an IDE's stop button, kill or a closing
// terminal (SIGHUP) do, reaches the server: prea run and prea dev forward
// it, the server shuts down gracefully and the CLI exits.
func TestSignalToCLIStopsServer(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a project several times")
	}
	framework, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(t.TempDir(), "prea")
	if out, err := exec.Command("go", "build", "-o", cli, ".").CombinedOutput(); err != nil {
		t.Fatalf("build the CLI: %v\n%s", err, out)
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
	up := func() bool {
		res, err := http.Get(url)
		if err == nil {
			res.Body.Close()
		}
		return err == nil
	}

	for _, tc := range []struct {
		command string
		signal  syscall.Signal
	}{{"run", syscall.SIGINT}, {"dev", syscall.SIGINT}, {"run", syscall.SIGHUP}, {"dev", syscall.SIGHUP}} {
		command := tc.command
		t.Run(command+" "+tc.signal.String(), func(t *testing.T) {
			var out syncBuffer
			c := exec.Command(cli, command)
			c.Stdout, c.Stderr = &out, &out
			if err := c.Start(); err != nil {
				t.Fatal(err)
			}
			exited := make(chan error, 1)
			go func() { exited <- c.Wait() }()
			defer c.Process.Kill()

			for i := 0; !up(); i++ {
				if i == 600 {
					t.Fatalf("prea %s never served:\n%s", command, out.String())
				}
				time.Sleep(100 * time.Millisecond)
			}

			c.Process.Signal(tc.signal) // the CLI only, not its process group
			select {
			case <-exited:
			case <-time.After(30 * time.Second):
				t.Fatalf("prea %s must exit after %v:\n%s", command, tc.signal, out.String())
			}
			if up() {
				t.Fatalf("prea %s left the server running", command)
			}
			if !strings.Contains(out.String(), "prea: stopped") {
				t.Fatalf("the server must shut down gracefully:\n%s", out.String())
			}
		})
	}
}

// syncBuffer is a bytes.Buffer the CLI and the server can write at once.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}
