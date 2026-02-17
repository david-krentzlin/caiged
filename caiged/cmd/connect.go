package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/david-krentzlin/caiged/caiged/internal/opencode"
	"github.com/spf13/cobra"
)

func newConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <container-name>",
		Short: "Connect to an OpenCode server with the TUI client",
		Long:  "Launch the OpenCode TUI client connected to a running container by container name (e.g., 'caiged-qa-my-project')",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerName := args[0]

			executor := exec.NewRealExecutor()
			dockerClient := docker.NewClient(executor)

			// Verify container exists and is running
			if !dockerClient.ContainerExists(containerName) {
				return fmt.Errorf("container '%s' does not exist", containerName)
			}

			if !dockerClient.ContainerIsRunning(containerName) {
				return fmt.Errorf("container '%s' is not running (resume with: caiged run .)", containerName)
			}

			// Get the port from the container label
			port, err := dockerClient.ContainerGetLabel(containerName, "opencode.port")
			if err != nil {
				return fmt.Errorf("inspect container: %w", err)
			}

			port = strings.TrimSpace(port)
			if port == "" {
				return fmt.Errorf("no port found for container: %s (container may be using legacy configuration)", containerName)
			}

			// Generate the password using the same method
			password, err := generateOpencodePassword(containerName)
			if err != nil {
				return fmt.Errorf("generate password: %w", err)
			}

			// Query the container for the most recent session and continue it
			lastSessionID, err := opencode.GetLastSessionFromContainer(
				func(name string, cmd []string) (string, error) {
					return dockerClient.ContainerExecCapture(name, cmd)
				},
				containerName,
			)
			if err == nil && lastSessionID != "" {
				fmt.Printf("%s\n", InfoStyle.Render(opencode.FormatSessionResumptionMessage(lastSessionID)))
			}

			// Connect to the OpenCode server using opencode client
			opencodeClient := opencode.NewClient(executor).WithOutput(os.Stdout, os.Stderr, os.Stdin)
			url := fmt.Sprintf("http://localhost:%s", port)

			return opencodeClient.Attach(opencode.AttachConfig{
				URL:       url,
				Workdir:   "/workspace",
				Password:  password,
				SessionID: lastSessionID,
			})
		},
	}
	return cmd
}
