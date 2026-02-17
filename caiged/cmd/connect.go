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
		Use:   "connect <project-name>",
		Short: "Connect to an OpenCode server with the TUI client",
		Long:  "Launch the OpenCode TUI client connected to a running container by project name (e.g., 'my-app' or 'code-myproject')",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectSlug := slugifyProjectName(args[0])
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")

			executor := exec.NewRealExecutor()
			dockerClient := docker.NewClient(executor)

			// Container names are in format: {prefix}-{spin}-{project}
			// We search for any container that ends with the project slug
			containers, err := dockerClient.ContainerList(fmt.Sprintf("name=-%s$", projectSlug), "{{.Names}}")
			if err != nil {
				return fmt.Errorf("search containers: %w", err)
			}

			containers = filterNonEmpty(containers)
			if len(containers) == 0 {
				return fmt.Errorf("no running container found for project: %s", args[0])
			}

			// Filter to only containers with the correct prefix
			var matchingContainers []string
			for _, name := range containers {
				if strings.HasPrefix(name, prefix+"-") {
					matchingContainers = append(matchingContainers, name)
				}
			}

			if len(matchingContainers) == 0 {
				return fmt.Errorf("no running container found for project: %s", args[0])
			}

			// If multiple matches, prefer the default "qa" spin
			containerName := matchingContainers[0]
			for _, name := range matchingContainers {
				if strings.HasPrefix(name, prefix+"-qa-") {
					containerName = name
					break
				}
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
