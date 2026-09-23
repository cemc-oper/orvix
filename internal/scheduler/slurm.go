package scheduler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/log"
)

// SLURM submits jobs via sbatch / squeue / scancel.
type SLURM struct{}

func (s *SLURM) Name() string { return "slurm" }

// sbatch returns a Generator that emits "#SBATCH --flag=value".
// The value is passed through unchanged (no quoting).
func sbatch(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("#SBATCH --%s=%s", flag, v), true
	}
}

// sbatchQ returns a Generator that emits "#SBATCH --flag=value"
// with the value wrapped in double quotes when it contains whitespace.
func sbatchQ(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("#SBATCH --%s=%s", flag, slurmQuote(v)), true
	}
}

// sbatchBare returns a Generator for bare SLURM flags.
// When the orvix value is empty it emits "#SBATCH --flag";
// otherwise "#SBATCH --flag=value".
func sbatchBare(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok {
			return "", false
		}
		if v == "" {
			return fmt.Sprintf("#SBATCH --%s", flag), true
		}
		return fmt.Sprintf("#SBATCH --%s=%s", flag, v), true
	}
}

// PreambleFor translates orvix generic directives to #SBATCH lines.
//
// Each generator queries the directive set for the orvix keys it cares about
// and emits one scheduler-specific line. Generators that find nothing return
// ("", false) and are skipped.
func (s *SLURM) PreambleFor(d *directive.Set) ([]string, error) {
	generators := []Generator{
		sbatchQ("job-name", "job-name"),
		sbatchQ("output", "output"),
		sbatchQ("error", "error"),
		sbatch("nodes", "nodes"),
		sbatch("ntasks", "ntasks"),
		sbatch("ntasks-per-node", "ntasks-per-node"),
		sbatch("cpus-per-task", "cpus-per-task"),
		sbatch("time", "time"),
		sbatchQ("queue", "partition"),
		sbatchQ("account", "account"),
		sbatchQ("project", "wckey"),
		sbatchQ("application", "comment"),
		sbatchBare("exclusive", "exclusive"),
		sbatchQ("nodelist", "nodelist"),
		sbatch("memory", "mem"),
		sbatch("dependency", "dependency"),
		slurmNoRequeue,
	}

	var lines []string
	for _, gen := range generators {
		if line, ok := gen(d); ok {
			lines = append(lines, line)
		}
	}
	log.Debugf("[slurm] preamble: %d line(s)", len(lines))
	return lines, nil
}

func slurmQuote(v string) string {
	if strings.ContainsAny(v, " \t") {
		return `"` + v + `"`
	}
	return v
}

// slurmNoRequeue emits "#SBATCH --no-requeue" when the requeue directive is
// explicitly false. requeue is a neutral boolean key (default true); when true
// or absent, SLURM's default requeue behavior is kept and nothing is emitted.
func slurmNoRequeue(d *directive.Set) (string, bool) {
	v, ok := d.GetOK("requeue")
	if !ok {
		return "", false
	}
	if strings.EqualFold(strings.TrimSpace(v), "false") {
		return "#SBATCH --no-requeue", true
	}
	return "", false
}

func (s *SLURM) Submit(scriptPath string) (string, string, error) {
	log.Debugf("[slurm] sbatch --parsable %s", scriptPath)
	cmd := exec.Command("sbatch", "--parsable", scriptPath)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", cmd.String(), fmt.Errorf("sbatch failed: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	// --parsable returns "<jobid>" or "<jobid>;<cluster>".
	id := strings.TrimSpace(out.String())
	if i := strings.IndexByte(id, ';'); i > 0 {
		id = id[:i]
	}
	log.Debugf("[slurm] sbatch returned job id: %s", id)
	return id, cmd.String(), nil
}

func (s *SLURM) Status(jobID string) (string, error) {
	log.Debugf("[slurm] squeue -j %s -h -o %%T", jobID)
	cmd := exec.Command("squeue", "-j", jobID, "-h", "-o", "%T")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return s.sacctState(jobID, fmt.Sprintf("squeue: %v", err))
	}
	state := strings.TrimSpace(out.String())
	if state != "" {
		log.Debugf("[slurm] squeue state: %s", state)
		return state, nil
	}

	log.Debug("[slurm] squeue returned empty, falling back to sacct")
	return s.sacctState(jobID, "squeue: empty state")
}

func (s *SLURM) sacctState(jobID string, reason string) (string, error) {
	log.Debugf("[slurm] sacct -j %s -n -X -o State", jobID)
	sacct := exec.Command("sacct", "-j", jobID, "-n", "-X", "-o", "State")
	var sout bytes.Buffer
	sacct.Stdout = &sout
	if serr := sacct.Run(); serr != nil {
		return "", fmt.Errorf("%s; sacct: %v", reason, serr)
	}
	state := strings.TrimSpace(sout.String())
	log.Debugf("[slurm] sacct state: %s", state)
	return state, nil
}

