package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/log"
)

// Local runs scripts directly on the current host without a scheduler.
type Local struct{}

func (l *Local) Name() string { return "local" }

func (l *Local) PreambleFor(_ *directive.Set) ([]string, error) {
	return nil, nil
}

func (l *Local) Submit(scriptPath string) (string, error) {
	log.Debugf("[local] starting script: %s", scriptPath)
	cmd := exec.Command(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	pid := cmd.Process.Pid
	log.Debugf("[local] started process, pid=%d", pid)
	// Reap the child in the background so we don't leave a zombie.
	go cmd.Wait()
	return strconv.Itoa(pid), nil
}

func (l *Local) Status(jobID string) (string, error) {
	pid, err := strconv.Atoi(jobID)
	if err != nil {
		return "", fmt.Errorf("invalid pid %q", jobID)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Debugf("[local] process %d not found, treating as finished", pid)
		return "FINISHED", nil
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		log.Debugf("[local] signal(0) to pid %d failed: %v, treating as finished", pid, err)
		return "FINISHED", nil
	}
	log.Debugf("[local] pid %d is running", pid)
	return "RUNNING", nil
}

func (l *Local) Kill(jobID string) error {
	pid, err := strconv.Atoi(jobID)
	if err != nil {
		return fmt.Errorf("invalid pid %q", jobID)
	}
	log.Debugf("[local] killing pid %d", pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func (l *Local) NormalizeState(raw string) JobState {
	switch strings.ToUpper(raw) {
	case "RUNNING":
		return StateRunning
	case "FINISHED":
		return StateCompleted
	default:
		return StateUnknown
	}
}
