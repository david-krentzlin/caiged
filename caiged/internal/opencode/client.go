package opencode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/david-krentzlin/caiged/caiged/internal/exec"
)

// Client wraps OpenCode CLI operations
type Client struct {
	executor exec.CmdExecutor
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
}

// NewClient creates a new OpenCode client with the given executor
func NewClient(executor exec.CmdExecutor) *Client {
	return &Client{
		executor: executor,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		stdin:    os.Stdin,
	}
}

// WithOutput sets custom stdout/stderr/stdin for the client
func (c *Client) WithOutput(stdout, stderr io.Writer, stdin io.Reader) *Client {
	return &Client{
		executor: c.executor,
		stdout:   stdout,
		stderr:   stderr,
		stdin:    stdin,
	}
}

// AttachConfig holds configuration for attaching to OpenCode server
type AttachConfig struct {
	URL       string
	Workdir   string
	Password  string
	SessionID string
}

// Attach connects to an OpenCode server interactively
func (c *Client) Attach(cfg AttachConfig) error {
	args := []string{"attach", cfg.URL, "--dir", cfg.Workdir, "--password", cfg.Password}

	if cfg.SessionID != "" {
		args = append(args, "--session", cfg.SessionID)
	}

	return c.executor.Run("opencode", args, exec.RunOptions{
		Stdin:  c.stdin,
		Stdout: c.stdout,
		Stderr: c.stderr,
	})
}

// Session represents an OpenCode session
type Session struct {
	ID        string
	Workdir   string
	CreatedAt string
}

// GetLastSessionFromContainer retrieves the most recent session ID from a container's storage
// Sessions are stored as files in /root/.local/share/opencode/storage/session_diff/
func GetLastSessionFromContainer(dockerExec func(containerName string, command []string) (string, error), containerName string) (string, error) {
	// Execute command in container to list session files sorted by modification time
	output, err := dockerExec(containerName, []string{
		"sh", "-c",
		"ls -t /root/.local/share/opencode/storage/session_diff/ses_*.json 2>/dev/null | head -n1",
	})

	if err != nil || output == "" {
		return "", nil
	}

	// Extract session ID from filename
	// Path format: /root/.local/share/opencode/storage/session_diff/ses_<id>.json
	filename := filepath.Base(strings.TrimSpace(output))
	if !strings.HasPrefix(filename, "ses_") || !strings.HasSuffix(filename, ".json") {
		return "", nil
	}

	// Remove "ses_" prefix and ".json" suffix to get the session ID
	sessionID := strings.TrimSuffix(strings.TrimPrefix(filename, "ses_"), ".json")
	return "ses_" + sessionID, nil
}

// FormatSessionResumptionMessage returns a formatted message for session resumption
func FormatSessionResumptionMessage(sessionID string) string {
	return fmt.Sprintf("📋 Resuming session: %s", sessionID)
}
