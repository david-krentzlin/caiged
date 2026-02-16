package cmd

import "github.com/spf13/cobra"

func newResetSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-session <workdir>",
		Short: "Reset the host tmux session for a spin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := resolveConfig(runOpts, args[0])
			if err != nil {
				return err
			}
			resetSession(config)
			return nil
		},
	}
	addRunFlags(cmd, &runOpts)
	return cmd
}
