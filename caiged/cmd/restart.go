package cmd

import "github.com/spf13/cobra"

func newRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart <workdir>",
		Short: "Stop container and reset tmux session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return restartCommand(args, runOpts)
		},
	}
	addRunFlags(cmd, &runOpts)
	return cmd
}

func restartCommand(args []string, opts RunOptions) error {
	config, err := resolveConfig(opts, args[0])
	if err != nil {
		return err
	}

	resetSession(config)
	stopContainer(config)

	return runCommand(args, opts, false)
}
