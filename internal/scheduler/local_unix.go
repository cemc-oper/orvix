//go:build !windows

package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// setProcGroup puts the child in its own process group, making it a group
// leader with pgid == pid. Kill can then signal the whole job tree via
// kill(-pid), and signals of the submit caller's own group no longer reach
// the job.
func setProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalGroup sends sig to every process in process group pgid
// (i.e. kill(-pgid, sig)).
func signalGroup(pgid int, sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return fmt.Errorf("unsupported signal type %T", sig)
	}
	return syscall.Kill(-pgid, s)
}
