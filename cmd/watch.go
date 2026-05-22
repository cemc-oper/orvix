package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/watch"
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
		return watch.FromInfoFile(args[0], watchInterval, cmd.OutOrStdout())
	},
}

func init() {
	watchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 5*time.Second, "Polling interval")
	rootCmd.AddCommand(watchCmd)
}
