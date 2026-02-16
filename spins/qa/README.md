QA Spin

This spin is configured for QA workflows. It ships with:
- QA-oriented instructions in AGENT.md
- QA skills under `spins/qa/skills`
- MCP config under `spins/qa/mcp`
- A tmux session that opens this README on start

Quick onboarding:
- Run `caiged-onboard` to set up gh/op/opencode auth
- Set `OPENCODE_AUTH_CMD` if opencode needs a specific auth command

In-container helpers:
- `,help` for environment info
- `,auth-tools` to authenticate gh and 1password

Notes:
- Network is enabled by default unless `--disable-network` is used
- Docker socket is enabled by default unless `--disable-docker-sock` is used
