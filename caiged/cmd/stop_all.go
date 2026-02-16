package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newStopAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all caiged containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")
			errorsList := make([]string, 0)

			containers, err := runCapture("docker", []string{"ps", "-a", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.ID}}"}, ExecOptions{})
			if err == nil {
				for _, line := range strings.Split(containers, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					if rmErr := execCommand("docker", []string{"rm", "-f", line}, ExecOptions{}); rmErr != nil {
						errorsList = append(errorsList, fmt.Sprintf("remove container %s: %v", line, rmErr))
					}
				}
			} else {
				errorsList = append(errorsList, fmt.Sprintf("list containers: %v", err))
			}

			if len(errorsList) > 0 {
				return fmt.Errorf("stop-all completed with errors: %s", strings.Join(errorsList, "; "))
			}
			return nil
		},
	}
	return cmd
}
