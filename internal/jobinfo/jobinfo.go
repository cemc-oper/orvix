// Package jobinfo records submitted-job metadata as a YAML sidecar next to the
// generated script.
package jobinfo

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/cemc-oper/orvix/internal/directive"
)

// JobInfo is the on-disk record of a submitted orvix job.
type JobInfo struct {
	Scheduler       string        `yaml:"scheduler"`
	JobID           string        `yaml:"job_id"`
	SubmittedAt     time.Time     `yaml:"submitted_at"`
	ScriptSource    string        `yaml:"script_source"`
	ScriptGenerated string        `yaml:"script_generated"`
	SubmitDir       string        `yaml:"submit_dir,omitempty"`
	Hostname        string        `yaml:"hostname,omitempty"`
	User            string        `yaml:"user,omitempty"`
	Directives      []DirectiveKV `yaml:"directives,omitempty"`
}

// DirectiveKV is one parsed `#ORVIX key=value` line, recorded for traceability.
type DirectiveKV struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value,omitempty"`
}

// FromDirectives copies a directive.Set into the YAML-friendly slice form.
func FromDirectives(d *directive.Set) []DirectiveKV {
	if d == nil {
		return nil
	}
	out := make([]DirectiveKV, 0, len(d.Items))
	for _, item := range d.Items {
		out = append(out, DirectiveKV{Key: item.Key, Value: item.Value})
	}
	return out
}

// Write serializes info as YAML to path (mode 0644).
func Write(path string, info JobInfo) error {
	data, err := yaml.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal job info: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Read parses the YAML at path into a JobInfo.
func Read(path string) (JobInfo, error) {
	var info JobInfo
	data, err := os.ReadFile(path)
	if err != nil {
		return info, fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &info); err != nil {
		return info, fmt.Errorf("parse %s: %w", path, err)
	}
	return info, nil
}
