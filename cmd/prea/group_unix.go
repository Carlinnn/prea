//go:build unix

package main

import (
	"os/exec"
	"syscall"
)

// ownGroup runs cmd in its own process group, so a Ctrl+C in the terminal
// reaches only the CLI, which forwards it: the server gets one signal
// whether it came from the terminal, an IDE or kill.
func ownGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
