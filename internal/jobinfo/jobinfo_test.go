package jobinfo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cemc-oper/orvix/internal/directive"
)

func TestWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "info.yaml")

	info := JobInfo{
		Scheduler:       "slurm",
		JobID:           "12345",
		SubmittedAt:     time.Date(2026, 5, 10, 7, 45, 0, 0, time.UTC),
		ScriptSource:    "/abs/orig.sh",
		ScriptGenerated: "/abs/orig.20260510-074500.sh",
		SubmitDir:       "/abs",
		Hostname:        "node01",
		User:            "wangdp",
		Directives: []DirectiveKV{
			{Key: "scheduler", Value: "slurm"},
			{Key: "queue", Value: "gpu"},
			{Key: "exclusive"},
		},
	}
	require.NoError(t, Write(p, info))

	data, err := os.ReadFile(p)
	require.NoError(t, err)
	s := string(data)

	for _, want := range []string{
		"scheduler: slurm",
		`job_id: "12345"`,
		"script_source: /abs/orig.sh",
		"submit_dir: /abs",
		"- key: scheduler",
	} {
		assert.Contains(t, s, want, "yaml missing %q", want)
	}
}

func TestFromDirectivesNil(t *testing.T) {
	assert.Nil(t, FromDirectives(nil))
}

func TestFromDirectivesPreservesOrder(t *testing.T) {
	src := []byte(`#ORVIX scheduler=slurm
#ORVIX queue=gpu
#ORVIX nodes=2
`)
	set, err := directive.Parse(src)
	require.NoError(t, err)

	got := FromDirectives(set)
	require.Len(t, got, 3)
	assert.Equal(t, "scheduler", got[0].Key)
	assert.Equal(t, "queue", got[1].Key)
	assert.Equal(t, "nodes", got[2].Key)
}

func TestReadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "demo.info.yaml")

	want := JobInfo{
		Scheduler:       "slurm",
		JobID:           "98765",
		SubmittedAt:     time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC),
		ScriptSource:    "/abs/orig.sh",
		ScriptGenerated: "/abs/orig.submit.sh",
		SubmitDir:       "/abs",
		Hostname:        "node02",
		User:            "wangdp",
		Directives: []DirectiveKV{
			{Key: "scheduler", Value: "slurm"},
			{Key: "queue", Value: "normal"},
		},
	}
	require.NoError(t, Write(p, want))

	got, err := Read(p)
	require.NoError(t, err)

	assert.Equal(t, want.Scheduler, got.Scheduler)
	assert.Equal(t, want.JobID, got.JobID)
	assert.Equal(t, want.ScriptSource, got.ScriptSource)
	assert.Equal(t, want.ScriptGenerated, got.ScriptGenerated)
	assert.Equal(t, want.SubmitDir, got.SubmitDir)
	assert.Equal(t, want.Hostname, got.Hostname)
	assert.Equal(t, want.User, got.User)
	assert.True(t, got.SubmittedAt.Equal(want.SubmittedAt))
	assert.Equal(t, want.Directives, got.Directives)
}

func TestReadMissingScheduler(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "noschedver.yaml")
	// YAML lacks scheduler field; consumers should treat it as default ("local").
	require.NoError(t, os.WriteFile(p, []byte("job_id: \"1\"\n"), 0o644))

	got, err := Read(p)
	require.NoError(t, err)
	assert.Empty(t, got.Scheduler)
	assert.Equal(t, "1", got.JobID)
}

func TestReadMissing(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "nope.yaml"))
	require.Error(t, err, "expected error for missing file")
}
