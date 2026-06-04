package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/submit"
)

var (
	generateScheduler string
	generateOutScript string
)

var generateCmd = &cobra.Command{
	Use:   "generate [flags] <script>",
	Short: "Generate a scheduler-specific runnable script without submitting",
	Long: `Generate reads <script>, parses #ORVIX directives, and generates a
scheduler-specific runnable script next to the original. Unlike submit, it
does not execute the submission command or create a job info sidecar.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := submit.Generate(submit.GenerateOptions{
			ScriptPath: args[0],
			Scheduler:  generateScheduler,
			OutScript:  generateOutScript,
			Out:        cmd.OutOrStdout(),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Generated: %s\n", path)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&generateScheduler, "scheduler", "", "Override the scheduler (slurm, local, etc.)")
	generateCmd.Flags().StringVar(&generateOutScript, "output-script", "", "Path for the generated runnable script (default: <orig>.submit next to the input script)")
	rootCmd.AddCommand(generateCmd)
}
