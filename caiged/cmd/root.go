package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caiged [command]",
	Short: "Run isolated agent spins in Docker",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runCommand(args, runOpts, false)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	addCommonFlags(rootCmd)
	addRunFlags(rootCmd, &runOpts)

	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newBuildCmd())
	rootCmd.AddCommand(newSessionCmd())
}
