package scheduler

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cemc-oper/orvix/internal/directive"
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

// writeJobScript writes an executable script emitting one line each to
// stdout and stderr, then exits.
func writeJobScript(t *testing.T, dir string) string {
	t.Helper()
	path := dir + "/job.sh"
	content := "#!/usr/bin/env bash\necho out-line\necho err-line 1>&2\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	return path
}

// awaitFileContent polls until the file contains want or the deadline passes.
func awaitFileContent(t *testing.T, path, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(data), want) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	data, _ := os.ReadFile(path)
	t.Fatalf("file %s never contained %q; got %q", path, want, string(data))
}

func TestLocalSubmitRedirectsOutputAndError(t *testing.T) {
	dir := t.TempDir()
	script := writeJobScript(t, dir)
	d, err := directive.Parse([]byte("#ORVIX scheduler=local\n#ORVIX output="+dir+"/job.out\n#ORVIX error="+dir+"/job.err\n"))
	require.NoError(t, err)

	_, _, err = NewLocal(d).Submit(script)
	require.NoError(t, err)

	awaitFileContent(t, dir+"/job.out", "out-line")
	awaitFileContent(t, dir+"/job.err", "err-line")
	data, _ := os.ReadFile(dir + "/job.out")
	assert.NotContains(t, string(data), "err-line")
}

func TestLocalSubmitMergesStderrIntoOutputWhenNoErrorDirective(t *testing.T) {
	dir := t.TempDir()
	script := writeJobScript(t, dir)
	d, err := directive.Parse([]byte("#ORVIX scheduler=local\n#ORVIX output="+dir+"/job.out\n"))
	require.NoError(t, err)

	_, _, err = NewLocal(d).Submit(script)
	require.NoError(t, err)

	awaitFileContent(t, dir+"/job.out", "out-line")
	awaitFileContent(t, dir+"/job.out", "err-line")
}

func TestLocalSubmitZeroValueKeepsLegacyStreams(t *testing.T) {
	dir := t.TempDir()
	script := writeJobScript(t, dir)
	pid, _, err := (&Local{}).Submit(script)
	require.NoError(t, err)
	assert.NotEmpty(t, pid)
	// no output file is created anywhere; streams stay with the parent
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1) // just job.sh
	// the detached child must finish before TempDir cleanup removes the script
	time.Sleep(300 * time.Millisecond)
}

func TestLocalSubmitMissingOutputDirFails(t *testing.T) {
	dir := t.TempDir()
	script := writeJobScript(t, dir)
	d, err := directive.Parse([]byte("#ORVIX scheduler=local\n#ORVIX output="+dir+"/no/such/dir/job.out\n"))
	require.NoError(t, err)

	_, _, err = NewLocal(d).Submit(script)
	assert.Error(t, err)
}
