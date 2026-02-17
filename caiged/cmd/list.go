package cmd

import (
	"fmt"
	"strings"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active caiged containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := envOrDefault("IMAGE_PREFIX", "caiged")

			executor := exec.NewRealExecutor()
			client := docker.NewClient(executor)

			runningLines, err := client.ContainerList(fmt.Sprintf("name=^/%s-", prefix), "{{.Names}}\t{{.Status}}")
			if err != nil {
				return fmt.Errorf("list running containers: %w", err)
			}
			runningLines = filterNonEmpty(runningLines)
			if len(runningLines) == 0 {
				fmt.Println("Running containers: none")
			} else {
				fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				fmt.Println(SectionDivider.Render("  RUNNING CONTAINERS"))
				fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				for _, line := range runningLines {
					parts := strings.Split(line, "\t")
					if len(parts) < 2 {
						continue
					}
					containerName := parts[0]
					status := parts[1]

					// Extract project name from container name (remove prefix)
					projectName := strings.TrimPrefix(containerName, prefix+"-")

					// Get the port from the container label
					port, _ := client.ContainerGetLabel(containerName, "opencode.port")
					port = strings.TrimSpace(port)

					// Generate the password
					password := ""
					if pwd, err := generateOpencodePassword(containerName); err == nil {
						password = pwd
					}

					fmt.Println()
					fmt.Printf("  📦 %s %s\n", LabelStyle.Render("Project:"), ProjectStyle.Render(projectName))
					fmt.Printf("     %s %s\n", LabelStyle.Render("Container:"), containerName)
					fmt.Printf("     %s %s\n", LabelStyle.Render("Status:"), RunningStyle.Render(status))
					if port != "" {
						fmt.Printf("     %s %s\n", LabelStyle.Render("Server:"), ValueStyle.Render(fmt.Sprintf("http://localhost:%s", port)))
						if password != "" {
							fmt.Printf("     %s %s\n", LabelStyle.Render("Password:"), InfoStyle.Render(password))
						}
					}
					fmt.Println()
					fmt.Printf("     %s  caiged connect %s\n", LabelStyle.Render("Connect:"), projectName)
					fmt.Printf("     %s    docker exec -it %s /bin/zsh\n", LabelStyle.Render("Shell:"), containerName)
					fmt.Println(DividerStyle.Render("  ──────────────────────────────────────────────────────────────────"))
				}
				fmt.Println()
			}

			allLines, err := client.ContainerListAll(fmt.Sprintf("name=^/%s-", prefix), "{{.Names}}\t{{.Status}}")
			if err != nil {
				return fmt.Errorf("list all containers: %w", err)
			}
			allLines = filterNonEmpty(allLines)
			if len(allLines) == 0 {
				fmt.Println("All containers: none")
			} else {
				fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				fmt.Println(SectionDivider.Render("  ALL CONTAINERS"))
				fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				for _, line := range allLines {
					parts := strings.Split(line, "\t")
					if len(parts) < 2 {
						continue
					}
					containerName := parts[0]
					status := parts[1]

					// Extract project name from container name (remove prefix)
					projectName := strings.TrimPrefix(containerName, prefix+"-")

					// Determine if container is running
					isRunning := strings.Contains(strings.ToLower(status), "up")
					statusStyle := RunningStyle
					if !isRunning {
						statusStyle = StoppedStyle
					}

					// Get the port from the container label (only for running containers)
					port := ""
					password := ""
					if isRunning {
						port, _ = client.ContainerGetLabel(containerName, "opencode.port")
						port = strings.TrimSpace(port)
						if pwd, err := generateOpencodePassword(containerName); err == nil {
							password = pwd
						}
					}

					fmt.Println()
					fmt.Printf("  📦 %s %s\n", LabelStyle.Render("Project:"), ProjectStyle.Render(projectName))
					fmt.Printf("     %s %s\n", LabelStyle.Render("Container:"), containerName)
					fmt.Printf("     %s %s\n", LabelStyle.Render("Status:"), statusStyle.Render(status))
					if port != "" {
						fmt.Printf("     %s %s\n", LabelStyle.Render("Server:"), ValueStyle.Render(fmt.Sprintf("http://localhost:%s", port)))
						if password != "" {
							fmt.Printf("     %s %s\n", LabelStyle.Render("Password:"), InfoStyle.Render(password))
						}
					}
					fmt.Println()

					// Only show connect command for running containers
					if isRunning {
						fmt.Printf("     %s  caiged connect %s\n", LabelStyle.Render("Connect:"), projectName)
					}
					fmt.Printf("     %s   docker rm -f %s\n", LabelStyle.Render("Remove:"), containerName)
					fmt.Println(DividerStyle.Render("  ──────────────────────────────────────────────────────────────────"))
				}
				fmt.Println()
			}

			return nil
		},
	}
	return cmd
}

// filterNonEmpty filters out empty strings from a slice
func filterNonEmpty(lines []string) []string {
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
