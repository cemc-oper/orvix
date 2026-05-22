package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cemc-oper/orvix/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print orvix version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
