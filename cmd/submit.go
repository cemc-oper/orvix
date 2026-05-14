package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/directive"
	"github.com/cemc-oper/orvix/internal/jobinfo"
	"github.com/cemc-oper/orvix/internal/scheduler"
	"github.com/cemc-oper/orvix/internal/script"
)

var (
	submitDryRun     bool
	submitScheduler  string
	submitOutScript  string
	submitOutInfo    string
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
		origPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}

		src, err := os.ReadFile(origPath)
		if err != nil {
			return fmt.Errorf("read script: %w", err)
		}

		directives, err := directive.ParseWithOverride(src, submitScheduler)
		if err != nil {
			return fmt.Errorf("parse directives: %w", err)
		}

		sched, err := scheduler.For(directives)
		if err != nil {
			return fmt.Errorf("resolve scheduler: %w", err)
		}

		generated, err := script.Generate(src, directives, sched)
		if err != nil {
			return fmt.Errorf("generate script: %w", err)
		}

		if submitDryRun {
			_, err := cmd.OutOrStdout().Write(generated)
			return err
		}

		now := time.Now()
		genScriptPath, yamlPath := derivePaths(origPath, submitOutScript, submitOutInfo)

		if err := os.WriteFile(genScriptPath, generated, 0o755); err != nil {
			return fmt.Errorf("write generated script: %w", err)
		}

		jobID, err := sched.Submit(genScriptPath)
		if err != nil {
			return fmt.Errorf("submit: %w", err)
		}

		cwd, _ := os.Getwd()
		info := jobinfo.JobInfo{
			Scheduler:       sched.Name(),
			JobID:           jobID,
			SubmittedAt:     now,
			ScriptSource:    origPath,
			ScriptGenerated: genScriptPath,
			SubmitDir:       cwd,
			Hostname:        hostnameOrEmpty(),
			User:            usernameOrEmpty(),
			Directives:      jobinfo.FromDirectives(directives),
		}
		if err := jobinfo.Write(yamlPath, info); err != nil {
			return fmt.Errorf("write job info: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), jobID)
		return nil
	},
}

// derivePaths returns the (script, yaml) sidecar paths.
//
// By default they are placed next to origPath:
//   - <orig>.submit      (translated runnable script)
//   - <orig>.info.yaml   (job metadata)
//
// If outScript or outInfo are non-empty they override the default paths.
// Resubmits overwrite.
func derivePaths(origPath, outScript, outInfo string) (string, string) {
	var scriptPath, yamlPath string
	if outScript != "" {
		scriptPath = outScript
	} else {
		dir := filepath.Dir(origPath)
		base := filepath.Base(origPath)
		scriptPath = filepath.Join(dir, base+".submit")
	}
	if outInfo != "" {
		yamlPath = outInfo
	} else {
		dir := filepath.Dir(origPath)
		base := filepath.Base(origPath)
		yamlPath = filepath.Join(dir, base+".info.yaml")
	}
	return scriptPath, yamlPath
}

func hostnameOrEmpty() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

func usernameOrEmpty() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

func init() {
	submitCmd.Flags().BoolVar(&submitDryRun, "dry-run", false, "Print the generated script and exit without submitting")
	submitCmd.Flags().StringVar(&submitScheduler, "scheduler", "", "Override the scheduler (slurm, local, etc.)")
	submitCmd.Flags().StringVar(&submitOutScript, "output-script", "", "Path for the generated runnable script (default: <orig>.<ext>.submit next to the input script)")
	submitCmd.Flags().StringVar(&submitOutInfo, "output-info", "", "Path for the job info YAML sidecar (default: <orig>.<ext>.info.yaml next to the input script)")
	rootCmd.AddCommand(submitCmd)
}
