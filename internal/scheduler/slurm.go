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

// slurmFlag maps every orvix generic directive name to its SLURM long-flag
// name. An empty value means the directive has no SLURM equivalent and is
// silently skipped. Entries whose key and value are identical are listed
// explicitly for clarity and auditability.
var slurmFlag = map[string]string{
	"scheduler":       "",                // consumed by orvix
	"job-name":        "job-name",        // --job-name
	"output":          "output",          // --output
	"error":           "error",           // --error
	"nodes":           "nodes",           // --nodes
	"ntasks":          "ntasks",          // --ntasks
	"ntasks-per-node": "ntasks-per-node", // --ntasks-per-node
	"cpus-per-task":   "cpus-per-task",   // --cpus-per-task
	"time":            "time",            // --time
	"queue":           "partition",       // --partition
	"account":         "account",         // --account
	"project":         "wckey",           // --wckey
	"application":     "comment",         // --comment
	"exclusive":       "exclusive",       // --exclusive
	"nodelist":        "nodelist",        // --nodelist
	"job-type":        "",                // no SLURM equivalent
	"memory":          "mem",             // --mem
	"dependency":      "dependency",      // --dependency
}

// PreambleFor translates orvix generic directives to #SBATCH lines.
//
// Each known directive is mapped via slurmFlag (e.g. project -> --wckey).
// Directives with no SLURM equivalent (job-type) are silently skipped.
// Bare keys emit #SBATCH --flag; values with whitespace are double-quoted.
func (s *SLURM) PreambleFor(d *directive.Set) ([]string, error) {
	var lines []string
	for _, item := range d.Items {
		flag := item.Key
		if mapped, ok := slurmFlag[item.Key]; ok {
			if mapped == "" {
				continue // skip directives with no SLURM equivalent
			}
			flag = mapped
		}
		if item.Value == "" {
			lines = append(lines, fmt.Sprintf("#SBATCH --%s", flag))
		} else {
			lines = append(lines, fmt.Sprintf("#SBATCH --%s=%s", flag, slurmQuote(item.Value)))
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
