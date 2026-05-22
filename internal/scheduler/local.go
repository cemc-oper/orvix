package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/cemc-oper/orvix/internal/directive"
)

// Local runs scripts directly on the current host without a scheduler.
type Local struct{}

func (l *Local) Name() string { return "local" }

func (l *Local) PreambleFor(_ *directive.Set) ([]string, error) {
	return nil, nil
}

func (l *Local) Submit(scriptPath string) (string, error) {
	cmd := exec.Command(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	pid := cmd.Process.Pid
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
		return "FINISHED", nil
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return "FINISHED", nil
	}
	return "RUNNING", nil
}

func (l *Local) Kill(jobID string) error {
	pid, err := strconv.Atoi(jobID)
	if err != nil {
		return fmt.Errorf("invalid pid %q", jobID)
	}
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
