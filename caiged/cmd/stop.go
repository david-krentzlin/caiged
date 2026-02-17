package cmd

import (
	"fmt"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <container-name>",
		Short: "Stop a running caiged container (without removing it)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerName := args[0]

			executor := exec.NewRealExecutor()
			client := docker.NewClient(executor)

			// Check if container exists
			if !client.ContainerExists(containerName) {
				return fmt.Errorf("container '%s' does not exist", containerName)
			}

			// Check if container is running
			if !client.ContainerIsRunning(containerName) {
				fmt.Printf("Container '%s' is already stopped\n", containerName)
				return nil
			}

			// Stop the container
			fmt.Printf("Stopping container '%s'...\n", containerName)
			if err := client.ContainerStop(containerName); err != nil {
				return fmt.Errorf("failed to stop container '%s': %w", containerName, err)
			}

			fmt.Printf("✓ Container '%s' stopped successfully\n", containerName)
			return nil
		},
	}
	return cmd
}
