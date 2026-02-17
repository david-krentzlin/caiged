package cmd

import "github.com/spf13/cobra"

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage OpenCode server sessions and containers",
	}
	cmd.AddCommand(newShellCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newStopAllCmd())
	return cmd
}
