// Package submit encapsulates the full job-submission pipeline.
package submit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
	"github.com/cemc-oper/orvix/internal/script"
)

// GenerateOptions controls script generation without submission.
type GenerateOptions struct {
	ScriptPath string
	Scheduler  string // override
	OutScript  string
	Out        io.Writer
}

// Generate reads a script, parses #ORVIX directives, generates a
// scheduler-specific runnable script, writes it to disk, and returns
// the path to the generated script.
func Generate(opts GenerateOptions) (string, error) {
	origPath, err := filepath.Abs(opts.ScriptPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	log.Debugf("[generate] script path: %s", origPath)

	logPath := deriveLogPath(origPath)
	wrap := func(err error) error {
		if err != nil {
			_ = writeSubmitLog(logPath, false, "", err.Error())
		}
		return err
	}

	src, err := os.ReadFile(origPath)
	if err != nil {
		return "", wrap(fmt.Errorf("read script: %w", err))
	}
	log.Debugf("[generate] read script: %d bytes", len(src))

	directives, err := directive.ParseWithOverride(src, opts.Scheduler)
	if err != nil {
		return "", wrap(fmt.Errorf("parse directives: %w", err))
	}
	log.Debugf("[generate] parsed %d directive(s), scheduler=%q", len(directives.Items), directives.Scheduler())

	sched, err := scheduler.For(directives)
	if err != nil {
		return "", wrap(fmt.Errorf("resolve scheduler: %w", err))
	}
	log.Debugf("[generate] using scheduler: %s", sched.Name())

	generated, err := script.Render(src, directives, sched)
	if err != nil {
		return "", wrap(fmt.Errorf("generate script: %w", err))
	}
	log.Debugf("[generate] generated script: %d bytes", len(generated))

	genScriptPath := deriveScriptPath(origPath, opts.OutScript)
	log.Debugf("[generate] output script: %s", genScriptPath)

	if err := os.WriteFile(genScriptPath, generated, 0o755); err != nil {
		return "", wrap(fmt.Errorf("write generated script: %w", err))
	}
	log.Debug("[generate] wrote generated script")

	return genScriptPath, nil
}

// deriveScriptPath returns the output path for the generated script.
// If outScript is non-empty it is used directly; otherwise the default
// <orig>.submit next to the original is returned.
func deriveScriptPath(origPath, outScript string) string {
	if outScript != "" {
		return outScript
	}
	dir := filepath.Dir(origPath)
	base := filepath.Base(origPath)
	return filepath.Join(dir, base+".submit")
}
