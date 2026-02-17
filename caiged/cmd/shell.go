package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell <container-name>",
		Short: "Open a shell in a container (for debugging)",
		Long:  "Open an interactive shell (zsh) in a running container for debugging and maintenance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shellCommand(args[0])
		},
	}
	return cmd
}

func shellCommand(containerName string) error {
	// Open a shell directly in the container (for debugging/maintenance)
	shell := envOrDefault("CONTAINER_SHELL", "/bin/zsh")
	return execCommand("docker", []string{"exec", "-it", containerName, shell}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}
