// Package cmd wires the orvix CLI on top of cobra.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/log"
)

var rootCmd = &cobra.Command{
	Use:   "orvix",
	Short: "Submit script tasks to HPC",
	Long: `orvix translates "# orvix" preprocessor directives in a script into
scheduler-specific directives (e.g. SLURM #SBATCH) and submits the rewritten
script for execution.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setDebug()
	},
}

// Execute runs the root command. main() prints any returned error.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugEnabled, "debug", false, "Enable debug logging")
}

var debugEnabled bool

func setDebug() {
	log.SetDebug(debugEnabled)
}
