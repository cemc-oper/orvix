package scheduler

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/log"
)

// Donau submits jobs via dsub / djob (Huawei Donau scheduler).
type Donau struct{}

func (d *Donau) Name() string { return "donau" }

// donauSimpleMapping describes a straightforward 1:1 directive-to-flag mapping.
type donauSimpleMapping struct {
	key      string // orvix directive name
	flag     string // Donau flag (without leading - or --)
	double   bool   // true if the flag needs -- prefix (e.g. --job_type)
	quote    bool   // whether the value should be quoted
	optional bool   // true if the directive may be omitted when value is empty
}

// PreambleFor translates orvix generic directives to #DSUB lines.
//
// Special handling that breaks the naive 1:1 pattern:
//   - project + application are combined into a single -d "project:application" line.
//   - cpus-per-task and memory are merged into a single -R "cpu=X;mem=Y" line.
//   - time is converted to seconds when it looks like HH:MM:SS or MM:SS;
//     otherwise it is passed through (Donau also accepts duration strings like 8h).
//   - ntasks has no Donau equivalent and is silently skipped.
//   - dependency has no Donau equivalent and is silently skipped.
func (d *Donau) PreambleFor(ds *directive.Set) ([]string, error) {
	var lines []string

	// Emit simple 1:1 mappings in a stable order.
	simples := []donauSimpleMapping{
		{"job-name", "n", false, false, true},
		{"output", "oo", false, false, true},
		{"error", "eo", false, false, true},
		{"nodes", "nn", false, false, true},
		{"ntasks-per-node", "tpn", false, false, true},
		{"queue", "q", false, false, true},
		{"account", "A", false, false, true},
		{"nodelist", "pn", false, true, true},
	}

	for _, m := range simples {
		if !ds.Has(m.key) {
			continue
		}
		val := ds.Get(m.key)
		if m.optional && val == "" {
			continue
		}
		prefix := "-"
		if m.double {
			prefix = "--"
		}
		if m.quote {
			lines = append(lines, fmt.Sprintf("#DSUB %s%s '%s'", prefix, m.flag, val))
		} else {
			lines = append(lines, fmt.Sprintf("#DSUB %s%s %s", prefix, m.flag, val))
		}
	}

	// job-type: --job_type [value].
	if ds.Has("job-type") {
		val := ds.Get("job-type")
		if val != "" {
			lines = append(lines, fmt.Sprintf("#DSUB --job_type %s", val))
		}
	}

	// time: -T (convert HH:MM:SS / MM:SS to seconds; pass through otherwise).
	if ds.Has("time") {
		t := ds.Get("time")
		if sec, err := timeToSeconds(t); err == nil {
			lines = append(lines, fmt.Sprintf("#DSUB -T %d", sec))
		} else {
			lines = append(lines, fmt.Sprintf("#DSUB -T %s", t))
		}
	}

	// Build the combined -R resource line from cpus-per-task and memory.
	var resParts []string
	if ds.Has("cpus-per-task") {
		resParts = append(resParts, fmt.Sprintf("cpu=%s", ds.Get("cpus-per-task")))
	}
	if ds.Has("memory") {
		resParts = append(resParts, fmt.Sprintf("mem=%s", ds.Get("memory")))
	}
	if len(resParts) > 0 {
		lines = append(lines, fmt.Sprintf("#DSUB -R \"%s\"", strings.Join(resParts, ";")))
	}

	// project + application combined into a single -d line.
	hasProj := ds.Has("project")
	hasApp := ds.Has("application")
	if hasProj || hasApp {
		proj := ds.Get("project")
		app := ds.Get("application")
		if proj != "" && app != "" {
			lines = append(lines, fmt.Sprintf("#DSUB -d \"%s:%s\"", proj, app))
		} else if proj != "" {
			lines = append(lines, fmt.Sprintf("#DSUB -d \"%s\"", proj))
		} else if app != "" {
			lines = append(lines, fmt.Sprintf("#DSUB -d \"%s\"", app))
		}
	}

	// exclusive: --exclusive [value]
	if ds.Has("exclusive") {
		val := ds.Get("exclusive")
		if val == "" {
			lines = append(lines, "#DSUB --exclusive")
		} else {
			lines = append(lines, fmt.Sprintf("#DSUB --exclusive %s", val))
		}
	}

	// ntasks and dependency have no Donau equivalent — skip silently.

	log.Debugf("[donau] preamble: %d line(s)", len(lines))
	return lines, nil
}

