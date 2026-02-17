package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/david-krentzlin/caiged/caiged/internal/docker"
	"github.com/david-krentzlin/caiged/caiged/internal/exec"
	"github.com/david-krentzlin/caiged/caiged/internal/opencode"
)

func runCommand(args []string, opts RunOptions, forceConnect bool) error {
	workdir := args[0]
	commandArgs := args[1:]

	// Warn and confirm if docker socket is enabled
	if opts.EnableDockerSock {
		fmt.Fprintf(os.Stderr, "\n%s\n", ErrorStyle.Render("⚠️  WARNING: Docker socket access enabled"))
		fmt.Fprintf(os.Stderr, "%s\n", ErrorStyle.Render("   The agent will have root-equivalent access to your host system"))
		fmt.Fprintf(os.Stderr, "%s\n", ErrorStyle.Render("   The agent can escape the container and access all files on your machine"))
		fmt.Fprintf(os.Stderr, "\n%s", InfoStyle.Render("Continue? (yes/no): "))

		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			return fmt.Errorf("operation cancelled by user")
		}
		fmt.Println()
	}

	config, err := resolveConfig(opts, workdir)
	if err != nil {
		return err
	}

	executor := exec.NewRealExecutor()
	dockerClient := docker.NewClient(executor).WithOutput(os.Stdout, os.Stderr)

	if err := ensureImages(config, dockerClient); err != nil {
		return err
	}

	if len(commandArgs) > 0 {
		return runContainerCommand(config, dockerClient, commandArgs)
	}

	// Check container state
	alreadyRunning := dockerClient.ContainerIsRunning(config.ContainerName)
	stoppedExists := !alreadyRunning && dockerClient.ContainerExists(config.ContainerName)

	if err := startContainerDetached(config, dockerClient); err != nil {
		return err
	}

	// Display connection information
	fmt.Println()
	if alreadyRunning {
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(SectionDivider.Render("  🔗 CONNECTING TO EXISTING CONTAINER"))
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	} else if stoppedExists {
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(SectionDivider.Render("  🔄 RESUMED PERSISTENT SESSION"))
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	} else {
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
		fmt.Println(SectionDivider.Render("  🚀 CONTAINER STARTED"))
		fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	}
	fmt.Println()
	fmt.Printf("  %s %s\n", LabelStyle.Render("Project:"), ProjectStyle.Render(config.Project))
	fmt.Printf("  %s %s\n", LabelStyle.Render("Container:"), ContainerStyle.Render(config.ContainerName))
	fmt.Printf("  %s %s\n", LabelStyle.Render("Server:"), ValueStyle.Render(fmt.Sprintf("http://localhost:%d", config.OpencodePort)))
	if config.ShowSessionPassword {
		fmt.Printf("  %s %s\n", LabelStyle.Render("Password:"), InfoStyle.Render(config.OpencodePassword))
	}
	fmt.Println()
	fmt.Printf("  %s\n", HeaderStyle.Render("Reconnect:"))
	fmt.Printf("    %s\n", CommandStyle.Render(fmt.Sprintf("caiged connect %s", config.Project)))
	fmt.Println()
	fmt.Printf("  %s\n", HeaderStyle.Render("Manual Connect:"))
	if config.ShowSessionPassword {
		fmt.Printf("    %s\n", CommandStyle.Render(fmt.Sprintf("opencode attach http://localhost:%d --dir /workspace --password %s", config.OpencodePort, config.OpencodePassword)))
	} else {
		fmt.Printf("    %s\n", CommandStyle.Render(fmt.Sprintf("opencode attach http://localhost:%d --dir /workspace --password <use --show-session-password>", config.OpencodePort)))
		fmt.Printf("    %s\n", InfoStyle.Render("💡 Add --show-session-password flag to display the password"))
	}
	fmt.Println(DividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if opts.NoConnect && !forceConnect {
		return nil
	}

	// By default, automatically connect to the OpenCode server
	return connectToOpenCode(config, dockerClient, executor)
}

func ensureImages(cfg Config, client *docker.Client) error {
	if cfg.ForceBuild {
		if err := buildImage(cfg, client, "base"); err != nil {
			return err
		}
		return buildImage(cfg, client, "spin")
	}

	if !client.ImageExists(cfg.BaseImage) {
		if err := buildImage(cfg, client, "base"); err != nil {
			return err
		}
	}
	if !client.ImageExists(cfg.SpinImage) {
		if err := buildImage(cfg, client, "spin"); err != nil {
			return err
		}
	}
	return nil
}

func buildImage(cfg Config, client *docker.Client, target string) error {
	imageName := cfg.BaseImage
	if target == "spin" {
		imageName = cfg.SpinImage
	}

	buildArgs := map[string]string{
		"ARCH":             cfg.Arch,
		"MISE_VERSION":     cfg.MiseVersion,
		"GH_VERSION":       cfg.GHVersion,
		"OPENCODE_VERSION": cfg.OpencodeVersion,
	}
	if target == "spin" {
		buildArgs["SPIN"] = cfg.Spin
	}

	return client.ImageBuild(docker.BuildConfig{
		Dockerfile: filepath.Join(cfg.DockerDir, "Dockerfile"),
		Context:    cfg.DockerDir,
		Target:     target,
		Tag:        imageName,
		BuildArgs:  buildArgs,
	})
}

