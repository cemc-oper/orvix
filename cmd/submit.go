package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/submit"
)

var (
	submitDryRun        bool
	submitScheduler     string
	submitOutScript     string
	submitOutInfo       string
	submitWatch         bool
	submitWatchInterval time.Duration
)

var submitCmd = &cobra.Command{
	Use:   "submit [flags] <script>",
	Short: "Submit a job script",
	Long: `Submit reads <script>, parses #ORVIX directives, generates a
scheduler-specific runnable script next to the original, submits it, and
writes a YAML sidecar with the job metadata. The job id is printed on
success.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return submit.Run(submit.Options{
			ScriptPath:    args[0],
			Scheduler:     submitScheduler,
			DryRun:        submitDryRun,
			OutScript:     submitOutScript,
			OutInfo:       submitOutInfo,
			Watch:         submitWatch,
			WatchInterval: submitWatchInterval,
			Out:           cmd.OutOrStdout(),
		})
	},
}

func init() {
	submitCmd.Flags().BoolVar(&submitDryRun, "dry-run", false, "Print the generated script and exit without submitting")
	submitCmd.Flags().StringVar(&submitScheduler, "scheduler", "", "Override the scheduler (slurm, local, etc.)")
	submitCmd.Flags().StringVar(&submitOutScript, "output-script", "", "Path for the generated runnable script (default: <orig>.<ext>.submit next to the input script)")
	submitCmd.Flags().StringVar(&submitOutInfo, "output-info", "", "Path for the job info YAML sidecar (default: <orig>.<ext>.info.yaml next to the input script)")
	submitCmd.Flags().BoolVar(&submitWatch, "watch", false, "After successful submission, poll job status until it reaches a terminal state")
	submitCmd.Flags().DurationVar(&submitWatchInterval, "watch-interval", 5*time.Second, "Polling interval when --watch is used")
	rootCmd.AddCommand(submitCmd)
}
