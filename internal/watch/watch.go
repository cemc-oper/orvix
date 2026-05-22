// Package watch provides the job-status polling logic shared by submit --watch
// and the standalone watch command.
package watch

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

// Run polls the scheduler for job status every interval until the job reaches a
// terminal state or the status query fails.  Status lines are written to out
// with an RFC3339 timestamp prefix.
func Run(sched scheduler.Scheduler, jobID string, interval time.Duration, out io.Writer) error {
	log.Debugf("[watch] start polling job %s every %s", jobID, interval)
	for {
		st, err := sched.Status(jobID)
		if err != nil {
			log.Debugf("[watch] status query failed: %v", err)
			fmt.Fprintf(out, "[%s] ERROR: %v\n", time.Now().Format(time.RFC3339), err)
			return fmt.Errorf("status query failed: %w", err)
		}

		state := sched.NormalizeState(st)
		log.Debugf("[watch] job %s state: %s (raw: %s)", jobID, state, st)
		fmt.Fprintf(out, "[%s] %s\n", time.Now().Format(time.RFC3339), state)

		if state.IsTerminal() {
			log.Debugf("[watch] job %s reached terminal state %s", jobID, state)
			return nil
		}

		time.Sleep(interval)
	}
}

// FromInfoFile reads a submit-generated info.yaml, resolves the scheduler, and
// calls Run to poll the job until it reaches a terminal state.
func FromInfoFile(infoPath string, interval time.Duration, out io.Writer) error {
	p, err := filepath.Abs(infoPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	log.Debugf("[watch] reading info file: %s", p)

	info, err := jobinfo.Read(p)
	if err != nil {
		return err
	}
	if info.JobID == "" {
		return fmt.Errorf("info %s missing job_id", p)
	}
	log.Debugf("[watch] info: scheduler=%s job_id=%s", info.Scheduler, info.JobID)

	sched, err := scheduler.ByName(info.Scheduler)
	if err != nil {
		return err
	}

	return Run(sched, info.JobID, interval, out)
}
