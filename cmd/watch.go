package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/scheduler"
)

var watchInterval time.Duration

var watchCmd = &cobra.Command{
	Use:   "watch [flags] <info.yaml>",
	Short: "Poll job status until the job finishes or fails",
	Long: `watch reads a submit-generated info.yaml, queries the scheduler
repeatedly, and prints the status on every poll. It exits once the job reaches
a terminal state (e.g. COMPLETED, FAILED, CANCELLED, TIMEOUT) or when the
status query itself fails.`,
	Args: cobra.ExactArgs(1),
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

		for {
			st, err := sched.Status(info.JobID)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] ERROR: %v\n", time.Now().Format(time.RFC3339), err)
				return fmt.Errorf("status query failed: %w", err)
			}

			state := sched.NormalizeState(st)
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", time.Now().Format(time.RFC3339), state)

			if state.IsTerminal() {
				return nil
			}

			time.Sleep(watchInterval)
		}
	},
}

func init() {
	watchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 5*time.Second, "Polling interval")
	rootCmd.AddCommand(watchCmd)
}
