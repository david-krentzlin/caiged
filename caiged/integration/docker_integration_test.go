//go:build integration
// +build integration

package integration

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
)

const (
	testImageName      = "caiged-test-image:latest"
	testContainerName  = "caiged-test-container"
	testDockerfilePath = "testdata/Dockerfile"
)

// TestMain sets up and tears down test resources
func TestMain(m *testing.M) {
	// Setup: Build test image
	fmt.Println("Building test image...")
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Current working directory: %s\n", cwd)

	// Build path to testdata - it should be in the same directory as this test file
	testdataPath := filepath.Join(cwd, "testdata")
	fmt.Printf("Looking for testdata at: %s\n", testdataPath)

	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		fmt.Printf("testdata directory not found at %s\n", testdataPath)
		os.Exit(1)
	}

	dockerfilePath := filepath.Join(testdataPath, "Dockerfile")

	buildCfg := docker.BuildConfig{
		Context:    testdataPath,
		Dockerfile: dockerfilePath,
		Tag:        testImageName,
	}

	if err := client.ImageBuild(buildCfg); err != nil {
		fmt.Printf("Failed to build test image: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup: Remove test containers and image
	fmt.Println("Cleaning up test resources...")

	// Remove any leftover containers
	containerNames := []string{
		testContainerName,
		testContainerName + "-detached",
		testContainerName + "-run",
		testContainerName + "-exec",
		testContainerName + "-labels",
		testContainerName + "-env",
	}

	for _, name := range containerNames {
		if client.ContainerExists(name) {
			client.ContainerRemove(name)
		}
	}

	// Remove test image
	executor.Run("docker", []string{"rmi", "-f", testImageName}, exec.RunOptions{})

	os.Exit(code)
}

func TestDockerIntegration_ImageBuild(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	// Verify test image exists
	if !client.ImageExists(testImageName) {
		t.Fatalf("Test image %s does not exist", testImageName)
	}

	// Note: With BuildKit (default in modern Docker), intermediate images
	// may not be kept in the local cache, so we don't check for alpine:3.19
}

func TestDockerIntegration_ContainerLifecycle(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-lifecycle"

	// Cleanup any existing container
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Test: Container should not exist initially
	if client.ContainerExists(containerName) {
		t.Error("Container should not exist before creation")
	}

	// Test: Create container
	runCfg := docker.RunConfig{
		Name:    containerName,
		Image:   testImageName,
		Detach:  true,
		Command: []string{"sleep", "10"},
	}

	err := client.ContainerRun(runCfg)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Give container a moment to start
	time.Sleep(500 * time.Millisecond)

	// Test: Container should exist
	if !client.ContainerExists(containerName) {
		t.Error("Container should exist after creation")
	}

	// Test: Container should be running
	if !client.ContainerIsRunning(containerName) {
		t.Error("Container should be running")
	}

	// Test: Remove container
	err = client.ContainerRemove(containerName)
	if err != nil {
		t.Fatalf("Failed to remove container: %v", err)
	}

	// Test: Container should not exist after removal
	if client.ContainerExists(containerName) {
		t.Error("Container should not exist after removal")
	}
}

func TestDockerIntegration_ContainerExec(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-exec"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container
	runCfg := docker.RunConfig{
		Name:    containerName,
		Image:   testImageName,
		Detach:  true,
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: Execute command and capture output
	output, err := client.ContainerExecCapture(containerName, []string{"cat", "/test-file.txt"})
	if err != nil {
		t.Fatalf("Failed to exec command: %v", err)
	}

	expectedOutput := "test-content"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain %q, got %q", expectedOutput, output)
	}

	// Test: Execute command that creates a file
	_, err = client.ContainerExecCapture(containerName, []string{"sh", "-c", "echo hello > /tmp/test.txt"})
	if err != nil {
		t.Fatalf("Failed to create file in container: %v", err)
	}

	// Verify file was created
	output, err = client.ContainerExecCapture(containerName, []string{"cat", "/tmp/test.txt"})
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if !strings.Contains(output, "hello") {
		t.Errorf("Expected file to contain 'hello', got %q", output)
	}
}

func TestDockerIntegration_ContainerInspect(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-inspect"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container
	runCfg := docker.RunConfig{
		Name:    containerName,
		Image:   testImageName,
		Detach:  true,
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: Inspect container state
	state, err := client.ContainerInspect(containerName, "{{.State.Running}}")
	if err != nil {
		t.Fatalf("Failed to inspect container: %v", err)
	}

	if state != "true" {
		t.Errorf("Expected container to be running, got state: %s", state)
	}

	// Test: Inspect container image
	image, err := client.ContainerInspect(containerName, "{{.Config.Image}}")
	if err != nil {
		t.Fatalf("Failed to inspect container image: %v", err)
	}

	if image != testImageName {
		t.Errorf("Expected image %q, got %q", testImageName, image)
	}
}

func TestDockerIntegration_ContainerLabels(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-labels"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container with labels
	runCfg := docker.RunConfig{
		Name:   containerName,
		Image:  testImageName,
		Detach: true,
		Labels: map[string]string{
			"test.app":     "caiged",
			"test.version": "1.0",
			"test.env":     "integration",
		},
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: Get label values
	tests := []struct {
		label string
		want  string
	}{
		{"test.app", "caiged"},
		{"test.version", "1.0"},
		{"test.env", "integration"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			value, err := client.ContainerGetLabel(containerName, tt.label)
			if err != nil {
				t.Fatalf("Failed to get label %q: %v", tt.label, err)
			}

			if value != tt.want {
				t.Errorf("Label %q = %q, want %q", tt.label, value, tt.want)
			}
		})
	}
}

func TestDockerIntegration_ContainerList(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-list"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container with unique label for filtering
	runCfg := docker.RunConfig{
		Name:   containerName,
		Image:  testImageName,
		Detach: true,
		Labels: map[string]string{
			"caiged.test": "list-test",
		},
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: List containers with filter
	containers, err := client.ContainerList("label=caiged.test=list-test", "{{.Names}}")
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	if len(containers) == 0 {
		t.Error("Expected to find at least one container")
	}

	found := false
	for _, name := range containers {
		if strings.Contains(name, containerName) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find container %q in list: %v", containerName, containers)
	}
}

func TestDockerIntegration_ContainerEnv(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-env"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container with environment variables
	runCfg := docker.RunConfig{
		Name:   containerName,
		Image:  testImageName,
		Detach: true,
		Env: []string{
			"TEST_VAR=test-value",
			"ANOTHER_VAR=another-value",
		},
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: Verify environment variables
	output, err := client.ContainerExecCapture(containerName, []string{"sh", "-c", "echo $TEST_VAR"})
	if err != nil {
		t.Fatalf("Failed to read env var: %v", err)
	}

	if !strings.Contains(output, "test-value") {
		t.Errorf("Expected TEST_VAR to be 'test-value', got %q", output)
	}

	output, err = client.ContainerExecCapture(containerName, []string{"sh", "-c", "echo $ANOTHER_VAR"})
	if err != nil {
		t.Fatalf("Failed to read env var: %v", err)
	}

	if !strings.Contains(output, "another-value") {
		t.Errorf("Expected ANOTHER_VAR to be 'another-value', got %q", output)
	}
}

func TestDockerIntegration_ContainerVolumes(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	containerName := testContainerName + "-volumes"

	// Create a temporary directory for volume mount
	tmpDir, err := os.MkdirTemp("", "caiged-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file in the temp directory
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("volume-test-content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Cleanup container
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create container with volume
	runCfg := docker.RunConfig{
		Name:   containerName,
		Image:  testImageName,
		Detach: true,
		Volumes: []string{
			fmt.Sprintf("%s:/mnt/test:ro", tmpDir),
		},
		Command: []string{"sleep", "30"},
	}

	if err := client.ContainerRun(runCfg); err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer client.ContainerRemove(containerName)

	time.Sleep(500 * time.Millisecond)

	// Test: Read file from mounted volume
	output, err := client.ContainerExecCapture(containerName, []string{"cat", "/mnt/test/test.txt"})
	if err != nil {
		t.Fatalf("Failed to read file from volume: %v", err)
	}

	if !strings.Contains(output, "volume-test-content") {
		t.Errorf("Expected file content 'volume-test-content', got %q", output)
	}
}

func TestDockerIntegration_ContainerWithOutput(t *testing.T) {
	executor := exec.NewRealExecutor()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	client := docker.NewClient(executor).WithOutput(stdout, stderr)

	containerName := testContainerName + "-output"

	// Cleanup
	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName)
	}

	// Create and immediately remove container (with --rm)
	runCfg := docker.RunConfig{
		Name:    containerName,
		Image:   testImageName,
		Remove:  true, // Auto-remove after exit
		Command: []string{"echo", "test output"},
	}

	err := client.ContainerRun(runCfg)
	if err != nil {
		t.Fatalf("Failed to run container: %v", err)
	}

	// Container should have auto-removed due to --rm flag
	time.Sleep(500 * time.Millisecond)

	if client.ContainerExists(containerName) {
		client.ContainerRemove(containerName) // Cleanup just in case
		t.Error("Container should have been auto-removed with --rm flag")
	}
}

func TestDockerIntegration_ImageExists(t *testing.T) {
	executor := exec.NewRealExecutor()
	client := docker.NewClient(executor)

	// Test: Image exists
	if !client.ImageExists(testImageName) {
		t.Errorf("Test image %s should exist", testImageName)
	}

	// Test: Image does not exist
	if client.ImageExists("nonexistent-image:impossible-tag-12345") {
		t.Error("Non-existent image should return false")
	}
}
