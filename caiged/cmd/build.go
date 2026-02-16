package cmd

import "github.com/spf13/cobra"

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build <workdir>",
		Short: "Build the base and spin container images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildCommand(args, runOpts)
		},
	}
	addBuildFlags(cmd, &runOpts)
	return cmd
}

func buildCommand(args []string, opts RunOptions) error {
	config, err := resolveConfig(opts, args[0])
	if err != nil {
		return err
	}

	if err := buildImage(config, "base"); err != nil {
		return err
	}
	return buildImage(config, "spin")
}
