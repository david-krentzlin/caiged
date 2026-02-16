package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	projectStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	labelStyle     = lipgloss.NewStyle().Bold(true)
	runningStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	stoppedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dividerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sectionDivider = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active caiged containers",
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
				fmt.Println(sectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				fmt.Println(sectionDivider.Render("  RUNNING CONTAINERS"))
				fmt.Println(sectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				for _, line := range runningLines {
					parts := strings.Split(line, "\t")
					if len(parts) < 2 {
						continue
					}
					containerName := parts[0]
					status := parts[1]

					// Extract project name from container name (remove prefix)
					projectName := strings.TrimPrefix(containerName, prefix+"-")

					fmt.Println()
					fmt.Printf("  📦 %s %s\n", labelStyle.Render("Project:"), projectStyle.Render(projectName))
					fmt.Printf("     Container: %s\n", containerName)
					fmt.Printf("     %s %s\n", labelStyle.Render("Status:"), runningStyle.Render(status))
					fmt.Println()
					fmt.Printf("     %s  caiged connect %s\n", labelStyle.Render("Connect:"), projectName)
					fmt.Printf("     %s    docker exec -it %s /bin/zsh\n", labelStyle.Render("Shell:"), containerName)
					fmt.Println(dividerStyle.Render("  ──────────────────────────────────────────────────────────────────"))
				}
				fmt.Println()
			}

			allOutput, err := runCapture("docker", []string{"ps", "-a", "--filter", fmt.Sprintf("name=^/%s-", prefix), "--format", "{{.Names}}\t{{.Status}}"}, ExecOptions{})
			if err != nil {
				return fmt.Errorf("list all containers: %w", err)
			}
			allLines := nonEmptyLines(allOutput)
			if len(allLines) == 0 {
				fmt.Println("All containers: none")
			} else {
				fmt.Println(sectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
				fmt.Println(sectionDivider.Render("  ALL CONTAINERS"))
				fmt.Println(sectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
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
					statusStyle := runningStyle
					if !isRunning {
						statusStyle = stoppedStyle
					}

					fmt.Println()
					fmt.Printf("  📦 %s %s\n", labelStyle.Render("Project:"), projectStyle.Render(projectName))
					fmt.Printf("     Container: %s\n", containerName)
					fmt.Printf("     %s %s\n", labelStyle.Render("Status:"), statusStyle.Render(status))
					fmt.Println()

					// Only show connect command for running containers
					if isRunning {
						fmt.Printf("     %s  caiged connect %s\n", labelStyle.Render("Connect:"), projectName)
					}
					fmt.Printf("     %s   docker rm -f %s\n", labelStyle.Render("Remove:"), containerName)
					fmt.Println(dividerStyle.Render("  ──────────────────────────────────────────────────────────────────"))
				}
				fmt.Println()
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
