package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

// fakeScheduler implements scheduler.Scheduler with a fixed normalization.
type fakeScheduler struct{ state scheduler.JobState }

func (f *fakeScheduler) Name() string                                     { return "fake" }
func (f *fakeScheduler) PreambleFor(d *directive.Set) ([]string, error)   { return nil, nil }
func (f *fakeScheduler) Submit(scriptPath string) (string, string, error) { return "", "", nil }
func (f *fakeScheduler) Status(jobID string) (string, error)              { return "", nil }
func (f *fakeScheduler) Kill(jobID string, sig os.Signal) error           { return nil }
func (f *fakeScheduler) NormalizeState(raw string) scheduler.JobState     { return f.state }

func TestNewStatusJSON(t *testing.T) {
	info := &jobinfo.JobInfo{Scheduler: "fake", JobID: "42"}

	out := newStatusJSON(info, &fakeScheduler{state: scheduler.StateCompleted}, "DONE")
	assert.Equal(t, "42", out.JobID)
	assert.Equal(t, "fake", out.Scheduler)
	assert.Equal(t, "COMPLETED", out.State)
	assert.Equal(t, "DONE", out.RawState)
	assert.True(t, out.Terminal)

	out = newStatusJSON(info, &fakeScheduler{state: scheduler.StateRunning}, "RUNNING")
	assert.Equal(t, "RUNNING", out.State)
	assert.False(t, out.Terminal)
}

func TestStatusJSONEncoding(t *testing.T) {
	info := &jobinfo.JobInfo{Scheduler: "slurm", JobID: "123"}
	var buf bytes.Buffer
	require.NoError(t, json.NewEncoder(&buf).Encode(
		newStatusJSON(info, &fakeScheduler{state: scheduler.StateCancelled}, "CANCELLED by 30255")))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, "123", decoded["job_id"])
	assert.Equal(t, "slurm", decoded["scheduler"])
	assert.Equal(t, "CANCELLED", decoded["state"])
	assert.Equal(t, "CANCELLED by 30255", decoded["raw_state"])
	assert.Equal(t, true, decoded["terminal"])
}

// TestStatusCmdJSONLocal runs the real command against the local backend
// using this test process as the "job", so no scheduler tools are needed.
func TestStatusCmdJSONLocal(t *testing.T) {
	dir := t.TempDir()
	infoPath := filepath.Join(dir, "job.info.yaml")
	require.NoError(t, jobinfo.Write(infoPath, jobinfo.JobInfo{
		Scheduler: "local",
		JobID:     strconv.Itoa(os.Getpid()),
	}))

	var buf bytes.Buffer
	cmd := statusCmd
	cmd.SetOut(&buf)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	defer func() {
		require.NoError(t, cmd.Flags().Set("format", "text"))
	}()

	require.NoError(t, cmd.RunE(cmd, []string{infoPath}))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, strconv.Itoa(os.Getpid()), decoded["job_id"])
	assert.Equal(t, "local", decoded["scheduler"])
	assert.Equal(t, "RUNNING", decoded["state"])
	assert.Equal(t, "RUNNING", decoded["raw_state"])
	assert.Equal(t, false, decoded["terminal"])
}

func TestStatusCmdRejectsUnknownFormat(t *testing.T) {
	cmd := statusCmd
	require.NoError(t, cmd.Flags().Set("format", "yaml"))
	defer func() {
		require.NoError(t, cmd.Flags().Set("format", "text"))
	}()
	err := cmd.RunE(cmd, []string{"whatever.info.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown --format")
}
