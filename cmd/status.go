package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

var statusFormat string

// statusJSON is the machine-readable output of `status --format json`.
// State is the normalized scheduler-agnostic state; RawState keeps the
// scheduler-native string for forensics.
type statusJSON struct {
	JobID     string `json:"job_id"`
	Scheduler string `json:"scheduler"`
	State     string `json:"state"`
	RawState  string `json:"raw_state"`
	Terminal  bool   `json:"terminal"`
}

func newStatusJSON(info *jobinfo.JobInfo, sched scheduler.Scheduler, raw string) statusJSON {
	normalized := sched.NormalizeState(raw)
	return statusJSON{
		JobID:     info.JobID,
		Scheduler: info.Scheduler,
		State:     string(normalized),
		RawState:  raw,
		Terminal:  normalized.IsTerminal(),
	}
}

var statusCmd = &cobra.Command{
	Use:   "status <info.yaml>",
	Short: "Query job status from a submit-generated info.yaml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if statusFormat != "text" && statusFormat != "json" {
			return fmt.Errorf("unknown --format %q (want text or json)", statusFormat)
		}

		p, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		log.Debugf("[status] reading info file: %s", p)

		info, err := jobinfo.Read(p)
		if err != nil {
			return err
		}
		if info.JobID == "" {
			return fmt.Errorf("info %s missing job_id", p)
		}
		log.Debugf("[status] info: scheduler=%s job_id=%s", info.Scheduler, info.JobID)

		sched, err := scheduler.ByName(info.Scheduler)
		if err != nil {
			return err
		}
		st, err := sched.Status(info.JobID)
		if err != nil {
			return err
		}
		log.Debugf("[status] raw state: %s, normalized: %s", st, sched.NormalizeState(st))

		if statusFormat == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(newStatusJSON(&info, sched, st))
		}
		fmt.Fprintln(cmd.OutOrStdout(), st)
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusFormat, "format", "text",
		"output format: text (raw scheduler state, default) or json (adds the normalized state)")
	rootCmd.AddCommand(statusCmd)
}
