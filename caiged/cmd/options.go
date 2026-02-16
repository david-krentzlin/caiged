package cmd

import "github.com/spf13/cobra"

type RunOptions struct {
	Spin                string
	Project             string
	Repo                string
	DisableNetwork      bool
	DisableDockerSock   bool
	SecretEnv           []string
	SecretEnvFile       string
	MountOpenCodeAuth   bool
	NoMountOpenCodeAuth bool
	MountGH             bool
	MountGHRW           bool
	NoMountGH           bool
	ForceBuild          bool
	DetachOnly          bool
}

var runOpts = RunOptions{}

func addCommonFlags(cmd *cobra.Command) {
	cmd.SilenceUsage = true
}

func addRunFlags(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().StringVar(&opts.Spin, "spin", "qa", "Spin name")
	cmd.Flags().StringVar(&opts.Project, "project", "", "Project name for container naming")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "Path to caiged repo (spins/Dockerfile/entrypoint.sh)")
	cmd.Flags().BoolVar(&opts.DisableNetwork, "disable-network", false, "Disable network access")
	cmd.Flags().BoolVar(&opts.DisableDockerSock, "disable-docker-sock", false, "Disable Docker socket mount")
	cmd.Flags().StringSliceVar(&opts.SecretEnv, "secret-env", nil, "Pass host secret env var into container (repeatable)")
	cmd.Flags().StringVar(&opts.SecretEnvFile, "secret-env-file", "", "Path to env file with secret values for container")
	cmd.Flags().BoolVar(&opts.MountOpenCodeAuth, "mount-opencode-auth", true, "Mount host OpenCode auth.json when available")
	cmd.Flags().BoolVar(&opts.NoMountOpenCodeAuth, "no-mount-opencode-auth", false, "Do not mount host OpenCode auth.json")
	cmd.Flags().BoolVar(&opts.MountGH, "mount-gh", true, "Mount host gh config when available")
	cmd.Flags().BoolVar(&opts.MountGHRW, "mount-gh-rw", false, "Mount host gh config read-write")
	cmd.Flags().BoolVar(&opts.NoMountGH, "no-mount-gh", false, "Do not mount host gh config")
	addRebuildImagesFlag(cmd, opts)
	cmd.Flags().BoolVar(&opts.DetachOnly, "no-attach", false, "Start container without attaching")
}

func addAttachFlags(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().StringVar(&opts.Spin, "spin", "qa", "Spin name")
	cmd.Flags().StringVar(&opts.Project, "project", "", "Project name for container naming")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "Path to caiged repo (spins/Dockerfile/entrypoint.sh)")
	cmd.Flags().BoolVar(&opts.DisableNetwork, "disable-network", false, "Disable network access")
	cmd.Flags().BoolVar(&opts.DisableDockerSock, "disable-docker-sock", false, "Disable Docker socket mount")
	cmd.Flags().StringSliceVar(&opts.SecretEnv, "secret-env", nil, "Pass host secret env var into container (repeatable)")
	cmd.Flags().StringVar(&opts.SecretEnvFile, "secret-env-file", "", "Path to env file with secret values for container")
	cmd.Flags().BoolVar(&opts.MountOpenCodeAuth, "mount-opencode-auth", true, "Mount host OpenCode auth.json when available")
	cmd.Flags().BoolVar(&opts.NoMountOpenCodeAuth, "no-mount-opencode-auth", false, "Do not mount host OpenCode auth.json")
	cmd.Flags().BoolVar(&opts.MountGH, "mount-gh", true, "Mount host gh config when available")
	cmd.Flags().BoolVar(&opts.MountGHRW, "mount-gh-rw", false, "Mount host gh config read-write")
	cmd.Flags().BoolVar(&opts.NoMountGH, "no-mount-gh", false, "Do not mount host gh config")
	addRebuildImagesFlag(cmd, opts)
}

func addBuildFlags(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().StringVar(&opts.Spin, "spin", "qa", "Spin name")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "Path to caiged repo (spins/Dockerfile/entrypoint.sh)")
}

func addRebuildImagesFlag(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().BoolVar(&opts.ForceBuild, "rebuild-images", false, "Rebuild base and spin images")
	cmd.Flags().BoolVar(&opts.ForceBuild, "force-build", false, "Deprecated: use --rebuild-images")
	_ = cmd.Flags().MarkHidden("force-build")
	_ = cmd.Flags().MarkDeprecated("force-build", "use --rebuild-images")
}

func normalizeOptions(opts RunOptions) RunOptions {
	if opts.NoMountOpenCodeAuth {
		opts.MountOpenCodeAuth = false
	}
	if opts.NoMountGH {
		opts.MountGH = false
		opts.MountGHRW = false
	}
	if opts.MountGHRW {
		opts.MountGH = true
	}
	if !opts.MountGH {
		opts.MountGHRW = false
	}
	return opts
}
