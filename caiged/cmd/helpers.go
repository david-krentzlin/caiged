package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var envVarNamePattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

type Config struct {
	WorkdirAbs        string
	RepoRoot          string
	Spin              string
	SpinDir           string
	Project           string
	ProjectSlug       string
	ImagePrefix       string
	BaseImage         string
	SpinImage         string
	ContainerName     string
	SessionName       string
	ContainerShell    string
	EnableNetwork     bool
	DisableDockerSock bool
	MountGH           bool
	MountGHRW         bool
	MountGHPath       string
	MountOpenCodeAuth bool
	OpenCodeAuthPath  string
	SecretEnvs        []string
	SecretEnvFile     string
	ForceBuild        bool
	Arch              string
	MiseVersion       string
	GHVersion         string
	OpencodeVersion   string
}

type ExecOptions struct {
	Dir    string
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

func resolveConfig(opts RunOptions, workdir string) (Config, error) {
	opts = normalizeOptions(opts)
	workdirAbs, err := filepath.Abs(workdir)
	if err != nil {
		return Config{}, err
	}

	repoRoot, err := resolveRepoRoot(workdirAbs, opts.Repo)
	if err != nil {
		return Config{}, err
	}

	spin := opts.Spin
	spinDir := filepath.Join(repoRoot, "spins", spin)
	if _, err := os.Stat(spinDir); err != nil {
		return Config{}, fmt.Errorf("unknown spin: %s (missing %s)", spin, spinDir)
	}
	if err := validateSpinDir(spinDir); err != nil {
		return Config{}, err
	}

	project := opts.Project
	if project == "" {
		project = deriveProjectName(workdirAbs)
	}
	projectSlug := slugifyProjectName(project)

	imagePrefix := envOrDefault("IMAGE_PREFIX", "caiged")
	containerName := fmt.Sprintf("%s-%s-%s", imagePrefix, spin, projectSlug)

	containerShell := envOrDefault("CONTAINER_SHELL", "/bin/zsh")

	mountGHPath := ""
	if opts.MountGH {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			candidate := filepath.Join(homeDir, ".config", "gh")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				mountGHPath = candidate
			}
		}
	}

	opencodeAuthPath := ""
	if opts.MountOpenCodeAuth {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			opencodeAuthPath = hostOpenCodeAuthPath(homeDir)
		}
	}

	secretEnvs, err := resolveSecretEnvs(opts.SecretEnv)
	if err != nil {
		return Config{}, err
	}

	secretEnvFile := ""
	if opts.SecretEnvFile != "" {
		candidate, err := filepath.Abs(opts.SecretEnvFile)
		if err != nil {
			return Config{}, err
		}
		info, err := os.Stat(candidate)
		if err != nil {
			return Config{}, fmt.Errorf("invalid secret env file: %s", candidate)
		}
		if info.IsDir() {
			return Config{}, fmt.Errorf("invalid secret env file: %s is a directory", candidate)
		}
		secretEnvFile = candidate
	}

	config := Config{
		WorkdirAbs:        workdirAbs,
		RepoRoot:          repoRoot,
		Spin:              spin,
		SpinDir:           spinDir,
		Project:           project,
		ProjectSlug:       projectSlug,
		ImagePrefix:       imagePrefix,
		BaseImage:         fmt.Sprintf("%s:base", imagePrefix),
		SpinImage:         fmt.Sprintf("%s:%s", imagePrefix, spin),
		ContainerName:     containerName,
		SessionName:       containerName,
		ContainerShell:    containerShell,
		EnableNetwork:     !opts.DisableNetwork,
		DisableDockerSock: opts.DisableDockerSock,
		MountGH:           opts.MountGH,
		MountGHRW:         opts.MountGHRW,
		MountGHPath:       mountGHPath,
		MountOpenCodeAuth: opts.MountOpenCodeAuth,
		OpenCodeAuthPath:  opencodeAuthPath,
		SecretEnvs:        secretEnvs,
		SecretEnvFile:     secretEnvFile,
		ForceBuild:        opts.ForceBuild,
		Arch:              envOrDefault("ARCH", "arm64"),
		MiseVersion:       envOrDefault("MISE_VERSION", "2026.2.13"),
		GHVersion:         envOrDefault("GH_VERSION", "2.86.0"),
		OpencodeVersion:   envOrDefault("OPENCODE_VERSION", "latest"),
	}

	return config, nil
}

