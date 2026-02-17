package docker

import (
	"bytes"
	"fmt"
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
}

func TestWithOutput(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	client := NewClient(mockExec).WithOutput(stdout, stderr)

	if client.stdout != stdout {
		t.Error("stdout not set correctly")
	}
	if client.stderr != stderr {
		t.Error("stderr not set correctly")
	}
	if client.executor != mockExec {
		t.Error("executor not preserved correctly")
	}
}

func TestContainerExists(t *testing.T) {
	tests := []struct {
		name      string
		container string
		mockError error
		want      bool
	}{
		{
			name:      "container exists",
			container: "my-container",
			mockError: nil,
			want:      true,
		},
		{
			name:      "container does not exist",
			container: "nonexistent",
			mockError: fmt.Errorf("container not found"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			mockExec.AddResponse("docker", []string{"inspect", tt.container}, "", tt.mockError)

			client := NewClient(mockExec)
			got := client.ContainerExists(tt.container)

			if got != tt.want {
				t.Errorf("ContainerExists() = %v, want %v", got, tt.want)
			}

			mockExec.AssertCommandExecuted(t, "docker", "inspect", tt.container)
		})
	}
}

func TestContainerIsRunning(t *testing.T) {
	tests := []struct {
		name      string
		container string
		output    string
		mockError error
		want      bool
	}{
		{
			name:      "container is running",
			container: "my-container",
			output:    "true\n",
			mockError: nil,
			want:      true,
		},
		{
			name:      "container is not running",
			container: "my-container",
			output:    "false\n",
			mockError: nil,
			want:      false,
		},
		{
			name:      "container does not exist",
			container: "nonexistent",
			output:    "",
			mockError: fmt.Errorf("not found"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			mockExec.AddResponse("docker", []string{"inspect", "-f", "{{.State.Running}}", tt.container}, tt.output, tt.mockError)

			client := NewClient(mockExec)
			got := client.ContainerIsRunning(tt.container)

			if got != tt.want {
				t.Errorf("ContainerIsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerRemove(t *testing.T) {
	tests := []struct {
		name      string
		container string
		wantErr   bool
	}{
		{
			name:      "successful removal",
			container: "my-container",
			wantErr:   false,
		},
		{
			name:      "container not found",
			container: "nonexistent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			var mockErr error
			if tt.wantErr {
				mockErr = fmt.Errorf("not found")
			}
			mockExec.AddResponse("docker", []string{"rm", "-f", tt.container}, "", mockErr)

			client := NewClient(mockExec).WithOutput(stdout, stderr)
			err := client.ContainerRemove(tt.container)

			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerRemove() error = %v, wantErr %v", err, tt.wantErr)
			}

			mockExec.AssertCommandExecuted(t, "docker", "rm", "-f", tt.container)
		})
	}
}

func TestContainerExecCapture(t *testing.T) {
	tests := []struct {
		name      string
		container string
		command   []string
		output    string
		wantErr   bool
	}{
		{
			name:      "successful exec",
			container: "my-container",
			command:   []string{"echo", "hello"},
			output:    "hello\n",
			wantErr:   false,
		},
		{
			name:      "exec with error",
			container: "my-container",
			command:   []string{"false"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()

			args := append([]string{"exec", tt.container}, tt.command...)
			var mockErr error
			if tt.wantErr {
				mockErr = fmt.Errorf("command failed")
			}
			mockExec.AddResponse("docker", args, tt.output, mockErr)

			client := NewClient(mockExec)
			output, err := client.ContainerExecCapture(tt.container, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerExecCapture() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.output {
				t.Errorf("ContainerExecCapture() output = %q, want %q", output, tt.output)
			}
		})
	}
}

func TestContainerInspect(t *testing.T) {
	tests := []struct {
		name      string
		container string
		format    string
		output    string
		wantErr   bool
	}{
		{
			name:      "get state",
			container: "my-container",
			format:    "{{.State.Running}}",
			output:    "true",
			wantErr:   false,
		},
		{
			name:      "container not found",
			container: "nonexistent",
			format:    "{{.State.Running}}",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()

			var mockErr error
			if tt.wantErr {
				mockErr = fmt.Errorf("not found")
			}
			mockExec.AddResponse("docker", []string{"inspect", "-f", tt.format, tt.container}, tt.output+"\n", mockErr)

			client := NewClient(mockExec)
			output, err := client.ContainerInspect(tt.container, tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerInspect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && output != tt.output {
				t.Errorf("ContainerInspect() = %q, want %q", output, tt.output)
			}
		})
	}
}

func TestContainerGetPort(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	format := "{{(index (index .NetworkSettings.Ports \"4096/tcp\") 0).HostPort}}"
	mockExec.AddResponse("docker", []string{"inspect", "-f", format, "my-container"}, "54321\n", nil)

	client := NewClient(mockExec)
	port, err := client.ContainerGetPort("my-container")

	if err != nil {
		t.Errorf("ContainerGetPort() error = %v", err)
	}
	if port != "54321" {
		t.Errorf("ContainerGetPort() = %q, want %q", port, "54321")
	}
}

func TestContainerGetLabel(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	format := "{{index .Config.Labels \"app\"}}"
	mockExec.AddResponse("docker", []string{"inspect", "-f", format, "my-container"}, "myapp\n", nil)

	client := NewClient(mockExec)
	label, err := client.ContainerGetLabel("my-container", "app")

	if err != nil {
		t.Errorf("ContainerGetLabel() error = %v", err)
	}
	if label != "myapp" {
		t.Errorf("ContainerGetLabel() = %q, want %q", label, "myapp")
	}
}

func TestContainerList(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		format  string
		output  string
		want    []string
		wantErr bool
	}{
		{
			name:   "list containers",
			filter: "",
			format: "{{.Names}}",
			output: "container1\ncontainer2\ncontainer3",
			want:   []string{"container1", "container2", "container3"},
		},
		{
			name:   "empty list",
			filter: "",
			format: "{{.Names}}",
			output: "",
			want:   []string{},
		},
		{
			name:   "with filter",
			filter: "name=test",
			format: "{{.Names}}",
			output: "test-container",
			want:   []string{"test-container"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()

			// Build the expected args
			args := []string{"ps"}
			if tt.filter != "" {
				args = append(args, "--filter", tt.filter)
			}
			if tt.format != "" {
				args = append(args, "--format", tt.format)
			}

			mockExec.AddResponse("docker", args, tt.output, nil)

			client := NewClient(mockExec)
			got, err := client.ContainerList(tt.filter, tt.format)

			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Errorf("ContainerList() returned %d items, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestContainerListAll(t *testing.T) {
	mockExec := exec.NewMockExecutor()
	mockExec.AddResponse("docker", []string{"ps", "-a", "--format", "{{.Names}}"}, "container1\ncontainer2", nil)

	client := NewClient(mockExec)
	containers, err := client.ContainerListAll("", "{{.Names}}")

	if err != nil {
		t.Errorf("ContainerListAll() error = %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("ContainerListAll() returned %d containers, want 2", len(containers))
	}
}

func TestImageExists(t *testing.T) {
	tests := []struct {
		name      string
		image     string
		mockError error
		want      bool
	}{
		{
			name:      "image exists",
			image:     "my-image:latest",
			mockError: nil,
			want:      true,
		},
		{
			name:      "image does not exist",
			image:     "nonexistent:latest",
			mockError: fmt.Errorf("not found"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			mockExec.AddResponse("docker", []string{"image", "inspect", tt.image}, "", tt.mockError)

			client := NewClient(mockExec)
			got := client.ImageExists(tt.image)

			if got != tt.want {
				t.Errorf("ImageExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerRun(t *testing.T) {
	tests := []struct {
		name   string
		config RunConfig
	}{
		{
			name: "minimal config",
			config: RunConfig{
				Image: "test:latest",
			},
		},
		{
			name: "full config",
			config: RunConfig{
				Image:       "test:latest",
				Name:        "my-container",
				Detach:      true,
				Remove:      true,
				Interactive: false,
				Volumes:     []string{"/host:/container"},
				Ports:       []string{"8080:80"},
				Network:     "bridge",
				Labels:      map[string]string{"app": "test"},
				Env:         []string{"FOO=bar"},
				Command:     []string{"sh", "-c", "echo hello"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			// Use prefix matching for docker run commands
			mockExec.AddResponseForPrefix("docker", "", nil)

			client := NewClient(mockExec)
			err := client.ContainerRun(tt.config)

			if err != nil {
				t.Errorf("ContainerRun() error = %v", err)
			}

			// Verify the command was executed
			if mockExec.CommandCount() == 0 {
				t.Error("No commands were executed")
			}
		})
	}
}

func TestImageBuild(t *testing.T) {
	tests := []struct {
		name   string
		config BuildConfig
	}{
		{
			name: "minimal config",
			config: BuildConfig{
				Context: ".",
				Tag:     "test:latest",
			},
		},
		{
			name: "full config",
			config: BuildConfig{
				Context:    ".",
				Dockerfile: "Dockerfile.custom",
				Tag:        "test:latest",
				BuildArgs:  map[string]string{"VERSION": "1.0"},
				Target:     "production",
				Platform:   "linux/amd64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := exec.NewMockExecutor()
			// Use prefix matching for docker build commands
			mockExec.AddResponseForPrefix("docker", "", nil)

			client := NewClient(mockExec)
			err := client.ImageBuild(tt.config)

			if err != nil {
				t.Errorf("ImageBuild() error = %v", err)
			}

			// Verify the command was executed
			if mockExec.CommandCount() == 0 {
				t.Error("No commands were executed")
			}
		})
	}
}
