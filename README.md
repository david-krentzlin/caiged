# Caiged Spins

Run your coding agent in a dockerized environment tailored to a specific role.

```bash
caiged "$(pwd)" --spin qa
```

This will spin up a docker container preconfigured for the QA agent.
Then it will create or attach to a tmux session that connects into that container.

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

Build is handled by the `caiged` CLI, which only rebuilds if images are missing or `--force-build` is set.

Build images explicitly:

```bash
caiged build "$(pwd)" --spin qa
```

When installed via `make install`, the CLI is compiled with the repo path it was built from. If the repo is moved or deleted, `caiged` will error and ask you to set `--repo` or `CAIGED_REPO`.

Primary subcommands:
- `run`: start a spin (builds images if needed) and attach via host tmux when available
- `build`: build the base and spin images without running a container
- `session`: manage host tmux sessions and related containers (attach/list/restart/reset/stop-all)

## Run examples

QA spin with network disabled (default):

```bash
caiged /path/to/workdir --spin qa
```

By default, `caiged` starts the container detached and then opens a host tmux session (if available) with a shell attached to the container. Use `--no-attach` to skip attaching.
Network is enabled by default. Use `--disable-network` to turn it off.

Host tmux sessions:
- Session name: `caiged-<spin>-<project>`
- If already inside tmux, it switches to the session

Session windows:
- `help`
- `opencode`
- `shell`

If you run `caiged` from outside the caiged repo, set the repo path:

```bash
caiged run . --spin qa --repo /path/to/caiged
```

Inside the container:
- `,help` for environment info
- `,auth-tools` to authenticate gh and 1password
- gh config is mounted from the host when available (read-only by default)

Disable network:

```bash
caiged /path/to/workdir --spin qa --disable-network
```

Docker socket is enabled by default. To disable it:

```bash
caiged /path/to/workdir --spin qa --disable-docker-sock
```

Run without attaching (container keeps running):

```bash
caiged "$(pwd)" --spin qa --no-attach
```

Attach to a host tmux session (or container shell fallback):

```bash
caiged session attach caiged-qa-<project>
```

List active caiged containers and tmux sessions:

```bash
caiged session list
```

Restart container and reset tmux session:

```bash
caiged session restart "$(pwd)" --spin qa
```

Reset tmux session only:

```bash
caiged session reset-session "$(pwd)" --spin qa
```

Stop all caiged containers and tmux sessions:

```bash
caiged session stop-all
```

Mount host gh config read-write:

```bash
caiged "$(pwd)" --spin qa --mount-gh-rw
```

Force rebuild:

```bash
OPENCODE_VERSION=latest caiged /path/to/workdir --spin qa --force-build
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
