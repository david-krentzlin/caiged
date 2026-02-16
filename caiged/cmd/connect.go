package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

			// Container names are in format: {prefix}-{spin}-{project}
			// We search for any container that ends with the project slug
			output, err := runCapture("docker", []string{"ps", "--filter", fmt.Sprintf("name=-%s$", projectSlug), "--format", "{{.Names}}"}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("search containers: %w", err)
			}

			containers := nonEmptyLines(output)
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

			// Execute opencode attach to connect to the server
			// Try to get the last session ID and continue it
			url := fmt.Sprintf("http://localhost:%s", port)
			connectArgs := []string{"attach", url, "--dir", "/workspace", "--password", password}

			// Query the container for the most recent session and continue it
			lastSessionID := getLastSessionIDForContainer(containerName)
			if lastSessionID != "" {
				connectArgs = append(connectArgs, "--session", lastSessionID)
			}

			return execCommand("opencode", connectArgs, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
		},
	}
	return cmd
}

func getLastSessionIDForContainer(containerName string) string {
	// Query the container for the most recent session
	// We run: docker exec <container> opencode session list --format json -n 1
	output, err := runCapture("docker", []string{
		"exec", containerName,
		"opencode", "session", "list", "--format", "json", "-n", "1",
	}, ExecOptions{})

	if err != nil {
		return ""
	}

	// Parse the JSON output to get the session ID
	// The output should be a JSON array with session objects
	var sessions []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &sessions); err != nil {
		return ""
	}

	if len(sessions) == 0 {
		return ""
	}

	// Get the ID from the first (most recent) session
	if id, ok := sessions[0]["id"].(string); ok {
		return id
	}

	return ""
}
