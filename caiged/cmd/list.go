package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active caiged containers and tmux sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")

			runningOutput, err := runCapture("docker", []string{"ps", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.Names}}\t{{.Status}}"}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("list running containers: %w", err)
			}
			runningLines := nonEmptyLines(runningOutput)
			if len(runningLines) == 0 {
				fmt.Println("Running containers: none")
			} else {
				fmt.Println("Running containers:")
				for _, line := range runningLines {
					fmt.Println(line)
				}
			}

			allOutput, err := runCapture("docker", []string{"ps", "-a", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.Names}}\t{{.Status}}"}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("list all containers: %w", err)
			}
			allLines := nonEmptyLines(allOutput)
			if len(allLines) == 0 {
				fmt.Println("All containers: none")
			} else {
				fmt.Println("All containers:")
				for _, line := range allLines {
					fmt.Println(line)
				}
			}

			if !commandExists("tmux") {
				fmt.Println("Tmux sessions: none")
				return nil
			}

			output, err := runCapture("tmux", []string{"list-sessions", "-F", "#{session_name}"}, ExecOptions{})
			if err != nil {
				if isBenignTmuxNoServerError(err) {
					fmt.Println("Tmux sessions: none")
					return nil
				}
				return fmt.Errorf("list tmux sessions: %w", err)
			}

			sessions := make([]string, 0)
			for _, line := range nonEmptyLines(output) {
				if strings.HasPrefix(line, prefix+"-") {
					sessions = append(sessions, line)
				}
			}
			if len(sessions) == 0 {
				fmt.Println("Tmux sessions: none")
			} else {
				fmt.Println("Tmux sessions:")
				for _, session := range sessions {
					fmt.Println(session)
				}
			}
			return nil
		},
	}
	return cmd
}

func nonEmptyLines(output string) []string {
	lines := strings.Split(output, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}
