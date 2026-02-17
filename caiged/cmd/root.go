package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caiged",
	Short: "Run isolated OpenCode agent spins in Docker",
	Long: `caiged - Run isolated OpenCode agent spins in Docker

Available commands:
  run       Start or resume a container with an OpenCode spin
  connect   Connect to an existing container's OpenCode server
  session   Manage container sessions (list, stop, shell)

Examples:
  caiged run . --spin qa      # Run qa spin in current directory
  caiged connect my-project   # Connect to existing project
  caiged session list         # List all containers
  caiged session shell <name> # Open shell in container`,
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

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <workdir>",
		Short: "Start or resume a container with an OpenCode spin",
		Long: `Start or resume a container with an OpenCode spin.

This command will:
  - Build images if they don't exist (or --rebuild-images is set)
  - Connect to an existing container for this project/spin if one exists
  - Create a new container and connect if none exists

Examples:
  caiged run . --spin qa                    # Run qa spin in current directory
  caiged run /path/to/project --spin dev    # Run dev spin for a specific path
  caiged run . --spin qa --no-connect       # Start container but don't connect`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(args, runOpts, false)
		},
	}
	addRunFlags(cmd, &runOpts)
	return cmd
}

func init() {
	addCommonFlags(rootCmd)

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newSessionCmd())
	rootCmd.AddCommand(newConnectCmd())
}
