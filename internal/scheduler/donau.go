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

// dsub returns a Generator that emits "#DSUB -flag value".
// The line is skipped when the orvix value is empty.
func dsub(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok || v == "" {
			return "", false
		}
		return fmt.Sprintf("#DSUB -%s %s", flag, v), true
	}
}

// dsubQ returns a Generator that emits "#DSUB -flag 'value'"
// with the value wrapped in single quotes.
func dsubQ(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok || v == "" {
			return "", false
		}
		return fmt.Sprintf("#DSUB -%s '%s'", flag, v), true
	}
}

// dsubDouble returns a Generator that emits "#DSUB --flag value".
func dsubDouble(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok || v == "" {
			return "", false
		}
		return fmt.Sprintf("#DSUB --%s %s", flag, v), true
	}
}

// dsubBare returns a Generator for bare Donau flags.
// When the orvix value is empty it emits "#DSUB --flag";
// otherwise "#DSUB --flag value".
func dsubBare(orvixKey, flag string) Generator {
	return func(d *directive.Set) (string, bool) {
		v, ok := d.GetOK(orvixKey)
		if !ok {
			return "", false
		}
		if v == "" {
			return fmt.Sprintf("#DSUB --%s", flag), true
		}
		return fmt.Sprintf("#DSUB --%s %s", flag, v), true
	}
}

// donauTime handles the special time-conversion logic.
func donauTime(d *directive.Set) (string, bool) {
	v, ok := d.GetOK("time")
	if !ok || v == "" {
		return "", false
	}
	if sec, err := timeToSeconds(v); err == nil {
		return fmt.Sprintf("#DSUB -T %d", sec), true
	}
	return fmt.Sprintf("#DSUB -T %s", v), true
}

// donauResources merges cpus-per-task and memory into a single -R line.
func donauResources(d *directive.Set) (string, bool) {
	var parts []string
	if v, ok := d.GetOK("cpus-per-task"); ok && v != "" {
		parts = append(parts, fmt.Sprintf("cpu=%s", v))
	}
	if v, ok := d.GetOK("memory"); ok && v != "" {
		parts = append(parts, fmt.Sprintf("mem=%s", v))
	}
	if len(parts) == 0 {
		return "", false
	}
	return fmt.Sprintf(`#DSUB -R "%s"`, strings.Join(parts, ";")), true
}

// donauDescription merges project and application into a single -d line.
func donauDescription(d *directive.Set) (string, bool) {
	proj, hasProj := d.GetOK("project")
	app, hasApp := d.GetOK("application")
	if !hasProj && !hasApp {
		return "", false
	}
	if proj != "" && app != "" {
		return fmt.Sprintf(`#DSUB -d "%s:%s"`, proj, app), true
	}
	if proj != "" {
		return fmt.Sprintf(`#DSUB -d "%s"`, proj), true
	}
	return fmt.Sprintf(`#DSUB -d "%s"`, app), true
}

// PreambleFor translates orvix generic directives to #DSUB lines.
//
// Each generator queries the directive set for the orvix keys it cares about
// and emits one scheduler-specific line. Generators that find nothing return
// ("", false) and are skipped.
//
// Special handling:
//   - project + application are combined into a single -d line.
//   - cpus-per-task and memory are merged into a single -R line.
//   - time is converted to seconds when possible.
//   - ntasks and dependency have no Donau equivalent — skipped silently.
func (d *Donau) PreambleFor(ds *directive.Set) ([]string, error) {
	generators := []Generator{
		dsub("job-name", "n"),
		dsub("output", "oo"),
		dsub("error", "eo"),
		dsub("nodes", "nn"),
		dsub("ntasks-per-node", "tpn"),
		dsub("queue", "q"),
		dsub("account", "A"),
		dsubQ("nodelist", "pn"),
		dsubDouble("job-type", "job_type"),
		donauTime,
		donauResources,   // merges cpus-per-task + memory
		donauDescription, // merges project + application
		dsubBare("exclusive", "exclusive"),
	}

	var lines []string
	for _, gen := range generators {
		if line, ok := gen(ds); ok {
			lines = append(lines, line)
		}
	}
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
