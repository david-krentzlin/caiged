package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <workdir> [-- <command>]",
		Short: "Run a spin and connect to OpenCode server",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(args, runOpts, false)
		},
	}
	addRunFlags(cmd, &runOpts)
	return cmd
}

func runCommand(args []string, opts RunOptions, isAttach bool) error {
	workdir := args[0]
	commandArgs := args[1:]

	config, err := resolveConfig(opts, workdir)
	if err != nil {
		return err
	}

	if err := ensureImages(config); err != nil {
		return err
	}

	if len(commandArgs) > 0 {
		return runContainerCommand(config, commandArgs)
	}

	if err := startContainerDetached(config); err != nil {
		return err
	}

	// Display connection information
	fmt.Println()
	fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(SectionDivider.Render("  🚀 CONTAINER STARTED"))
	fmt.Println(SectionDivider.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Printf("  %s %s\n", LabelStyle.Render("Project:"), ProjectStyle.Render(config.Project))
	fmt.Printf("  %s %s\n", LabelStyle.Render("Container:"), ContainerStyle.Render(config.ContainerName))
	fmt.Printf("  %s %s\n", LabelStyle.Render("Server:"), ValueStyle.Render(fmt.Sprintf("http://localhost:%d", config.OpencodePort)))
	fmt.Printf("  %s %s\n", LabelStyle.Render("Password:"), InfoStyle.Render(config.OpencodePassword))
	fmt.Println()
	fmt.Printf("  %s\n", HeaderStyle.Render("Quick Connect:"))
	fmt.Printf("    %s\n", CommandStyle.Render(fmt.Sprintf("caiged connect %s", config.Project)))
	fmt.Println()
	fmt.Printf("  %s\n", HeaderStyle.Render("Manual Attach:"))
	fmt.Printf("    %s\n", CommandStyle.Render(fmt.Sprintf("opencode attach http://localhost:%d --dir /workspace --password %s", config.OpencodePort, config.OpencodePassword)))
	fmt.Println(DividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()

	if opts.DetachOnly && !isAttach {
		return nil
	}

	// By default, automatically connect to the OpenCode server
	return connectToOpenCode(config)
}

func ensureImages(cfg Config) error {
	if cfg.ForceBuild {
		if err := buildImage(cfg, "base"); err != nil {
			return err
		}
		return buildImage(cfg, "spin")
	}

	if !imageExists(cfg.BaseImage) {
		if err := buildImage(cfg, "base"); err != nil {
			return err
		}
	}
	if !imageExists(cfg.SpinImage) {
		if err := buildImage(cfg, "spin"); err != nil {
			return err
		}
	}
	return nil
}

func imageExists(image string) bool {
	err := execCommand("docker", []string{"image", "inspect", image}, ExecOptions{})
	return err == nil
}

func buildImage(cfg Config, target string) error {
	args := []string{"build", "--target", target, "-t"}
	if target == "base" {
		args = append(args, cfg.BaseImage)
	} else {
		args = append(args, cfg.SpinImage)
	}
	args = append(args,
		"--build-arg", fmt.Sprintf("ARCH=%s", cfg.Arch),
		"--build-arg", fmt.Sprintf("MISE_VERSION=%s", cfg.MiseVersion),
		"--build-arg", fmt.Sprintf("GH_VERSION=%s", cfg.GHVersion),
		"--build-arg", fmt.Sprintf("OPENCODE_VERSION=%s", cfg.OpencodeVersion),
	)
	if target == "spin" {
		args = append(args, "--build-arg", fmt.Sprintf("SPIN=%s", cfg.Spin))
	}
	args = append(args, cfg.RepoRoot)

	return execCommand("docker", args, ExecOptions{Stdout: os.Stdout, Stderr: os.Stderr})
}

func containerExists(cfg Config) bool {
	return execCommand("docker", []string{"inspect", cfg.ContainerName}, ExecOptions{}) == nil
}

func containerRunning(cfg Config) bool {
	output, err := runCapture("docker", []string{"inspect", "-f", "{{.State.Running}}", cfg.ContainerName}, ExecOptions{})
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == "true"
}

type dockerRunMode int

const (
	dockerRunDetached dockerRunMode = iota
	dockerRunOneShot
)

func dockerRunArgs(cfg Config, mode dockerRunMode) []string {
	args := []string{"run"}
	if mode == dockerRunDetached {
		args = append(args, "-d", "--rm", "--name", cfg.ContainerName)
		args = append(args, "--label", fmt.Sprintf("opencode.port=%d", cfg.OpencodePort))
	} else {
		args = append(args, "--rm", "-it")
	}

	args = append(args, "-v", fmt.Sprintf("%s:/workspace", cfg.WorkdirAbs))
	if !cfg.EnableNetwork {
		args = append(args, "--network=none")
	} else {
		args = append(args, "--network=bridge")
		args = append(args, "-p", fmt.Sprintf("%d:4096", cfg.OpencodePort))
	}
	if !cfg.DisableDockerSock {
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

func wrapNetworkRunError(cfg Config, err error) error {
	if err == nil {
		return nil
	}
	if cfg.EnableNetwork {
		return fmt.Errorf("docker run failed with host networking: %w (host networking is required for OAuth callbacks; if unsupported in Docker Desktop, enable host networking or use --disable-network)", err)
	}
	return err
}

func startContainerDetached(cfg Config) error {
	if containerRunning(cfg) {
		return nil
	}
	if containerExists(cfg) {
		_ = execCommand("docker", []string{"rm", "-f", cfg.ContainerName}, ExecOptions{})
	}

	args := dockerRunArgs(cfg, dockerRunDetached)
	args = append(args,
		"-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin),
		"-e", "AGENT_DAEMON=1",
		"-e", fmt.Sprintf("OPENCODE_SERVER_PASSWORD=%s", cfg.OpencodePassword),
		cfg.SpinImage)

	return wrapNetworkRunError(cfg, execCommand("docker", args, ExecOptions{Stdout: os.Stdout, Stderr: os.Stderr}))
}

func runContainerCommand(cfg Config, command []string) error {
	args := dockerRunArgs(cfg, dockerRunOneShot)
	args = append(args, "-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin), cfg.SpinImage)
	args = append(args, command...)

	return wrapNetworkRunError(cfg, execCommand("docker", args, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}))
}

func connectToOpenCode(cfg Config) error {
	// Wait for the OpenCode server to be ready
	fmt.Printf("%s", InfoStyle.Render("⏳ Waiting for OpenCode server to start"))
	maxWait := 60 * time.Second
	checkInterval := 500 * time.Millisecond
	deadline := time.Now().Add(maxWait)

	url := fmt.Sprintf("http://localhost:%d", cfg.OpencodePort)
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	serverReady := false
	attempts := 0
	for time.Now().Before(deadline) {
		attempts++
		// Try to make an HTTP request to see if server is responding
		resp, err := client.Get(url)

		// Debug: print what we got every 10 attempts
		if attempts%10 == 0 {
			fmt.Printf("\n[debug attempt %d] resp=%v err=%v\n", attempts, resp != nil, err)
			fmt.Printf("%s", InfoStyle.Render("⏳ Waiting for OpenCode server to start"))
		}

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
	}

	if !serverReady {
		fmt.Printf("\n")
		return fmt.Errorf("OpenCode server failed to start within %v (attempted %d times)", maxWait, attempts)
	}

	// Now connect with the OpenCode client
	connectArgs := []string{"attach", url, "--dir", "/workspace", "--password", cfg.OpencodePassword}
	return execCommand("opencode", connectArgs, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}
