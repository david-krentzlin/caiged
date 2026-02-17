package docker

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/david-krentzlin/caiged/caiged/internal/exec"
)

// Client wraps Docker CLI operations with domain-specific methods
type Client struct {
	executor exec.CmdExecutor
	stdout   io.Writer
	stderr   io.Writer
}

// NewClient creates a new Docker client with the given executor
func NewClient(executor exec.CmdExecutor) *Client {
	return &Client{
		executor: executor,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

// WithOutput sets custom stdout/stderr for the client
func (c *Client) WithOutput(stdout, stderr io.Writer) *Client {
	return &Client{
		executor: c.executor,
		stdout:   stdout,
		stderr:   stderr,
	}
}

// Container represents a Docker container
type Container struct {
	Name   string
	ID     string
	Status string
	Labels map[string]string
}

// Image represents a Docker image
type Image struct {
	Name string
	ID   string
}

// ContainerExists checks if a container exists (running or stopped)
func (c *Client) ContainerExists(name string) bool {
	err := c.executor.Run("docker", []string{"inspect", name}, exec.RunOptions{})
	return err == nil
}

// ContainerIsRunning checks if a container is currently running
func (c *Client) ContainerIsRunning(name string) bool {
	output, err := c.executor.Output("docker", []string{"inspect", "-f", "{{.State.Running}}", name})
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// ContainerRemove forcefully removes a container
func (c *Client) ContainerRemove(name string) error {
	return c.executor.Run("docker", []string{"rm", "-f", name}, exec.RunOptions{
		Stdout: c.stdout,
		Stderr: c.stderr,
	})
}

// ContainerExec executes a command in a running container interactively
func (c *Client) ContainerExec(name string, command []string, interactive bool) error {
	args := []string{"exec"}
	if interactive {
		args = append(args, "-it")
	}
	args = append(args, name)
	args = append(args, command...)

	return c.executor.Run("docker", args, exec.RunOptions{
		Stdin:  os.Stdin,
		Stdout: c.stdout,
		Stderr: c.stderr,
	})
}

// ContainerExecCapture executes a command and captures its output
func (c *Client) ContainerExecCapture(name string, command []string) (string, error) {
	args := append([]string{"exec", name}, command...)
	output, err := c.executor.Output("docker", args)
	return string(output), err
}

// ContainerInspect inspects a container with a given format template
func (c *Client) ContainerInspect(name, format string) (string, error) {
	output, err := c.executor.Output("docker", []string{"inspect", "-f", format, name})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ContainerGetPort gets the host port mapped to container port 4096
func (c *Client) ContainerGetPort(name string) (string, error) {
	return c.ContainerInspect(name, "{{(index (index .NetworkSettings.Ports \"4096/tcp\") 0).HostPort}}")
}

// ContainerGetLabel gets a specific label value from a container
func (c *Client) ContainerGetLabel(name, label string) (string, error) {
	format := fmt.Sprintf("{{index .Config.Labels \"%s\"}}", label)
	return c.ContainerInspect(name, format)
}

// ContainerList lists containers matching a filter
func (c *Client) ContainerList(filter, format string) ([]string, error) {
	args := []string{"ps"}
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	if format != "" {
		args = append(args, "--format", format)
	}

	output, err := c.executor.Output("docker", args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

// ContainerListAll lists all containers (including stopped ones)
func (c *Client) ContainerListAll(filter, format string) ([]string, error) {
	args := []string{"ps", "-a"}
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	if format != "" {
		args = append(args, "--format", format)
	}

	output, err := c.executor.Output("docker", args)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

// RunConfig holds configuration for docker run
type RunConfig struct {
	Name        string
	Image       string
	Detach      bool
	Remove      bool
	Interactive bool
	Volumes     []string
	Ports       []string
	Network     string
	Labels      map[string]string
	Env         []string
	EnvFile     string
	Command     []string
}

// ContainerRun starts a new container
func (c *Client) ContainerRun(cfg RunConfig) error {
	args := []string{"run"}

	if cfg.Detach {
		args = append(args, "-d")
	}
	if cfg.Remove {
		args = append(args, "--rm")
	}
	if cfg.Interactive {
		args = append(args, "-it")
	}
	if cfg.Name != "" {
		args = append(args, "--name", cfg.Name)
	}

	for _, vol := range cfg.Volumes {
		args = append(args, "-v", vol)
	}
	for _, port := range cfg.Ports {
		args = append(args, "-p", port)
	}
	if cfg.Network != "" {
		args = append(args, "--network", cfg.Network)
	}
	for key, value := range cfg.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}
	for _, env := range cfg.Env {
		args = append(args, "-e", env)
	}
	if cfg.EnvFile != "" {
		args = append(args, "--env-file", cfg.EnvFile)
	}

	args = append(args, cfg.Image)
	args = append(args, cfg.Command...)

	opts := exec.RunOptions{
		Stdout: c.stdout,
		Stderr: c.stderr,
	}
	if cfg.Interactive {
		opts.Stdin = os.Stdin
	}

	return c.executor.Run("docker", args, opts)
}

// BuildConfig holds configuration for docker build
type BuildConfig struct {
	Context    string
	Dockerfile string
	Tag        string
	BuildArgs  map[string]string
	Target     string
	Platform   string
}

// ImageBuild builds a Docker image
func (c *Client) ImageBuild(cfg BuildConfig) error {
	args := []string{"build"}

	if cfg.Dockerfile != "" {
		args = append(args, "-f", cfg.Dockerfile)
	}
	if cfg.Tag != "" {
		args = append(args, "-t", cfg.Tag)
	}
	for key, value := range cfg.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}
	if cfg.Target != "" {
		args = append(args, "--target", cfg.Target)
	}
	if cfg.Platform != "" {
		args = append(args, "--platform", cfg.Platform)
	}

	args = append(args, cfg.Context)

	return c.executor.Run("docker", args, exec.RunOptions{
		Stdout: c.stdout,
		Stderr: c.stderr,
	})
}

// ImageExists checks if an image exists
func (c *Client) ImageExists(name string) bool {
	err := c.executor.Run("docker", []string{"image", "inspect", name}, exec.RunOptions{})
	return err == nil
}
