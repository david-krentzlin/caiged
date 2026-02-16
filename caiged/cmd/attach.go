package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <session-or-workdir>",
		Short: "Attach to a container shell (for debugging)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return attachCommand(args, runOpts)
		},
	}
	addAttachFlags(cmd, &runOpts)
	return cmd
}

func attachCommand(args []string, opts RunOptions) error {
	target := args[0]
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return runCommand(args, opts, true)
	}

	// Attach directly to the container shell (for debugging/maintenance)
	shell := envOrDefault("CONTAINER_SHELL", "/bin/zsh")
	return execCommand("docker", []string{"exec", "-it", target, shell}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}
