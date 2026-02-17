package opencode

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/david-krentzlin/caiged/caiged/internal/exec"
)

func TestNewClient(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	client := NewClient(mockExec)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.executor == nil {
		t.Error("client.executor is nil")
	}
	if client.stdout == nil {
		t.Error("client.stdout is nil")
	}
	if client.stderr == nil {
		t.Error("client.stderr is nil")
	}
	if client.stdin == nil {
		t.Error("client.stdin is nil")
	}
}

func TestWithOutput(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	stdin := &bytes.Buffer{}

	client := NewClient(mockExec).WithOutput(stdout, stderr, stdin)

	if client.stdout != stdout {
		t.Error("stdout not set correctly")
	}
	if client.stderr != stderr {
		t.Error("stderr not set correctly")
	}
	if client.stdin != stdin {
		t.Error("stdin not set correctly")
	}
	if client.executor != mockExec {
		t.Error("executor not preserved correctly")
	}
}

func TestAttach(t *testing.T) {
	tests := []struct {
		name    string
		config  AttachConfig
		wantErr bool
	}{
		{
			name: "minimal config",
			config: AttachConfig{
				URL:      "http://localhost:4096",
				Workdir:  "/workspace",
				Password: "test-password",
			},
			wantErr: false,
		},
		{
			name: "with session ID",
			config: AttachConfig{
				URL:       "http://localhost:4096",
				Workdir:   "/workspace",
				Password:  "test-password",
				SessionID: "ses_abc123",
			},
			wantErr: false,
		},
		{
			name: "different port",
			config: AttachConfig{
				URL:      "http://localhost:8080",
				Workdir:  "/custom/path",
				Password: "secure-pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			// Use prefix matching for opencode attach commands
			var mockErr error
			if tt.wantErr {
				mockErr = fmt.Errorf("attach failed")
			}
			mockExec.AddResponseForPrefix("opencode", "", mockErr)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			stdin := &bytes.Buffer{}

			client := NewClient(mockExec).WithOutput(stdout, stderr, stdin)
			err := client.Attach(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Attach() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the command was executed
			if mockExec.CommandCount() == 0 {
				t.Error("No commands were executed")
			}

			// Verify command structure
			cmd, ok := mockExec.GetLastCommand()
			if !ok {
				t.Fatal("Failed to get last command")
			}
			if cmd.Name != "opencode" {
				t.Errorf("command name = %q, want %q", cmd.Name, "opencode")
			}
			if len(cmd.Args) < 6 {
				t.Errorf("expected at least 6 args, got %d", len(cmd.Args))
			}
		})
	}
}

func TestAttachCommandConstruction(t *testing.T) {
	tests := []struct {
		name       string
		config     AttachConfig
		wantInArgs []string
	}{
		{
			name: "without session ID",
			config: AttachConfig{
				URL:      "http://localhost:4096",
				Workdir:  "/workspace",
				Password: "pass123",
			},
			wantInArgs: []string{"attach", "http://localhost:4096", "--dir", "/workspace", "--password", "pass123"},
		},
		{
			name: "with session ID",
			config: AttachConfig{
				URL:       "http://localhost:4096",
				Workdir:   "/workspace",
				Password:  "pass123",
				SessionID: "ses_abc",
			},
			wantInArgs: []string{"attach", "http://localhost:4096", "--dir", "/workspace", "--password", "pass123", "--session", "ses_abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			mockExec.AddResponseForPrefix("opencode", "", nil)

			client := NewClient(mockExec)
			err := client.Attach(tt.config)

			if err != nil {
				t.Errorf("Attach() error = %v", err)
			}

			// Verify command was executed
			cmd, ok := mockExec.GetLastCommand()
			if !ok {
				t.Fatal("Failed to get last command")
			}

			// Verify all expected args are present
			argsStr := strings.Join(cmd.Args, " ")
			for _, wantArg := range tt.wantInArgs {
				if !strings.Contains(argsStr, wantArg) {
					t.Errorf("expected arg %q not found in command args: %v", wantArg, cmd.Args)
				}
			}
		})
	}
}

func TestGetLastSessionFromContainer(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		mockOutput    string
		mockError     error
		wantSessionID string
		wantError     bool
	}{
		{
			name:          "valid session file",
			containerName: "test-container",
			mockOutput:    "/root/.local/share/opencode/storage/session_diff/ses_398c2e0c9ffeCJppZlLGpFvGS3.json",
			mockError:     nil,
			wantSessionID: "ses_398c2e0c9ffeCJppZlLGpFvGS3",
			wantError:     false,
		},
		{
			name:          "valid session file with whitespace",
			containerName: "test-container",
			mockOutput:    "/root/.local/share/opencode/storage/session_diff/ses_abc123xyz.json\n",
			mockError:     nil,
			wantSessionID: "ses_abc123xyz",
			wantError:     false,
		},
		{
			name:          "no session files",
			containerName: "test-container",
			mockOutput:    "",
			mockError:     nil,
			wantSessionID: "",
			wantError:     false,
		},
		{
			name:          "docker exec error",
			containerName: "test-container",
			mockOutput:    "",
			mockError:     errors.New("container not found"),
			wantSessionID: "",
			wantError:     false, // We return empty string, not error
		},
		{
			name:          "invalid filename format",
			containerName: "test-container",
			mockOutput:    "/root/.local/share/opencode/storage/session_diff/invalid.json",
			mockError:     nil,
			wantSessionID: "",
			wantError:     false,
		},
		{
			name:          "missing json extension",
			containerName: "test-container",
			mockOutput:    "/root/.local/share/opencode/storage/session_diff/ses_abc123",
			mockError:     nil,
			wantSessionID: "",
			wantError:     false,
		},
		{
			name:          "missing ses prefix",
			containerName: "test-container",
			mockOutput:    "/root/.local/share/opencode/storage/session_diff/abc123.json",
			mockError:     nil,
			wantSessionID: "",
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock docker exec function
			mockExec := func(containerName string, command []string) (string, error) {
				if containerName != tt.containerName {
					t.Errorf("containerName = %q, want %q", containerName, tt.containerName)
				}
				return tt.mockOutput, tt.mockError
			}

			sessionID, err := GetLastSessionFromContainer(mockExec, tt.containerName)

			if (err != nil) != tt.wantError {
				t.Errorf("GetLastSessionFromContainer() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if sessionID != tt.wantSessionID {
				t.Errorf("GetLastSessionFromContainer() = %q, want %q", sessionID, tt.wantSessionID)
			}
		})
	}
}

