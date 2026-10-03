package scheduler

import (
	"errors"
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
//
// Unlike the slurm/donau backends there is no scheduler layer to honour the
// output/error directives, so Submit redirects the child's stdout/stderr to
// those paths itself (matching slurm semantics: no error directive means
// stderr merges into the output file). Without this the child inherits
// whatever stdout the submit caller had — under ecFlow's ECF_JOB_CMD that
// output is lost entirely, leaving failed local jobs with no log at all.
// A zero-value Local (e.g. from ByName on status/kill paths) keeps the
// legacy inherit-parent behaviour.
type Local struct {
	outputPath string
	errorPath  string
}

// NewLocal builds a Local backend honouring the output/error directives.
func NewLocal(d *directive.Set) *Local {
	return &Local{outputPath: d.Get("output"), errorPath: d.Get("error")}
}

func (l *Local) Name() string { return "local" }

func (l *Local) PreambleFor(_ *directive.Set) ([]string, error) {
	return nil, nil
}

func (l *Local) Submit(scriptPath string) (string, string, error) {
	log.Debugf("[local] starting script: %s", scriptPath)
	cmd := exec.Command(scriptPath)
	setProcGroup(cmd)
	closers, err := l.redirect(cmd)
	if err != nil {
		return "", cmd.String(), err
	}
	if err := cmd.Start(); err != nil {
		for _, c := range closers {
			_ = c.Close()
		}
		return "", cmd.String(), err
	}
	// The child holds its own copy of the redirected fds; close ours.
	for _, c := range closers {
		_ = c.Close()
	}
	pid := cmd.Process.Pid
	log.Debugf("[local] started process, pid=%d", pid)
	// Reap the child in the background so we don't leave a zombie.
	go cmd.Wait()
	return strconv.Itoa(pid), cmd.String(), nil
}

// redirect wires the child's stdout/stderr to the output/error directive
// paths, falling back to the parent's streams when unset. Returned closers
// must be closed once the child has started (or failed to start).
func (l *Local) redirect(cmd *exec.Cmd) ([]*os.File, error) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if l.outputPath == "" && l.errorPath == "" {
		return nil, nil
	}
	var files []*os.File
	fail := func(err error) ([]*os.File, error) {
		for _, f := range files {
			_ = f.Close()
		}
		return nil, err
	}
	out := os.Stdout
	if l.outputPath != "" {
		f, err := os.Create(l.outputPath)
		if err != nil {
			return fail(fmt.Errorf("open output %s: %w", l.outputPath, err))
		}
		files = append(files, f)
		out = f
	}
	cmd.Stdout = out
	switch {
	case l.errorPath == "" || l.errorPath == l.outputPath:
		// slurm semantics: stderr merges into the output file
		cmd.Stderr = out
	default:
		f, err := os.Create(l.errorPath)
		if err != nil {
			return fail(fmt.Errorf("open error %s: %w", l.errorPath, err))
		}
		files = append(files, f)
		cmd.Stderr = f
	}
	return files, nil
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

// Kill sends a signal to the job's whole process tree. A nil sig defaults
// to SIGTERM (matching `kill -15`) so scripts can run their trap/cleanup
// handlers.
//
// Submit places the job in its own process group with pgid == pid, so the
// signal first goes to the entire group: a wrapper bash waiting on a
// foreground child would otherwise defer its trap until the child exits
// and never forward the signal, leaving the tree running. Jobs submitted
// before process groups were introduced have no such group; there the
// signal falls back to the single recorded pid (legacy behaviour). A job
// that has already finished is reported as success so repeated kills are
// idempotent.
func (l *Local) Kill(jobID string, sig os.Signal) error {
	pid, err := strconv.Atoi(jobID)
	if err != nil {
		return fmt.Errorf("invalid pid %q", jobID)
	}
	if sig == nil {
		sig = syscall.SIGTERM
	}
	log.Debugf("[local] sending signal %v to process group %d", sig, pid)
	err = signalGroup(pid, sig)
	if err == nil {
		return nil
	}
	if !errors.Is(err, syscall.ESRCH) {
		return err
	}
	log.Debugf("[local] no process group %d, falling back to single pid", pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(sig); err != nil {
		if errors.Is(err, syscall.ESRCH) || errors.Is(err, os.ErrProcessDone) {
			log.Debugf("[local] pid %d already finished, treating kill as success", pid)
			return nil
		}
		return err
	}
	return nil
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
