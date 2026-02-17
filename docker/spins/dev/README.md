Dev Spin

This spin is configured for development workflows. It ships with:
- Development-oriented instructions in AGENTS.md
- Dev skills under `spins/dev/skills`
- MCP config under `spins/dev/mcp`
- A host tmux session with `help`, `opencode`, and `shell` windows

Quick onboarding:
- OpenCode auth is reused from host `~/.local/share/opencode/auth.json` when available
- If host OpenCode auth is missing, run `/connect` inside the OpenCode TUI
- Pass provider/service secrets using `--secret-env` (for example `JFROG_OIDC_USER`, `JFROG_OIDC_TOKEN`)

In-container helpers:
- `,help` for environment info

Notes:
- Network is enabled by default unless `--disable-network` is used
- Docker socket is disabled by default for security; enable with `--enable-docker-sock` if docker-in-docker is required
