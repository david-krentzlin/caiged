package cmd

import (
	"fmt"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	var remove bool

	cmd := &cobra.Command{
		Use:   "stop <container-name>",
		Short: "Stop a running caiged container",
		Long:  "Stop a running caiged container. By default, the container is preserved for persistent sessions. Use --remove to delete it.",
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
				if remove {
					// Container is stopped, just remove it
					fmt.Printf("Removing stopped container '%s'...\n", containerName)
					if err := client.ContainerRemove(containerName); err != nil {
						return fmt.Errorf("failed to remove container '%s': %w", containerName, err)
					}
					fmt.Printf("✓ Container '%s' removed successfully\n", containerName)
					return nil
				}
				fmt.Printf("Container '%s' is already stopped\n", containerName)
				return nil
			}

			// Stop the container
			fmt.Printf("Stopping container '%s'...\n", containerName)
			if err := client.ContainerStop(containerName); err != nil {
				return fmt.Errorf("failed to stop container '%s': %w", containerName, err)
			}

			if remove {
				// Also remove the container
				fmt.Printf("Removing container '%s'...\n", containerName)
				if err := client.ContainerRemove(containerName); err != nil {
					return fmt.Errorf("failed to remove container '%s': %w", containerName, err)
				}
				fmt.Printf("✓ Container '%s' stopped and removed successfully\n", containerName)
			} else {
				fmt.Printf("✓ Container '%s' stopped successfully (persistent session preserved)\n", containerName)
				fmt.Printf("  Resume with: caiged run . --spin <spin-name>\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&remove, "remove", "r", false, "Remove the container after stopping (deletes persistent session)")

	return cmd
}
