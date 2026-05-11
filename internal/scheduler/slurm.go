package scheduler

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cemc-oper/orvix/internal/directive"
)

// SLURM submits jobs via sbatch / squeue / scancel.
type SLURM struct{}

func (s *SLURM) Name() string { return "slurm" }

// PreambleFor translates orvix key=value directives to #SBATCH lines using
// long-flag form: `#SBATCH --key=value`. Bare keys (empty value) emit
// `#SBATCH --key`. Values containing whitespace are double-quoted.
func (s *SLURM) PreambleFor(d *directive.Set) ([]string, error) {
	var lines []string
	for _, item := range d.Items {
		if item.Key == "scheduler" {
			continue
		}
		if item.Value == "" {
			lines = append(lines, fmt.Sprintf("#SBATCH --%s", item.Key))
		} else {
			lines = append(lines, fmt.Sprintf("#SBATCH --%s=%s", item.Key, slurmQuote(item.Value)))
		}
	}
	return lines, nil
}

func slurmQuote(v string) string {
	if strings.ContainsAny(v, " \t") {
		return `"` + v + `"`
	}
	return v
}

func (s *SLURM) Submit(scriptPath string) (string, error) {
	cmd := exec.Command("sbatch", "--parsable", scriptPath)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("sbatch failed: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	// --parsable returns "<jobid>" or "<jobid>;<cluster>".
	id := strings.TrimSpace(out.String())
	if i := strings.IndexByte(id, ';'); i > 0 {
		id = id[:i]
	}
	return id, nil
}

func (s *SLURM) Status(jobID string) (string, error) {
	cmd := exec.Command("squeue", "-j", jobID, "-h", "-o", "%T")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// squeue forgets completed jobs quickly; fall back to sacct.
		sacct := exec.Command("sacct", "-j", jobID, "-n", "-X", "-o", "State")
		var sout bytes.Buffer
		sacct.Stdout = &sout
		if serr := sacct.Run(); serr != nil {
			return "", fmt.Errorf("squeue: %w; sacct: %v", err, serr)
		}
		return strings.TrimSpace(sout.String()), nil
	}
	state := strings.TrimSpace(out.String())
	if state == "" {
		return "UNKNOWN", nil
	}
	return state, nil
}

func (s *SLURM) Kill(jobID string) error {
	cmd := exec.Command("scancel", jobID)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("scancel: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	return nil
}
