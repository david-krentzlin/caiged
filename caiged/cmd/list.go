package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active caiged containers and tmux sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")
			fmt.Println("Running containers:")
			_ = execCommand("docker", []string{"ps", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.Names}}\t{{.Status}}"}, ExecOptions{Stdout: os.Stdout, Stderr: os.Stderr})
			fmt.Println()
			fmt.Println("All containers:")
			_ = execCommand("docker", []string{"ps", "-a", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.Names}}\t{{.Status}}"}, ExecOptions{Stdout: os.Stdout, Stderr: os.Stderr})
			fmt.Println()
			if commandExists("tmux") {
				fmt.Println("Tmux sessions:")
				output, err := runCapture("tmux", []string{"list-sessions", "-F", "#{session_name}"}, ExecOptions{})
				if err == nil {
					lines := strings.Split(output, "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" {
							continue
						}
						if strings.HasPrefix(line, prefix+"-") {
							fmt.Println(line)
						}
					}
				} else {
					fmt.Println("(none)")
				}
			} else {
				fmt.Println("Tmux sessions: tmux not available")
			}
			return nil
		},
	}
	return cmd
}
