# Caiged Spins

Run isolated, tool-rich agent containers with explicit host mounts and purpose-built spins.

Most common use-case (QA spin on a repo):

```bash
./scripts/caiged /path/to/workdir --task qa
```

Container naming:
- Name format: `caiged-<spin>-<project>`
- Default project name uses the last two path segments of the working dir
- Override with `--project <name>`

## Spins

Spins live under `spins/` and include their own `AGENT.md`, skills, and MCP config.

Current spins:
- `spins/qa`

## Tooling and versions

Tools are installed via `mise` with pinned versions in `config/target_mise.toml`. `opencode-ai` is installed via `bun add -g` during build (default `OPENCODE_VERSION=latest`).

## Build

Build is handled by the `caiged` wrapper, which only rebuilds if images are missing or `--force-build` is set.

## Run examples

QA spin with network disabled (default):

```bash
./scripts/caiged /path/to/workdir --task qa
```

On start, the container opens a tmux session with:
- README tab showing the spin overview
- opencode tab attempting to run `opencode`

Enable network:

```bash
./scripts/caiged /path/to/workdir --task qa --enable-network
```

Docker socket is enabled by default. To disable it:

```bash
./scripts/caiged /path/to/workdir --task qa --disable-docker-sock
```

Force rebuild:

```bash
OPENCODE_VERSION=latest ./scripts/caiged /path/to/workdir --task qa --force-build
```

Optional tmux:

```bash
./scripts/caiged /path/to/workdir --task qa --tmux
```

## Acceptance test

```bash
./scripts/acceptance.sh
```

Or via make:

```bash
make acceptance
```

## Credentials

The image includes `op` (1Password CLI). You can authenticate inside the container:

```bash
op signin
```

If you prefer explicit auth flows for tools (like `gh`), run their standard login commands inside the container.

Quick onboarding script:

```bash
caiged-onboard
```

If opencode needs a specific auth flow, set `OPENCODE_AUTH_CMD` before running `caiged-onboard`.