func TestFormatSessionResumptionMessage(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		want      string
	}{
		{
			name:      "valid session ID",
			sessionID: "ses_abc123",
			want:      "📋 Resuming session: ses_abc123",
		},
		{
			name:      "long session ID",
			sessionID: "ses_398c2e0c9ffeCJppZlLGpFvGS3",
			want:      "📋 Resuming session: ses_398c2e0c9ffeCJppZlLGpFvGS3",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			want:      "📋 Resuming session: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSessionResumptionMessage(tt.sessionID)
			if got != tt.want {
				t.Errorf("FormatSessionResumptionMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSession(t *testing.T) {
	// Test Session struct
	session := Session{
		ID:        "ses_abc123",
		Workdir:   "/workspace",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.Workdir == "" {
		t.Error("Session Workdir should not be empty")
	}
	if session.CreatedAt == "" {
		t.Error("Session CreatedAt should not be empty")
	}
}

func TestSessionIDParsing(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantID   string
	}{
		{
			name:     "standard format",
			filename: "ses_abc123.json",
			wantID:   "ses_abc123",
		},
		{
			name:     "with underscores in ID",
			filename: "ses_abc_123_xyz.json",
			wantID:   "ses_abc_123_xyz",
		},
		{
			name:     "with dashes in ID",
			filename: "ses_abc-123-xyz.json",
			wantID:   "ses_abc-123-xyz",
		},
		{
			name:     "alphanumeric ID",
			filename: "ses_398c2e0c9ffeCJppZlLGpFvGS3.json",
			wantID:   "ses_398c2e0c9ffeCJppZlLGpFvGS3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parsing logic
			if !strings.HasPrefix(tt.filename, "ses_") {
				t.Skip("Invalid prefix, skipping")
			}
			if !strings.HasSuffix(tt.filename, ".json") {
				t.Skip("Invalid suffix, skipping")
			}

			id := strings.TrimSuffix(strings.TrimPrefix(tt.filename, "ses_"), ".json")
			result := "ses_" + id

			if result != tt.wantID {
				t.Errorf("parsed ID = %q, want %q", result, tt.wantID)
			}
		})
	}
}

func TestMockDockerExecScenarios(t *testing.T) {
	tests := []struct {
		name          string
		execFunc      func(string, []string) (string, error)
		containerName string
		wantSessionID string
	}{
		{
			name: "successful execution",
			execFunc: func(name string, cmd []string) (string, error) {
				return "/root/.local/share/opencode/storage/session_diff/ses_test123.json", nil
			},
			containerName: "test",
			wantSessionID: "ses_test123",
		},
		{
			name: "execution returns error",
			execFunc: func(name string, cmd []string) (string, error) {
				return "", errors.New("exec failed")
			},
			containerName: "test",
			wantSessionID: "",
		},
		{
			name: "execution returns empty",
			execFunc: func(name string, cmd []string) (string, error) {
				return "", nil
			},
			containerName: "test",
			wantSessionID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := GetLastSessionFromContainer(tt.execFunc, tt.containerName)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if sessionID != tt.wantSessionID {
				t.Errorf("sessionID = %q, want %q", sessionID, tt.wantSessionID)
			}
		})
	}
}