// Kill terminates a SLURM job.
//
// Default behaviour (sig == nil) is a staged graceful kill:
//  1. scancel --full --signal=TERM <jobid>
//     --full delivers TERM to the batch script *and* its children, so a
//     batch script waiting on a foreground child gets to run its signal
//     trap (a plain scancel only signals the batch shell, whose trap is
//     then deferred until the child exits on its own; see man scancel).
//  2. Poll the job state for up to the grace period (default 30s,
//     overridable with ORVIX_SLURM_KILL_GRACE in seconds), giving trap
//     handlers time to run.
//  3. If the job is still not in a terminal state, fall back to a plain
//     scancel <jobid> (controller cancel path: SIGCONT+SIGTERM, KillWait,
//     then SIGKILL; it also marks the job CANCELLED, which the --signal
//     path never does because it bypasses slurmctld).
//
// An explicit signal keeps single-shot semantics but adds --full so the
// signal reaches the batch script's children as well.
//
// Killing a job that has already finished is treated as success.
func (s *SLURM) Kill(jobID string, sig os.Signal) error {
	if sig != nil {
		num, err := signalNumber(sig)
		if err != nil {
			return err
		}
		if err := s.scancel(jobID, "--full", "--signal="+num); err != nil {
			return s.ignoreGone(jobID, err)
		}
		return nil
	}

	if err := s.scancel(jobID, "--full", "--signal=TERM"); err != nil {
		return s.ignoreGone(jobID, err)
	}

	deadline := time.Now().Add(slurmKillGrace())
	for {
		raw, err := s.Status(jobID)
		if err != nil {
			log.Debugf("[slurm] job %s no longer visible, treating kill as done: %v", jobID, err)
			return nil
		}
		if state := s.NormalizeState(raw); state.IsTerminal() {
			log.Debugf("[slurm] job %s reached terminal state %s", jobID, state)
			return nil
		}
		if !time.Now().Before(deadline) {
			break
		}
		time.Sleep(time.Second)
	}

	log.Debugf("[slurm] job %s still alive after grace, escalating to plain scancel", jobID)
	if err := s.scancel(jobID); err != nil {
		return s.ignoreGone(jobID, err)
	}
	return nil
}

// scancel runs scancel with args followed by the job id.
func (s *SLURM) scancel(jobID string, args ...string) error {
	args = append(args, jobID)
	log.Debugf("[slurm] scancel %s", strings.Join(args, " "))
	cmd := exec.Command("scancel", args...)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("scancel: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	return nil
}

// ignoreGone converts a failed scancel into success when the job is already
// gone (scancel's own error message, or a state cross-check).
func (s *SLURM) ignoreGone(jobID string, err error) error {
	if killGone(err) || s.jobGone(jobID) {
		log.Debugf("[slurm] job %s already gone, treating kill as success (%v)", jobID, err)
		return nil
	}
	return err
}

// killGone reports whether a failed scancel invocation refers to a job that
// is already finished or unknown to the controller.
func killGone(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid job id") ||
		strings.Contains(msg, "already completing or completed")
}

// jobGone reports whether the job has reached a terminal state or is no
// longer visible to squeue/sacct at all.
func (s *SLURM) jobGone(jobID string) bool {
	raw, err := s.Status(jobID)
	if err != nil {
		return true
	}
	return s.NormalizeState(raw).IsTerminal()
}

// slurmKillGrace is the grace period between the TERM stage and the cancel
// fallback of a default Kill. Overridable with ORVIX_SLURM_KILL_GRACE
// (seconds); 0 disables the wait and escalates immediately.
func slurmKillGrace() time.Duration {
	if v := strings.TrimSpace(os.Getenv("ORVIX_SLURM_KILL_GRACE")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Second
		}
		log.Debugf("[slurm] ignoring invalid ORVIX_SLURM_KILL_GRACE=%q", v)
	}
	return 30 * time.Second
}

func (s *SLURM) NormalizeState(raw string) JobState {
	switch strings.ToUpper(raw) {
	case "RUNNING":
		return StateRunning
	case "PENDING", "CONFIGURING":
		return StatePending
	case "COMPLETED":
		return StateCompleted
	case "FAILED", "NODE_FAIL", "BOOT_FAIL", "DEADLINE", "PREEMPTED":
		return StateFailed
	case "CANCELLED":
		return StateCancelled
	case "TIMEOUT":
		return StateTimeout
	case "OUT_OF_MEMORY", "OOM":
		return StateFailed
	default:
		return StateUnknown
	}
}
