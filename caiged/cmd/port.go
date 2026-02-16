package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newPortCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port <project-name>",
		Short: "Show the OpenCode server port for a project",
		Long:  "Show connection information for a running container by project name (e.g., 'my-app' or 'code-myproject')",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectSlug := slugifyProjectName(args[0])
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")

			// Try to find a container matching the project slug
			output, err := runCapture("docker", []string{"ps", "--filter", fmt.Sprintf("name=^/%s-.*-%s$", prefix, projectSlug), "--format", "{{.Names}}"}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("search containers: %w", err)
			}

			containers := nonEmptyLines(output)
			if len(containers) == 0 {
				return fmt.Errorf("no running container found for project: %s", args[0])
			}

			containerName := containers[0]

			// Get the port from the container label
			labelOutput, err := runCapture("docker", []string{"inspect", "--format", "{{index .Config.Labels \"opencode.port\"}}", containerName}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("inspect container: %w", err)
			}

			port := strings.TrimSpace(labelOutput)
			if port == "" {
				return fmt.Errorf("no port found for container: %s (container may be using legacy configuration)", containerName)
			}

			// Generate the password using the same method
			password, err := generateOpencodePassword(containerName)
			if err != nil {
				return fmt.Errorf("generate password: %w", err)
			}

			fmt.Printf("Container: %s\n", containerName)
			fmt.Printf("OpenCode server: http://localhost:%s\n", port)
			fmt.Printf("Attach command: opencode attach http://localhost:%s --dir /workspace --password %s\n", port, password)

			return nil
		},
	}
	return cmd
}
