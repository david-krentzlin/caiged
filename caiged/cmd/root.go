package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caiged [workdir] [flags]",
	Short: "Run isolated OpenCode agent spins in Docker",
	Long: `caiged - Run isolated OpenCode agent spins in Docker

Without a subcommand, caiged will:
  - Attach to an existing container for this project/spin if one exists
  - Create a new container and attach if none exists

Examples:
  caiged .                    # Run/attach to default spin (qa) in current directory
  caiged . --spin dev         # Run/attach to dev spin
  caiged connect <project>    # Connect to a project by name (from any directory)
  caiged build                # Build Docker images
  caiged session list         # List all containers`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no args provided, default to current directory
		if len(args) == 0 {
			args = []string{"."}
		}
		// Default behavior: run (which will attach if exists, or create if not)
		return runCommand(args, runOpts, false)
	},
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
	addRunFlags(rootCmd, &runOpts) // Add run flags to root command

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newSessionCmd())
	rootCmd.AddCommand(newPortCmd())
	rootCmd.AddCommand(newConnectCmd())
}
