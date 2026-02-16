package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caiged [command] [flags]",
	Short: "Run isolated OpenCode agent spins in Docker",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func Execute() {
	// Silence cobra's built-in error printing, we'll style our own
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s %s\n\n", ErrorStyle.Render("✗ Error:"), err.Error())
		os.Exit(1)
	}
}

func init() {
	addCommonFlags(rootCmd)

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newSessionCmd())
	rootCmd.AddCommand(newPortCmd())
	rootCmd.AddCommand(newConnectCmd())
}
