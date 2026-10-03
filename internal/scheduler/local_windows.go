//go:build windows

package scheduler

import (
	"os"
	"os/exec"
	"syscall"
)

// setProcGroup is a no-op on Windows: there is no POSIX process-group
// concept, so Kill falls back to signalling the single process.
func setProcGroup(cmd *exec.Cmd) {}

// signalGroup always reports ESRCH on Windows so Kill falls back to the
// single-pid path, preserving the previous behaviour.
func signalGroup(_ int, _ os.Signal) error {
	return syscall.ESRCH
}
