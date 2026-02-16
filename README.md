# Caiged Spins

Run isolated, tool-rich agent containers with explicit host mounts and purpose-built spins.

Most common use-case (QA spin on the current repo):

```bash
./scripts/caiged "$(pwd)" --task qa
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

By default, `caiged` starts the container detached and then opens a host tmux session (if available) with a shell attached to the container. Use `--detach` to skip attaching.

Host tmux sessions:
- Session name: `caiged-<spin>-<project>`
- If already inside tmux, it switches to the session

Inside the container:
- `,help` for environment info
- `,auth-tools` to authenticate gh and 1password
- gh config is mounted from the host when available (read-only by default)

Enable network:

```bash
./scripts/caiged /path/to/workdir --task qa --enable-network
```

Docker socket is enabled by default. To disable it:

```bash
./scripts/caiged /path/to/workdir --task qa --disable-docker-sock
```

Run detached (container keeps running):

```bash
./scripts/caiged "$(pwd)" --task qa --detach
```

Attach to container shell (host tmux session if available):

```bash
./scripts/caiged "$(pwd)" --task qa --attach
```

List active caiged containers and tmux sessions:

```bash
./scripts/caiged --list
```

Mount host gh config read-write:

```bash
./scripts/caiged "$(pwd)" --task qa --mount-gh-rw
```

Force rebuild:

```bash
OPENCODE_VERSION=latest ./scripts/caiged /path/to/workdir --task qa --force-build
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
