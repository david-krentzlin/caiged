package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func createFakeRepoRoot(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, "spins"), 0o755); err != nil {
		t.Fatalf("mkdir spins: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "entrypoint.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write entrypoint.sh: %v", err)
	}

	return repoRoot
}

func TestDeriveProjectName(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "two segments", path: "/a/b", want: "a-b"},
		{name: "many segments", path: "/one/two/three", want: "two-three"},
		{name: "single segment", path: "project", want: "project"},
		{name: "root fallback", path: "/", want: "workspace"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveProjectName(tc.path); got != tc.want {
				t.Fatalf("deriveProjectName(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestSlugifyProjectName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "spaces and case", in: "My Project", want: "my-project"},
		{name: "special chars", in: "proj@123!", want: "proj-123"},
		{name: "leading punctuation", in: "--..Name", want: "name"},
		{name: "all punctuation fallback", in: "!!!", want: "workspace"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := slugifyProjectName(tc.in); got != tc.want {
				t.Fatalf("slugifyProjectName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestDockerRunArgsModes(t *testing.T) {
	cfg := Config{
		WorkdirAbs:        "/tmp/work",
		ContainerName:     "caiged-qa-demo",
		EnableNetwork:     true,
		DisableDockerSock: false,
		MountGH:           true,
		MountGHRW:         false,
		MountGHPath:       "/tmp/gh",
	}

	detached := dockerRunArgs(cfg, dockerRunDetached)
	if !slices.Contains(detached, "--name") || !slices.Contains(detached, cfg.ContainerName) {
		t.Fatalf("detached args should include container name: %v", detached)
	}
	if !slices.Contains(detached, "-d") {
		t.Fatalf("detached args should include -d: %v", detached)
	}
	if !slices.Contains(detached, "--network=host") {
		t.Fatalf("detached args should use host networking by default: %v", detached)
	}

	oneshoot := dockerRunArgs(cfg, dockerRunOneShot)
	if slices.Contains(oneshoot, "--name") {
		t.Fatalf("one-shot args should not include fixed container name: %v", oneshoot)
	}
}

func TestDockerRunArgsDisableNetwork(t *testing.T) {
	cfg := Config{WorkdirAbs: "/tmp/work", EnableNetwork: false}
	args := dockerRunArgs(cfg, dockerRunDetached)
	if !slices.Contains(args, "--network=none") {
		t.Fatalf("expected --network=none when network disabled: %v", args)
	}
	if slices.Contains(args, "--network=host") {
		t.Fatalf("did not expect --network=host when network disabled: %v", args)
	}
}

func TestValidateSpinDir(t *testing.T) {
	root := t.TempDir()
	spin := filepath.Join(root, "demo")
	if err := os.MkdirAll(spin, 0o755); err != nil {
		t.Fatalf("mkdir spin: %v", err)
	}

	if err := validateSpinDir(spin); err == nil {
		t.Fatalf("expected validation error when AGENTS.md is missing")
	}

	agents := filepath.Join(spin, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("# Agent"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := validateSpinDir(spin); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestResolveRepoRootPrecedence(t *testing.T) {
	originalDefaultRepoPath := defaultRepoPath
	originalEnv, hadEnv := os.LookupEnv("CAIGED_REPO")
	t.Cleanup(func() {
		defaultRepoPath = originalDefaultRepoPath
		if hadEnv {
			_ = os.Setenv("CAIGED_REPO", originalEnv)
		} else {
			_ = os.Unsetenv("CAIGED_REPO")
		}
	})

	overrideRepo := createFakeRepoRoot(t)
	envRepo := createFakeRepoRoot(t)
	defaultRepo := createFakeRepoRoot(t)

	defaultRepoPath = defaultRepo
	if err := os.Setenv("CAIGED_REPO", envRepo); err != nil {
		t.Fatalf("set env: %v", err)
	}

	resolved, err := resolveRepoRoot(t.TempDir(), overrideRepo)
	if err != nil {
		t.Fatalf("resolveRepoRoot with override: %v", err)
	}
	if resolved != overrideRepo {
		t.Fatalf("expected override repo %q, got %q", overrideRepo, resolved)
	}

	resolved, err = resolveRepoRoot(t.TempDir(), "")
	if err != nil {
		t.Fatalf("resolveRepoRoot with env fallback: %v", err)
	}
	if resolved != envRepo {
		t.Fatalf("expected env repo %q, got %q", envRepo, resolved)
	}

	if err := os.Unsetenv("CAIGED_REPO"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	resolved, err = resolveRepoRoot(t.TempDir(), "")
	if err != nil {
		t.Fatalf("resolveRepoRoot with default fallback: %v", err)
	}
	if resolved != defaultRepo {
		t.Fatalf("expected default repo %q, got %q", defaultRepo, resolved)
	}
}

func TestResolveConfigAcceptsNewSpinWithoutCodeChanges(t *testing.T) {
	repoRoot := createFakeRepoRoot(t)
	spinName := "engineer"
	spinDir := filepath.Join(repoRoot, "spins", spinName)
	if err := os.MkdirAll(spinDir, 0o755); err != nil {
		t.Fatalf("mkdir spin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(spinDir, "AGENTS.md"), []byte("# Engineer Agent\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	workdir := t.TempDir()
	opts := RunOptions{Spin: spinName, Repo: repoRoot}

	cfg, err := resolveConfig(opts, workdir)
	if err != nil {
		t.Fatalf("resolveConfig for new spin: %v", err)
	}
	if cfg.Spin != spinName {
		t.Fatalf("expected spin %q, got %q", spinName, cfg.Spin)
	}
	if cfg.SpinDir != spinDir {
		t.Fatalf("expected spin dir %q, got %q", spinDir, cfg.SpinDir)
	}
	if cfg.SpinImage != "caiged:"+spinName {
		t.Fatalf("expected spin image caiged:%s, got %q", spinName, cfg.SpinImage)
	}
}

func TestHostOpenCodeAuthPath(t *testing.T) {
	home := t.TempDir()
	authDir := filepath.Join(home, ".local", "share", "opencode")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatalf("mkdir auth dir: %v", err)
	}

	if got := hostOpenCodeAuthPath(home); got != "" {
		t.Fatalf("expected empty auth path when file is missing, got %q", got)
	}

	authFile := filepath.Join(authDir, "auth.json")
	if err := os.WriteFile(authFile, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write auth.json: %v", err)
	}

	if got := hostOpenCodeAuthPath(home); got != authFile {
		t.Fatalf("expected OpenCode auth path %q, got %q", authFile, got)
	}
}

func TestResolveSecretEnvs(t *testing.T) {
	original, hadOriginal := os.LookupEnv("JFROG_OIDC_USER")
	t.Cleanup(func() {
		if hadOriginal {
			_ = os.Setenv("JFROG_OIDC_USER", original)
		} else {
			_ = os.Unsetenv("JFROG_OIDC_USER")
		}
	})

	if err := os.Setenv("JFROG_OIDC_USER", "ci-user"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	values, err := resolveSecretEnvs([]string{"JFROG_OIDC_USER"})
	if err != nil {
		t.Fatalf("resolveSecretEnvs: %v", err)
	}
	if len(values) != 1 || values[0] != "JFROG_OIDC_USER=ci-user" {
		t.Fatalf("unexpected values: %v", values)
	}

	if _, err := resolveSecretEnvs([]string{"invalid-name"}); err == nil {
		t.Fatalf("expected invalid env name error")
	}
	if _, err := resolveSecretEnvs([]string{"MISSING_SECRET_ENV"}); err == nil {
		t.Fatalf("expected missing host secret env error")
	}
}

func TestDockerRunArgsIncludesSecretEnvs(t *testing.T) {
	cfg := Config{
		WorkdirAbs:    "/tmp/work",
		EnableNetwork: true,
		SecretEnvs:    []string{"JFROG_OIDC_USER=ci-user", "JFROG_OIDC_TOKEN=topsecret"},
	}

	args := dockerRunArgs(cfg, dockerRunDetached)
	if !slices.Contains(args, "JFROG_OIDC_USER=ci-user") {
		t.Fatalf("expected secret env to be present in docker args: %v", args)
	}
	if !slices.Contains(args, "JFROG_OIDC_TOKEN=topsecret") {
		t.Fatalf("expected secret env to be present in docker args: %v", args)
	}
}
