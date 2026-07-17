package scheduler

import (
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startSleep launches a long-running child process and returns its pid and a
// channel reporting how the process exited.
func startSleep(t *testing.T) (int, <-chan error) {
	t.Helper()
	cmd := exec.Command("sleep", "60")
	require.NoError(t, cmd.Start())
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	return cmd.Process.Pid, done
}

// awaitSignal asserts the process exits within the timeout because of sig.
func awaitSignal(t *testing.T, wait <-chan error, sig syscall.Signal) {
	t.Helper()
	select {
	case err := <-wait:
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		require.True(t, ok, "exit status is not a WaitStatus")
		require.True(t, status.Signaled(), "process did not die by signal")
		assert.Equal(t, sig, status.Signal())
	case <-time.After(5 * time.Second):
		t.Fatalf("process did not exit after signal %v", sig)
	}
}

func TestLocalKillDefaultsToSIGTERM(t *testing.T) {
	pid, wait := startSleep(t)
	require.NoError(t, (&Local{}).Kill(strconv.Itoa(pid), nil))
	awaitSignal(t, wait, syscall.SIGTERM)
}

func TestLocalKillWithExplicitSignal(t *testing.T) {
	pid, wait := startSleep(t)
	require.NoError(t, (&Local{}).Kill(strconv.Itoa(pid), syscall.SIGKILL))
	awaitSignal(t, wait, syscall.SIGKILL)
}

func TestLocalKillInvalidPid(t *testing.T) {
	assert.Error(t, (&Local{}).Kill("not-a-pid", nil))
}
