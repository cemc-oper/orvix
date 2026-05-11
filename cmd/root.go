// Package cmd wires the orvix CLI on top of cobra.
package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "orvix",
	Short: "Submit script tasks to HPC",
	Long: `orvix translates "# orvix" preprocessor directives in a script into
scheduler-specific directives (e.g. SLURM #SBATCH) and submits the rewritten
script for execution.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command. main() prints any returned error.
func Execute() error {
	return rootCmd.Execute()
}
