package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

var killCmd = &cobra.Command{
	Use:   "kill <info.yaml>",
	Short: "Kill a running job referenced by a submit-generated info.yaml",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		info, err := jobinfo.Read(p)
		if err != nil {
			return err
		}
		if info.JobID == "" {
			return fmt.Errorf("info %s missing job_id", p)
		}
		sched, err := scheduler.ByName(info.Scheduler)
		if err != nil {
			return err
		}
		return sched.Kill(info.JobID)
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
}