func resolveRepoRoot(workdir string, override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		if !isCaigedRoot(abs) {
			return "", fmt.Errorf("invalid repo path: %s", abs)
		}
		return abs, nil
	}
	if envOverride := os.Getenv("CAIGED_REPO"); envOverride != "" {
		abs, err := filepath.Abs(envOverride)
		if err != nil {
			return "", err
		}
		if !isCaigedRoot(abs) {
			return "", fmt.Errorf("invalid repo path: %s", abs)
		}
		return abs, nil
	}

	if defaultRepoPath != "" {
		if isCaigedRoot(defaultRepoPath) {
			return defaultRepoPath, nil
		}
		return "", fmt.Errorf("compiled repo path not found: %s (did you move or remove the repository?) use --repo or set CAIGED_REPO", defaultRepoPath)
	}

	if repoRoot, ok := findRepoRoot(workdir); ok {
		return repoRoot, nil
	}

	if cwd, err := os.Getwd(); err == nil {
		if repoRoot, ok := findRepoRoot(cwd); ok {
			return repoRoot, nil
		}
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if repoRoot, ok := findRepoRoot(exeDir); ok {
			return repoRoot, nil
		}
	}

	return "", fmt.Errorf("unable to locate caiged repo; set --repo or CAIGED_REPO")
}

func findRepoRoot(start string) (string, bool) {
	current := start
	for {
		if isCaigedRoot(current) {
			return current, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", false
}

func isCaigedRoot(path string) bool {
	spins := filepath.Join(path, "spins")
	if info, err := os.Stat(spins); err != nil || !info.IsDir() {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "Dockerfile")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "entrypoint.sh")); err != nil {
		return false
	}
	return true
}

func validateSpinDir(spinDir string) error {
	agentsPath := filepath.Join(spinDir, "AGENTS.md")
	legacyAgentPath := filepath.Join(spinDir, "AGENT.md")
	if _, err := os.Stat(agentsPath); err != nil {
		if _, legacyErr := os.Stat(legacyAgentPath); legacyErr != nil {
			return fmt.Errorf("invalid spin: missing AGENTS.md (or legacy AGENT.md) in %s", spinDir)
		}
	}

	skillsPath := filepath.Join(spinDir, "skills")
	if info, err := os.Stat(skillsPath); err == nil && !info.IsDir() {
		return fmt.Errorf("invalid spin: skills is not a directory in %s", spinDir)
	}

	mcpPath := filepath.Join(spinDir, "mcp")
	if info, err := os.Stat(mcpPath); err == nil && !info.IsDir() {
		return fmt.Errorf("invalid spin: mcp is not a directory in %s", spinDir)
	}

	return nil
}

func hostOpenCodeAuthPath(homeDir string) string {
	candidate := filepath.Join(homeDir, ".local", "share", "opencode", "auth.json")
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}

func resolveSecretEnvs(names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	values := make([]string, 0, len(names))
	for _, name := range names {
		clean := strings.TrimSpace(name)
		if clean == "" {
			continue
		}
		if !envVarNamePattern.MatchString(clean) {
			return nil, fmt.Errorf("invalid secret env name: %s", clean)
		}
		value, ok := os.LookupEnv(clean)
		if !ok {
			return nil, fmt.Errorf("missing host secret env: %s", clean)
		}
		values = append(values, clean+"="+value)
	}

	return values, nil
}

func deriveProjectName(path string) string {
	clean := filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(clean, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	count := len(filtered)
	if count >= 2 {
		return fmt.Sprintf("%s-%s", filtered[count-2], filtered[count-1])
	}
	if count == 1 {
		return filtered[0]
	}
	return "workspace"
}

func slugifyProjectName(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")

	var builder strings.Builder
	for _, char := range slug {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '-' || char == '_' || char == '.':
			builder.WriteRune(char)
		default:
			builder.WriteRune('-')
		}
	}

	clean := builder.String()
	clean = strings.TrimLeftFunc(clean, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	clean = strings.TrimRightFunc(clean, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	if clean == "" {
		return "workspace"
	}
	return clean
}

func envOrDefault(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func runCapture(name string, args []string, opts ExecOptions) (string, error) {
	cmd := exec.Command(name, args...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w (%s)", name, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func execCommand(name string, args []string, opts ExecOptions) error {
	cmd := exec.Command(name, args...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	}
	return cmd.Run()
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func stopContainer(cfg Config) {
	_ = execCommand("docker", []string{"rm", "-f", cfg.ContainerName}, ExecOptions{})
}

func resetSession(cfg Config) {
	if !commandExists("tmux") {
		return
	}
	_ = execCommand("tmux", []string{"kill-session", "-t", cfg.SessionName}, ExecOptions{})
}
