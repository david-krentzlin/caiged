# Creating Spins

## What is a Spin?

A **spin** is a role-specific container configuration that tailors the AI agent environment for a particular task (QA, engineering, code review, security audit, etc.).

Each spin defines:
- **Agent instructions** (`AGENTS.md` or `AGENT.md`) - the persona, rules, and workflows
- **Skills** (`skills/`) - domain-specific capabilities loaded by OpenCode
- **MCP servers** (`mcp/`) - Model Context Protocol server configurations
- **Documentation** (`README.md`) - spin-specific usage notes

## Directory Structure

```
spins/
└── <spin-name>/
    ├── AGENTS.md          # Agent instructions (preferred, falls back to AGENT.md)
    ├── README.md          # Spin documentation
    ├── skills/            # OpenCode skills
    │   ├── skill-name-1/
    │   │   └── SKILL.md
    │   ├── skill-name-2/
    │   │   └── SKILL.md
    │   └── README.md      # Optional: skills overview
    └── mcp/               # MCP server configs
        ├── config.json    # MCP server definitions
        └── README.md      # Optional: MCP setup notes
```

## File Specifications

### `AGENTS.md` (or `AGENT.md`)

The agent instructions file defines:
- **Purpose**: What role does this agent play?
- **Hard rules**: Non-negotiable constraints (e.g., "no production code changes")
- **Process**: Step-by-step workflows
- **Tone and conduct**: How the agent should communicate
- **Decision criteria**: What makes a good/bad outcome?

**Example structure:**
```markdown
# Agent: <Role Name>

## Purpose
Brief description of what this agent does.

## Hard Rules (Non-Negotiable)
1. Rule one
2. Rule two

## Process
1. Step one
2. Step two

## Operating Mode
How the agent approaches tasks...
```

See `spins/qa/AGENTS.md` for a complete example.

### `skills/`

Skills extend the agent with specialized capabilities. Each skill lives in its own directory:

```
skills/
└── skill-name/
    └── SKILL.md
```

OpenCode loads skills from this directory. Each `SKILL.md` defines:
- Skill name
- Description
- When to use it
- How to invoke it
- Expected inputs/outputs

**Example skill structure:**
```markdown
---
name: security review
description: Evaluate the change for security risks using layered controls and explicit threat modeling.
---

## When to use
Situations where this skill applies...

## What to do
How to invoke and what to expect...
```

See `spins/qa/skills/` for examples (security_review, performance_testing, etc.).

### `mcp/`

Model Context Protocol server configurations. This directory contains JSON files defining MCP servers the agent can use.

**Example `mcp/config.json`:**
```json
{
  "servers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-package"]
    }
  }
}
```

See OpenCode's MCP documentation for full configuration options.

### `README.md`

Spin-specific documentation covering:
- Quick overview of the spin's purpose
- Key skills and capabilities
- Onboarding instructions
- Usage notes and tips

## Creating a New Spin

Creating a new spin is just a matter of creating the directory structure and the files you want to have availble.

### Testing Your Spin

```bash
# Run the spin
caiged run . --spin <spin-name>
```

### Removing a Spin

```bash
rm -rf spins/<name>
docker rmi caiged:<name>
```


## FAQ

**Q: Can I have spin-specific tools/dependencies?**
A: Not directly. Spins share the base image. For tool differences, use conditional logic in scripts or skills.

**Q: Can spins share skills?**
A: Skills are copied per-spin. To share, symlink or copy common skills into multiple spin directories.

**Q: How do I pass data between spins?**
A: Spins run in separate containers. Share data via the mounted workspace or external services.
