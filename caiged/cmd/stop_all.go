package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newStopAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all caiged containers and tmux sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")
			if commandExists("tmux") {
				output, err := runCapture("tmux", []string{"list-sessions", "-F", "#{session_name}"}, ExecOptions{})
				if err == nil {
					for _, line := range strings.Split(output, "\n") {
						line = strings.TrimSpace(line)
						if line == "" {
							continue
						}
						if strings.HasPrefix(line, prefix+"-") {
							_ = execCommand("tmux", []string{"kill-session", "-t", line}, ExecOptions{})
						}
					}
				}
			}

			containers, err := runCapture("docker", []string{"ps", "-a", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.ID}}"}, ExecOptions{})
			if err == nil {
				for _, line := range strings.Split(containers, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					_ = execCommand("docker", []string{"rm", "-f", line}, ExecOptions{})
				}
			}
			return nil
		},
	}
	return cmd
}
