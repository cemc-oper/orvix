package scheduler

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

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

func (s *SLURM) Kill(jobID string) error {
	log.Debugf("[slurm] scancel %s", jobID)
	cmd := exec.Command("scancel", jobID)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("scancel: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	log.Debugf("[slurm] scancel succeeded for job %s", jobID)
	return nil
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