func dockerRunArgs(cfg Config, mode dockerRunMode) []string {
	args := []string{"run"}
	if mode == dockerRunDetached {
		// Note: removed --rm to enable persistent sessions
		args = append(args, "-d", "--name", cfg.ContainerName)
		args = append(args, "--label", fmt.Sprintf("opencode.port=%d", cfg.OpencodePort))
	} else {
		args = append(args, "--rm", "-it")
	}

	args = append(args, "-v", fmt.Sprintf("%s:/workspace", cfg.WorkdirAbs))
	// Always enable networking - OpenCode needs network access for LLM APIs
	args = append(args, "--network=bridge")
	args = append(args, "-p", fmt.Sprintf("%d:4096", cfg.OpencodePort))
	// Set hostname to container name for better shell identification
	args = append(args, "--hostname", cfg.ContainerName)

	if cfg.EnableDockerSock {
		args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock")
	}
	if cfg.MountGH && cfg.MountGHPath != "" {
		mount := fmt.Sprintf("%s:/root/.config/gh", cfg.MountGHPath)
		if !cfg.MountGHRW {
			mount = mount + ":ro"
		}
		args = append(args, "-v", mount)
	}
	if cfg.MountOpenCodeAuth && cfg.OpenCodeAuthPath != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/root/.local/share/opencode/auth.json:ro", cfg.OpenCodeAuthPath))
	}
	for _, secret := range cfg.SecretEnvs {
		args = append(args, "-e", secret)
	}
	if cfg.SecretEnvFile != "" {
		args = append(args, "--env-file", cfg.SecretEnvFile)
	}

	return args
}

type dockerRunMode int

const (
	dockerRunDetached dockerRunMode = iota
	dockerRunOneShot
)

func wrapNetworkRunError(cfg Config, err error) error {
	if err == nil {
		return nil
	}
	// Network is always enabled for OpenCode to access LLM APIs
	return fmt.Errorf("docker run failed: %w", err)
}

func startContainerDetached(cfg Config, client *docker.Client) error {
	// If container is already running, nothing to do
	if client.ContainerIsRunning(cfg.ContainerName) {
		return nil
	}

	// If container exists but is stopped, restart it
	if client.ContainerExists(cfg.ContainerName) {
		fmt.Printf("%s\n", InfoStyle.Render("🔄 Resuming existing container (persistent session)..."))
		return client.ContainerStart(cfg.ContainerName)
	}

	// Container doesn't exist, create a new one
	args := dockerRunArgs(cfg, dockerRunDetached)
	args = append(args,
		"-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin),
		"-e", "AGENT_DAEMON=1",
		"-e", fmt.Sprintf("OPENCODE_SERVER_PASSWORD=%s", cfg.OpencodePassword),
		cfg.SpinImage)

	// Use ContainerRun with the args (note: we're still building args manually for now)
	// TODO: Eventually migrate to using RunConfig directly
	executor := exec.NewRealExecutor()
	return wrapNetworkRunError(cfg, executor.Run("docker", args, exec.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}))
}

func runContainerCommand(cfg Config, client *docker.Client, command []string) error {
	args := dockerRunArgs(cfg, dockerRunOneShot)
	args = append(args, "-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin), cfg.SpinImage)
	args = append(args, command...)

	executor := exec.NewRealExecutor()
	return wrapNetworkRunError(cfg, executor.Run("docker", args, exec.RunOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}))
}

func connectToOpenCode(cfg Config, dockerClient *docker.Client, executor exec.CmdExecutor) error {
	// Wait for the OpenCode server to be ready
	fmt.Printf("%s", InfoStyle.Render("⏳ Waiting for OpenCode server to start"))
	maxWait := 60 * time.Second
	checkInterval := 500 * time.Millisecond
	maxCheckInterval := 5 * time.Second
	deadline := time.Now().Add(maxWait)

	url := fmt.Sprintf("http://localhost:%d", cfg.OpencodePort)
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	serverReady := false
	for time.Now().Before(deadline) {
		// Try to make an HTTP request to see if server is responding
		resp, err := client.Get(url)

		// If we got a response object, the server is responding (even if it's an error like 401)
		if resp != nil {
			resp.Body.Close()
			serverReady = true
			fmt.Printf(" %s\n", SuccessStyle.Render("✓ ready!"))
			break
		}

		// If there's no error and no response, something is very wrong, but let's treat it as ready
		if err == nil {
			serverReady = true
			fmt.Printf(" %s\n", SuccessStyle.Render("✓ ready!"))
			break
		}

		fmt.Printf(".")
		time.Sleep(checkInterval)

		// Exponential backoff: double the interval, but cap at max
		checkInterval = checkInterval * 2
		if checkInterval > maxCheckInterval {
			checkInterval = maxCheckInterval
		}
	}

	if !serverReady {
		fmt.Printf("\n")
		return fmt.Errorf("OpenCode server failed to start within %v", maxWait)
	}

	// Now connect with the OpenCode client
	// Try to get the last session ID from the container
	lastSessionID, err := opencode.GetLastSessionFromContainer(
		func(name string, cmd []string) (string, error) {
			return dockerClient.ContainerExecCapture(name, cmd)
		},
		cfg.ContainerName,
	)
	if err == nil && lastSessionID != "" {
		fmt.Printf("%s\n", InfoStyle.Render(opencode.FormatSessionResumptionMessage(lastSessionID)))
	}

	opencodeClient := opencode.NewClient(executor).WithOutput(os.Stdout, os.Stderr, os.Stdin)
	return opencodeClient.Attach(opencode.AttachConfig{
		URL:       url,
		Workdir:   "/workspace",
		Password:  cfg.OpencodePassword,
		SessionID: lastSessionID,
	})
}
