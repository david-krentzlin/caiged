package cmd

import "github.com/spf13/cobra"

func newContainersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "containers",
		Short: "Manage caiged containers",
	}
	cmd.AddCommand(newShellCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newStopAllCmd())
	return cmd
}
