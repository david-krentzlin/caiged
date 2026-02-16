package cmd

import "github.com/spf13/cobra"

type RunOptions struct {
	Spin                string
	Project             string
	Repo                string
	DisableDockerSock   bool
	SecretEnv           []string
	SecretEnvFile       string
	NoMountOpenCodeAuth bool
	MountGHRW           bool
	NoMountGH           bool
	ForceBuild          bool
	DetachOnly          bool
	// Computed fields (not set by flags)
	MountOpenCodeAuth bool
	MountGH           bool
}

var runOpts = RunOptions{}

func addCommonFlags(cmd *cobra.Command) {
	cmd.SilenceUsage = true
}

func addRunFlags(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().StringVar(&opts.Spin, "spin", "qa", "Spin name")
	cmd.Flags().StringVar(&opts.Project, "project", "", "Project name for container naming")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "Path to caiged repo (spins/Dockerfile/entrypoint.sh)")
	cmd.Flags().BoolVar(&opts.DisableDockerSock, "disable-docker-sock", false, "Disable Docker socket mount")
	cmd.Flags().StringSliceVar(&opts.SecretEnv, "secret-env", nil, "Pass host secret env var into container (repeatable)")
	cmd.Flags().StringVar(&opts.SecretEnvFile, "secret-env-file", "", "Path to env file with secret values for container")
	cmd.Flags().BoolVar(&opts.NoMountOpenCodeAuth, "no-mount-opencode-auth", false, "Do not mount host OpenCode auth.json")
	cmd.Flags().BoolVar(&opts.MountGHRW, "mount-gh-rw", false, "Mount host gh config read-write")
	cmd.Flags().BoolVar(&opts.NoMountGH, "no-mount-gh", false, "Do not mount host gh config")
	addRebuildImagesFlag(cmd, opts)
	cmd.Flags().BoolVar(&opts.DetachOnly, "no-attach", false, "Start container without attaching")
}

func addAttachFlags(cmd *cobra.Command, opts *RunOptions) {
	cmd.Flags().StringVar(&opts.Spin, "spin", "qa", "Spin name")
	cmd.Flags().StringVar(&opts.Project, "project", "", "Project name for container naming")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "Path to caiged repo (spins/Dockerfile/entrypoint.sh)")
	cmd.Flags().BoolVar(&opts.DisableDockerSock, "disable-docker-sock", false, "Disable Docker socket mount")
	cmd.Flags().StringSliceVar(&opts.SecretEnv, "secret-env", nil, "Pass host secret env var into container (repeatable)")
	cmd.Flags().StringVar(&opts.SecretEnvFile, "secret-env-file", "", "Path to env file with secret values for container")
	cmd.Flags().BoolVar(&opts.NoMountOpenCodeAuth, "no-mount-opencode-auth", false, "Do not mount host OpenCode auth.json")
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
	// Compute mount flags based on --no- flags and defaults
	opts.MountOpenCodeAuth = !opts.NoMountOpenCodeAuth

	// If NoMountGH is set, disable both MountGH and MountGHRW
	if opts.NoMountGH {
		opts.MountGH = false
		opts.MountGHRW = false
	} else {
		// Default is to mount (read-only unless MountGHRW is set)
		opts.MountGH = true
	}

	return opts
}
