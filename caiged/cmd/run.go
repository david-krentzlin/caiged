package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <workdir> [-- <command>]",
		Short: "Run a spin and attach via host tmux",
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

	if opts.DetachOnly && !isAttach {
		return nil
	}

	return attachShell(config)
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
		"--build-arg", fmt.Sprintf("OP_VERSION=%s", cfg.OPVersion),
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
	} else {
		args = append(args, "--rm", "-it")
	}

	args = append(args, "-v", fmt.Sprintf("%s:/workspace", cfg.WorkdirAbs))
	if !cfg.EnableNetwork {
		args = append(args, "--network=none")
	} else {
		args = append(args, "--network=host")
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
	args = append(args, "-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin), "-e", "AGENT_DAEMON=1", cfg.SpinImage)

	return wrapNetworkRunError(cfg, execCommand("docker", args, ExecOptions{Stdout: os.Stdout, Stderr: os.Stderr}))
}

func runContainerCommand(cfg Config, command []string) error {
	args := dockerRunArgs(cfg, dockerRunOneShot)
	args = append(args, "-e", fmt.Sprintf("AGENT_SPIN=%s", cfg.Spin), cfg.SpinImage)
	args = append(args, command...)

	return wrapNetworkRunError(cfg, execCommand("docker", args, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}))
}

type tmuxWindowSet struct {
	shell    string
	help     string
	opencode string
}

func tmuxWindowCommands(cfg Config) tmuxWindowSet {
	return tmuxWindowSet{
		shell:    fmt.Sprintf("docker exec -it -e CAIGED_WINDOW=shell %s %s", cfg.ContainerName, cfg.ContainerShell),
		help:     fmt.Sprintf("docker exec -it -e CAIGED_WINDOW=help %s /bin/zsh -lc '/usr/local/bin/,help; exec %s'", cfg.ContainerName, cfg.ContainerShell),
		opencode: fmt.Sprintf("docker exec -it -e CAIGED_WINDOW=opencode %s /bin/zsh -lc '/usr/local/bin/start-opencode; exec %s'", cfg.ContainerName, cfg.ContainerShell),
	}
}

func ensureTmuxSession(cfg Config) bool {
	if !commandExists("tmux") {
		return false
	}
	if execCommand("tmux", []string{"has-session", "-t", cfg.SessionName}, ExecOptions{}) == nil {
		_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "automatic-rename", "off"}, ExecOptions{})
		_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "allow-rename", "off"}, ExecOptions{})
		ensureTmuxWindows(cfg)
		return true
	}

	commands := tmuxWindowCommands(cfg)

	if err := execCommand("tmux", []string{"new-session", "-d", "-s", cfg.SessionName, "-n", "help", commands.help}, ExecOptions{}); err != nil {
		return false
	}
	_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "automatic-rename", "off"}, ExecOptions{})
	_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "allow-rename", "off"}, ExecOptions{})
	_ = execCommand("tmux", []string{"new-window", "-t", cfg.SessionName, "-n", "opencode", commands.opencode}, ExecOptions{})
	_ = execCommand("tmux", []string{"new-window", "-t", cfg.SessionName, "-n", "shell", commands.shell}, ExecOptions{})
	orderTmuxWindows(cfg)
	return true
}

func ensureTmuxWindows(cfg Config) {
	_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "automatic-rename", "off"}, ExecOptions{})
	_ = execCommand("tmux", []string{"set-option", "-t", cfg.SessionName, "allow-rename", "off"}, ExecOptions{})

	output, err := runCapture("tmux", []string{"list-windows", "-t", cfg.SessionName, "-F", "#{window_name}"}, ExecOptions{})
	if err != nil {
		return
	}

	windowNames := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		windowNames[line] = true
	}

	commands := tmuxWindowCommands(cfg)

	if !windowNames["help"] {
		_ = execCommand("tmux", []string{"new-window", "-t", cfg.SessionName, "-n", "help", commands.help}, ExecOptions{})
	}
	if !windowNames["opencode"] {
		_ = execCommand("tmux", []string{"new-window", "-t", cfg.SessionName, "-n", "opencode", commands.opencode}, ExecOptions{})
	}
	if !windowNames["shell"] {
		_ = execCommand("tmux", []string{"new-window", "-t", cfg.SessionName, "-n", "shell", commands.shell}, ExecOptions{})
	}

	orderTmuxWindows(cfg)
}

func orderTmuxWindows(cfg Config) {
	baseIndex := tmuxBaseIndex()
	_ = execCommand("tmux", []string{"move-window", "-s", cfg.SessionName + ":help", "-t", fmt.Sprintf("%s:%d", cfg.SessionName, baseIndex)}, ExecOptions{})
	_ = execCommand("tmux", []string{"move-window", "-s", cfg.SessionName + ":opencode", "-t", fmt.Sprintf("%s:%d", cfg.SessionName, baseIndex+1)}, ExecOptions{})
	_ = execCommand("tmux", []string{"move-window", "-s", cfg.SessionName + ":shell", "-t", fmt.Sprintf("%s:%d", cfg.SessionName, baseIndex+2)}, ExecOptions{})
	renameWindowIndices(cfg, baseIndex)
	_ = execCommand("tmux", []string{"select-window", "-t", fmt.Sprintf("%s:%d", cfg.SessionName, baseIndex)}, ExecOptions{})
}

func renameWindowIndices(cfg Config, baseIndex int) {
	renameWindow(cfg, baseIndex, "help")
	renameWindow(cfg, baseIndex+1, "opencode")
	renameWindow(cfg, baseIndex+2, "shell")
}

func renameWindow(cfg Config, index int, name string) {
	target := fmt.Sprintf("%s:%d", cfg.SessionName, index)
	_ = execCommand("tmux", []string{"rename-window", "-t", target, name}, ExecOptions{})
	_ = execCommand("tmux", []string{"set-window-option", "-t", target, "automatic-rename", "off"}, ExecOptions{})
	_ = execCommand("tmux", []string{"set-window-option", "-t", target, "allow-rename", "off"}, ExecOptions{})
}

func tmuxBaseIndex() int {
	output, err := runCapture("tmux", []string{"show-options", "-gqv", "base-index"}, ExecOptions{})
	if err != nil {
		return 0
	}
	value := strings.TrimSpace(output)
	if value == "" {
		return 0
	}
	index, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return index
}

func attachShell(cfg Config) error {
	if ensureTmuxSession(cfg) {
		if os.Getenv("TMUX") != "" {
			return execCommand("tmux", []string{"switch-client", "-t", cfg.SessionName}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
		}
		return execCommand("tmux", []string{"attach", "-t", cfg.SessionName}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
	}

	return execCommand("docker", []string{"exec", "-it", cfg.ContainerName, cfg.ContainerShell}, ExecOptions{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr})
}
