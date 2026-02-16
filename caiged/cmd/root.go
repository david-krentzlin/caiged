package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caiged [command] [flags]",
	Short: "Run isolated agent spins in Docker",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	addCommonFlags(rootCmd)

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newSessionCmd())
}
