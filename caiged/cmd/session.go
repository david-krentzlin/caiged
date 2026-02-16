package cmd

import "github.com/spf13/cobra"

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage host tmux sessions",
	}
	cmd.AddCommand(newAttachCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRestartCmd())
	cmd.AddCommand(newResetSessionCmd())
	cmd.AddCommand(newStopAllCmd())
	return cmd
}
