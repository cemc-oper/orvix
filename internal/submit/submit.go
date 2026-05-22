// Package submit encapsulates the full job-submission pipeline.
package submit

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
	"github.com/cemc-oper/orvix/internal/script"
	"github.com/cemc-oper/orvix/internal/version"
	"github.com/cemc-oper/orvix/internal/watch"
)

// Options controls a single submission.
type Options struct {
	ScriptPath    string
	Scheduler     string // override
	DryRun        bool
	OutScript     string
	OutInfo       string
	Watch         bool
	WatchInterval time.Duration
	Out           io.Writer
}

// Run executes the full submit pipeline.
// On success the job ID is written to Out.
func Run(opts Options) error {
	origPath, err := filepath.Abs(opts.ScriptPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	log.Debugf("[submit] script path: %s", origPath)

	logPath := deriveLogPath(origPath)
	wrap := func(err error) error {
		if err != nil {
			_ = writeSubmitLog(logPath, err)
		}
		return err
	}

	src, err := os.ReadFile(origPath)
	if err != nil {
		return wrap(fmt.Errorf("read script: %w", err))
	}
	log.Debugf("[submit] read script: %d bytes", len(src))

	directives, err := directive.ParseWithOverride(src, opts.Scheduler)
	if err != nil {
		return wrap(fmt.Errorf("parse directives: %w", err))
	}
	log.Debugf("[submit] parsed %d directive(s), scheduler=%q", len(directives.Items), directives.Scheduler())

	sched, err := scheduler.For(directives)
	if err != nil {
		return wrap(fmt.Errorf("resolve scheduler: %w", err))
	}
	log.Debugf("[submit] using scheduler: %s", sched.Name())

	generated, err := script.Generate(src, directives, sched)
	if err != nil {
		return wrap(fmt.Errorf("generate script: %w", err))
	}
	log.Debugf("[submit] generated script: %d bytes", len(generated))

	if opts.DryRun {
		log.Debug("[submit] dry-run mode: printing generated script")
		_, err = opts.Out.Write(generated)
		return wrap(err)
	}

	now := time.Now()
	genScriptPath, yamlPath := derivePaths(origPath, opts.OutScript, opts.OutInfo)
	log.Debugf("[submit] output script: %s", genScriptPath)
	log.Debugf("[submit] output info:   %s", yamlPath)

	if err := os.WriteFile(genScriptPath, generated, 0o755); err != nil {
		return wrap(fmt.Errorf("write generated script: %w", err))
	}
	log.Debug("[submit] wrote generated script")

	jobID, submitCmd, err := sched.Submit(genScriptPath)
	if err != nil {
		return wrap(fmt.Errorf("submit: %w", err))
	}
	log.Debugf("[submit] job submitted, id=%s", jobID)

	cwd, _ := os.Getwd()
	info := jobinfo.JobInfo{
		Version:         version.Version,
		Scheduler:       sched.Name(),
		JobID:           jobID,
		SubmittedAt:     now,
		ScriptSource:    origPath,
		ScriptGenerated: genScriptPath,
		SubmitDir:       cwd,
		Hostname:        hostnameOrEmpty(),
		User:            usernameOrEmpty(),
		SubmitCommand:   submitCmd,
		Directives:      jobinfo.FromDirectives(directives),
	}
	if err := jobinfo.Write(yamlPath, info); err != nil {
		return wrap(fmt.Errorf("write job info: %w", err))
	}
	log.Debugf("[submit] wrote job info: %s", yamlPath)

	fmt.Fprintln(opts.Out, jobID)

	if opts.Watch {
		log.Debugf("[submit] entering watch mode, interval=%s", opts.WatchInterval)
		return watch.Run(sched, jobID, opts.WatchInterval, opts.Out)
	}
	return nil
}

// derivePaths returns the (script, yaml) sidecar paths.
//
// By default they are placed next to origPath:
//   - <orig>.submit      (translated runnable script)
//   - <orig>.info.yaml   (job metadata)
//
// If outScript or outInfo are non-empty they override the default paths.
// Resubmits overwrite.
func derivePaths(origPath, outScript, outInfo string) (string, string) {
	var scriptPath, yamlPath string
	if outScript != "" {
		scriptPath = outScript
	} else {
		dir := filepath.Dir(origPath)
		base := filepath.Base(origPath)
		scriptPath = filepath.Join(dir, base+".submit")
	}
	if outInfo != "" {
		yamlPath = outInfo
	} else {
		dir := filepath.Dir(origPath)
		base := filepath.Base(origPath)
		yamlPath = filepath.Join(dir, base+".info.yaml")
	}
	return scriptPath, yamlPath
}

// deriveLogPath returns the default log path for a failed submit:
//   - <orig>.submit.log
func deriveLogPath(origPath string) string {
	dir := filepath.Dir(origPath)
	base := filepath.Base(origPath)
	return filepath.Join(dir, base+".submit.log")
}

// writeSubmitLog writes err to path with a timestamp.
// Failures are silently ignored so logging never masks the original error.
func writeSubmitLog(path string, err error) error {
	if err == nil {
		return nil
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = fmt.Fprintf(f, "[%s] ERROR: %v\n", time.Now().Format(time.RFC3339), err)
	return e
}

func hostnameOrEmpty() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

func usernameOrEmpty() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}