// timeToSeconds converts an orvix time value to seconds when possible.
// Supported formats:
//   - pure integer (already seconds)
//   - MM:SS
//   - HH:MM:SS
// Anything else (e.g. "8h", "1h30m") returns an error so the caller can
// pass it through unchanged.
func timeToSeconds(t string) (int, error) {
	// Pure integer = seconds.
	if sec, err := strconv.Atoi(t); err == nil {
		return sec, nil
	}

	parts := strings.Split(t, ":")
	switch len(parts) {
	case 2: // MM:SS
		m, err1 := strconv.Atoi(parts[0])
		s, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || s < 0 || s > 59 || m < 0 {
			return 0, fmt.Errorf("invalid time format %q", t)
		}
		return m*60 + s, nil
	case 3: // HH:MM:SS
		h, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		s, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil ||
			m < 0 || m > 59 || s < 0 || s > 59 || h < 0 {
			return 0, fmt.Errorf("invalid time format %q", t)
		}
		return h*3600 + m*60 + s, nil
	}
	return 0, fmt.Errorf("unsupported time format %q", t)
}

func (d *Donau) Submit(scriptPath string) (string, string, error) {
	log.Debugf("[donau] dsub -s %s", scriptPath)
	cmd := exec.Command("dsub", "-s", scriptPath)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		outStr := strings.TrimSpace(out.String())
		errStr := strings.TrimSpace(errBuf.String())
		var b strings.Builder
		b.WriteString("dsub failed")
		if errStr != "" {
			b.WriteString("\nstderr: ")
			b.WriteString(errStr)
		}
		if outStr != "" {
			b.WriteString("\nstdout: ")
			b.WriteString(outStr)
		}
		return "", cmd.String(), fmt.Errorf("%s: %w", b.String(), err)
	}

	// dsub stdout contains a table:
	//   JOBID        MESSAGE
	//   8678302      Submit job successfully...
	// We extract the first field of the line immediately after the JOBID header.
	var id string
	lines := strings.Split(out.String(), "\n")
	for i, line := range lines {
		if strings.Contains(line, "JOBID") && i+1 < len(lines) {
			fields := strings.Fields(lines[i+1])
			if len(fields) > 0 {
				id = fields[0]
				break
			}
		}
	}
	// Fallback: first line whose first field is a pure number.
	if id == "" {
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				if _, err := strconv.Atoi(fields[0]); err == nil {
					id = fields[0]
					break
				}
			}
		}
	}

	log.Debugf("[donau] dsub returned job id: %s", id)
	return id, cmd.String(), nil
}

func (d *Donau) Status(jobID string) (string, error) {
	log.Debugf("[donau] djob -L %s.0", jobID)
	cmd := exec.Command("djob", "-L", jobID+".0")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("djob failed: %w", err)
	}
	// Parse djob -L output looking for a STATE line.
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(line), "STATE") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				state := parts[1]
				log.Debugf("[donau] djob state: %s", state)
				return state, nil
			}
		}
	}
	return "", fmt.Errorf("djob: no STATE field found for job %s", jobID)
}

func (d *Donau) Kill(jobID string) error {
	log.Debugf("[donau] djob -T %s", jobID)
	cmd := exec.Command("djob", "-T", jobID)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("djob -T: %s: %w", strings.TrimSpace(errBuf.String()), err)
	}
	log.Debugf("[donau] djob -T succeeded for job %s", jobID)
	return nil
}

func (d *Donau) NormalizeState(raw string) JobState {
	switch strings.ToUpper(raw) {
	case "RUNNING":
		return StateRunning
	case "PENDING", "WAITING", "QUEUED", "CONFIGURING":
		return StatePending
	case "COMPLETED", "DONE", "FINISHED":
		return StateCompleted
	case "FAILED", "FAILURE", "ABORTED", "BOOT_FAIL", "NODE_FAIL", "DEADLINE", "PREEMPTED", "OUT_OF_MEMORY", "OOM":
		return StateFailed
	case "CANCELLED", "CANCELED":
		return StateCancelled
	case "TIMEOUT":
		return StateTimeout
	default:
		return StateUnknown
	}
}
