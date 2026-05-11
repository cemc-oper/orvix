package jobinfo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			{Key: "partition", Value: "gpu"},
			{Key: "exclusive"},
		},
	}
	if err := Write(p, info); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		"scheduler: slurm",
		`job_id: "12345"`,
		"script_source: /abs/orig.sh",
		"submit_dir: /abs",
		"- key: scheduler",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("yaml missing %q\n--- yaml ---\n%s", want, s)
		}
	}
}

func TestFromDirectivesNil(t *testing.T) {
	if got := FromDirectives(nil); got != nil {
		t.Errorf("FromDirectives(nil) = %v, want nil", got)
	}
}

func TestFromDirectivesPreservesOrder(t *testing.T) {
	src := []byte(`#ORVIX scheduler=slurm
#ORVIX partition=gpu
#ORVIX nodes=2
`)
	set, err := directive.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	got := FromDirectives(set)
	want := []string{"scheduler", "partition", "nodes"}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Key != w {
			t.Errorf("got[%d].Key = %q, want %q", i, got[i].Key, w)
		}
	}
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
			{Key: "partition", Value: "normal"},
		},
	}
	if err := Write(p, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Scheduler != want.Scheduler ||
		got.JobID != want.JobID ||
		got.ScriptSource != want.ScriptSource ||
		got.ScriptGenerated != want.ScriptGenerated ||
		got.SubmitDir != want.SubmitDir ||
		got.Hostname != want.Hostname ||
		got.User != want.User {
		t.Errorf("Read scalar fields differ:\n got=%+v\nwant=%+v", got, want)
	}
	if !got.SubmittedAt.Equal(want.SubmittedAt) {
		t.Errorf("SubmittedAt got=%v want=%v", got.SubmittedAt, want.SubmittedAt)
	}
	if len(got.Directives) != len(want.Directives) {
		t.Fatalf("directives len got=%d want=%d", len(got.Directives), len(want.Directives))
	}
	for i := range got.Directives {
		if got.Directives[i] != want.Directives[i] {
			t.Errorf("Directives[%d] got=%+v want=%+v", i, got.Directives[i], want.Directives[i])
		}
	}
}

func TestReadMissingScheduler(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "noschedver.yaml")
	// YAML lacks scheduler field; consumers should treat it as default ("local").
	if err := os.WriteFile(p, []byte("job_id: \"1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Scheduler != "" {
		t.Errorf("Scheduler got=%q want empty", got.Scheduler)
	}
	if got.JobID != "1" {
		t.Errorf("JobID got=%q want %q", got.JobID, "1")
	}
}

func TestReadMissing(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatal("Read: expected error for missing file, got nil")
	}
}
