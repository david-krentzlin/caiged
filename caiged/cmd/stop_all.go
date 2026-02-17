package cmd

import (
	"fmt"
	"strings"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/spf13/cobra"
)

func newStopAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all caiged containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")
			errorsList := make([]string, 0)

			executor := exec.NewRealExecutor()
			client := docker.NewClient(executor)

			containerIDs, err := client.ContainerListAll(fmt.Sprintf("name=^/%s-", prefix), "{{.ID}}")
			if err == nil {
				for _, containerID := range containerIDs {
					containerID = strings.TrimSpace(containerID)
					if containerID == "" {
						continue
					}
					if rmErr := client.ContainerRemove(containerID); rmErr != nil {
						errorsList = append(errorsList, fmt.Sprintf("remove container %s: %v", containerID, rmErr))
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
