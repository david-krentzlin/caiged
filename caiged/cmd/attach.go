package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <session-or-workdir>",
		Short: "Attach to a host tmux session or container shell",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return attachCommand(args, runOpts)
		},
	}
	addRunFlags(cmd, &runOpts)
	return cmd
}

func attachCommand(args []string, opts RunOptions) error {
	target := args[0]
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return runCommand(nil, args, opts, true)
	}

	if commandExists("tmux") {
		if execCommand("tmux", []string{"has-session", "-t", target}, ExecOptions{}) == nil {
			cfg := Config{
				SessionName:    target,
				ContainerName:  target,
				ContainerShell: envOrDefault("CONTAINER_SHELL", "/bin/zsh"),
			}
			ensureTmuxWindows(cfg)
			if os.Getenv("TMUX") != "" {
				return execCommand("tmux", []string{"switch-client", "-t", target}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
			}
			return execCommand("tmux", []string{"attach", "-t", target}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
		}
	}

	shell := envOrDefault("CONTAINER_SHELL", "/bin/zsh")
	return execCommand("docker", []string{"exec", "-it", target, shell}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}
