package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/log"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

var statusCmd = &cobra.Command{
	Use:   "status <info.yaml>",
	Short: "Query job status from a submit-generated info.yaml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		fmt.Fprintln(cmd.OutOrStdout(), st)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
